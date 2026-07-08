package frostmage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"strconv"
)

const (
	heroID    = "frostmage"
	passiveID = "frostmage_passive"
	qID       = "frostmage_q"
	wID       = "frostmage_w"
	eID       = "frostmage_e"
	rID       = "frostmage_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		TickEntity:    Tick,
		OnKill:        OnKill,
		OnDamaged:     IgnoreSelfRDamage,
		SpecialRecast: RecastE,
		FrostQDamage:  QDamage,
		FrostQHit:     QHit,
	})
}

func IgnoreSelfRDamage(w *world.World, source *world.Entity, target *world.Entity, basicAttack bool, pet bool, skipBleed bool, tick uint64, tickRate int) {
	if target == nil || target.HeroID != heroID || target.Passive.FrostRSelfUntil == 0 || tick >= target.Passive.FrostRSelfUntil || target.Combat.LastDamage <= 0 {
		return
	}
	target.Stats.HP += float64(target.Combat.LastDamage)
	if target.Stats.HP > target.Stats.MaxHP {
		target.Stats.HP = target.Stats.MaxHP
	}
	target.Combat.LastDamage = 0
	if len(target.Combat.DamageEvents) > 0 {
		target.Combat.DamageEvents[len(target.Combat.DamageEvents)-1].Damage = 0
	}
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.FrostRPending || entity.Passive.FrostRSelfUntil > tick || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < cost {
		return
	}
	if selfCast(entity, cast, skill) {
		entity.Stats.MP -= cost
		state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{120000, 100000, 80000})), tickRate)
		entity.Skills[rID] = state
		startSelfR(w, entity, skill, state.Level, tick, tickRate)
		return
	}
	target := rTarget(w, entity, cast, skill)
	if target == nil {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{120000, 100000, 80000})), tickRate)
	entity.Skills[rID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "enemyCastWindupSeconds", 0.375), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.FrostRPending = true
	entity.Passive.FrostRRelease = tick + windupTicks
	entity.Passive.FrostRTargetID = target.ID
	entity.Passive.FrostRLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.FrostRRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.FrostEPending || entity.Passive.FrostEProjectileID != "" || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 40)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{24000, 21000, 18000, 15000, 12000})), tickRate)
	entity.Skills[eID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.1), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.FrostEPending = true
	entity.Passive.FrostERelease = tick + windupTicks
	entity.Passive.FrostETarget = eTargetPoint(w, entity, cast, skill)
	entity.Passive.FrostELevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.FrostERelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func RecastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || entity.HeroID != heroID || cast.SkillID != eID || entity.Passive.FrostEProjectileID == "" {
		return false
	}
	if tick < entity.Passive.FrostERecastTick {
		return true
	}
	projectile := w.ProjectileByID(entity.Passive.FrostEProjectileID)
	if projectile != nil {
		entity.Position = w.ClampWorldPoint(projectile.Position)
		w.RemoveProjectile(projectile.ID)
	}
	clearE(entity)
	return true
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{14000, 13000, 12000, 11000, 10000})), tickRate)
	entity.Skills[wID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)

	radius := skillMeta(skill, "radius", 450)
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:frostmage_w:"),
		Kind:         "frostmage_w",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + 2,
	})

	raw := skillList(skill, "baseDamage", state.Level, []float64{70, 105, 140, 175, 210}) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.7)
	rootTicks := secondsToTicks(skillList(skill, "rootSeconds", state.Level, []float64{1.1, 1.2, 1.3, 1.4, 1.5}), tickRate)
	for _, target := range w.TargetsInRadius(entity, entity.Position, radius) {
		damage := w.MagicDamageAfterResistance(entity, target, raw, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			target.Control.RootedUntilTick = tick + rootTicks
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(entity, target, damage, tickRate)
		target.Control.RootedUntilTick = tick + world.ControlTicksAfterTenacity(target, rootTicks, tick)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.FrostQPending || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{60, 63, 66, 69, 72})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{8000, 7000, 6000, 5000, 4000})), tickRate)
	entity.Skills[qID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.FrostQPending = true
	entity.Passive.FrostQRelease = tick + windupTicks
	entity.Passive.FrostQTarget = qTargetPoint(w, entity, cast, skill)
	entity.Passive.FrostQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.FrostQRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FrostQPending || tick < entity.Passive.FrostQRelease {
		return
	}
	target := entity.Passive.FrostQTarget
	level := entity.Passive.FrostQLevel
	entity.Passive.FrostQPending = false
	entity.Passive.FrostQRelease = 0
	entity.Passive.FrostQTarget = world.Vector2{}
	entity.Passive.FrostQLevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	speed := skillMeta(skill, "projectileSpeed", 2200)
	shatterDistance := skillMeta(skill, "shatterDistance", 700)
	qRange := skillRange(skill, 825)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:frostmage_q:"),
		Kind:         "frostmage_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 75),
		Damage:       level,
		EffectRatio:  shatterDistance,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	raw := skillList(skill, "baseDamage", skillLevel, []float64{70, 100, 130, 160, 190})
	raw += float64(attacker.Stats.AbilityPower) * skillMeta(skill, "apRatio", 0.8)
	return w.MagicDamageAfterResistance(attacker, target, raw, tick)
}

