package sword

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func CritChanceMultiplier(w *world.World, entity *world.Entity) float64 {
	if entity == nil || entity.HeroID != heroID {
		return 1
	}
	return skillMeta(w.HeroPassiveSkill(entity), "critChanceMultiplier", 2)
}

func ApplyCritFinalMultiplier(w *world.World, attacker *world.Entity, damage int, crit bool) int {
	if !crit || attacker == nil || attacker.HeroID != heroID || damage <= 0 {
		return damage
	}
	result := int(math.Round(float64(damage) * skillMeta(w.HeroPassiveSkill(attacker), "critFinalDamageMultiplier", 0.9)))
	if result < 1 {
		return 1
	}
	return result
}

func ApplyShield(w *world.World, source *world.Entity, target *world.Entity, tickRate int) {
	if target == nil || target.HeroID != heroID || target.Passive.Shield > 0 || target.Passive.SwordIntent < target.Passive.MaxSwordIntent || !shieldTriggers(source) {
		return
	}
	skill := w.HeroPassiveSkill(target)
	target.Passive.MaxShield = ShieldValue(w, target)
	target.Passive.Shield = target.Passive.MaxShield
	target.Passive.ShieldExpireTick = target.Combat.LastHitTick + secondsToTicks(skillMeta(skill, "shieldDurationSeconds", 1), tickRate)
	target.Passive.SwordIntent = 0
}

func ShieldValue(w *world.World, entity *world.Entity) int {
	if entity == nil {
		return 0
	}
	level := entity.Level
	if level < world.MinHeroLevel {
		level = world.MinHeroLevel
	}
	if level > world.MaxHeroLevel {
		level = world.MaxHeroLevel
	}
	return int(math.Round(skillMetaCurve(w.HeroPassiveSkill(entity), "shieldValue", "shieldValueLevels", level, 125)))
}

func shieldTriggers(source *world.Entity) bool {
	if source == nil {
		return false
	}
	return source.Kind == world.EntityKindPlayer || source.Kind == world.EntityKindEnemyHero
}

func PassiveState(w *world.World, hero config.HeroConfig) world.PassiveState {
	if hero.HeroID != heroID {
		return world.PassiveState{}
	}
	return world.PassiveState{MaxSwordIntent: skillMeta(w.SkillConfig(hero.Skills.Passive), "intentMax", 100)}
}

func StateForHero(id string) world.SwordState {
	if id != heroID {
		return world.SwordState{}
	}
	return world.SwordState{SweepingBladeTargetUntil: make(map[string]uint64)}
}

func ChargeIntent(w *world.World, entity *world.Entity, moved float64) {
	if entity == nil || entity.HeroID != heroID || moved <= 0 {
		return
	}
	skill := w.HeroPassiveSkill(entity)
	if entity.Passive.MaxSwordIntent <= 0 {
		entity.Passive.MaxSwordIntent = skillMeta(skill, "intentMax", 100)
	}
	if entity.Passive.SwordIntent >= entity.Passive.MaxSwordIntent {
		return
	}
	moveUnitsPerPercent := skillMetaCurve(skill, "intentMoveUnitsPerPercent", "intentMoveUnitLevels", entity.Level, 59)
	if moveUnitsPerPercent <= 0 {
		moveUnitsPerPercent = 59
	}
	entity.Passive.SwordIntent += moved / moveUnitsPerPercent
	if entity.Passive.SwordIntent > entity.Passive.MaxSwordIntent {
		entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	}
}

func TickShield(entity *world.Entity, tick uint64) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.Shield <= 0 || entity.Passive.ShieldExpireTick == 0 || tick < entity.Passive.ShieldExpireTick {
		return
	}
	entity.Passive.Shield = 0
	entity.Passive.MaxShield = 0
	entity.Passive.ShieldExpireTick = 0
}

func ApplyCritOverflowStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || stats == nil || entity.HeroID != heroID {
		return
	}
	skill := w.HeroPassiveSkill(entity)
	effectiveCrit := stats.CritChance * skillMeta(skill, "critChanceMultiplier", 2)
	if effectiveCrit <= 1 {
		return
	}
	bonusAttack := (effectiveCrit - 1) * 100 * skillMeta(skill, "critOverflowAttackPerPercent", 0.5)
	stats.Attack += bonusAttack
	stats.BonusAttack += bonusAttack
}
