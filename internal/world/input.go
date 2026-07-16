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
	if input.DebugAbilityHaste != nil {
		w.setDebugAbilityHasteBuff(entity, clamp(*input.DebugAbilityHaste, 0, 10000))
		w.recalculatePlayerStats(entity)
	}
	if input.DebugGold > 0 {
		w.addGold(entity, input.DebugGold)
	}
	if input.UpgradeSkill != nil {
		w.upgradeSkill(entity, input.UpgradeSkill.Slot)
	}
	if input.BuyEquipment != nil {
		w.buyEquipment(entity, input.BuyEquipment.EquipmentID, tick)
	}
	if input.SellEquipment != nil {
		w.sellEquipment(entity, input.SellEquipment.Slot)
	}
	moving := input.Move != nil || (input.MoveX != 0 || input.MoveY != 0)
	if moving {
		w.cancelGunnerRChannel(entity)
		if h := heroHooksForEntity(entity).OnMoveInput; h != nil {
			h(w, entity, tick)
		}
	}
	if tick < entity.Control.AirborneUntilTick || tick < entity.Control.ActionLockedUntilTick || tick < entity.Control.StunnedUntilTick || tick < entity.Control.SuppressedUntilTick || tick < entity.Control.TauntedUntilTick {
		return
	}
	rooted := tick < entity.Control.RootedUntilTick
	if input.Move != nil && !rooted {
		w.cancelTankRPreparedCast(entity)
		entity.Combat.PendingAttackTargetID = ""
		entity.Combat.AttackReleaseTick = 0
		target := Vector2{
			X: clamp(input.Move.TargetX, 0, w.width),
			Y: clamp(input.Move.TargetY, 0, w.height),
		}
		entity.Intent.MoveTarget = &target
		entity.Intent.AttackPausedTill = tick + uint64(tickRate*3)
	}
	if input.Move == nil && !rooted && (input.MoveX != 0 || input.MoveY != 0) {
		w.cancelTankRPreparedCast(entity)
		entity.Combat.PendingAttackTargetID = ""
		entity.Combat.AttackReleaseTick = 0
		dx, dy := normalize(input.MoveX, input.MoveY)
		before := entity.Position
		step := movementStepAtTick(entity, tickRate, tick)
		entity.Position = w.resolveCollisionPosition(entity, Vector2{
			X: clamp(entity.Position.X+dx*step, 0, w.width),
			Y: clamp(entity.Position.Y+dy*step, 0, w.height),
		})
		moved := distance(before, entity.Position)
		w.chargeSwordIntent(entity, moved)
		w.chargeEquipmentOnMove(entity, moved)
	}
	if input.Attack != nil {
		if input.Attack.Clear {
			entity.Intent.AttackTargetID = ""
			entity.Combat.PendingAttackTargetID = ""
			entity.Combat.AttackReleaseTick = 0
		} else if input.Attack.TargetID != "" && tick >= entity.Control.DashUntilTick {
			if entity.Intent.AttackTargetID != input.Attack.TargetID {
				entity.Combat.PendingAttackTargetID = ""
				entity.Combat.AttackReleaseTick = 0
			}
			entity.Intent.AttackTargetID = input.Attack.TargetID
			entity.Intent.AttackPausedTill = 0
			entity.Intent.MoveTarget = nil
		}
	}
	if input.Cast != nil {
		if !canCastDuringSwordEDash(entity, tick, tickRate) {
			return
		}
		if skills == nil {
			skills = w.skills
		}
		w.applyCast(entity, *input.Cast, tick, skills, tickRate)
	}
}

func canCastDuringSwordEDash(entity *Entity, tick uint64, tickRate int) bool {
	if entity == nil || entity.HeroID != swordHeroID || tick >= entity.Control.DashUntilTick {
		return true
	}
	return entity.Control.DashUntilTick-tick <= secondsToTicks(0.2, tickRate)
}

