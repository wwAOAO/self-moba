package config

import "fmt"

type RewardConfig struct {
	Minion        RewardGroup           `json:"minion"`
	Jungle        RewardGroup           `json:"jungle"`
	JungleScaling JungleScalingReward   `json:"jungleScaling"`
	Epic          map[string]EpicReward `json:"epic"`
	HeroKill      HeroKillReward        `json:"heroKill"`
	Structure     StructureReward       `json:"structure"`
}

type RewardGroup struct {
	ShareMultiplier float64            `json:"shareMultiplier,omitempty"`
	ShareMinPlayers int                `json:"shareMinPlayers,omitempty"`
	ShareRadius     float64            `json:"shareRadius,omitempty"`
	KillExp         map[string]float64 `json:"killExp"`
	KillGold        map[string]int     `json:"killGold,omitempty"`
}

type StructureReward struct {
	TeamExp  map[string]int `json:"teamExp"`
	TeamGold map[string]int `json:"teamGold,omitempty"`
}

type JungleScalingReward struct {
	StartAverageLevel int     `json:"startAverageLevel"`
	CapAverageLevel   int     `json:"capAverageLevel"`
	MaxMultiplier     float64 `json:"maxMultiplier"`
}

type EpicReward struct {
	ParticipantExp            int    `json:"participantExp,omitempty"`
	NonParticipantTeamPoolExp int    `json:"nonParticipantTeamPoolExp,omitempty"`
	NonParticipantSplit       string `json:"nonParticipantSplit,omitempty"`
	TeamGold                  int    `json:"teamGold,omitempty"`
	MinExp                    int    `json:"minExp,omitempty"`
	MaxExp                    int    `json:"maxExp,omitempty"`
	Split                     string `json:"split,omitempty"`
	CatchUpBonus              bool   `json:"catchUpBonus,omitempty"`
	ScalesWithGameTime        bool   `json:"scalesWithGameTime,omitempty"`
}

type HeroKillReward struct {
	Gold                         int             `json:"gold,omitempty"`
	BaseNextLevelExpMultiplier   float64         `json:"baseNextLevelExpMultiplier"`
	LevelDiffStepMultiplier      float64         `json:"levelDiffStepMultiplier"`
	HigherLevelMinMultiplier     float64         `json:"higherLevelMinMultiplier"`
	LowerLevelMaxBonusMultiplier float64         `json:"lowerLevelMaxBonusMultiplier"`
	NearbyRadius                 float64         `json:"nearbyRadius"`
	DeadGraceSeconds             int             `json:"deadGraceSeconds"`
	NearbyAliveHeroShare         bool            `json:"nearbyAliveHeroShare"`
	AssistExpTiers               []AssistExpTier `json:"assistExpTiers"`
}

type AssistExpTier struct {
	MinLevel   int     `json:"minLevel"`
	MaxLevel   int     `json:"maxLevel"`
	Multiplier float64 `json:"multiplier"`
}

func NewRewardConfig(config RewardConfig) (*RewardConfig, error) {
	if err := validateRewardGroup("minion", config.Minion, true); err != nil {
		return nil, err
	}
	if err := validateRewardGroup("jungle", config.Jungle, false); err != nil {
		return nil, err
	}
	if err := validateJungleScaling(config.JungleScaling); err != nil {
		return nil, err
	}
	if err := validateEpicRewards(config.Epic); err != nil {
		return nil, err
	}
	if err := validateHeroKillReward(config.HeroKill); err != nil {
		return nil, err
	}
	if err := validateStructureReward(config.Structure); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *RewardConfig) MinionExp(kind string, players int) (float64, bool) {
	if c == nil {
		return 0, false
	}
	base, ok := c.Minion.KillExp[kind]
	if !ok {
		return 0, false
	}
	return sharedExp(base, players, c.Minion.ShareMultiplier, c.Minion.ShareMinPlayers), true
}

