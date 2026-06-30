package world

func (w *World) expireWindWalls(tick uint64) {
	for id, wall := range w.windWalls {
		if tick >= wall.ExpiresAt {
			delete(w.windWalls, id)
		}
	}
}

func (w *World) expireSkillEffects(tick uint64) {
	for id, effect := range w.skillEffects {
		if tick >= effect.ExpiresAt {
			delete(w.skillEffects, id)
		}
	}
}

func (w *World) tickProjectiles(tick uint64, tickRate int) {
	for id, projectile := range w.projectiles {
		if tick >= projectile.ExpiresAt || projectile.Traveled >= projectile.Range {
			delete(w.projectiles, id)
			continue
		}
		step := projectile.SpeedPerTick
		remaining := projectile.Range - projectile.Traveled
		if step > remaining {
			step = remaining
		}
		if projectile.SkillID == tankQSkillID {
			updateTrackingProjectileDir(projectile, w.entities[projectile.TargetID])
		}
		projectile.Position.X = clamp(projectile.Position.X+projectile.Dir.X*step, 0, w.width)
		projectile.Position.Y = clamp(projectile.Position.Y+projectile.Dir.Y*step, 0, w.height)
		projectile.Traveled += step
		source := w.entities[projectile.SourceID]
		for _, target := range w.entities {
			if projectile.SkillID == tankQSkillID && target.ID != projectile.TargetID {
				continue
			}
			if projectile.HitIDs[target.ID] || !canAttackTarget(source, target) {
				continue
			}
			if distance(projectile.Position, target.Position) > projectile.Radius+target.Radius {
				continue
			}
			projectile.HitIDs[target.ID] = true
			damage := projectile.Damage
			if projectile.SkillID == swordQSkillID && source != nil {
				damage = w.swordQDamage(source, target, w.skillConfig(projectile.SkillID), tick)
			} else if projectile.SkillID == tankQSkillID && source != nil {
				damage = tankQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
			}
			target.Combat.LastHitTick = tick
			if target.Kind != EntityKindDummy {
				wasAlive := target.Stats.HP > 0
				if projectile.SkillID == tankQSkillID {
					w.applyMagicDamage(source, target, damage, tickRate)
					applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
					delete(w.projectiles, id)
				} else {
					w.applyDamage(source, target, damage, tickRate)
				}
				if projectile.KnockupTicks > 0 {
					target.Control.AirborneUntilTick = tick + projectile.KnockupTicks
				}
				if wasAlive && target.Stats.HP == 0 {
					w.applyKillReward(source, target)
					w.killPlayer(target, tick, tickRate)
					w.removeDeadUnit(target)
				}
			} else {
				target.Combat.LastDamage = damage
				target.Combat.LastDamageType = projectileDamageType(projectile.SkillID)
				if projectile.SkillID == tankQSkillID {
					applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
					delete(w.projectiles, id)
				}
			}
		}
		if projectile.Traveled >= projectile.Range {
			delete(w.projectiles, id)
		}
	}
}

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
		effects = append(effects, effect)
	}
	for _, projectile := range w.projectiles {
		start := projectile.Start
		createdAt := projectile.CreatedAt
		if projectile.SkillID == tankQSkillID {
			start = projectile.Position
			createdAt = 0
		}
		effects = append(effects, SkillEffect{
			ID:        projectile.ID,
			Kind:      projectile.Kind,
			Team:      projectile.Team,
			Start:     start,
			Dir:       projectile.Dir,
			Range:     projectile.Range,
			Radius:    projectile.Radius,
			Speed:     projectile.SpeedPerTick,
			CreatedAt: createdAt,
			ExpiresAt: projectile.ExpiresAt,
		})
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
	if skillID == tankQSkillID {
		return "magic"
	}
	return "physical"
}

func (w *World) BlocksProjectile(team Team, from Vector2, to Vector2) bool {
	for _, wall := range w.windWalls {
		if wall.Team == team {
			continue
		}
		if segmentsIntersect(from, to, windWallStart(wall), windWallEnd(wall)) {
			return true
		}
	}
	return false
}
