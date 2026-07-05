package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

func init() {
	registerHeroCastHandlers(tankHeroID, map[string]HeroCastHandler{
		tankWSkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applyTankW(w, entity, state, skill, tick, tickRate)
		},
		tankESkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applyTankE(w, entity, state, skill, tick, tickRate)
		},
	})
}
