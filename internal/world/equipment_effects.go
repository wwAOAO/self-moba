package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) applyEquipmentLevelUpRestore(entity *Entity) {
	if entity == nil || w.equipment == nil {
		return
	}
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
		if item.Effects.LevelUpRestoreHPRatio > 0 {
			entity.Stats.HP += int(math.Floor(float64(entity.Stats.MaxHP) * item.Effects.LevelUpRestoreHPRatio))
			if entity.Stats.HP > entity.Stats.MaxHP {
				entity.Stats.HP = entity.Stats.MaxHP
			}
		}
		if item.Effects.LevelUpRestoreMPRatio > 0 {
			entity.Stats.MP += entity.Stats.MaxMP * item.Effects.LevelUpRestoreMPRatio
			if entity.Stats.MP > entity.Stats.MaxMP {
				entity.Stats.MP = entity.Stats.MaxMP
			}
		}
	}
}

func (w *World) applyEquipmentUnitKillGrowth(killer *Entity, target *Entity) {
	if killer == nil || target == nil || w.equipment == nil || killer.Kind != EntityKindPlayer {
		return
	}
	changed := false
	seen := make(map[string]bool, len(killer.Equipment))
	for index, equipped := range killer.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || (item.Effects.UnitKillPhysicalDefenseGain <= 0 && item.Effects.UnitKillAbilityPowerGain <= 0) {
			continue
		}
		maxGain := item.Effects.UnitKillMaxGain
		if maxGain <= 0 {
			continue
		}
		currentGain := math.Max(
			equipped.Stacks*item.Effects.UnitKillPhysicalDefenseGain,
			equipped.Stacks*item.Effects.UnitKillAbilityPowerGain,
		)
		if currentGain >= maxGain {
			continue
		}
		nextStacks := equipped.Stacks + 1
		if nextStacks*item.Effects.UnitKillPhysicalDefenseGain > maxGain ||
			nextStacks*item.Effects.UnitKillAbilityPowerGain > maxGain {
			nextStacks = math.Floor(maxGain / math.Max(item.Effects.UnitKillPhysicalDefenseGain, item.Effects.UnitKillAbilityPowerGain))
		}
		killer.Equipment[index].Stacks = nextStacks
		changed = true
	}
	if changed {
		w.recalculatePlayerStats(killer)
	}
}

func (w *World) equipmentBasicAttackBonus(attacker *Entity, damageType string) float64 {
	if attacker == nil || w.equipment == nil {
		return 0
	}
	var bonus float64
	seen := make(map[string]bool, len(attacker.Equipment))
	for _, equipped := range attacker.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.BasicAttackBonusDamageType == damageType {
			bonus += item.Effects.BasicAttackBonusDamage
		}
		if damageType == "magic" && item.Effects.BasicAttackBonusByCritMax > 0 {
			onHit := math.Min(item.Effects.BasicAttackBonusByCritMax, w.rawEquipmentCritChance(attacker)*item.Effects.BasicAttackBonusByCritMax)
			bonus += onHit
			if item.Effects.EveryNthBasicAttackBonusHit > 0 && equipped.RagebladeHitCount > 0 && equipped.RagebladeHitCount%int(item.Effects.EveryNthBasicAttackBonusHit) == 0 {
				bonus += item.Effects.BasicAttackBonusDamage + onHit
			}
		}
	}
	return bonus
}

func (w *World) equipmentMinionBasicAttackBonus(attacker *Entity, damageType string) float64 {
	if attacker == nil || w.equipment == nil {
		return 0
	}
	var bonus float64
	seen := make(map[string]bool, len(attacker.Equipment))
	for _, equipped := range attacker.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.MinionBasicAttackBonusDamageType == damageType {
			bonus += item.Effects.MinionBasicAttackBonusDamage
		}
	}
	return bonus
}

func (w *World) equipmentCritDamageBonus(attacker *Entity) float64 {
	if attacker == nil || w.equipment == nil {
		return 0
	}
	bonus := 0.0
	seen := make(map[string]bool, len(attacker.Equipment))
	for _, equipped := range attacker.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.CritDamageBonus <= bonus {
			continue
		}
		bonus = item.Effects.CritDamageBonus
	}
	return bonus
}

func (w *World) triggerEquipmentLowHealthShield(target *Entity, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer || target.Stats.MaxHP <= 0 || target.Passive.Shield > 0 {
		return
	}
	seen := make(map[string]bool, len(target.Equipment))
	for index, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		if equipped.LowHealthShieldUsed || equipped.LowHealthShieldMax <= 0 {
			continue
		}
		threshold := equipped.LowHealthShieldThreshold
		if threshold <= 0 {
			threshold = 0.3
		}
		if float64(target.Stats.HP)/float64(target.Stats.MaxHP) > threshold {
			continue
		}
		shield := equipmentShieldByLevel(target.Level, equipped.LowHealthShieldMin, equipped.LowHealthShieldMax)
		if shield <= 0 {
			continue
		}
		target.Equipment[index].LowHealthShieldUsed = true
		target.Passive.Shield = shield
		target.Passive.MaxShield = shield
		target.Passive.ShieldExpireTick = 0
		return
	}
}

