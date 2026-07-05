package world

import (
	"l-battle/internal/config"
	"l-battle/internal/world/equipcalc"
	"math"
)

func (w *World) triggerSunfireCombat(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.SunfireBurnSeconds <= 0 {
			continue
		}
		entity.Equipment[index].SunfireActiveUntil = tick + secondsToTicks(item.Effects.SunfireBurnSeconds, tickRate)
		if entity.Equipment[index].SunfireNextTick == 0 || tick >= entity.Equipment[index].SunfireNextTick {
			entity.Equipment[index].SunfireNextTick = tick + secondsToTicks(1, tickRate)
		}
	}
}

func (w *World) tickSunfire(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 || w.equipment == nil {
		return
	}
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.SunfireBurnSeconds <= 0 || tick >= equipped.SunfireActiveUntil || tick < equipped.SunfireNextTick {
			continue
		}
		for _, target := range w.entities {
			if !canAttackTarget(entity, target) || distance(entity.Position, target.Position) > item.Effects.SunfireRadius+target.Radius {
				continue
			}
			raw := equipmentSunfireDamage(entity, item, equipped)
			if isMinion(target) {
				raw *= item.Effects.SunfireMinionMultiplier
			}
			if isMonster(target) {
				raw *= item.Effects.SunfireMonsterMultiplier
			}
			target.Combat.LastHitTick = tick
			wasAlive := target.Stats.HP > 0
			w.applyAOEDamage(entity, target, magicDamageAfterResistance(entity, target, raw, tick), "magic", tickRate)
			if target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero || target.Kind == EntityKindBaronNashor {
				entity.Equipment[index].Stacks = math.Min(item.Effects.SunfireMaxStacks, entity.Equipment[index].Stacks+1)
				entity.Equipment[index].SunfireStackExpireTick = tick + secondsToTicks(item.Effects.SunfireStackSeconds, tickRate)
			}
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		}
		entity.Equipment[index].SunfireNextTick = tick + secondsToTicks(1, tickRate)
	}
}

func equipmentSunfireDamage(entity *Entity, item config.EquipmentConfig, equipped EquipmentSlot) float64 {
	return equipcalc.SunfireDamage(entity.Level, MinHeroLevel, MaxHeroLevel, entity.Stats.BonusHP, item, equipped)
}
