package config

import "fmt"

type LevelConfig struct {
	MaxLevel int          `json:"maxLevel"`
	TotalExp int          `json:"totalExp"`
	Levels   []LevelEntry `json:"levels"`
	nextExp  map[int]int
}

type LevelEntry struct {
	Level   int `json:"level"`
	NextExp int `json:"nextExp"`
}

func NewLevelConfig(config LevelConfig) (*LevelConfig, error) {
	if config.MaxLevel <= 1 {
		return nil, fmt.Errorf("max level must be greater than 1")
	}
	if len(config.Levels) != config.MaxLevel {
		return nil, fmt.Errorf("level config must contain %d levels", config.MaxLevel)
	}
	config.nextExp = make(map[int]int, len(config.Levels))
	totalExp := 0
	for _, level := range config.Levels {
		if level.Level < 1 || level.Level > config.MaxLevel {
			return nil, fmt.Errorf("level %d out of range", level.Level)
		}
		if _, exists := config.nextExp[level.Level]; exists {
			return nil, fmt.Errorf("duplicate level %d", level.Level)
		}
		if level.Level == config.MaxLevel {
			if level.NextExp != 0 {
				return nil, fmt.Errorf("max level %d next exp must be 0", level.Level)
			}
		} else if level.NextExp <= 0 {
			return nil, fmt.Errorf("level %d next exp must be positive", level.Level)
		}
		config.nextExp[level.Level] = level.NextExp
		totalExp += level.NextExp
	}
	for level := 1; level <= config.MaxLevel; level++ {
		if _, ok := config.nextExp[level]; !ok {
			return nil, fmt.Errorf("missing level %d", level)
		}
	}
	if config.TotalExp != totalExp {
		return nil, fmt.Errorf("total exp = %d, want %d", config.TotalExp, totalExp)
	}
	return &config, nil
}

func (c *LevelConfig) NextExp(level int) (int, bool) {
	if c == nil {
		return 0, false
	}
	nextExp, ok := c.nextExp[level]
	return nextExp, ok
}
