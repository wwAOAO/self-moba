package config

import "fmt"

type HeroConfig struct {
	HeroID   string     `json:"heroId"`
	Name     string     `json:"name"`
	Resource string     `json:"resource"`
	Base     BaseStats  `json:"base"`
	Growth   BaseStats  `json:"growth"`
	Radius   float64    `json:"radius"`
	Skills   HeroSkills `json:"skills"`
}

type BaseStats struct {
	HP                   int     `json:"hp"`
	BonusHP              int     `json:"bonusHp"`
	MP                   int     `json:"mp"`
	HPRegen5             float64 `json:"hpRegen5"`
	Attack               float64 `json:"attack"`
	BonusAttack          float64 `json:"bonusAttack"`
	AbilityPower         int     `json:"abilityPower"`
	DamageReduce         float64 `json:"damageReduce"`
	PhysicalDefense      float64 `json:"physicalDefense"`
	BonusPhysicalDefense float64 `json:"bonusPhysicalDefense"`
	PhysicalPenPercent   float64 `json:"physicalPenPercent"`
	PhysicalPenFlat      float64 `json:"physicalPenFlat"`
	PhysicalDamageReduce float64 `json:"physicalDamageReduce"`
	MagicDefense         float64 `json:"magicDefense"`
	BonusMagicDefense    float64 `json:"bonusMagicDefense"`
	MagicPenPercent      float64 `json:"magicPenPercent"`
	MagicPenFlat         float64 `json:"magicPenFlat"`
	MagicDamageReduce    float64 `json:"magicDamageReduce"`
	MoveSpeed            float64 `json:"moveSpeed"`
	AttackRange          float64 `json:"attackRange"`
	AttackSpeed          float64 `json:"attackSpeed"`
	AttackSpeedRatio     float64 `json:"attackSpeedRatio"`
	BonusAttackSpeed     float64 `json:"bonusAttackSpeed"`
	AttackSpeedSlow      float64 `json:"attackSpeedSlow"`
	CritChance           float64 `json:"critChance"`
}

type HeroSkills struct {
	Passive string `json:"passive"`
	Q       string `json:"q"`
	W       string `json:"w"`
	E       string `json:"e"`
	R       string `json:"r"`
}

func (s HeroSkills) ActiveSkillIDs() []string {
	return []string{s.Q, s.W, s.E, s.R}
}

func (s HeroSkills) SkillIDs() []string {
	return []string{s.Passive, s.Q, s.W, s.E, s.R}
}

func (s HeroSkills) SkillIDForSlot(slot string) string {
	switch slot {
	case "q":
		return s.Q
	case "w":
		return s.W
	case "e":
		return s.E
	case "r":
		return s.R
	default:
		return ""
	}
}

type HeroStore struct {
	heroes map[string]HeroConfig
}

