package mage

import "l-battle/internal/world"

func ApplyIllumination(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if !canApplyIllumination(source, target) {
		return
	}
	skill := w.MagePassiveSkill(source)
	target.Control.MageIlluminationBy = source.ID
	target.Control.MageIlluminationUntil = tick + secondsToTicks(skillMeta(skill, "debuffSeconds", 6), tickRate)
}

func ApplyUltimateIllumination(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if !canApplyIllumination(source, target) {
		return
	}
	DetonateIllumination(w, source, target, tick, tickRate)
	ApplyIllumination(w, source, target, tick, tickRate)
}

func TriggerIllumination(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID {
		return
	}
	DetonateIllumination(w, source, target, tick, tickRate)
}

func canApplyIllumination(source *world.Entity, target *world.Entity) bool {
	if source == nil || target == nil || source.HeroID != heroID {
		return false
	}
	if source.ID == target.ID || target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == world.EntityKindPlayer && target.Death.Dead {
		return false
	}
	return target.Team != source.Team || target.Team == world.TeamNeutral
}

func DetonateIllumination(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) bool {
	if !illuminationActive(source, target, tick) {
		return false
	}
	target.Control.MageIlluminationBy = ""
	target.Control.MageIlluminationUntil = 0
	damage := illuminationDamage(w, source, target, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != world.EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	}
	return true
}

func illuminationActive(source *world.Entity, target *world.Entity, tick uint64) bool {
	return source != nil &&
		target != nil &&
		target.Control.MageIlluminationBy == source.ID &&
		target.Control.MageIlluminationUntil > tick
}

func illuminationDamage(w *world.World, source *world.Entity, target *world.Entity, tick uint64) int {
	skill := w.MagePassiveSkill(source)
	baseDamage := world.MageSkillMetaCurveByLevel(skill, "detonateDamage", "detonateDamageLevels", source.Level, 20)
	rawDamage := baseDamage + float64(source.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.3)
	return w.MageMagicDamageAfterResistance(source, target, rawDamage, tick)
}
