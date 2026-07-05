package world

import "l-battle/internal/config"

func (w *World) releaseWarriorR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || !entity.Warrior.JusticePending || tick < entity.Warrior.JusticeReleaseTick {
		return
	}
	target := w.EntityByID(entity.Warrior.JusticeTargetID)
	level := entity.Warrior.JusticeLevel
	entity.Warrior.JusticePending = false
	entity.Warrior.JusticeReleaseTick = 0
	entity.Warrior.JusticeTargetID = ""
	entity.Warrior.JusticeLevel = 0
	if !canAttackTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(warriorRSkillID)
	state := entity.Skills[warriorRSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{120000, 100000, 80000}), tickRate)
	entity.Skills[warriorRSkillID] = state
	damage := warriorRDamage(target, skill, level)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.ApplyTrueDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = trueDamageAfterReduction(target, damage, tick)
		target.Combat.LastDamageType = "true"
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func warriorRDamage(target *Entity, skill config.SkillConfig, level int) float64 {
	if target == nil {
		return 0
	}
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{150, 250, 350})
	missingHPRatio := skillMetaListByLevel(skill, "missingHPRatio", level, []float64{0.25, 0.3, 0.35})
	missingHP := target.Stats.MaxHP - target.Stats.HP
	if missingHP < 0 {
		missingHP = 0
	}
	return baseDamage + float64(missingHP)*missingHPRatio
}
