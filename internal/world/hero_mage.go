package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
	"strconv"
)

func (w *World) applyMageQ(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.LightBindingPending {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.LightBindingPending = true
	entity.Mage.LightBindingReleaseTick = tick + windupTicks
	entity.Mage.LightBindingTarget = Vector2{
		X: clamp(cast.TargetX, 0, w.width),
		Y: clamp(cast.TargetY, 0, w.height),
	}
	entity.Mage.LightBindingLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.LightBindingReleaseTick
	entity.Skills[mageQSkillID] = state
}

func (w *World) releaseMageQ(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.LightBindingPending || tick < entity.Mage.LightBindingReleaseTick {
		return
	}
	targetPoint := entity.Mage.LightBindingTarget
	level := entity.Mage.LightBindingLevel
	entity.Mage.LightBindingPending = false
	entity.Mage.LightBindingReleaseTick = 0
	entity.Mage.LightBindingTarget = Vector2{}
	entity.Mage.LightBindingLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.skillConfig(mageQSkillID)
	state := entity.Skills[mageQSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{15000, 14000, 13000, 12000, 11000}), tickRate)
	entity.Skills[mageQSkillID] = state
	qRange := skillRange(skill, 1175)
	speedPerSecond := skillMetaRange(skill, "projectileSpeed", 1400)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:mage_q:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "mage_light_binding",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      mageQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMetaRange(skill, "projectileRadius", 45),
		Damage:       level,
		EffectTicks:  secondsToTicks(skillMetaRange(skill, "rootSeconds", 2), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) applyMageW(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.PrismaticBarrierPending {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 60)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.2), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.PrismaticBarrierPending = true
	entity.Mage.PrismaticBarrierReleaseTick = tick + windupTicks
	entity.Mage.PrismaticBarrierTarget = Vector2{
		X: clamp(cast.TargetX, 0, w.width),
		Y: clamp(cast.TargetY, 0, w.height),
	}
	entity.Mage.PrismaticBarrierLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.PrismaticBarrierReleaseTick
	entity.Skills[mageWSkillID] = state
}

func (w *World) releaseMageW(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.PrismaticBarrierPending || tick < entity.Mage.PrismaticBarrierReleaseTick {
		return
	}
	targetPoint := entity.Mage.PrismaticBarrierTarget
	level := entity.Mage.PrismaticBarrierLevel
	entity.Mage.PrismaticBarrierPending = false
	entity.Mage.PrismaticBarrierReleaseTick = 0
	entity.Mage.PrismaticBarrierTarget = Vector2{}
	entity.Mage.PrismaticBarrierLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.skillConfig(mageWSkillID)
	state := entity.Skills[mageWSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{14000, 13000, 12000, 11000, 10000}), tickRate)
	entity.Skills[mageWSkillID] = state
	w.addMageShieldLayer(entity, mageWShieldValue(entity, skill, level), tick+secondsToTicks(skillMetaRange(skill, "shieldSeconds", 3), tickRate))
	w.spawnMageWProjectile(entity, Vector2{X: dx, Y: dy}, level, skill, tick, tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) spawnMageWProjectile(entity *Entity, dir Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	wRange := skillRange(skill, 1075)
	speedPerSecond := skillMetaRange(skill, "projectileSpeed", 1450)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = wRange
	}
	lifeTicks := secondsToTicks(10, tickRate)
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:mage_w:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "mage_prismatic_barrier",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      mageWSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          dir,
		SpeedPerTick: speedPerTick,
		SpeedMin:     speedPerTick,
		Range:        wRange * 2,
		Radius:       skillMetaRange(skill, "projectileRadius", 55),
		Damage:       level,
		EffectTicks:  secondsToTicks(skillMetaRange(skill, "shieldSeconds", 3), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
}

func mageWShieldValue(entity *Entity, skill config.SkillConfig, level int) int {
	baseShield := skillMetaListByLevel(skill, "shieldValue", level, []float64{50, 65, 80, 95, 110})
	return int(math.Round(baseShield + float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.2)))
}

