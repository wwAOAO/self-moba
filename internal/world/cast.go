package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

func (w *World) applyCast(entity *Entity, cast protocol.CastInput, tick uint64, skills *config.SkillStore, tickRate int) {
	if tick < entity.Control.SilencedUntilTick {
		return
	}
	state, ok := entity.Skills[cast.SkillID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.castSkillConfig(cast.SkillID, skills)
	if w.trySpecialRecast(entity, cast, state, skill, tick, tickRate) {
		return
	}
	if tick < state.CooldownUntilTick {
		return
	}
	if handler := heroCastHandlerFor(entity.HeroID, cast.SkillID); handler != nil {
		handler(w, entity, cast, state, skill, tick, tickRate)
		return
	}
	if skill.SkillID == "" {
		return
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skill.CooldownMS, tickRate)
	entity.Skills[cast.SkillID] = state
}

func (w *World) castSkillConfig(skillID string, skills *config.SkillStore) config.SkillConfig {
	if skills != nil {
		if skill, ok := skills.Get(skillID); ok {
			return skill
		}
	}
	return w.skillConfig(skillID)
}

func (w *World) trySpecialRecast(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorESkillID && tick < entity.Warrior.JudgmentUntilTick {
		w.stopWarriorE(entity, state, skill, tick, tickRate)
		return true
	}
	if entity.HeroID == mageHeroID && cast.SkillID == mageESkillID && entity.Mage.LucentSingularityActive {
		w.detonateMageE(entity, skill, tick, tickRate)
		return true
	}
	if h := heroHooksForEntity(entity).SpecialRecast; h != nil {
		return h(w, entity, cast, state, skill, tick, tickRate)
	}
	return false
}
