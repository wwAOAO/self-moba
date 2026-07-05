package world

import "math"

func (w *World) releaseTankQ(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.SeismicShardPending || tick < entity.Tank.SeismicShardReleaseTick {
		return
	}
	target := w.EntityByID(entity.Tank.SeismicShardTargetID)
	level := entity.Tank.SeismicShardLevel
	entity.Tank.SeismicShardPending = false
	entity.Tank.SeismicShardReleaseTick = 0
	entity.Tank.SeismicShardTargetID = ""
	entity.Tank.SeismicShardLevel = 0
	if !canAttackTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(tankQSkillID)
	state := entity.Skills[tankQSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{8000, 8000, 8000, 8000, 8000}), tickRate)
	entity.Skills[tankQSkillID] = state
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
	id := w.NextProjectileID("projectile:tank_q:")
	w.PutProjectile(&Projectile{
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
	})
	w.LockAttackAfterCast(entity, tick, tickRate)
}
