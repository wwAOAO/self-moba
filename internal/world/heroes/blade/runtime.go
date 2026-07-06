package blade

import "l-battle/internal/world"

const (
	heroID = "blade"
	qID    = "blade_q"
	wID    = "blade_w"
	eID    = "blade_e"
	rID    = "blade_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: ApplyQ,
			wID: ApplyW,
			eID: ApplyE,
			rID: ApplyR,
		},
		Tick:           Tick,
		OnBasicHit:     OnBasicHit,
		OnSkillHit:     OnSkillHit,
		OnKill:         OnKill,
		RageCritChance: RageCritChance,
		ApplyStats:     ApplyBloodlustStats,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	TickRageDecay(w, entity, tick, tickRate)
	ReleaseW(w, entity, tick, tickRate)
}

func OnBasicHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	GainBasicAttackRage(w, source, target, tick)
	RefundEOnCrit(w, source, target, tick, tickRate)
}

func OnSkillHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	GainSkillHitRage(w, source, tick)
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	GainKillRage(w, killer)
}
