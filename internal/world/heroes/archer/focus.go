package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func AddFocusStack(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	state, ok := entity.Skills[qID]
	if !ok || state.Level <= 0 {
		return
	}
	if tick < entity.Archer.FocusActiveUntil {
		return
	}
	skill := w.SkillConfig(qID)
	maxStacks := archerFocusMaxStacks(skill)
	if entity.Archer.FocusStacks < maxStacks {
		entity.Archer.FocusStacks++
	}
	entity.Archer.FocusExpireTick = tick + secondsToTicks(skillMeta(skill, "stackDurationSeconds", 4), tickRate)
	entity.Skills[qID] = state
}

func ExpireFocus(w *world.World, entity *world.Entity, tick uint64) {
	if entity == nil || entity.HeroID != heroID {
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

func FocusBonusDamage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64) int {
	if attacker == nil || target == nil || attacker.HeroID != heroID || tick >= attacker.Archer.FocusActiveUntil {
		return 0
	}
	ratio := attacker.Archer.FocusBonusADRatio
	if ratio <= 0 {
		ratio = skillList(w.SkillConfig(qID), "bonusAdDamageRatio", attacker.Archer.FocusActiveLevel, []float64{1.05, 1.1, 1.15, 1.2, 1.25})
	}
	return w.ArcherPhysicalDamageAfterResistance(attacker, target, attacker.Stats.Attack*ratio, tick)
}

func ApplyFocusOnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.HeroID != heroID {
		return
	}
	if tick < attacker.Archer.FocusActiveUntil {
		damage := FocusBonusDamage(w, attacker, target, tick)
		if damage > 0 {
			previousDamage := target.Combat.LastDamage
			target.Combat.LastHitTick = tick
			if target.Kind != world.EntityKindDummy {
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
	AddFocusStack(w, attacker, tick, tickRate)
}

func archerFocusMaxStacks(skill config.SkillConfig) int {
	value := int(math.Round(skillMeta(skill, "maxStacks", 4)))
	if value < 1 {
		return 1
	}
	return value
}
