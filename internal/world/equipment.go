package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) buyEquipment(entity *Entity, equipmentID string, tick uint64) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil || equipmentID == "" {
		return
	}
	item, ok := w.equipment.Get(equipmentID)
	if !ok {
		return
	}
	if len(item.Components) > 0 {
		w.buyCompositeEquipment(entity, item, tick)
		return
	}
	if w.hasUniqueEquipmentGroup(entity, item.UniqueGroup) {
		w.setMessage(entity, "该类型装备只能装备一件", tick)
		return
	}
	if w.hasEquipmentCategory(entity, item.Category) && item.Category == "shoes" {
		w.setMessage(entity, "该类型装备只能装备一件", tick)
		return
	}
	if len(entity.Equipment) >= config.MaxEquipmentSlots {
		w.setMessage(entity, "装备栏已满", tick)
		return
	}
	if entity.Gold < float64(item.Price) {
		w.setMessage(entity, "金币不足", tick)
		return
	}
	w.addEquipment(entity, item, float64(item.Price))
}

func (w *World) buyCompositeEquipment(entity *Entity, item config.EquipmentConfig, tick uint64) {
	componentIndexes := w.findComponentIndexes(entity, item.Components)
	ownedCost := w.componentCost(entity, componentIndexes)
	combineCost := float64(item.Price - ownedCost)
	if combineCost < 0 {
		combineCost = 0
	}
	if len(componentIndexes) > 0 && entity.Gold >= combineCost {
		w.replaceComponentsWithEquipment(entity, item, componentIndexes, combineCost)
		return
	}
	if len(componentIndexes) == 0 && entity.Gold >= float64(item.Price) {
		if w.hasUniqueEquipmentGroup(entity, item.UniqueGroup) || (item.Category == "shoes" && w.hasEquipmentCategory(entity, "shoes")) {
			w.setMessage(entity, "该类型装备只能装备一件", tick)
			return
		}
		if len(entity.Equipment) >= config.MaxEquipmentSlots {
			w.setMessage(entity, "装备栏已满", tick)
			return
		}
		w.addEquipment(entity, item, float64(item.Price))
		return
	}
	if w.tryBuyMissingComponent(entity, item.Components, componentIndexes, tick) {
		return
	}
	w.setMessage(entity, "金币不足", tick)
}

func (w *World) addEquipment(entity *Entity, item config.EquipmentConfig, cost float64) {
	entity.Gold -= cost
	entity.Equipment = append(entity.Equipment, equipmentSlotFromConfig(item))
	w.recalculatePlayerStats(entity)
	w.refreshStoneplateShield(entity)
}

func (w *World) replaceComponentsWithEquipment(entity *Entity, item config.EquipmentConfig, componentIndexes []int, cost float64) {
	entity.Gold -= cost
	indexSet := make(map[int]bool, len(componentIndexes))
	for _, index := range componentIndexes {
		indexSet[index] = true
	}
	next := entity.Equipment[:0]
	for index, equipped := range entity.Equipment {
		if !indexSet[index] {
			next = append(next, equipped)
		}
	}
	entity.Equipment = append(next, equipmentSlotFromConfig(item))
	w.recalculatePlayerStats(entity)
	w.refreshStoneplateShield(entity)
}

func equipmentSlotFromConfig(item config.EquipmentConfig) EquipmentSlot {
	seconds := item.Effects.OutOfCombatSeconds
	if seconds <= 0 {
		seconds = 5
	}
	return EquipmentSlot{
		EquipmentID:              item.EquipmentID,
		Name:                     item.Name,
		LowHealthDamageReduce:    item.Effects.LowHealthDamageReduce,
		LowHealthShieldThreshold: item.Effects.LowHealthShieldThreshold,
		LowHealthShieldMin:       item.Effects.LowHealthShieldMin,
		LowHealthShieldMax:       item.Effects.LowHealthShieldMax,
		OutOfCombatMoveSpeed:     item.Effects.OutOfCombatMoveSpeed,
		OutOfCombatRequiredTicks: uint64(math.Ceil(seconds * 20)),
		StoneplateShieldRatio:    item.Effects.StoneplateShieldMaxHPRatio,
		StoneplateResistPercent:  item.Effects.StoneplateResistPercent,
		StoneplateCooldownTicks:  uint64(math.Ceil(item.Effects.StoneplateCooldownSeconds * 20)),
	}
}

