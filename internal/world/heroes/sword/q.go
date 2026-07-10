package sword

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const qStackTicks = 6

func ApplyQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.QPending {
		return
	}
	if tick > state.StacksExpireTick {
		state.Stacks = 0
	}
	hasWhirlwindStack := state.Stacks >= 2
	form := "line"
	qRange := skillRange(skill, 475)
	if eqWindowActive(entity, skill, tick, tickRate) {
		form = "circle"
		qRange = skillMeta(skill, "eqRadius", 375)
	} else if hasWhirlwindStack {
		form = "whirlwind"
		qRange = skillMeta(skill, "whirlwindRange", 900)
	}
	windupTicks := windupTicks(entity, skill, tickRate)
	entity.Sword.QPending = true
	entity.Sword.QReleaseTick = tick + windupTicks
	entity.Sword.QTarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Sword.QForm = form
	entity.Sword.QRange = qRange
	entity.Control.ActionLockedUntilTick = tick + windupTicks
	entity.Skills[qID] = state
}

func ReleaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Sword.QPending || tick < entity.Sword.QReleaseTick {
		return
	}
	skill := w.SkillConfig(qID)
	state := entity.Skills[qID]
	form := entity.Sword.QForm
	qRange := entity.Sword.QRange
	targetPoint := entity.Sword.QTarget
	entity.Sword.QPending = false
	entity.Sword.QReleaseTick = 0
	entity.Sword.QTarget = world.Vector2{}
	entity.Sword.QForm = ""
	entity.Sword.QRange = 0
	hasWhirlwindStack := state.Stacks >= 2
	state.CooldownUntilTick = tick + w.SwordQCooldownTicks(entity, skill, state.Level, tickRate)
	if form == "whirlwind" {
		spawnWhirlwind(w, entity, targetPoint, qRange, skill, tick, tickRate)
		w.LockAttackAfterCast(entity, tick, tickRate)
		state.Stacks = 0
		state.StacksExpireTick = 0
		entity.Skills[qID] = state
		return
	}
	targets := w.SwordQTargets(entity, targetPoint, qRange, form, skill)
	for _, target := range targets {
		damage := w.SwordQDamage(entity, target, skill, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			if form == "circle" {
				w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
			} else {
				w.ApplyDamage(entity, target, damage, tickRate)
			}
			if form == "circle" && hasWhirlwindStack {
				w.ApplyAirborne(target, tick+secondsToTicks(skillMeta(skill, "knockupSeconds", 1), tickRate), tick, tickRate)
			}
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
		}
	}
	if len(targets) > 0 {
		if form == "circle" && hasWhirlwindStack {
			state.Stacks = 0
			state.StacksExpireTick = 0
		} else {
			state.Stacks++
			if state.Stacks > 2 {
				state.Stacks = 2
			}
			state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "stackDurationSeconds", qStackTicks), tickRate)
		}
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[qID] = state
}

func ExpireQStacks(entity *world.Entity, tick uint64) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	state := entity.Skills[qID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[qID] = state
}

func windupTicks(entity *world.Entity, skill config.SkillConfig, tickRate int) uint64 {
	ticks := secondsToTicks(windupSeconds(entity, skill), tickRate)
	if ticks < 1 {
		return 1
	}
	return ticks
}

func windupSeconds(entity *world.Entity, skill config.SkillConfig) float64 {
	base := skillMeta(skill, "castWindupSeconds", 0.328)
	minimum := skillMeta(skill, "minCastWindupSeconds", 0.09)
	bonus := 0.0
	if entity != nil {
		bonus = entity.Stats.AttackSpeedBonus
	}
	bonus = math.Max(0, bonus)
	seconds := base / (1 + bonus)
	if seconds < minimum {
		return minimum
	}
	return seconds
}

func eqWindowActive(entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || tick >= entity.Control.DashUntilTick {
		return false
	}
	windowTicks := secondsToTicks(skillMeta(skill, "eqWindowSeconds", 0.15), tickRate)
	if windowTicks == 0 {
		return false
	}
	return entity.Control.DashUntilTick-tick <= windowTicks
}

func spawnWhirlwind(w *world.World, entity *world.Entity, targetPoint world.Vector2, qRange float64, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	radius := skillMeta(skill, "whirlwindRadius", 70)
	speedPerSecond := skillMeta(skill, "whirlwindSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	id := w.NextProjectileID("projectile:sword_whirlwind:")
	w.PutProjectile(&world.Projectile{
		ID:           id,
		Kind:         "sword_whirlwind",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       radius,
		Damage:       w.SwordQDamage(entity, &world.Entity{ID: id}, skill, tick),
		KnockupTicks: secondsToTicks(skillMeta(skill, "knockupSeconds", 1), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	})
}
