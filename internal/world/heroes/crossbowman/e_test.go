package crossbowman

import (
	"math"
	"testing"

	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func TestCondemnWindupDamageKnockbackAndSilverBoltsHit(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 100
	source.Stats.BonusAttack = 100
	learnCondemn(source, 2)
	learnSilverBolts(source, 1)
	target := spawnCondemnTarget(t, w, 1500, 1000)

	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID, TargetID: target.ID}}, 10, nil, 20)
	if source.Stats.MP != 10 {
		t.Fatalf("mana after condemn = %v, want 10", source.Stats.MP)
	}
	if got, want := source.Skills[eID].CooldownUntilTick, uint64(370); got != want {
		t.Fatalf("rank 2 cooldown = %d, want %d", got, want)
	}
	if !source.Crossbowman.CondemnPending || source.Crossbowman.CondemnReleaseTick != 14 {
		t.Fatalf("pending condemn = %+v, want release at 14", source.Crossbowman)
	}

	w.Tick(13, 20)
	if hasEffect(w.SkillEffects(), "crossbowman_condemn") {
		t.Fatal("condemn projectile released before windup completed")
	}
	w.Tick(14, 20)
	if !hasEffect(w.SkillEffects(), "crossbowman_condemn") {
		t.Fatal("condemn projectile missing after windup")
	}

	hitTick := tickUntilCondemnHit(t, w, target, 15, 30)
	if got, want := target.Combat.LastDamage, 135; got != want {
		t.Fatalf("rank 2 condemn damage = %d, want %d", got, want)
	}
	if target.Stats.HP != 1865 {
		t.Fatalf("target hp after condemn = %v, want 1865", target.Stats.HP)
	}
	if source.Crossbowman.SilverBoltsStacks != 1 || source.Crossbowman.SilverBoltsTargetID != target.ID {
		t.Fatalf("silver bolts after condemn = %+v, want target/1", source.Crossbowman)
	}
	if target.Control.StunnedUntilTick != 0 || source.Crossbowman.CondemnImpactTick != 0 {
		t.Fatalf("open-ground condemn stunned or queued impact: target=%+v source=%+v", target.Control, source.Crossbowman)
	}

	dashUntil := target.Control.DashUntilTick
	for tick := hitTick + 1; tick <= dashUntil; tick++ {
		w.Tick(tick, 20)
	}
	if math.Abs(target.Position.X-1970) > 0.000001 || math.Abs(target.Position.Y-1000) > 0.000001 {
		t.Fatalf("knockback end = %+v, want 1970/1000", target.Position)
	}
}

func TestCondemnWallImpactDealsEqualDamageAndStuns(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 7400, Y: 1000}
	source.Stats.MP = 100
	learnCondemn(source, 1)
	target := spawnCondemnTarget(t, w, 7900, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000

	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID, TargetID: target.ID}}, 10, nil, 20)
	w.Tick(14, 20)
	hitTick := tickUntilCondemnHit(t, w, target, 15, 30)
	if target.Stats.HP != 950 {
		t.Fatalf("hp after initial condemn hit = %v, want 950", target.Stats.HP)
	}
	impactTick := source.Crossbowman.CondemnImpactTick
	if impactTick <= hitTick {
		t.Fatalf("impact tick = %d, want after hit tick %d", impactTick, hitTick)
	}
	for tick := hitTick + 1; tick < impactTick; tick++ {
		w.Tick(tick, 20)
	}
	if target.Stats.HP != 950 || target.Control.StunnedUntilTick != 0 {
		t.Fatalf("wall bonus resolved early: hp=%v stun=%d", target.Stats.HP, target.Control.StunnedUntilTick)
	}

	w.Tick(impactTick, 20)
	if target.Stats.HP != 900 || target.Combat.LastDamage != 50 {
		t.Fatalf("wall impact hp/damage = %v/%d, want 900/50", target.Stats.HP, target.Combat.LastDamage)
	}
	if got, want := target.Control.StunnedUntilTick, impactTick+30; got != want {
		t.Fatalf("wall impact stun until = %d, want %d", got, want)
	}
	if math.Abs(target.Position.X-8000) > 0.000001 || math.Abs(target.Position.Y-1000) > 0.000001 {
		t.Fatalf("wall knockback end = %+v, want 8000/1000", target.Position)
	}
	if !hasEffect(w.SkillEffects(), "crossbowman_condemn_impact") {
		t.Fatal("condemn wall impact effect missing")
	}
}

