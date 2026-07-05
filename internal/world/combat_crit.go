package world

import (
	"l-battle/internal/world/formula"
	"math"
)

func (w *World) attackCrits(attacker *Entity, target *Entity, tick uint64) bool {
	chance := w.critChance(attacker)
	if chance <= 0 {
		return false
	}
	if chance >= 1 {
		return true
	}
	roll := deterministicCritRoll(attacker.ID, target.ID, tick)
	return roll < chance
}

func (w *World) critChance(attacker *Entity) float64 {
	chance := attacker.Stats.CritChance
	if attacker.HeroID == swordHeroID {
		chance *= skillMetaRange(w.heroPassiveSkill(attacker), "critChanceMultiplier", 2)
	}
	if attacker.HeroID == bladeHeroID {
		chance += bladeRageCritChance(attacker, w.heroPassiveSkill(attacker))
	}
	if chance > 1 {
		return 1
	}
	if chance < 0 {
		return 0
	}
	return chance
}

func (w *World) DisplayCritChance(attacker *Entity) float64 {
	return w.critChance(attacker)
}

func (w *World) critDamageMultiplier(attacker *Entity) float64 {
	return 2.0 + w.equipmentCritDamageBonus(attacker)
}

func (w *World) applyCritFinalDamageMultiplier(attacker *Entity, damage int, crit bool) int {
	if !crit || attacker == nil || attacker.HeroID != swordHeroID || damage <= 0 {
		return damage
	}
	multiplier := skillMetaRange(w.heroPassiveSkill(attacker), "critFinalDamageMultiplier", 0.9)
	result := int(math.Round(float64(damage) * multiplier))
	if result < 1 {
		return 1
	}
	return result
}

func deterministicCritRoll(attackerID string, targetID string, tick uint64) float64 {
	return formula.DeterministicCritRoll(attackerID, targetID, tick)
}
