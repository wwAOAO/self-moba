package world

func (w *World) tickProjectiles(tick uint64, tickRate int) {
	for id, projectile := range w.projectiles {
		source := w.entities[projectile.SourceID]
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
