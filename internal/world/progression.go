package world

const (
	passiveGoldStartSeconds    = 120
	passiveGoldIntervalSeconds = 10
	passiveGoldAmount          = 20.4
)

func (w *World) applyKillReward(killer *Entity, target *Entity) {
	if target == nil {
		return
	}
	w.onHeroKill(killer, target)
	if w.rewards == nil {
		return
	}
	switch target.Kind {
	case EntityKindMeleeMinion, EntityKindRangedMinion, EntityKindSiegeMinion, EntityKindSuperMinion:
		w.addNearbyMinionExperience(target)
		w.applyEquipmentUnitKillGrowth(killer, target)
		if gold, ok := w.rewards.MinionGold(string(target.Kind)); ok {
			w.addGold(killer, float64(gold))
		}
		return
	}
	if killer == nil {
		return
	}
	w.applyEquipmentUnitKillGrowth(killer, target)
	switch target.Kind {
	case EntityKindBlueBuff, EntityKindRedBuff, EntityKindGromp, EntityKindRaptor, EntityKindMurkWolf, EntityKindKrugCamp:
		if exp, ok := w.rewards.JungleExp(string(target.Kind)); ok {
			w.addExperience(killer, exp)
		}
		if gold, ok := w.rewards.JungleGold(string(target.Kind)); ok {
			w.addGold(killer, float64(gold))
		}
	case EntityKindBaronNashor:
		if reward, ok := w.rewards.EpicReward(string(target.Kind)); ok && reward.TeamGold > 0 {
			w.addTeamGold(killer.Team, float64(reward.TeamGold))
		}
	case EntityKindTower:
		if exp, ok := w.rewards.StructureTeamExp(string(target.Kind)); ok {
			w.addTeamExperience(killer.Team, float64(exp))
		}
		if gold, ok := w.rewards.StructureTeamGold(string(target.Kind)); ok {
			w.addTeamGold(killer.Team, float64(gold))
		}
	case EntityKindEnemyHero, EntityKindPlayer:
		targetNextExp := w.nextLevelExp(target.Level)
		if targetNextExp > 0 {
			w.addExperience(killer, w.rewards.HeroKillExp(int(targetNextExp), killer.Level, target.Level))
		}
		w.addGold(killer, float64(w.rewards.HeroKillGold()))
	}
}

func (w *World) addNearbyMinionExperience(target *Entity) {
	if target == nil || w.rewards == nil {
		return
	}
	receivers := w.nearbyMinionExperienceReceivers(target)
	if len(receivers) == 0 {
		return
	}
	exp, ok := w.rewards.MinionExp(string(target.Kind), len(receivers))
	if !ok {
		return
	}
	for _, receiver := range receivers {
		w.addExperience(receiver, exp)
	}
}

func (w *World) nearbyMinionExperienceReceivers(target *Entity) []*Entity {
	if target == nil || w.rewards == nil {
		return nil
	}
	team := oppositeTeam(target.Team)
	radius := w.rewards.Minion.ShareRadius
	receivers := make([]*Entity, 0, 2)
	for _, entity := range w.entities {
		if entity.Kind != EntityKindPlayer || entity.Team != team || entity.Stats.HP <= 0 || entity.Death.Dead {
			continue
		}
		if distance(target.Position, entity.Position) <= radius+entity.Radius {
			receivers = append(receivers, entity)
		}
	}
	return receivers
}

func (w *World) addTeamExperience(team Team, exp float64) {
	for _, entity := range w.entities {
		if entity.Kind == EntityKindPlayer && entity.Team == team && entity.Stats.HP > 0 {
			w.addExperience(entity, exp)
		}
	}
}

func (w *World) addTeamGold(team Team, gold float64) {
	for _, entity := range w.entities {
		if entity.Kind == EntityKindPlayer && entity.Team == team {
			w.addGold(entity, gold)
		}
	}
}

func (w *World) addGold(entity *Entity, gold float64) {
	if entity == nil || entity.Kind != EntityKindPlayer || gold <= 0 {
		return
	}
	entity.Gold += gold
}

func (w *World) tickPassiveGold(tick uint64, tickRate int) {
	if tickRate <= 0 {
		return
	}
	interval := secondsToTicks(passiveGoldIntervalSeconds, tickRate)
	if interval == 0 {
		return
	}
	if w.nextPassiveGoldTick == 0 {
		w.nextPassiveGoldTick = secondsToTicks(passiveGoldStartSeconds, tickRate)
	}
	for tick >= w.nextPassiveGoldTick {
		for _, entity := range w.entities {
			w.addGold(entity, passiveGoldAmount)
		}
		w.nextPassiveGoldTick += interval
	}
}

