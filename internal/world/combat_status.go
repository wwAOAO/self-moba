package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) applyArcherFrostShot(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != archerHeroID {
		return
	}
	skill := w.heroPassiveSkill(source)
	slow := archerFrostSlowRatio(source.Level, skill)
	if w.attackCrits(source, target, tick) {
		slow *= skillMetaRange(skill, "critSlowMultiplier", 2)
	}
	duration := secondsToTicks(skillMetaRange(skill, "slowSeconds", 2), tickRate)
	applyMoveSpeedSlow(target, slow, tick+duration)
}

func reduceCritDamage(target *Entity, damage int, crit bool) int {
	if target == nil || !crit || target.Stats.CritDamageReduce <= 0 {
		return damage
	}
	reduced := int(math.Round(float64(damage) * (1 - clamp(target.Stats.CritDamageReduce, 0, 1))))
	if reduced < 1 {
		return 1
	}
	return reduced
}

func archerFrostSlowRatio(level int, skill config.SkillConfig) float64 {
	minSlow := skillMetaRange(skill, "slowMin", 0.2)
	maxSlow := skillMetaRange(skill, "slowMax", 0.3)
	level = clampInt(level, MinHeroLevel, MaxHeroLevel)
	if MaxHeroLevel <= MinHeroLevel {
		return maxSlow
	}
	progress := float64(level-MinHeroLevel) / float64(MaxHeroLevel-MinHeroLevel)
	return minSlow + (maxSlow-minSlow)*progress
}

func applyMoveSpeedSlow(target *Entity, slow float64, until uint64) {
	if target == nil || slow <= 0 || until == 0 {
		return
	}
	slow = clamp(slow*(1-clamp(target.Stats.SlowResist, 0, 1)), 0, 1)
	if until < target.Control.MoveSpeedSlowUntil && slow <= target.Control.MoveSpeedSlow {
		return
	}
	target.Control.MoveSpeedSlow = slow
	target.Control.MoveSpeedSlowUntil = until
}

func (w *World) skillConfig(skillID string) config.SkillConfig {
	if w == nil || w.skills == nil || skillID == "" {
		return config.SkillConfig{}
	}
	skill, _ := w.skills.Get(skillID)
	return skill
}

func (w *World) heroPassiveSkill(entity *Entity) config.SkillConfig {
	if entity == nil || w == nil || w.heroes == nil {
		return config.SkillConfig{}
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return config.SkillConfig{}
	}
	return w.skillConfig(hero.Skills.Passive)
}

func (w *World) breakWarriorToughness(source *Entity, target *Entity, tick uint64) {
	if target == nil || target.HeroID != warriorHeroID || !warriorToughnessBreaksRegen(source) {
		return
	}
	target.Passive.LastRegenBreakTick = tick
	target.Passive.NextRegenTick = 0
}

func (w *World) breakTankGraniteShield(target *Entity, tick uint64) {
	if target == nil || target.HeroID != tankHeroID {
		return
	}
	target.Passive.LastRegenBreakTick = tick
	target.Passive.NextRegenTick = 0
}

func warriorToughnessBreaksRegen(source *Entity) bool {
	if source == nil {
		return false
	}
	switch source.Kind {
	case EntityKindPlayer, EntityKindEnemyHero, EntityKindTower:
		return true
	default:
		return false
	}
}
