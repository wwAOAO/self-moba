package world

func (w *World) tickProjectiles(tick uint64, tickRate int) {
	for id, projectile := range w.projectiles {
		source := w.entities[projectile.SourceID]
		if tick >= projectile.ExpiresAt || (projectile.SkillID != mageWSkillID && projectile.Traveled >= projectile.Range) {
			if projectile.SkillID == mageESkillID {
				w.finishMageEProjectile(source, projectile, tick, tickRate)
			}
			if projectile.SkillID == tankQSkillID {
				w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
			}
			delete(w.projectiles, id)
			w.cleanupProjectileGroup(projectile)
			continue
		}
		if projectile.SkillID == mageWSkillID {
			updateMageWProjectileSpeed(projectile)
		}
		if projectile.SkillID != mageWSkillID {
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
		if step > remaining {
			step = remaining
		}
		if projectile.SkillID == tankQSkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot" {
			updateTrackingProjectileDir(projectile, w.entities[projectile.TargetID])
		}
		previousPosition := projectile.Position
		projectile.Position.X = clamp(projectile.Position.X+projectile.Dir.X*step, 0, w.width)
		projectile.Position.Y = clamp(projectile.Position.Y+projectile.Dir.Y*step, 0, w.height)
		projectile.Traveled += step
		if w.projectileBlockedByWindWall(projectile, previousPosition) {
			delete(w.projectiles, id)
			w.cleanupProjectileGroup(projectile)
			continue
		}
		if projectile.SkillID == mageESkillID && projectile.Traveled >= projectile.Range {
			w.finishMageEProjectile(source, projectile, tick, tickRate)
			delete(w.projectiles, id)
			continue
		}
		if projectile.SkillID == mageWSkillID && !projectile.Returning && projectile.Traveled >= projectile.Range/2 {
			projectile.Returning = true
			projectile.Dir.X = -projectile.Dir.X
			projectile.Dir.Y = -projectile.Dir.Y
			projectile.HitIDs = make(map[string]bool)
		}
		if projectile.SkillID == mageESkillID {
			continue
		}
		removeProjectile := false
		for _, target := range w.entities {
			if projectile.SkillID == mageWSkillID {
				if projectile.HitIDs[target.ID] || !canShieldTarget(source, target) || !projectileIntersectsTarget(projectile, previousPosition, target) {
					continue
				}
				projectile.HitIDs[target.ID] = true
				w.addMageShieldLayer(target, mageWShieldValue(source, w.skillConfig(projectile.SkillID), projectile.Damage), tick+projectile.EffectTicks)
				continue
			}
			if projectile.SkillID == archerRSkillID && target.Kind != EntityKindPlayer && target.Kind != EntityKindEnemyHero {
				continue
			}
			if (projectile.SkillID == tankQSkillID || projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot") && target.ID != projectile.TargetID {
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
			target.Combat.DamageEvents = nil
			if target.Kind != EntityKindDummy {
				wasAlive := target.Stats.HP > 0
				if projectile.SkillID == tankQSkillID {
					w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
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
					if projectile.Kind == "fountain_shot" {
						w.applyFountainShotDamage(source, target, projectile, tickRate)
					} else if projectile.Kind == "basic_arrow" {
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
						w.gainBladeBasicAttackRage(source, target, tick)
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
			} else {
				target.Combat.LastDamage = damage
				target.Combat.LastDamageType = projectileDamageType(projectile.SkillID)
				if projectile.Kind == "basic_arrow" && source != nil && source.HeroID == archerHeroID {
					w.applyArcherFocusOnBasicHit(source, target, tick, tickRate)
				}
				if projectile.Kind == "basic_arrow" {
					w.gainBladeBasicAttackRage(source, target, tick)
				}
				if projectile.SkillID == tankQSkillID {
					w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
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
				} else if projectile.Kind == "basic_arrow" || projectile.Kind == "fountain_shot" || projectile.SkillID == archerWSkillID {
					delete(w.projectiles, id)
					removeProjectile = true
				}
			}
			if removeProjectile {
				break
			}
		}
		if projectile.SkillID == mageWSkillID && projectile.Returning && distance(projectile.Position, projectile.Start) <= 1 {
			if source != nil {
				w.addMageShieldLayer(source, mageWShieldValue(source, w.skillConfig(projectile.SkillID), projectile.Damage), tick+projectile.EffectTicks)
			}
			delete(w.projectiles, id)
			continue
		}
		if projectile.SkillID != mageWSkillID && projectile.Traveled >= projectile.Range {
			if projectile.SkillID == tankQSkillID {
				w.resolveTankQProjectileHit(source, projectile, tick, tickRate)
			}
			delete(w.projectiles, id)
			w.cleanupProjectileGroup(projectile)
		}
	}
}
