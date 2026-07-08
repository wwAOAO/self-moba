package world

func (w *World) triggerEquipmentBasicAttackAttackerSlow(source *Entity, target *Entity, tickRate int) {
	if source == nil || target == nil || target.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	seen := make(map[string]bool, len(target.Equipment))
	for _, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.BasicAttackAttackerSlow <= 0 {
			continue
		}
		seconds := item.Effects.BasicAttackAttackerSlowSeconds
		if seconds <= 0 {
			seconds = 1
		}
		applyMoveSpeedSlow(source, item.Effects.BasicAttackAttackerSlow, target.Combat.LastHitTick+secondsToTicks(seconds, tickRate))
		return
	}
}

func (w *World) triggerEquipmentMagicHitStacks(target *Entity) {
	if target == nil || target.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	changed := false
	seen := make(map[string]bool, len(target.Equipment))
	for index, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.MagicHitMaxStacks <= 0 || equipped.Stacks >= item.Effects.MagicHitMaxStacks {
			continue
		}
		target.Equipment[index].Stacks++
		changed = true
	}
	if changed {
		w.recalculatePlayerStats(target)
	}
}

func (w *World) triggerEquipmentSkillDamageSlow(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.Kind != EntityKindPlayer || target.Stats.MaxHP <= 0 || target.Stats.HP <= 0 || w.equipment == nil {
		return
	}
	seen := make(map[string]bool, len(source.Equipment))
	for _, equipped := range source.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.SkillDamageLowHealthSlow <= 0 {
			continue
		}
		threshold := item.Effects.SkillDamageLowHealthSlowThreshold
		if threshold <= 0 {
			threshold = 0.5
		}
		if target.Stats.HP/target.Stats.MaxHP >= threshold {
			continue
		}
		seconds := item.Effects.SkillDamageLowHealthSlowSeconds
		if seconds <= 0 {
			seconds = 1
		}
		applyMoveSpeedSlow(target, item.Effects.SkillDamageLowHealthSlow, tick+secondsToTicks(seconds, tickRate))
		return
	}
}
