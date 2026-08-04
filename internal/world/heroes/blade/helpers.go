package blade

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func beginAction(entity *world.Entity, action string, skillID string, tick uint64, endsAt uint64) {
	if entity == nil {
		return
	}
	entity.Action = world.ActionState{
		Name:          action,
		SkillID:       skillID,
		StartedAtTick: tick,
		EndsAtTick:    endsAt,
	}
}
