package battle

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
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
