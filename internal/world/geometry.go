package world

import (
	"math"
)

func canAttackTarget(attacker *Entity, target *Entity) bool {
	if attacker == nil || target == nil {
		return false
	}
	if target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
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
	if attackSpeed <= 0 {
		return uint64(tickRate)
	}
	ticks := math.Ceil(float64(tickRate) / attackSpeed)
	if ticks < 1 {
		return 1
	}
	return uint64(ticks)
}

func distance(a Vector2, b Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func projectPoint(origin Vector2, direction Vector2, point Vector2) (float64, float64) {
	dx := point.X - origin.X
	dy := point.Y - origin.Y
	along := dx*direction.X + dy*direction.Y
	perpX := dx - along*direction.X
	perpY := dy - along*direction.Y
	return along, math.Hypot(perpX, perpY)
}

func windWallStart(wall WindWall) Vector2 {
	half := wall.Width / 2
	return Vector2{
		X: wall.Center.X - wall.Dir.X*half,
		Y: wall.Center.Y - wall.Dir.Y*half,
	}
}

func windWallEnd(wall WindWall) Vector2 {
	half := wall.Width / 2
	return Vector2{
		X: wall.Center.X + wall.Dir.X*half,
		Y: wall.Center.Y + wall.Dir.Y*half,
	}
}

func segmentsIntersect(a Vector2, b Vector2, c Vector2, d Vector2) bool {
	ab1 := orientation(a, b, c)
	ab2 := orientation(a, b, d)
	cd1 := orientation(c, d, a)
	cd2 := orientation(c, d, b)
	return ab1*ab2 <= 0 && cd1*cd2 <= 0
}

func orientation(a Vector2, b Vector2, c Vector2) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}