func (w *World) findComponentIndexes(entity *Entity, components []string) []int {
	if entity == nil || len(components) == 0 {
		return nil
	}
	used := make(map[int]bool, len(components))
	indexes := make([]int, 0, len(components))
	for _, componentID := range components {
		for index, equipped := range entity.Equipment {
			if used[index] || equipped.EquipmentID != componentID {
				continue
			}
			used[index] = true
			indexes = append(indexes, index)
			break
		}
	}
	return indexes
}

func (w *World) componentCost(entity *Entity, componentIndexes []int) int {
	cost := 0
	for _, index := range componentIndexes {
		if index < 0 || index >= len(entity.Equipment) {
			continue
		}
		item, ok := w.equipment.Get(entity.Equipment[index].EquipmentID)
		if ok {
			cost += item.Price
		}
	}
	return cost
}

func (w *World) tryBuyMissingComponent(entity *Entity, components []string, ownedIndexes []int, tick uint64) bool {
	if len(entity.Equipment) >= config.MaxEquipmentSlots {
		w.setMessage(entity, "装备栏已满", tick)
		return true
	}
	owned := make(map[string]int, len(ownedIndexes))
	for _, index := range ownedIndexes {
		if index >= 0 && index < len(entity.Equipment) {
			owned[entity.Equipment[index].EquipmentID]++
		}
	}
	for _, componentID := range components {
		if owned[componentID] > 0 {
			owned[componentID]--
			continue
		}
		component, ok := w.equipment.Get(componentID)
		if !ok {
			continue
		}
		if w.hasUniqueEquipmentGroup(entity, component.UniqueGroup) || (component.Category == "shoes" && w.hasEquipmentCategory(entity, "shoes")) {
			continue
		}
		if entity.Gold >= float64(component.Price) {
			w.addEquipment(entity, component, float64(component.Price))
			return true
		}
	}
	return false
}

func (w *World) sellEquipment(entity *Entity, slot int) {
	if entity == nil || entity.Kind != EntityKindPlayer || w.equipment == nil {
		return
	}
	index := slot - 1
	if index < 0 || index >= len(entity.Equipment) {
		return
	}
	equipped := entity.Equipment[index]
	item, ok := w.equipment.Get(equipped.EquipmentID)
	if !ok {
		return
	}
	entity.Gold += math.Floor(float64(item.Price)*item.SellRatio + 0.000000001)
	removeStoneplateShieldFromSlot(entity, index)
	entity.Equipment = append(entity.Equipment[:index], entity.Equipment[index+1:]...)
	w.recalculatePlayerStats(entity)
	w.refreshStoneplateShield(entity)
}

func (w *World) recalculatePlayerStats(entity *Entity) {
	if entity == nil || entity.Kind != EntityKindPlayer {
		return
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return
	}
	oldMaxHP := entity.Stats.MaxHP
	oldMaxMP := entity.Stats.MaxMP
	oldHP := entity.Stats.HP
	oldMP := entity.Stats.MP
	nextStats := heroStatsAtLevel(hero, entity.Level)
	w.applyEquipmentStats(entity, &nextStats)
	w.applySwordCritOverflowStats(entity, &nextStats)
	nextStats.AbilityHaste += abilityHasteFromBuffs(entity)
	nextStats.HP = oldHP + (nextStats.MaxHP - oldMaxHP)
	if nextStats.HP > nextStats.MaxHP {
		nextStats.HP = nextStats.MaxHP
	}
	if nextStats.HP < 1 {
		nextStats.HP = 1
	}
	nextStats.MP = oldMP + (nextStats.MaxMP - oldMaxMP)
	if nextStats.MP > nextStats.MaxMP {
		nextStats.MP = nextStats.MaxMP
	}
	if nextStats.MP < 0 {
		nextStats.MP = 0
	}
	entity.Tank.ThunderclapArmorBonus = 0
	entity.Stats = nextStats
	w.refreshTankGraniteShieldMax(entity)
	w.refreshTankWPassive(entity)
}

