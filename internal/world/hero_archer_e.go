package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) refreshArcherSkillOnUpgrade(entity *Entity, skillID string) {
	if entity == nil || entity.HeroID != archerHeroID || skillID != archerESkillID {
		return
	}
	state := entity.Skills[archerESkillID]
	if state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(archerESkillID)
	maxCharges := archerHawkMaxCharges(skill)
	if state.Stacks <= 0 {
		state.Stacks = maxCharges
		state.StacksExpireTick = 0
		entity.Skills[archerESkillID] = state
	}
}

func (w *World) tickArcherHawkCharges(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	state, ok := entity.Skills[archerESkillID]
	if !ok || state.Level <= 0 {
		return
	}
	maxCharges := archerHawkMaxCharges(w.SkillConfig(archerESkillID))
	if state.Stacks >= maxCharges {
		state.Stacks = maxCharges
		state.StacksExpireTick = 0
		entity.Skills[archerESkillID] = state
		return
	}
	if state.StacksExpireTick == 0 {
		state.StacksExpireTick = tick + archerHawkRechargeTicks(entity, w.SkillConfig(archerESkillID), state.Level, tickRate)
		entity.Skills[archerESkillID] = state
		return
	}
	for state.Stacks < maxCharges && state.StacksExpireTick > 0 && tick >= state.StacksExpireTick {
		state.Stacks++
		if state.Stacks >= maxCharges {
			state.StacksExpireTick = 0
			break
		}
		state.StacksExpireTick += archerHawkRechargeTicks(entity, w.SkillConfig(archerESkillID), state.Level, tickRate)
	}
	entity.Skills[archerESkillID] = state
}

func archerHawkMaxCharges(skill config.SkillConfig) int {
	maxCharges := int(math.Round(skillMetaRange(skill, "maxCharges", 2)))
	if maxCharges < 1 {
		return 1
	}
	return maxCharges
}

func archerHawkRechargeTicks(entity *Entity, skill config.SkillConfig, level int, tickRate int) uint64 {
	return cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "rechargeMs", level, []float64{90000, 80000, 70000, 60000, 50000}), tickRate)
}
