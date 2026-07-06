package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
)

func ApplyFrostShot(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID {
		return
	}
	skill := w.ArcherPassiveSkill(source)
	slow := frostSlowRatio(source.Level, skill)
	if w.ArcherAttackCrits(source, target, tick) {
		slow *= skillMeta(skill, "critSlowMultiplier", 2)
	}
	duration := secondsToTicks(skillMeta(skill, "slowSeconds", 2), tickRate)
	w.ApplyArcherMoveSpeedSlow(target, slow, tick+duration)
}

func frostSlowRatio(level int, skill config.SkillConfig) float64 {
	minSlow := skillMeta(skill, "slowMin", 0.2)
	maxSlow := skillMeta(skill, "slowMax", 0.3)
	level = clampInt(level, world.MinHeroLevel, world.MaxHeroLevel)
	if world.MaxHeroLevel <= world.MinHeroLevel {
		return maxSlow
	}
	progress := float64(level-world.MinHeroLevel) / float64(world.MaxHeroLevel-world.MinHeroLevel)
	return minSlow + (maxSlow-minSlow)*progress
}