func equipmentShieldByLevel(level int, minShield int, maxShield int) int {
	if maxShield <= minShield {
		return minShield
	}
	level = clampInt(level, MinHeroLevel, MaxHeroLevel)
	progress := float64(level-MinHeroLevel) / float64(MaxHeroLevel-MinHeroLevel)
	return int(math.Round(float64(minShield) + float64(maxShield-minShield)*progress))
}

func equipmentLowHealthDamageReduce(entity *Entity) float64 {
	if entity == nil || entity.Passive.Shield <= 0 {
		return 0
	}
	reduce := 0.0
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		if equipped.LowHealthShieldUsed && equipped.LowHealthDamageReduce > reduce {
			reduce = equipped.LowHealthDamageReduce
		}
	}
	return reduce
}

func (w *World) triggerEquipmentHeroHitHeal(source *Entity, target *Entity) {
	if source == nil || target == nil || target.Kind != EntityKindPlayer || !isHeroDamageSource(source) || w.equipment == nil {
		return
	}
	var heal int
	seen := make(map[string]bool, len(target.Equipment))
	for _, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || !item.Effects.HeroHitSmallHeal {
			continue
		}
		heal += item.Effects.HeroHitHeal
	}
	if heal <= 0 || target.Stats.HP <= 0 {
		return
	}
	target.Stats.HP += heal
	if target.Stats.HP > target.Stats.MaxHP {
		target.Stats.HP = target.Stats.MaxHP
	}
}

func (w *World) triggerEquipmentHeroDamageManaShield(source *Entity, target *Entity, tickRate int) {
	if source == nil || target == nil || target.Kind != EntityKindPlayer || target.Stats.MP <= 0 || !isHeroDamageSource(source) || w.equipment == nil {
		return
	}
	tick := target.Combat.LastHitTick
	seen := make(map[string]bool, len(target.Equipment))
	for index, equipped := range target.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.HeroDamageManaShieldRatio <= 0 || tick < equipped.ManaShieldCooldownUntil {
			continue
		}
		shield := int(math.Floor(target.Stats.MP * item.Effects.HeroDamageManaShieldRatio))
		if shield <= 0 {
			return
		}
		target.Stats.MP -= float64(shield)
		target.Passive.Shield += shield
		target.Passive.MaxShield += shield
		cooldownTicks := uint64(math.Ceil(float64(item.Effects.HeroDamageManaShieldCooldownMS) / 1000 * float64(tickRate)))
		target.Equipment[index].ManaShieldCooldownUntil = tick + cooldownTicks
		return
	}
}

func equipmentAbilityPowerMultiplier(entity *Entity, equipment *config.EquipmentStore) float64 {
	if entity == nil || equipment == nil {
		return 0
	}
	multiplier := 0.0
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		item, ok := equipment.Get(equipped.EquipmentID)
		if ok && item.Effects.AbilityPowerMultiplier > multiplier {
			multiplier = item.Effects.AbilityPowerMultiplier
		}
	}
	return multiplier
}

func equipmentZeroesCritChance(entity *Entity, equipment *config.EquipmentStore) bool {
	if entity == nil || equipment == nil {
		return false
	}
	for _, equipped := range entity.Equipment {
		item, ok := equipment.Get(equipped.EquipmentID)
		if ok && item.Effects.ZeroCritChance {
			return true
		}
	}
	return false
}

func (w *World) rawEquipmentCritChance(entity *Entity) float64 {
	if entity == nil || w.equipment == nil {
		return 0
	}
	chance := entity.Stats.CritChance
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.Effects.ZeroCritChance {
			chance += item.Stats.CritChance
		}
	}
	return clamp(chance, 0, 1)
}

func (w *World) triggerEquipmentBasicAttackStacks(source *Entity, tick uint64, tickRate int) {
	if source == nil || source.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	changed := false
	for index, equipped := range source.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.BasicAttackAttackSpeedMaxStacks <= 0 {
			continue
		}
		if equipped.Stacks < item.Effects.BasicAttackAttackSpeedMaxStacks {
			source.Equipment[index].Stacks++
			changed = true
		}
		source.Equipment[index].RagebladeHitCount++
		source.Equipment[index].StackExpireTick = tick + secondsToTicks(5, tickRate)
	}
	if changed {
		w.recalculatePlayerStats(source)
	}
}

