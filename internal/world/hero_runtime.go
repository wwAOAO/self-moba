package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

type HeroCastHandler func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int)

var heroCastHandlers = map[string]map[string]HeroCastHandler{}

func RegisterHeroCastHandlers(heroID string, handlers map[string]HeroCastHandler) {
	if heroCastHandlers[heroID] == nil {
		heroCastHandlers[heroID] = map[string]HeroCastHandler{}
	}
	for skillID, handler := range handlers {
		heroCastHandlers[heroID][skillID] = handler
	}
}

func heroCastHandlerFor(heroID string, skillID string) HeroCastHandler {
	if handlers := heroCastHandlers[heroID]; handlers != nil {
		return handlers[skillID]
	}
	return nil
}