func (w *World) applyEquipmentStats(entity *Entity, stats *Stats) {
	if entity == nil || stats == nil || w.equipment == nil {
		return
	}
	baseMPRegen5 := stats.MPRegen5
	baseMPRegenBonus := 0.0
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if !ok {
			continue
		}
		stats.MaxHP += item.Stats.HP
		stats.BonusHP += item.Stats.HP
		stats.MaxMP += item.Stats.MP
		stats.HPRegen5 += item.Stats.HPRegen5
		stats.MPRegen5 += item.Stats.MPRegen5
		baseMPRegenBonus += item.Stats.BaseMPRegenBonus
		stats.Attack += item.Stats.Attack
		stats.BonusAttack += item.Stats.Attack
		stats.AbilityPower += item.Stats.AbilityPower
		stats.AbilityHaste += item.Stats.AbilityHaste
		stats.PhysicalDefense += item.Stats.PhysicalDefense
		stats.BonusPhysicalDefense += item.Stats.PhysicalDefense
		stats.MagicDefense += item.Stats.MagicDefense
		stats.BonusMagicDefense += item.Stats.MagicDefense
		stats.MagicPenFlat += item.Stats.MagicPenFlat
		stats.Tenacity += item.Stats.Tenacity
		stats.SlowResist += item.Stats.SlowResist
		stats.BasicAttackBlock += item.Stats.BasicAttackBlock
		stats.CritDamageReduce += item.Stats.CritDamageReduce
		stats.PhysicalDefense += equipped.Stacks * item.Effects.UnitKillPhysicalDefenseGain
		stats.BonusPhysicalDefense += equipped.Stacks * item.Effects.UnitKillPhysicalDefenseGain
		stats.AbilityPower += int(equipped.Stacks * item.Effects.UnitKillAbilityPowerGain)
		stats.MagicDefense += equipped.Stacks * item.Effects.MagicHitMagicDefensePerStack
		stats.BonusMagicDefense += equipped.Stacks * item.Effects.MagicHitMagicDefensePerStack
		stats.MoveSpeed += item.Stats.MoveSpeed
		stats.MoveSpeed *= 1 + equipped.Stacks*item.Effects.MagicHitMoveSpeedPercentPerStack
		stats.MoveSpeed *= 1 + item.Stats.MoveSpeedPercent
		stats.AttackSpeedBonus += item.Stats.AttackSpeedBonus
		stats.AttackSpeedBonus += equipped.Stacks * item.Effects.BasicAttackAttackSpeedPerStack
		stats.CritChance += item.Stats.CritChance
		stats.Omnivamp += item.Stats.Omnivamp
		stats.LifeSteal += item.Stats.LifeSteal
	}
	if equipmentZeroesCritChance(entity, w.equipment) {
		stats.CritChance = 0
	}
	stats.MPRegen5 += baseMPRegen5 * baseMPRegenBonus
	stats.AbilityPower = int(math.Round(float64(stats.AbilityPower) * (1 + equipmentAbilityPowerMultiplier(entity, w.equipment))))
	w.applyStoneplateResists(entity, stats)
	stats.AttackSpeed = finalAttackSpeed(stats.BaseAttackSpeed, stats.AttackSpeedBonus, stats.AttackSpeedRatio, stats.AttackSpeedSlow)
}