func QHit(w *world.World, source *world.Entity, target *world.Entity, projectile *world.Projectile, damage int, tick uint64, tickRate int) {
	if source == nil || target == nil || projectile == nil {
		return
	}
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	} else {
		w.ApplyMagicDamage(source, target, damage, tickRate)
	}
	if projectile.Returning {
		return
	}
	skill := w.SkillConfig(qID)
	level := projectile.Damage
	w.ApplyMoveSpeedSlow(target, skillList(skill, "slow", level, []float64{0.16, 0.19, 0.22, 0.25, 0.28}), tick+secondsToTicks(skillMeta(skill, "slowSeconds", 1.5), tickRate))
	fireShard(w, source, target, projectile, skill, tick, tickRate)
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	if w == nil || target == nil || !world.IsHeroUnit(target) {
		return
	}
	tick := target.Combat.LastHitTick
	tickRate := target.Death.RespawnTickRate
	if tickRate <= 0 {
		tickRate = 20
	}
	deathPosition := target.Position
	w.ForEachEntity(func(entity *world.Entity) {
		if entity == nil || entity.HeroID != heroID || entity.Team == target.Team || entity.Stats.HP <= 0 || entity.Death.Dead {
			return
		}
		skill := w.SkillConfig(passiveID)
		if distance(entity.Position, deathPosition) > skillMeta(skill, "spawnRadius", 1000)+target.Radius {
			return
		}
		id := w.NextEffectID("effect:frostmage_servant:")
		entity.Passive.FrostServants = append(entity.Passive.FrostServants, world.FrostServantState{
			ID:        strconv.Itoa(len(entity.Passive.FrostServants) + 1),
			Position:  deathPosition,
			ExpiresAt: tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate),
			EffectID:  id,
		})
		putServantEffect(w, entity, id, deathPosition, skill, tick)
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || tickRate <= 0 {
		return
	}
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		w.RemoveProjectile(entity.Passive.FrostEProjectileID)
		w.RemoveSkillEffect(entity.Passive.FrostRSelfEffectID)
		clearE(entity)
		clearR(entity)
		removeServants(w, entity)
		return
	}
	releaseQ(w, entity, tick, tickRate)
	releaseE(w, entity, tick, tickRate)
	releaseR(w, entity, tick, tickRate)
	tickSelfR(w, entity, tick, tickRate)
	if entity.Passive.FrostEProjectileID != "" && w.ProjectileByID(entity.Passive.FrostEProjectileID) == nil {
		clearE(entity)
	}
	skill := w.SkillConfig(passiveID)
	kept := entity.Passive.FrostServants[:0]
	for _, servant := range entity.Passive.FrostServants {
		if tick >= servant.ExpiresAt {
			explode(w, entity, servant.Position, skill, tick, tickRate)
			w.RemoveSkillEffect(servant.EffectID)
			continue
		}
		servant.Position = moveServant(w, entity, servant.Position, skill, tick, tickRate)
		putServantEffect(w, entity, servant.EffectID, servant.Position, skill, tick)
		kept = append(kept, servant)
	}
	entity.Passive.FrostServants = kept
}

