package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID = "archer"
	qID    = "shot"
	eID    = "trap"
	rID    = "arrow_rain"
)

func init() {
	world.RegisterHeroCastHandlers(heroID, map[string]world.HeroCastHandler{
		qID: ApplyQ,
		eID: ApplyE,
		rID: ApplyR,
	})
}

func ApplyQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	maxStacks := archerFocusMaxStacks(skill)
	if entity.Archer.FocusStacks < maxStacks {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 50)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Archer.FocusStacks = 0
	entity.Archer.FocusExpireTick = 0
	entity.Archer.FocusActiveLevel = state.Level
	entity.Archer.FocusAttackSpeed = skillList(skill, "attackSpeedBonus", state.Level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
	entity.Archer.FocusBonusADRatio = skillList(skill, "bonusAdDamageRatio", state.Level, []float64{1.05, 1.1, 1.15, 1.2, 1.25})
	entity.Archer.FocusActiveUntil = tick + secondsToTicks(skillMeta(skill, "activeDurationSeconds", 5), tickRate)
	entity.Skills[qID] = state
}

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

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Skills[rID] = state
	delayTicks := secondsToTicks(skillMeta(skill, "castDelaySeconds", 0.25), tickRate)
	entity.Control.ActionLockedUntilTick = tick + delayTicks
	entity.Archer.CrystalArrowPending = true
	entity.Archer.CrystalArrowReleaseTick = tick + delayTicks
	entity.Archer.CrystalArrowTarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Archer.CrystalArrowLevel = state.Level
}

func archerFocusMaxStacks(skill config.SkillConfig) int {
	value := int(math.Round(skillMeta(skill, "maxStacks", 4)))
	if value < 1 {
		return 1
	}
	return value
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

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}
