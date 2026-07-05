package world

import "math"

func (w *World) triggerEquipmentPhysicalDamageEffects(source *Entity, target *Entity, damage int, tick uint64, tickRate int) {
	if source == nil || target == nil || source.Kind != EntityKindPlayer || damage <= 0 || w.equipment == nil {
		return
	}
	for index, equipped := range source.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.PhysicalHitArmorShredMaxStacks > 0 && (target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero) {
			target.Combat.BlackCleaverStacks++
			if target.Combat.BlackCleaverStacks > int(item.Effects.PhysicalHitArmorShredMaxStacks) {
				target.Combat.BlackCleaverStacks = int(item.Effects.PhysicalHitArmorShredMaxStacks)
			}
			target.Combat.BlackCleaverUntil = tick + secondsToTicks(item.Effects.PhysicalHitArmorShredSeconds, tickRate)
		}
		if item.Effects.PhysicalHitMoveSpeed > 0 {
			bonus := item.Effects.PhysicalHitMoveSpeed
			if isRangedBasicAttacker(source) {
				bonus *= 0.5
			}
			source.Control.MoveSpeedBonusFlat = bonus
			source.Control.MoveSpeedBonusUntil = tick + secondsToTicks(item.Effects.PhysicalHitMoveSpeedSeconds, tickRate)
		}
		if item.Effects.PhysicalDamageShieldRatio > 0 {
			shield := int(math.Floor(float64(damage) * item.Effects.PhysicalDamageShieldRatio))
			if shield <= 0 {
				continue
			}
			seconds := item.Effects.PhysicalDamageShieldDecaySeconds
			if seconds <= 0 {
				seconds = 3
			}
			source.Passive.Shield += shield
			source.Passive.MaxShield += shield
			source.Equipment[index].PhysicalShieldMaxAmount += shield
			source.Equipment[index].PhysicalShieldAmount += shield
			source.Equipment[index].PhysicalShieldStartTick = tick
			source.Equipment[index].PhysicalShieldExpireTick = tick + secondsToTicks(seconds, tickRate)
		}
	}
}
