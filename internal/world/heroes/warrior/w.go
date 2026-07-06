package warrior

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func ApplyW(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.CourageUntilTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate)
	entity.Warrior.CourageFrontUntilTick = tick + secondsToTicks(skillMeta(skill, "frontDurationSeconds", 0.75), tickRate)
	entity.Warrior.CourageFrontDamageReduce = skillMeta(skill, "frontDamageReduce", 0.6)
	entity.Warrior.CourageFrontTenacity = skillMeta(skill, "frontTenacity", 0.6)
	entity.Warrior.CourageBackDamageReduce = skillMeta(skill, "backDamageReduce", 0.3)
	entity.Control.TenacityUntilTick = entity.Warrior.CourageFrontUntilTick
	entity.Passive.MaxShield = warriorWShieldValue(entity, skill, state.Level)
	entity.Passive.Shield = entity.Passive.MaxShield
	entity.Passive.ShieldExpireTick = entity.Warrior.CourageUntilTick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(math.Round(skillList(skill, "cooldownMs", state.Level, []float64{24000, 22000, 20000, 18000, 16000}))), tickRate)
	entity.Skills[wID] = state
}

func warriorWShieldValue(entity *world.Entity, skill config.SkillConfig, skillLevel int) int {
	baseShield := skillList(skill, "shieldValue", skillLevel, []float64{70, 95, 120, 145, 170})
	return int(math.Round(baseShield + float64(entity.Stats.BonusHP)*skillMeta(skill, "bonusHealthRatio", 0.2)))
}

func ApplyWPassiveKill(w *world.World, killer *world.Entity, target *world.Entity) {
	if killer == nil || target == nil || killer.HeroID != heroID {
		return
	}
	state, ok := killer.Skills[wID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(wID)
	gain := skillMeta(skill, "passiveMinionResistGain", 0.2)
	if target.Kind == world.EntityKindPlayer || target.Kind == world.EntityKindEnemyHero {
		gain = skillMeta(skill, "passiveHeroResistGain", 1)
	}
	maxGain := skillMeta(skill, "passiveMaxResistGain", 40)
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
	killer.Skills[wID] = state
}
