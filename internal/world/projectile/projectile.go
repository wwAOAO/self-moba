package projectile

import (
	"l-battle/internal/world/formula"
	"l-battle/internal/world/geom"
	"l-battle/internal/world/model"
)

type Projectile struct {
	ID           string
	Kind         string
	Team         model.Team
	SourceID     string
	TargetID     string
	SkillID      string
	GroupID      string
	Position     geom.Vector2
	Start        geom.Vector2
	Dir          geom.Vector2
	SpeedPerTick float64
	SpeedMin     float64
	SpeedMax     float64
	Range        float64
	Radius       float64
	DisplayRange float64
	DisplayCount int
	Traveled     float64
	Damage       int
	MagicDamage  int
	KnockupTicks uint64
	EffectRatio  float64
	EffectTicks  uint64
	Returning    bool
	CreatedAt    uint64
	ExpiresAt    uint64
	HitIDs       map[string]bool
}

func UpdateSpeed(projectile *Projectile, tickRate int) {
	if projectile == nil || projectile.SpeedMin <= 0 || projectile.SpeedMax <= projectile.SpeedMin {
		return
	}
	if projectile.Range <= 0 {
		return
	}
	progress := formula.Clamp(projectile.Traveled/projectile.Range, 0, 1)
	speed := projectile.SpeedMin + (projectile.SpeedMax-projectile.SpeedMin)*progress
	if tickRate > 0 {
		projectile.SpeedPerTick = speed / float64(tickRate)
		return
	}
	projectile.SpeedPerTick = speed
}

func UpdateMageWProjectileSpeed(projectile *Projectile) {
	if projectile == nil || projectile.SpeedMin <= 0 || projectile.Range <= 0 {
		return
	}
	halfRange := projectile.Range / 2
	if halfRange <= 0 {
		return
	}
	if projectile.Returning {
		progress := formula.Clamp((projectile.Traveled-halfRange)/halfRange, 0, 1)
		projectile.SpeedPerTick = projectile.SpeedMin * (0.35 + 0.65*progress)
		return
	}
	progress := formula.Clamp(projectile.Traveled/halfRange, 0, 1)
	projectile.SpeedPerTick = projectile.SpeedMin * (1 - 0.65*progress)
}