func (w *World) tickPlayer(entity *Entity, tick uint64, tickRate int) {
	if tick < entity.Control.AirborneUntilTick || tick < entity.Control.ActionLockedUntilTick || tick < entity.Control.StunnedUntilTick || tick < entity.Control.SuppressedUntilTick || tick < entity.Control.TauntedUntilTick {
		return
	}
	if tick < entity.Control.DashUntilTick {
		return
	}
	rooted := tick < entity.Control.RootedUntilTick
	if entity.Combat.PendingAttackTargetID != "" {
		return
	}
	target := w.entities[entity.Intent.AttackTargetID]
	attackPaused := tick < entity.Intent.AttackPausedTill
	if !attackPaused && canAttackTarget(entity, target) {
		if distance(entity.Position, target.Position) <= w.attackReachAtTick(entity, target, tick) {
			w.applyAttack(entity, target, tick, tickRate)
			return
		}
		if rooted {
			return
		}
		w.moveToward(entity, target.Position, movementStepAtTick(entity, tickRate, tick), 0)
		return
	}
	if entity.Intent.MoveTarget != nil && !rooted {
		if w.moveToward(entity, *entity.Intent.MoveTarget, movementStepAtTick(entity, tickRate, tick), 8) {
			entity.Intent.MoveTarget = nil
			w.releasePreparedTankR(entity, tick, tickRate)
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
	moveSpeed *= heroMoveSpeedMultiplier(entity, tick)
	if entity.Control.MoveSpeedBonusUntil > 0 && (tick == 0 || tick < entity.Control.MoveSpeedBonusUntil) {
		moveSpeed += entity.Control.MoveSpeedBonusFlat
	}
	moveSpeed += equipmentOutOfCombatMoveSpeed(entity, tick)
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
	if entity.HeroID == archerHeroID && entity.Archer.FocusActiveUntil > 0 && (tick == 0 || tick < entity.Archer.FocusActiveUntil) {
		attackSpeed *= 1 + entity.Archer.FocusAttackSpeed
	}
	if entity.HeroID == gunnerHeroID && entity.Passive.GunnerWActiveUntil > 0 && (tick == 0 || tick < entity.Passive.GunnerWActiveUntil) {
		attackSpeed *= 1 + entity.Passive.GunnerWAttackSpeed
	}
	attackSpeed *= heroAttackSpeedMultiplier(entity, tick)
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
		entity.Position = w.resolveCollisionPosition(entity, Vector2{
			X: clamp(entity.Position.X+dx*ratio, 0, w.width),
			Y: clamp(entity.Position.Y+dy*ratio, 0, w.height),
		})
		moved := distance(before, entity.Position)
		w.chargeSwordIntent(entity, moved)
		w.chargeEquipmentOnMove(entity, moved)
		return true
	}
	before := entity.Position
	entity.Position = w.resolveCollisionPosition(entity, Vector2{
		X: clamp(entity.Position.X+dx/dist*step, 0, w.width),
		Y: clamp(entity.Position.Y+dy/dist*step, 0, w.height),
	})
	moved := distance(before, entity.Position)
	w.chargeSwordIntent(entity, moved)
	w.chargeEquipmentOnMove(entity, moved)
	return false
}

func (w *World) resolveCollisionPosition(entity *Entity, candidate Vector2) Vector2 {
	if !isCollisionEntity(entity) {
		return candidate
	}
	for _, other := range w.entities {
		if other == nil || other.ID == entity.ID || !isCollisionEntity(other) {
			continue
		}
		minDistance := entity.Radius + other.Radius
		if minDistance <= 0 {
			continue
		}
		dx := candidate.X - other.Position.X
		dy := candidate.Y - other.Position.Y
		dist := math.Hypot(dx, dy)
		if dist <= 0 || dist >= minDistance {
			continue
		}
		push := minDistance - dist
		candidate.X = clamp(candidate.X+dx/dist*push, 0, w.width)
		candidate.Y = clamp(candidate.Y+dy/dist*push, 0, w.height)
	}
	return candidate
}

func isCollisionEntity(entity *Entity) bool {
	if entity == nil || entity.Stats.HP <= 0 {
		return false
	}
	if entity.Kind == EntityKindPlayer {
		return !entity.Death.Dead
	}
	return entity.Kind == EntityKindEnemyHero || isMinion(entity) || isStructure(entity)
}
