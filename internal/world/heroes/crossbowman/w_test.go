package crossbowman

import (
	"testing"

	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func TestSilverBoltsTriggersOnThirdBasicOrSkillHit(t *testing.T) {
	w, source := tumbleTestWorld(t)
	learnSilverBolts(source, 1)
	target := &world.Entity{
		ID:   "target",
		Kind: world.EntityKindEnemyHero,
		Team: world.TeamRed,
		Stats: world.Stats{
			HP:    1000,
			MaxHP: 1000,
		},
	}

	OnBasicHit(w, source, target, 10, 20)
	if target.Stats.HP != 1000 || source.Crossbowman.SilverBoltsStacks != 1 {
		t.Fatalf("first hit hp/stacks = %v/%d, want 1000/1", target.Stats.HP, source.Crossbowman.SilverBoltsStacks)
	}
	ApplySilverBoltsHit(w, source, target, 11, 20)
	if target.Stats.HP != 1000 || source.Crossbowman.SilverBoltsStacks != 2 {
		t.Fatalf("second hit hp/stacks = %v/%d, want 1000/2", target.Stats.HP, source.Crossbowman.SilverBoltsStacks)
	}
	OnBasicHit(w, source, target, 12, 20)
	if target.Stats.HP != 940 {
		t.Fatalf("third hit hp = %v, want 940 after 60 true damage", target.Stats.HP)
	}
	if source.Crossbowman.SilverBoltsTargetID != "" || source.Crossbowman.SilverBoltsStacks != 0 {
		t.Fatalf("state after proc = %+v, want cleared", source.Crossbowman)
	}
	if target.Combat.LastDamage != 60 || target.Combat.LastDamageType != "true" {
		t.Fatalf("third hit damage = %d %s, want 60 true", target.Combat.LastDamage, target.Combat.LastDamageType)
	}
	if !hasEffect(w.SkillEffects(), "crossbowman_silver_bolts_proc") {
		t.Fatal("silver bolts proc effect was not created")
	}
	if hasEffect(w.SkillEffects(), "crossbowman_silver_bolts") {
		t.Fatal("persistent silver bolts mark remained after proc")
	}
}

func TestSilverBoltsLastsFiveSecondsRefreshesAndKeepsEffect(t *testing.T) {
	w, source := tumbleTestWorld(t)
	learnSilverBolts(source, 1)
	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1500, 1000)
	if !ok {
		t.Fatal("failed to spawn silver bolts target")
	}
	target := w.EntityByID(targetID)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000

	ApplySilverBoltsHit(w, source, target, 10, 20)
	firstEffectID := source.Crossbowman.SilverBoltsEffectID
	if got, want := source.Crossbowman.SilverBoltsUntilTick, uint64(110); got != want {
		t.Fatalf("first mark expiry = %d, want %d", got, want)
	}
	assertSilverBoltsEffect(t, w, target.ID, firstEffectID, 1, 110)
	buffs := ActiveBuffs(w, target, 10)
	if len(buffs) != 1 || !buffs[0].Negative || buffs[0].Stacks != 1 || buffs[0].ExpiresAtTick != 110 {
		t.Fatalf("target silver bolts buff = %+v", buffs)
	}

	ApplySilverBoltsHit(w, source, target, 50, 20)
	if got, want := source.Crossbowman.SilverBoltsUntilTick, uint64(150); got != want {
		t.Fatalf("refreshed mark expiry = %d, want %d", got, want)
	}
	assertSilverBoltsEffect(t, w, target.ID, firstEffectID, 2, 150)
	w.Tick(149, 20)
	if source.Crossbowman.SilverBoltsStacks != 2 || !hasEffect(w.SkillEffects(), "crossbowman_silver_bolts") {
		t.Fatal("silver bolts state or effect disappeared before five seconds elapsed")
	}
	w.Tick(150, 20)
	if source.Crossbowman.SilverBoltsStacks != 0 || source.Crossbowman.SilverBoltsTargetID != "" || hasEffect(w.SkillEffects(), "crossbowman_silver_bolts") {
		t.Fatalf("expired silver bolts state/effect was not cleared: %+v", source.Crossbowman)
	}
	if buffs := ActiveBuffs(w, target, 150); len(buffs) != 0 {
		t.Fatalf("expired target buff = %+v, want none", buffs)
	}
}

