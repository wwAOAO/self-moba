package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

func (w *World) ApplyInput(playerID string, input protocol.PlayerInput, tick uint64, skills *config.SkillStore, tickRate int) {
	entity := w.entities[playerEntityID(playerID)]
	if entity == nil {
		entity = w.entities[playerID]
	}
	if entity == nil {
		return
	}
	if entity.Death.Dead {
		return
	}
	if input.DebugLevelUp {
		w.debugLevelUp(entity)
	}
	if input.UpgradeSkill != nil {
		w.upgradeSkill(entity, input.UpgradeSkill.Slot)
	}
	if tick < entity.Control.AirborneUntilTick || tick < entity.Control.ActionLockedUntilTick {
		return
	}
	if input.Move != nil {
		w.cancelTankRPreparedCast(entity)
		target := Vector2{
			X: clamp(input.Move.TargetX, 0, w.width),
			Y: clamp(input.Move.TargetY, 0, w.height),
		}
		entity.Intent.MoveTarget = &target
		entity.Intent.AttackPausedTill = tick + uint64(tickRate*3)
	}
	if input.Move == nil && (input.MoveX != 0 || input.MoveY != 0) {
		w.cancelTankRPreparedCast(entity)
		dx, dy := normalize(input.MoveX, input.MoveY)
		before := entity.Position
		step := movementStepAtTick(entity, tickRate, tick)
		entity.Position.X += dx * step
		entity.Position.Y += dy * step
		entity.Position.X = clamp(entity.Position.X, 0, w.width)
		entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
		w.chargeSwordIntent(entity, distance(before, entity.Position))
	}
	if input.Attack != nil {
		if input.Attack.Clear {
			entity.Intent.AttackTargetID = ""
		} else if input.Attack.TargetID != "" && tick >= entity.Control.DashUntilTick {
			entity.Intent.AttackTargetID = input.Attack.TargetID
			entity.Intent.AttackPausedTill = 0
			entity.Intent.MoveTarget = nil
		}
	}
	if input.Cast != nil {
		if skills == nil {
			skills = w.skills
		}
		w.applyCast(entity, *input.Cast, tick, skills, tickRate)
	}
}

func (w *World) tickPlayer(entity *Entity, tick uint64, tickRate int) {
	if tick < entity.Control.AirborneUntilTick || tick < entity.Control.ActionLockedUntilTick {
		return
	}
	if tick < entity.Control.DashUntilTick {
		return
	}
	target := w.entities[entity.Intent.AttackTargetID]
	attackPaused := tick < entity.Intent.AttackPausedTill
	if !attackPaused && canAttackTarget(entity, target) {
		if distance(entity.Position, target.Position) <= w.attackReachAtTick(entity, target, tick) {
			w.applyAttack(entity, target, tick, tickRate)
			return
		}
		w.moveToward(entity, target.Position, movementStepAtTick(entity, tickRate, tick), 0)
		return
	}
	if entity.Intent.MoveTarget != nil {
		if w.moveToward(entity, *entity.Intent.MoveTarget, movementStepAtTick(entity, tickRate, tick), 8) {
			entity.Intent.MoveTarget = nil
			if entity.HeroID == tankHeroID && entity.Tank.UnstoppableCastPending {
				w.releasePreparedTankR(entity, tick, tickRate)
			}
		}
	}
}

func movementStep(entity *Entity, tickRate int) float64 {
	return movementStepAtTick(entity, tickRate, 0)
}

func movementStepAtTick(entity *Entity, tickRate int, tick uint64) float64 {
	moveSpeed := EffectiveMoveSpeedAtTick(entity, tick)
	if tickRate <= 0 {
		return moveSpeed
	}
	return moveSpeed / float64(tickRate)
}

func EffectiveMoveSpeedAtTick(entity *Entity, tick uint64) float64 {
	if entity == nil {
		return 0
	}
	moveSpeed := entity.Stats.MoveSpeed
	if entity.HeroID == warriorHeroID && tick > 0 && tick < entity.Warrior.DecisiveStrikeSpeedUntilTick {
		moveSpeed *= 1 + entity.Warrior.DecisiveStrikeMoveSpeedBonus
	}
	if entity.Control.MoveSpeedBonusUntil > 0 && (tick == 0 || tick < entity.Control.MoveSpeedBonusUntil) {
		moveSpeed += entity.Control.MoveSpeedBonusFlat
	}
	if entity.Control.MoveSpeedSlowUntil > 0 && (tick == 0 || tick < entity.Control.MoveSpeedSlowUntil) {
		moveSpeed *= 1 - clamp(entity.Control.MoveSpeedSlow, 0, 1)
	}
	return moveSpeed
}

func EffectiveAttackSpeedAtTick(entity *Entity, tick uint64) float64 {
	if entity == nil {
		return 0
	}
	attackSpeed := entity.Stats.AttackSpeed
	if entity.Control.AttackSpeedSlowUntil > 0 && (tick == 0 || tick < entity.Control.AttackSpeedSlowUntil) {
		attackSpeed *= 1 - clamp(entity.Control.AttackSpeedSlow, 0, 1)
	}
	return clamp(attackSpeed, 0, 2.5)
}

func (w *World) moveToward(entity *Entity, destination Vector2, step float64, stopDistance float64) bool {
	dx := destination.X - entity.Position.X
	dy := destination.Y - entity.Position.Y
	dist := math.Hypot(dx, dy)
	if dist <= stopDistance {
		return true
	}
	if dist <= step+stopDistance {
		ratio := math.Max(0, dist-stopDistance) / dist
		before := entity.Position
		entity.Position.X += dx * ratio
		entity.Position.Y += dy * ratio
		w.chargeSwordIntent(entity, distance(before, entity.Position))
		return true
	}
	before := entity.Position
	entity.Position.X += dx / dist * step
	entity.Position.Y += dy / dist * step
	entity.Position.X = clamp(entity.Position.X, 0, w.width)
	entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
	w.chargeSwordIntent(entity, distance(before, entity.Position))
	return false
}