func releaseE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FrostEPending || tick < entity.Passive.FrostERelease {
		return
	}
	target := entity.Passive.FrostETarget
	level := entity.Passive.FrostELevel
	entity.Passive.FrostEPending = false
	entity.Passive.FrostERelease = 0
	entity.Passive.FrostETarget = world.Vector2{}
	entity.Passive.FrostELevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(eID)
	projectileRange := distance(entity.Position, target)
	if projectileRange <= 0 {
		clearE(entity)
		return
	}
	speedMax := skillMeta(skill, "projectileMaxSpeed", 1600)
	id := w.NextProjectileID("projectile:frostmage_e:")
	entity.Passive.FrostEProjectileID = id
	entity.Passive.FrostERecastTick = tick + secondsToTicks(skillMeta(skill, "recastDelaySeconds", 0.5), tickRate)
	w.PutProjectile(&world.Projectile{
		ID:           id,
		Kind:         "frostmage_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      eID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedMax / float64(tickRate),
		SpeedMin:     skillMeta(skill, "projectileMinSpeed", 400),
		SpeedMax:     speedMax,
		Range:        projectileRange,
		Radius:       skillMeta(skill, "projectileRadius", 90),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "durationSeconds", 1.25)+0.5, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func releaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FrostRPending || tick < entity.Passive.FrostRRelease {
		return
	}
	target := w.EntityByID(entity.Passive.FrostRTargetID)
	level := entity.Passive.FrostRLevel
	entity.Passive.FrostRPending = false
	entity.Passive.FrostRRelease = 0
	entity.Passive.FrostRTargetID = ""
	entity.Passive.FrostRLevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead || !validRTarget(entity, target, w.SkillConfig(rID)) {
		return
	}
	skill := w.SkillConfig(rID)
	damage := rDamage(w, entity, target, skill, level, tick)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	} else {
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(entity, target, damage, tickRate)
		target.Control.StunnedUntilTick = tick + world.ControlTicksAfterTenacity(target, secondsToTicks(skillMeta(skill, "stunSeconds", 1.5), tickRate), tick)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	putREffect(w, entity, target.Position, "frostmage_r_enemy", target.Radius, tick, tick+secondsToTicks(0.35, tickRate))
}

func startSelfR(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	durationTicks := secondsToTicks(skillMeta(skill, "selfDurationSeconds", 2.5), tickRate)
	if durationTicks < 1 {
		durationTicks = 1
	}
	missing := 0.0
	if entity.Stats.MaxHP > 0 {
		missing = math.Max(0, entity.Stats.MaxHP-entity.Stats.HP) / entity.Stats.MaxHP
	}
	heal := skillList(skill, "selfHeal", level, []float64{90, 140, 190}) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "selfHealAPRatio", 0.25)
	heal *= 1 + missing
	entity.Passive.FrostRSelfUntil = tick + durationTicks
	entity.Passive.FrostRSelfLevel = level
	entity.Passive.FrostRSelfHealLeft = heal
	entity.Passive.FrostRSelfHealTicks = durationTicks
	entity.Passive.FrostROldDamageReduce = entity.Stats.DamageReduce
	entity.Stats.DamageReduce = 1
	entity.Control.UntargetableUntilTick = entity.Passive.FrostRSelfUntil
	entity.Control.ActionLockedUntilTick = entity.Passive.FrostRSelfUntil
	entity.Passive.FrostRSelfEffectID = putREffect(w, entity, entity.Position, "frostmage_r_self", skillMeta(skill, "selfRadius", 550), tick, entity.Passive.FrostRSelfUntil)
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func tickSelfR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Passive.FrostRSelfUntil == 0 {
		return
	}
	if tick < entity.Passive.FrostRSelfUntil {
		healSelfR(entity)
		return
	}
	healSelfR(entity)
	skill := w.SkillConfig(rID)
	level := entity.Passive.FrostRSelfLevel
	entity.Stats.DamageReduce = entity.Passive.FrostROldDamageReduce
	clearR(entity)
	explodeSelfR(w, entity, skill, level, tick, tickRate)
}

