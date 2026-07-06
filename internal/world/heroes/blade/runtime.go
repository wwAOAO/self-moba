package blade

import "l-battle/internal/world"

const heroID = "blade"

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		GainBasicAttackRage: GainBasicAttackRage,
		GainKillRage:        GainKillRage,
		GainSkillHitRage:    GainSkillHitRage,
		TickRageDecay:       TickRageDecay,
		RageCritChance:      RageCritChance,
	})
}
