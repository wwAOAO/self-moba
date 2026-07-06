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
		ReleaseQ:              ReleaseQ,
		ReleaseW:              ReleaseW,
		ReleaseE:              ReleaseE,
		ActivateEZone:         ActivateEZone,
		TickE:                 TickE,
		DetonateE:             DetonateE,
		ReleaseR:              ReleaseR,
		ApplyFinalSparkRefund: ApplyFinalSparkRefund,
		MageQDamage:           QDamage,
		ApplyIllumination:     ApplyIllumination,
		ApplyUltimateIllum:    ApplyUltimateIllumination,
		TriggerIllumination:   TriggerIllumination,
		DetonateIllumination:  DetonateIllumination,
	})
}
