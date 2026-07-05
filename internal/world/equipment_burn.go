package world

func (w *World) applyEquipmentSkillBurn(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.Kind != EntityKindPlayer || target.Stats.HP <= 0 || w.equipment == nil {
		return
	}
	seen := make(map[string]bool, len(source.Equipment))
	for _, equipped := range source.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.SkillBurnSeconds <= 0 {
			continue
		}
		tickSeconds := item.Effects.SkillBurnTickSeconds
		if tickSeconds <= 0 {
			tickSeconds = 1
		}
		key := source.ID + "->" + target.ID
		w.equipmentBurns[key] = EquipmentBurn{
			SourceID:           source.ID,
			TargetID:           target.ID,
			NextTick:           tick + secondsToTicks(tickSeconds, tickRate),
			ExpiresAt:          tick + secondsToTicks(item.Effects.SkillBurnSeconds, tickRate),
			FlatDamage:         item.Effects.SkillBurnFlatDamage,
			BaseMaxHPRatio:     item.Effects.SkillBurnBaseMaxHPRatio,
			APMaxHPRatioPer100: item.Effects.SkillBurnAPMaxHPRatioPer100AP,
		}
		return
	}
}

func (w *World) tickEquipmentBurns(tick uint64, tickRate int) {
	for key, burn := range w.equipmentBurns {
		source := w.entities[burn.SourceID]
		target := w.entities[burn.TargetID]
		if tick > burn.ExpiresAt || source == nil || target == nil || target.Stats.HP <= 0 {
			delete(w.equipmentBurns, key)
			continue
		}
		if tick < burn.NextTick {
			continue
		}
		rawDamage := burn.FlatDamage + float64(target.Stats.MaxHP)*(burn.BaseMaxHPRatio+float64(source.Stats.AbilityPower)/100*burn.APMaxHPRatioPer100)
		damage := magicDamageAfterResistance(source, target, rawDamage, tick)
		target.Combat.LastHitTick = tick
		wasAlive := target.Stats.HP > 0
		w.applyResolvedDamage(source, target, damage, "magic", sustainPetDamage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(source, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
		burn.NextTick += secondsToTicks(1, tickRate)
		if burn.NextTick > burn.ExpiresAt {
			delete(w.equipmentBurns, key)
			continue
		}
		w.equipmentBurns[key] = burn
	}
}
