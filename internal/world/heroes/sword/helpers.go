package sword

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func canAttackTarget(attacker *world.Entity, target *world.Entity) bool {
	return attacker != nil && target != nil && attacker.Team != target.Team && target.Stats.HP > 0
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}

func cooldownTicks(cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 {
		return 0
	}
	return secondsToTicks(float64(cooldownMS)/1000, tickRate)
}

func skillMetaList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if skill.MetaLists != nil && len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
	}
	if level < 1 {
		level = 1
	}
	if level > len(values) {
		level = len(values)
	}
	return values[level-1]
}

func skillMetaListMS(skill config.SkillConfig, key string, level int, fallback []float64) int {
	return int(math.Round(skillMetaList(skill, key, level, fallback)))
}

func skillMetaCurve(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	if skill.MetaLists == nil {
		return fallback
	}
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	currentLevel := float64(level)
	if currentLevel <= levels[0] {
		return values[0]
	}
	last := len(values) - 1
	if currentLevel >= levels[last] {
		return values[last]
	}
	for i := 1; i < len(values); i++ {
		if currentLevel > levels[i] {
			continue
		}
		fromLevel := levels[i-1]
		toLevel := levels[i]
		if toLevel <= fromLevel {
			return values[i]
		}
		t := (currentLevel - fromLevel) / (toLevel - fromLevel)
		return values[i-1] + (values[i]-values[i-1])*t
	}
	return values[last]
}