func (c *RewardConfig) MinionGold(kind string) (int, bool) {
	if c == nil {
		return 0, false
	}
	gold, ok := c.Minion.KillGold[kind]
	return gold, ok
}

func (c *RewardConfig) JungleExp(kind string) (float64, bool) {
	if c == nil {
		return 0, false
	}
	exp, ok := c.Jungle.KillExp[kind]
	return exp, ok
}

func (c *RewardConfig) JungleGold(kind string) (int, bool) {
	if c == nil {
		return 0, false
	}
	gold, ok := c.Jungle.KillGold[kind]
	return gold, ok
}

func (c *RewardConfig) StructureTeamExp(kind string) (int, bool) {
	if c == nil {
		return 0, false
	}
	exp, ok := c.Structure.TeamExp[kind]
	return exp, ok
}

func (c *RewardConfig) StructureTeamGold(kind string) (int, bool) {
	if c == nil {
		return 0, false
	}
	gold, ok := c.Structure.TeamGold[kind]
	return gold, ok
}

func (c *RewardConfig) EpicReward(kind string) (EpicReward, bool) {
	if c == nil {
		return EpicReward{}, false
	}
	reward, ok := c.Epic[kind]
	return reward, ok
}

func (c *RewardConfig) HeroKillGold() int {
	if c == nil {
		return 0
	}
	return c.HeroKill.Gold
}

func (c *RewardConfig) HeroKillExp(targetNextLevelExp int, killerLevel int, targetLevel int) float64 {
	if c == nil {
		return 0
	}
	base := float64(targetNextLevelExp) * c.HeroKill.BaseNextLevelExpMultiplier
	levelDiff := killerLevel - targetLevel
	multiplier := 1.0
	if levelDiff > 0 {
		multiplier -= float64(levelDiff) * c.HeroKill.LevelDiffStepMultiplier
		if multiplier < c.HeroKill.HigherLevelMinMultiplier {
			multiplier = c.HeroKill.HigherLevelMinMultiplier
		}
	}
	if levelDiff < 0 {
		bonus := float64(-levelDiff) * c.HeroKill.LevelDiffStepMultiplier
		if bonus > c.HeroKill.LowerLevelMaxBonusMultiplier {
			bonus = c.HeroKill.LowerLevelMaxBonusMultiplier
		}
		multiplier += bonus
	}
	return base * multiplier
}

func (c *RewardConfig) AssistExpMultiplier(level int) (float64, bool) {
	if c == nil {
		return 0, false
	}
	for _, tier := range c.HeroKill.AssistExpTiers {
		if level >= tier.MinLevel && level <= tier.MaxLevel {
			return tier.Multiplier, true
		}
	}
	return 0, false
}

func validateRewardGroup(name string, group RewardGroup, shared bool) error {
	if len(group.KillExp) == 0 {
		return fmt.Errorf("%s kill exp is required", name)
	}
	for kind, exp := range group.KillExp {
		if kind == "" {
			return fmt.Errorf("%s kill exp kind is required", name)
		}
		if exp < 0 {
			return fmt.Errorf("%s %s kill exp must not be negative", name, kind)
		}
	}
	for kind, gold := range group.KillGold {
		if kind == "" {
			return fmt.Errorf("%s kill gold kind is required", name)
		}
		if gold < 0 {
			return fmt.Errorf("%s %s kill gold must not be negative", name, kind)
		}
	}
	if !shared {
		return nil
	}
	if group.ShareMultiplier <= 0 {
		return fmt.Errorf("%s share multiplier must be positive", name)
	}
	if group.ShareMinPlayers < 2 {
		return fmt.Errorf("%s share min players must be at least 2", name)
	}
	if group.ShareRadius <= 0 {
		return fmt.Errorf("%s share radius must be positive", name)
	}
	return nil
}

