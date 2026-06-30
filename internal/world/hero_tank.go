package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
	"strconv"
)

func (w *World) applyTankQ(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	if entity.Tank.SeismicShardPending {
		return
	}
	target := w.entities[cast.TargetID]
	if target == nil {
		target = w.tankQTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	}
	if !canAttackTarget(entity, target) {
		return
	}
	if distance(entity.Position, target.Position) > skillRange(skill, 625)+target.Radius {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{70, 75, 80, 85, 90})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Tank.SeismicShardPending = true
	entity.Tank.SeismicShardReleaseTick = tick + windupTicks
	entity.Tank.SeismicShardTargetID = target.ID
	entity.Tank.SeismicShardLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Tank.SeismicShardReleaseTick
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000}), tickRate)
	entity.Skills[tankQSkillID] = state
}

func (w *World) releaseTankQ(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.SeismicShardPending || tick < entity.Tank.SeismicShardReleaseTick {
		return
	}
	target := w.entities[entity.Tank.SeismicShardTargetID]
	level := entity.Tank.SeismicShardLevel
	entity.Tank.SeismicShardPending = false
	entity.Tank.SeismicShardReleaseTick = 0
	entity.Tank.SeismicShardTargetID = ""
	entity.Tank.SeismicShardLevel = 0
	if !canAttackTarget(entity, target) {
		return
	}
	skill := w.skillConfig(tankQSkillID)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	qRange := skillRange(skill, 625)
	speedPerSecond := skillMetaRange(skill, "projectileSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:tank_q:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "tank_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      tankQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMetaRange(skill, "projectileRadius", 45),
		Damage:       level,
		EffectRatio:  skillMetaListByLevel(skill, "moveSpeedSteal", level, []float64{0.2, 0.25, 0.3, 0.35, 0.4}),
		EffectTicks:  secondsToTicks(skillMetaRange(skill, "moveSpeedStealSeconds", 3), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) tankQTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 625)
	pickPadding := skillMetaRange(skill, "targetPickPadding", 90)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) > castRange+target.Radius {
			continue
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+pickPadding {
			continue
		}
		if distToPoint < bestDistance {
			bestDistance = distToPoint
			best = target
		}
	}
	return best
}

func (w *World) applyTankW(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{30, 35, 40, 45, 50})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Tank.ThunderclapEmpowerUntil = tick + secondsToTicks(skillMetaRange(skill, "aftershockDurationSeconds", 5), tickRate)
	entity.Tank.ThunderclapAftershockUntil = entity.Tank.ThunderclapEmpowerUntil
	entity.Tank.ThunderclapLevel = state.Level
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{10000, 9500, 9000, 8500, 8000}), tickRate)
	entity.Skills[tankWSkillID] = state
}

func (w *World) refreshTankWPassive(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	if entity.Tank.ThunderclapArmorBonus != 0 {
		entity.Stats.PhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Stats.BonusPhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Tank.ThunderclapArmorBonus = 0
	}
	state, ok := entity.Skills[tankWSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.skillConfig(tankWSkillID)
	ratio := skillMetaListByLevel(skill, "passiveArmorRatio", state.Level, []float64{0.1, 0.15, 0.2, 0.25, 0.3})
	if entity.Passive.Shield > 0 {
		ratio = skillMetaRange(skill, "shieldArmorRatio", 0.3)
	}
	baseArmor := entity.Stats.PhysicalDefense - entity.Stats.BonusPhysicalDefense
	bonus := baseArmor * ratio
	entity.Tank.ThunderclapArmorBonus = bonus
	entity.Stats.PhysicalDefense += bonus
	entity.Stats.BonusPhysicalDefense += bonus
}

func (w *World) applyTankE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	damage := tankEDamage(entity, skill, state.Level)
	slow := skillMetaListByLevel(skill, "attackSpeedSlow", state.Level, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	slowUntil := tick + secondsToTicks(skillMetaRange(skill, "attackSpeedSlowSeconds", 3), tickRate)
	for _, target := range w.targetsInRadius(entity, entity.Position, skillRange(skill, 400)) {
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyMagicDamage(entity, target, magicDamageAfterResistance(entity, target, damage, tick), tickRate)
			applyAttackSpeedSlow(target, slow, slowUntil)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			applyAttackSpeedSlow(target, slow, slowUntil)
		}
	}
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{7000, 7000, 7000, 7000, 7000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[tankESkillID] = state
}

func tankEDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{60, 95, 130, 165, 200}) +
		entity.Stats.PhysicalDefense*skillMetaRange(skill, "armorRatio", 0.4) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
}

func applyAttackSpeedSlow(target *Entity, slow float64, until uint64) {
	if target == nil || slow <= 0 || until == 0 {
		return
	}
	target.Control.AttackSpeedSlow = clamp(slow, 0, 1)
	target.Control.AttackSpeedSlowUntil = until
}

