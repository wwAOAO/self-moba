package equipcalc

import (
	"l-battle/internal/config"
	"l-battle/internal/world/formula"
	"l-battle/internal/world/state"
	"math"
)

func ShieldByLevel(level int, minLevel int, maxLevel int, minShield int, maxShield int) int {
	if maxShield <= minShield {
		return minShield
	}
	level = formula.ClampInt(level, minLevel, maxLevel)
	progress := float64(level-minLevel) / float64(maxLevel-minLevel)
	return int(math.Round(float64(minShield) + float64(maxShield-minShield)*progress))
}

func SunfireDamage(level int, minLevel int, maxLevel int, bonusHP int, item config.EquipmentConfig, equipped state.EquipmentSlot) float64 {
	level = formula.ClampInt(level, minLevel, maxLevel)
	base := item.Effects.SunfireBurnFlatMin
	if item.Effects.SunfireBurnFlatMax > base {
		base += (item.Effects.SunfireBurnFlatMax - base) * float64(level-minLevel) / float64(maxLevel-minLevel)
	}
	return (base + float64(bonusHP)*item.Effects.SunfireBurnBonusHPRatio) * (1 + equipped.Stacks*item.Effects.SunfireStackDamageBonus)
}
