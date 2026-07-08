package world

import "math"

func (w *World) triggerEquipmentThorns(attacker *Entity, target *Entity, tickRate int) {
	if attacker == nil || target == nil || target.Kind != EntityKindPlayer || attacker.Stats.HP <= 0 || w.equipment == nil {
		return
	}
	flat, ratio, wounds, seconds := w.equipmentThorns(target)
	if flat <= 0 && ratio <= 0 {
		return
	}
	tick := target.Combat.LastHitTick
	raw := flat + target.Stats.BonusPhysicalDefense*ratio
	damage := magicDamageAfterResistance(target, attacker, raw, tick)
	if damage > 0 {
		attacker.Combat.LastHitTick = tick
		w.applyResolvedDamage(target, attacker, damage, "magic", sustainEquipmentDamage, tickRate)
	}
	if wounds > 0 && IsHeroUnit(attacker) {
		if seconds <= 0 {
			seconds = 3
		}
		w.applyGrievousWounds(attacker, wounds, tick+secondsToTicks(seconds, tickRate))
	}
}

func (w *World) equipmentThorns(entity *Entity) (float64, float64, float64, float64) {
	flat, ratio, wounds, seconds := 0.0, 0.0, 0.0, 0.0
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || (item.Effects.ThornsDamageFlat <= 0 && item.Effects.ThornsBonusArmorRatio <= 0) {
			continue
		}
		flat = math.Max(flat, item.Effects.ThornsDamageFlat)
		ratio = math.Max(ratio, item.Effects.ThornsBonusArmorRatio)
		wounds = math.Max(wounds, item.Effects.ThornsGrievousWounds)
		seconds = math.Max(seconds, item.Effects.ThornsGrievousWoundsSeconds)
	}
	return flat, ratio, wounds, seconds
}
