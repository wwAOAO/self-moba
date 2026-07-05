package world

import (
	"l-battle/internal/config"
	"l-battle/internal/world/formula"
	"l-battle/internal/world/statcalc"
	"math"
)

func normalize(x float64, y float64) (float64, float64) {
	return formula.Normalize(x, y)
}

func clamp(value float64, min float64, max float64) float64 {
	return formula.Clamp(value, min, max)
}

func heroStatsAtLevel(hero config.HeroConfig, level int) Stats {
	return statcalc.HeroStatsAtLevel(hero, level, MinHeroLevel, MaxHeroLevel)
}

func finalAttackSpeed(baseAttackSpeed float64, attackSpeedBonus float64, attackSpeedRatio float64, attackSpeedSlow float64) float64 {
	return formula.FinalAttackSpeed(baseAttackSpeed, attackSpeedBonus, attackSpeedRatio, attackSpeedSlow)
}

func (w *World) passiveStateForHero(hero config.HeroConfig) PassiveState {
	if hero.HeroID == swordHeroID {
		skill := w.skillConfig(hero.Skills.Passive)
		return PassiveState{
			MaxSwordIntent: skillMetaRange(skill, "intentMax", 100),
		}
	}
	if hero.HeroID == tankHeroID {
		stats := heroStatsAtLevel(hero, MinHeroLevel)
		shield := tankGraniteShieldValue(stats.MaxHP, w.skillConfig(hero.Skills.Passive))
		return PassiveState{
			Shield:    shield,
			MaxShield: shield,
		}
	}
	return PassiveState{}
}

func swordStateForHero(heroID string) SwordState {
	if heroID != swordHeroID {
		return SwordState{}
	}
	return SwordState{
		SweepingBladeTargetUntil: make(map[string]uint64),
	}
}

func (w *World) chargeSwordIntent(entity *Entity, moved float64) {
	if entity == nil || entity.HeroID != swordHeroID || moved <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	if entity.Passive.MaxSwordIntent <= 0 {
		entity.Passive.MaxSwordIntent = skillMetaRange(skill, "intentMax", 100)
	}
	if entity.Passive.SwordIntent >= entity.Passive.MaxSwordIntent {
		return
	}
	moveUnitsPerPercent := skillMetaCurveByLevel(skill, "intentMoveUnitsPerPercent", "intentMoveUnitLevels", entity.Level, 59)
	if moveUnitsPerPercent <= 0 {
		moveUnitsPerPercent = 59
	}
	entity.Passive.SwordIntent += moved / moveUnitsPerPercent
	if entity.Passive.SwordIntent > entity.Passive.MaxSwordIntent {
		entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	}
}

func (w *World) tickSwordShield(entity *Entity, tick uint64) {
	tickShieldLayers(entity, tick)
	if entity == nil || entity.Passive.Shield <= 0 || entity.Passive.ShieldExpireTick == 0 {
		return
	}
	if tick < entity.Passive.ShieldExpireTick {
		return
	}
	entity.Passive.Shield = 0
	entity.Passive.MaxShield = 0
	entity.Passive.ShieldExpireTick = 0
}

func tickShieldLayers(entity *Entity, tick uint64) {
	if entity == nil || len(entity.Passive.ShieldLayers) == 0 {
		return
	}
	layers := entity.Passive.ShieldLayers[:0]
	total := 0
	for _, layer := range entity.Passive.ShieldLayers {
		if layer.Amount <= 0 || tick >= layer.ExpiresAt {
			continue
		}
		layers = append(layers, layer)
		total += layer.Amount
	}
	entity.Passive.ShieldLayers = layers
	if total == 0 {
		entity.Passive.Shield = 0
		entity.Passive.MaxShield = 0
		return
	}
	entity.Passive.Shield = total
	entity.Passive.MaxShield = total
}

func (w *World) tickTankGraniteShield(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || entity.Stats.HP <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	maxShield := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	if maxShield <= 0 || entity.Passive.Shield >= maxShield {
		entity.Passive.MaxShield = maxShield
		return
	}
	resetTicks := secondsToTicks(skillMetaRange(skill, "resetSeconds", 10), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+resetTicks {
		return
	}
	entity.Passive.MaxShield = maxShield
	entity.Passive.Shield = maxShield
}

func tankGraniteShieldValue(maxHP int, skill config.SkillConfig) int {
	return int(math.Round(float64(maxHP) * skillMetaRange(skill, "shieldMaxHPRatio", 0.1)))
}

func (w *World) tickWarriorToughness(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || entity.Stats.HP <= 0 || entity.Stats.HP >= entity.Stats.MaxHP {
		return
	}
	skill := w.heroPassiveSkill(entity)
	outOfCombatTicks := secondsToTicks(skillMetaRange(skill, "outOfCombatSeconds", 8), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+outOfCombatTicks {
		return
	}
	intervalTicks := secondsToTicks(skillMetaRange(skill, "regenIntervalSeconds", 5), tickRate)
	if intervalTicks == 0 {
		intervalTicks = uint64(tickRate * 5)
	}
	if entity.Passive.NextRegenTick == 0 {
		entity.Passive.NextRegenTick = tick
	}
	if tick < entity.Passive.NextRegenTick {
		return
	}
	ratio := warriorToughnessRegenRatio(entity.Level, skill)
	heal := int(math.Round(float64(entity.Stats.MaxHP) * ratio))
	if heal < 1 {
		heal = 1
	}
	entity.Stats.HP += heal
	if entity.Stats.HP > entity.Stats.MaxHP {
		entity.Stats.HP = entity.Stats.MaxHP
	}
	entity.Passive.NextRegenTick = tick + intervalTicks
}

func tickBaseRegen(entity *Entity, tickRate int) {
	if entity == nil || tickRate <= 0 || entity.Stats.HP <= 0 {
		return
	}
	if entity.Stats.HP < entity.Stats.MaxHP && entity.Stats.HPRegen5 > 0 {
		entity.Regen.HPRemainder += entity.Stats.HPRegen5 / 5 / float64(tickRate)
		heal := int(math.Floor(entity.Regen.HPRemainder + 0.000000001))
		if heal > 0 {
			entity.Stats.HP += heal
			entity.Regen.HPRemainder -= float64(heal)
			if entity.Stats.HP > entity.Stats.MaxHP {
				entity.Stats.HP = entity.Stats.MaxHP
				entity.Regen.HPRemainder = 0
			}
		}
	} else {
		entity.Regen.HPRemainder = 0
	}
	if entity.Stats.MP < entity.Stats.MaxMP && entity.Stats.MPRegen5 > 0 {
		entity.Stats.MP += entity.Stats.MPRegen5 / 5 / float64(tickRate)
		if entity.Stats.MP > entity.Stats.MaxMP {
			entity.Stats.MP = entity.Stats.MaxMP
			entity.Regen.MPRemainder = 0
		}
	} else {
		entity.Regen.MPRemainder = 0
	}
}

func warriorToughnessRegenRatio(level int, skill config.SkillConfig) float64 {
	return skillMetaListByLevel(skill, "regenMaxHPRatio", level, []float64{
		0.015, 0.0198, 0.0246, 0.0294, 0.0342, 0.039,
		0.0438, 0.0486, 0.0534, 0.0582, 0.063, 0.0678,
		0.0726, 0.0774, 0.0822, 0.087, 0.0918, 0.101,
	})
}

func clampInt(value int, min int, max int) int {
	return formula.ClampInt(value, min, max)
}
