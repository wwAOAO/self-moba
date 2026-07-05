package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) stopWarriorE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || tick >= entity.Warrior.JudgmentUntilTick {
		return
	}
	remainingTicks := entity.Warrior.JudgmentUntilTick - tick
	cooldownTicks := cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000}), tickRate)
	if cooldownTicks > remainingTicks {
		state.CooldownUntilTick = tick + cooldownTicks - remainingTicks
	} else {
		state.CooldownUntilTick = tick
	}
	clearWarriorE(entity)
	entity.Skills[warriorESkillID] = state
}

func (w *World) tickWarriorJudgment(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || entity.Warrior.JudgmentUntilTick == 0 {
		return
	}
	skill := w.SkillConfig(warriorESkillID)
	if tick >= entity.Warrior.JudgmentUntilTick {
		w.finishWarriorE(entity, skill, tick, tickRate)
		return
	}
	if entity.Warrior.JudgmentSpinsRemaining <= 0 {
		return
	}
	if tick < entity.Warrior.JudgmentNextSpinTick {
		return
	}
	if skill.SkillID == "" {
		return
	}
	applyWarriorESpin(w, entity, skill, tick, tickRate)
	entity.Warrior.JudgmentSpinsRemaining--
	if entity.Warrior.JudgmentSpinsRemaining <= 0 {
		entity.Warrior.JudgmentNextSpinTick = 0
		return
	}
	entity.Warrior.JudgmentNextSpinTick += entity.Warrior.JudgmentSpinIntervalTicks
}

func (w *World) finishWarriorE(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	state := entity.Skills[warriorESkillID]
	if skill.SkillID != "" {
		state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000}), tickRate)
		entity.Skills[warriorESkillID] = state
	}
	clearWarriorE(entity)
}

func clearWarriorE(entity *Entity) {
	entity.Warrior.JudgmentUntilTick = 0
	entity.Warrior.JudgmentNextSpinTick = 0
	entity.Warrior.JudgmentSpinIntervalTicks = 0
	entity.Warrior.JudgmentSpinsRemaining = 0
	entity.Warrior.JudgmentLevel = 0
	entity.Warrior.JudgmentHits = nil
}

func applyWarriorESpin(w *World, entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	targets := w.warriorETargets(entity, skill)
	if len(targets) == 0 {
		return
	}
	nearest := nearestEntity(entity, targets)
	for _, target := range targets {
		damage := w.warriorEDamage(entity, target, skill, entity.Warrior.JudgmentLevel, target == nearest, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
			w.recordWarriorEHit(entity, target, skill, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
			w.recordWarriorEHit(entity, target, skill, tick, tickRate)
		}
	}
}

func (w *World) warriorETargets(entity *Entity, skill config.SkillConfig) []*Entity {
	targets := []*Entity{}
	w.ForEachEntity(func(target *Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		if distance(entity.Position, target.Position) <= skillRange(skill, 180)+target.Radius {
			targets = append(targets, target)
		}
	})
	return targets
}

func nearestEntity(source *Entity, targets []*Entity) *Entity {
	var nearest *Entity
	nearestDistance := math.MaxFloat64
	for _, target := range targets {
		dist := distance(source.Position, target.Position)
		if dist < nearestDistance {
			nearestDistance = dist
			nearest = target
		}
	}
	return nearest
}

func (w *World) warriorEDamage(attacker *Entity, target *Entity, skill config.SkillConfig, level int, nearest bool, tick uint64) int {
	if level <= 0 {
		level = 1
	}
	rawDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{4, 7, 10, 13, 16}) + attacker.Stats.Attack*skillMetaListByLevel(skill, "adRatio", level, []float64{0.4, 0.43, 0.46, 0.49, 0.52})
	if w.attackCrits(attacker, target, tick) {
		rawDamage = skillMetaListByLevel(skill, "critBaseDamage", level, []float64{5.2, 9.1, 13, 16.9, 20.8}) + attacker.Stats.Attack*skillMetaListByLevel(skill, "critAdRatio", level, []float64{0.52, 0.559, 0.598, 0.637, 0.676})
	}
	if nearest {
		rawDamage *= 1 + skillMetaRange(skill, "nearestDamageBonus", 0.25)
	}
	return physicalDamageAfterResistance(attacker, target, rawDamage, tick)
}

func (w *World) recordWarriorEHit(attacker *Entity, target *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer && target.Kind != EntityKindEnemyHero {
		return
	}
	if attacker.Warrior.JudgmentHits == nil {
		attacker.Warrior.JudgmentHits = make(map[string]int)
	}
	attacker.Warrior.JudgmentHits[target.ID]++
	if attacker.Warrior.JudgmentHits[target.ID] != int(skillMetaRange(skill, "armorShredHitCount", 6)) {
		return
	}
	applyPhysicalDefenseShred(w, target, skillMetaRange(skill, "armorShredPercent", 0.25), tick+secondsToTicks(skillMetaRange(skill, "armorShredSeconds", 6), tickRate))
}

func applyPhysicalDefenseShred(w *World, target *Entity, percent float64, untilTick uint64) {
	if target == nil || percent <= 0 {
		return
	}
	if target.Combat.PhysicalDefenseShredAmount > 0 {
		target.Stats.PhysicalDefense += target.Combat.PhysicalDefenseShredAmount
	}
	shred := target.Stats.PhysicalDefense * clamp(percent, 0, 1)
	target.Stats.PhysicalDefense -= shred
	if target.Stats.PhysicalDefense < 0 {
		target.Stats.PhysicalDefense = 0
	}
	target.Combat.PhysicalDefenseShredAmount = shred
	target.Combat.PhysicalDefenseShredUntil = untilTick
}
