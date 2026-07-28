package shadowassassin

import (
	"l-battle/internal/world"
)

const (
	heroID    = "shadow_assassin"
	passiveID = "shadow_assassin_passive"
	qID       = "shadow_assassin_q"
	wID       = "shadow_assassin_w"
	eID       = "shadow_assassin_e"
	rID       = "shadow_assassin_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		TickEntity:                     Tick,
		ActiveBuffs:                    ActiveBuffs,
		OnBasicHit:                     OnBasicHit,
		OnBasicAttackRelease:           BreakROnAttack,
		BasicAttackBonusPhysicalDamage: QBonusPhysicalDamage,
		PhysicalDamageMultiplier:       PhysicalDamageMultiplier,
		DamageMultiplier:               EDamageMultiplier,
		ResolveProjectile:              ResolveProjectile,
		MoveSpeedMultiplier:            RMoveSpeedMultiplier,
		SpecialRecast:                  SpecialRecast,
	})
}

func PhysicalDamageMultiplier(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64) float64 {
	if w == nil || attacker == nil || attacker.HeroID != heroID || !passiveTarget(target, tick) {
		return 1
	}
	skill := w.SkillConfig(passiveID)
	bonus := 0.1
	if value, ok := skill.Meta["physicalDamageBonus"]; ok {
		bonus = value
	}
	return 1 + bonus
}

func passiveTarget(target *world.Entity, tick uint64) bool {
	if target == nil {
		return false
	}
	control := target.Control
	return (control.MoveSpeedSlow > 0 && tick < control.MoveSpeedSlowUntil) ||
		tick < control.StunnedUntilTick ||
		tick < control.RootedUntilTick ||
		tick < control.SuppressedUntilTick
}