func (w *World) applyTankR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	start := entity.Position
	rRange := skillRange(skill, 1000)
	targetPoint := Vector2{
		X: clamp(cast.TargetX, 0, w.width),
		Y: clamp(cast.TargetY, 0, w.height),
	}
	if distance(start, targetPoint) > rRange {
		dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
		if dx == 0 && dy == 0 {
			return
		}
		castPosition := Vector2{
			X: clamp(targetPoint.X-dx*rRange, 0, w.width),
			Y: clamp(targetPoint.Y-dy*rRange, 0, w.height),
		}
		entity.Tank.UnstoppableCastPending = true
		entity.Tank.UnstoppableCastTarget = targetPoint
		entity.Tank.UnstoppableCastLevel = state.Level
		entity.Intent.MoveTarget = &castPosition
		entity.Intent.AttackTargetID = ""
		entity.Intent.AttackPausedTill = 0
		return
	}
	w.startTankRDash(entity, targetPoint, state, skill, tick, tickRate)
}

func (w *World) startTankRDash(entity *Entity, targetPoint Vector2, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	start := entity.Position
	dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
	if dx == 0 && dy == 0 {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	rRange := skillRange(skill, 1000)
	if distance(start, targetPoint) > rRange {
		targetPoint = Vector2{
			X: clamp(start.X+dx*rRange, 0, w.width),
			Y: clamp(start.Y+dy*rRange, 0, w.height),
		}
	}
	entity.Stats.MP -= manaCost
	entity.Intent = IntentState{}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
	landingRadius := skillMetaRange(skill, "landingRadius", 250)
	dashSpeed := skillMetaRange(skill, "dashSpeed", 1600)
	if dashSpeed <= 0 {
		dashSpeed = rRange
	}
	travelTicks := uint64(math.Ceil(distance(start, targetPoint) / dashSpeed * float64(tickRate)))
	if travelTicks < 1 {
		travelTicks = 1
	}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = start
	entity.Control.DashEnd = targetPoint
	entity.Control.DashUntilTick = tick + travelTicks
	entity.Control.ActionLockedUntilTick = entity.Control.DashUntilTick
	entity.Tank.UnstoppableImpactPending = true
	entity.Tank.UnstoppableImpactTick = entity.Control.DashUntilTick
	entity.Tank.UnstoppableImpactLevel = state.Level
	entity.Tank.UnstoppableImpactRadius = landingRadius
	entity.Tank.UnstoppableKnockupTicks = secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1.5), tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{130000, 105000, 80000}), tickRate)
	entity.Skills[tankRSkillID] = state
}

func (w *World) releasePreparedTankR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	state := entity.Skills[tankRSkillID]
	if state.Level <= 0 {
		w.cancelTankRPreparedCast(entity)
		return
	}
	if state.CooldownUntilTick > tick {
		w.cancelTankRPreparedCast(entity)
		return
	}
	state.Level = entity.Tank.UnstoppableCastLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	w.startTankRDash(entity, entity.Tank.UnstoppableCastTarget, state, w.skillConfig(tankRSkillID), tick, tickRate)
}

func (w *World) cancelTankRPreparedCast(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
}

func (w *World) resolveTankRImpact(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableImpactPending {
		return
	}
	if tick < entity.Tank.UnstoppableImpactTick {
		return
	}
	skill := w.skillConfig(tankRSkillID)
	level := entity.Tank.UnstoppableImpactLevel
	if level <= 0 {
		level = 1
	}
	radius := entity.Tank.UnstoppableImpactRadius
	if radius <= 0 {
		radius = skillMetaRange(skill, "landingRadius", 250)
	}
	knockupTicks := entity.Tank.UnstoppableKnockupTicks
	if knockupTicks == 0 {
		knockupTicks = secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1.5), tickRate)
	}
	damage := tankRDamage(entity, skill, level)
	w.addTankRImpactEffect(entity, entity.Position, radius, tick, tickRate)
	for _, target := range w.targetsInRadius(entity, entity.Position, radius) {
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyMagicDamage(entity, target, magicDamageAfterResistance(entity, target, damage, tick), tickRate)
			target.Control.AirborneUntilTick = tick + controlTicksAfterTenacity(target, knockupTicks, tick)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			target.Control.AirborneUntilTick = tick + knockupTicks
		}
	}
	entity.Tank.UnstoppableImpactPending = false
	entity.Tank.UnstoppableImpactTick = 0
	entity.Tank.UnstoppableImpactLevel = 0
	entity.Tank.UnstoppableImpactRadius = 0
	entity.Tank.UnstoppableKnockupTicks = 0
}

func tankRDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{200, 300, 400}) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.8)
}

func (w *World) addTankRImpactEffect(entity *Entity, center Vector2, radius float64, tick uint64, tickRate int) {
	if entity == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.35))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.nextEffectID++
	id := "effect:tank_r_impact:" + strconv.Itoa(w.nextEffectID)
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "tank_r_impact",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	}
}
