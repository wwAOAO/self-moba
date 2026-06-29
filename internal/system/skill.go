package system

import "l-battle/internal/world"

type SkillCast struct {
	Caster *world.Entity
}

func ApplySkill(_ SkillCast) {
}
