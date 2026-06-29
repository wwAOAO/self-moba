package system

import (
	"math"

	"l-battle/internal/world"
)

func ApplyMovement(entity *world.Entity, direction world.Vector2, speed float64) {
	length := math.Hypot(direction.X, direction.Y)
	if length == 0 {
		return
	}
	entity.Position.X += direction.X / length * speed
	entity.Position.Y += direction.Y / length * speed
}
