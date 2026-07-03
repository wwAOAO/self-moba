package world

import "l-battle/internal/config"

func (w *World) gainBladeBasicAttackRage(attacker *Entity, target *Entity, tick uint64) {
	if attacker == nil || target == nil || attacker.HeroID != bladeHeroID {
		return
	}
	skill := w.heroPassiveSkill(attacker)
	gain := skillMetaRange(skill, "basicAttackRage", 5)
	if w.attackCrits(attacker, target, tick) {
		gain += skillMetaRange(skill, "critAttackRage", 5)
	}
	w.gainBladeRage(attacker, gain, tick)
}

func (w *World) gainBladeKillRage(killer *Entity) {
	if killer == nil || killer.HeroID != bladeHeroID {
		return
	}
	w.gainBladeRage(killer, skillMetaRange(w.heroPassiveSkill(killer), "killRage", 10), 0)
}

func (w *World) gainBladeSkillHitRage(source *Entity, tick uint64) {
	if source == nil || source.HeroID != bladeHeroID {
		return
	}
	w.gainBladeRage(source, skillMetaRange(w.heroPassiveSkill(source), "skillHitRage", 5), tick)
}

func (w *World) gainBladeRage(entity *Entity, amount float64, tick uint64) {
	if entity == nil || entity.HeroID != bladeHeroID || amount <= 0 {
		return
	}
	maxRage := bladeMaxRage(entity)
	if maxRage <= 0 {
		return
	}
	entity.Stats.MP += amount
	if entity.Stats.MP > maxRage {
		entity.Stats.MP = maxRage
	}
	if tick > 0 {
		entity.Combat.LastHitTick = tick
	}
}

func (w *World) tickBladeRageDecay(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != bladeHeroID || entity.Stats.HP <= 0 || entity.Stats.MP <= 0 || tickRate <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	outOfCombatTicks := secondsToTicks(skillMetaRange(skill, "outOfCombatSeconds", 5), tickRate)
	if tick < entity.Combat.LastHitTick+outOfCombatTicks {
		return
	}
	entity.Stats.MP -= skillMetaRange(skill, "rageDecayPerSecond", 5) / float64(tickRate)
	if entity.Stats.MP < 0 {
		entity.Stats.MP = 0
	}
}

func bladeMaxRage(entity *Entity) float64 {
	if entity == nil {
		return 0
	}
	if entity.Stats.MaxMP > 0 {
		return entity.Stats.MaxMP
	}
	return 100
}

func bladeRageCritChance(entity *Entity, skill config.SkillConfig) float64 {
	if entity == nil || entity.HeroID != bladeHeroID {
		return 0
	}
	rage := entity.Stats.MP
	maxRage := bladeMaxRage(entity)
	if rage > maxRage {
		rage = maxRage
	}
	if rage < 0 {
		rage = 0
	}
	return rage * skillMetaRange(skill, "critChancePerRage", 0.0035)
}
