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
	if !ok {
		return
	}
	if state.Level <= 0 {
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorESkillID && tick < entity.Warrior.JudgmentUntilTick {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.stopWarriorE(entity, state, skill, tick, tickRate)
		return
	}
	if tick < state.CooldownUntilTick {
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordQSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordQ(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordWSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordW(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordESkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordE(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordRSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordR(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorQSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorQ(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorWSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorW(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorESkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorE(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorRSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorR(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankQSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankQ(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankWSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankW(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankESkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankE(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankRSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankR(entity, cast, state, skill, tick, tickRate)
		return
	}
	if skills == nil {
		return
	}
	skill, ok := skills.Get(cast.SkillID)
	if !ok {
		return
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skill.CooldownMS, tickRate)
	entity.Skills[cast.SkillID] = state
}
