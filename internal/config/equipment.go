package config

import "fmt"

const MaxEquipmentSlots = 6

type EquipmentConfig struct {
	EquipmentID string           `json:"equipmentId"`
	Name        string           `json:"name"`
	Category    string           `json:"category,omitempty"`
	Tier        int              `json:"tier,omitempty"`
	Description []string         `json:"description,omitempty"`
	Price       int              `json:"price"`
	SellRatio   float64          `json:"sellRatio"`
	UniqueGroup string           `json:"uniqueGroup,omitempty"`
	Components  []string         `json:"components,omitempty"`
	Stats       EquipmentStats   `json:"stats"`
	Effects     EquipmentEffects `json:"effects,omitempty"`
}

type EquipmentStats struct {
	HP               float64 `json:"hp,omitempty"`
	MP               float64 `json:"mp,omitempty"`
	HPRegen5         float64 `json:"hpRegen5,omitempty"`
	MPRegen5         float64 `json:"mpRegen5,omitempty"`
	BaseMPRegenBonus float64 `json:"baseMpRegenBonus,omitempty"`
	Attack           float64 `json:"attack,omitempty"`
	AbilityPower     int     `json:"abilityPower,omitempty"`
	AbilityHaste     float64 `json:"abilityHaste,omitempty"`
	PhysicalDefense  float64 `json:"physicalDefense,omitempty"`
	MagicDefense     float64 `json:"magicDefense,omitempty"`
	MagicPenFlat     float64 `json:"magicPenFlat,omitempty"`
	Tenacity         float64 `json:"tenacity,omitempty"`
	SlowResist       float64 `json:"slowResist,omitempty"`
	BasicAttackBlock float64 `json:"basicAttackBlock,omitempty"`
	CritDamageReduce float64 `json:"critDamageReduce,omitempty"`
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
	HeroDamageManaShieldRatio        float64 `json:"heroDamageManaShieldRatio,omitempty"`
	HeroDamageManaShieldCooldownMS   int     `json:"heroDamageManaShieldCooldownMs,omitempty"`
	AbilityPowerMultiplier           float64 `json:"abilityPowerMultiplier,omitempty"`
	SkillBurnSeconds                 float64 `json:"skillBurnSeconds,omitempty"`
	SkillBurnTickSeconds             float64 `json:"skillBurnTickSeconds,omitempty"`
	SkillBurnFlatDamage              float64 `json:"skillBurnFlatDamage,omitempty"`
	SkillBurnBaseMaxHPRatio          float64 `json:"skillBurnBaseMaxHpRatio,omitempty"`
	SkillBurnAPMaxHPRatioPer100AP    float64 `json:"skillBurnApMaxHpRatioPer100Ap,omitempty"`
	SunfireBurnSeconds               float64 `json:"sunfireBurnSeconds,omitempty"`
	SunfireBurnFlatMin               float64 `json:"sunfireBurnFlatMin,omitempty"`
	SunfireBurnFlatMax               float64 `json:"sunfireBurnFlatMax,omitempty"`
	SunfireBurnBonusHPRatio          float64 `json:"sunfireBurnBonusHpRatio,omitempty"`
	SunfireRadius                    float64 `json:"sunfireRadius,omitempty"`
	SunfireMinionMultiplier          float64 `json:"sunfireMinionMultiplier,omitempty"`
	SunfireMonsterMultiplier         float64 `json:"sunfireMonsterMultiplier,omitempty"`
	SunfireStackDamageBonus          float64 `json:"sunfireStackDamageBonus,omitempty"`
	SunfireMaxStacks                 float64 `json:"sunfireMaxStacks,omitempty"`
	SunfireStackSeconds              float64 `json:"sunfireStackSeconds,omitempty"`
	PhysicalHitArmorShredPerStack    float64 `json:"physicalHitArmorShredPerStack,omitempty"`
	PhysicalHitArmorShredMaxStacks   float64 `json:"physicalHitArmorShredMaxStacks,omitempty"`
	PhysicalHitArmorShredSeconds     float64 `json:"physicalHitArmorShredSeconds,omitempty"`
	PhysicalHitMoveSpeed             float64 `json:"physicalHitMoveSpeed,omitempty"`
	PhysicalHitMoveSpeedSeconds      float64 `json:"physicalHitMoveSpeedSeconds,omitempty"`
	BasicAttackBonusByCritMax        float64 `json:"basicAttackBonusByCritMax,omitempty"`
	ZeroCritChance                   bool    `json:"zeroCritChance,omitempty"`
	BasicAttackAttackSpeedPerStack   float64 `json:"basicAttackAttackSpeedPerStack,omitempty"`
	BasicAttackAttackSpeedMaxStacks  float64 `json:"basicAttackAttackSpeedMaxStacks,omitempty"`
	EveryNthBasicAttackBonusHit      float64 `json:"everyNthBasicAttackBonusHit,omitempty"`
	PhysicalDamageShieldRatio        float64 `json:"physicalDamageShieldRatio,omitempty"`
	PhysicalDamageShieldDecaySeconds float64 `json:"physicalDamageShieldDecaySeconds,omitempty"`
	BasicAttackAttackerSlow          float64 `json:"basicAttackAttackerSlow,omitempty"`
	BasicAttackAttackerSlowSeconds   float64 `json:"basicAttackAttackerSlowSeconds,omitempty"`
	MagicHitMoveSpeedPercentPerStack float64 `json:"magicHitMoveSpeedPercentPerStack,omitempty"`
	MagicHitMagicDefensePerStack     float64 `json:"magicHitMagicDefensePerStack,omitempty"`
	MagicHitMaxStacks                float64 `json:"magicHitMaxStacks,omitempty"`
	StoneplateShieldMaxHPRatio       float64 `json:"stoneplateShieldMaxHpRatio,omitempty"`
	StoneplateResistPercent          float64 `json:"stoneplateResistPercent,omitempty"`
	StoneplateCooldownSeconds        float64 `json:"stoneplateCooldownSeconds,omitempty"`
	CombatHPRegenMaxHPRatio5         float64 `json:"combatHpRegenMaxHpRatio5,omitempty"`
	CombatMPRegenMaxMPRatio5         float64 `json:"combatMpRegenMaxMpRatio5,omitempty"`
	OutOfCombatHPRegenMaxHPRatio5    float64 `json:"outOfCombatHpRegenMaxHpRatio5,omitempty"`
	OutOfCombatMPRegenMaxMPRatio5    float64 `json:"outOfCombatMpRegenMaxMpRatio5,omitempty"`
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
		if item.Category != "" && item.Category != "physical" && item.Category != "magic" && item.Category != "defense" && item.Category != "shoes" {
			return nil, fmt.Errorf("equipment %s category must be physical, magic, defense, or shoes", item.EquipmentID)
		}
		if item.Tier < 0 || item.Tier > 3 {
			return nil, fmt.Errorf("equipment %s tier must be in [1, 3]", item.EquipmentID)
		}
		if item.Tier == 0 {
			item.Tier = 1
			if len(item.Components) > 0 {
				item.Tier = 2
			}
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
			item.Stats.BaseMPRegenBonus < 0 ||
			item.Stats.Attack < 0 || item.Stats.AbilityPower < 0 || item.Stats.AbilityHaste < 0 ||
			item.Stats.PhysicalDefense < 0 || item.Stats.MagicDefense < 0 || item.Stats.MagicPenFlat < 0 ||
			item.Stats.Tenacity < 0 || item.Stats.SlowResist < 0 || item.Stats.BasicAttackBlock < 0 ||
			item.Stats.CritDamageReduce < 0 || item.Stats.MoveSpeed < 0 || item.Stats.MoveSpeedPercent < 0 ||
			item.Stats.AttackSpeedBonus < 0 || item.Stats.CritChance < 0 || item.Stats.Omnivamp < 0 ||
			item.Stats.LifeSteal < 0 {
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
			item.Effects.LowHealthShieldThreshold < 0 || item.Effects.LowHealthDamageReduce < 0 ||
			item.Effects.HeroDamageManaShieldRatio < 0 || item.Effects.HeroDamageManaShieldCooldownMS < 0 ||
			item.Effects.AbilityPowerMultiplier < 0 || item.Effects.SkillBurnSeconds < 0 ||
			item.Effects.SkillBurnTickSeconds < 0 || item.Effects.SkillBurnFlatDamage < 0 ||
			item.Effects.SkillBurnBaseMaxHPRatio < 0 || item.Effects.SkillBurnAPMaxHPRatioPer100AP < 0 ||
			item.Effects.SunfireBurnSeconds < 0 || item.Effects.SunfireBurnFlatMin < 0 ||
			item.Effects.SunfireBurnFlatMax < 0 || item.Effects.SunfireBurnBonusHPRatio < 0 ||
			item.Effects.SunfireRadius < 0 || item.Effects.SunfireMinionMultiplier < 0 ||
			item.Effects.SunfireMonsterMultiplier < 0 || item.Effects.SunfireStackDamageBonus < 0 ||
			item.Effects.SunfireMaxStacks < 0 || item.Effects.SunfireStackSeconds < 0 ||
			item.Effects.PhysicalHitArmorShredPerStack < 0 || item.Effects.PhysicalHitArmorShredMaxStacks < 0 ||
			item.Effects.PhysicalHitArmorShredSeconds < 0 || item.Effects.PhysicalHitMoveSpeed < 0 ||
			item.Effects.PhysicalHitMoveSpeedSeconds < 0 || item.Effects.BasicAttackBonusByCritMax < 0 ||
			item.Effects.BasicAttackAttackSpeedPerStack < 0 || item.Effects.BasicAttackAttackSpeedMaxStacks < 0 ||
			item.Effects.EveryNthBasicAttackBonusHit < 0 || item.Effects.PhysicalDamageShieldRatio < 0 ||
			item.Effects.PhysicalDamageShieldDecaySeconds < 0 ||
			item.Effects.BasicAttackAttackerSlow < 0 || item.Effects.BasicAttackAttackerSlowSeconds < 0 ||
			item.Effects.MagicHitMoveSpeedPercentPerStack < 0 || item.Effects.MagicHitMagicDefensePerStack < 0 ||
			item.Effects.MagicHitMaxStacks < 0 || item.Effects.StoneplateShieldMaxHPRatio < 0 ||
			item.Effects.StoneplateResistPercent < 0 || item.Effects.StoneplateCooldownSeconds < 0 ||
			item.Effects.CombatHPRegenMaxHPRatio5 < 0 || item.Effects.CombatMPRegenMaxMPRatio5 < 0 ||
			item.Effects.OutOfCombatHPRegenMaxHPRatio5 < 0 || item.Effects.OutOfCombatMPRegenMaxMPRatio5 < 0 {
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
