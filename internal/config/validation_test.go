package config

import "testing"

func TestValidateGameConfigLoadsCurrentTables(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes.json")
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
