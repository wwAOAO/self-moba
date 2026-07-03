package world

import "strconv"

const debugAbilityHasteBuffID = "debug_ability_haste"

func (w *World) setDebugAbilityHasteBuff(entity *Entity, value float64) {
	removeBuff(entity, debugAbilityHasteBuffID)
	if value <= 0 {
		return
	}
	entity.Buffs = append(entity.Buffs, BuffState{
		ID:           debugAbilityHasteBuffID,
		Name:         "+200技能急速",
		AbilityHaste: value,
	})
}

func removeBuff(entity *Entity, id string) {
	if entity == nil {
		return
	}
	for i := 0; i < len(entity.Buffs); i++ {
		if entity.Buffs[i].ID != id {
			continue
		}
		entity.Buffs = append(entity.Buffs[:i], entity.Buffs[i+1:]...)
		i--
	}
}

func abilityHasteFromBuffs(entity *Entity) float64 {
	if entity == nil {
		return 0
	}
	total := 0.0
	for _, buff := range entity.Buffs {
		total += buff.AbilityHaste
	}
	return total
}

func (w *World) ActiveBuffs(entity *Entity, tick uint64) []BuffState {
	if entity == nil {
		return nil
	}
	buffs := make([]BuffState, 0, len(entity.Buffs)+8)
	for _, buff := range entity.Buffs {
		if buff.ExpiresAtTick > 0 && tick >= buff.ExpiresAtTick {
			continue
		}
		buffs = append(buffs, buff)
	}
	buffs = append(buffs, w.activeEquipmentBurnBuffs(entity, tick)...)
	buffs = append(buffs, activeCombatBuffs(entity, tick)...)
	buffs = append(buffs, activeEquipmentSlotBuffs(entity, tick)...)
	return buffs
}

func (w *World) activeEquipmentBurnBuffs(entity *Entity, tick uint64) []BuffState {
	if w == nil || len(w.equipmentBurns) == 0 {
		return nil
	}
	buffs := make([]BuffState, 0, 1)
	for _, burn := range w.equipmentBurns {
		if burn.TargetID != entity.ID || tick >= burn.ExpiresAt {
			continue
		}
		buffs = append(buffs, BuffState{
			ID:            "liandrys_burn:" + burn.SourceID,
			Name:          "兰德里灼烧",
			ExpiresAtTick: burn.ExpiresAt,
			Negative:      true,
		})
	}
	return buffs
}

func activeCombatBuffs(entity *Entity, tick uint64) []BuffState {
	buffs := make([]BuffState, 0, 4)
	if entity.Control.MoveSpeedBonusUntil > tick && entity.Control.MoveSpeedBonusFlat > 0 {
		buffs = append(buffs, BuffState{
			ID:            "move_speed_bonus",
			Name:          "移动速度提升",
			ExpiresAtTick: entity.Control.MoveSpeedBonusUntil,
		})
	}
	if entity.Control.AttackSpeedSlowUntil > tick && entity.Control.AttackSpeedSlow > 0 {
		buffs = append(buffs, BuffState{
			ID:            "attack_speed_slow",
			Name:          "攻速降低",
			ExpiresAtTick: entity.Control.AttackSpeedSlowUntil,
			Negative:      true,
		})
	}
	if entity.Combat.BlackCleaverUntil > tick && entity.Combat.BlackCleaverStacks > 0 {
		buffs = append(buffs, BuffState{
			ID:            "black_cleaver_shred",
			Name:          "切割 " + strconv.Itoa(entity.Combat.BlackCleaverStacks) + "层",
			ExpiresAtTick: entity.Combat.BlackCleaverUntil,
			Negative:      true,
		})
	}
	return buffs
}

func activeEquipmentSlotBuffs(entity *Entity, tick uint64) []BuffState {
	buffs := make([]BuffState, 0, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if equipped.EquipmentID == "" {
			continue
		}
		if equipped.StackExpireTick > tick && equipped.Stacks > 0 {
			buffs = append(buffs, BuffState{
				ID:            equipped.EquipmentID + "_stacks",
				Name:          equipmentBuffName(equipped.EquipmentID) + " " + strconv.Itoa(int(equipped.Stacks)) + "层",
				ExpiresAtTick: equipped.StackExpireTick,
			})
		}
		if equipped.SunfireActiveUntil > tick {
			name := "献祭"
			if equipped.Stacks > 0 && equipped.SunfireStackExpireTick > tick {
				name += " " + strconv.Itoa(int(equipped.Stacks)) + "层"
			}
			buffs = append(buffs, BuffState{
				ID:            equipped.EquipmentID + "_sunfire",
				Name:          name,
				ExpiresAtTick: equipped.SunfireActiveUntil,
			})
		}
		if equipped.PhysicalShieldAmount > 0 && equipped.PhysicalShieldExpireTick > tick {
			buffs = append(buffs, BuffState{
				ID:            equipped.EquipmentID + "_physical_shield",
				Name:          "锋焰护盾",
				ExpiresAtTick: equipped.PhysicalShieldExpireTick,
			})
		}
		if equipped.StoneplateShieldActive && equipped.StoneplateBreakTick > tick {
			buffs = append(buffs, BuffState{
				ID:            equipped.EquipmentID + "_stoneplate",
				Name:          "石像鬼护盾",
				ExpiresAtTick: equipped.StoneplateBreakTick,
			})
		}
	}
	return buffs
}

func equipmentBuffName(equipmentID string) string {
	switch equipmentID {
	case "guinsoos_rageblade":
		return "鬼索"
	default:
		return "装备"
	}
}