func (w *World) applyMageE(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.LucentSingularityPending || entity.Mage.LucentSingularityActive || entity.Mage.LucentSingularityProjectileID != "" {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{70, 80, 90, 100, 110})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	target := Vector2{X: clamp(cast.TargetX, 0, w.width), Y: clamp(cast.TargetY, 0, w.height)}
	if distance(entity.Position, target) > skillRange(skill, 1100) {
		dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
		target = Vector2{X: entity.Position.X + dx*skillRange(skill, 1100), Y: entity.Position.Y + dy*skillRange(skill, 1100)}
	}
	entity.Mage.LucentSingularityPending = true
	entity.Mage.LucentSingularityReleaseTick = tick + windupTicks
	entity.Mage.LucentSingularityTarget = target
	entity.Mage.LucentSingularityLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.LucentSingularityReleaseTick
	entity.Skills[mageESkillID] = state
}

func (w *World) applyMageR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.FinalSparkPending {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.5), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.FinalSparkPending = true
	entity.Mage.FinalSparkReleaseTick = tick + windupTicks
	entity.Mage.FinalSparkTarget = Vector2{X: clamp(cast.TargetX, 0, w.width), Y: clamp(cast.TargetY, 0, w.height)}
	entity.Mage.FinalSparkLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.FinalSparkReleaseTick
	entity.Skills[mageRSkillID] = state
}

func (w *World) releaseMageR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.FinalSparkPending || tick < entity.Mage.FinalSparkReleaseTick {
		return
	}
	targetPoint := entity.Mage.FinalSparkTarget
	level := entity.Mage.FinalSparkLevel
	entity.Mage.FinalSparkPending = false
	entity.Mage.FinalSparkReleaseTick = 0
	entity.Mage.FinalSparkTarget = Vector2{}
	entity.Mage.FinalSparkLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.skillConfig(mageRSkillID)
	state := entity.Skills[mageRSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{80000, 65000, 50000}), tickRate)
	entity.Skills[mageRSkillID] = state
	w.addMageREffect(entity, Vector2{X: dx, Y: dy}, skillRange(skill, 3400), skillMetaRange(skill, "beamWidth", 200), tick, tickRate)
	for _, target := range w.mageRTargets(entity, Vector2{X: dx, Y: dy}, skill, tick) {
		damage := mageRDamage(entity, target, skill, level, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyAOEDamage(entity, target, damage, "magic", tickRate)
			w.applyMageIlluminationOnUltimateHit(entity, target, tick, tickRate)
			if target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero {
				target.Control.MageFinalSparkBy = entity.ID
				target.Control.MageFinalSparkUntil = tick + secondsToTicks(skillMetaRange(skill, "refundWindowSeconds", 1.75), tickRate)
				target.Control.MageFinalSparkRefund = skillMetaListByLevel(skill, "cooldownRefund", level, []float64{0.3, 0.4, 0.5})
			}
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			w.applyMageIlluminationOnUltimateHit(entity, target, tick, tickRate)
		}
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) releaseMageE(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.LucentSingularityPending || tick < entity.Mage.LucentSingularityReleaseTick {
		return
	}
	level := entity.Mage.LucentSingularityLevel
	center := entity.Mage.LucentSingularityTarget
	entity.Mage.LucentSingularityPending = false
	entity.Mage.LucentSingularityReleaseTick = 0
	entity.Mage.LucentSingularityTarget = Vector2{}
	if level <= 0 {
		level = 1
	}
	skill := w.skillConfig(mageESkillID)
	w.spawnMageEProjectile(entity, center, level, skill, tick, tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) spawnMageEProjectile(entity *Entity, target Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	travelRange := distance(entity.Position, target)
	if travelRange <= 0 || dx == 0 && dy == 0 {
		w.activateMageEZone(entity, target, level, skill, tick, tickRate)
		return
	}
	speedPerTick := skillMetaRange(skill, "projectileSpeed", 1200) / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = travelRange
	}
	lifeTicks := uint64(math.Ceil(travelRange / speedPerTick))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:mage_e:" + strconv.Itoa(w.nextProjectileID)
	entity.Mage.LucentSingularityProjectileID = id
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "mage_lucent_singularity_orb",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      mageESkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        travelRange,
		Radius:       skillMetaRange(skill, "projectileRadius", 34),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 2,
		HitIDs:       make(map[string]bool),
	}
}

