package world

import "math"

func (w *World) tickEquipmentPercentRegen(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 || tickRate <= 0 || tick%uint64(5*tickRate) != 0 || w.equipment == nil {
		return
	}
	beforeHP := entity.Stats.HP
	hpRatio, mpRatio := w.equipmentPercentRegenRatios(entity, tick, tickRate)
	if hpRatio > 0 && entity.Stats.HP < entity.Stats.MaxHP {
		entity.Stats.HP += int(math.Floor(float64(entity.Stats.MaxHP) * hpRatio))
		if entity.Stats.HP > entity.Stats.MaxHP {
			entity.Stats.HP = entity.Stats.MaxHP
		}
	}
	if mpRatio > 0 && entity.Stats.MP < entity.Stats.MaxMP {
		entity.Stats.MP += entity.Stats.MaxMP * mpRatio
		if entity.Stats.MP > entity.Stats.MaxMP {
			entity.Stats.MP = entity.Stats.MaxMP
		}
	}
	w.refreshPlayerStatsAfterHPChange(entity, beforeHP)
}

func (w *World) equipmentPercentRegenRatios(entity *Entity, tick uint64, tickRate int) (float64, float64) {
	outOfCombat := tick >= entity.Combat.LastHitTick+uint64(5*tickRate)
	hpRatio, mpRatio := 0.0, 0.0
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
		if outOfCombat {
			hpRatio += item.Effects.OutOfCombatHPRegenMaxHPRatio5
			mpRatio += item.Effects.OutOfCombatMPRegenMaxMPRatio5
			continue
		}
		hpRatio += item.Effects.CombatHPRegenMaxHPRatio5
		mpRatio += item.Effects.CombatMPRegenMaxMPRatio5
	}
	return hpRatio, mpRatio
}
