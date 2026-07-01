package world

import (
	"l-battle/internal/config"
	"math"
)

func normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func heroStatsAtLevel(hero config.HeroConfig, level int) Stats {
	level = clampInt(level, MinHeroLevel, MaxHeroLevel)
	growthSteps := level - MinHeroLevel
	growthStepValue := float64(growthSteps)
	bonusHP := hero.Base.BonusHP + hero.Growth.BonusHP*growthSteps
	hp := hero.Base.HP + hero.Growth.HP*growthSteps + bonusHP
	mp := hero.Base.MP + hero.Growth.MP*growthStepValue
	baseAttackSpeed := hero.Base.AttackSpeed * (1 + hero.Growth.AttackSpeed*growthStepValue)
	attackSpeedRatio := hero.Base.AttackSpeedRatio
	attackSpeedBonus := hero.Base.BonusAttackSpeed
	attackSpeedSlow := hero.Base.AttackSpeedSlow
	attackSpeed := finalAttackSpeed(baseAttackSpeed, attackSpeedBonus, attackSpeedRatio, attackSpeedSlow)
	attackWindupSeconds := hero.Base.AttackWindupSeconds + hero.Growth.AttackWindupSeconds*growthStepValue
	if attackWindupSeconds <= 0 {
		attackWindupSeconds = 0.25
	}
	return Stats{
		HP:                   hp,
		MaxHP:                hp,
		BonusHP:              bonusHP,
		MP:                   mp,
		MaxMP:                mp,
		HPRegen5:             hero.Base.HPRegen5 + hero.Growth.HPRegen5*growthStepValue,
		MPRegen5:             hero.Base.MPRegen5 + hero.Growth.MPRegen5*growthStepValue,
		Attack:               hero.Base.Attack + hero.Growth.Attack*growthStepValue,
		BonusAttack:          hero.Base.BonusAttack + hero.Growth.BonusAttack*growthStepValue,
		AbilityPower:         hero.Base.AbilityPower + hero.Growth.AbilityPower*growthSteps,
		AbilityHaste:         hero.Base.AbilityHaste + hero.Growth.AbilityHaste*growthStepValue,
		DamageReduce:         hero.Base.DamageReduce + hero.Growth.DamageReduce*growthStepValue,
		PhysicalDefense:      hero.Base.PhysicalDefense + hero.Growth.PhysicalDefense*growthStepValue,
		BonusPhysicalDefense: hero.Base.BonusPhysicalDefense + hero.Growth.BonusPhysicalDefense*growthStepValue,
		PhysicalPenPercent:   hero.Base.PhysicalPenPercent + hero.Growth.PhysicalPenPercent*growthStepValue,
		PhysicalPenFlat:      hero.Base.PhysicalPenFlat + hero.Growth.PhysicalPenFlat*growthStepValue,
		PhysicalDamageReduce: hero.Base.PhysicalDamageReduce + hero.Growth.PhysicalDamageReduce*growthStepValue,
		MagicDefense:         hero.Base.MagicDefense + hero.Growth.MagicDefense*growthStepValue,
		BonusMagicDefense:    hero.Base.BonusMagicDefense + hero.Growth.BonusMagicDefense*growthStepValue,
		MagicPenPercent:      hero.Base.MagicPenPercent + hero.Growth.MagicPenPercent*growthStepValue,
		MagicPenFlat:         hero.Base.MagicPenFlat + hero.Growth.MagicPenFlat*growthStepValue,
		MagicDamageReduce:    hero.Base.MagicDamageReduce + hero.Growth.MagicDamageReduce*growthStepValue,
		MoveSpeed:            hero.Base.MoveSpeed + hero.Growth.MoveSpeed*growthStepValue,
		AttackRange:          hero.Base.AttackRange + hero.Growth.AttackRange*growthStepValue,
		AttackSpeed:          attackSpeed,
		AttackWindupSeconds:  attackWindupSeconds,
		BaseAttackSpeed:      baseAttackSpeed,
		AttackSpeedBonus:     attackSpeedBonus,
		AttackSpeedRatio:     attackSpeedRatio,
		AttackSpeedSlow:      attackSpeedSlow,
		CritChance:           hero.Base.CritChance + hero.Growth.CritChance*growthStepValue,
	}
}

func finalAttackSpeed(baseAttackSpeed float64, attackSpeedBonus float64, attackSpeedRatio float64, attackSpeedSlow float64) float64 {
	if baseAttackSpeed < 0 {
		baseAttackSpeed = 0
	}
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	if attackSpeedRatio < 0 {
		attackSpeedRatio = 0
	}
	if attackSpeedSlow < 0 {
		attackSpeedSlow = 0
	}
	if attackSpeedSlow > 1 {
		attackSpeedSlow = 1
	}
	attackSpeed := baseAttackSpeed * (1 + attackSpeedBonus*attackSpeedRatio) * (1 - attackSpeedSlow)
	return clamp(attackSpeed, 0, 2.5)
}

func (w *World) passiveStateForHero(hero config.HeroConfig) PassiveState {
	if hero.HeroID == swordHeroID {
		skill := w.skillConfig(hero.Skills.Passive)
		return PassiveState{
			MaxSwordIntent: skillMetaRange(skill, "intentMax", 100),
		}
	}
	if hero.HeroID == tankHeroID {
		stats := heroStatsAtLevel(hero, MinHeroLevel)
		shield := tankGraniteShieldValue(stats.MaxHP, w.skillConfig(hero.Skills.Passive))
		return PassiveState{
			Shield:    shield,
			MaxShield: shield,
		}
	}
	return PassiveState{}
}

