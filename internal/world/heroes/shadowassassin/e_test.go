package shadowassassin

import (
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"testing"
)

func TestCutthroatWarBlinksBehindRootsAndMarksHero(t *testing.T) {
	w, source, target := qTestWorld(t)
	state := source.Skills[eID]
	state.Level = 3
	source.Skills[eID] = state
	source.Stats.MP = 100

	CastE(w, source, protocol.CastInput{SkillID: eID, TargetID: target.ID}, state, w.SkillConfig(eID), 10, 20)

	if source.Stats.MP != 55 {
		t.Fatalf("mana = %v, want 55", source.Stats.MP)
	}
	if got := source.Skills[eID].CooldownUntilTick; got != 290 {
		t.Fatalf("cooldown = %d, want 290", got)
	}
	if math.Abs(source.Position.X-5240) > 0.001 || math.Abs(source.Position.Y-5000) > 0.001 {
		t.Fatalf("position = %+v, want behind target at (5240, 5000)", source.Position)
	}
	if target.Control.RootedUntilTick != 30 {
		t.Fatalf("root until = %d, want 30", target.Control.RootedUntilTick)
	}
	if source.ShadowAssassin.ETargetID != target.ID || source.ShadowAssassin.EDamageAmp != 0.09 || source.ShadowAssassin.EDamageUntil != 70 {
		t.Fatalf("damage mark = %+v", source.ShadowAssassin)
	}
	if !hasEffect(w.SkillEffects(), "shadow_assassin_e") {
		t.Fatal("blink effect is missing")
	}
}

func TestCutthroatWarRejectsNonHeroAndOutOfRangeTargets(t *testing.T) {
	w, source, target := qTestWorld(t)
	state := source.Skills[eID]
	state.Level = 1
	source.Skills[eID] = state
	source.Stats.MP = 100
	start := source.Position
	minion := &world.Entity{ID: "minion", Kind: world.EntityKindMeleeMinion, Team: world.TeamRed, Position: target.Position, Stats: world.Stats{HP: 100}}
	if validETarget(source, minion, 10) {
		t.Fatal("minion should not be a valid e target")
	}

	CastE(w, source, protocol.CastInput{SkillID: eID, TargetID: minion.ID}, state, w.SkillConfig(eID), 10, 20)
	if source.Stats.MP != 100 || source.Position != start || source.Skills[eID].CooldownUntilTick != 0 {
		t.Fatal("non-hero target consumed the skill")
	}

	target.Position = world.Vector2{X: start.X + 800, Y: start.Y}
	CastE(w, source, protocol.CastInput{SkillID: eID, TargetID: target.ID}, state, w.SkillConfig(eID), 10, 20)
	if source.Stats.MP != 100 || source.Position != start || source.Skills[eID].CooldownUntilTick != 0 {
		t.Fatal("out-of-range target consumed the skill")
	}
}

func TestCutthroatWarAmplifiesAllDamageOnlyAgainstMarkedTarget(t *testing.T) {
	w, source, target := qTestWorld(t)
	state := source.Skills[eID]
	state.Level = 5
	source.Skills[eID] = state
	CastE(w, source, protocol.CastInput{SkillID: eID, TargetID: target.ID}, state, w.SkillConfig(eID), 10, 20)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Combat.LastHitTick = 11

	w.ApplyMagicDamage(source, target, 100, 20)
	if got := target.Combat.LastDamage; got != 115 {
		t.Fatalf("marked magic damage = %d, want 115", got)
	}

	other := &world.Entity{ID: "other", Kind: world.EntityKindPlayer, Team: world.TeamRed, Stats: world.Stats{HP: 1000}}
	other.Combat.LastHitTick = 11
	w.ApplyMagicDamage(source, other, 100, 20)
	if got := other.Combat.LastDamage; got != 100 {
		t.Fatalf("unmarked damage = %d, want 100", got)
	}

	target.Stats.HP = 1000
	target.Combat.LastHitTick = 70
	w.ApplyMagicDamage(source, target, 100, 20)
	if got := target.Combat.LastDamage; got != 100 {
		t.Fatalf("expired mark damage = %d, want 100", got)
	}
}
