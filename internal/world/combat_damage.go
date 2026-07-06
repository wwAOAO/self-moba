package world

import (
	"l-battle/internal/config"
	"l-battle/internal/world/formula"
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

func (w *World) tankQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	if heroHooksFor(tankHeroID).TankQDamage != nil {
		return heroHooksFor(tankHeroID).TankQDamage(w, attacker, target, skill, skillLevel, tick)
	}
	return 0
}

func (w *World) ninjaQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, hitNumber int, tick uint64) int {
	if heroHooksFor(ninjaHeroID).NinjaQDamage != nil {
		return heroHooksFor(ninjaHeroID).NinjaQDamage(w, attacker, target, skill, skillLevel, hitNumber, tick)
	}
	return 0
}

func (w *World) gunnerQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, crit bool, tick uint64) int {
	rawDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{20, 45, 70, 95, 120})
	rawDamage += attacker.Stats.Attack * skillMetaRange(skill, "totalAdRatio", 1)
	rawDamage += float64(attacker.Stats.AbilityPower) * skillMetaRange(skill, "apRatio", 0.35)
	return w.PhysicalCritDamageAfterResistance(attacker, target, rawDamage, crit, tick)
}

func (w *World) gunnerRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	rawDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{20, 30, 40})
	rawDamage += attacker.Stats.Attack * skillMetaRange(skill, "adRatio", 0.6)
	rawDamage += float64(attacker.Stats.AbilityPower) * skillMetaRange(skill, "apRatio", 0.25)
	return w.PhysicalCritDamageAfterResistance(attacker, target, rawDamage, w.attackCrits(attacker, target, tick), tick)
}

func (w *World) ninjaSkillHit(source *Entity, target *Entity, skillID string, groupID string, fromShadow bool, tick uint64, tickRate int) {
	if h := heroHooksFor(ninjaHeroID).NinjaSkillHit; h != nil {
		h(w, source, target, skillID, groupID, fromShadow, tick, tickRate)
	}
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
		reductions = append(reductions, warriorCourageDamageReductionAtTick(entity.Warrior, tick))
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

func warriorCourageDamageReductionAtTick(state WarriorState, tick uint64) float64 {
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
	return formula.EffectiveResistance(resistance, percentPen, flatPen)
}

func damageAfterResistance(rawDamage float64, resistance float64, damageReduce float64) int {
	return formula.DamageAfterResistance(rawDamage, resistance, damageReduce)
}

func stackDamageReduction(reductions ...float64) float64 {
	return formula.StackDamageReduction(reductions...)
}

func stackTenacity(tenacityValues ...float64) float64 {
	return formula.StackTenacity(tenacityValues...)
}
