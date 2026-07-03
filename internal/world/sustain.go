package world

import "math"

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
