package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID = "mage"
	qID    = "mage_q"
	wID    = "mage_w"
	rID    = "mage_r"
)

func init() {
	world.RegisterHeroCastHandlers(heroID, map[string]world.HeroCastHandler{
		qID: ApplyQ,
		wID: ApplyW,
		rID: ApplyR,
	})
}

func ApplyQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.LightBindingPending {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.LightBindingPending = true
	entity.Mage.LightBindingReleaseTick = tick + windupTicks
	entity.Mage.LightBindingTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Mage.LightBindingLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.LightBindingReleaseTick
	entity.Skills[qID] = state
}

func ApplyW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.PrismaticBarrierPending {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 60)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.2), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.PrismaticBarrierPending = true
	entity.Mage.PrismaticBarrierReleaseTick = tick + windupTicks
	entity.Mage.PrismaticBarrierTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Mage.PrismaticBarrierLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.PrismaticBarrierReleaseTick
	entity.Skills[wID] = state
}

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.FinalSparkPending {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.5), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.FinalSparkPending = true
	entity.Mage.FinalSparkReleaseTick = tick + windupTicks
	entity.Mage.FinalSparkTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Mage.FinalSparkLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.FinalSparkReleaseTick
	entity.Skills[rID] = state
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

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}
