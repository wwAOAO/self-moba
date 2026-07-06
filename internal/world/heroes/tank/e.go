package tank

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Tank.GroundSlamPending {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.242), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Tank.GroundSlamPending = true
	entity.Tank.GroundSlamReleaseTick = tick + windupTicks
	entity.Tank.GroundSlamLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Tank.GroundSlamReleaseTick
	entity.Skills[eID] = state
}

func ReleaseE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Tank.GroundSlamPending || tick < entity.Tank.GroundSlamReleaseTick {
		return
	}
	level := entity.Tank.GroundSlamLevel
	entity.Tank.GroundSlamPending = false
	entity.Tank.GroundSlamReleaseTick = 0
	entity.Tank.GroundSlamLevel = 0
	if level <= 0 {
		level = 1
	}
	skill := w.SkillConfig(eID)
	state := entity.Skills[eID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{7000, 7000, 7000, 7000, 7000})), tickRate)
	entity.Skills[eID] = state
	damage := eDamage(entity, skill, level)
	slow := skillList(skill, "attackSpeedSlow", level, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	slowUntil := tick + secondsToTicks(skillMeta(skill, "attackSpeedSlowSeconds", 3), tickRate)
	for _, target := range w.TankTargetsInRadius(entity, entity.Position, skillRange(skill, 400)) {
		target.Combat.LastHitTick = tick
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, w.TankMagicDamageAfterResistance(entity, target, damage, tick), "magic", tickRate)
			applyAttackSpeedSlow(target, slow, slowUntil)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = w.TankMagicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			applyAttackSpeedSlow(target, slow, slowUntil)
		}
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func eDamage(entity *world.Entity, skill config.SkillConfig, level int) float64 {
	return skillList(skill, "baseDamage", level, []float64{60, 95, 130, 165, 200}) +
		entity.Stats.PhysicalDefense*skillMeta(skill, "armorRatio", 0.4) +
		float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.6)
}

func applyAttackSpeedSlow(target *world.Entity, slow float64, until uint64) {
	if target == nil || slow <= 0 || until == 0 {
		return
	}
	target.Control.AttackSpeedSlow = math.Max(0, math.Min(slow, 1))
	target.Control.AttackSpeedSlowUntil = until
}