func healSelfR(entity *world.Entity) {
	if entity.Passive.FrostRSelfHealTicks == 0 || entity.Passive.FrostRSelfHealLeft <= 0 {
		return
	}
	heal := entity.Passive.FrostRSelfHealLeft / float64(entity.Passive.FrostRSelfHealTicks)
	entity.Passive.FrostRSelfHealLeft -= heal
	entity.Passive.FrostRSelfHealTicks--
	entity.Stats.HP += heal
	if entity.Stats.HP > entity.Stats.MaxHP {
		entity.Stats.HP = entity.Stats.MaxHP
	}
}

func explodeSelfR(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	radius := skillMeta(skill, "selfRadius", 550)
	slow := skillList(skill, "slow", level, []float64{0.3, 0.45, 0.75})
	slowUntil := tick + secondsToTicks(skillMeta(skill, "slowSeconds", 1.5), tickRate)
	for _, target := range w.TargetsInRadius(entity, entity.Position, radius) {
		damage := rDamage(w, entity, target, skill, level, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			w.ApplyMoveSpeedSlow(target, slow, slowUntil)
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyAOEDamage(entity, target, damage, "magic", tickRate)
		w.ApplyMoveSpeedSlow(target, slow, slowUntil)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
}

func fireShard(w *world.World, source *world.Entity, target *world.Entity, projectile *world.Projectile, skill config.SkillConfig, tick uint64, tickRate int) {
	speed := skillMeta(skill, "projectileSpeed", 2200)
	effectRange := skillMeta(skill, "effectRange", 950)
	rangeLeft := effectRange - distance(projectile.Start, target.Position)
	if rangeLeft <= 0 || tickRate <= 0 {
		return
	}
	hitIDs := map[string]bool{target.ID: true}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:frostmage_q_shard:"),
		Kind:         "frostmage_q_shard",
		Team:         source.Team,
		SourceID:     source.ID,
		SkillID:      qID,
		Position:     target.Position,
		Start:        target.Position,
		Dir:          projectile.Dir,
		SpeedPerTick: speed / float64(tickRate),
		Range:        rangeLeft,
		Radius:       skillMeta(skill, "shardProjectileRadius", 90),
		Damage:       projectile.Damage,
		Returning:    true,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(rangeLeft/speed+0.2, tickRate),
		HitIDs:       hitIDs,
	})
}

func qTargetPoint(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) world.Vector2 {
	target := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	qRange := skillRange(skill, 825)
	if distance(entity.Position, target) <= qRange {
		return target
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		return target
	}
	return w.ClampWorldPoint(world.Vector2{X: entity.Position.X + dx*qRange, Y: entity.Position.Y + dy*qRange})
}

func eTargetPoint(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) world.Vector2 {
	target := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	eRange := skillRange(skill, 1050)
	if distance(entity.Position, target) <= eRange {
		return target
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		return target
	}
	return w.ClampWorldPoint(world.Vector2{X: entity.Position.X + dx*eRange, Y: entity.Position.Y + dy*eRange})
}

func rTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) *world.Entity {
	if target := w.EntityByID(cast.TargetID); validRTarget(entity, target, skill) {
		return target
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	padding := skillMeta(skill, "targetPickPadding", 80)
	w.ForEachEntity(func(target *world.Entity) {
		if !validRTarget(entity, target, skill) {
			return
		}
		dist := distance(point, target.Position)
		if dist > target.Radius+padding || dist >= bestDistance {
			return
		}
		best = target
		bestDistance = dist
	})
	return best
}

func validRTarget(entity *world.Entity, target *world.Entity, skill config.SkillConfig) bool {
	return world.CanAttackTarget(entity, target) && world.IsHeroUnit(target) && distance(entity.Position, target.Position) <= skillRange(skill, 550)+target.Radius
}

func selfCast(entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) bool {
	if entity == nil {
		return false
	}
	if cast.TargetID == entity.ID {
		return true
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	return cast.TargetID == "" && distance(entity.Position, point) <= entity.Radius+skillMeta(skill, "targetPickPadding", 80)
}

func rDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	raw := skillList(skill, "baseDamage", level, []float64{150, 250, 350}) + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.75)
	return w.MagicDamageAfterResistance(attacker, target, raw, tick)
}

