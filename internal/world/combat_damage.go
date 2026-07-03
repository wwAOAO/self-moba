package world

import (
	"l-battle/internal/config"
	"math"
)

func physicalDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	damageReduce := target.damageReductionForType("physical", tick)
	resistance := target.Stats.PhysicalDefense
	if tick < target.Combat.BlackCleaverUntil && target.Combat.BlackCleaverStacks > 0 {
		resistance *= 1 - float64(target.Combat.BlackCleaverStacks)*0.05
	}
	return damageAfterResistance(rawDamage, effectiveResistance(resistance, attacker.Stats.PhysicalPenPercent, attacker.Stats.PhysicalPenFlat), damageReduce)
}

func magicDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	damageReduce := target.damageReductionForType("magic", tick)
	return damageAfterResistance(rawDamage, effectiveResistance(target.Stats.MagicDefense, attacker.Stats.MagicPenPercent, attacker.Stats.MagicPenFlat), damageReduce)
}

func tankQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{70, 120, 170, 220, 270})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func applyTankQMoveSpeedSteal(source *Entity, target *Entity, ratio float64, until uint64) {
	if source == nil || target == nil || ratio <= 0 || until == 0 {
		return
	}
	ratio = clamp(ratio, 0, 1)
	stolen := EffectiveMoveSpeedAtTick(target, 0) * ratio
	source.Control.MoveSpeedBonusFlat = stolen
	source.Control.MoveSpeedBonusUntil = until
	target.Control.MoveSpeedSlow = ratio
	target.Control.MoveSpeedSlowUntil = until
}

func trueDamageAfterReduction(target *Entity, rawDamage float64, tick uint64) int {
	return damageAfterResistance(rawDamage, 0, target.damageReductionForType("true", tick))
}

func (entity *Entity) damageReductionForType(damageType string, tick uint64) float64 {
	reductions := []float64{entity.Stats.DamageReduce}
	switch damageType {
	case "physical":
		reductions = append(reductions, entity.Stats.PhysicalDamageReduce)
	case "magic":
		reductions = append(reductions, entity.Stats.MagicDamageReduce)
	}
	if entity.HeroID == warriorHeroID && entity.Warrior.CourageUntilTick > 0 {
		reductions = append(reductions, entity.Warrior.courageDamageReductionAtTick(tick))
	}
	reductions = append(reductions, equipmentLowHealthDamageReduce(entity))
	return stackDamageReduction(reductions...)
}

func (entity *Entity) tenacityAtTick(tick uint64) float64 {
	tenacity := []float64{entity.Stats.Tenacity}
	if entity.HeroID == warriorHeroID && entity.Warrior.CourageFrontUntilTick > 0 && (tick == 0 || tick < entity.Warrior.CourageFrontUntilTick) {
		tenacity = append(tenacity, entity.Warrior.CourageFrontTenacity)
	}
	return stackTenacity(tenacity...)
}

func controlTicksAfterTenacity(target *Entity, ticks uint64, tick uint64) uint64 {
	if target == nil || ticks == 0 {
		return ticks
	}
	remainingRatio := 1 - target.tenacityAtTick(tick)
	adjusted := uint64(math.Ceil(float64(ticks) * remainingRatio))
	if adjusted < 1 {
		return 1
	}
	return adjusted
}

func (state WarriorState) courageDamageReductionAtTick(tick uint64) float64 {
	if state.CourageUntilTick == 0 {
		return 0
	}
	if tick > 0 && tick >= state.CourageUntilTick {
		return 0
	}
	if tick == 0 || tick < state.CourageFrontUntilTick {
		return state.CourageFrontDamageReduce
	}
	return state.CourageBackDamageReduce
}

func warriorWShieldValue(entity *Entity, skill config.SkillConfig, skillLevel int) int {
	baseShield := skillMetaListByLevel(skill, "shieldValue", skillLevel, []float64{70, 95, 120, 145, 170})
	return int(math.Round(baseShield + float64(entity.Stats.BonusHP)*skillMetaRange(skill, "bonusHealthRatio", 0.2)))
}

func effectiveResistance(resistance float64, percentPen float64, flatPen float64) float64 {
	if resistance < 0 {
		return resistance
	}
	if percentPen < 0 {
		percentPen = 0
	}
	if percentPen > 1 {
		percentPen = 1
	}
	if flatPen < 0 {
		flatPen = 0
	}
	effective := resistance*(1-percentPen) - flatPen
	if effective < 0 {
		return 0
	}
	return effective
}

func damageAfterResistance(rawDamage float64, resistance float64, damageReduce float64) int {
	if rawDamage <= 0 {
		return 0
	}
	multiplier := 100 / (resistance + 100)
	if resistance < 0 {
		denominator := 100 + resistance
		if denominator < 1 {
			denominator = 1
		}
		multiplier = 100 / denominator
	}
	damageReduce = clamp(damageReduce, 0, 1)
	damage := int(math.Round(rawDamage * multiplier * (1 - damageReduce)))
	if damage < 1 {
		return 1
	}
	return damage
}

func stackDamageReduction(reductions ...float64) float64 {
	multiplier := 1.0
	for _, reduction := range reductions {
		reduction = clamp(reduction, 0, 1)
		multiplier *= 1 - reduction
	}
	return 1 - multiplier
}

func stackTenacity(tenacityValues ...float64) float64 {
	multiplier := 1.0
	for _, tenacity := range tenacityValues {
		tenacity = clamp(tenacity, 0, 1)
		multiplier *= 1 - tenacity
	}
	return 1 - multiplier
}
