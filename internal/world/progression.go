package world

func (w *World) applyKillReward(killer *Entity, target *Entity) {
	if target == nil {
		return
	}
	w.applyMageFinalSparkRefund(target)
	if killer == nil || w.rewards == nil {
		return
	}
	w.applyWarriorWPassiveKill(killer, target)
	w.applyEquipmentUnitKillGrowth(killer, target)
	switch target.Kind {
	case EntityKindMeleeMinion, EntityKindRangedMinion, EntityKindSiegeMinion, EntityKindSuperMinion:
		if exp, ok := w.rewards.MinionExp(string(target.Kind), 1); ok {
			w.addExperience(killer, exp)
		}
		if gold, ok := w.rewards.MinionGold(string(target.Kind)); ok {
			w.addGold(killer, float64(gold))
		}
	case EntityKindBlueBuff, EntityKindRedBuff, EntityKindGromp, EntityKindRaptor, EntityKindMurkWolf, EntityKindKrugCamp:
		if exp, ok := w.rewards.JungleExp(string(target.Kind)); ok {
			w.addExperience(killer, float64(exp))
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

func (w *World) applyWarriorWPassiveKill(killer *Entity, target *Entity) {
	if killer == nil || target == nil || killer.HeroID != warriorHeroID {
		return
	}
	state, ok := killer.Skills[warriorWSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.skillConfig(warriorWSkillID)
	gain := skillMetaRange(skill, "passiveMinionResistGain", 0.2)
	if target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero {
		gain = skillMetaRange(skill, "passiveHeroResistGain", 1)
	}
	maxGain := skillMetaRange(skill, "passiveMaxResistGain", 40)
	before := killer.Warrior.CouragePassiveResistGain
	after := before + gain
	if after > maxGain {
		after = maxGain
	}
	if after <= before {
		return
	}
	delta := after - before
	killer.Warrior.CouragePassiveResistGain = after
	killer.Stats.PhysicalDefense += delta
	killer.Stats.BonusPhysicalDefense += delta
	killer.Stats.MagicDefense += delta
	killer.Stats.BonusMagicDefense += delta
	killer.Skills[warriorWSkillID] = state
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
	entity.Tank.ThunderclapArmorBonus = 0
	entity.Stats = nextStats
	w.refreshTankGraniteShieldMax(entity)
	w.refreshTankWPassive(entity)
	w.applyEquipmentLevelUpRestore(entity)
	entity.SkillPoints++
}

func (w *World) refreshTankGraniteShield(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	skill := w.heroPassiveSkill(entity)
	shield := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	entity.Passive.MaxShield = shield
	entity.Passive.Shield = shield
	entity.Passive.ShieldExpireTick = 0
}

func (w *World) refreshTankGraniteShieldMax(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	oldMax := entity.Passive.MaxShield
	skill := w.heroPassiveSkill(entity)
	nextMax := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	entity.Passive.MaxShield = nextMax
	if entity.Passive.Shield >= oldMax {
		entity.Passive.Shield = nextMax
	}
	if entity.Passive.Shield > nextMax {
		entity.Passive.Shield = nextMax
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
	state.Level++
	entity.SkillPoints--
	entity.Skills[skillID] = state
	w.refreshArcherSkillOnUpgrade(entity, skillID)
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
