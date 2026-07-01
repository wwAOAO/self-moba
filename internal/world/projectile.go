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
			w.cleanupProjectileGroup(projectile)
			continue
		}
		updateProjectileSpeed(projectile, tickRate)
		step := projectile.SpeedPerTick
		remaining := projectile.Range - projectile.Traveled
		if step > remaining {
			step = remaining
		}
		if projectile.SkillID == tankQSkillID || projectile.Kind == "basic_arrow" {
			updateTrackingProjectileDir(projectile, w.entities[projectile.TargetID])
		}
		previousPosition := projectile.Position
		projectile.Position.X = clamp(projectile.Position.X+projectile.Dir.X*step, 0, w.width)
		projectile.Position.Y = clamp(projectile.Position.Y+projectile.Dir.Y*step, 0, w.height)
		projectile.Traveled += step
		source := w.entities[projectile.SourceID]
		removeProjectile := false
		for _, target := range w.entities {
			if projectile.SkillID == archerRSkillID && target.Kind != EntityKindPlayer && target.Kind != EntityKindEnemyHero {
				continue
			}
			if (projectile.SkillID == tankQSkillID || projectile.Kind == "basic_arrow") && target.ID != projectile.TargetID {
				continue
			}
			if projectile.HitIDs[target.ID] || !canAttackTarget(source, target) {
				continue
			}
			if !projectileIntersectsTarget(projectile, previousPosition, target) {
				continue
			}
			if w.projectileGroupHit(projectile, target.ID) {
				if projectile.SkillID == archerWSkillID {
					delete(w.projectiles, id)
					removeProjectile = true
					break
				}
				continue
			}
			projectile.HitIDs[target.ID] = true
			w.markProjectileGroupHit(projectile, target.ID)
			damage := projectile.Damage
			if projectile.Kind == "basic_arrow" && source != nil {
				damage = w.attackDamage(source, target, tick)
			} else if projectile.SkillID == archerWSkillID && source != nil {
				damage = archerWDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
			} else if projectile.SkillID == archerRSkillID && source != nil {
				damage = archerRDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick, 1)
			} else if projectile.SkillID == swordQSkillID && source != nil {
				damage = w.swordQDamage(source, target, w.skillConfig(projectile.SkillID), tick)
			} else if projectile.SkillID == tankQSkillID && source != nil {
				damage = tankQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
			} else if projectile.SkillID == mageQSkillID && source != nil {
				hitNumber := len(projectile.HitIDs)
				multiplier := 1.0
				if hitNumber >= 2 {
					multiplier = skillMetaRange(w.skillConfig(projectile.SkillID), "secondHitDamageMultiplier", 0.5)
				}
				damage = mageQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, multiplier, tick)
			}
			target.Combat.LastHitTick = tick
			if target.Kind != EntityKindDummy {
				wasAlive := target.Stats.HP > 0
				if projectile.SkillID == tankQSkillID {
					w.applyMagicDamage(source, target, damage, tickRate)
					applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
					delete(w.projectiles, id)
					removeProjectile = true
				} else if projectile.SkillID == mageQSkillID {
					w.applyMagicDamage(source, target, damage, tickRate)
					target.Control.RootedUntilTick = tick + controlTicksAfterTenacity(target, projectile.EffectTicks, tick)
					w.applyMageIlluminationOnSkillHit(source, target, tick, tickRate)
					if len(projectile.HitIDs) >= int(skillMetaRange(w.skillConfig(projectile.SkillID), "maxHits", 2)) {
						delete(w.projectiles, id)
						removeProjectile = true
					}
				} else if projectile.SkillID == archerRSkillID {
					w.applyMagicDamage(source, target, damage, tickRate)
					target.Control.StunnedUntilTick = tick + controlTicksAfterTenacity(target, archerRStunTicks(projectile, w.skillConfig(projectile.SkillID), tickRate), tick)
					w.applyArcherRSplash(source, target, projectile, w.skillConfig(projectile.SkillID), tick, tickRate)
					delete(w.projectiles, id)
					removeProjectile = true
				} else {
					if projectile.Kind == "basic_arrow" {
						w.applyBasicAttackDamage(source, target, damage, tickRate)
					} else if projectile.SkillID == archerWSkillID || projectile.SkillID == swordQSkillID {
						w.applyAOEDamage(source, target, damage, "physical", tickRate)
					} else {
						w.applyDamage(source, target, damage, tickRate)
					}
					if projectile.Kind == "basic_arrow" && source != nil && source.HeroID == archerHeroID {
						w.applyArcherFocusOnBasicHit(source, target, tick, tickRate)
					}
					if projectile.Kind == "basic_arrow" {
						w.triggerMageIlluminationOnBasicAttack(source, target, tick, tickRate)
					}
					if projectile.Kind == "basic_arrow" || projectile.SkillID == archerWSkillID {
						delete(w.projectiles, id)
						removeProjectile = true
					}
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
				if projectile.Kind == "basic_arrow" && source != nil && source.HeroID == archerHeroID {
					w.applyArcherFocusOnBasicHit(source, target, tick, tickRate)
				}
				if projectile.SkillID == tankQSkillID {
					applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
					delete(w.projectiles, id)
					removeProjectile = true
				} else if projectile.SkillID == mageQSkillID {
					target.Control.RootedUntilTick = tick + projectile.EffectTicks
					w.applyMageIlluminationOnSkillHit(source, target, tick, tickRate)
					if len(projectile.HitIDs) >= int(skillMetaRange(w.skillConfig(projectile.SkillID), "maxHits", 2)) {
						delete(w.projectiles, id)
						removeProjectile = true
					}
				} else if projectile.SkillID == archerRSkillID {
					delete(w.projectiles, id)
					removeProjectile = true
				} else if projectile.Kind == "basic_arrow" || projectile.SkillID == archerWSkillID {
					delete(w.projectiles, id)
					removeProjectile = true
				}
			}
			if removeProjectile {
				break
			}
		}
		if projectile.Traveled >= projectile.Range {
			delete(w.projectiles, id)
			w.cleanupProjectileGroup(projectile)
		}
	}
}

