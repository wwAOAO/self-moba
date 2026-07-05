package effect

import (
	"l-battle/internal/world/geom"
	"l-battle/internal/world/model"
)

type WindWall struct {
	ID        string
	Team      model.Team
	Center    geom.Vector2
	Dir       geom.Vector2
	Width     float64
	ExpiresAt uint64
}

type EquipmentBurn struct {
	SourceID           string
	TargetID           string
	NextTick           uint64
	ExpiresAt          uint64
	FlatDamage         float64
	BaseMaxHPRatio     float64
	APMaxHPRatioPer100 float64
}

type SkillEffect struct {
	ID           string
	Kind         string
	Team         model.Team
	SourceHeroID string
	Start        geom.Vector2
	End          geom.Vector2
	Dir          geom.Vector2
	Range        float64
	Radius       float64
	Width        float64
	Height       float64
	Count        int
	Speed        float64
	CreatedAt    uint64
	ExpiresAt    uint64
}
