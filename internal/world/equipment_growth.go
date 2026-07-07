package world

import "math"

func (w *World) applyEquipmentLevelUpRestore(entity *Entity) {
	if entity == nil || w.equipment == nil {
		return
	}
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.LevelUpRestoreHPRatio > 0 {
			entity.Stats.HP += entity.Stats.MaxHP * item.Effects.LevelUpRestoreHPRatio
			if entity.Stats.HP > entity.Stats.MaxHP {
				entity.Stats.HP = entity.Stats.MaxHP
			}
		}
		if item.Effects.LevelUpRestoreMPRatio > 0 {
			entity.Stats.MP += entity.Stats.MaxMP * item.Effects.LevelUpRestoreMPRatio
			if entity.Stats.MP > entity.Stats.MaxMP {
				entity.Stats.MP = entity.Stats.MaxMP
			}
		}
	}
}

func (w *World) applyEquipmentUnitKillGrowth(killer *Entity, target *Entity) {
	if killer == nil || target == nil || w.equipment == nil || killer.Kind != EntityKindPlayer {
		return
	}
	changed := false
	seen := make(map[string]bool, len(killer.Equipment))
	for index, equipped := range killer.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || (item.Effects.UnitKillPhysicalDefenseGain <= 0 && item.Effects.UnitKillAbilityPowerGain <= 0) {
			continue
		}
		maxGain := item.Effects.UnitKillMaxGain
		if maxGain <= 0 {
			continue
		}
		currentGain := math.Max(
			equipped.Stacks*item.Effects.UnitKillPhysicalDefenseGain,
			equipped.Stacks*item.Effects.UnitKillAbilityPowerGain,
		)
		if currentGain >= maxGain {
			continue
		}
		nextStacks := equipped.Stacks + 1
		if nextStacks*item.Effects.UnitKillPhysicalDefenseGain > maxGain ||
			nextStacks*item.Effects.UnitKillAbilityPowerGain > maxGain {
			nextStacks = math.Floor(maxGain / math.Max(item.Effects.UnitKillPhysicalDefenseGain, item.Effects.UnitKillAbilityPowerGain))
		}
		killer.Equipment[index].Stacks = nextStacks
		changed = true
	}
	if changed {
		w.recalculatePlayerStats(killer)
	}
}