func projectileIntersectsTarget(projectile *Projectile, previousPosition Vector2, target *Entity) bool {
	if projectile == nil || target == nil {
		return false
	}
	if projectile.SkillID == archerWSkillID {
		return archerVolleyArrowIntersectsTarget(projectile, previousPosition, target)
	}
	return distancePointToSegment(target.Position, previousPosition, projectile.Position) <= projectile.Radius+target.Radius
}

func archerVolleyArrowIntersectsTarget(projectile *Projectile, previousPosition Vector2, target *Entity) bool {
	length := projectileArrowLength(projectile)
	radius := projectileArrowCollisionRadius(projectile) + target.Radius
	segments := [][2]Vector2{
		arrowBodySegment(previousPosition, projectile.Dir, length),
		arrowBodySegment(projectile.Position, projectile.Dir, length),
		{arrowTail(previousPosition, projectile.Dir, length), arrowTail(projectile.Position, projectile.Dir, length)},
		{arrowHead(previousPosition, projectile.Dir, length), arrowHead(projectile.Position, projectile.Dir, length)},
	}
	for _, segment := range segments {
		if distancePointToSegment(target.Position, segment[0], segment[1]) <= radius {
			return true
		}
	}
	return false
}

func arrowBodySegment(center Vector2, direction Vector2, length float64) [2]Vector2 {
	return [2]Vector2{
		arrowTail(center, direction, length),
		arrowHead(center, direction, length),
	}
}

func arrowTail(center Vector2, direction Vector2, length float64) Vector2 {
	return Vector2{
		X: center.X - direction.X*length*0.5,
		Y: center.Y - direction.Y*length*0.5,
	}
}

func arrowHead(center Vector2, direction Vector2, length float64) Vector2 {
	return Vector2{
		X: center.X + direction.X*length*0.5,
		Y: center.Y + direction.Y*length*0.5,
	}
}

func projectileArrowLength(projectile *Projectile) float64 {
	if projectile == nil || projectile.Radius <= 0 {
		return 26
	}
	return projectile.Radius * 1.625
}

func projectileArrowCollisionRadius(projectile *Projectile) float64 {
	if projectile == nil || projectile.Radius <= 0 {
		return 16
	}
	return projectile.Radius
}

func updateProjectileSpeed(projectile *Projectile, tickRate int) {
	if projectile == nil || projectile.SpeedMin <= 0 || projectile.SpeedMax <= projectile.SpeedMin {
		return
	}
	if projectile.Range <= 0 {
		return
	}
	progress := clamp(projectile.Traveled/projectile.Range, 0, 1)
	speed := projectile.SpeedMin + (projectile.SpeedMax-projectile.SpeedMin)*progress
	if tickRate > 0 {
		projectile.SpeedPerTick = speed / float64(tickRate)
		return
	}
	projectile.SpeedPerTick = speed
}

func (w *World) projectileGroupHit(projectile *Projectile, targetID string) bool {
	if projectile == nil || projectile.GroupID == "" || targetID == "" {
		return false
	}
	return w.projectileHits[projectile.GroupID][targetID]
}

func (w *World) markProjectileGroupHit(projectile *Projectile, targetID string) {
	if projectile == nil || projectile.GroupID == "" || targetID == "" {
		return
	}
	if w.projectileHits[projectile.GroupID] == nil {
		w.projectileHits[projectile.GroupID] = make(map[string]bool)
	}
	w.projectileHits[projectile.GroupID][targetID] = true
}

func (w *World) cleanupProjectileGroup(projectile *Projectile) {
	if projectile == nil || projectile.GroupID == "" {
		return
	}
	for _, other := range w.projectiles {
		if other.GroupID == projectile.GroupID {
			return
		}
	}
	delete(w.projectileHits, projectile.GroupID)
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
		if projectile.SkillID == tankQSkillID || projectile.SkillID == archerWSkillID || projectile.SkillID == archerRSkillID {
			start = projectile.Position
		}
		if projectile.SkillID == tankQSkillID {
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
			Count:     projectile.DisplayCount,
			Speed:     projectile.SpeedPerTick,
			CreatedAt: createdAt,
			ExpiresAt: projectile.ExpiresAt,
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
	if skillID == tankQSkillID || skillID == mageQSkillID {
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
