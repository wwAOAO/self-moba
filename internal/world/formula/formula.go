package formula

import (
	"math"
	"strconv"
)

const CritRollModulo = 10000

func Normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func Clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func ClampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func FinalAttackSpeed(baseAttackSpeed float64, attackSpeedBonus float64, attackSpeedRatio float64, attackSpeedSlow float64) float64 {
	if baseAttackSpeed < 0 {
		baseAttackSpeed = 0
	}
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	if attackSpeedRatio < 0 {
		attackSpeedRatio = 0
	}
	attackSpeedSlow = Clamp(attackSpeedSlow, 0, 1)
	attackSpeed := baseAttackSpeed * (1 + attackSpeedBonus*attackSpeedRatio) * (1 - attackSpeedSlow)
	return Clamp(attackSpeed, 0, 2.5)
}

func AttackCooldownTicks(attackSpeed float64, tickRate int) uint64 {
	if attackSpeed <= 0 {
		return uint64(tickRate)
	}
	ticks := math.Ceil(float64(tickRate) / attackSpeed)
	if ticks < 1 {
		return 1
	}
	return uint64(ticks)
}

func EffectiveResistance(resistance float64, percentPen float64, flatPen float64) float64 {
	if resistance < 0 {
		return resistance
	}
	percentPen = Clamp(percentPen, 0, 1)
	if flatPen < 0 {
		flatPen = 0
	}
	effective := resistance*(1-percentPen) - flatPen
	if effective < 0 {
		return 0
	}
	return effective
}

func DamageAfterResistance(rawDamage float64, resistance float64, damageReduce float64) int {
	if rawDamage <= 0 {
		return 0
	}
	multiplier := 100 / (resistance + 100)
	if resistance < 0 {
		denominator := 100 + resistance
		if denominator < 1 {
			denominator = 1
		}
		multiplier = 100 / denominator
	}
	damageReduce = Clamp(damageReduce, 0, 1)
	damage := int(math.Round(rawDamage * multiplier * (1 - damageReduce)))
	if damage < 1 {
		return 1
	}
	return damage
}

func StackDamageReduction(reductions ...float64) float64 {
	multiplier := 1.0
	for _, reduction := range reductions {
		reduction = Clamp(reduction, 0, 1)
		multiplier *= 1 - reduction
	}
	return 1 - multiplier
}

func StackTenacity(tenacityValues ...float64) float64 {
	multiplier := 1.0
	for _, tenacity := range tenacityValues {
		tenacity = Clamp(tenacity, 0, 1)
		multiplier *= 1 - tenacity
	}
	return 1 - multiplier
}

func DeterministicCritRoll(attackerID string, targetID string, tick uint64) float64 {
	hash := uint64(1469598103934665603)
	for _, value := range []string{attackerID, targetID, strconv.FormatUint(tick, 10)} {
		for i := 0; i < len(value); i++ {
			hash ^= uint64(value[i])
			hash *= 1099511628211
		}
	}
	return float64(hash%CritRollModulo) / CritRollModulo
}
