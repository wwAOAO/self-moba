package world

import (
	"l-battle/internal/config"
	"strconv"
)

func (w *World) SkillConfig(skillID string) config.SkillConfig {
	return w.skillConfig(skillID)
}

func (w *World) EntityByID(id string) *Entity {
	if w == nil || id == "" {
		return nil
	}
	return w.entities[id]
}

func (w *World) ForEachEntity(fn func(*Entity)) {
	if w == nil || fn == nil {
		return
	}
	for _, entity := range w.entities {
		fn(entity)
	}
}

func (w *World) ClampWorldPoint(point Vector2) Vector2 {
	if w == nil {
		return point
	}
	return Vector2{
		X: clamp(point.X, 0, w.width),
		Y: clamp(point.Y, 0, w.height),
	}
}

func (w *World) NextProjectileID(prefix string) string {
	w.nextProjectileID++
	return prefix + strconv.Itoa(w.nextProjectileID)
}

func (w *World) PutProjectile(projectile *Projectile) {
	if projectile != nil && projectile.ID != "" {
		w.projectiles[projectile.ID] = projectile
	}
}

func (w *World) NextEffectID(prefix string) string {
	w.nextEffectID++
	return prefix + strconv.Itoa(w.nextEffectID)
}

func (w *World) PutSkillEffect(effect SkillEffect) {
	if effect.ID != "" {
		w.skillEffects[effect.ID] = effect
	}
}

func (w *World) RemoveSkillEffect(id string) {
	if id != "" {
		delete(w.skillEffects, id)
	}
}

func (w *World) LockAttackAfterCast(entity *Entity, tick uint64, tickRate int) {
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) RefreshPlayerStats(entity *Entity) {
	w.recalculatePlayerStats(entity)
}

func (w *World) TickDashMovement(entity *Entity, tick uint64, tickRate int) {
	w.tickDashMovement(entity, tick, tickRate)
}

func (w *World) ApplyDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyDamage(source, target, damage, tickRate)
}

func (w *World) ApplyMagicDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyMagicDamage(source, target, damage, tickRate)
}

func (w *World) ApplyTrueDamage(source *Entity, target *Entity, rawDamage float64, tickRate int) {
	w.applyTrueDamage(source, target, rawDamage, tickRate)
}

func (w *World) ApplyAOEDamage(source *Entity, target *Entity, damage int, damageType string, tickRate int) {
	w.applyAOEDamage(source, target, damage, damageType, tickRate)
}

func (w *World) PhysicalDamageAfterResistance(source *Entity, target *Entity, rawDamage float64, tick uint64) int {
	return physicalDamageAfterResistance(source, target, rawDamage, tick)
}

func (w *World) ApplyKillReward(killer *Entity, target *Entity) {
	w.applyKillReward(killer, target)
}

func (w *World) KillPlayer(target *Entity, tick uint64, tickRate int) {
	w.killPlayer(target, tick, tickRate)
}

func (w *World) RemoveDeadUnit(target *Entity) {
	w.removeDeadUnit(target)
}

func (w *World) StartTankRDash(entity *Entity, targetPoint Vector2, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	w.startTankRDash(entity, targetPoint, state, skill, tick, tickRate)
}

func (w *World) NextWindWallID(prefix string) string {
	w.nextWallID++
	return prefix + strconv.Itoa(w.nextWallID)
}

func (w *World) PutWindWall(wall WindWall) {
	if wall.ID != "" {
		w.windWalls[wall.ID] = wall
	}
}