func (w *World) applySwordCritOverflowStats(entity *Entity, stats *Stats) {
	if entity == nil || stats == nil || entity.HeroID != swordHeroID {
		return
	}
	skill := w.heroPassiveSkill(entity)
	effectiveCrit := stats.CritChance * skillMetaRange(skill, "critChanceMultiplier", 2)
	if effectiveCrit <= 1 {
		return
	}
	bonusAttack := (effectiveCrit - 1) * 100 * skillMetaRange(skill, "critOverflowAttackPerPercent", 0.5)
	stats.Attack += bonusAttack
	stats.BonusAttack += bonusAttack
}

func (w *World) hasUniqueEquipmentGroup(entity *Entity, group string) bool {
	if entity == nil || group == "" || w.equipment == nil {
		return false
	}
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.UniqueGroup == group {
			return true
		}
	}
	return false
}

func (w *World) hasEquipmentCategory(entity *Entity, category string) bool {
	if entity == nil || category == "" || w.equipment == nil {
		return false
	}
	for _, equipped := range entity.Equipment {
		item, ok := w.equipment.Get(equipped.EquipmentID)
		if ok && item.Category == category {
			return true
		}
	}
	return false
}

func (w *World) setMessage(entity *Entity, message string, tick uint64) {
	if entity == nil {
		return
	}
	entity.Message = message
	entity.MessageTick = tick
}

func equipmentOutOfCombatMoveSpeed(entity *Entity, tick uint64) float64 {
	if entity == nil || tick == 0 {
		return 0
	}
	var bonus float64
	seen := make(map[string]bool, len(entity.Equipment))
	for _, equipped := range entity.Equipment {
		if seen[equipped.EquipmentID] {
			continue
		}
		seen[equipped.EquipmentID] = true
		if equipped.OutOfCombatMoveSpeed <= 0 || equipped.OutOfCombatRequiredTicks == 0 {
			continue
		}
		if tick >= entity.Combat.LastHitTick+equipped.OutOfCombatRequiredTicks {
			bonus += equipped.OutOfCombatMoveSpeed
		}
	}
	return bonus
}

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
	for _, equipped := range source.Equipment {
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
			source.Passive.Shield += int(math.Floor(float64(damage) * item.Effects.PhysicalDamageShieldRatio))
			source.Passive.MaxShield = source.Passive.Shield
		}
	}
}

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

type sustainContext struct {
	BasicAttack bool
	AOE         bool
	Pet         bool
}

var (
	sustainSingleTargetSkill = sustainContext{}
	sustainBasicAttack       = sustainContext{BasicAttack: true}
	sustainAOESkill          = sustainContext{AOE: true}
	sustainPetDamage         = sustainContext{Pet: true}
)

func (w *World) applySustain(source *Entity, actualDamage int, context sustainContext) {
	if source == nil || source.Kind != EntityKindPlayer || actualDamage <= 0 || source.Stats.HP <= 0 {
		return
	}
	ratio := source.Stats.Omnivamp
	if context.BasicAttack {
		ratio += source.Stats.LifeSteal
	}
	if ratio <= 0 {
		return
	}
	decay := 1.0
	if context.AOE || context.Pet {
		decay = 0.33
	}
	healValue := float64(actualDamage) * ratio * decay * (1 + source.Stats.HealingPower) * (1 - clamp(source.Stats.GrievousWounds, 0, 1))
	heal := int(math.Floor(healValue + 0.000000001))
	if heal <= 0 {
		return
	}
	source.Stats.HP += heal
	if source.Stats.HP > source.Stats.MaxHP {
		source.Stats.HP = source.Stats.MaxHP
	}
}

func isHeroDamageSource(entity *Entity) bool {
	return entity != nil && (entity.Kind == EntityKindPlayer || entity.Kind == EntityKindEnemyHero)
}
