package world

func (w *World) tickProjectiles(tick uint64, tickRate int) {
	for id, projectile := range w.projectiles {
		source := w.entities[projectile.SourceID]
		if w.removeProjectileIfTargetDead(id, projectile) {
			continue
		}
		if w.expireProjectileIfNeeded(id, source, projectile, tick, tickRate) {
			continue
		}
		previousPosition := w.moveProjectile(projectile, tickRate)
		if w.handleProjectileAfterMove(id, source, projectile, previousPosition, tick, tickRate) {
			continue
		}
		w.resolveProjectileTargets(id, source, projectile, previousPosition, tick, tickRate)
		w.finishProjectileIfNeeded(id, source, projectile, tick, tickRate)
	}
}

func (w *World) removeProjectileIfTargetDead(id string, projectile *Projectile) bool {
	if projectile == nil || projectile.TargetID == "" {
		return false
	}
	target := w.entities[projectile.TargetID]
	if target != nil && target.Stats.HP > 0 && !target.Death.Dead {
		return false
	}
	delete(w.projectiles, id)
	w.cleanupProjectileGroup(projectile)
	return true
}
