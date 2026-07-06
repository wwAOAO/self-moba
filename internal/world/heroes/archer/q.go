package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

const (
	heroID = "archer"
	qID    = "shot"
	wID    = "roll"
	eID    = "trap"
	rID    = "arrow_rain"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: ApplyQ,
			wID: ApplyW,
			eID: ApplyE,
			rID: ApplyR,
		},
		Tick:             Tick,
		OnBasicHit:       OnBasicHit,
		OnSkillHit:       ApplyFrostShot,
		OnSkillUpgrade:   RefreshSkillOnUpgrade,
		FocusBonusDamage: FocusBonusDamage,
		ApplyFrostShot:   ApplyFrostShot,
		WDamage:          WDamage,
		ArcherRDamage:    RDamage,
		RStunTicks:       RStunTicks,
		ApplyRSplash:     ApplyRSplash,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	ExpireFocus(w, entity, tick)
	TickHawkCharges(w, entity, tick, tickRate)
	ReleaseR(w, entity, tick, tickRate)
}

func OnBasicHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	ApplyFrostShot(w, source, target, tick, tickRate)
	ApplyFocusOnBasicHit(w, source, target, tick, tickRate)
}

func ApplyQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	maxStacks := archerFocusMaxStacks(skill)
	if entity.Archer.FocusStacks < maxStacks {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 50)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Archer.FocusStacks = 0
	entity.Archer.FocusExpireTick = 0
	entity.Archer.FocusActiveLevel = state.Level
	entity.Archer.FocusAttackSpeed = skillList(skill, "attackSpeedBonus", state.Level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
	entity.Archer.FocusBonusADRatio = skillList(skill, "bonusAdDamageRatio", state.Level, []float64{1.05, 1.1, 1.15, 1.2, 1.25})
	entity.Archer.FocusActiveUntil = tick + secondsToTicks(skillMeta(skill, "activeDurationSeconds", 5), tickRate)
	entity.Skills[qID] = state
}
