package config

import "fmt"

const MaxEquipmentSlots = 6

type EquipmentConfig struct {
	EquipmentID string           `json:"equipmentId"`
	Name        string           `json:"name"`
	Price       int              `json:"price"`
	SellRatio   float64          `json:"sellRatio"`
	UniqueGroup string           `json:"uniqueGroup,omitempty"`
	Components  []string         `json:"components,omitempty"`
	Stats       EquipmentStats   `json:"stats"`
	Effects     EquipmentEffects `json:"effects,omitempty"`
}

type EquipmentStats struct {
	HP               int     `json:"hp,omitempty"`
	MP               float64 `json:"mp,omitempty"`
	HPRegen5         float64 `json:"hpRegen5,omitempty"`
	MPRegen5         float64 `json:"mpRegen5,omitempty"`
	Attack           float64 `json:"attack,omitempty"`
	AbilityPower     int     `json:"abilityPower,omitempty"`
	AbilityHaste     float64 `json:"abilityHaste,omitempty"`
	PhysicalDefense  float64 `json:"physicalDefense,omitempty"`
	MagicDefense     float64 `json:"magicDefense,omitempty"`
	MoveSpeed        float64 `json:"moveSpeed,omitempty"`
	MoveSpeedPercent float64 `json:"moveSpeedPercent,omitempty"`
	AttackSpeedBonus float64 `json:"attackSpeedBonus,omitempty"`
	CritChance       float64 `json:"critChance,omitempty"`
	Omnivamp         float64 `json:"omnivamp,omitempty"`
	LifeSteal        float64 `json:"lifeSteal,omitempty"`
}

type EquipmentEffects struct {
	BasicAttackBonusDamage           float64 `json:"basicAttackBonusDamage,omitempty"`
	BasicAttackBonusDamageType       string  `json:"basicAttackBonusDamageType,omitempty"`
	MinionBasicAttackBonusDamage     float64 `json:"minionBasicAttackBonusDamage,omitempty"`
	MinionBasicAttackBonusDamageType string  `json:"minionBasicAttackBonusDamageType,omitempty"`
	HeroHitSmallHeal                 bool    `json:"heroHitSmallHeal,omitempty"`
	HeroHitHeal                      int     `json:"heroHitHeal,omitempty"`
	LevelUpRestoreHPRatio            float64 `json:"levelUpRestoreHpRatio,omitempty"`
	LevelUpRestoreMPRatio            float64 `json:"levelUpRestoreMpRatio,omitempty"`
	OutOfCombatMoveSpeed             float64 `json:"outOfCombatMoveSpeed,omitempty"`
	OutOfCombatSeconds               float64 `json:"outOfCombatSeconds,omitempty"`
	UnitKillPhysicalDefenseGain      float64 `json:"unitKillPhysicalDefenseGain,omitempty"`
	UnitKillAbilityPowerGain         float64 `json:"unitKillAbilityPowerGain,omitempty"`
	UnitKillMaxGain                  float64 `json:"unitKillMaxGain,omitempty"`
	CritDamageBonus                  float64 `json:"critDamageBonus,omitempty"`
	LowHealthShieldMin               int     `json:"lowHealthShieldMin,omitempty"`
	LowHealthShieldMax               int     `json:"lowHealthShieldMax,omitempty"`
	LowHealthShieldThreshold         float64 `json:"lowHealthShieldThreshold,omitempty"`
	LowHealthDamageReduce            float64 `json:"lowHealthDamageReduce,omitempty"`
}

type EquipmentStore struct {
	equipment map[string]EquipmentConfig
}

func NewEquipmentStore(items []EquipmentConfig) (*EquipmentStore, error) {
	store := &EquipmentStore{equipment: make(map[string]EquipmentConfig, len(items))}
	for _, item := range items {
		if item.EquipmentID == "" {
			return nil, fmt.Errorf("equipment id is required")
		}
		if item.Name == "" {
			return nil, fmt.Errorf("equipment %s name is required", item.EquipmentID)
		}
		if item.Price < 0 {
			return nil, fmt.Errorf("equipment %s price must not be negative", item.EquipmentID)
		}
		if item.SellRatio < 0 || item.SellRatio > 1 {
			return nil, fmt.Errorf("equipment %s sell ratio must be in [0, 1]", item.EquipmentID)
		}
		if item.SellRatio == 0 {
			item.SellRatio = 0.5
		}
		if item.Stats.HP < 0 || item.Stats.MP < 0 || item.Stats.HPRegen5 < 0 || item.Stats.MPRegen5 < 0 ||
			item.Stats.Attack < 0 || item.Stats.AbilityPower < 0 || item.Stats.AbilityHaste < 0 ||
			item.Stats.PhysicalDefense < 0 || item.Stats.MagicDefense < 0 || item.Stats.MoveSpeed < 0 ||
			item.Stats.MoveSpeedPercent < 0 || item.Stats.AttackSpeedBonus < 0 || item.Stats.CritChance < 0 ||
			item.Stats.Omnivamp < 0 || item.Stats.LifeSteal < 0 {
			return nil, fmt.Errorf("equipment %s stats must not be negative", item.EquipmentID)
		}
		if item.Effects.BasicAttackBonusDamage < 0 {
			return nil, fmt.Errorf("equipment %s basic attack bonus damage must not be negative", item.EquipmentID)
		}
		if item.Effects.MinionBasicAttackBonusDamage < 0 {
			return nil, fmt.Errorf("equipment %s minion bonus damage must not be negative", item.EquipmentID)
		}
		if item.Effects.HeroHitHeal < 0 {
			return nil, fmt.Errorf("equipment %s hero hit heal must not be negative", item.EquipmentID)
		}
		if item.Effects.LevelUpRestoreHPRatio < 0 || item.Effects.LevelUpRestoreMPRatio < 0 ||
			item.Effects.OutOfCombatMoveSpeed < 0 || item.Effects.OutOfCombatSeconds < 0 ||
			item.Effects.UnitKillPhysicalDefenseGain < 0 || item.Effects.UnitKillAbilityPowerGain < 0 ||
			item.Effects.UnitKillMaxGain < 0 || item.Effects.CritDamageBonus < 0 ||
			item.Effects.LowHealthShieldMin < 0 || item.Effects.LowHealthShieldMax < 0 ||
			item.Effects.LowHealthShieldThreshold < 0 || item.Effects.LowHealthDamageReduce < 0 {
			return nil, fmt.Errorf("equipment %s effects must not be negative", item.EquipmentID)
		}
		if _, exists := store.equipment[item.EquipmentID]; exists {
			return nil, fmt.Errorf("duplicate equipment %s", item.EquipmentID)
		}
		store.equipment[item.EquipmentID] = item
	}
	for _, item := range store.equipment {
		for _, componentID := range item.Components {
			if componentID == "" {
				return nil, fmt.Errorf("equipment %s component id is required", item.EquipmentID)
			}
			if _, ok := store.equipment[componentID]; !ok {
				return nil, fmt.Errorf("equipment %s component %s not found", item.EquipmentID, componentID)
			}
		}
	}
	if len(store.equipment) == 0 {
		return nil, fmt.Errorf("at least one equipment item is required")
	}
	return store, nil
}

func (s *EquipmentStore) Get(equipmentID string) (EquipmentConfig, bool) {
	if s == nil {
		return EquipmentConfig{}, false
	}
	item, ok := s.equipment[equipmentID]
	return item, ok
}

func (s *EquipmentStore) Count() int {
	if s == nil {
		return 0
	}
	return len(s.equipment)
}
