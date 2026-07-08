package world

import "math"

func (w *World) recalculatePlayerStats(entity *Entity) {
	if entity == nil || entity.Kind != EntityKindPlayer {
		return
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return
	}
	oldMaxHP := entity.Stats.MaxHP
	oldMaxMP := entity.Stats.MaxMP
	oldHP := entity.Stats.HP
	oldMP := entity.Stats.MP
	nextStats := heroStatsAtLevel(hero, entity.Level)
	w.applyEquipmentStats(entity, &nextStats)
	w.applySwordCritOverflowStats(entity, &nextStats)
	nextStats.AbilityHaste += abilityHasteFromBuffs(entity)
	nextStats.HP = oldHP + (nextStats.MaxHP - oldMaxHP)
	if nextStats.HP > nextStats.MaxHP {
		nextStats.HP = nextStats.MaxHP
	}
	if nextStats.HP < 1 {
		nextStats.HP = 1
	}
	nextStats.MP = oldMP + (nextStats.MaxMP - oldMaxMP)
	if nextStats.MP > nextStats.MaxMP {
		nextStats.MP = nextStats.MaxMP
	}
	if nextStats.MP < 0 {
		nextStats.MP = 0
	}
	w.applyHeroStats(entity, &nextStats)
	applyControlStats(entity, &nextStats)
	entity.Tank.ThunderclapArmorBonus = 0
	entity.Stats = nextStats
	w.refreshTankGraniteShieldMax(entity)
	w.refreshTankWPassive(entity)
}

func applyControlStats(entity *Entity, stats *Stats) {
	if entity == nil || stats == nil {
		return
	}
	if entity.Control.AttackDamageReduceUntil > 0 && entity.Control.AttackDamageReduction > 0 {
		stats.Attack -= entity.Control.AttackDamageReduction
		if stats.Attack < 0 {
			stats.Attack = 0
		}
		stats.BonusAttack -= entity.Control.AttackDamageReduction
		if stats.BonusAttack < 0 {
			stats.BonusAttack = 0
		}
	}
	if entity.Control.GrievousWoundsUntil > 0 && entity.Control.GrievousWounds > stats.GrievousWounds {
		stats.GrievousWounds = entity.Control.GrievousWounds
	}
}

func (w *World) refreshPlayerStatsAfterHPChange(entity *Entity, beforeHP float64) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 || entity.Stats.HP == beforeHP {
		return
	}
	if heroHooksForEntity(entity).ApplyStats == nil {
		return
	}
	w.recalculatePlayerStats(entity)
}

