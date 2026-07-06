package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	maxCharges := archerHawkMaxCharges(skill)
	if state.Stacks <= 0 {
		return
	}
	state.Stacks--
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skill.CooldownMS, tickRate)
	if state.Stacks < maxCharges && state.StacksExpireTick == 0 {
		state.StacksExpireTick = tick + archerHawkRechargeTicks(entity, skill, state.Level, tickRate)
	}
	entity.Skills[eID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)

	target := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	speed := skillMeta(skill, "projectileSpeed", 1800)
	travelSeconds := distance(entity.Position, target) / speed
	arriveTick := tick + secondsToTicks(travelSeconds, tickRate)
	expiresAt := arriveTick + secondsToTicks(skillMeta(skill, "lingerSeconds", 5), tickRate)
	effectSpeed := speed
	if tickRate > 0 {
		effectSpeed /= float64(tickRate)
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:        w.NextEffectID("effect:archer_hawk:"),
		Kind:      "archer_hawk",
		Team:      entity.Team,
		Start:     entity.Position,
		End:       target,
		Dir:       world.Vector2{X: dx, Y: dy},
		Speed:     effectSpeed,
		Height:    float64(arriveTick),
		Radius:    80,
		CreatedAt: tick,
		ExpiresAt: expiresAt,
	})
}

func RefreshSkillOnUpgrade(w *world.World, entity *world.Entity, skillID string) {
	if entity == nil || entity.HeroID != heroID || skillID != eID {
		return
	}
	state := entity.Skills[eID]
	if state.Level <= 0 {
		return
	}
	maxCharges := archerHawkMaxCharges(w.SkillConfig(eID))
	if state.Stacks <= 0 {
		state.Stacks = maxCharges
		state.StacksExpireTick = 0
		entity.Skills[eID] = state
	}
}

func TickHawkCharges(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	state, ok := entity.Skills[eID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(eID)
	maxCharges := archerHawkMaxCharges(skill)
	if state.Stacks >= maxCharges {
		state.Stacks = maxCharges
		state.StacksExpireTick = 0
		entity.Skills[eID] = state
		return
	}
	if state.StacksExpireTick == 0 {
		state.StacksExpireTick = tick + archerHawkRechargeTicks(entity, skill, state.Level, tickRate)
		entity.Skills[eID] = state
		return
	}
	for state.Stacks < maxCharges && state.StacksExpireTick > 0 && tick >= state.StacksExpireTick {
		state.Stacks++
		if state.Stacks >= maxCharges {
			state.StacksExpireTick = 0
			break
		}
		state.StacksExpireTick += archerHawkRechargeTicks(entity, skill, state.Level, tickRate)
	}
	entity.Skills[eID] = state
}

func archerHawkMaxCharges(skill config.SkillConfig) int {
	value := int(math.Round(skillMeta(skill, "maxCharges", 2)))
	if value < 1 {
		return 1
	}
	return value
}

func archerHawkRechargeTicks(entity *world.Entity, skill config.SkillConfig, level int, tickRate int) uint64 {
	return cooldownTicksFor(entity, int(math.Round(skillList(skill, "rechargeMs", level, []float64{90000, 80000, 70000, 60000, 50000}))), tickRate)
}
