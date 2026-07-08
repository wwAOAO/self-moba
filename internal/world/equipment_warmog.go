package world

import "math"

func (w *World) triggerEquipmentDamageTaken(target *Entity, source *Entity, tick uint64, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	for index, equipped := range target.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.WarmogHealMaxHPRatio <= 0 {
			continue
		}
		seconds := item.Effects.WarmogOutOfCombatSeconds
		if isMinion(source) || isMonster(source) {
			seconds = item.Effects.WarmogMinionOutOfCombatSeconds
		}
		if seconds <= 0 {
			seconds = 8
		}
		target.Equipment[index].WarmogRequiredTicks = secondsToTicks(seconds, tickRate)
		target.Equipment[index].WarmogNextTick = 0
	}
}

func (w *World) tickWarmog(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 || entity.Stats.HP >= entity.Stats.MaxHP || w.equipment == nil {
		return
	}
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.WarmogHealMaxHPRatio <= 0 {
			continue
		}
		requiredBonusHP := item.Effects.WarmogRequiredBonusHP
		if requiredBonusHP <= 0 {
			requiredBonusHP = 2000
		}
		if entity.Stats.BonusHP < requiredBonusHP {
			continue
		}
		requiredTicks := equipped.WarmogRequiredTicks
		if requiredTicks == 0 {
			seconds := item.Effects.WarmogOutOfCombatSeconds
			if seconds <= 0 {
				seconds = 8
			}
			requiredTicks = secondsToTicks(seconds, tickRate)
			entity.Equipment[index].WarmogRequiredTicks = requiredTicks
		}
		if tick < entity.Combat.LastHitTick+requiredTicks {
			continue
		}
		if equipped.WarmogNextTick == 0 {
			entity.Equipment[index].WarmogNextTick = tick
		}
		if tick < entity.Equipment[index].WarmogNextTick {
			continue
		}
		beforeHP := entity.Stats.HP
		entity.Stats.HP += entity.Stats.MaxHP * item.Effects.WarmogHealMaxHPRatio
		if entity.Stats.HP > entity.Stats.MaxHP {
			entity.Stats.HP = entity.Stats.MaxHP
		}
		seconds := item.Effects.WarmogTickSeconds
		if seconds <= 0 {
			seconds = 0.5
		}
		entity.Equipment[index].WarmogNextTick = tick + uint64(math.Ceil(seconds*float64(tickRate)))
		w.refreshPlayerStatsAfterHPChange(entity, beforeHP)
		return
	}
}