func swordStateForHero(heroID string) SwordState {
	if heroID != swordHeroID {
		return SwordState{}
	}
	return SwordState{
		SweepingBladeTargetUntil: make(map[string]uint64),
	}
}

func (w *World) chargeSwordIntent(entity *Entity, moved float64) {
	if entity == nil || entity.HeroID != swordHeroID || moved <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	if entity.Passive.MaxSwordIntent <= 0 {
		entity.Passive.MaxSwordIntent = skillMetaRange(skill, "intentMax", 100)
	}
	if entity.Passive.SwordIntent >= entity.Passive.MaxSwordIntent {
		return
	}
	moveUnitsPerPercent := skillMetaCurveByLevel(skill, "intentMoveUnitsPerPercent", "intentMoveUnitLevels", entity.Level, 59)
	if moveUnitsPerPercent <= 0 {
		moveUnitsPerPercent = 59
	}
	entity.Passive.SwordIntent += moved / moveUnitsPerPercent
	if entity.Passive.SwordIntent > entity.Passive.MaxSwordIntent {
		entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	}
}

func (w *World) tickSwordShield(entity *Entity, tick uint64) {
	if entity == nil || entity.Passive.Shield <= 0 || entity.Passive.ShieldExpireTick == 0 {
		return
	}
	if tick < entity.Passive.ShieldExpireTick {
		return
	}
	entity.Passive.Shield = 0
	entity.Passive.MaxShield = 0
	entity.Passive.ShieldExpireTick = 0
}

func (w *World) tickTankGraniteShield(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || entity.Stats.HP <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	maxShield := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	if maxShield <= 0 || entity.Passive.Shield >= maxShield {
		entity.Passive.MaxShield = maxShield
		return
	}
	resetTicks := secondsToTicks(skillMetaRange(skill, "resetSeconds", 10), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+resetTicks {
		return
	}
	entity.Passive.MaxShield = maxShield
	entity.Passive.Shield = maxShield
}

func tankGraniteShieldValue(maxHP int, skill config.SkillConfig) int {
	return int(math.Round(float64(maxHP) * skillMetaRange(skill, "shieldMaxHPRatio", 0.1)))
}

func (w *World) tickWarriorToughness(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || entity.Stats.HP <= 0 || entity.Stats.HP >= entity.Stats.MaxHP {
		return
	}
	skill := w.heroPassiveSkill(entity)
	outOfCombatTicks := secondsToTicks(skillMetaRange(skill, "outOfCombatSeconds", 8), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+outOfCombatTicks {
		return
	}
	intervalTicks := secondsToTicks(skillMetaRange(skill, "regenIntervalSeconds", 5), tickRate)
	if intervalTicks == 0 {
		intervalTicks = uint64(tickRate * 5)
	}
	if entity.Passive.NextRegenTick == 0 {
		entity.Passive.NextRegenTick = tick
	}
	if tick < entity.Passive.NextRegenTick {
		return
	}
	ratio := warriorToughnessRegenRatio(entity.Level, skill)
	heal := int(math.Round(float64(entity.Stats.MaxHP) * ratio))
	if heal < 1 {
		heal = 1
	}
	entity.Stats.HP += heal
	if entity.Stats.HP > entity.Stats.MaxHP {
		entity.Stats.HP = entity.Stats.MaxHP
	}
	entity.Passive.NextRegenTick = tick + intervalTicks
}

func tickBaseRegen(entity *Entity, tickRate int) {
	if entity == nil || tickRate <= 0 || entity.Stats.HP <= 0 {
		return
	}
	if entity.Stats.HP < entity.Stats.MaxHP && entity.Stats.HPRegen5 > 0 {
		entity.Regen.HPRemainder += entity.Stats.HPRegen5 / 5 / float64(tickRate)
		heal := int(math.Floor(entity.Regen.HPRemainder + 0.000000001))
		if heal > 0 {
			entity.Stats.HP += heal
			entity.Regen.HPRemainder -= float64(heal)
			if entity.Stats.HP > entity.Stats.MaxHP {
				entity.Stats.HP = entity.Stats.MaxHP
				entity.Regen.HPRemainder = 0
			}
		}
	} else {
		entity.Regen.HPRemainder = 0
	}
	if entity.Stats.MP < entity.Stats.MaxMP && entity.Stats.MPRegen5 > 0 {
		entity.Stats.MP += entity.Stats.MPRegen5 / 5 / float64(tickRate)
		if entity.Stats.MP > entity.Stats.MaxMP {
			entity.Stats.MP = entity.Stats.MaxMP
			entity.Regen.MPRemainder = 0
		}
	} else {
		entity.Regen.MPRemainder = 0
	}
}

func warriorToughnessRegenRatio(level int, skill config.SkillConfig) float64 {
	return skillMetaListByLevel(skill, "regenMaxHPRatio", level, []float64{
		0.015, 0.0198, 0.0246, 0.0294, 0.0342, 0.039,
		0.0438, 0.0486, 0.0534, 0.0582, 0.063, 0.0678,
		0.0726, 0.0774, 0.0822, 0.087, 0.0918, 0.101,
	})
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
