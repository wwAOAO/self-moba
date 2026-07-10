package world

import "math"

type sustainContext struct {
	BasicAttack            bool
	AOE                    bool
	Pet                    bool
	Nonlethal              bool
	SkipBerserkerBleed     bool
	SkipEquipmentSkillSlow bool
	SkipEquipmentEffects   bool
}

var (
	sustainSingleTargetSkill = sustainContext{}
	sustainBasicAttack       = sustainContext{BasicAttack: true}
	sustainAOESkill          = sustainContext{AOE: true}
	sustainPetDamage         = sustainContext{Pet: true}
	sustainEquipmentDamage   = sustainContext{Pet: true, SkipBerserkerBleed: true, SkipEquipmentSkillSlow: true, SkipEquipmentEffects: true}
)

func (w *World) applySustain(source *Entity, actualDamage int, context sustainContext) {
	if source == nil || source.Kind != EntityKindPlayer || actualDamage <= 0 || source.Stats.HP <= 0 {
		return
	}
	ratio := source.Stats.Omnivamp
	lifeStealRatio := 0.0
	if context.BasicAttack {
		lifeStealRatio = source.Stats.LifeSteal
		ratio += lifeStealRatio
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
	missingHP := int(math.Ceil(source.Stats.MaxHP - source.Stats.HP))
	if missingHP < 0 {
		missingHP = 0
	}
	overheal := heal - missingHP
	source.Stats.HP += float64(heal)
	if source.Stats.HP > source.Stats.MaxHP {
		source.Stats.HP = source.Stats.MaxHP
	}
	if overheal > 0 && lifeStealRatio > 0 {
		lifeStealHeal := int(math.Floor(float64(actualDamage)*lifeStealRatio*decay*(1+source.Stats.HealingPower)*(1-clamp(source.Stats.GrievousWounds, 0, 1)) + 0.000000001))
		if lifeStealHeal > overheal {
			lifeStealHeal = overheal
		}
		w.applyEquipmentLifeStealOverhealShield(source, lifeStealHeal)
	}
}

func isHeroDamageSource(entity *Entity) bool {
	return entity != nil && (entity.Kind == EntityKindPlayer || entity.Kind == EntityKindEnemyHero)
}
