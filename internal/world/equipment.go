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
	if w.hasUniqueEquipmentGroup(entity, item.UniqueGroup) {
		w.setMessage(entity, "该类型装备只能装备一件", tick)
		return
	}
	if len(item.Components) > 0 {
		w.buyCompositeEquipment(entity, item, tick)
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
		if w.hasUniqueEquipmentGroup(entity, component.UniqueGroup) {
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
	entity.Equipment = append(entity.Equipment[:index], entity.Equipment[index+1:]...)
	w.recalculatePlayerStats(entity)
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
		stats.Attack += item.Stats.Attack
		stats.BonusAttack += item.Stats.Attack
		stats.AbilityPower += item.Stats.AbilityPower
		stats.AbilityHaste += item.Stats.AbilityHaste
		stats.PhysicalDefense += item.Stats.PhysicalDefense
		stats.BonusPhysicalDefense += item.Stats.PhysicalDefense
		stats.MagicDefense += item.Stats.MagicDefense
		stats.BonusMagicDefense += item.Stats.MagicDefense
		stats.PhysicalDefense += equipped.Stacks * item.Effects.UnitKillPhysicalDefenseGain
		stats.BonusPhysicalDefense += equipped.Stacks * item.Effects.UnitKillPhysicalDefenseGain
		stats.AbilityPower += int(equipped.Stacks * item.Effects.UnitKillAbilityPowerGain)
		stats.MoveSpeed += item.Stats.MoveSpeed
		stats.MoveSpeed *= 1 + item.Stats.MoveSpeedPercent
		stats.AttackSpeedBonus += item.Stats.AttackSpeedBonus
		stats.CritChance += item.Stats.CritChance
		stats.Omnivamp += item.Stats.Omnivamp
		stats.LifeSteal += item.Stats.LifeSteal
	}
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
