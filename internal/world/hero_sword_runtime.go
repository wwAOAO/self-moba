package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

func init() {
	registerHeroCastHandlers(swordHeroID, map[string]HeroCastHandler{
		swordQSkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applySwordQ(w, entity, cast, state, skill, tick, tickRate)
		},
		swordWSkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applySwordW(w, entity, cast, state, skill, tick, tickRate)
		},
		swordESkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applySwordE(w, entity, cast, state, skill, tick, tickRate)
		},
		swordRSkillID: func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			applySwordR(w, entity, cast, state, skill, tick, tickRate)
		},
	})
}
