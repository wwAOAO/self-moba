package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

func applyMageE(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
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
	target := w.ClampWorldPoint(Vector2{X: cast.TargetX, Y: cast.TargetY})
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
	skill := w.SkillConfig(mageESkillID)
	w.spawnMageEProjectile(entity, center, level, skill, tick, tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
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
	id := w.NextProjectileID("projectile:mage_e:")
	entity.Mage.LucentSingularityProjectileID = id
	w.PutProjectile(&Projectile{
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
	})
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
	skill := w.SkillConfig(mageESkillID)
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
	w.RemoveSkillEffect(entity.Mage.LucentSingularityEffectID)
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
			w.ApplyAOEDamage(entity, target, magicDamageAfterResistance(entity, target, rawDamage, tick), "magic", tickRate)
			applyMoveSpeedSlow(target, slow, slowUntil)
			w.applyMageIlluminationOnSkillHit(entity, target, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
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

func (w *World) addMageEEffect(entity *Entity, center Vector2, radius float64, tick uint64, expiresAt uint64) string {
	id := w.NextEffectID("effect:mage_e:")
	w.PutSkillEffect(SkillEffect{
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
