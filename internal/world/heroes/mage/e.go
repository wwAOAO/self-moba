package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ReleaseE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Mage.LucentSingularityPending || tick < entity.Mage.LucentSingularityReleaseTick {
		return
	}
	level := entity.Mage.LucentSingularityLevel
	center := entity.Mage.LucentSingularityTarget
	entity.Mage.LucentSingularityPending = false
	entity.Mage.LucentSingularityReleaseTick = 0
	entity.Mage.LucentSingularityTarget = world.Vector2{}
	if level <= 0 {
		level = 1
	}
	skill := w.SkillConfig(eID)
	spawnEProjectile(w, entity, center, level, skill, tick, tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func spawnEProjectile(w *world.World, entity *world.Entity, target world.Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	travelRange := distance(entity.Position, target)
	if travelRange <= 0 || dx == 0 && dy == 0 {
		ActivateEZone(w, entity, target, level, skill, tick, tickRate)
		return
	}
	speedPerTick := skillMeta(skill, "projectileSpeed", 1200) / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = travelRange
	}
	lifeTicks := uint64(math.Ceil(travelRange / speedPerTick))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	id := w.NextProjectileID("projectile:mage_e:")
	entity.Mage.LucentSingularityProjectileID = id
	w.PutProjectile(&world.Projectile{
		ID:           id,
		Kind:         "mage_lucent_singularity_orb",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      eID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        travelRange,
		Radius:       skillMeta(skill, "projectileRadius", 34),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 2,
		HitIDs:       make(map[string]bool),
	})
}

func ActivateEZone(w *world.World, entity *world.Entity, center world.Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	entity.Mage.LucentSingularityProjectileID = ""
	entity.Mage.LucentSingularityActive = true
	entity.Mage.LucentSingularityCenter = center
	entity.Mage.LucentSingularityExpireTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 5), tickRate)
	entity.Mage.LucentSingularityLevel = level
	entity.Mage.LucentSingularityEffectID = addEEffect(w, entity, center, skillMeta(skill, "radius", 300), tick, entity.Mage.LucentSingularityExpireTick)
}

func TickE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Mage.LucentSingularityActive {
		return
	}
	skill := w.SkillConfig(eID)
	if tick >= entity.Mage.LucentSingularityExpireTick {
		DetonateE(w, entity, skill, tick, tickRate)
		return
	}
	slow := skillList(skill, "slow", entity.Mage.LucentSingularityLevel, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	for _, target := range w.MageTargetsInRadius(entity, entity.Mage.LucentSingularityCenter, skillMeta(skill, "radius", 300)) {
		w.ApplyMageMoveSpeedSlow(target, slow, tick+2)
	}
}

func DetonateE(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || !entity.Mage.LucentSingularityActive {
		return
	}
	center := entity.Mage.LucentSingularityCenter
	level := entity.Mage.LucentSingularityLevel
	if level <= 0 {
		level = 1
	}
	state := entity.Skills[eID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{10000, 9500, 9000, 8500, 8000})), tickRate)
	entity.Skills[eID] = state
	w.RemoveSkillEffect(entity.Mage.LucentSingularityEffectID)
	entity.Mage.LucentSingularityActive = false
	entity.Mage.LucentSingularityCenter = world.Vector2{}
	entity.Mage.LucentSingularityExpireTick = 0
	entity.Mage.LucentSingularityLevel = 0
	entity.Mage.LucentSingularityEffectID = ""
	rawDamage := eRawDamage(entity, skill, level)
	slow := skillList(skill, "slow", level, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	slowUntil := tick + secondsToTicks(skillMeta(skill, "detonateSlowSeconds", 1), tickRate)
	for _, target := range w.MageTargetsInRadius(entity, center, skillMeta(skill, "radius", 300)) {
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		damage := w.MageMagicDamageAfterResistance(entity, target, rawDamage, tick)
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, damage, "magic", tickRate)
			w.ApplyMageMoveSpeedSlow(target, slow, slowUntil)
			w.ApplyMageIlluminationOnSkillHit(entity, target, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			w.ApplyMageMoveSpeedSlow(target, slow, slowUntil)
			w.ApplyMageIlluminationOnSkillHit(entity, target, tick, tickRate)
		}
	}
}

func eRawDamage(entity *world.Entity, skill config.SkillConfig, level int) float64 {
	return skillList(skill, "baseDamage", level, []float64{80, 120, 160, 200, 240}) +
		float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.8)
}

func addEEffect(w *world.World, entity *world.Entity, center world.Vector2, radius float64, tick uint64, expiresAt uint64) string {
	id := w.NextEffectID("effect:mage_e:")
	w.PutSkillEffect(world.SkillEffect{
		ID:        id,
		Kind:      "mage_lucent_singularity",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: expiresAt,
	})
	return id
}

func ApplyE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.LucentSingularityPending || entity.Mage.LucentSingularityActive || entity.Mage.LucentSingularityProjectileID != "" {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{70, 80, 90, 100, 110})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	target := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	if distance(entity.Position, target) > skillRange(skill, 1100) {
		dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
		target = world.Vector2{X: entity.Position.X + dx*skillRange(skill, 1100), Y: entity.Position.Y + dy*skillRange(skill, 1100)}
	}
	entity.Mage.LucentSingularityPending = true
	entity.Mage.LucentSingularityReleaseTick = tick + windupTicks
	entity.Mage.LucentSingularityTarget = target
	entity.Mage.LucentSingularityLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.LucentSingularityReleaseTick
	entity.Skills[eID] = state
}
