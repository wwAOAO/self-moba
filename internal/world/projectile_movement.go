package world

func (w *World) moveProjectile(projectile *Projectile, tickRate int) Vector2 {
	if projectile.SkillID == mageWSkillID {
		updateMageWProjectileSpeed(projectile)
	}
	if projectile.SkillID == frostmageESkillID {
		updateFrostMageEProjectileSpeed(projectile, tickRate)
	}
	if projectile.SkillID != mageWSkillID && projectile.SkillID != frostmageESkillID {
		updateProjectileSpeed(projectile, tickRate)
	}
	step := projectile.SpeedPerTick
	remaining := projectile.Range - projectile.Traveled
	if projectile.SkillID == mageWSkillID && !projectile.Returning {
		remaining = projectile.Range/2 - projectile.Traveled
	}
	if projectile.SkillID == mageWSkillID && projectile.Returning {
		remaining = distance(projectile.Position, projectile.Start)
	}
	if projectile.Boomerang && !projectile.Returning {
		remaining = projectile.Range/2 - projectile.Traveled
	}
	if projectile.Boomerang && projectile.Returning {
		remaining = distance(projectile.Position, projectile.Start)
	}
	if projectile.Recallable && projectile.Returning {
		if source := w.entities[projectile.SourceID]; source != nil {
			updateReturningProjectileDir(projectile, source.Position)
			remaining = distance(projectile.Position, source.Position)
		}
	}
	if isShadowAssassinBlade(projectile) && projectile.Returning {
		if source := w.entities[projectile.SourceID]; source != nil {
			updateReturningProjectileDir(projectile, source.Position)
			remaining = distance(projectile.Position, source.Position)
		}
	}
	if step > remaining {
		step = remaining
	}
	if projectile.SkillID == tankQSkillID || projectile.SkillID == gunnerQSkillID || projectile.SkillID == explorerESkillID || projectile.SkillID == fireMageRSkillID || projectile.SkillID == killerQSkillID || projectile.SkillID == killerRSkillID || projectile.Kind == "crossbowman_condemn" || isBasicAttackProjectileKind(projectile.Kind) || projectile.Kind == "fountain_shot" {
		updateTrackingProjectileDir(projectile, w.entities[projectile.TargetID])
	}
	previousPosition := projectile.Position
	projectile.Position.X = clamp(projectile.Position.X+projectile.Dir.X*step, 0, w.width)
	projectile.Position.Y = clamp(projectile.Position.Y+projectile.Dir.Y*step, 0, w.height)
	projectile.Traveled += step
	return previousPosition
}

func (w *World) handleProjectileAfterMove(id string, source *Entity, projectile *Projectile, previousPosition Vector2, tick uint64, tickRate int) bool {
	if w.projectileBlockedByWindWall(projectile, previousPosition) {
		delete(w.projectiles, id)
		w.cleanupProjectileGroup(projectile)
		return true
	}
	if projectile.SkillID == mageESkillID && projectile.Traveled >= projectile.Range {
		w.finishMageEProjectile(source, projectile, tick, tickRate)
		delete(w.projectiles, id)
		return true
	}
	if projectile.SkillID == mageWSkillID && !projectile.Returning && projectile.Traveled >= projectile.Range/2 {
		projectile.Returning = true
		projectile.Dir.X = -projectile.Dir.X
		projectile.Dir.Y = -projectile.Dir.Y
		projectile.HitIDs = make(map[string]bool)
	}
	if projectile.Boomerang && !projectile.Returning && projectile.Traveled >= projectile.Range/2 {
		projectile.Returning = true
		projectile.Dir.X = -projectile.Dir.X
		projectile.Dir.Y = -projectile.Dir.Y
		projectile.EffectTicks = tick
	}
	if isShadowAssassinBlade(projectile) && !projectile.Returning && projectileAtWorldBoundary(w, projectile.Position) {
		projectile.Returning = true
		projectile.EffectTicks = tick
	}
	if isShadowAssassinBlade(projectile) && projectile.Returning && source != nil {
		updateReturningProjectileDir(projectile, source.Position)
	}
	return projectile.SkillID == mageESkillID
}

func (w *World) expireProjectileIfNeeded(id string, source *Entity, projectile *Projectile, tick uint64, tickRate int) bool {
	if tick < projectile.ExpiresAt && (projectile.SkillID == mageWSkillID || projectile.Boomerang || projectile.Recallable || projectile.Traveled < projectile.Range) {
		return false
	}
	if projectile.SkillID == mageESkillID {
		w.finishMageEProjectile(source, projectile, tick, tickRate)
	}
	if projectile.SkillID == tankQSkillID {
		w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
	}
	delete(w.projectiles, id)
	w.cleanupProjectileGroup(projectile)
	return true
}

func (w *World) finishProjectileIfNeeded(id string, source *Entity, projectile *Projectile, tick uint64, tickRate int) {
	if projectile.SkillID == frostmageQSkillID && !projectile.Returning && projectile.EffectRatio > 0 && projectile.Traveled >= projectile.EffectRatio {
		w.resolveFrostQShatter(id, source, projectile, tick, tickRate)
		return
	}
	if projectile.SkillID == mageWSkillID && projectile.Returning && distance(projectile.Position, projectile.Start) <= 1 {
		if source != nil {
			w.addMageShieldLayer(source, mageWShieldValue(source, w.skillConfig(projectile.SkillID), projectile.Damage), tick+projectile.EffectTicks)
		}
		delete(w.projectiles, id)
		return
	}
	if isShadowAssassinBlade(projectile) && projectile.Returning && source != nil && distance(projectile.Position, source.Position) <= max(1, source.Radius) {
		delete(w.projectiles, id)
		w.cleanupProjectileGroup(projectile)
		return
	}
	if projectile.Boomerang && projectile.Returning && distance(projectile.Position, projectile.Start) <= 1 {
		delete(w.projectiles, id)
		w.cleanupProjectileGroup(projectile)
		return
	}
	if projectile.Recallable && projectile.Returning {
		if source := w.entities[projectile.SourceID]; source != nil && distance(projectile.Position, source.Position) <= 1 {
			delete(w.projectiles, id)
			w.cleanupProjectileGroup(projectile)
			return
		}
	}
	if projectile.SkillID != mageWSkillID && !projectile.Boomerang && !projectile.Recallable && projectile.Traveled >= projectile.Range {
		if projectile.SkillID == tankQSkillID {
			w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
		}
		delete(w.projectiles, id)
		w.cleanupProjectileGroup(projectile)
	}
}

func isShadowAssassinBlade(projectile *Projectile) bool {
	return projectile != nil && (projectile.SkillID == shadowAssassinWSkillID || projectile.SkillID == shadowAssassinRSkillID)
}

func projectileAtWorldBoundary(w *World, position Vector2) bool {
	return w != nil && (position.X <= 0 || position.X >= w.width || position.Y <= 0 || position.Y >= w.height)
}

func updateReturningProjectileDir(projectile *Projectile, target Vector2) {
	dx, dy := normalize(target.X-projectile.Position.X, target.Y-projectile.Position.Y)
	if dx == 0 && dy == 0 {
		return
	}
	projectile.Dir = Vector2{X: dx, Y: dy}
}
