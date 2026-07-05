package world

import "l-battle/internal/config"

func applyTankW(w *World, entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{30, 35, 40, 45, 50})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Tank.ThunderclapEmpowerUntil = tick + secondsToTicks(skillMetaRange(skill, "aftershockDurationSeconds", 5), tickRate)
	entity.Tank.ThunderclapAftershockUntil = entity.Tank.ThunderclapEmpowerUntil
	entity.Tank.ThunderclapLevel = state.Level
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{10000, 9500, 9000, 8500, 8000}), tickRate)
	entity.Skills[tankWSkillID] = state
}

func (w *World) refreshTankWPassive(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	if entity.Tank.ThunderclapArmorBonus != 0 {
		entity.Stats.PhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Stats.BonusPhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Tank.ThunderclapArmorBonus = 0
	}
	state, ok := entity.Skills[tankWSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(tankWSkillID)
	ratio := skillMetaListByLevel(skill, "passiveArmorRatio", state.Level, []float64{0.1, 0.15, 0.2, 0.25, 0.3})
	if entity.Passive.Shield > 0 {
		ratio = skillMetaRange(skill, "shieldArmorRatio", 0.3)
	}
	baseArmor := entity.Stats.PhysicalDefense - entity.Stats.BonusPhysicalDefense
	bonus := baseArmor * ratio
	entity.Tank.ThunderclapArmorBonus = bonus
	entity.Stats.PhysicalDefense += bonus
	entity.Stats.BonusPhysicalDefense += bonus
}
