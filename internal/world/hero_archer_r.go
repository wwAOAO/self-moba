package world

import (
	"l-battle/internal/config"
)

func (w *World) releaseArcherCrystalArrow(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID || !entity.Archer.CrystalArrowPending || tick < entity.Archer.CrystalArrowReleaseTick {
		return
	}
	skill := w.SkillConfig(archerRSkillID)
	target := entity.Archer.CrystalArrowTarget
	level := entity.Archer.CrystalArrowLevel
	state := entity.Skills[archerRSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{100000, 90000, 80000}), tickRate)
	entity.Skills[archerRSkillID] = state
	entity.Archer.CrystalArrowPending = false
	entity.Archer.CrystalArrowReleaseTick = 0
	entity.Archer.CrystalArrowTarget = Vector2{}
	entity.Archer.CrystalArrowLevel = 0

	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	minSpeed := skillMetaRange(skill, "projectileMinSpeed", 1500)
	maxSpeed := skillMetaRange(skill, "projectileMaxSpeed", 2100)
	speedPerTick := minSpeed
	if tickRate > 0 {
		speedPerTick /= float64(tickRate)
	}
	id := w.NextProjectileID("projectile:archer_r:")
	w.PutProjectile(&Projectile{
		ID:           id,
		Kind:         "archer_crystal_arrow",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      archerRSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		SpeedMin:     minSpeed,
		SpeedMax:     maxSpeed,
		Range:        skillRange(skill, DefaultMapWidth),
		Radius:       skillMetaRange(skill, "projectileRadius", 28),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillRange(skill, DefaultMapWidth)/minSpeed+0.5, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func archerRDamage(entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64, multiplier float64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{200, 400, 600})
	rawDamage := (baseDamage + float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 1)) * multiplier
	return magicDamageAfterResistance(entity, target, rawDamage, tick)
}

func archerRStunTicks(projectile *Projectile, skill config.SkillConfig, tickRate int) uint64 {
	if projectile == nil {
		return secondsToTicks(skillMetaRange(skill, "minStunSeconds", 1), tickRate)
	}
	minSeconds := skillMetaRange(skill, "minStunSeconds", 1)
	maxSeconds := skillMetaRange(skill, "maxStunSeconds", 3.5)
	maxDistance := skillMetaRange(skill, "maxStunDistance", 1400)
	progress := 1.0
	if maxDistance > 0 {
		progress = clamp(projectile.Traveled/maxDistance, 0, 1)
	}
	return secondsToTicks(minSeconds+(maxSeconds-minSeconds)*progress, tickRate)
}

func applyArcherRSplash(w *World, source *Entity, primary *Entity, projectile *Projectile, skill config.SkillConfig, tick uint64, tickRate int) {
	if source == nil || primary == nil || projectile == nil {
		return
	}
	radius := skillMetaRange(skill, "splashRadius", 260)
	multiplier := skillMetaRange(skill, "splashDamageMultiplier", 0.5)
	w.ForEachEntity(func(target *Entity) {
		if target == nil || target.ID == primary.ID || !canAttackTarget(source, target) {
			return
		}
		if distance(target.Position, primary.Position) > radius+target.Radius {
			return
		}
		damage := archerRDamage(source, target, skill, projectile.Damage, tick, multiplier)
		target.Combat.LastHitTick = tick
		wasAlive := target.Stats.HP > 0
		w.ApplyAOEDamage(source, target, damage, "magic", tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	})
}
