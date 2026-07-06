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
		Tick:                     Tick,
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

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	ExpireQStacks(entity, tick)
	if entity != nil && entity.Sword.QPending {
		w.TickDashMovement(entity, tick, tickRate)
	}
	ReleaseQ(w, entity, tick, tickRate)
}