func (w *World) addExperience(entity *Entity, exp float64) {
	if entity == nil || entity.Kind != EntityKindPlayer || exp <= 0 || entity.Level >= MaxHeroLevel {
		return
	}
	entity.Exp += exp
	entity.TotalExp += exp
	for entity.Level < MaxHeroLevel {
		nextExp := w.nextLevelExp(entity.Level)
		if nextExp <= 0 || entity.Exp < nextExp {
			break
		}
		entity.Exp -= nextExp
		w.levelUp(entity)
	}
	entity.NextLevelExp = w.nextLevelExp(entity.Level)
}

func (w *World) levelUp(entity *Entity) {
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return
	}
	oldMaxHP := entity.Stats.MaxHP
	oldMaxMP := entity.Stats.MaxMP
	entity.Level = clampInt(entity.Level+1, MinHeroLevel, MaxHeroLevel)
	nextStats := heroStatsAtLevel(hero, entity.Level)
	w.applyEquipmentStats(entity, &nextStats)
	w.applySwordCritOverflowStats(entity, &nextStats)
	nextStats.AbilityHaste += abilityHasteFromBuffs(entity)
	hpGain := nextStats.MaxHP - oldMaxHP
	mpGain := nextStats.MaxMP - oldMaxMP
	nextStats.HP = entity.Stats.HP + hpGain
	nextStats.MP = entity.Stats.MP + mpGain
	if nextStats.HP > nextStats.MaxHP {
		nextStats.HP = nextStats.MaxHP
	}
	if nextStats.MP > nextStats.MaxMP {
		nextStats.MP = nextStats.MaxMP
	}
	w.applyHeroStats(entity, &nextStats)
	applyControlStats(entity, &nextStats)
	entity.Tank.ThunderclapArmorBonus = 0
	entity.Stats = nextStats
	w.refreshTankGraniteShieldMax(entity)
	w.refreshTankWPassive(entity)
	w.applyEquipmentLevelUpRestore(entity)
	w.refreshPlayerStatsAfterHPChange(entity, nextStats.HP)
	entity.SkillPoints++
}

func (w *World) refreshTankGraniteShield(entity *Entity) {
	if heroHooksFor(tankHeroID).RefreshGranite != nil {
		heroHooksFor(tankHeroID).RefreshGranite(w, entity)
	}
}

func (w *World) refreshTankGraniteShieldMax(entity *Entity) {
	if heroHooksFor(tankHeroID).RefreshGraniteMax != nil {
		heroHooksFor(tankHeroID).RefreshGraniteMax(w, entity)
	}
}

func (w *World) debugLevelUp(entity *Entity) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Level >= MaxHeroLevel {
		return
	}
	w.levelUp(entity)
	entity.NextLevelExp = w.nextLevelExp(entity.Level)
	if entity.Level >= MaxHeroLevel {
		entity.Exp = 0
	}
}

func (w *World) nextLevelExp(level int) float64 {
	if w.levels == nil {
		return 0
	}
	nextExp, ok := w.levels.NextExp(level)
	if !ok {
		return 0
	}
	return float64(nextExp)
}

func (w *World) upgradeSkill(entity *Entity, slot string) {
	if entity == nil || entity.SkillPoints <= 0 {
		return
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return
	}
	skillID := hero.Skills.SkillIDForSlot(slot)
	if skillID == "" {
		return
	}
	state, ok := entity.Skills[skillID]
	if !ok {
		return
	}
	maxLevel := maxSkillLevel(slot)
	if maxLevel <= 0 || state.Level >= maxLevel {
		return
	}
	if !canUpgradeSkillAtLevel(slot, state.Level, entity.Level) {
		return
	}
	state.Level++
	entity.SkillPoints--
	entity.Skills[skillID] = state
	w.onHeroSkillUpgrade(entity, skillID)
	if heroHooksForEntity(entity).ApplyStats != nil {
		w.recalculatePlayerStats(entity)
	}
	w.refreshTankWPassive(entity)
}

func maxSkillLevel(slot string) int {
	switch slot {
	case "q", "w", "e":
		return MaxBasicSkillLevel
	case "r":
		return MaxUltSkillLevel
	default:
		return 0
	}
}

func canUpgradeSkillAtLevel(slot string, skillLevel int, heroLevel int) bool {
	if slot != "r" {
		return true
	}
	required := []int{6, 11, 16}
	return skillLevel >= 0 && skillLevel < len(required) && heroLevel >= required[skillLevel]
}
