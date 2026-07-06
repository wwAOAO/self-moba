package world

import (
	"l-battle/internal/protocol"
	"testing"
)

func assertPlayerTeam(t *testing.T, w *World, playerID string, want Team) {
	t.Helper()
	entity := w.entities[playerEntityID(playerID)]
	if entity == nil {
		t.Fatalf("player %s not found", playerID)
	}
	if entity.Team != want {
		t.Fatalf("player %s team = %s, want %s", playerID, entity.Team, want)
	}
}

func learnSkill(entity *Entity, skillID string, level int) {
	state := entity.Skills[skillID]
	state.SkillID = skillID
	state.Level = level
	entity.Skills[skillID] = state
}

func protocolPlayerInputMove(x float64, y float64) protocol.PlayerInput {
	return protocol.PlayerInput{
		Move: &protocol.MoveInput{
			TargetX: x,
			TargetY: y,
		},
	}
}

func protocolPlayerInputAttack(targetID string) protocol.PlayerInput {
	return protocol.PlayerInput{
		Attack: &protocol.AttackInput{
			TargetID: targetID,
		},
	}
}

func protocolPlayerInputCast(skillID string, targetX float64, targetY float64) protocol.PlayerInput {
	return protocolPlayerInputCastTarget(skillID, "", targetX, targetY)
}

func protocolPlayerInputCastTarget(skillID string, targetID string, targetX float64, targetY float64) protocol.PlayerInput {
	return protocol.PlayerInput{
		Cast: &protocol.CastInput{
			SkillID:  skillID,
			TargetID: targetID,
			TargetX:  targetX,
			TargetY:  targetY,
		},
	}
}

func protocolPlayerInputUpgrade(slot string) protocol.PlayerInput {
	return protocol.PlayerInput{
		UpgradeSkill: &protocol.UpgradeSkillInput{
			Slot: slot,
		},
	}
}

func protocolPlayerInputBuyEquipment(equipmentID string) protocol.PlayerInput {
	return protocol.PlayerInput{
		BuyEquipment: &protocol.BuyEquipmentInput{
			EquipmentID: equipmentID,
		},
	}
}

func protocolPlayerInputSellEquipment(slot int) protocol.PlayerInput {
	return protocol.PlayerInput{
		SellEquipment: &protocol.SellEquipmentInput{
			Slot: slot,
		},
	}
}

func protocolPlayerInputDebugLevelUp() protocol.PlayerInput {
	return protocol.PlayerInput{
		DebugLevelUp: true,
	}
}

func protocolPlayerInputDebugAbilityHaste(value float64) protocol.PlayerInput {
	return protocol.PlayerInput{
		DebugAbilityHaste: &value,
	}
}

func assertSkillEffect(t *testing.T, effects []SkillEffect, kind string) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			return
		}
	}
	t.Fatalf("missing skill effect %s", kind)
}

func tickSwordQRelease(t *testing.T, w *World, entity *Entity, tickRate int) uint64 {
	t.Helper()
	if entity == nil || !entity.Sword.QPending {
		t.Fatal("sword q is not pending")
	}
	releaseTick := entity.Sword.QReleaseTick
	w.Tick(releaseTick, tickRate)
	if entity.Sword.QPending {
		t.Fatalf("sword q still pending after release tick %d", releaseTick)
	}
	return releaseTick
}

func tickAttackRelease(t *testing.T, w *World, entity *Entity, tickRate int) uint64 {
	t.Helper()
	if entity == nil || entity.Combat.PendingAttackTargetID == "" {
		t.Fatal("basic attack is not pending")
	}
	releaseTick := entity.Combat.AttackReleaseTick
	w.Tick(releaseTick, tickRate)
	if entity.Combat.PendingAttackTargetID != "" {
		t.Fatalf("basic attack still pending after release tick %d", releaseTick)
	}
	return releaseTick
}

func tickUntilDamage(t *testing.T, w *World, entity *Entity, from uint64, to uint64, tickRate int) {
	t.Helper()
	for tick := from + 1; tick <= to; tick++ {
		w.Tick(tick, tickRate)
		if entity.Combat.LastDamage > 0 {
			return
		}
	}
	t.Fatalf("entity %s was not damaged by tick %d", entity.ID, to)
}

func countProjectilesByKind(w *World, kind string) int {
	count := 0
	for _, projectile := range w.projectiles {
		if projectile.Kind == kind {
			count++
		}
	}
	return count
}

func placeEntity(entity *Entity, x float64, y float64) {
	if entity != nil {
		entity.Position = Vector2{X: x, Y: y}
	}
}
