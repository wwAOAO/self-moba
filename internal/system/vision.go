package system

import "l-battle/internal/world"

func VisibleEntities(observer world.Entity, entities []world.Entity, radius float64) []world.Entity {
	visible := make([]world.Entity, 0, len(entities))
	radius2 := radius * radius
	for _, entity := range entities {
		dx := observer.Position.X - entity.Position.X
		dy := observer.Position.Y - entity.Position.Y
		if dx*dx+dy*dy <= radius2 {
			visible = append(visible, entity)
		}
	}
	return visible
}