func TestCondemnProjectileIsBlockedByWindWall(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 100
	learnCondemn(source, 1)
	learnSilverBolts(source, 1)
	target := spawnCondemnTarget(t, w, 1500, 1000)
	w.PutWindWall(world.WindWall{
		ID:        "windwall:condemn",
		Team:      world.TeamRed,
		Center:    world.Vector2{X: 1250, Y: 1000},
		Dir:       world.Vector2{X: 0, Y: 1},
		Width:     400,
		ExpiresAt: 100,
	})

	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID, TargetID: target.ID}}, 10, nil, 20)
	for tick := uint64(14); tick <= 30; tick++ {
		w.Tick(tick, 20)
	}
	if target.Combat.LastDamage != 0 || target.Position != (world.Vector2{X: 1500, Y: 1000}) {
		t.Fatalf("blocked condemn affected target: damage=%d position=%+v", target.Combat.LastDamage, target.Position)
	}
	if source.Crossbowman.SilverBoltsStacks != 0 {
		t.Fatalf("blocked condemn added silver bolts stack: %+v", source.Crossbowman)
	}
	if hasEffect(w.SkillEffects(), "crossbowman_condemn") {
		t.Fatal("blocked condemn projectile was not removed")
	}
}

func TestCondemnDamageAndCooldownByRank(t *testing.T) {
	baseDamage := []int{50, 85, 120, 155, 190}
	cooldownSeconds := []uint64{20, 18, 16, 14, 12}
	for index := range baseDamage {
		w, source := tumbleTestWorld(t)
		source.Position = world.Vector2{X: 1000, Y: 1000}
		source.Stats.MP = 100
		source.Stats.BonusAttack = 100
		learnCondemn(source, index+1)
		target := spawnCondemnTarget(t, w, 1500, 1000)

		if got, want := condemnDamage(w, source, target, w.SkillConfig(eID), index+1, 10), baseDamage[index]+50; got != want {
			t.Fatalf("rank %d damage = %d, want %d", index+1, got, want)
		}
		w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID, TargetID: target.ID}}, 10, nil, 20)
		wantCooldown := uint64(10) + cooldownSeconds[index]*20
		if got := source.Skills[eID].CooldownUntilTick; got != wantCooldown {
			t.Fatalf("rank %d cooldown = %d, want %d", index+1, got, wantCooldown)
		}
	}
}

func TestCondemnRejectsInvalidTargetAndInsufficientMana(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 89
	learnCondemn(source, 1)
	target := spawnCondemnTarget(t, w, 1500, 1000)

	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID, TargetID: target.ID}}, 10, nil, 20)
	if source.Crossbowman.CondemnPending || source.Skills[eID].CooldownUntilTick != 0 || source.Stats.MP != 89 {
		t.Fatalf("low-mana condemn changed state: stats=%+v skill=%+v state=%+v", source.Stats, source.Skills[eID], source.Crossbowman)
	}

	source.Stats.MP = 100
	target.Position = world.Vector2{X: 1700, Y: 1000}
	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID, TargetID: target.ID}}, 11, nil, 20)
	if source.Crossbowman.CondemnPending || source.Skills[eID].CooldownUntilTick != 0 || source.Stats.MP != 100 {
		t.Fatalf("out-of-range condemn changed state: stats=%+v skill=%+v state=%+v", source.Stats, source.Skills[eID], source.Crossbowman)
	}
}

func learnCondemn(source *world.Entity, level int) {
	state := source.Skills[eID]
	state.Level = level
	source.Skills[eID] = state
}

func spawnCondemnTarget(t *testing.T, w *world.World, x float64, y float64) *world.Entity {
	t.Helper()
	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, x, y)
	if !ok {
		t.Fatal("failed to spawn condemn target")
	}
	target := w.EntityByID(targetID)
	target.Stats.HP = 2000
	target.Stats.MaxHP = 2000
	target.Stats.PhysicalDefense = 0
	return target
}

func tickUntilCondemnHit(t *testing.T, w *world.World, target *world.Entity, start uint64, end uint64) uint64 {
	t.Helper()
	for tick := start; tick <= end; tick++ {
		w.Tick(tick, 20)
		if target.Combat.LastDamage > 0 {
			return tick
		}
	}
	t.Fatalf("condemn did not hit by tick %d", end)
	return 0
}
