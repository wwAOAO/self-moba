package blade

import "l-battle/internal/world"

const heroID = "blade"

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Tick:           TickRageDecay,
		OnBasicHit:     OnBasicHit,
		OnSkillHit:     OnSkillHit,
		OnKill:         OnKill,
		RageCritChance: RageCritChance,
	})
}

func OnBasicHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	GainBasicAttackRage(w, source, target, tick)
}

func OnSkillHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	GainSkillHitRage(w, source, tick)
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	GainKillRage(w, killer)
}
