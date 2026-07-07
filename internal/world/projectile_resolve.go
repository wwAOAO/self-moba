package world

func (w *World) resolveProjectileTargets(id string, source *Entity, projectile *Projectile, previousPosition Vector2, tick uint64, tickRate int) {
	if projectile.SkillID == robotQSkillID {
		w.resolveRobotQTarget(id, source, projectile, previousPosition, tick, tickRate)
		return
	}
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
		damage := w.projectileDamage(source, target, projectile, tick, tickRate)
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
	if projectile.SkillID == explorerWSkillID && !canAttachExplorerW(target) {
		return false, false
	}
	if (projectile.SkillID == tankQSkillID || projectile.SkillID == gunnerQSkillID || projectile.SkillID == explorerESkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot") && target.ID != projectile.TargetID {
		return false, false
	}
	if projectile.HitIDs[target.ID] || !canAttackTarget(source, target) {
		return false, false
	}
	if !projectileIntersectsTarget(projectile, previousPosition, target) {
		return false, false
	}
	if projectile.SkillID != ninjaQSkillID && w.projectileGroupHit(projectile, target.ID) {
		if projectile.SkillID == archerWSkillID {
			delete(w.projectiles, id)
			return false, true
		}
		return false, false
	}
	projectile.HitIDs[target.ID] = true
	if projectile.SkillID != ninjaQSkillID {
		w.markProjectileGroupHit(projectile, target.ID)
	}
	return true, false
}

func (w *World) projectileDamage(source *Entity, target *Entity, projectile *Projectile, tick uint64, tickRate int) int {
	damage := projectile.Damage
	if projectile.Kind == "basic_arrow" && source != nil {
		damage = w.attackDamage(source, target, tick, tickRate)
	} else if projectile.SkillID == archerWSkillID && source != nil {
		damage = archerWDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == archerRSkillID && source != nil {
		damage = archerRDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick, 1)
	} else if projectile.SkillID == swordQSkillID && source != nil {
		damage = w.swordQDamage(source, target, w.skillConfig(projectile.SkillID), tick)
	} else if projectile.SkillID == tankQSkillID && source != nil {
		damage = w.tankQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == gunnerQSkillID && source != nil {
		crit := projectile.Returning && (projectile.EffectRatio > 0 || w.attackCrits(source, target, tick))
		damage = w.gunnerQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, crit, tick)
	} else if projectile.SkillID == gunnerRSkillID && source != nil {
		damage = w.gunnerRDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == explorerQSkillID && source != nil {
		damage = w.explorerQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == explorerESkillID && source != nil {
		damage = w.explorerEDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == explorerRSkillID && source != nil {
		damage = w.explorerRDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	} else if projectile.SkillID == mageQSkillID && source != nil {
		hitNumber := len(projectile.HitIDs)
		multiplier := 1.0
		if hitNumber >= 2 {
			multiplier = skillMetaRange(w.skillConfig(projectile.SkillID), "secondHitDamageMultiplier", 0.5)
		}
		damage = w.mageQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, multiplier, tick)
	} else if projectile.SkillID == ninjaQSkillID && source != nil {
		damage = w.ninjaQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, len(projectile.HitIDs), tick)
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
	} else if projectile.SkillID == gunnerQSkillID {
		w.applyDamage(source, target, damage, tickRate)
		if !projectile.Returning {
			w.fireGunnerQBounce(source, target, projectile, wasAlive && target.Stats.HP == 0, tick, tickRate)
		}
		delete(w.projectiles, id)
		removeProjectile = true
	} else if projectile.SkillID == gunnerRSkillID {
		w.applyAOEDamage(source, target, damage, "physical", tickRate)
	} else if projectile.SkillID == explorerQSkillID {
		w.applyBasicAttackDamage(source, target, damage, tickRate)
		w.explorerQHit(source, target, w.skillConfig(projectile.SkillID), tick, tickRate)
		w.onHeroSkillHit(source, target, tick, tickRate)
		delete(w.projectiles, id)
		removeProjectile = true
	} else if projectile.SkillID == explorerWSkillID {
		w.onHeroSkillHit(source, target, tick, tickRate)
		w.explorerWAttach(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick, tickRate)
		delete(w.projectiles, id)
		removeProjectile = true
	} else if projectile.SkillID == explorerESkillID {
		w.applyMagicDamage(source, target, damage, tickRate)
		w.explorerEHit(source, target, w.skillConfig(projectile.SkillID), tick, tickRate)
		w.onHeroSkillHit(source, target, tick, tickRate)
		delete(w.projectiles, id)
		removeProjectile = true
	} else if projectile.SkillID == explorerRSkillID {
		w.applyMagicDamage(source, target, damage, tickRate)
		w.explorerRHit(source, target, w.skillConfig(projectile.SkillID), tick, tickRate)
		w.onHeroSkillHit(source, target, tick, tickRate)
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
		if projectile.SkillID == ninjaQSkillID {
			w.ninjaSkillHit(source, target, projectile.SkillID, projectile.GroupID, projectile.FromShadow, tick, tickRate)
		}
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
	if projectile.SkillID == gunnerQSkillID {
		if !projectile.Returning {
			w.fireGunnerQBounce(source, target, projectile, false, tick, tickRate)
		}
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
	if projectile.SkillID == explorerQSkillID {
		w.explorerQHit(source, target, w.skillConfig(projectile.SkillID), tick, tickRate)
		w.onHeroSkillHit(source, target, tick, tickRate)
		delete(w.projectiles, id)
		return true
	}
	if projectile.SkillID == explorerWSkillID {
		w.onHeroSkillHit(source, target, tick, tickRate)
		w.explorerWAttach(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick, tickRate)
		delete(w.projectiles, id)
		return true
	}
	if projectile.SkillID == explorerESkillID {
		w.explorerEHit(source, target, w.skillConfig(projectile.SkillID), tick, tickRate)
		w.onHeroSkillHit(source, target, tick, tickRate)
		delete(w.projectiles, id)
		return true
	}
	if projectile.SkillID == explorerRSkillID {
		w.explorerRHit(source, target, w.skillConfig(projectile.SkillID), tick, tickRate)
		w.onHeroSkillHit(source, target, tick, tickRate)
		return false
	}
	if projectile.SkillID == archerRSkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot" || projectile.SkillID == archerWSkillID {
		delete(w.projectiles, id)
		return true
	}
	return false
}

func canAttachExplorerW(target *Entity) bool {
	if target == nil {
		return false
	}
	return IsHeroUnit(target) || target.Kind == EntityKindTower || target.Kind == EntityKindBarracks || target.Kind == EntityKindCrystal || target.Kind == EntityKindBaronNashor
}
