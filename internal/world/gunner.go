package world

import "math"

func (w *World) fireGunnerQBounce(source *Entity, first *Entity, projectile *Projectile, forceCrit bool, tick uint64, tickRate int) {
	if source == nil || first == nil || projectile == nil || tickRate <= 0 {
		return
	}
	target := w.gunnerQSecondTarget(source, first)
	if target == nil {
		return
	}
	dx, dy := normalize(target.Position.X-first.Position.X, target.Position.Y-first.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	id := w.NextProjectileID("projectile:gunner_q:")
	w.PutProjectile(&Projectile{
		ID:           id,
		Kind:         "gunner_q",
		Team:         source.Team,
		SourceID:     source.ID,
		TargetID:     target.ID,
		SkillID:      gunnerQSkillID,
		Position:     first.Position,
		Start:        first.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: projectile.SpeedPerTick,
		Range:        skillMetaRange(w.skillConfig(gunnerQSkillID), "bounceRange", 500) + target.Radius,
		Radius:       projectile.Radius,
		Damage:       projectile.Damage,
		EffectRatio:  boolRatio(forceCrit),
		Returning:    true,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func (w *World) gunnerQSecondTarget(source *Entity, first *Entity) *Entity {
	skill := w.skillConfig(gunnerQSkillID)
	dx, dy := normalize(first.Position.X-source.Position.X, first.Position.Y-source.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	coneRange := skillMetaRange(skill, "bounceRange", 500)
	cosLimit := math.Cos(skillMetaRange(skill, "bounceAngleDegrees", 60) * math.Pi / 360)
	var best *Entity
	bestAlong := math.MaxFloat64
	w.ForEachEntity(func(target *Entity) {
		if target == first || !canAttackTarget(source, target) {
			return
		}
		tx := target.Position.X - first.Position.X
		ty := target.Position.Y - first.Position.Y
		dist := math.Hypot(tx, ty)
		if dist == 0 || dist > coneRange+target.Radius {
			return
		}
		along := tx*dx + ty*dy
		if along <= 0 || along/dist < cosLimit {
			return
		}
		if along < bestAlong {
			bestAlong = along
			best = target
		}
	})
	return best
}

func boolRatio(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
