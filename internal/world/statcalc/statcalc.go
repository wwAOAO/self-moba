package statcalc

import (
	"l-battle/internal/config"
	"l-battle/internal/world/formula"
	"l-battle/internal/world/state"
)

func HeroStatsAtLevel(hero config.HeroConfig, level int, minLevel int, maxLevel int) state.Stats {
	level = formula.ClampInt(level, minLevel, maxLevel)
	growthSteps := level - minLevel
	growthStepValue := float64(growthSteps)
	bonusHP := hero.Base.BonusHP + hero.Growth.BonusHP*growthStepValue
	hp := hero.Base.HP + hero.Growth.HP*growthStepValue + bonusHP
	mp := hero.Base.MP + hero.Growth.MP*growthStepValue
	baseAttackSpeed := hero.Base.AttackSpeed * (1 + hero.Growth.AttackSpeed*growthStepValue)
	attackSpeedRatio := hero.Base.AttackSpeedRatio
	attackSpeedBonus := hero.Base.BonusAttackSpeed
	attackSpeedSlow := hero.Base.AttackSpeedSlow
	attackSpeed := formula.FinalAttackSpeed(baseAttackSpeed, attackSpeedBonus, attackSpeedRatio, attackSpeedSlow)
	attackWindupSeconds := hero.Base.AttackWindupSeconds + hero.Growth.AttackWindupSeconds*growthStepValue
	if attackWindupSeconds <= 0 {
		attackWindupSeconds = 0.25
	}
	return state.Stats{
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
