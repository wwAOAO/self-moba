package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ReleaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Mage.LightBindingPending || tick < entity.Mage.LightBindingReleaseTick {
		return
	}
	targetPoint := entity.Mage.LightBindingTarget
	level := entity.Mage.LightBindingLevel
	entity.Mage.LightBindingPending = false
	entity.Mage.LightBindingReleaseTick = 0
	entity.Mage.LightBindingTarget = world.Vector2{}
	entity.Mage.LightBindingLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	state := entity.Skills[qID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{15000, 14000, 13000, 12000, 11000})), tickRate)
	entity.Skills[qID] = state
	qRange := skillRange(skill, 1175)
	speedPerSecond := skillMeta(skill, "projectileSpeed", 1400)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	id := w.NextProjectileID("projectile:mage_q:")
	w.PutProjectile(&world.Projectile{
		ID:           id,
		Kind:         "mage_light_binding",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 45),
		Damage:       level,
		EffectTicks:  secondsToTicks(skillMeta(skill, "rootSeconds", 2), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	})
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int {
	baseDamage := skillList(skill, "baseDamage", skillLevel, []float64{50, 100, 150, 200, 250})
	rawDamage := (baseDamage + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.7)) * multiplier
	return w.MageMagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func ApplyQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.LightBindingPending {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.LightBindingPending = true
	entity.Mage.LightBindingReleaseTick = tick + windupTicks
	entity.Mage.LightBindingTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Mage.LightBindingLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.LightBindingReleaseTick
	entity.Skills[qID] = state
}
