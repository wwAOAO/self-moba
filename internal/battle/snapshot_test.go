package battle

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	_ "l-battle/internal/world/heroes/berserker"
	_ "l-battle/internal/world/heroes/ninja"
	_ "l-battle/internal/world/heroes/sword"
	"testing"
)

func TestSwordCritChanceSnapshotUsesPassiveMultiplier(t *testing.T) {
	w, heroes := testSnapshotWorld(t)
	hero, ok := heroes.Get("sword")
	if !ok {
		t.Fatal("missing sword hero")
	}
	w.SpawnHero("p1", hero, world.TeamBlue)

	w.ApplyInput("p1", protocol.PlayerInput{DebugGold: 2800}, 1, nil, 20)
	w.ApplyInput("p1", protocol.PlayerInput{
		BuyEquipment: &protocol.BuyEquipmentInput{EquipmentID: "phantom_dancer"},
	}, 2, nil, 20)
	snapshot := BuildSnapshot("room-1", 2, w)

	if len(snapshot.Players) != 1 {
		t.Fatalf("players = %d, want 1", len(snapshot.Players))
	}
	if got := snapshot.Players[0].Stats.CritChance; got != 0.6 {
		t.Fatalf("snapshot crit chance = %f, want 0.6", got)
	}
}

func TestUnitSnapshotIncludesNegativeBuffs(t *testing.T) {
	w, heroes := testSnapshotWorld(t)
	hero, ok := heroes.Get("berserker")
	if !ok {
		t.Fatal("missing berserker hero")
	}
	w.SpawnHero("darius", hero, world.TeamBlue)
	source := w.EntityByID("player:darius")
	target := w.EntityByID("enemy:hero-1")
	if source == nil || target == nil {
		t.Fatal("missing test entities")
	}
	target.Combat.LastHitTick = 1
	w.ApplyDamage(source, target, 1, 20)

	snapshot := BuildSnapshot("room-1", 1, w)

	for _, unit := range snapshot.Units {
		if unit.ID != target.ID {
			continue
		}
		if len(unit.Buffs) != 1 || !unit.Buffs[0].Negative || unit.Buffs[0].Name != "出血1层" {
			t.Fatalf("unit buffs = %+v", unit.Buffs)
		}
		return
	}
	t.Fatal("enemy hero unit not found")
}

func TestSnapshotIncludesNinjaPassiveTargetCooldowns(t *testing.T) {
	w, heroes := testSnapshotWorld(t)
	hero, ok := heroes.Get("ninja")
	if !ok {
		t.Fatal("missing ninja hero")
	}
	w.SpawnHero("zed", hero, world.TeamBlue)
	player := w.EntityByID("player:zed")
	player.Passive.NinjaSoulCooldowns = map[string]uint64{"enemy:hero-1": 210}

	snapshot := BuildSnapshot("room-1", 20, w)

	got := snapshot.Players[0].Passive.NinjaSoulCooldowns["enemy:hero-1"]
	if got != 210 {
		t.Fatalf("ninja passive cooldown = %d, want 210", got)
	}
}

func TestSnapshotIncludesNinjaShadow(t *testing.T) {
	w, heroes := testSnapshotWorld(t)
	hero, ok := heroes.Get("ninja")
	if !ok {
		t.Fatal("missing ninja hero")
	}
	w.SpawnHero("zed", hero, world.TeamBlue)
	player := w.EntityByID("player:zed")
	player.Ninja.ShadowPosition = world.Vector2{X: 300, Y: 400}
	player.Ninja.ShadowExpiresAt = 99

	snapshot := BuildSnapshot("room-1", 20, w)

	got := snapshot.Players[0].Ninja
	if got.ShadowX != 300 || got.ShadowY != 400 || got.ShadowExpiresAt != 99 {
		t.Fatalf("ninja shadow snapshot = %+v, want 300/400/99", got)
	}
}

func testSnapshotWorld(t *testing.T) (*world.World, *config.HeroStore) {
	t.Helper()
	heroes, err := config.LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	levels, err := config.LoadLevels("../../configs/levels.json")
	if err != nil {
		t.Fatal(err)
	}
	rewards, err := config.LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}
	equipment, err := config.LoadEquipment("../../configs/equipment")
	if err != nil {
		t.Fatal(err)
	}
	return world.NewWorld(heroes, skills, levels, rewards, equipment), heroes
}
