package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) addArcherFocusStack(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	state, ok := entity.Skills[archerQSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	if tick < entity.Archer.FocusActiveUntil {
		return
	}
	skill := w.SkillConfig(archerQSkillID)
	maxStacks := archerFocusMaxStacks(skill)
	if entity.Archer.FocusStacks < maxStacks {
		entity.Archer.FocusStacks++
	}
	entity.Archer.FocusExpireTick = tick + secondsToTicks(skillMetaRange(skill, "stackDurationSeconds", 4), tickRate)
	entity.Skills[archerQSkillID] = state
}

func archerFocusMaxStacks(skill config.SkillConfig) int {
	maxStacks := int(math.Round(skillMetaRange(skill, "maxStacks", 4)))
	if maxStacks < 1 {
		return 1
	}
	return maxStacks
}

func (w *World) expireArcherFocus(entity *Entity, tick uint64) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	if entity.Archer.FocusStacks > 0 && entity.Archer.FocusExpireTick > 0 && tick >= entity.Archer.FocusExpireTick {
		entity.Archer.FocusStacks = 0
		entity.Archer.FocusExpireTick = 0
	}
	if entity.Archer.FocusActiveUntil > 0 && tick >= entity.Archer.FocusActiveUntil {
		entity.Archer.FocusActiveUntil = 0
		entity.Archer.FocusActiveLevel = 0
		entity.Archer.FocusAttackSpeed = 0
		entity.Archer.FocusBonusADRatio = 0
	}
}

func (w *World) archerFocusBonusDamage(attacker *Entity, target *Entity, tick uint64) int {
	if attacker == nil || target == nil || attacker.HeroID != archerHeroID || tick >= attacker.Archer.FocusActiveUntil {
		return 0
	}
	ratio := attacker.Archer.FocusBonusADRatio
	if ratio <= 0 {
		skill := w.SkillConfig(archerQSkillID)
		ratio = skillMetaListByLevel(skill, "bonusAdDamageRatio", attacker.Archer.FocusActiveLevel, []float64{1.05, 1.1, 1.15, 1.2, 1.25})
	}
	return physicalDamageAfterResistance(attacker, target, attacker.Stats.Attack*ratio, tick)
}

func (w *World) applyArcherFocusOnBasicHit(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.HeroID != archerHeroID {
		return
	}
	if tick < attacker.Archer.FocusActiveUntil {
		damage := w.archerFocusBonusDamage(attacker, target, tick)
		if damage > 0 {
			previousDamage := target.Combat.LastDamage
			target.Combat.LastHitTick = tick
			if target.Kind != EntityKindDummy {
				wasAlive := target.Stats.HP > 0
				w.ApplyDamage(attacker, target, damage, tickRate)
				if previousDamage > 0 {
					target.Combat.LastDamage += previousDamage
				}
				if wasAlive && target.Stats.HP == 0 {
					w.ApplyKillReward(attacker, target)
					w.KillPlayer(target, tick, tickRate)
					w.RemoveDeadUnit(target)
				}
			} else {
				target.Combat.LastDamage = damage
				target.Combat.LastDamageType = "physical"
				if previousDamage > 0 {
					target.Combat.LastDamage += previousDamage
				}
			}
		}
		return
	}
	w.addArcherFocusStack(attacker, tick, tickRate)
}
