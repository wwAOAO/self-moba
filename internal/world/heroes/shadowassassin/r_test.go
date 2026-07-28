package shadowassassin

import (
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestShadowAssaultCastEntersStealthAndCreatesBladeRing(t *testing.T) {
	w, source, _ := qTestWorld(t)
	state := source.Skills[rID]
	state.Level = 2
	source.Skills[rID] = state
	source.Stats.MP = 100

	CastR(w, source, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)

	if source.Stats.MP != 10 {
		t.Fatalf("mana = %v, want 10", source.Stats.MP)
	}
	if got := source.Skills[rID].CooldownUntilTick; got != 1310 {
		t.Fatalf("cooldown = %d, want 1310", got)
	}
	if !source.ShadowAssassin.RActive || source.ShadowAssassin.RUntil != 60 || source.Control.InvisibleUntilTick != 60 {
		t.Fatalf("stealth state = %+v invisible=%d", source.ShadowAssassin, source.Control.InvisibleUntilTick)
	}
	if got := RMoveSpeedMultiplier(source, 20); got != 1.4 {
		t.Fatalf("move speed multiplier = %v, want 1.4", got)
	}
	count := 0
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "shadow_assassin_r" {
			count++
		}
	}
	if count != 24 {
		t.Fatalf("blade count = %d, want 24", count)
	}
}

func TestShadowAssaultAutomaticallyRecallsAndHitsTwice(t *testing.T) {
	w, source, target := qTestWorld(t)
	state := source.Skills[rID]
	state.Level = 1
	source.Skills[rID] = state
	source.Stats.BonusAttack = 100
	target.Position = world.Vector2{X: source.Position.X + 300, Y: source.Position.Y}
	target.Stats.PhysicalDefense = 0
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000

	CastR(w, source, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)
	for tick := uint64(11); tick <= 75; tick++ {
		w.Tick(tick, 20)
	}

	if source.ShadowAssassin.RActive || source.Control.InvisibleUntilTick != 0 {
		t.Fatal("stealth did not end at 2.5 seconds")
	}
	if got := 1000 - target.Stats.HP; got != 420 {
		t.Fatalf("two-pass damage = %v, want 420", got)
	}
}

func TestShadowAssaultBreaksOnAttackSkillAndRecast(t *testing.T) {
	tests := []struct {
		name         string
		breakStealth func(*world.World, *world.Entity, *world.Entity)
	}{
		{name: "attack", breakStealth: func(w *world.World, source *world.Entity, target *world.Entity) {
			BreakROnAttack(w, source, target, 20, 20)
		}},
		{name: "skill", breakStealth: func(w *world.World, source *world.Entity, _ *world.Entity) {
			state := source.Skills[qID]
			state.Level = 1
			source.Skills[qID] = state
			CastQ(w, source, protocol.CastInput{SkillID: qID}, state, w.SkillConfig(qID), 20, 20)
		}},
		{name: "recast", breakStealth: func(w *world.World, source *world.Entity, _ *world.Entity) {
			SpecialRecast(w, source, protocol.CastInput{SkillID: rID}, source.Skills[rID], w.SkillConfig(rID), 20, 20)
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w, source, target := qTestWorld(t)
			state := source.Skills[rID]
			state.Level = 1
			source.Skills[rID] = state
			CastR(w, source, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)

			test.breakStealth(w, source, target)

			if source.ShadowAssassin.RActive || source.Control.InvisibleUntilTick != 0 {
				t.Fatal("action did not break stealth")
			}
		})
	}
}
