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
	if heroHooksFor(swordHeroID).PassiveState != nil {
		if state := heroHooksFor(swordHeroID).PassiveState(w, hero); state.MaxSwordIntent > 0 {
			return state
		}
	}
	if hero.HeroID == tankHeroID {
		if heroHooksFor(tankHeroID).PassiveState != nil {
			return heroHooksFor(tankHeroID).PassiveState(w, hero)
		}
	}
	return PassiveState{}
}

func swordStateForHero(heroID string) SwordState {
	if heroHooksFor(swordHeroID).StateForHero != nil {
		return heroHooksFor(swordHeroID).StateForHero(heroID)
	}
	return SwordState{}
}

func (w *World) chargeSwordIntent(entity *Entity, moved float64) {
	if heroHooksFor(swordHeroID).ChargeIntent != nil {
		heroHooksFor(swordHeroID).ChargeIntent(w, entity, moved)
	}
}

func (w *World) tickSwordShield(entity *Entity, tick uint64) {
	tickShieldLayers(entity, tick)
	if entity == nil || entity.Passive.Shield <= 0 || entity.Passive.ShieldExpireTick == 0 {
		return
	}
	if heroHooksFor(swordHeroID).TickShield != nil {
		heroHooksFor(swordHeroID).TickShield(entity, tick)
	}
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

func (w *World) WarriorPassiveSkill(entity *Entity) config.SkillConfig {
	return w.heroPassiveSkill(entity)
}

func clampInt(value int, min int, max int) int {
	return formula.ClampInt(value, min, max)
}
