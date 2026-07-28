package shadowassassin

import (
	"l-battle/internal/world"
	"testing"
)

func TestMercifulHeartAmplifiesPhysicalDamageAgainstControlledTargets(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*world.Entity)
	}{
		{name: "slow", apply: func(target *world.Entity) {
			target.Control.MoveSpeedSlow = 0.3
			target.Control.MoveSpeedSlowUntil = 20
		}},
		{name: "stun", apply: func(target *world.Entity) { target.Control.StunnedUntilTick = 20 }},
		{name: "root", apply: func(target *world.Entity) { target.Control.RootedUntilTick = 20 }},
		{name: "suppression", apply: func(target *world.Entity) { target.Control.SuppressedUntilTick = 20 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := world.NewWorld(nil, nil, nil, nil, nil)
			attacker := &world.Entity{HeroID: heroID, Stats: world.Stats{HP: 1000}}
			target := &world.Entity{Stats: world.Stats{HP: 1000}}
			target.Combat.LastHitTick = 10
			test.apply(target)

			w.ApplyDamage(attacker, target, 100, 20)

			if got := target.Combat.LastDamage; got != 110 {
				t.Fatalf("physical damage = %d, want 110", got)
			}
		})
	}
}

func TestMercifulHeartIgnoresOtherControlAndNonPhysicalDamage(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	attacker := &world.Entity{HeroID: heroID, Stats: world.Stats{HP: 1000}}
	target := &world.Entity{Stats: world.Stats{HP: 1000}}
	target.Combat.LastHitTick = 10
	target.Control.AirborneUntilTick = 20
	target.Control.TauntedUntilTick = 20

	w.ApplyDamage(attacker, target, 100, 20)
	if got := target.Combat.LastDamage; got != 100 {
		t.Fatalf("physical damage against other control = %d, want 100", got)
	}

	target.Stats.HP = 1000
	target.Control.StunnedUntilTick = 20
	w.ApplyMagicDamage(attacker, target, 100, 20)
	if got := target.Combat.LastDamage; got != 100 {
		t.Fatalf("magic damage = %d, want 100", got)
	}
}
