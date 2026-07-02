package world

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
