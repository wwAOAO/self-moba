package world

import "math"

func (w *World) refreshStoneplateShield(entity *Entity) {
	if entity == nil || entity.Kind != EntityKindPlayer {
		return
	}
	for index := range entity.Equipment {
		equipped := &entity.Equipment[index]
		if equipped.StoneplateShieldRatio <= 0 || equipped.StoneplateShieldActive || equipped.StoneplateCooldownUntil > 0 {
			continue
		}
		shield := int(math.Round(entity.Stats.MaxHP * equipped.StoneplateShieldRatio))
		if shield <= 0 {
			continue
		}
		equipped.StoneplateShieldActive = true
		equipped.StoneplateShieldAmount = shield
		entity.Passive.Shield += shield
		entity.Passive.MaxShield += shield
		w.recalculatePlayerStats(entity)
		return
	}
}

func (w *World) tickStoneplateShield(entity *Entity, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer {
		return
	}
	for index := range entity.Equipment {
		equipped := &entity.Equipment[index]
		if equipped.StoneplateShieldActive && equipped.StoneplateBreakTick > 0 && tick >= equipped.StoneplateBreakTick {
			targetShield := equipped.StoneplateShieldAmount
			if targetShield > entity.Passive.Shield {
				targetShield = entity.Passive.Shield
			}
			entity.Passive.Shield -= targetShield
			if entity.Passive.Shield < 0 {
				entity.Passive.Shield = 0
			}
			deactivateStoneplateShield(entity)
			w.recalculatePlayerStats(entity)
		}
		if equipped.StoneplateShieldRatio <= 0 || equipped.StoneplateShieldActive || equipped.StoneplateCooldownUntil == 0 || tick < equipped.StoneplateCooldownUntil {
			continue
		}
		equipped.StoneplateCooldownUntil = 0
		equipped.StoneplateBreakTick = 0
	}
	w.refreshStoneplateShield(entity)
}

func (w *World) triggerStoneplateCooldown(target *Entity, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer {
		return
	}
	tick := target.Combat.LastHitTick
	for index := range target.Equipment {
		equipped := &target.Equipment[index]
		if !equipped.StoneplateShieldActive || equipped.StoneplateCooldownTicks == 0 || equipped.StoneplateCooldownUntil > tick {
			continue
		}
		equipped.StoneplateCooldownUntil = tick + equipped.StoneplateCooldownTicks
		equipped.StoneplateBreakTick = tick + secondsToTicks(5, tickRate)
		if tickRate <= 0 {
			equipped.StoneplateCooldownUntil = tick + 2400
			equipped.StoneplateBreakTick = tick + 100
		}
	}
}

func deactivateStoneplateShield(target *Entity) bool {
	if target == nil || target.Kind != EntityKindPlayer {
		return false
	}
	changed := false
	for index := range target.Equipment {
		equipped := &target.Equipment[index]
		if equipped.StoneplateShieldActive && target.Passive.Shield <= 0 {
			equipped.StoneplateShieldActive = false
			equipped.StoneplateShieldAmount = 0
			equipped.StoneplateBreakTick = 0
			changed = true
		}
	}
	if changed {
		target.Passive.MaxShield = target.Passive.Shield
	}
	return changed
}

func removeStoneplateShieldFromSlot(entity *Entity, index int) {
	if entity == nil || index < 0 || index >= len(entity.Equipment) {
		return
	}
	equipped := &entity.Equipment[index]
	if !equipped.StoneplateShieldActive || equipped.StoneplateShieldAmount <= 0 {
		return
	}
	removed := equipped.StoneplateShieldAmount
	if removed > entity.Passive.Shield {
		removed = entity.Passive.Shield
	}
	entity.Passive.Shield -= removed
	if entity.Passive.Shield < 0 {
		entity.Passive.Shield = 0
	}
	entity.Passive.MaxShield = entity.Passive.Shield
	equipped.StoneplateShieldActive = false
	equipped.StoneplateShieldAmount = 0
	equipped.StoneplateBreakTick = 0
}

func removePhysicalDamageShieldFromSlot(entity *Entity, index int) {
	if entity == nil || index < 0 || index >= len(entity.Equipment) {
		return
	}
	removed := entity.Equipment[index].PhysicalShieldAmount
	if removed <= 0 {
		return
	}
	if removed > entity.Passive.Shield {
		removed = entity.Passive.Shield
	}
	entity.Passive.Shield -= removed
	if entity.Passive.Shield < 0 {
		entity.Passive.Shield = 0
	}
	if entity.Passive.MaxShield > entity.Passive.Shield {
		entity.Passive.MaxShield = entity.Passive.Shield
	}
	entity.Equipment[index].PhysicalShieldAmount = 0
	entity.Equipment[index].PhysicalShieldMaxAmount = 0
}

func removeLifeStealOverhealShieldFromSlot(entity *Entity, index int) {
	if entity == nil || index < 0 || index >= len(entity.Equipment) {
		return
	}
	removed := entity.Equipment[index].LifeStealOverhealShield
	if removed <= 0 {
		return
	}
	if removed > entity.Passive.Shield {
		removed = entity.Passive.Shield
	}
	entity.Passive.Shield -= removed
	if entity.Passive.Shield < 0 {
		entity.Passive.Shield = 0
	}
	if entity.Passive.MaxShield > entity.Passive.Shield {
		entity.Passive.MaxShield = entity.Passive.Shield
	}
	entity.Equipment[index].LifeStealOverhealShield = 0
}

func (w *World) applyStoneplateResists(entity *Entity, stats *Stats) {
	if entity == nil || stats == nil {
		return
	}
	for _, equipped := range entity.Equipment {
		if !equipped.StoneplateShieldActive || equipped.StoneplateResistPercent <= 0 {
			continue
		}
		stats.PhysicalDefense *= 1 + equipped.StoneplateResistPercent
		stats.MagicDefense *= 1 + equipped.StoneplateResistPercent
		return
	}
}
