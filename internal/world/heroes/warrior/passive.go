package warrior

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func TickToughness(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Stats.HP <= 0 || entity.Stats.HP >= entity.Stats.MaxHP {
		return
	}
	skill := w.WarriorPassiveSkill(entity)
	outOfCombatTicks := secondsToTicks(skillMeta(skill, "outOfCombatSeconds", 8), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+outOfCombatTicks {
		return
	}
	intervalTicks := secondsToTicks(skillMeta(skill, "regenIntervalSeconds", 5), tickRate)
	if intervalTicks == 0 {
		intervalTicks = uint64(tickRate * 5)
	}
	if entity.Passive.NextRegenTick == 0 {
		entity.Passive.NextRegenTick = tick
	}
	if tick < entity.Passive.NextRegenTick {
		return
	}
	ratio := ToughnessRegenRatio(entity.Level, skill)
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

func ToughnessRegenRatio(level int, skill config.SkillConfig) float64 {
	return skillList(skill, "regenMaxHPRatio", level, []float64{
		0.015, 0.0198, 0.0246, 0.0294, 0.0342, 0.039,
		0.0438, 0.0486, 0.0534, 0.0582, 0.063, 0.0678,
		0.0726, 0.0774, 0.0822, 0.087, 0.0918, 0.101,
	})
}
