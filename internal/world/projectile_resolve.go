package world

func (w *World) resolveProjectileTargets(id string, source *Entity, projectile *Projectile, previousPosition Vector2, tick uint64, tickRate int) {
	removeProjectile := false
	for _, target := range w.entities {
		canHit, shouldRemove := w.projectileCanHitTarget(id, source, projectile, previousPosition, target, tick)
		if shouldRemove {
			removeProjectile = true
			break
		}
		if !canHit {
			continue
		}
		damage := w.projectileDamage(source, target, projectile, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind != EntityKindDummy {
			removeProjectile = w.resolveProjectileUnitHit(id, source, target, projectile, damage, tick, tickRate)
		} else {
			removeProjectile = w.resolveProjectileDummyHit(id, source, target, projectile, damage, tick, tickRate)
		}
		if removeProjectile {
			break
		}
	}
}

func (w *World) projectileCanHitTarget(id string, source *Entity, projectile *Projectile, previousPosition Vector2, target *Entity, tick uint64) (bool, bool) {
	if projectile.SkillID == mageWSkillID {
		if projectile.HitIDs[target.ID] || !canShieldTarget(source, target) || !projectileIntersectsTarget(projectile, previousPosition, target) {
			return false, false
		}
		projectile.HitIDs[target.ID] = true
		w.addMageShieldLayer(target, mageWShieldValue(source, w.skillConfig(projectile.SkillID), projectile.Damage), tick+projectile.EffectTicks)
		return false, false
	}
	if projectile.SkillID == archerRSkillID && target.Kind != EntityKindPlayer && target.Kind != EntityKindEnemyHero {
		return false, false
	}
	if (projectile.SkillID == tankQSkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot") && target.ID != projectile.TargetID {
		return false, false
	}
	if projectile.HitIDs[target.ID] || !canAttackTarget(source, target) {
		return false, false
	}
	if !projectileIntersectsTarget(projectile, previousPosition, target) {
		return false, false
	}
	if w.projectileGroupHit(projectile, target.ID) {
		if projectile.SkillID == archerWSkillID {
			delete(w.projectiles, id)
			return false, true
		}
		return false, false
	}
	projectile.HitIDs[target.ID] = true
	w.markProjectileGroupHit(projectile, target.ID)
	return true, false
}

func (w *World) projectileDamage(source *Entity, target *Entity, projectile *Projectile, tick uint64) int {
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
		damage = w.tankQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == mageQSkillID && source != nil {
		hitNumber := len(projectile.HitIDs)
		multiplier := 1.0
		if hitNumber >= 2 {
			multiplier = skillMetaRange(w.skillConfig(projectile.SkillID), "secondHitDamageMultiplier", 0.5)
		}
		damage = w.mageQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, multiplier, tick)
	}
	return damage
}

func (w *World) resolveProjectileUnitHit(id string, source *Entity, target *Entity, projectile *Projectile, damage int, tick uint64, tickRate int) bool {
	wasAlive := target.Stats.HP > 0
	removeProjectile := false
	if projectile.SkillID == tankQSkillID {
		w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
		delete(w.projectiles, id)
		removeProjectile = true
	} else if projectile.SkillID == mageQSkillID {
		w.applyMagicDamage(source, target, damage, tickRate)
		target.Control.RootedUntilTick = tick + controlTicksAfterTenacity(target, projectile.EffectTicks, tick)
		w.onHeroSkillHit(source, target, tick, tickRate)
		if len(projectile.HitIDs) >= int(skillMetaRange(w.skillConfig(projectile.SkillID), "maxHits", 2)) {
			delete(w.projectiles, id)
			removeProjectile = true
		}
	} else if projectile.SkillID == archerRSkillID {
		w.applyMagicDamage(source, target, damage, tickRate)
		target.Control.StunnedUntilTick = tick + controlTicksAfterTenacity(target, archerRStunTicks(projectile, w.skillConfig(projectile.SkillID), tickRate), tick)
		applyArcherRSplash(w, source, target, projectile, w.skillConfig(projectile.SkillID), tick, tickRate)
		delete(w.projectiles, id)
		removeProjectile = true
	} else {
		w.applyGenericProjectileDamage(source, target, projectile, damage, tick, tickRate)
		if projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot" || projectile.SkillID == archerWSkillID {
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
	return removeProjectile
}

func (w *World) applyGenericProjectileDamage(source *Entity, target *Entity, projectile *Projectile, damage int, tick uint64, tickRate int) {
	if projectile.Kind == "fountain_shot" {
		w.applyFountainShotDamage(source, target, projectile, tickRate)
	} else if projectile.Kind == "basic_arrow" {
		w.applyBasicAttackDamage(source, target, damage, tickRate)
	} else if projectile.SkillID == archerWSkillID || projectile.SkillID == swordQSkillID {
		w.applyAOEDamage(source, target, damage, "physical", tickRate)
	} else {
		w.applyDamage(source, target, damage, tickRate)
	}
	if projectile.Kind == "basic_arrow" {
		w.onHeroBasicHit(source, target, tick, tickRate)
	}
}

func (w *World) resolveProjectileDummyHit(id string, source *Entity, target *Entity, projectile *Projectile, damage int, tick uint64, tickRate int) bool {
	target.Combat.LastDamage = damage
	target.Combat.LastDamageType = projectileDamageType(projectile.SkillID)
	if projectile.Kind == "basic_arrow" {
		w.onHeroBasicHit(source, target, tick, tickRate)
	}
	if projectile.SkillID == tankQSkillID {
		w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
		delete(w.projectiles, id)
		return true
	}
	if projectile.SkillID == mageQSkillID {
		target.Control.RootedUntilTick = tick + projectile.EffectTicks
		w.onHeroSkillHit(source, target, tick, tickRate)
		if len(projectile.HitIDs) >= int(skillMetaRange(w.skillConfig(projectile.SkillID), "maxHits", 2)) {
			delete(w.projectiles, id)
			return true
		}
		return false
	}
	if projectile.SkillID == archerRSkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot" || projectile.SkillID == archerWSkillID {
		delete(w.projectiles, id)
		return true
	}
	return false
}
