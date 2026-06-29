package config

import (
	"math"
	"testing"
)

func TestRewardConfigMinionSharedExp(t *testing.T) {
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}

	assertFloat(t, rewards, "melee_minion", 2, 38.44)
	assertFloat(t, rewards, "ranged_minion", 2, 19.22)
	assertFloat(t, rewards, "siege_minion", 2, 46.5)
	assertFloat(t, rewards, "melee_minion", 3, 62*1.24/3)
	if rewards.Minion.ShareRadius != 1500 {
		t.Fatalf("minion share radius = %f, want 1500", rewards.Minion.ShareRadius)
	}
}

func TestRewardConfigJungleExp(t *testing.T) {
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]int{
		"blue_buff":     100,
		"red_buff":      100,
		"gromp":         130,
		"murk_wolf":     100,
		"raptor":        105,
		"krug_camp":     160,
		"rift_scuttler": 70,
	}
	for kind, want := range cases {
		got, ok := rewards.JungleExp(kind)
		if !ok || got != want {
			t.Fatalf("%s exp = %d, ok=%v, want %d", kind, got, ok, want)
		}
	}
	if rewards.JungleScaling.StartAverageLevel != 3 {
		t.Fatalf("jungle scaling start = %d, want 3", rewards.JungleScaling.StartAverageLevel)
	}
	if rewards.JungleScaling.CapAverageLevel != 9 {
		t.Fatalf("jungle scaling cap = %d, want 9", rewards.JungleScaling.CapAverageLevel)
	}
	if rewards.JungleScaling.MaxMultiplier != 1.5 {
		t.Fatalf("jungle scaling max multiplier = %f, want 1.5", rewards.JungleScaling.MaxMultiplier)
	}
}

func TestRewardConfigEpicRewards(t *testing.T) {
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}

	herald, ok := rewards.EpicReward("rift_herald")
	if !ok {
		t.Fatal("rift herald reward not found")
	}
	if herald.ParticipantExp != 600 || herald.NonParticipantTeamPoolExp != 800 || herald.NonParticipantSplit != "equal" {
		t.Fatalf("unexpected herald reward: %+v", herald)
	}

	dragon, ok := rewards.EpicReward("elemental_dragon")
	if !ok {
		t.Fatal("elemental dragon reward not found")
	}
	if dragon.MinExp != 150 || dragon.MaxExp != 510 || dragon.Split != "nearby_allies" || !dragon.CatchUpBonus {
		t.Fatalf("unexpected dragon reward: %+v", dragon)
	}

	baron, ok := rewards.EpicReward("baron_nashor")
	if !ok {
		t.Fatal("baron reward not found")
	}
	if baron.Split != "team_equal" || !baron.ScalesWithGameTime {
		t.Fatalf("unexpected baron reward: %+v", baron)
	}
}

func TestRewardConfigHeroKillExp(t *testing.T) {
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}

	assertNear(t, rewards.HeroKillExp(1150, 7, 7), 862.5)
	assertNear(t, rewards.HeroKillExp(1150, 8, 7), 690)
	assertNear(t, rewards.HeroKillExp(1150, 12, 7), 345)
	assertNear(t, rewards.HeroKillExp(1150, 6, 7), 1035)
	assertNear(t, rewards.HeroKillExp(1150, 3, 7), 1380)

	if rewards.HeroKill.NearbyRadius != 1600 {
		t.Fatalf("nearby radius = %f, want 1600", rewards.HeroKill.NearbyRadius)
	}
	if rewards.HeroKill.DeadGraceSeconds != 10 {
		t.Fatalf("dead grace seconds = %d, want 10", rewards.HeroKill.DeadGraceSeconds)
	}
	if !rewards.HeroKill.NearbyAliveHeroShare {
		t.Fatal("nearby alive hero share should be true")
	}
}

func TestRewardConfigAssistExpMultiplier(t *testing.T) {
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}

	cases := map[int]float64{
		1:  0.66,
		6:  0.66,
		7:  0.82,
		8:  0.82,
		9:  0.9,
		18: 0.9,
	}
	for level, want := range cases {
		got, ok := rewards.AssistExpMultiplier(level)
		if !ok {
			t.Fatalf("assist multiplier for level %d not found", level)
		}
		assertNear(t, got, want)
	}
}

func TestRewardConfigStructureTeamExp(t *testing.T) {
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}

	got, ok := rewards.StructureTeamExp("tower")
	if !ok || got != 300 {
		t.Fatalf("tower team exp = %d, ok=%v, want 300", got, ok)
	}
}

func assertFloat(t *testing.T, rewards *RewardConfig, kind string, players int, want float64) {
	t.Helper()
	got, ok := rewards.MinionExp(kind, players)
	if !ok {
		t.Fatalf("%s exp not found", kind)
	}
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("%s players=%d exp = %.4f, want %.4f", kind, players, got, want)
	}
}

func assertNear(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("got %.4f, want %.4f", got, want)
	}
}
