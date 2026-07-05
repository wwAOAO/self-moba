package world

import (
	"l-battle/internal/world/equipcalc"
	"math"
)

func (w *World) triggerEquipmentLowHealthShield(target *Entity, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer || target.Stats.MaxHP <= 0 || target.Passive.Shield > 0 {
		return
	}
	seen := make(map[string]bool, len(target.Equipment))
	for index, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		if equipped.LowHealthShieldUsed || equipped.LowHealthShieldMax <= 0 {
			continue
		}
		threshold := equipped.LowHealthShieldThreshold
		if threshold <= 0 {
			threshold = 0.3
		}
		if float64(target.Stats.HP)/float64(target.Stats.MaxHP) > threshold {
			continue
		}
		shield := equipmentShieldByLevel(target.Level, equipped.LowHealthShieldMin, equipped.LowHealthShieldMax)
		if shield <= 0 {
			continue
		}
		target.Equipment[index].LowHealthShieldUsed = true
		target.Passive.Shield = shield
		target.Passive.MaxShield = shield
		target.Passive.ShieldExpireTick = 0
		return
	}
}

func equipmentShieldByLevel(level int, minShield int, maxShield int) int {
	return equipcalc.ShieldByLevel(level, MinHeroLevel, MaxHeroLevel, minShield, maxShield)
}

func equipmentLowHealthDamageReduce(entity *Entity) float64 {
	if entity == nil || entity.Passive.Shield <= 0 {
		return 0
	}
	reduce := 0.0
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		if equipped.LowHealthShieldUsed && equipped.LowHealthDamageReduce > reduce {
			reduce = equipped.LowHealthDamageReduce
		}
	}
	return reduce
}

func (w *World) triggerEquipmentHeroHitHeal(source *Entity, target *Entity) {
	if source == nil || target == nil || target.Kind != EntityKindPlayer || !isHeroDamageSource(source) || w.equipment == nil {
		return
	}
	var heal int
	seen := make(map[string]bool, len(target.Equipment))
	for _, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || !item.Effects.HeroHitSmallHeal {
			continue
		}
		heal += item.Effects.HeroHitHeal
	}
	if heal <= 0 || target.Stats.HP <= 0 {
		return
	}
	target.Stats.HP += heal
	if target.Stats.HP > target.Stats.MaxHP {
		target.Stats.HP = target.Stats.MaxHP
	}
}

func (w *World) triggerEquipmentHeroDamageManaShield(source *Entity, target *Entity, tickRate int) {
	if source == nil || target == nil || target.Kind != EntityKindPlayer || target.Stats.MP <= 0 || !isHeroDamageSource(source) || w.equipment == nil {
		return
	}
	tick := target.Combat.LastHitTick
	seen := make(map[string]bool, len(target.Equipment))
	for index, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.HeroDamageManaShieldRatio <= 0 || tick < equipped.ManaShieldCooldownUntil {
			continue
		}
		shield := int(math.Floor(target.Stats.MP * item.Effects.HeroDamageManaShieldRatio))
		if shield <= 0 {
			return
		}
		target.Stats.MP -= float64(shield)
		target.Passive.Shield += shield
		target.Passive.MaxShield += shield
		cooldownTicks := uint64(math.Ceil(float64(item.Effects.HeroDamageManaShieldCooldownMS) / 1000 * float64(tickRate)))
		target.Equipment[index].ManaShieldCooldownUntil = tick + cooldownTicks
		return
	}
}

func tickEquipmentPhysicalDamageShield(entity *Entity, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Passive.Shield <= 0 {
		return
	}
	decayed := 0
	for index := range entity.Equipment {
		equipped := &entity.Equipment[index]
		if equipped.PhysicalShieldAmount <= 0 {
			continue
		}
		if tick >= equipped.PhysicalShieldExpireTick || equipped.PhysicalShieldExpireTick <= equipped.PhysicalShieldStartTick {
			decayed += equipped.PhysicalShieldAmount
			equipped.PhysicalShieldMaxAmount = 0
			equipped.PhysicalShieldAmount = 0
			continue
		}
		remaining := float64(equipped.PhysicalShieldExpireTick - tick)
		duration := float64(equipped.PhysicalShieldExpireTick - equipped.PhysicalShieldStartTick)
		next := int(math.Ceil(float64(equipped.PhysicalShieldMaxAmount) * remaining / duration))
		if next < equipped.PhysicalShieldAmount {
			decayed += equipped.PhysicalShieldAmount - next
			equipped.PhysicalShieldAmount = next
		}
	}
	if decayed > entity.Passive.Shield {
		decayed = entity.Passive.Shield
	}
	entity.Passive.Shield -= decayed
	if entity.Passive.MaxShield > entity.Passive.Shield {
		entity.Passive.MaxShield = entity.Passive.Shield
	}
}

func consumeEquipmentPhysicalDamageShield(entity *Entity, absorbed int) {
	if entity == nil || absorbed <= 0 {
		return
	}
	remaining := absorbed
	for index := range entity.Equipment {
		if remaining <= 0 {
			return
		}
		equipped := &entity.Equipment[index]
		if equipped.PhysicalShieldAmount <= 0 {
			continue
		}
		if equipped.PhysicalShieldAmount <= remaining {
			remaining -= equipped.PhysicalShieldAmount
			equipped.PhysicalShieldMaxAmount = 0
			equipped.PhysicalShieldAmount = 0
			continue
		}
		equipped.PhysicalShieldAmount -= remaining
		equipped.PhysicalShieldMaxAmount -= remaining
		return
	}
}
