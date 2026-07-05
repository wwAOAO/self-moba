package world

func (w *World) applyMageIlluminationOnSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !canApplyMageIllumination(source, target) {
		return
	}
	skill := w.heroPassiveSkill(source)
	target.Control.MageIlluminationBy = source.ID
	target.Control.MageIlluminationUntil = tick + secondsToTicks(skillMetaRange(skill, "debuffSeconds", 6), tickRate)
}

func (w *World) applyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !canApplyMageIllumination(source, target) {
		return
	}
	w.detonateMageIllumination(source, target, tick, tickRate)
	w.applyMageIlluminationOnSkillHit(source, target, tick, tickRate)
}

func (w *World) triggerMageIlluminationOnBasicAttack(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != mageHeroID {
		return
	}
	w.detonateMageIllumination(source, target, tick, tickRate)
}

func canApplyMageIllumination(source *Entity, target *Entity) bool {
	if source == nil || target == nil || source.HeroID != mageHeroID {
		return false
	}
	if source.ID == target.ID || target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
		return false
	}
	return target.Team != source.Team || target.Team == TeamNeutral
}

func (w *World) detonateMageIllumination(source *Entity, target *Entity, tick uint64, tickRate int) bool {
	if !mageIlluminationActive(source, target, tick) {
		return false
	}
	target.Control.MageIlluminationBy = ""
	target.Control.MageIlluminationUntil = 0
	damage := w.mageIlluminationDamage(source, target, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
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

func mageIlluminationActive(source *Entity, target *Entity, tick uint64) bool {
	return source != nil &&
		target != nil &&
		target.Control.MageIlluminationBy == source.ID &&
		target.Control.MageIlluminationUntil > tick
}

func (w *World) mageIlluminationDamage(source *Entity, target *Entity, tick uint64) int {
	skill := w.heroPassiveSkill(source)
	baseDamage := skillMetaCurveByLevel(skill, "detonateDamage", "detonateDamageLevels", source.Level, 20)
	rawDamage := baseDamage + float64(source.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.3)
	return magicDamageAfterResistance(source, target, rawDamage, tick)
}
