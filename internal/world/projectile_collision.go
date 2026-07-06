package world

import "math"

func projectileIntersectsTarget(projectile *Projectile, previousPosition Vector2, target *Entity) bool {
	if projectile == nil || target == nil {
		return false
	}
	if projectile.SkillID == archerWSkillID {
		return archerVolleyArrowIntersectsTarget(projectile, previousPosition, target)
	}
	if projectile.SkillID == gunnerRSkillID {
		return gunnerRWaveIntersectsTarget(projectile, previousPosition, target)
	}
	return distancePointToSegment(target.Position, previousPosition, projectile.Position) <= projectile.Radius+target.Radius
}

func gunnerRWaveIntersectsTarget(projectile *Projectile, previousPosition Vector2, target *Entity) bool {
	prevAlong, _ := projectPoint(projectile.Start, projectile.Dir, previousPosition)
	targetAlong, perpendicular := projectPoint(projectile.Start, projectile.Dir, target.Position)
	if targetAlong < prevAlong-target.Radius || targetAlong > projectile.Traveled+target.Radius {
		return false
	}
	coneRadius := math.Tan(projectile.EffectRatio*math.Pi/360)*targetAlong + projectile.Radius
	return perpendicular <= coneRadius+target.Radius
}

func canShieldTarget(source *Entity, target *Entity) bool {
	if source == nil || target == nil || target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
		return false
	}
	return source.Team == target.Team
}

func archerVolleyArrowIntersectsTarget(projectile *Projectile, previousPosition Vector2, target *Entity) bool {
	length := projectileArrowLength(projectile)
	radius := projectileArrowCollisionRadius(projectile) + target.Radius
	segments := [][2]Vector2{
		arrowBodySegment(previousPosition, projectile.Dir, length),
		arrowBodySegment(projectile.Position, projectile.Dir, length),
		{arrowTail(previousPosition, projectile.Dir, length), arrowTail(projectile.Position, projectile.Dir, length)},
		{arrowHead(previousPosition, projectile.Dir, length), arrowHead(projectile.Position, projectile.Dir, length)},
	}
	for _, segment := range segments {
		if distancePointToSegment(target.Position, segment[0], segment[1]) <= radius {
			return true
		}
	}
	return false
}

func arrowBodySegment(center Vector2, direction Vector2, length float64) [2]Vector2 {
	return [2]Vector2{
		arrowTail(center, direction, length),
		arrowHead(center, direction, length),
	}
}

func arrowTail(center Vector2, direction Vector2, length float64) Vector2 {
	return Vector2{
		X: center.X - direction.X*length*0.5,
		Y: center.Y - direction.Y*length*0.5,
	}
}

func arrowHead(center Vector2, direction Vector2, length float64) Vector2 {
	return Vector2{
		X: center.X + direction.X*length*0.5,
		Y: center.Y + direction.Y*length*0.5,
	}
}

func projectileArrowLength(projectile *Projectile) float64 {
	if projectile == nil || projectile.Radius <= 0 {
		return 26
	}
	return projectile.Radius * 1.625
}

func projectileArrowCollisionRadius(projectile *Projectile) float64 {
	if projectile == nil || projectile.Radius <= 0 {
		return 16
	}
	return projectile.Radius
}
