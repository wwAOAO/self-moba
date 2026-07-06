package blade

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func GainBasicAttackRage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64) {
	if attacker == nil || target == nil || attacker.HeroID != heroID {
		return
	}
	skill := w.HeroPassiveSkill(attacker)
	gain := skillMeta(skill, "basicAttackRage", 5)
	if w.AttackCrits(attacker, target, tick) {
		gain += skillMeta(skill, "critAttackRage", 5)
	}
	gainRage(attacker, gain, tick)
}

func GainKillRage(w *world.World, killer *world.Entity) {
	if killer == nil || killer.HeroID != heroID {
		return
	}
	gainRage(killer, skillMeta(w.HeroPassiveSkill(killer), "killRage", 10), 0)
}

func GainSkillHitRage(w *world.World, source *world.Entity, tick uint64) {
	if source == nil || source.HeroID != heroID {
		return
	}
	gainRage(source, skillMeta(w.HeroPassiveSkill(source), "skillHitRage", 5), tick)
}

func gainRage(entity *world.Entity, amount float64, tick uint64) {
	if entity == nil || entity.HeroID != heroID || amount <= 0 {
		return
	}
	maxRage := maxRage(entity)
	if maxRage <= 0 {
		return
	}
	entity.Stats.MP = math.Min(entity.Stats.MP+amount, maxRage)
	if tick > 0 {
		entity.Combat.LastHitTick = tick
	}
}

func TickRageDecay(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Stats.HP <= 0 || entity.Stats.MP <= 0 || tickRate <= 0 {
		return
	}
	skill := w.HeroPassiveSkill(entity)
	outOfCombatTicks := secondsToTicks(skillMeta(skill, "outOfCombatSeconds", 5), tickRate)
	if tick < entity.Combat.LastHitTick+outOfCombatTicks {
		return
	}
	entity.Stats.MP = math.Max(0, entity.Stats.MP-skillMeta(skill, "rageDecayPerSecond", 5)/float64(tickRate))
}

func RageCritChance(entity *world.Entity, skill config.SkillConfig) float64 {
	if entity == nil || entity.HeroID != heroID {
		return 0
	}
	rage := math.Max(0, math.Min(entity.Stats.MP, maxRage(entity)))
	return rage * skillMeta(skill, "critChancePerRage", 0.0035)
}

func maxRage(entity *world.Entity) float64 {
	if entity == nil {
		return 0
	}
	if entity.Stats.MaxMP > 0 {
		return entity.Stats.MaxMP
	}
	return 100
}
