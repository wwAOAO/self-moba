package config

import "fmt"

type HeroConfig struct {
	HeroID string     `json:"heroId"`
	Name   string     `json:"name"`
	Base   BaseStats  `json:"base"`
	Growth BaseStats  `json:"growth"`
	Radius float64    `json:"radius"`
	Skills HeroSkills `json:"skills"`
}

type BaseStats struct {
	HP              int     `json:"hp"`
	MP              int     `json:"mp"`
	Attack          int     `json:"attack"`
	PhysicalDefense int     `json:"physicalDefense"`
	MagicDefense    int     `json:"magicDefense"`
	MoveSpeed       float64 `json:"moveSpeed"`
	AttackRange     float64 `json:"attackRange"`
	AttackSpeed     float64 `json:"attackSpeed"`
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
		if hero.Base.MP < 0 {
			return nil, fmt.Errorf("hero %s mp must not be negative", hero.HeroID)
		}
		if hero.Base.Attack < 0 {
			return nil, fmt.Errorf("hero %s attack must not be negative", hero.HeroID)
		}
		if hero.Base.PhysicalDefense < 0 {
			return nil, fmt.Errorf("hero %s physical defense must not be negative", hero.HeroID)
		}
		if hero.Base.MagicDefense < 0 {
			return nil, fmt.Errorf("hero %s magic defense must not be negative", hero.HeroID)
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
		if hero.Growth.HP < 0 {
			return nil, fmt.Errorf("hero %s hp growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MP < 0 {
			return nil, fmt.Errorf("hero %s mp growth must not be negative", hero.HeroID)
		}
		if hero.Growth.Attack < 0 {
			return nil, fmt.Errorf("hero %s attack growth must not be negative", hero.HeroID)
		}
		if hero.Growth.PhysicalDefense < 0 {
			return nil, fmt.Errorf("hero %s physical defense growth must not be negative", hero.HeroID)
		}
		if hero.Growth.MagicDefense < 0 {
			return nil, fmt.Errorf("hero %s magic defense growth must not be negative", hero.HeroID)
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