func TestSilverBoltsUsesMinimumAndPercentageByRank(t *testing.T) {
	w, _ := tumbleTestWorld(t)
	skill := w.SkillConfig(wID)
	tests := []struct {
		level int
		maxHP float64
		want  float64
	}{
		{level: 1, maxHP: 500, want: 50},
		{level: 1, maxHP: 1000, want: 60},
		{level: 2, maxHP: 1000, want: 70},
		{level: 3, maxHP: 1000, want: 80},
		{level: 4, maxHP: 1000, want: 95},
		{level: 5, maxHP: 2000, want: 200},
	}
	for _, test := range tests {
		target := &world.Entity{Kind: world.EntityKindEnemyHero, Team: world.TeamRed, Stats: world.Stats{MaxHP: test.maxHP}}
		if got := silverBoltsDamage(target, skill, test.level); got != test.want {
			t.Fatalf("rank %d at %v max hp damage = %v, want %v", test.level, test.maxHP, got, test.want)
		}
	}
}

func TestSilverBoltsMonsterDamageCapsByRank(t *testing.T) {
	w, _ := tumbleTestWorld(t)
	skill := w.SkillConfig(wID)
	wants := []float64{140, 155, 170, 185, 200}
	for level, want := range wants {
		monster := &world.Entity{
			Kind: world.EntityKindBaronNashor,
			Team: world.TeamNeutral,
			Stats: world.Stats{
				MaxHP: 10000,
			},
		}
		if got := silverBoltsDamage(monster, skill, level+1); got != want {
			t.Fatalf("rank %d monster damage = %v, want cap %v", level+1, got, want)
		}
	}
}

func TestSilverBoltsSwitchingTargetsResetsSequence(t *testing.T) {
	w, source := tumbleTestWorld(t)
	learnSilverBolts(source, 1)
	targetA := silverBoltsTarget("a")
	targetB := silverBoltsTarget("b")

	ApplySilverBoltsHit(w, source, targetA, 10, 20)
	ApplySilverBoltsHit(w, source, targetA, 11, 20)
	ApplySilverBoltsHit(w, source, targetB, 12, 20)
	if source.Crossbowman.SilverBoltsTargetID != "b" || source.Crossbowman.SilverBoltsStacks != 1 {
		t.Fatalf("state after switching to B = %+v, want b/1", source.Crossbowman)
	}
	ApplySilverBoltsHit(w, source, targetA, 13, 20)
	if source.Crossbowman.SilverBoltsTargetID != "a" || source.Crossbowman.SilverBoltsStacks != 1 {
		t.Fatalf("state after switching back to A = %+v, want a/1", source.Crossbowman)
	}
	if targetA.Stats.HP != 1000 || targetB.Stats.HP != 1000 {
		t.Fatalf("switching targets triggered damage: A=%v B=%v", targetA.Stats.HP, targetB.Stats.HP)
	}
}

func TestSilverBoltsCannotBeActivelyCast(t *testing.T) {
	w, source := tumbleTestWorld(t)
	learnSilverBolts(source, 1)
	source.Stats.MP = 100
	source.Combat.NextAttackTick = 77

	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID, TargetX: 2000, TargetY: 1000}}, 10, nil, 20)

	if source.Stats.MP != 100 || source.Combat.NextAttackTick != 77 || source.Control.ActionLockedUntilTick != 0 {
		t.Fatalf("passive W cast changed entity state: mp=%v combat=%+v control=%+v", source.Stats.MP, source.Combat, source.Control)
	}
	if source.Skills[wID].CooldownUntilTick != 0 {
		t.Fatalf("passive W entered cooldown: %+v", source.Skills[wID])
	}
}

func TestSilverBoltsUnlearnedDoesNotStack(t *testing.T) {
	w, source := tumbleTestWorld(t)
	target := silverBoltsTarget("target")
	ApplySilverBoltsHit(w, source, target, 10, 20)
	if source.Crossbowman.SilverBoltsStacks != 0 || target.Stats.HP != 1000 {
		t.Fatalf("unlearned W state=%+v hp=%v", source.Crossbowman, target.Stats.HP)
	}
}

func learnSilverBolts(source *world.Entity, level int) {
	state := source.Skills[wID]
	state.Level = level
	source.Skills[wID] = state
}

func silverBoltsTarget(id string) *world.Entity {
	return &world.Entity{
		ID:   id,
		Kind: world.EntityKindEnemyHero,
		Team: world.TeamRed,
		Stats: world.Stats{
			HP:    1000,
			MaxHP: 1000,
		},
	}
}

func assertSilverBoltsEffect(t *testing.T, w *world.World, targetID string, effectID string, stacks int, expiresAt uint64) {
	t.Helper()
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "crossbowman_silver_bolts" {
			if effect.ID != effectID || effect.TargetID != targetID || effect.Count != stacks || effect.ExpiresAt != expiresAt {
				t.Fatalf("silver bolts effect = %+v, want id=%s target=%s stacks=%d expires=%d", effect, effectID, targetID, stacks, expiresAt)
			}
			return
		}
	}
	t.Fatal("persistent silver bolts effect is missing")
}
