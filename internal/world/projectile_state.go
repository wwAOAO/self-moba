package world

import proj "l-battle/internal/world/projectile"

func updateProjectileSpeed(projectile *Projectile, tickRate int) {
	proj.UpdateSpeed(projectile, tickRate)
}

func updateMageWProjectileSpeed(projectile *Projectile) {
	if projectile == nil || projectile.SkillID != mageWSkillID || projectile.SpeedMin <= 0 || projectile.Range <= 0 {
		return
	}
	proj.UpdateMageWProjectileSpeed(projectile)
}

func updateFrostMageEProjectileSpeed(projectile *Projectile, tickRate int) {
	if projectile == nil || projectile.SpeedMin <= 0 || projectile.SpeedMax <= projectile.SpeedMin || projectile.Range <= 0 {
		return
	}
	progress := clamp(projectile.Traveled/projectile.Range, 0, 1)
	speed := projectile.SpeedMax - (projectile.SpeedMax-projectile.SpeedMin)*progress
	if tickRate > 0 {
		speed /= float64(tickRate)
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
