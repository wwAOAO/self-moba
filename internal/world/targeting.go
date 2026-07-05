package world

import (
	"l-battle/internal/config"
	"l-battle/internal/world/model"
	"math"
)

func (w *World) swordQTargets(entity *Entity, targetPoint Vector2, qRange float64, form string, skill config.SkillConfig) []*Entity {
	if form == "circle" {
		return w.targetsInRadius(entity, entity.Position, qRange)
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	if form == "whirlwind" {
		return w.targetsAlongMovingCircle(entity, entity.Position, Vector2{X: dx, Y: dy}, qRange, skillMetaRange(skill, "whirlwindRadius", 70))
	}
	width := skillMetaRange(skill, "lineWidth", 55)
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(entity.Position, Vector2{X: dx, Y: dy}, target.Position)
		if along < 0 || along > qRange+target.Radius {
			continue
		}
		if perpendicular > width+target.Radius {
			continue
		}
		hits = append(hits, target)
	}
	return hits
}

func (w *World) targetsAlongMovingCircle(entity *Entity, origin Vector2, direction Vector2, travelRange float64, radius float64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(origin, direction, target.Position)
		if along < -target.Radius || along > travelRange+target.Radius {
			continue
		}
		if perpendicular <= radius+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func (w *World) targetsInRadius(entity *Entity, center Vector2, radius float64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(center, target.Position) <= radius+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func (w *World) targetsInCone(entity *Entity, direction Vector2, coneRange float64, angleDegrees float64) []*Entity {
	hits := make([]*Entity, 0)
	if direction.X == 0 && direction.Y == 0 {
		direction = Vector2{X: 1, Y: 0}
	}
	cosLimit := math.Cos((angleDegrees / 2) * math.Pi / 180)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		toTarget := Vector2{X: target.Position.X - entity.Position.X, Y: target.Position.Y - entity.Position.Y}
		dist := math.Hypot(toTarget.X, toTarget.Y)
		if dist > coneRange+target.Radius || dist == 0 {
			continue
		}
		dot := (toTarget.X*direction.X + toTarget.Y*direction.Y) / dist
		if dot >= cosLimit {
			hits = append(hits, target)
		}
	}
	return hits
}

func isMonster(entity *Entity) bool {
	if entity == nil {
		return false
	}
	return model.IsMonsterKind(entity.Kind, entity.Team)
}

func isMinion(entity *Entity) bool {
	if entity == nil {
		return false
	}
	return model.IsMinionKind(entity.Kind)
}
