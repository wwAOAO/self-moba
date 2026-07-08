package world

func (w *World) applyFountainShotDamage(source *Entity, target *Entity, projectile *Projectile, tickRate int) {
	if source == nil || target == nil || projectile == nil {
		return
	}
	maxHP := target.Stats.MaxHP
	trueDamage := trueDamageAfterReduction(target, fountainShotTrueBase+maxHP*fountainShotTrueRate, target.Combat.LastHitTick)
	magicDamage := magicDamageAfterResistance(source, target, fountainShotMagicBase+maxHP*fountainShotMagicRate, target.Combat.LastHitTick)
	physicalDamage := physicalDamageAfterResistance(source, target, fountainShotPhysBase+maxHP*fountainShotPhysRate, target.Combat.LastHitTick)
	w.applyResolvedDamage(source, target, trueDamage, "true", sustainSingleTargetSkill, tickRate)
	w.applyResolvedDamage(source, target, magicDamage, "magic", sustainSingleTargetSkill, tickRate)
	w.applyResolvedDamage(source, target, physicalDamage, "physical", sustainSingleTargetSkill, tickRate)
}

func (w *World) resolveTankQProjectileHit(source *Entity, projectile *Projectile, tick uint64, tickRate int) {
	if source == nil || projectile == nil {
		return
	}
	target := w.entities[projectile.TargetID]
	if !canAttackTarget(source, target) {
		return
	}
	damage := w.tankQDamage(source, target, w.skillConfig(tankQSkillID), projectile.Damage, tick)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	if target.Kind == EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
		applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
		return
	}
	wasAlive := target.Stats.HP > 0
	w.applyMagicDamage(source, target, damage, tickRate)
	applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
	if wasAlive && target.Stats.HP == 0 {
		w.applyKillReward(source, target)
		w.killPlayer(target, tick, tickRate)
		w.removeDeadUnit(target)
	}
}

func (w *World) finishMageEProjectile(source *Entity, projectile *Projectile, tick uint64, tickRate int) {
	if source == nil || projectile == nil {
		return
	}
	center := Vector2{
		X: clamp(projectile.Start.X+projectile.Dir.X*projectile.Range, 0, w.width),
		Y: clamp(projectile.Start.Y+projectile.Dir.Y*projectile.Range, 0, w.height),
	}
	w.activateMageEZone(source, center, projectile.Damage, w.skillConfig(mageESkillID), tick, tickRate)
}

func (w *World) resolveFrostQShatter(id string, source *Entity, projectile *Projectile, tick uint64, tickRate int) {
	if source != nil && projectile != nil {
		if target := w.frostQShatterTarget(source, projectile); target != nil {
			damage := w.frostQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
			target.Combat.LastHitTick = tick
			target.Combat.DamageEvents = nil
			wasAlive := target.Stats.HP > 0
			w.frostQHit(source, target, projectile, damage, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(source, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		}
	}
	delete(w.projectiles, id)
	w.cleanupProjectileGroup(projectile)
}

func (w *World) frostQShatterTarget(source *Entity, projectile *Projectile) *Entity {
	if projectile == nil {
		return nil
	}
	skill := w.skillConfig(frostmageQSkillID)
	radius := skillMetaRange(skill, "shatterSearchRadius", 100)
	center := Vector2{
		X: projectile.Position.X + projectile.Dir.X*radius,
		Y: projectile.Position.Y + projectile.Dir.Y*radius,
	}
	var best *Entity
	bestDistance := 0.0
	for _, target := range w.entities {
		if !canAttackTarget(source, target) {
			continue
		}
		dist := distance(center, target.Position)
		if dist > radius+target.Radius {
			continue
		}
		if best == nil || dist < bestDistance {
			best = target
			bestDistance = dist
		}
	}
	return best
}
