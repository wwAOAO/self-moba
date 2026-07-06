package warrior

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func ApplyE(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	durationTicks := secondsToTicks(skillMeta(skill, "durationSeconds", 3), tickRate)
	spins := warriorESpinCount(entity, skill)
	if durationTicks == 0 || spins <= 0 {
		return
	}
	intervalTicks := durationTicks / uint64(spins)
	if intervalTicks < 1 {
		intervalTicks = 1
	}
	entity.Warrior.JudgmentUntilTick = tick + durationTicks
	entity.Warrior.JudgmentNextSpinTick = tick
	entity.Warrior.JudgmentSpinIntervalTicks = intervalTicks
	entity.Warrior.JudgmentSpinsRemaining = spins
	entity.Warrior.JudgmentLevel = state.Level
	entity.Warrior.JudgmentHits = make(map[string]int)
	state.CooldownUntilTick = 0
	entity.Skills[eID] = state
}

func StopE(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || tick >= entity.Warrior.JudgmentUntilTick {
		return
	}
	remainingTicks := entity.Warrior.JudgmentUntilTick - tick
	cooldownTicks := cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000})), tickRate)
	if cooldownTicks > remainingTicks {
		state.CooldownUntilTick = tick + cooldownTicks - remainingTicks
	} else {
		state.CooldownUntilTick = tick
	}
	clearE(entity)
	entity.Skills[eID] = state
}

func TickE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Warrior.JudgmentUntilTick == 0 {
		return
	}
	skill := w.SkillConfig(eID)
	if tick >= entity.Warrior.JudgmentUntilTick {
		finishE(w, entity, skill, tick, tickRate)
		return
	}
	if entity.Warrior.JudgmentSpinsRemaining <= 0 || tick < entity.Warrior.JudgmentNextSpinTick || skill.SkillID == "" {
		return
	}
	applyESpin(w, entity, skill, tick, tickRate)
	entity.Warrior.JudgmentSpinsRemaining--
	if entity.Warrior.JudgmentSpinsRemaining <= 0 {
		entity.Warrior.JudgmentNextSpinTick = 0
		return
	}
	entity.Warrior.JudgmentNextSpinTick += entity.Warrior.JudgmentSpinIntervalTicks
}

func finishE(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	state := entity.Skills[eID]
	if skill.SkillID != "" {
		state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000})), tickRate)
		entity.Skills[eID] = state
	}
	clearE(entity)
}

func clearE(entity *world.Entity) {
	entity.Warrior.JudgmentUntilTick = 0
	entity.Warrior.JudgmentNextSpinTick = 0
	entity.Warrior.JudgmentSpinIntervalTicks = 0
	entity.Warrior.JudgmentSpinsRemaining = 0
	entity.Warrior.JudgmentLevel = 0
	entity.Warrior.JudgmentHits = nil
}

func applyESpin(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	targets := eTargets(w, entity, skill)
	if len(targets) == 0 {
		return
	}
	nearest := nearestEntity(entity, targets)
	for _, target := range targets {
		damage := eDamage(w, entity, target, skill, entity.Warrior.JudgmentLevel, target == nearest, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
			recordEHit(w, entity, target, skill, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
			recordEHit(w, entity, target, skill, tick, tickRate)
		}
	}
}

func eTargets(w *world.World, entity *world.Entity, skill config.SkillConfig) []*world.Entity {
	targets := []*world.Entity{}
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		if distance(entity.Position, target.Position) <= skillRange(skill, 180)+target.Radius {
			targets = append(targets, target)
		}
	})
	return targets
}

func nearestEntity(source *world.Entity, targets []*world.Entity) *world.Entity {
	var nearest *world.Entity
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

func eDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, nearest bool, tick uint64) int {
	if level <= 0 {
		level = 1
	}
	rawDamage := skillList(skill, "baseDamage", level, []float64{4, 7, 10, 13, 16}) + attacker.Stats.Attack*skillList(skill, "adRatio", level, []float64{0.4, 0.43, 0.46, 0.49, 0.52})
	if w.WarriorAttackCrits(attacker, target, tick) {
		rawDamage = skillList(skill, "critBaseDamage", level, []float64{5.2, 9.1, 13, 16.9, 20.8}) + attacker.Stats.Attack*skillList(skill, "critAdRatio", level, []float64{0.52, 0.559, 0.598, 0.637, 0.676})
	}
	if nearest {
		rawDamage *= 1 + skillMeta(skill, "nearestDamageBonus", 0.25)
	}
	return w.WarriorPhysicalDamageAfterResistance(attacker, target, rawDamage, tick)
}

func recordEHit(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if target == nil || target.Kind != world.EntityKindPlayer && target.Kind != world.EntityKindEnemyHero {
		return
	}
	if attacker.Warrior.JudgmentHits == nil {
		attacker.Warrior.JudgmentHits = make(map[string]int)
	}
	attacker.Warrior.JudgmentHits[target.ID]++
	if attacker.Warrior.JudgmentHits[target.ID] != int(skillMeta(skill, "armorShredHitCount", 6)) {
		return
	}
	world.ApplyWarriorPhysicalDefenseShred(w, target, skillMeta(skill, "armorShredPercent", 0.25), tick+secondsToTicks(skillMeta(skill, "armorShredSeconds", 6), tickRate))
}

func warriorESpinCount(entity *world.Entity, skill config.SkillConfig) int {
	baseSpins := int(skillMeta(skill, "baseSpins", 7))
	if entity == nil {
		return baseSpins
	}
	bonusPerSpin := skillMeta(skill, "attackSpeedBonusPerExtraSpin", 0.25)
	if bonusPerSpin <= 0 {
		return baseSpins
	}
	attackSpeedBonus := entity.Stats.AttackSpeedBonus
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	return baseSpins + int(math.Floor(attackSpeedBonus/bonusPerSpin))
}
