package world

import "math"

func (w *World) applyEquipmentLowHealthMagicTrueDamageBonus(source *Entity, target *Entity, damage int, damageType string, context sustainContext) int {
	if source == nil || target == nil || damage <= 0 || source.Kind != EntityKindPlayer || w.equipment == nil || (damageType != "magic" && damageType != "true") {
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
		if !ok || item.Effects.LowHealthMagicTrueDamageBonus <= 0 || target.Stats.MaxHP <= 0 {
			continue
		}
		threshold := item.Effects.LowHealthDamageThreshold
		if threshold <= 0 {
			threshold = 0.4
		}
		if target.Stats.HP/target.Stats.MaxHP >= threshold {
			continue
		}
		itemBonus := item.Effects.LowHealthMagicTrueDamageBonus
		if context.Pet && item.Effects.LowHealthDotPetDamageBonus > itemBonus {
			itemBonus = item.Effects.LowHealthDotPetDamageBonus
		}
		if itemBonus > bonus {
			bonus = itemBonus
		}
	}
	if bonus <= 0 {
		return damage
	}
	return int(math.Round(float64(damage) * (1 + bonus)))
}

func (w *World) triggerEquipmentHeroDamageBonus(source *Entity, target *Entity, tickRate int) {
	if source == nil || target == nil || source.Kind != EntityKindPlayer || !IsHeroUnit(target) || w.equipment == nil {
		return
	}
	tick := target.Combat.LastHitTick
	for index, equipped := range source.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.HeroDamageBonusMagicDamage <= 0 || tick < equipped.HeroDamageBonusUntil {
			continue
		}
		damage := magicDamageAfterResistance(source, target, item.Effects.HeroDamageBonusMagicDamage, tick)
		seconds := item.Effects.HeroDamageBonusCooldownSeconds
		if seconds <= 0 {
			seconds = 40
		}
		source.Equipment[index].HeroDamageBonusUntil = tick + secondsToTicks(seconds, tickRate)
		w.applyResolvedDamage(source, target, damage, "magic", sustainEquipmentDamage, tickRate)
		return
	}
}

func (w *World) chargeEquipmentOnCast(entity *Entity) {
	w.addEquipmentEchoCharge(entity, func(e itemEffectsEcho) float64 { return e.castCharge })
}

func (w *World) chargeEquipmentOnMove(entity *Entity, moved float64) {
	if moved <= 0 {
		return
	}
	w.addEquipmentEchoCharge(entity, func(e itemEffectsEcho) float64 {
		if e.moveDistancePerCharge <= 0 {
			return 0
		}
		return moved / e.moveDistancePerCharge
	})
}

type itemEffectsEcho struct {
	castCharge            float64
	moveDistancePerCharge float64
	chargePerStack        float64
	maxStacks             float64
}

func (w *World) addEquipmentEchoCharge(entity *Entity, chargeFor func(itemEffectsEcho) float64) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.EchoMaxStacks <= 0 {
			continue
		}
		echo := itemEffectsEcho{
			castCharge:            item.Effects.EchoCastCharge,
			moveDistancePerCharge: item.Effects.EchoMoveDistancePerCharge,
			chargePerStack:        item.Effects.EchoChargePerStack,
			maxStacks:             item.Effects.EchoMaxStacks,
		}
		charge := chargeFor(echo)
		if charge <= 0 {
			continue
		}
		if echo.chargePerStack <= 0 {
			echo.chargePerStack = 100
		}
		entity.Equipment[index].EchoCharge += charge
		for entity.Equipment[index].EchoCharge >= echo.chargePerStack && entity.Equipment[index].Stacks < echo.maxStacks {
			entity.Equipment[index].EchoCharge -= echo.chargePerStack
			entity.Equipment[index].Stacks++
		}
		if entity.Equipment[index].Stacks >= echo.maxStacks {
			entity.Equipment[index].Stacks = echo.maxStacks
			entity.Equipment[index].EchoCharge = 0
		}
	}
}

func (w *World) triggerEquipmentEcho(source *Entity, target *Entity, context sustainContext, tick uint64, tickRate int) {
	if source == nil || target == nil || source.Kind != EntityKindPlayer || context.BasicAttack || context.Pet || w.equipment == nil {
		return
	}
	for index, equipped := range source.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.EchoDamage <= 0 || equipped.Stacks <= 0 || tick < equipped.EchoCooldownUntil {
			continue
		}
		stacks := int(equipped.Stacks)
		source.Equipment[index].Stacks = 0
		source.Equipment[index].EchoCharge = 0
		seconds := item.Effects.EchoCooldownSeconds
		if seconds <= 0 {
			seconds = 12
		}
		source.Equipment[index].EchoCooldownUntil = tick + secondsToTicks(seconds, tickRate)
		w.applyEquipmentEchoDamage(source, target, item.Effects.EchoDamage, item.Effects.EchoAPRatio, tickRate)
		for _, bounce := range w.equipmentEchoBounces(source, target, item.Effects.EchoBounceRadius, stacks-1) {
			w.applyEquipmentEchoDamage(source, bounce, item.Effects.EchoDamage, item.Effects.EchoAPRatio, tickRate)
		}
		return
	}
}

func (w *World) applyEquipmentEchoDamage(source *Entity, target *Entity, base float64, apRatio float64, tickRate int) {
	damage := magicDamageAfterResistance(source, target, base+float64(source.Stats.AbilityPower)*apRatio, target.Combat.LastHitTick)
	w.applyResolvedDamage(source, target, damage, "magic", sustainEquipmentDamage, tickRate)
}

func (w *World) equipmentEchoBounces(source *Entity, primary *Entity, radius float64, count int) []*Entity {
	if count <= 0 || primary == nil {
		return nil
	}
	if radius <= 0 {
		radius = 600
	}
	hits := make([]*Entity, 0, count)
	for _, target := range w.targetsInRadius(source, primary.Position, radius) {
		if target.ID == primary.ID {
			continue
		}
		hits = append(hits, target)
		if len(hits) >= count {
			return hits
		}
	}
	return hits
}
