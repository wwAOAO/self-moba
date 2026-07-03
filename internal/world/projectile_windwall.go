package world

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

func (w *World) projectileBlockedByWindWall(projectile *Projectile, previousPosition Vector2) bool {
	if projectile == nil || projectile.SkillID == mageWSkillID {
		return false
	}
	return w.BlocksProjectile(projectile.Team, previousPosition, projectile.Position)
}