func NewHeroStore(heroes []HeroConfig) (*HeroStore, error) {
	store := &HeroStore{
		heroes: make(map[string]HeroConfig, len(heroes)),
	}

	for _, hero := range heroes {
		if hero.HeroID == "" {
			return nil, fmt.Errorf("hero id is required")
		}
		if hero.Base.HP <= 0 {
			return nil, fmt.Errorf("hero %s hp must be positive", hero.HeroID)
		}
		if hero.Base.BonusHP < 0 {
			return nil, fmt.Errorf("hero %s bonus hp must not be negative", hero.HeroID)
		}
		if hero.Base.MP < 0 {
			return nil, fmt.Errorf("hero %s mp must not be negative", hero.HeroID)
		}
		if hero.Base.HPRegen5 < 0 {
			return nil, fmt.Errorf("hero %s hp regen must not be negative", hero.HeroID)
		}
		if hero.Base.Attack < 0 {
			return nil, fmt.Errorf("hero %s attack must not be negative", hero.HeroID)
		}
		if hero.Base.BonusAttack < 0 {
			return nil, fmt.Errorf("hero %s bonus attack must not be negative", hero.HeroID)
		}
		if hero.Base.AbilityPower < 0 {
			return nil, fmt.Errorf("hero %s ability power must not be negative", hero.HeroID)
		}
		if hero.Base.DamageReduce < 0 || hero.Base.DamageReduce > 1 {
			return nil, fmt.Errorf("hero %s damage reduce must be in [0, 1]", hero.HeroID)
		}
		if hero.Base.PhysicalDefense < 0 {
			return nil, fmt.Errorf("hero %s physical defense must not be negative", hero.HeroID)
		}
		if hero.Base.BonusPhysicalDefense < 0 {
			return nil, fmt.Errorf("hero %s bonus physical defense must not be negative", hero.HeroID)
		}
		if hero.Base.PhysicalPenPercent < 0 || hero.Base.PhysicalPenPercent > 1 {
			return nil, fmt.Errorf("hero %s physical pen percent must be in [0, 1]", hero.HeroID)
		}
		if hero.Base.PhysicalPenFlat < 0 {
			return nil, fmt.Errorf("hero %s physical pen flat must not be negative", hero.HeroID)
		}
		if hero.Base.PhysicalDamageReduce < 0 || hero.Base.PhysicalDamageReduce > 1 {
			return nil, fmt.Errorf("hero %s physical damage reduce must be in [0, 1]", hero.HeroID)
		}
		if hero.Base.MagicDefense < 0 {
			return nil, fmt.Errorf("hero %s magic defense must not be negative", hero.HeroID)
		}
		if hero.Base.BonusMagicDefense < 0 {
			return nil, fmt.Errorf("hero %s bonus magic defense must not be negative", hero.HeroID)
		}
		if hero.Base.MagicPenPercent < 0 || hero.Base.MagicPenPercent > 1 {
			return nil, fmt.Errorf("hero %s magic pen percent must be in [0, 1]", hero.HeroID)
		}
		if hero.Base.MagicPenFlat < 0 {
			return nil, fmt.Errorf("hero %s magic pen flat must not be negative", hero.HeroID)
		}
		if hero.Base.MagicDamageReduce < 0 || hero.Base.MagicDamageReduce > 1 {
			return nil, fmt.Errorf("hero %s magic damage reduce must be in [0, 1]", hero.HeroID)
		}
		if hero.Base.MoveSpeed <= 0 {
			return nil, fmt.Errorf("hero %s move speed must be positive", hero.HeroID)
		}
		if hero.Base.AttackRange < 0 {
			return nil, fmt.Errorf("hero %s attack range must not be negative", hero.HeroID)
		}
		if hero.Base.AttackSpeed <= 0 {
			return nil, fmt.Errorf("hero %s attack speed must be positive", hero.HeroID)
		}
		if hero.Base.AttackSpeedRatio < 0 {
			return nil, fmt.Errorf("hero %s attack speed ratio must not be negative", hero.HeroID)
		}
		if hero.Base.BonusAttackSpeed < 0 {
			return nil, fmt.Errorf("hero %s bonus attack speed must not be negative", hero.HeroID)
		}
		if hero.Base.AttackSpeedSlow < 0 || hero.Base.AttackSpeedSlow > 1 {
			return nil, fmt.Errorf("hero %s attack speed slow must be in [0, 1]", hero.HeroID)
		}
		if hero.Base.CritChance < 0 || hero.Base.CritChance > 1 {
			return nil, fmt.Errorf("hero %s crit chance must be in [0, 1]", hero.HeroID)
		}
		if hero.Growth.HP < 0 {
			return nil, fmt.Errorf("hero %s hp growth must not be negative", hero.HeroID)
		}
		if hero.Growth.BonusHP < 0 {
			return nil, fmt.Errorf("hero %s bonus hp growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MP < 0 {
			return nil, fmt.Errorf("hero %s mp growth must not be negative", hero.HeroID)
		}
		if hero.Growth.HPRegen5 < 0 {
			return nil, fmt.Errorf("hero %s hp regen growth must not be negative", hero.HeroID)
		}
		if hero.Growth.Attack < 0 {
			return nil, fmt.Errorf("hero %s attack growth must not be negative", hero.HeroID)
		}
		if hero.Growth.BonusAttack < 0 {
			return nil, fmt.Errorf("hero %s bonus attack growth must not be negative", hero.HeroID)
		}
		if hero.Growth.AbilityPower < 0 {
			return nil, fmt.Errorf("hero %s ability power growth must not be negative", hero.HeroID)
		}
		if hero.Growth.DamageReduce < 0 {
			return nil, fmt.Errorf("hero %s damage reduce growth must not be negative", hero.HeroID)
		}
		if hero.Growth.PhysicalDefense < 0 {
			return nil, fmt.Errorf("hero %s physical defense growth must not be negative", hero.HeroID)
		}
		if hero.Growth.BonusPhysicalDefense < 0 {
			return nil, fmt.Errorf("hero %s bonus physical defense growth must not be negative", hero.HeroID)
		}
		if hero.Growth.PhysicalPenPercent < 0 {
			return nil, fmt.Errorf("hero %s physical pen percent growth must not be negative", hero.HeroID)
		}
		if hero.Growth.PhysicalPenFlat < 0 {
			return nil, fmt.Errorf("hero %s physical pen flat growth must not be negative", hero.HeroID)
		}
		if hero.Growth.PhysicalDamageReduce < 0 {
			return nil, fmt.Errorf("hero %s physical damage reduce growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MagicDefense < 0 {
			return nil, fmt.Errorf("hero %s magic defense growth must not be negative", hero.HeroID)
		}
		if hero.Growth.BonusMagicDefense < 0 {
			return nil, fmt.Errorf("hero %s bonus magic defense growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MagicPenPercent < 0 {
			return nil, fmt.Errorf("hero %s magic pen percent growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MagicPenFlat < 0 {
			return nil, fmt.Errorf("hero %s magic pen flat growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MagicDamageReduce < 0 {
			return nil, fmt.Errorf("hero %s magic damage reduce growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MoveSpeed < 0 {
			return nil, fmt.Errorf("hero %s move speed growth must not be negative", hero.HeroID)
		}
		if hero.Growth.AttackRange < 0 {
			return nil, fmt.Errorf("hero %s attack range growth must not be negative", hero.HeroID)
		}
		if hero.Growth.AttackSpeed < 0 {
			return nil, fmt.Errorf("hero %s attack speed growth must not be negative", hero.HeroID)
		}
		if hero.Growth.CritChance < 0 {
			return nil, fmt.Errorf("hero %s crit chance growth must not be negative", hero.HeroID)
		}
		if hero.Radius <= 0 {
			return nil, fmt.Errorf("hero %s radius must be positive", hero.HeroID)
		}
		if _, exists := store.heroes[hero.HeroID]; exists {
			return nil, fmt.Errorf("duplicate hero %s", hero.HeroID)
		}
		store.heroes[hero.HeroID] = hero
	}

	if len(store.heroes) == 0 {
		return nil, fmt.Errorf("at least one hero is required")
	}
	return store, nil
}

func (s *HeroStore) Get(heroID string) (HeroConfig, bool) {
	hero, ok := s.heroes[heroID]
	return hero, ok
}

func (s *HeroStore) Count() int {
	return len(s.heroes)
}

func (s *HeroStore) All() []HeroConfig {
	heroes := make([]HeroConfig, 0, len(s.heroes))
	for _, hero := range s.heroes {
		heroes = append(heroes, hero)
	}
	return heroes
}

func ValidateHeroSkills(heroes *HeroStore, skills *SkillStore) error {
	for _, hero := range heroes.All() {
		for _, skillID := range hero.Skills.SkillIDs() {
			if skillID == "" {
				return fmt.Errorf("hero %s must have passive, q, w, e and r skills", hero.HeroID)
			}
			if _, ok := skills.Get(skillID); !ok {
				return fmt.Errorf("hero %s references missing skill %s", hero.HeroID, skillID)
			}
		}
	}
	return nil
}
