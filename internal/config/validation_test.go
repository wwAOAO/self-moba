package config

import "testing"

func TestValidateGameConfigLoadsCurrentTables(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	levels, err := LoadLevels("../../configs/levels.json")
	if err != nil {
		t.Fatal(err)
	}
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}
	equipment, err := LoadEquipment("../../configs/equipment")
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateGameConfig(GameConfig{
		Heroes:    heroes,
		Skills:    skills,
		Levels:    levels,
		Rewards:   rewards,
		Equipment: equipment,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestKillerStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("killer")
	if !ok {
		t.Fatal("killer hero is missing")
	}

	if hero.Resource != "none" {
		t.Fatalf("resource = %q, want none", hero.Resource)
	}
	if hero.Base.HP != 602 || hero.Growth.HP != 94 ||
		hero.Base.HPRegen5 != 7.5 || hero.Growth.HPRegen5 != 0.7 ||
		hero.Base.Attack != 53 || hero.Growth.Attack != 3.2 ||
		hero.Base.AttackSpeed != 0.656 || hero.Growth.AttackSpeed != 0.0244 ||
		hero.Base.PhysicalDefense != 28 || hero.Growth.PhysicalDefense != 4 ||
		hero.Base.MagicDefense != 34 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.AttackRange != 125 {
		t.Fatalf("killer stats do not match the configured base attributes: %+v", hero)
	}
}

func TestMonkStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("monk")
	if !ok {
		t.Fatal("monk hero is missing")
	}

	if hero.Resource != "energy" {
		t.Fatalf("resource = %q, want energy", hero.Resource)
	}
	if hero.Base.HP != 428 || hero.Growth.HP != 85 ||
		hero.Base.HPRegen5 != 6.25 || hero.Growth.HPRegen5 != 0.7 ||
		hero.Base.MP != 200 || hero.Base.MPRegen5 != 50 ||
		hero.Base.Attack != 55.8 || hero.Growth.Attack != 3.2 ||
		hero.Base.AttackSpeed != 0.651 || hero.Growth.AttackSpeed != 0.03 ||
		hero.Base.PhysicalDefense != 16 || hero.Growth.PhysicalDefense != 3.7 ||
		hero.Base.MagicDefense != 30 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.MoveSpeed != 350 || hero.Base.AttackRange != 125 {
		t.Fatalf("monk stats do not match the configured base attributes: %+v", hero)
	}
}

func TestButcherStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("butcher")
	if !ok {
		t.Fatal("butcher hero is missing")
	}

	if hero.Resource != "mp" {
		t.Fatalf("resource = %q, want mp", hero.Resource)
	}
	if hero.Base.HP != 432 || hero.Growth.HP != 86 ||
		hero.Base.HPRegen5 != 7.45 || hero.Growth.HPRegen5 != 0.55 ||
		hero.Base.MP != 200 || hero.Growth.MP != 50 ||
		hero.Base.MPRegen5 != 7.45 || hero.Growth.MPRegen5 != 0.7 ||
		hero.Base.Attack != 52 || hero.Growth.Attack != 3.3 ||
		hero.Base.AttackSpeed != 0.613 || hero.Growth.AttackSpeed != 0.0098 ||
		hero.Base.PhysicalDefense != 12 || hero.Growth.PhysicalDefense != 1.25 ||
		hero.Base.MagicDefense != 30 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.MoveSpeed != 325 || hero.Base.AttackRange != 175 {
		t.Fatalf("butcher stats do not match the configured base attributes: %+v", hero)
	}
}

func TestValidateEquipmentComponentsRejectsCycles(t *testing.T) {
	equipment, err := NewEquipmentStore([]EquipmentConfig{
		{EquipmentID: "a", Name: "A", Price: 100, Tier: 2, Components: []string{"b"}},
		{EquipmentID: "b", Name: "B", Price: 100, Tier: 2, Components: []string{"a"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateEquipmentComponents(equipment); err == nil {
		t.Fatal("cycle should be rejected")
	}
}

func TestValidateEquipmentComponentsRejectsHigherTierComponents(t *testing.T) {
	equipment, err := NewEquipmentStore([]EquipmentConfig{
		{EquipmentID: "tier2", Name: "Tier 2", Price: 100, Tier: 2, Components: []string{"tier3"}},
		{EquipmentID: "tier3", Name: "Tier 3", Price: 100, Tier: 3},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateEquipmentComponents(equipment); err == nil {
		t.Fatal("higher-tier component should be rejected")
	}
}
