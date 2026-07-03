package world

import (
	"l-battle/internal/config"
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
	level := clampInt(entity.Level, MinHeroLevel, MaxHeroLevel)
	base := item.Effects.SunfireBurnFlatMin
	if item.Effects.SunfireBurnFlatMax > base {
		base += (item.Effects.SunfireBurnFlatMax - base) * float64(level-MinHeroLevel) / float64(MaxHeroLevel-MinHeroLevel)
	}
	return (base + float64(entity.Stats.BonusHP)*item.Effects.SunfireBurnBonusHPRatio) * (1 + equipped.Stacks*item.Effects.SunfireStackDamageBonus)
}

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

func (w *World) tickEquipmentPercentRegen(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 || tickRate <= 0 || tick%uint64(5*tickRate) != 0 || w.equipment == nil {
		return
	}
	hpRatio, mpRatio := w.equipmentPercentRegenRatios(entity, tick, tickRate)
	if hpRatio > 0 && entity.Stats.HP < entity.Stats.MaxHP {
		entity.Stats.HP += int(math.Floor(float64(entity.Stats.MaxHP) * hpRatio))
		if entity.Stats.HP > entity.Stats.MaxHP {
			entity.Stats.HP = entity.Stats.MaxHP
		}
	}
	if mpRatio > 0 && entity.Stats.MP < entity.Stats.MaxMP {
		entity.Stats.MP += entity.Stats.MaxMP * mpRatio
		if entity.Stats.MP > entity.Stats.MaxMP {
			entity.Stats.MP = entity.Stats.MaxMP
		}
	}
}

func (w *World) equipmentPercentRegenRatios(entity *Entity, tick uint64, tickRate int) (float64, float64) {
	outOfCombat := tick >= entity.Combat.LastHitTick+uint64(5*tickRate)
	hpRatio, mpRatio := 0.0, 0.0
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if outOfCombat {
			hpRatio += item.Effects.OutOfCombatHPRegenMaxHPRatio5
			mpRatio += item.Effects.OutOfCombatMPRegenMaxMPRatio5
			continue
		}
		hpRatio += item.Effects.CombatHPRegenMaxHPRatio5
		mpRatio += item.Effects.CombatMPRegenMaxMPRatio5
	}
	return hpRatio, mpRatio
}

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

func (w *World) refreshStoneplateShield(entity *Entity) {
	if entity == nil || entity.Kind != EntityKindPlayer {
		return
	}
	for index := range entity.Equipment {
		equipped := &entity.Equipment[index]
		if equipped.StoneplateShieldRatio <= 0 || equipped.StoneplateShieldActive || equipped.StoneplateCooldownUntil > 0 {
			continue
		}
		shield := int(math.Round(float64(entity.Stats.MaxHP) * equipped.StoneplateShieldRatio))
		if shield <= 0 {
			continue
		}
		equipped.StoneplateShieldActive = true
		equipped.StoneplateShieldAmount = shield
		entity.Passive.Shield += shield
		entity.Passive.MaxShield += shield
		w.recalculatePlayerStats(entity)
		return
	}
}

func (w *World) tickStoneplateShield(entity *Entity, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer {
		return
	}
	for index := range entity.Equipment {
		equipped := &entity.Equipment[index]
		if equipped.StoneplateShieldActive && equipped.StoneplateBreakTick > 0 && tick >= equipped.StoneplateBreakTick {
			targetShield := equipped.StoneplateShieldAmount
			if targetShield > entity.Passive.Shield {
				targetShield = entity.Passive.Shield
			}
			entity.Passive.Shield -= targetShield
			if entity.Passive.Shield < 0 {
				entity.Passive.Shield = 0
			}
			deactivateStoneplateShield(entity)
			w.recalculatePlayerStats(entity)
		}
		if equipped.StoneplateShieldRatio <= 0 || equipped.StoneplateShieldActive || equipped.StoneplateCooldownUntil == 0 || tick < equipped.StoneplateCooldownUntil {
			continue
		}
		equipped.StoneplateCooldownUntil = 0
		equipped.StoneplateBreakTick = 0
	}
	w.refreshStoneplateShield(entity)
}

func (w *World) triggerStoneplateCooldown(target *Entity, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer {
		return
	}
	tick := target.Combat.LastHitTick
	for index := range target.Equipment {
		equipped := &target.Equipment[index]
		if !equipped.StoneplateShieldActive || equipped.StoneplateCooldownTicks == 0 || equipped.StoneplateCooldownUntil > tick {
			continue
		}
		equipped.StoneplateCooldownUntil = tick + equipped.StoneplateCooldownTicks
		equipped.StoneplateBreakTick = tick + secondsToTicks(5, tickRate)
		if tickRate <= 0 {
			equipped.StoneplateCooldownUntil = tick + 2400
			equipped.StoneplateBreakTick = tick + 100
		}
	}
}

func deactivateStoneplateShield(target *Entity) bool {
	if target == nil || target.Kind != EntityKindPlayer {
		return false
	}
	changed := false
	for index := range target.Equipment {
		equipped := &target.Equipment[index]
		if equipped.StoneplateShieldActive && target.Passive.Shield <= 0 {
			equipped.StoneplateShieldActive = false
			equipped.StoneplateShieldAmount = 0
			equipped.StoneplateBreakTick = 0
			changed = true
		}
	}
	if changed {
		target.Passive.MaxShield = target.Passive.Shield
	}
	return changed
}

func removeStoneplateShieldFromSlot(entity *Entity, index int) {
	if entity == nil || index < 0 || index >= len(entity.Equipment) {
		return
	}
	equipped := &entity.Equipment[index]
	if !equipped.StoneplateShieldActive || equipped.StoneplateShieldAmount <= 0 {
		return
	}
	removed := equipped.StoneplateShieldAmount
	if removed > entity.Passive.Shield {
		removed = entity.Passive.Shield
	}
	entity.Passive.Shield -= removed
	if entity.Passive.Shield < 0 {
		entity.Passive.Shield = 0
	}
	entity.Passive.MaxShield = entity.Passive.Shield
	equipped.StoneplateShieldActive = false
	equipped.StoneplateShieldAmount = 0
	equipped.StoneplateBreakTick = 0
}

func removePhysicalDamageShieldFromSlot(entity *Entity, index int) {
	if entity == nil || index < 0 || index >= len(entity.Equipment) {
		return
	}
	removed := entity.Equipment[index].PhysicalShieldAmount
	if removed <= 0 {
		return
	}
	if removed > entity.Passive.Shield {
		removed = entity.Passive.Shield
	}
	entity.Passive.Shield -= removed
	if entity.Passive.Shield < 0 {
		entity.Passive.Shield = 0
	}
	if entity.Passive.MaxShield > entity.Passive.Shield {
		entity.Passive.MaxShield = entity.Passive.Shield
	}
	entity.Equipment[index].PhysicalShieldAmount = 0
	entity.Equipment[index].PhysicalShieldMaxAmount = 0
}

func (w *World) applyStoneplateResists(entity *Entity, stats *Stats) {
	if entity == nil || stats == nil {
		return
	}
	for _, equipped := range entity.Equipment {
		if !equipped.StoneplateShieldActive || equipped.StoneplateResistPercent <= 0 {
			continue
		}
		stats.PhysicalDefense *= 1 + equipped.StoneplateResistPercent
		stats.MagicDefense *= 1 + equipped.StoneplateResistPercent
		return
	}
}
