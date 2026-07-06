package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Skills[rID] = state
	delayTicks := secondsToTicks(skillMeta(skill, "castDelaySeconds", 0.25), tickRate)
	entity.Control.ActionLockedUntilTick = tick + delayTicks
	entity.Archer.CrystalArrowPending = true
	entity.Archer.CrystalArrowReleaseTick = tick + delayTicks
	entity.Archer.CrystalArrowTarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Archer.CrystalArrowLevel = state.Level
}

func ReleaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Archer.CrystalArrowPending || tick < entity.Archer.CrystalArrowReleaseTick {
		return
	}
	skill := w.SkillConfig(rID)
	target := entity.Archer.CrystalArrowTarget
	level := entity.Archer.CrystalArrowLevel
	state := entity.Skills[rID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{100000, 90000, 80000})), tickRate)
	entity.Skills[rID] = state
	entity.Archer.CrystalArrowPending = false
	entity.Archer.CrystalArrowReleaseTick = 0
	entity.Archer.CrystalArrowTarget = world.Vector2{}
	entity.Archer.CrystalArrowLevel = 0

	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	minSpeed := skillMeta(skill, "projectileMinSpeed", 1500)
	maxSpeed := skillMeta(skill, "projectileMaxSpeed", 2100)
	speedPerTick := minSpeed
	if tickRate > 0 {
		speedPerTick /= float64(tickRate)
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:archer_r:"),
		Kind:         "archer_crystal_arrow",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      rID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		SpeedMin:     minSpeed,
		SpeedMax:     maxSpeed,
		Range:        skillRange(skill, world.DefaultMapWidth),
		Radius:       skillMeta(skill, "projectileRadius", 28),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillRange(skill, world.DefaultMapWidth)/minSpeed+0.5, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func RDamage(w *world.World, entity *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64, multiplier float64) int {
	baseDamage := skillList(skill, "baseDamage", level, []float64{200, 400, 600})
	rawDamage := (baseDamage + float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 1)) * multiplier
	return w.ArcherMagicDamageAfterResistance(entity, target, rawDamage, tick)
}

func RStunTicks(projectile *world.Projectile, skill config.SkillConfig, tickRate int) uint64 {
	if projectile == nil {
		return secondsToTicks(skillMeta(skill, "minStunSeconds", 1), tickRate)
	}
	minSeconds := skillMeta(skill, "minStunSeconds", 1)
	maxSeconds := skillMeta(skill, "maxStunSeconds", 3.5)
	maxDistance := skillMeta(skill, "maxStunDistance", 1400)
	progress := 1.0
	if maxDistance > 0 {
		progress = clamp(projectile.Traveled/maxDistance, 0, 1)
	}
	return secondsToTicks(minSeconds+(maxSeconds-minSeconds)*progress, tickRate)
}

func ApplyRSplash(w *world.World, source *world.Entity, primary *world.Entity, projectile *world.Projectile, skill config.SkillConfig, tick uint64, tickRate int) {
	if source == nil || primary == nil || projectile == nil {
		return
	}
	radius := skillMeta(skill, "splashRadius", 260)
	multiplier := skillMeta(skill, "splashDamageMultiplier", 0.5)
	w.ForEachEntity(func(target *world.Entity) {
		if target == nil || target.ID == primary.ID || !canAttackTarget(source, target) {
			return
		}
		if distance(target.Position, primary.Position) > radius+target.Radius {
			return
		}
		damage := RDamage(w, source, target, skill, projectile.Damage, tick, multiplier)
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
