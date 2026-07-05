package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

func applySwordQ(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.QPending {
		return
	}
	if tick > state.StacksExpireTick {
		state.Stacks = 0
	}
	hasWhirlwindStack := state.Stacks >= 2
	form := "line"
	qRange := skillRange(skill, 475)
	if swordEQWindowActive(entity, skill, tick, tickRate) {
		form = "circle"
		qRange = skillMetaRange(skill, "eqRadius", 375)
	} else if hasWhirlwindStack {
		form = "whirlwind"
		qRange = skillMetaRange(skill, "whirlwindRange", 900)
	}
	windupTicks := w.swordQWindupTicks(entity, skill, tickRate)
	entity.Sword.QPending = true
	entity.Sword.QReleaseTick = tick + windupTicks
	entity.Sword.QTarget = Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Sword.QForm = form
	entity.Sword.QRange = qRange
	entity.Control.ActionLockedUntilTick = tick + windupTicks
	entity.Skills[swordQSkillID] = state
}

func (w *World) releaseSwordQ(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != swordHeroID || !entity.Sword.QPending || tick < entity.Sword.QReleaseTick {
		return
	}
	skill := w.SkillConfig(swordQSkillID)
	state := entity.Skills[swordQSkillID]
	form := entity.Sword.QForm
	qRange := entity.Sword.QRange
	targetPoint := entity.Sword.QTarget
	entity.Sword.QPending = false
	entity.Sword.QReleaseTick = 0
	entity.Sword.QTarget = Vector2{}
	entity.Sword.QForm = ""
	entity.Sword.QRange = 0
	hasWhirlwindStack := state.Stacks >= 2
	state.CooldownUntilTick = tick + w.swordQCooldownTicks(entity, skill, state.Level, tickRate)
	if form == "whirlwind" {
		w.spawnSwordWhirlwind(entity, targetPoint, qRange, skill, tick, tickRate)
		w.LockAttackAfterCast(entity, tick, tickRate)
		state.Stacks = 0
		state.StacksExpireTick = 0
		entity.Skills[swordQSkillID] = state
		return
	}
	targets := w.swordQTargets(entity, targetPoint, qRange, form, skill)
	for _, target := range targets {
		damage := w.swordQDamage(entity, target, skill, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			if form == "circle" {
				w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
			} else {
				w.ApplyDamage(entity, target, damage, tickRate)
			}
			if form == "circle" && hasWhirlwindStack {
				target.Control.AirborneUntilTick = tick + secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1), tickRate)
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
			state.StacksExpireTick = tick + secondsToTicks(skillMetaRange(skill, "stackDurationSeconds", swordQStackTicks), tickRate)
		}
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordQSkillID] = state
}

func (w *World) expireSwordQStacks(entity *Entity, tick uint64) {
	if entity == nil || entity.HeroID != swordHeroID {
		return
	}
	state := entity.Skills[swordQSkillID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[swordQSkillID] = state
}

func (w *World) swordQWindupTicks(entity *Entity, skill config.SkillConfig, tickRate int) uint64 {
	seconds := swordQWindupSeconds(entity, skill)
	ticks := secondsToTicks(seconds, tickRate)
	if ticks < 1 {
		return 1
	}
	return ticks
}

func swordQWindupSeconds(entity *Entity, skill config.SkillConfig) float64 {
	base := skillMetaRange(skill, "castWindupSeconds", 0.328)
	minimum := skillMetaRange(skill, "minCastWindupSeconds", 0.09)
	bonus := 0.0
	if entity != nil {
		bonus = entity.Stats.AttackSpeedBonus
	}
	bonus = clamp(bonus, 0, math.MaxFloat64)
	seconds := base / (1 + bonus)
	if seconds < minimum {
		return minimum
	}
	return seconds
}

func swordEQWindowActive(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || tick >= entity.Control.DashUntilTick {
		return false
	}
	windowTicks := secondsToTicks(skillMetaRange(skill, "eqWindowSeconds", 0.15), tickRate)
	if windowTicks == 0 {
		return false
	}
	return entity.Control.DashUntilTick-tick <= windowTicks
}

func (w *World) spawnSwordWhirlwind(entity *Entity, targetPoint Vector2, qRange float64, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	radius := skillMetaRange(skill, "whirlwindRadius", 70)
	speedPerSecond := skillMetaRange(skill, "whirlwindSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	id := w.NextProjectileID("projectile:sword_whirlwind:")
	w.PutProjectile(&Projectile{
		ID:           id,
		Kind:         "sword_whirlwind",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      swordQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       radius,
		Damage:       w.swordQDamage(entity, &Entity{ID: id}, skill, tick),
		KnockupTicks: secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	})
}