func (w *World) activateMageEZone(entity *Entity, center Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID {
		return
	}
	entity.Mage.LucentSingularityProjectileID = ""
	entity.Mage.LucentSingularityActive = true
	entity.Mage.LucentSingularityCenter = center
	entity.Mage.LucentSingularityExpireTick = tick + secondsToTicks(skillMetaRange(skill, "durationSeconds", 5), tickRate)
	entity.Mage.LucentSingularityLevel = level
	entity.Mage.LucentSingularityEffectID = w.addMageEEffect(entity, center, skillMetaRange(skill, "radius", 300), tick, entity.Mage.LucentSingularityExpireTick)
}

func (w *World) tickMageE(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.LucentSingularityActive {
		return
	}
	skill := w.skillConfig(mageESkillID)
	if tick >= entity.Mage.LucentSingularityExpireTick {
		w.detonateMageE(entity, skill, tick, tickRate)
		return
	}
	slow := skillMetaListByLevel(skill, "slow", entity.Mage.LucentSingularityLevel, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	for _, target := range w.targetsInRadius(entity, entity.Mage.LucentSingularityCenter, skillMetaRange(skill, "radius", 300)) {
		applyMoveSpeedSlow(target, slow, tick+2)
	}
}

func (w *World) detonateMageE(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || !entity.Mage.LucentSingularityActive {
		return
	}
	center := entity.Mage.LucentSingularityCenter
	level := entity.Mage.LucentSingularityLevel
	if level <= 0 {
		level = 1
	}
	state := entity.Skills[mageESkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{10000, 9500, 9000, 8500, 8000}), tickRate)
	entity.Skills[mageESkillID] = state
	delete(w.skillEffects, entity.Mage.LucentSingularityEffectID)
	entity.Mage.LucentSingularityActive = false
	entity.Mage.LucentSingularityCenter = Vector2{}
	entity.Mage.LucentSingularityExpireTick = 0
	entity.Mage.LucentSingularityLevel = 0
	entity.Mage.LucentSingularityEffectID = ""
	rawDamage := mageERawDamage(entity, skill, level)
	slow := skillMetaListByLevel(skill, "slow", level, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	slowUntil := tick + secondsToTicks(skillMetaRange(skill, "detonateSlowSeconds", 1), tickRate)
	for _, target := range w.targetsInRadius(entity, center, skillMetaRange(skill, "radius", 300)) {
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyAOEDamage(entity, target, magicDamageAfterResistance(entity, target, rawDamage, tick), "magic", tickRate)
			applyMoveSpeedSlow(target, slow, slowUntil)
			w.applyMageIlluminationOnSkillHit(entity, target, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, rawDamage, tick)
			target.Combat.LastDamageType = "magic"
			applyMoveSpeedSlow(target, slow, slowUntil)
			w.applyMageIlluminationOnSkillHit(entity, target, tick, tickRate)
		}
	}
}

func mageERawDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{80, 120, 160, 200, 240}) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.8)
}

func (w *World) mageRTargets(entity *Entity, direction Vector2, skill config.SkillConfig, tick uint64) []*Entity {
	hits := make([]*Entity, 0)
	castRange := skillRange(skill, 3400)
	halfWidth := skillMetaRange(skill, "beamWidth", 200) / 2
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(entity.Position, direction, target.Position)
		if along < -target.Radius || along > castRange+target.Radius {
			continue
		}
		if perpendicular <= halfWidth+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func mageRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{300, 400, 500})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.75)
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func (w *World) addMageREffect(entity *Entity, direction Vector2, beamRange float64, beamWidth float64, tick uint64, tickRate int) {
	w.nextEffectID++
	id := "effect:mage_r:" + strconv.Itoa(w.nextEffectID)
	lifeTicks := secondsToTicks(0.25, tickRate)
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "mage_final_spark",
		Team:      entity.Team,
		Start:     entity.Position,
		End:       Vector2{X: clamp(entity.Position.X+direction.X*beamRange, 0, w.width), Y: clamp(entity.Position.Y+direction.Y*beamRange, 0, w.height)},
		Dir:       direction,
		Range:     beamRange,
		Width:     beamWidth,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	}
}

