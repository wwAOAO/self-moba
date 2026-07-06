package sword

import "l-battle/internal/world"

const (
	heroID = "sword"
	qID    = "sword_cut"
	wID    = "sword_wind_wall"
	eID    = "sword_sweeping_blade"
	rID    = "sword_storm"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: ApplyQ,
			wID: ApplyW,
			eID: ApplyE,
			rID: ApplyR,
		},
		ReleaseQ:                 ReleaseQ,
		ExpireQStacks:            ExpireQStacks,
		CritChanceMultiplier:     CritChanceMultiplier,
		ApplyCritFinalMultiplier: ApplyCritFinalMultiplier,
		ApplyShield:              ApplyShield,
		ShieldValue:              ShieldValue,
		PassiveState:             PassiveState,
		StateForHero:             StateForHero,
		ChargeIntent:             ChargeIntent,
		TickShield:               TickShield,
		ApplyCritOverflowStats:   ApplyCritOverflowStats,
	})
}
