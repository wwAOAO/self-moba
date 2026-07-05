package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

func init() {
	registerHeroCastHandlers(mageHeroID, map[string]HeroCastHandler{
		mageESkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applyMageE(w, entity, cast, state, skill, tick, tickRate)
		},
	})
}