func (w *World) applyMageFinalSparkRefund(target *Entity) {
	if target == nil || target.Control.MageFinalSparkBy == "" || target.Combat.LastHitTick > target.Control.MageFinalSparkUntil {
		return
	}
	caster := w.entities[target.Control.MageFinalSparkBy]
	if caster == nil {
		return
	}
	state := caster.Skills[mageRSkillID]
	remaining := int64(state.CooldownUntilTick) - int64(target.Combat.LastHitTick)
	if remaining <= 0 {
		return
	}
	refund := uint64(math.Round(float64(remaining) * clamp(target.Control.MageFinalSparkRefund, 0, 1)))
	if refund >= uint64(remaining) {
		state.CooldownUntilTick = target.Combat.LastHitTick
	} else {
		state.CooldownUntilTick -= refund
	}
	caster.Skills[mageRSkillID] = state
	target.Control.MageFinalSparkBy = ""
	target.Control.MageFinalSparkUntil = 0
	target.Control.MageFinalSparkRefund = 0
}

func (w *World) addMageEEffect(entity *Entity, center Vector2, radius float64, tick uint64, expiresAt uint64) string {
	w.nextEffectID++
	id := "effect:mage_e:" + strconv.Itoa(w.nextEffectID)
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "mage_lucent_singularity",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: expiresAt,
	}
	return id
}

func (w *World) addMageShieldLayer(target *Entity, amount int, expiresAt uint64) {
	if target == nil || amount <= 0 || expiresAt == 0 {
		return
	}
	target.Passive.ShieldLayers = append(target.Passive.ShieldLayers, ShieldLayer{Amount: amount, ExpiresAt: expiresAt})
	target.Passive.Shield += amount
	target.Passive.MaxShield += amount
}

func (w *World) applyMageIlluminationOnSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !canApplyMageIllumination(source, target) {
		return
	}
	skill := w.heroPassiveSkill(source)
	target.Control.MageIlluminationBy = source.ID
	target.Control.MageIlluminationUntil = tick + secondsToTicks(skillMetaRange(skill, "debuffSeconds", 6), tickRate)
}

func (w *World) applyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !canApplyMageIllumination(source, target) {
		return
	}
	w.detonateMageIllumination(source, target, tick, tickRate)
	w.applyMageIlluminationOnSkillHit(source, target, tick, tickRate)
}

func (w *World) triggerMageIlluminationOnBasicAttack(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != mageHeroID {
		return
	}
	w.detonateMageIllumination(source, target, tick, tickRate)
}

func canApplyMageIllumination(source *Entity, target *Entity) bool {
	if source == nil || target == nil || source.HeroID != mageHeroID {
		return false
	}
	if source.ID == target.ID || target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
		return false
	}
	return target.Team != source.Team || target.Team == TeamNeutral
}

func (w *World) detonateMageIllumination(source *Entity, target *Entity, tick uint64, tickRate int) bool {
	if !mageIlluminationActive(source, target, tick) {
		return false
	}
	target.Control.MageIlluminationBy = ""
	target.Control.MageIlluminationUntil = 0
	damage := w.mageIlluminationDamage(source, target, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyMagicDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(source, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	}
	return true
}

func mageIlluminationActive(source *Entity, target *Entity, tick uint64) bool {
	return source != nil &&
		target != nil &&
		target.Control.MageIlluminationBy == source.ID &&
		target.Control.MageIlluminationUntil > tick
}

func (w *World) mageIlluminationDamage(source *Entity, target *Entity, tick uint64) int {
	skill := w.heroPassiveSkill(source)
	baseDamage := skillMetaCurveByLevel(skill, "detonateDamage", "detonateDamageLevels", source.Level, 20)
	rawDamage := baseDamage + float64(source.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.3)
	return magicDamageAfterResistance(source, target, rawDamage, tick)
}

func mageQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{50, 100, 150, 200, 250})
	rawDamage := (baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.7)) * multiplier
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}
