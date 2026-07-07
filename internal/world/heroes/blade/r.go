package blade

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyR(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Stats.HP <= 0 {
		return
	}
	entity.Control.UndyingRageUntil = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 5), tickRate)
	entity.Control.UndyingRageMinHP = math.Round(skillList(skill, "minHP", state.Level, []float64{30, 50, 70}))
	gainRage(entity, skillList(skill, "rageGain", state.Level, []float64{50, 75, 100}), tick)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{110000, 100000, 90000})), tickRate)
	entity.Skills[rID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
}