func putREffect(w *world.World, entity *world.Entity, center world.Vector2, kind string, radius float64, tick uint64, expiresAt uint64) string {
	id := w.NextEffectID("effect:" + kind + ":")
	w.PutSkillEffect(world.SkillEffect{
		ID:           id,
		Kind:         kind,
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        center,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
	return id
}

func moveServant(w *world.World, source *world.Entity, position world.Vector2, skill config.SkillConfig, tick uint64, tickRate int) world.Vector2 {
	target := servantTarget(w, source, position, skillMeta(skill, "seekRadius", 700))
	if target == nil {
		return position
	}
	step := skillMeta(skill, "moveSpeed", 325) / float64(tickRate)
	dist := distance(position, target.Position)
	if dist > 0 && step > 0 {
		ratio := math.Min(1, step/dist)
		position.X += (target.Position.X - position.X) * ratio
		position.Y += (target.Position.Y - position.Y) * ratio
	}
	if distance(position, target.Position) <= skillMeta(skill, "slowRadius", 180)+target.Radius {
		w.ApplyMoveSpeedSlow(target, skillMeta(skill, "slow", 0.3), tick+secondsToTicks(skillMeta(skill, "slowSeconds", 0.25), tickRate))
	}
	return position
}

func servantTarget(w *world.World, source *world.Entity, position world.Vector2, seekRadius float64) *world.Entity {
	var best *world.Entity
	bestDistance := math.Inf(1)
	bestIsHero := false
	w.ForEachEntity(func(target *world.Entity) {
		if !world.CanAttackTarget(source, target) {
			return
		}
		dist := distance(position, target.Position)
		if dist > seekRadius+target.Radius {
			return
		}
		isHero := world.IsHeroUnit(target)
		if best != nil && (bestIsHero && !isHero || bestIsHero == isHero && dist >= bestDistance) {
			return
		}
		best = target
		bestDistance = dist
		bestIsHero = isHero
	})
	return best
}

func explode(w *world.World, source *world.Entity, center world.Vector2, skill config.SkillConfig, tick uint64, tickRate int) {
	raw := float64(source.Stats.AbilityPower) * skillMeta(skill, "apRatio", 0.5)
	for _, target := range w.TargetsInRadius(source, center, skillMeta(skill, "explosionRadius", 280)) {
		damage := w.MagicDamageAfterResistance(source, target, raw, tick)
		target.Combat.LastHitTick = tick
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyPetDamage(source, target, damage, "magic", tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
}

func putServantEffect(w *world.World, source *world.Entity, id string, position world.Vector2, skill config.SkillConfig, tick uint64) {
	w.PutSkillEffect(world.SkillEffect{
		ID:           id,
		Kind:         "frostmage_servant",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		Start:        position,
		Radius:       skillMeta(skill, "slowRadius", 180),
		CreatedAt:    tick,
		ExpiresAt:    tick + 1,
	})
}

func removeServants(w *world.World, entity *world.Entity) {
	for _, servant := range entity.Passive.FrostServants {
		w.RemoveSkillEffect(servant.EffectID)
	}
	entity.Passive.FrostServants = nil
}

func clearE(entity *world.Entity) {
	entity.Passive.FrostEPending = false
	entity.Passive.FrostERelease = 0
	entity.Passive.FrostETarget = world.Vector2{}
	entity.Passive.FrostELevel = 0
	entity.Passive.FrostEProjectileID = ""
	entity.Passive.FrostERecastTick = 0
}

func clearR(entity *world.Entity) {
	if entity.Passive.FrostRSelfUntil > 0 {
		entity.Stats.DamageReduce = entity.Passive.FrostROldDamageReduce
	}
	entity.Passive.FrostRPending = false
	entity.Passive.FrostRRelease = 0
	entity.Passive.FrostRTargetID = ""
	entity.Passive.FrostRLevel = 0
	entity.Passive.FrostRSelfUntil = 0
	entity.Passive.FrostRSelfLevel = 0
	entity.Passive.FrostRSelfEffectID = ""
	entity.Passive.FrostRSelfHealLeft = 0
	entity.Passive.FrostRSelfHealTicks = 0
	entity.Passive.FrostROldDamageReduce = 0
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if skill.MetaLists != nil && len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
	}
	if level < 1 {
		level = 1
	}
	if level > len(values) {
		level = len(values)
	}
	return values[level-1]
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}
