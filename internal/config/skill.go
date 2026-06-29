package config

import "fmt"

type SkillConfig struct {
	SkillID    string  `json:"skillId"`
	Name       string  `json:"name"`
	CooldownMS int     `json:"cooldownMs"`
	Range      float64 `json:"range"`
	Type       string  `json:"type"`
}

type SkillStore struct {
	skills map[string]SkillConfig
}

func NewSkillStore(skills []SkillConfig) (*SkillStore, error) {
	store := &SkillStore{
		skills: make(map[string]SkillConfig, len(skills)),
	}

	for _, skill := range skills {
		if skill.SkillID == "" {
			return nil, fmt.Errorf("skill id is required")
		}
		if skill.CooldownMS < 0 {
			return nil, fmt.Errorf("skill %s cooldown must not be negative", skill.SkillID)
		}
		if skill.Range < 0 {
			return nil, fmt.Errorf("skill %s range must not be negative", skill.SkillID)
		}
		if _, exists := store.skills[skill.SkillID]; exists {
			return nil, fmt.Errorf("duplicate skill %s", skill.SkillID)
		}
		store.skills[skill.SkillID] = skill
	}

	if len(store.skills) == 0 {
		return nil, fmt.Errorf("at least one skill is required")
	}
	return store, nil
}

func (s *SkillStore) Get(skillID string) (SkillConfig, bool) {
	skill, ok := s.skills[skillID]
	return skill, ok
}

func (s *SkillStore) Count() int {
	return len(s.skills)
}
