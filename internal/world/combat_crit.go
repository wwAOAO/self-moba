package world

import (
	"l-battle/internal/world/formula"
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
	if heroHooksFor(swordHeroID).CritChanceMultiplier != nil {
		chance *= heroHooksFor(swordHeroID).CritChanceMultiplier(w, attacker)
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
	if heroHooksFor(swordHeroID).ApplyCritFinalMultiplier != nil {
		return heroHooksFor(swordHeroID).ApplyCritFinalMultiplier(w, attacker, damage, crit)
	}
	return damage
}

func deterministicCritRoll(attackerID string, targetID string, tick uint64) float64 {
	return formula.DeterministicCritRoll(attackerID, targetID, tick)
}
