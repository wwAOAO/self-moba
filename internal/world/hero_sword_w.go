package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

func applySwordW(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	id := w.NextWindWallID("wind_wall:")
	width := skillMetaListByLevel(skill, "width", state.Level, []float64{300, 350, 400, 450, 500})
	placeDistance := skillMetaRange(skill, "placeDistance", 180)
	center := w.ClampWorldPoint(Vector2{X: entity.Position.X + dx*placeDistance, Y: entity.Position.Y + dy*placeDistance})
	w.PutWindWall(WindWall{
		ID:        id,
		Team:      entity.Team,
		Center:    center,
		Dir:       Vector2{X: -dy, Y: dx},
		Width:     width,
		ExpiresAt: tick + secondsToTicks(skillMetaRange(skill, "durationSeconds", windWallDuration), tickRate),
	})
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{26000, 24000, 22000, 20000, 18000}), tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordWSkillID] = state
}
