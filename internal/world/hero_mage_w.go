package world

import (
	"l-battle/internal/config"
	"math"
)

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
	skill := w.SkillConfig(mageWSkillID)
	state := entity.Skills[mageWSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{14000, 13000, 12000, 11000, 10000}), tickRate)
	entity.Skills[mageWSkillID] = state
	w.addMageShieldLayer(entity, mageWShieldValue(entity, skill, level), tick+secondsToTicks(skillMetaRange(skill, "shieldSeconds", 3), tickRate))
	w.spawnMageWProjectile(entity, Vector2{X: dx, Y: dy}, level, skill, tick, tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
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
	id := w.NextProjectileID("projectile:mage_w:")
	w.PutProjectile(&Projectile{
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
	})
}

func mageWShieldValue(entity *Entity, skill config.SkillConfig, level int) int {
	baseShield := skillMetaListByLevel(skill, "shieldValue", level, []float64{50, 65, 80, 95, 110})
	return int(math.Round(baseShield + float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.2)))
}

func (w *World) addMageShieldLayer(target *Entity, amount int, expiresAt uint64) {
	if target == nil || amount <= 0 || expiresAt == 0 {
		return
	}
	target.Passive.ShieldLayers = append(target.Passive.ShieldLayers, ShieldLayer{Amount: amount, ExpiresAt: expiresAt})
	target.Passive.Shield += amount
	target.Passive.MaxShield += amount
}
