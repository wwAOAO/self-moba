package blade

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyW(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || state.Stacks > 0 {
		return
	}
	if len(mockingShoutTargets(w, entity, skill.Range)) == 0 {
		return
	}
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.3), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	state.Stacks = 1
	state.StacksExpireTick = tick + windupTicks
	entity.Control.ActionLockedUntilTick = state.StacksExpireTick
	entity.Skills[wID] = state
}

func ReleaseW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Stats.HP <= 0 {
		return
	}
	state := entity.Skills[wID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	skill := w.SkillConfig(wID)
	until := tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate)
	attackReduction := skillList(skill, "attackReduction", state.Level, []float64{20, 35, 50, 65, 80})
	slow := skillList(skill, "moveSpeedSlow", state.Level, []float64{0.3, 0.375, 0.45, 0.525, 0.6})
	for _, target := range mockingShoutTargets(w, entity, skill.Range) {
		w.ApplyAttackDamageReduction(target, attackReduction, until)
		w.ApplyMoveSpeedSlow(target, slow, until)
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{14000, 14000, 14000, 14000, 14000})), tickRate)
	entity.Skills[wID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func mockingShoutTargets(w *world.World, entity *world.Entity, radius float64) []*world.Entity {
	if radius <= 0 {
		radius = 700
	}
	targets := make([]*world.Entity, 0)
	w.ForEachEntity(func(target *world.Entity) {
		if target == nil || target.Team == entity.Team || target.Stats.HP <= 0 {
			return
		}
		if target.Kind != world.EntityKindPlayer && target.Kind != world.EntityKindEnemyHero {
			return
		}
		if math.Hypot(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y) <= radius+target.Radius {
			targets = append(targets, target)
		}
	})
	return targets
}
