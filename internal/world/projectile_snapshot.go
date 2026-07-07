package world

func (w *World) WindWalls() []WindWall {
	walls := make([]WindWall, 0, len(w.windWalls))
	for _, wall := range w.windWalls {
		walls = append(walls, wall)
	}
	return walls
}

func (w *World) SkillEffects() []SkillEffect {
	effects := make([]SkillEffect, 0, len(w.projectiles)+len(w.skillEffects))
	for _, effect := range w.skillEffects {
		if effect.Kind == "berserker_q" || effect.Kind == "berserker_r" {
			if source := w.entities[effect.SourceID]; source != nil {
				effect.Start = source.Position
			}
		}
		effects = append(effects, effect)
	}
	for _, projectile := range w.projectiles {
		start := projectile.Start
		createdAt := projectile.CreatedAt
		sourceHeroID := ""
		if source := w.entities[projectile.SourceID]; source != nil {
			sourceHeroID = source.HeroID
		}
		if projectile.SkillID == tankQSkillID || projectile.SkillID == gunnerQSkillID || projectile.SkillID == gunnerRSkillID || projectile.SkillID == archerWSkillID || projectile.SkillID == archerRSkillID || projectile.SkillID == mageQSkillID || projectile.SkillID == mageWSkillID || projectile.SkillID == mageESkillID || projectile.SkillID == ninjaQSkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot" {
			start = projectile.Position
		}
		if projectile.SkillID == tankQSkillID {
			createdAt = 0
		}
		effects = append(effects, SkillEffect{
			ID:           projectile.ID,
			Kind:         projectile.Kind,
			Team:         projectile.Team,
			SourceID:     projectile.SourceID,
			SourceHeroID: sourceHeroID,
			Start:        start,
			Dir:          projectile.Dir,
			Range:        projectile.Range,
			Radius:       projectile.Radius,
			Count:        projectile.DisplayCount,
			Speed:        projectile.SpeedPerTick,
			CreatedAt:    createdAt,
			ExpiresAt:    projectile.ExpiresAt,
		})
		if projectile.DisplayRange > 0 {
			effects[len(effects)-1].Width = projectile.DisplayRange
		}
	}
	return effects
}

func updateTrackingProjectileDir(projectile *Projectile, target *Entity) {
	if projectile == nil || target == nil || target.Stats.HP <= 0 {
		return
	}
	dx, dy := normalize(target.Position.X-projectile.Position.X, target.Position.Y-projectile.Position.Y)
	if dx == 0 && dy == 0 {
		return
	}
	projectile.Dir = Vector2{X: dx, Y: dy}
}

func projectileDamageType(skillID string) string {
	if skillID == tankQSkillID || skillID == mageQSkillID || skillID == mageESkillID {
		return "magic"
	}
	return "physical"
}
