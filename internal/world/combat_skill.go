package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) swordQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, tick uint64) int {
	state := attacker.Skills[swordQSkillID]
	baseDamage := skillMetaListByLevel(skill, "baseDamage", state.Level, []float64{20, 45, 70, 95, 120})
	attack := baseDamage + attacker.Stats.Attack*skillMetaRange(skill, "adRatio", 1)
	crit := w.attackCrits(attacker, target, tick)
	if crit {
		attack *= w.critDamageMultiplier(attacker)
	}
	return reduceCritDamage(target, w.applyCritFinalDamageMultiplier(attacker, physicalDamageAfterResistance(attacker, target, attack, tick), crit), crit)
}

func (w *World) swordQCooldownTicks(entity *Entity, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	attackSpeedBonus := 0.0
	if entity != nil {
		attackSpeedBonus = entity.Stats.AttackSpeedBonus
	}
	return swordQCooldownTicksByBonus(attackSpeedBonus, skill, skillLevel, tickRate)
}

func swordQCooldownTicksByBonus(attackSpeedBonus float64, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	baseCooldownMS := skillMetaListByLevelMS(skill, "cooldownMs", skillLevel, []float64{6000, 5500, 5000, 4500, 4000})
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	seconds := float64(baseCooldownMS) / 1000 * (1 - attackSpeedBonus*0.6)
	minSeconds := skillMetaRange(skill, "minCooldownSeconds", 1.33)
	if seconds < minSeconds {
		seconds = minSeconds
	}
	return uint64(math.Ceil(seconds*float64(tickRate) - 1e-6))
}
