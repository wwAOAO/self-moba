package world

import "l-battle/internal/config"

func equipmentAbilityPowerMultiplier(entity *Entity, equipment *config.EquipmentStore) float64 {
	if entity == nil || equipment == nil {
		return 0
	}
	multiplier := 0.0
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := equipment.Get(equipped.EquipmentID)
		if ok && item.Effects.AbilityPowerMultiplier > multiplier {
			multiplier = item.Effects.AbilityPowerMultiplier
		}
	}
	return multiplier
}

func equipmentZeroesCritChance(entity *Entity, equipment *config.EquipmentStore) bool {
	if entity == nil || equipment == nil {
		return false
	}
	for _, equipped := range entity.Equipment {
		item, ok := equipment.Get(equipped.EquipmentID)
		if ok && item.Effects.ZeroCritChance {
			return true
		}
	}
	return false
}

func (w *World) rawEquipmentCritChance(entity *Entity) float64 {
	if entity == nil || w.equipment == nil {
		return 0
	}
	chance := entity.Stats.CritChance
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.Effects.ZeroCritChance {
			chance += item.Stats.CritChance
		}
	}
	return clamp(chance, 0, 1)
}
