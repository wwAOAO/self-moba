package world

import (
	"l-battle/internal/world/formula"
	"l-battle/internal/world/geom"
	"math"
)

func canAttackTarget(attacker *Entity, target *Entity) bool {
	if attacker == nil || target == nil {
		return false
	}
	if target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindFountain {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
		return false
	}
	if target.Control.UntargetableUntilTick > 0 {
		return false
	}
	if target.ID == attacker.ID || target.Team == attacker.Team {
		return false
	}
	return true
}

func attackReach(attacker *Entity, target *Entity) float64 {
	return attacker.Stats.AttackRange + attacker.Radius + target.Radius
}

func (w *World) attackReachAtTick(attacker *Entity, target *Entity, tick uint64) float64 {
	attackRange := attacker.Stats.AttackRange
	if attacker.HeroID == warriorHeroID && tick < attacker.Warrior.DecisiveStrikeUntilTick {
		attackRange = math.Max(attackRange, skillRange(w.skillConfig(warriorQSkillID), 300))
	}
	return attackRange + attacker.Radius + target.Radius
}

func attackCooldownTicks(attackSpeed float64, tickRate int) uint64 {
	return formula.AttackCooldownTicks(attackSpeed, tickRate)
}

func distance(a Vector2, b Vector2) float64 {
	return geom.Distance(a, b)
}

func distancePointToSegment(point Vector2, start Vector2, end Vector2) float64 {
	return geom.DistancePointToSegment(point, start, end)
}

func closestPointOnSegment(point Vector2, start Vector2, end Vector2) Vector2 {
	return geom.ClosestPointOnSegment(point, start, end)
}

func projectPoint(origin Vector2, direction Vector2, point Vector2) (float64, float64) {
	return geom.ProjectPoint(origin, direction, point)
}

func windWallStart(wall WindWall) Vector2 {
	start, _ := geom.SegmentEndpoints(wall.Center, wall.Dir, wall.Width)
	return start
}

func windWallEnd(wall WindWall) Vector2 {
	_, end := geom.SegmentEndpoints(wall.Center, wall.Dir, wall.Width)
	return end
}

func segmentsIntersect(a Vector2, b Vector2, c Vector2, d Vector2) bool {
	return geom.SegmentsIntersect(a, b, c, d)
}

func orientation(a Vector2, b Vector2, c Vector2) float64 {
	return geom.Orientation(a, b, c)
}
