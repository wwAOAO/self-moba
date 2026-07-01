package config

import (
	"encoding/json"
	"os"
)

func LoadHeroes(path string) (*HeroStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var heroes []HeroConfig
	if err := json.Unmarshal(data, &heroes); err != nil {
		return nil, err
	}
	return NewHeroStore(heroes)
}

func LoadSkills(path string) (*SkillStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var skills []SkillConfig
	if err := json.Unmarshal(data, &skills); err != nil {
		return nil, err
	}
	return NewSkillStore(skills)
}

func LoadLevels(path string) (*LevelConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var levels LevelConfig
	if err := json.Unmarshal(data, &levels); err != nil {
		return nil, err
	}
	return NewLevelConfig(levels)
}

func LoadRewards(path string) (*RewardConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rewards RewardConfig
	if err := json.Unmarshal(data, &rewards); err != nil {
		return nil, err
	}
	return NewRewardConfig(rewards)
}

func LoadEquipment(path string) (*EquipmentStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var equipment []EquipmentConfig
	if err := json.Unmarshal(data, &equipment); err != nil {
		return nil, err
	}
	return NewEquipmentStore(equipment)
}