func (w *World) tickEquipmentStacks(entity *Entity, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	changed := false
	for index, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok || item.Effects.BasicAttackAttackSpeedMaxStacks <= 0 || equipped.StackExpireTick == 0 || tick < equipped.StackExpireTick {
			continue
		}
		entity.Equipment[index].Stacks = 0
		entity.Equipment[index].RagebladeHitCount = 0
		entity.Equipment[index].StackExpireTick = 0
		changed = true
	}
	if changed {
		w.recalculatePlayerStats(entity)
	}
}

func (w *World) triggerEquipmentPhysicalDamageEffects(source *Entity, target *Entity, damage int, tick uint64, tickRate int) {
	if source == nil || target == nil || source.Kind != EntityKindPlayer || damage <= 0 || w.equipment == nil {
		return
	}
	for index, equipped := range source.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		if item.Effects.PhysicalHitArmorShredMaxStacks > 0 && (target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero) {
			target.Combat.BlackCleaverStacks++
			if target.Combat.BlackCleaverStacks > int(item.Effects.PhysicalHitArmorShredMaxStacks) {
				target.Combat.BlackCleaverStacks = int(item.Effects.PhysicalHitArmorShredMaxStacks)
			}
			target.Combat.BlackCleaverUntil = tick + secondsToTicks(item.Effects.PhysicalHitArmorShredSeconds, tickRate)
		}
		if item.Effects.PhysicalHitMoveSpeed > 0 {
			bonus := item.Effects.PhysicalHitMoveSpeed
			if isRangedBasicAttacker(source) {
				bonus *= 0.5
			}
			source.Control.MoveSpeedBonusFlat = bonus
			source.Control.MoveSpeedBonusUntil = tick + secondsToTicks(item.Effects.PhysicalHitMoveSpeedSeconds, tickRate)
		}
		if item.Effects.PhysicalDamageShieldRatio > 0 {
			shield := int(math.Floor(float64(damage) * item.Effects.PhysicalDamageShieldRatio))
			if shield <= 0 {
				continue
			}
			seconds := item.Effects.PhysicalDamageShieldDecaySeconds
			if seconds <= 0 {
				seconds = 3
			}
			source.Passive.Shield += shield
			source.Passive.MaxShield += shield
			source.Equipment[index].PhysicalShieldMaxAmount += shield
			source.Equipment[index].PhysicalShieldAmount += shield
			source.Equipment[index].PhysicalShieldStartTick = tick
			source.Equipment[index].PhysicalShieldExpireTick = tick + secondsToTicks(seconds, tickRate)
		}
	}
}

func tickEquipmentPhysicalDamageShield(entity *Entity, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Passive.Shield <= 0 {
		return
	}
	decayed := 0
	for index := range entity.Equipment {
		equipped := &entity.Equipment[index]
		if equipped.PhysicalShieldAmount <= 0 {
			continue
		}
		if tick >= equipped.PhysicalShieldExpireTick || equipped.PhysicalShieldExpireTick <= equipped.PhysicalShieldStartTick {
			decayed += equipped.PhysicalShieldAmount
			equipped.PhysicalShieldMaxAmount = 0
			equipped.PhysicalShieldAmount = 0
			continue
		}
		remaining := float64(equipped.PhysicalShieldExpireTick - tick)
		duration := float64(equipped.PhysicalShieldExpireTick - equipped.PhysicalShieldStartTick)
		next := int(math.Ceil(float64(equipped.PhysicalShieldMaxAmount) * remaining / duration))
		if next < equipped.PhysicalShieldAmount {
			decayed += equipped.PhysicalShieldAmount - next
			equipped.PhysicalShieldAmount = next
		}
	}
	if decayed > entity.Passive.Shield {
		decayed = entity.Passive.Shield
	}
	entity.Passive.Shield -= decayed
	if entity.Passive.MaxShield > entity.Passive.Shield {
		entity.Passive.MaxShield = entity.Passive.Shield
	}
}

func consumeEquipmentPhysicalDamageShield(entity *Entity, absorbed int) {
	if entity == nil || absorbed <= 0 {
		return
	}
	remaining := absorbed
	for index := range entity.Equipment {
		if remaining <= 0 {
			return
		}
		equipped := &entity.Equipment[index]
		if equipped.PhysicalShieldAmount <= 0 {
			continue
		}
		if equipped.PhysicalShieldAmount <= remaining {
			remaining -= equipped.PhysicalShieldAmount
			equipped.PhysicalShieldMaxAmount = 0
			equipped.PhysicalShieldAmount = 0
			continue
		}
		equipped.PhysicalShieldAmount -= remaining
		equipped.PhysicalShieldMaxAmount -= remaining
		return
	}
}
