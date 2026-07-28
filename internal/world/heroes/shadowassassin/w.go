package shadowassassin

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{60, 65, 70, 75, 80})
	if entity.Stats.MP < cost {
		return
	}
	dx, dy := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}

	startRRecall(w, entity)
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, 10000, tickRate)
	entity.Skills[wID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)

	count := int(skillMeta(skill, "bladeCount", 7))
	if count < 1 {
		count = 1
	}
	angle := skillMeta(skill, "coneAngleDegrees", 50)
	rangeValue := skill.Range
	if rangeValue <= 0 {
		rangeValue = 850
	}
	groupID := w.NextEffectID("effect:shadow_assassin_w_group:")
	hitIDs := make(map[string]bool)
	for index := 0; index < count; index++ {
		offset := 0.0
		if count > 1 {
			offset = -angle/2 + angle*float64(index)/float64(count-1)
		}
		bladeX, bladeY := rotate(dx, dy, offset*math.Pi/180)
		w.PutProjectile(&world.Projectile{
			ID:           w.NextProjectileID("projectile:shadow_assassin_w:"),
			Kind:         "shadow_assassin_w",
			Team:         entity.Team,
			SourceID:     entity.ID,
			SkillID:      wID,
			GroupID:      groupID,
			Position:     entity.Position,
			Start:        entity.Position,
			Dir:          world.Vector2{X: bladeX, Y: bladeY},
			SpeedPerTick: skillMeta(skill, "projectileSpeed", 1700) / float64(tickRate),
			Range:        rangeValue * 2,
			Radius:       skillMeta(skill, "projectileRadius", 55),
			Damage:       state.Level,
			Boomerang:    true,
			CreatedAt:    tick,
			ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "projectileLifetimeSeconds", 2.5), tickRate),
			HitIDs:       hitIDs,
			DisplayCount: index + 1,
		})
	}
}

func ResolveWProjectile(w *world.World, source *world.Entity, projectile *world.Projectile, previousPosition world.Vector2, tick uint64, tickRate int) bool {
	if projectile == nil || projectile.SkillID != wID {
		return false
	}
	if w == nil || source == nil || source.HeroID != heroID || tickRate <= 0 {
		return true
	}
	returning := projectile.Returning && tick > projectile.EffectTicks
	phase := "out:"
	if returning {
		phase = "return:"
	}
	w.ForEachEntity(func(target *world.Entity) {
		if !world.CanAttackTarget(source, target) || !segmentHitsCircle(previousPosition, projectile.Position, target.Position, projectile.Radius+target.Radius) {
			return
		}
		key := phase + target.ID
		if projectile.HitIDs[key] {
			return
		}
		projectile.HitIDs[key] = true
		applyWHit(w, source, target, projectile.Damage, tick, tickRate)
	})
	return true
}

func applyWHit(w *world.World, source *world.Entity, target *world.Entity, level int, tick uint64, tickRate int) {
	skill := w.SkillConfig(wID)
	raw := skillList(skill, "baseDamage", level, []float64{30, 55, 80, 105, 130})
	raw += source.Stats.BonusAttack * skillMeta(skill, "bonusAdRatio", 0.6)
	damage := w.PhysicalDamageAfterResistance(source, target, raw, tick)
	target.Combat.LastHitTick = tick
	wasAlive := target.Stats.HP > 0
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
	} else {
		w.ApplyAOEDamage(source, target, damage, "physical", tickRate)
	}
	slow := skillList(skill, "slow", level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
	w.ApplyMoveSpeedSlow(target, slow, tick+secondsToTicks(skillMeta(skill, "slowSeconds", 2), tickRate))
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(source, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}

func rotate(x float64, y float64, radians float64) (float64, float64) {
	cosine := math.Cos(radians)
	sine := math.Sin(radians)
	return x*cosine - y*sine, x*sine + y*cosine
}

func normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func segmentHitsCircle(start world.Vector2, end world.Vector2, center world.Vector2, radius float64) bool {
	dx := end.X - start.X
	dy := end.Y - start.Y
	lengthSquared := dx*dx + dy*dy
	if lengthSquared == 0 {
		return math.Hypot(center.X-start.X, center.Y-start.Y) <= radius
	}
	t := ((center.X-start.X)*dx + (center.Y-start.Y)*dy) / lengthSquared
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closestX := start.X + dx*t
	closestY := start.Y + dy*t
	return math.Hypot(center.X-closestX, center.Y-closestY) <= radius
}
