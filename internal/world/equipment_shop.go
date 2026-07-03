package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) buyEquipment(entity *Entity, equipmentID string, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil || equipmentID == "" {
		return
	}
	item, ok := w.equipment.Get(equipmentID)
	if !ok {
		return
	}
	if len(item.Components) > 0 {
		w.buyCompositeEquipment(entity, item, tick)
		return
	}
	if w.hasUniqueEquipmentGroup(entity, item.UniqueGroup) {
		w.setMessage(entity, "该类型装备只能装备一件", tick)
		return
	}
	if w.hasEquipmentCategory(entity, item.Category) && item.Category == "shoes" {
		w.setMessage(entity, "该类型装备只能装备一件", tick)
		return
	}
	if len(entity.Equipment) >= config.MaxEquipmentSlots {
		w.setMessage(entity, "装备栏已满", tick)
		return
	}
	if entity.Gold < float64(item.Price) {
		w.setMessage(entity, "金币不足", tick)
		return
	}
	w.addEquipment(entity, item, float64(item.Price))
}

func (w *World) buyCompositeEquipment(entity *Entity, item config.EquipmentConfig, tick uint64) {
	componentIndexes := w.findComponentIndexes(entity, item.Components)
	ownedCost := w.componentCost(entity, componentIndexes)
	combineCost := float64(item.Price - ownedCost)
	if combineCost < 0 {
		combineCost = 0
	}
	if len(componentIndexes) > 0 && entity.Gold >= combineCost {
		w.replaceComponentsWithEquipment(entity, item, componentIndexes, combineCost)
		return
	}
	if len(componentIndexes) == 0 && entity.Gold >= float64(item.Price) {
		if w.hasUniqueEquipmentGroup(entity, item.UniqueGroup) || (item.Category == "shoes" && w.hasEquipmentCategory(entity, "shoes")) {
			w.setMessage(entity, "该类型装备只能装备一件", tick)
			return
		}
		if len(entity.Equipment) >= config.MaxEquipmentSlots {
			w.setMessage(entity, "装备栏已满", tick)
			return
		}
		w.addEquipment(entity, item, float64(item.Price))
		return
	}
	if w.tryBuyMissingComponent(entity, item.Components, componentIndexes, tick) {
		return
	}
	w.setMessage(entity, "金币不足", tick)
}

func (w *World) addEquipment(entity *Entity, item config.EquipmentConfig, cost float64) {
	entity.Gold -= cost
	entity.Equipment = append(entity.Equipment, equipmentSlotFromConfig(item))
	w.recalculatePlayerStats(entity)
	w.refreshStoneplateShield(entity)
}

func (w *World) replaceComponentsWithEquipment(entity *Entity, item config.EquipmentConfig, componentIndexes []int, cost float64) {
	entity.Gold -= cost
	indexSet := make(map[int]bool, len(componentIndexes))
	for _, index := range componentIndexes {
		indexSet[index] = true
	}
	next := entity.Equipment[:0]
	for index, equipped := range entity.Equipment {
		if !indexSet[index] {
			next = append(next, equipped)
		}
	}
	entity.Equipment = append(next, equipmentSlotFromConfig(item))
	w.recalculatePlayerStats(entity)
	w.refreshStoneplateShield(entity)
}

func equipmentSlotFromConfig(item config.EquipmentConfig) EquipmentSlot {
	seconds := item.Effects.OutOfCombatSeconds
	if seconds <= 0 {
		seconds = 5
	}
	return EquipmentSlot{
		EquipmentID:              item.EquipmentID,
		Name:                     item.Name,
		LowHealthDamageReduce:    item.Effects.LowHealthDamageReduce,
		LowHealthShieldThreshold: item.Effects.LowHealthShieldThreshold,
		LowHealthShieldMin:       item.Effects.LowHealthShieldMin,
		LowHealthShieldMax:       item.Effects.LowHealthShieldMax,
		OutOfCombatMoveSpeed:     item.Effects.OutOfCombatMoveSpeed,
		OutOfCombatRequiredTicks: uint64(math.Ceil(seconds * 20)),
		StoneplateShieldRatio:    item.Effects.StoneplateShieldMaxHPRatio,
		StoneplateResistPercent:  item.Effects.StoneplateResistPercent,
		StoneplateCooldownTicks:  uint64(math.Ceil(item.Effects.StoneplateCooldownSeconds * 20)),
	}
}

func (w *World) findComponentIndexes(entity *Entity, components []string) []int {
	if entity == nil || len(components) == 0 {
		return nil
	}
	used := make(map[int]bool, len(components))
	indexes := make([]int, 0, len(components))
	for _, componentID := range components {
		for index, equipped := range entity.Equipment {
			if used[index] || equipped.EquipmentID != componentID {
				continue
			}
			used[index] = true
			indexes = append(indexes, index)
			break
		}
	}
	return indexes
}

func (w *World) componentCost(entity *Entity, componentIndexes []int) int {
	cost := 0
	for _, index := range componentIndexes {
		if index < 0 || index >= len(entity.Equipment) {
			continue
		}
		item, ok := w.equipment.Get(entity.Equipment[index].EquipmentID)
		if ok {
			cost += item.Price
		}
	}
	return cost
}

func (w *World) tryBuyMissingComponent(entity *Entity, components []string, ownedIndexes []int, tick uint64) bool {
	if len(entity.Equipment) >= config.MaxEquipmentSlots {
		w.setMessage(entity, "装备栏已满", tick)
		return true
	}
	owned := make(map[string]int, len(ownedIndexes))
	for _, index := range ownedIndexes {
		if index >= 0 && index < len(entity.Equipment) {
			owned[entity.Equipment[index].EquipmentID]++
		}
	}
	for _, componentID := range components {
		if owned[componentID] > 0 {
			owned[componentID]--
			continue
		}
		component, ok := w.equipment.Get(componentID)
		if !ok {
			continue
		}
		if w.hasUniqueEquipmentGroup(entity, component.UniqueGroup) || (component.Category == "shoes" && w.hasEquipmentCategory(entity, "shoes")) {
			continue
		}
		if entity.Gold >= float64(component.Price) {
			w.addEquipment(entity, component, float64(component.Price))
			return true
		}
	}
	return false
}

func (w *World) sellEquipment(entity *Entity, slot int) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	index := slot - 1
	if index < 0 || index >= len(entity.Equipment) {
		return
	}
	equipped := entity.Equipment[index]
	item, ok := w.equipment.Get(equipped.EquipmentID)
	if !ok {
		return
	}
	entity.Gold += math.Floor(float64(item.Price)*item.SellRatio + 0.000000001)
	removeStoneplateShieldFromSlot(entity, index)
	removePhysicalDamageShieldFromSlot(entity, index)
	entity.Equipment = append(entity.Equipment[:index], entity.Equipment[index+1:]...)
	w.recalculatePlayerStats(entity)
	w.refreshStoneplateShield(entity)
}
