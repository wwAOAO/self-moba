package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func LoadHeroes(path string) (*HeroStore, error) {
	heroes, err := loadJSONList[HeroConfig](path)
	if err != nil {
		return nil, err
	}
	return NewHeroStore(heroes)
}

func LoadSkills(path string) (*SkillStore, error) {
	skills, err := loadJSONList[SkillConfig](path)
	if err != nil {
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
	equipment, err := loadJSONList[EquipmentConfig](path)
	if err != nil {
		return nil, err
	}
	return NewEquipmentStore(equipment)
}

func loadJSONList[T any](path string) ([]T, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return readJSONListFile[T](path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var out []T
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" || entry.Name() == "manifest.json" {
			continue
		}
		items, err := readJSONListFile[T](filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func readJSONListFile[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}
