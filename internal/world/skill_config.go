package world

import (
	"l-battle/internal/config"
	"math"
)

func cooldownTicks(cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 {
		return 0
	}
	ticks := math.Ceil(float64(cooldownMS) / 1000 * float64(tickRate))
	return uint64(ticks)
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillMetaRange(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	value, ok := skill.Meta[key]
	if !ok {
		return fallback
	}
	return value
}

func skillMetaListByLevel(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if skill.MetaLists != nil && len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
	}
	rank := skillRank(level, len(values))
	return values[rank-1]
}

func skillMetaCurveByLevel(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	if skill.MetaLists == nil {
		return fallback
	}
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	currentLevel := float64(clampInt(level, MinHeroLevel, MaxHeroLevel))
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

func skillMetaListByLevelMS(skill config.SkillConfig, key string, level int, fallback []float64) int {
	return int(math.Round(skillMetaListByLevel(skill, key, level, fallback)))
}

func skillRank(level int, count int) int {
	return clampInt(level, 1, count)
}