func (w *World) applyEquipmentStats(entity *Entity, stats *Stats) {
	if entity == nil || stats == nil || w.equipment == nil {
		return
	}
	baseHPRegen5 := stats.HPRegen5
	baseMPRegen5 := stats.MPRegen5
	baseHPRegenBonus := 0.0
	baseMPRegenBonus := 0.0
	equipmentHP := 0.0
	equipmentHPMultiplier := 0.0
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		stats.MaxHP += item.Stats.HP
		stats.BonusHP += item.Stats.HP
		equipmentHP += item.Stats.HP
		stats.MaxMP += item.Stats.MP
		stats.HPRegen5 += item.Stats.HPRegen5
		stats.MPRegen5 += item.Stats.MPRegen5
		baseHPRegenBonus += item.Stats.BaseHPRegenBonus
		baseMPRegenBonus += item.Stats.BaseMPRegenBonus
		if item.Effects.EquipmentHPMultiplier > equipmentHPMultiplier {
			equipmentHPMultiplier = item.Effects.EquipmentHPMultiplier
		}
		stats.Attack += item.Stats.Attack
		stats.BonusAttack += item.Stats.Attack
		stats.AbilityPower += item.Stats.AbilityPower
		stats.AbilityHaste += item.Stats.AbilityHaste
		stats.PhysicalDefense += item.Stats.PhysicalDefense
		stats.BonusPhysicalDefense += item.Stats.PhysicalDefense
		stats.MagicDefense += item.Stats.MagicDefense
		stats.BonusMagicDefense += item.Stats.MagicDefense
		stats.PhysicalPenPercent += item.Stats.PhysicalPenPercent
		stats.MagicPenPercent += item.Stats.MagicPenPercent
		stats.MagicPenFlat += item.Stats.MagicPenFlat
		stats.Tenacity += item.Stats.Tenacity
		stats.SlowResist += item.Stats.SlowResist
		stats.BasicAttackBlock += item.Stats.BasicAttackBlock
		stats.CritDamageReduce += item.Stats.CritDamageReduce
		stats.PhysicalDefense += equipped.Stacks * item.Effects.UnitKillPhysicalDefenseGain
		stats.BonusPhysicalDefense += equipped.Stacks * item.Effects.UnitKillPhysicalDefenseGain
		stats.AbilityPower += int(equipped.Stacks * item.Effects.UnitKillAbilityPowerGain)
		stats.MagicDefense += equipped.Stacks * item.Effects.MagicHitMagicDefensePerStack
		stats.BonusMagicDefense += equipped.Stacks * item.Effects.MagicHitMagicDefensePerStack
		stats.MoveSpeed += item.Stats.MoveSpeed
		stats.MoveSpeed *= 1 + equipped.Stacks*item.Effects.MagicHitMoveSpeedPercentPerStack
		stats.MoveSpeed *= 1 + item.Stats.MoveSpeedPercent
		stats.AttackSpeedBonus += item.Stats.AttackSpeedBonus
		stats.AttackSpeedBonus += equipped.Stacks * item.Effects.BasicAttackAttackSpeedPerStack
		stats.CritChance += item.Stats.CritChance
		stats.Omnivamp += item.Stats.Omnivamp
		stats.LifeSteal += item.Stats.LifeSteal
	}
	if equipmentZeroesCritChance(entity, w.equipment) {
		stats.CritChance = 0
	}
	extraEquipmentHP := equipmentHP * equipmentHPMultiplier
	stats.MaxHP += extraEquipmentHP
	stats.BonusHP += extraEquipmentHP
	stats.HPRegen5 += baseHPRegen5 * baseHPRegenBonus
	stats.MPRegen5 += baseMPRegen5 * baseMPRegenBonus
	stats.AbilityPower = int(math.Round(float64(stats.AbilityPower) * (1 + equipmentAbilityPowerMultiplier(entity, w.equipment))))
	w.applyStoneplateResists(entity, stats)
	stats.AttackSpeed = finalAttackSpeed(stats.BaseAttackSpeed, stats.AttackSpeedBonus, stats.AttackSpeedRatio, stats.AttackSpeedSlow)
}

func (w *World) applySwordCritOverflowStats(entity *Entity, stats *Stats) {
	if heroHooksFor(swordHeroID).ApplyCritOverflowStats != nil {
		heroHooksFor(swordHeroID).ApplyCritOverflowStats(w, entity, stats)
	}
}

func (w *World) hasUniqueEquipmentGroup(entity *Entity, group string) bool {
	if entity == nil || group == "" || w.equipment == nil {
		return false
	}
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.UniqueGroup == group {
			return true
		}
	}
	return false
}

func (w *World) hasEquipmentCategory(entity *Entity, category string) bool {
	if entity == nil || category == "" || w.equipment == nil {
		return false
	}
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.Category == category {
			return true
		}
	}
	return false
}

func (w *World) hasUniqueEquipmentGroupOutsideIndexes(entity *Entity, group string, ignored []int) bool {
	if entity == nil || group == "" || w.equipment == nil {
		return false
	}
	ignore := make(map[int]bool, len(ignored))
	for _, index := range ignored {
		ignore[index] = true
	}
	for index, equipped := range entity.Equipment {
		if ignore[index] {
			continue
		}
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.UniqueGroup == group {
			return true
		}
	}
	return false
}

func (w *World) hasEquipmentCategoryOutsideIndexes(entity *Entity, category string, ignored []int) bool {
	if entity == nil || category == "" || w.equipment == nil {
		return false
	}
	ignore := make(map[int]bool, len(ignored))
	for _, index := range ignored {
		ignore[index] = true
	}
	for index, equipped := range entity.Equipment {
		if ignore[index] {
			continue
		}
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.Category == category {
			return true
		}
	}
	return false
}

func (w *World) setMessage(entity *Entity, message string, tick uint64) {
	if entity == nil {
		return
	}
	entity.Message = message
	entity.MessageTick = tick
}

func equipmentOutOfCombatMoveSpeed(entity *Entity, tick uint64) float64 {
	if entity == nil || tick == 0 {
		return 0
	}
	var bonus float64
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		if equipped.OutOfCombatMoveSpeed <= 0 || equipped.OutOfCombatRequiredTicks == 0 {
			continue
		}
		if tick >= entity.Combat.LastHitTick+equipped.OutOfCombatRequiredTicks {
			bonus += equipped.OutOfCombatMoveSpeed
		}
	}
	return bonus
}
