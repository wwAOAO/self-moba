package blade

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyQ(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Stats.HP <= 0 {
		return
	}
	rage := math.Max(0, math.Min(entity.Stats.MP, maxRage(entity)))
	healValue := skillList(skill, "baseHeal", state.Level, []float64{30, 40, 50, 60, 70}) +
		float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 1.5) +
		rage*skillList(skill, "healPerRage", state.Level, []float64{0.5, 0.95, 1.4, 1.85, 2.3})
	entity.Stats.MP = 0
	beforeHP := entity.Stats.HP
	if heal := math.Round(healValue); heal > 0 {
		entity.Stats.HP += heal
		if entity.Stats.HP > entity.Stats.MaxHP {
			entity.Stats.HP = entity.Stats.MaxHP
		}
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:blade_q_heal:"),
		Kind:         "blade_q_heal",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       skillMeta(skill, "healEffectRadius", 90),
		Count:        int(math.Round(entity.Stats.HP - beforeHP)),
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "healEffectSeconds", 0.7), tickRate),
	})
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skill.CooldownMS, tickRate)
	entity.Skills[qID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
	w.RefreshPlayerStats(entity)
}

func ApplyBloodlustStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if w == nil || entity == nil || stats == nil || entity.HeroID != heroID {
		return
	}
	state := entity.Skills[qID]
	if state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(qID)
	missingPercent := 0.0
	if stats.MaxHP > 0 && stats.HP < stats.MaxHP {
		missingPercent = float64(stats.MaxHP-stats.HP) * 100 / float64(stats.MaxHP)
	}
	bonus := skillList(skill, "baseAttack", state.Level, []float64{5, 10, 15, 20, 25}) +
		missingPercent*skillList(skill, "missingHPAttackPerPercent", state.Level, []float64{0.15, 0.2, 0.25, 0.3, 0.35})
	stats.Attack += bonus
	stats.BonusAttack += bonus
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
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

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}
