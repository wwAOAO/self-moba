package world

import "math"

func (w *World) applyEquipmentDamageMultiplier(source *Entity, target *Entity, damage int) int {
	if source == nil || target == nil || damage <= 0 || source.Kind != EntityKindPlayer || !IsHeroUnit(target) || w.equipment == nil {
		return damage
	}
	bonus := 0.0
	seen := make(map[string]bool, len(source.Equipment))
	for _, equipped := range source.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.TargetBonusHPDamageMaxRatio <= bonus {
			continue
		}
		fullAt := item.Effects.TargetBonusHPDamageFullAt
		if fullAt <= 0 {
			fullAt = 1500
		}
		ratio := clamp(target.Stats.BonusHP/fullAt, 0, 1) * item.Effects.TargetBonusHPDamageMaxRatio
		if ratio > bonus {
			bonus = ratio
		}
	}
	if bonus <= 0 {
		return damage
	}
	return int(math.Round(float64(damage) * (1 + bonus)))
}

func (w *World) equipmentBasicAttackBonus(attacker *Entity, damageType string) float64 {
	if attacker == nil || w.equipment == nil {
		return 0
	}
	var bonus float64
	seen := make(map[string]bool, len(attacker.Equipment))
	for _, equipped := range attacker.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.BasicAttackBonusDamageType == damageType {
			bonus += item.Effects.BasicAttackBonusDamage
		}
		if damageType == "magic" && item.Effects.BasicAttackBonusByCritMax > 0 {
			onHit := math.Min(item.Effects.BasicAttackBonusByCritMax, w.rawEquipmentCritChance(attacker)*item.Effects.BasicAttackBonusByCritMax)
			bonus += onHit
			if item.Effects.EveryNthBasicAttackBonusHit > 0 && equipped.RagebladeHitCount > 0 && equipped.RagebladeHitCount%int(item.Effects.EveryNthBasicAttackBonusHit) == 0 {
				bonus += item.Effects.BasicAttackBonusDamage + onHit
			}
		}
	}
	return bonus
}

func (w *World) equipmentMinionBasicAttackBonus(attacker *Entity, damageType string) float64 {
	if attacker == nil || w.equipment == nil {
		return 0
	}
	var bonus float64
	seen := make(map[string]bool, len(attacker.Equipment))
	for _, equipped := range attacker.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.MinionBasicAttackBonusDamageType == damageType {
			bonus += item.Effects.MinionBasicAttackBonusDamage
		}
	}
	return bonus
}

func (w *World) equipmentCritDamageBonus(attacker *Entity) float64 {
	if attacker == nil || w.equipment == nil {
		return 0
	}
	bonus := 0.0
	seen := make(map[string]bool, len(attacker.Equipment))
	for _, equipped := range attacker.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.CritDamageBonus <= bonus {
			continue
		}
		bonus = item.Effects.CritDamageBonus
	}
	return bonus
}

func (w *World) triggerEquipmentBasicAttackStacks(source *Entity, tick uint64, tickRate int) {
	if source == nil || source.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	changed := false
	for index, equipped := range source.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.BasicAttackAttackSpeedMaxStacks <= 0 {
			continue
		}
		if equipped.Stacks < item.Effects.BasicAttackAttackSpeedMaxStacks {
			source.Equipment[index].Stacks++
			changed = true
		}
		source.Equipment[index].RagebladeHitCount++
		source.Equipment[index].StackExpireTick = tick + secondsToTicks(5, tickRate)
	}
	if changed {
		w.recalculatePlayerStats(source)
	}
}

func (w *World) tickEquipmentStacks(entity *Entity, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	changed := false
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.BasicAttackAttackSpeedMaxStacks <= 0 || equipped.StackExpireTick == 0 || tick < equipped.StackExpireTick {
			continue
		}
		entity.Equipment[index].Stacks = 0
		entity.Equipment[index].RagebladeHitCount = 0
		entity.Equipment[index].StackExpireTick = 0
		changed = true
	}
	if changed {
		w.recalculatePlayerStats(entity)
	}
}
