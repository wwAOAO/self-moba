package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

func init() {
	registerHeroCastHandlers(archerHeroID, map[string]HeroCastHandler{
		archerWSkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applyArcherW(w, entity, cast, state, skill, tick, tickRate)
		},
	})
}
