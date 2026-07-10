package world

import "math"

func canRobotQTarget(target *Entity) bool {
	return target != nil && (IsHeroUnit(target) || isMinion(target) || isMonster(target))
}

func (w *World) resolveRobotQTarget(id string, source *Entity, projectile *Projectile, previousPosition Vector2, tick uint64, tickRate int) {
	var hit *Entity
	bestAlong := math.MaxFloat64
	for _, target := range w.entities {
		if projectile.HitIDs[target.ID] || !canRobotQTarget(target) || !canAttackTarget(source, target) || !projectileIntersectsTarget(projectile, previousPosition, target) {
			continue
		}
		along, _ := projectPoint(previousPosition, projectile.Dir, target.Position)
		if along < bestAlong {
			bestAlong = along
			hit = target
		}
	}
	if hit == nil {
		return
	}
	projectile.HitIDs[hit.ID] = true
	hit.Combat.LastHitTick = tick
	hit.Combat.DamageEvents = nil
	wasAlive := hit.Stats.HP > 0
	damage := w.robotQDamage(source, hit, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
	w.applyMagicDamage(source, hit, damage, tickRate)
	if wasAlive && hit.Stats.HP > 0 {
		w.pullRobotQTarget(source, hit, projectile, tick, tickRate)
	}
	if wasAlive && hit.Stats.HP == 0 {
		w.applyKillReward(source, hit)
		w.killPlayer(hit, tick, tickRate)
		w.removeDeadUnit(hit)
	}
	delete(w.projectiles, id)
	w.cleanupProjectileGroup(projectile)
}

func (w *World) pullRobotQTarget(source *Entity, target *Entity, projectile *Projectile, tick uint64, tickRate int) {
	if source == nil || target == nil || projectile == nil {
		return
	}
	dx, dy := normalize(source.Position.X-target.Position.X, source.Position.Y-target.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.skillConfig(robotQSkillID)
	end := w.ClampWorldPoint(Vector2{
		X: source.Position.X - dx*(source.Radius+target.Radius+skillMetaRange(skill, "frontPadding", 1)),
		Y: source.Position.Y - dy*(source.Radius+target.Radius+skillMetaRange(skill, "frontPadding", 1)),
	})
	speed := skillMetaRange(skill, "pullSpeed", 1800)
	ticks := uint64(1)
	if speed > 0 && tickRate > 0 {
		ticks = uint64(math.Ceil(distance(target.Position, end) / speed * float64(tickRate)))
		if ticks < 1 {
			ticks = 1
		}
	}
	w.PutSkillEffect(SkillEffect{
		ID:           w.NextEffectID("effect:robot_q_pull:"),
		Kind:         "robot_q_pull",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		Start:        source.Position,
		End:          target.Position,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + ticks,
	})
	w.startForcedDisplacement(target, end, ticks, tick, tickRate)
}

func (w *World) startForcedDisplacement(target *Entity, end Vector2, ticks uint64, tick uint64, tickRate int) {
	if target == nil {
		return
	}
	target.Intent = IntentState{}
	target.Control.DashStartTick = tick
	target.Control.DashStart = target.Position
	target.Control.DashEnd = end
	target.Control.DashUntilTick = tick + ticks
	w.ApplyStun(target, tick+ticks, tick, tickRate)
	target.Control.ActionLockedUntilTick = tick + ticks
}
