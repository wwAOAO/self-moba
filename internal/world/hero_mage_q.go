package world

import (
	"l-battle/internal/config"
	"math"
)

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
	skill := w.SkillConfig(mageQSkillID)
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
	id := w.NextProjectileID("projectile:mage_q:")
	w.PutProjectile(&Projectile{
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
	})
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func mageQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{50, 100, 150, 200, 250})
	rawDamage := (baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.7)) * multiplier
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}