func validateStructureReward(reward StructureReward) error {
	if len(reward.TeamExp) == 0 {
		return fmt.Errorf("structure team exp is required")
	}
	for kind, exp := range reward.TeamExp {
		if kind == "" {
			return fmt.Errorf("structure team exp kind is required")
		}
		if exp < 0 {
			return fmt.Errorf("structure %s team exp must not be negative", kind)
		}
	}
	for kind, gold := range reward.TeamGold {
		if kind == "" {
			return fmt.Errorf("structure team gold kind is required")
		}
		if gold < 0 {
			return fmt.Errorf("structure %s team gold must not be negative", kind)
		}
	}
	return nil
}

func validateJungleScaling(reward JungleScalingReward) error {
	if reward.StartAverageLevel <= 0 {
		return fmt.Errorf("jungle scaling start average level must be positive")
	}
	if reward.CapAverageLevel < reward.StartAverageLevel {
		return fmt.Errorf("jungle scaling cap average level must be >= start average level")
	}
	if reward.MaxMultiplier < 1 {
		return fmt.Errorf("jungle scaling max multiplier must be >= 1")
	}
	return nil
}

func validateEpicRewards(rewards map[string]EpicReward) error {
	if len(rewards) == 0 {
		return fmt.Errorf("epic rewards are required")
	}
	required := []string{"rift_herald", "elemental_dragon", "baron_nashor"}
	for _, kind := range required {
		if _, ok := rewards[kind]; !ok {
			return fmt.Errorf("missing epic reward %s", kind)
		}
	}
	for kind, reward := range rewards {
		if kind == "" {
			return fmt.Errorf("epic reward kind is required")
		}
		if reward.ParticipantExp < 0 || reward.NonParticipantTeamPoolExp < 0 || reward.MinExp < 0 || reward.MaxExp < 0 {
			return fmt.Errorf("epic reward %s exp must not be negative", kind)
		}
		if reward.TeamGold < 0 {
			return fmt.Errorf("epic reward %s team gold must not be negative", kind)
		}
		if reward.MaxExp > 0 && reward.MaxExp < reward.MinExp {
			return fmt.Errorf("epic reward %s max exp must be >= min exp", kind)
		}
	}
	return nil
}

func validateHeroKillReward(reward HeroKillReward) error {
	if reward.Gold < 0 {
		return fmt.Errorf("hero kill gold must not be negative")
	}
	if reward.BaseNextLevelExpMultiplier <= 0 {
		return fmt.Errorf("hero kill base next level exp multiplier must be positive")
	}
	if reward.LevelDiffStepMultiplier <= 0 {
		return fmt.Errorf("hero kill level diff step multiplier must be positive")
	}
	if reward.HigherLevelMinMultiplier <= 0 || reward.HigherLevelMinMultiplier > 1 {
		return fmt.Errorf("hero kill higher level min multiplier must be in (0, 1]")
	}
	if reward.LowerLevelMaxBonusMultiplier < 0 {
		return fmt.Errorf("hero kill lower level max bonus multiplier must not be negative")
	}
	if reward.NearbyRadius <= 0 {
		return fmt.Errorf("hero kill nearby radius must be positive")
	}
	if reward.DeadGraceSeconds < 0 {
		return fmt.Errorf("hero kill dead grace seconds must not be negative")
	}
	if len(reward.AssistExpTiers) == 0 {
		return fmt.Errorf("hero kill assist exp tiers are required")
	}
	for _, tier := range reward.AssistExpTiers {
		if tier.MinLevel <= 0 || tier.MaxLevel < tier.MinLevel {
			return fmt.Errorf("hero kill assist exp tier level range is invalid")
		}
		if tier.Multiplier <= 0 {
			return fmt.Errorf("hero kill assist exp tier multiplier must be positive")
		}
	}
	return nil
}

func sharedExp(base float64, players int, multiplier float64, minPlayers int) float64 {
	if players >= minPlayers {
		return base * multiplier / float64(players)
	}
	return base
}
