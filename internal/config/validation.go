package config

import "fmt"

type GameConfig struct {
	Heroes    *HeroStore
	Skills    *SkillStore
	Levels    *LevelConfig
	Rewards   *RewardConfig
	Equipment *EquipmentStore
}

func ValidateGameConfig(cfg GameConfig) error {
	if cfg.Heroes == nil {
		return fmt.Errorf("hero config is required")
	}
	if cfg.Skills == nil {
		return fmt.Errorf("skill config is required")
	}
	if cfg.Levels == nil {
		return fmt.Errorf("level config is required")
	}
	if cfg.Rewards == nil {
		return fmt.Errorf("reward config is required")
	}
	if cfg.Equipment == nil {
		return fmt.Errorf("equipment config is required")
	}
	if err := ValidateHeroSkills(cfg.Heroes, cfg.Skills); err != nil {
		return err
	}
	if err := ValidateEquipmentComponents(cfg.Equipment); err != nil {
		return err
	}
	return nil
}

func ValidateEquipmentComponents(store *EquipmentStore) error {
	if store == nil {
		return fmt.Errorf("equipment config is required")
	}
	visiting := make(map[string]bool, len(store.equipment))
	visited := make(map[string]bool, len(store.equipment))
	for id := range store.equipment {
		if err := validateEquipmentComponent(id, store, visiting, visited); err != nil {
			return err
		}
	}
	return nil
}

func validateEquipmentComponent(id string, store *EquipmentStore, visiting map[string]bool, visited map[string]bool) error {
	if visited[id] {
		return nil
	}
	if visiting[id] {
		return fmt.Errorf("equipment %s has circular component dependency", id)
	}
	item, ok := store.equipment[id]
	if !ok {
		return fmt.Errorf("equipment %s not found", id)
	}
	visiting[id] = true
	for _, componentID := range item.Components {
		component, ok := store.equipment[componentID]
		if !ok {
			return fmt.Errorf("equipment %s component %s not found", id, componentID)
		}
		if component.Tier > item.Tier {
			return fmt.Errorf("equipment %s tier %d cannot use higher-tier component %s tier %d", id, item.Tier, componentID, component.Tier)
		}
		if err := validateEquipmentComponent(componentID, store, visiting, visited); err != nil {
			return err
		}
	}
	visiting[id] = false
	visited[id] = true
	return nil
}
