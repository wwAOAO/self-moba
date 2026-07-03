package world

import (
	"math"
	"testing"
)

func TestBladePassiveBasicAttackKillAndCritChance(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	hero.Base.Attack = 1000
	hero.Base.AttackRange = 1000
	w.SpawnHero("blade", hero, TeamBlue)
	player := w.entities[playerEntityID("blade")]
	target := w.entities["minion:red-melee-1"]
	target.Position = Vector2{X: player.Position.X + 100, Y: player.Position.Y}

	w.applyAttack(player, target, 1, 20)
	w.releasePendingAttack(player, 6, 20)

	if player.Stats.MP != 15 {
		t.Fatalf("rage after basic attack kill = %f, want 15", player.Stats.MP)
	}
	if got := w.DisplayCritChance(player); math.Abs(got-0.0525) > 0.000001 {
		t.Fatalf("crit chance = %f, want 0.0525", got)
	}
}

func TestBladePassiveRageCritAttackGrantsExtraRage(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	hero.Base.AttackRange = 1000
	hero.Base.CritChance = 1
	w.SpawnHero("blade", hero, TeamBlue)
	player := w.entities[playerEntityID("blade")]
	target := w.entities["dummy:training-1"]
	target.Position = Vector2{X: player.Position.X + 100, Y: player.Position.Y}

	w.applyAttack(player, target, 1, 20)
	w.releasePendingAttack(player, 6, 20)

	if player.Stats.MP != 10 {
		t.Fatalf("rage after crit basic attack = %f, want 10", player.Stats.MP)
	}
}

func TestBladePassiveRageDecaysAfterOutOfCombat(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", hero, TeamBlue)
	player := w.entities[playerEntityID("blade")]
	player.Stats.MP = 20
	player.Combat.LastHitTick = 100

	w.tickBladeRageDecay(player, 199, 20)
	if player.Stats.MP != 20 {
		t.Fatalf("rage before out of combat = %f, want 20", player.Stats.MP)
	}
	w.tickBladeRageDecay(player, 200, 20)
	if math.Abs(player.Stats.MP-19.75) > 0.000001 {
		t.Fatalf("rage after decay tick = %f, want 19.75", player.Stats.MP)
	}
}
