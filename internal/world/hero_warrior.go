package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

func (w *World) applyWarriorQ(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.DecisiveStrikeUntilTick = tick + secondsToTicks(skillMetaRange(skill, "empowerDurationSeconds", 4.5), tickRate)
	entity.Warrior.DecisiveStrikeSpeedUntilTick = tick + secondsToTicks(skillMetaListByLevel(skill, "moveSpeedDurationSeconds", state.Level, []float64{1.5, 2, 2.5, 3, 3.5}), tickRate)
	entity.Warrior.DecisiveStrikeLevel = state.Level
	entity.Warrior.DecisiveStrikeMoveSpeedBonus = skillMetaRange(skill, "moveSpeedBonus", 0.3)
	entity.Combat.NextAttackTick = tick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000}), tickRate)
	entity.Skills[warriorQSkillID] = state
}

func (w *World) applyWarriorW(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.CourageUntilTick = tick + secondsToTicks(skillMetaRange(skill, "durationSeconds", 4), tickRate)
	entity.Warrior.CourageFrontUntilTick = tick + secondsToTicks(skillMetaRange(skill, "frontDurationSeconds", 0.75), tickRate)
	entity.Warrior.CourageFrontDamageReduce = skillMetaRange(skill, "frontDamageReduce", 0.6)
	entity.Warrior.CourageFrontTenacity = skillMetaRange(skill, "frontTenacity", 0.6)
	entity.Warrior.CourageBackDamageReduce = skillMetaRange(skill, "backDamageReduce", 0.3)
	entity.Control.TenacityUntilTick = entity.Warrior.CourageFrontUntilTick
	entity.Passive.MaxShield = warriorWShieldValue(entity, skill, state.Level)
	entity.Passive.Shield = entity.Passive.MaxShield
	entity.Passive.ShieldExpireTick = entity.Warrior.CourageUntilTick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{24000, 22000, 20000, 18000, 16000}), tickRate)
	entity.Skills[warriorWSkillID] = state
}

func (w *World) applyWarriorE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	durationTicks := secondsToTicks(skillMetaRange(skill, "durationSeconds", 3), tickRate)
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
	entity.Skills[warriorESkillID] = state
}

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
	skill := w.skillConfig(warriorESkillID)
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
	w.applyWarriorESpin(entity, skill, tick, tickRate)
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

func (w *World) applyWarriorESpin(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
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
			w.applyDamage(entity, target, damage, tickRate)
			w.recordWarriorEHit(entity, target, skill, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
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
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) <= skillRange(skill, 180)+target.Radius {
			targets = append(targets, target)
		}
	}
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

func warriorESpinCount(entity *Entity, skill config.SkillConfig) int {
	baseSpins := int(skillMetaRange(skill, "baseSpins", 7))
	if entity == nil {
		return baseSpins
	}
	bonusPerSpin := skillMetaRange(skill, "attackSpeedBonusPerExtraSpin", 0.25)
	if bonusPerSpin <= 0 {
		return baseSpins
	}
	attackSpeedBonus := entity.Stats.AttackSpeedBonus
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	return baseSpins + int(math.Floor(attackSpeedBonus/bonusPerSpin))
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
	w.applyPhysicalDefenseShred(target, skillMetaRange(skill, "armorShredPercent", 0.25), tick+secondsToTicks(skillMetaRange(skill, "armorShredSeconds", 6), tickRate))
}

func (w *World) applyPhysicalDefenseShred(target *Entity, percent float64, untilTick uint64) {
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

func (w *World) applyWarriorR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	if entity.Warrior.JusticePending {
		return
	}
	target := w.warriorRTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.435), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Warrior.JusticePending = true
	entity.Warrior.JusticeReleaseTick = tick + windupTicks
	entity.Warrior.JusticeTargetID = target.ID
	entity.Warrior.JusticeLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Warrior.JusticeReleaseTick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{120000, 100000, 80000}), tickRate)
	entity.Skills[warriorRSkillID] = state
}

func (w *World) releaseWarriorR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || !entity.Warrior.JusticePending || tick < entity.Warrior.JusticeReleaseTick {
		return
	}
	target := w.entities[entity.Warrior.JusticeTargetID]
	level := entity.Warrior.JusticeLevel
	entity.Warrior.JusticePending = false
	entity.Warrior.JusticeReleaseTick = 0
	entity.Warrior.JusticeTargetID = ""
	entity.Warrior.JusticeLevel = 0
	if !canAttackTarget(entity, target) {
		return
	}
	skill := w.skillConfig(warriorRSkillID)
	damage := warriorRDamage(target, skill, level)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyTrueDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(entity, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = trueDamageAfterReduction(target, damage, tick)
		target.Combat.LastDamageType = "true"
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) warriorRTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 400)
	pickPadding := skillMetaRange(skill, "targetPickPadding", 80)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) > castRange+target.Radius {
			continue
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+pickPadding {
			continue
		}
		if distToPoint < bestDistance {
			bestDistance = distToPoint
			best = target
		}
	}
	return best
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
