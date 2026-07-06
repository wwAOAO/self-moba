package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func ReleaseW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Mage.PrismaticBarrierPending || tick < entity.Mage.PrismaticBarrierReleaseTick {
		return
	}
	targetPoint := entity.Mage.PrismaticBarrierTarget
	level := entity.Mage.PrismaticBarrierLevel
	entity.Mage.PrismaticBarrierPending = false
	entity.Mage.PrismaticBarrierReleaseTick = 0
	entity.Mage.PrismaticBarrierTarget = world.Vector2{}
	entity.Mage.PrismaticBarrierLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(wID)
	state := entity.Skills[wID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{14000, 13000, 12000, 11000, 10000})), tickRate)
	entity.Skills[wID] = state
	w.AddMageShieldLayer(entity, w.MageWShieldValue(entity, skill, level), tick+secondsToTicks(skillMeta(skill, "shieldSeconds", 3), tickRate))
	spawnWProjectile(w, entity, world.Vector2{X: dx, Y: dy}, level, skill, tick, tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func spawnWProjectile(w *world.World, entity *world.Entity, dir world.Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	wRange := skillRange(skill, 1075)
	speedPerSecond := skillMeta(skill, "projectileSpeed", 1450)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = wRange
	}
	lifeTicks := secondsToTicks(10, tickRate)
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	id := w.NextProjectileID("projectile:mage_w:")
	w.PutProjectile(&world.Projectile{
		ID:           id,
		Kind:         "mage_prismatic_barrier",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      wID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          dir,
		SpeedPerTick: speedPerTick,
		SpeedMin:     speedPerTick,
		Range:        wRange * 2,
		Radius:       skillMeta(skill, "projectileRadius", 55),
		Damage:       level,
		EffectTicks:  secondsToTicks(skillMeta(skill, "shieldSeconds", 3), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	})
}

func ApplyW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.PrismaticBarrierPending {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 60)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.2), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.PrismaticBarrierPending = true
	entity.Mage.PrismaticBarrierReleaseTick = tick + windupTicks
	entity.Mage.PrismaticBarrierTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Mage.PrismaticBarrierLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.PrismaticBarrierReleaseTick
	entity.Skills[wID] = state
}
