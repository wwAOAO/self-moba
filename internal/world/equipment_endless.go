package world

func (w *World) triggerEquipmentHeroCombat(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !IsHeroUnit(source) || !IsHeroUnit(target) {
		return
	}
	w.triggerEndlessDespairCombat(source, tick, tickRate)
	w.triggerEndlessDespairCombat(target, tick, tickRate)
}

func (w *World) triggerEndlessDespairCombat(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.EndlessDespairRadius <= 0 {
			continue
		}
		seconds := item.Effects.EndlessDespairTickSeconds
		if seconds <= 0 {
			seconds = 4
		}
		duration := secondsToTicks(seconds, tickRate)
		if tick >= equipped.EndlessDespairActiveUntil {
			entity.Equipment[index].EndlessDespairNextTick = tick + duration
		}
		entity.Equipment[index].EndlessDespairActiveUntil = tick + duration
	}
}

func (w *World) tickEndlessDespair(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 || w.equipment == nil {
		return
	}
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.EndlessDespairRadius <= 0 {
			continue
		}
		if tick > equipped.EndlessDespairActiveUntil || equipped.EndlessDespairNextTick == 0 || tick < equipped.EndlessDespairNextTick {
			if tick > equipped.EndlessDespairActiveUntil {
				entity.Equipment[index].EndlessDespairNextTick = 0
			}
			continue
		}
		totalDamage := 0
		for _, target := range w.entities {
			if !IsHeroUnit(target) || !canAttackTarget(entity, target) || distance(entity.Position, target.Position) > item.Effects.EndlessDespairRadius+target.Radius {
				continue
			}
			raw := entity.Stats.BonusHP * item.Effects.EndlessDespairBonusHPRatio
			damage := magicDamageAfterResistance(entity, target, raw, tick)
			if damage <= 0 {
				continue
			}
			target.Combat.LastHitTick = tick
			wasAlive := target.Stats.HP > 0
			w.applyResolvedDamage(entity, target, damage, "magic", sustainEquipmentDamage, tickRate)
			totalDamage += target.Combat.LastDamage
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		}
		if totalDamage > 0 && item.Effects.EndlessDespairHealDamageRatio > 0 {
			beforeHP := entity.Stats.HP
			entity.Stats.HP += float64(totalDamage) * item.Effects.EndlessDespairHealDamageRatio
			if entity.Stats.HP > entity.Stats.MaxHP {
				entity.Stats.HP = entity.Stats.MaxHP
			}
			w.refreshPlayerStatsAfterHPChange(entity, beforeHP)
		}
		seconds := item.Effects.EndlessDespairTickSeconds
		if seconds <= 0 {
			seconds = 4
		}
		entity.Equipment[index].EndlessDespairNextTick = tick + secondsToTicks(seconds, tickRate)
	}
}
