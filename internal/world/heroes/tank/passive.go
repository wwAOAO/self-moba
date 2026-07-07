package tank

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func PassiveState(w *world.World, hero config.HeroConfig) world.PassiveState {
	if hero.HeroID != heroID {
		return world.PassiveState{}
	}
	shield := graniteShieldValue(hero.Base.HP, w.SkillConfig(hero.Skills.Passive))
	return world.PassiveState{Shield: shield, MaxShield: shield}
}

func TickGranite(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Stats.HP <= 0 {
		return
	}
	maxShield := graniteShieldValue(entity.Stats.MaxHP, w.HeroPassiveSkill(entity))
	if maxShield <= 0 || entity.Passive.Shield >= maxShield {
		entity.Passive.MaxShield = maxShield
		return
	}
	resetTicks := secondsToTicks(skillMeta(w.HeroPassiveSkill(entity), "resetSeconds", 10), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+resetTicks {
		return
	}
	entity.Passive.MaxShield = maxShield
	entity.Passive.Shield = maxShield
}

func RefreshGranite(w *world.World, entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	shield := graniteShieldValue(entity.Stats.MaxHP, w.HeroPassiveSkill(entity))
	entity.Passive.MaxShield = shield
	entity.Passive.Shield = shield
	entity.Passive.ShieldExpireTick = 0
}

func RefreshGraniteMax(w *world.World, entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	oldMax := entity.Passive.MaxShield
	nextMax := graniteShieldValue(entity.Stats.MaxHP, w.HeroPassiveSkill(entity))
	entity.Passive.MaxShield = nextMax
	if entity.Passive.Shield >= oldMax {
		entity.Passive.Shield = nextMax
	}
	if entity.Passive.Shield > nextMax {
		entity.Passive.Shield = nextMax
	}
}

func graniteShieldValue(maxHP float64, skill config.SkillConfig) int {
	return int(math.Round(maxHP * skillMeta(skill, "shieldMaxHPRatio", 0.1)))
}
