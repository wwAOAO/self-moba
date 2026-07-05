package world

import "l-battle/internal/config"

func applyTankE(w *World, entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	if entity.Tank.GroundSlamPending {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.242), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Tank.GroundSlamPending = true
	entity.Tank.GroundSlamReleaseTick = tick + windupTicks
	entity.Tank.GroundSlamLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Tank.GroundSlamReleaseTick
	entity.Skills[tankESkillID] = state
}

func (w *World) releaseTankE(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.GroundSlamPending || tick < entity.Tank.GroundSlamReleaseTick {
		return
	}
	level := entity.Tank.GroundSlamLevel
	entity.Tank.GroundSlamPending = false
	entity.Tank.GroundSlamReleaseTick = 0
	entity.Tank.GroundSlamLevel = 0
	if level <= 0 {
		level = 1
	}
	skill := w.SkillConfig(tankESkillID)
	state := entity.Skills[tankESkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{7000, 7000, 7000, 7000, 7000}), tickRate)
	entity.Skills[tankESkillID] = state
	damage := tankEDamage(entity, skill, level)
	slow := skillMetaListByLevel(skill, "attackSpeedSlow", level, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	slowUntil := tick + secondsToTicks(skillMetaRange(skill, "attackSpeedSlowSeconds", 3), tickRate)
	for _, target := range w.targetsInRadius(entity, entity.Position, skillRange(skill, 400)) {
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, magicDamageAfterResistance(entity, target, damage, tick), "magic", tickRate)
			applyAttackSpeedSlow(target, slow, slowUntil)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			applyAttackSpeedSlow(target, slow, slowUntil)
		}
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func tankEDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{60, 95, 130, 165, 200}) +
		entity.Stats.PhysicalDefense*skillMetaRange(skill, "armorRatio", 0.4) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
}

func applyAttackSpeedSlow(target *Entity, slow float64, until uint64) {
	if target == nil || slow <= 0 || until == 0 {
		return
	}
	target.Control.AttackSpeedSlow = clamp(slow, 0, 1)
	target.Control.AttackSpeedSlowUntil = until
}
