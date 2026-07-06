package mage

import "l-battle/internal/world"

const (
	heroID = "mage"
	qID    = "mage_q"
	wID    = "mage_w"
	eID    = "mage_e"
	rID    = "mage_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: ApplyQ,
			wID: ApplyW,
			eID: ApplyE,
			rID: ApplyR,
		},
		Tick:               Tick,
		OnBasicHit:         TriggerIllumination,
		OnSkillHit:         ApplyIllumination,
		OnKill:             OnKill,
		ActivateEZone:      ActivateEZone,
		DetonateE:          DetonateE,
		MageQDamage:        QDamage,
		ApplyUltimateIllum: ApplyUltimateIllumination,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	ReleaseQ(w, entity, tick, tickRate)
	ReleaseW(w, entity, tick, tickRate)
	ReleaseE(w, entity, tick, tickRate)
	ReleaseR(w, entity, tick, tickRate)
	TickE(w, entity, tick, tickRate)
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	ApplyFinalSparkRefund(w, target)
}
