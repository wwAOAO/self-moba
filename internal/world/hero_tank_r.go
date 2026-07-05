package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) startTankRDash(entity *Entity, targetPoint Vector2, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	start := entity.Position
	dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
	if dx == 0 && dy == 0 {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	rRange := skillRange(skill, 1000)
	if distance(start, targetPoint) > rRange {
		targetPoint = w.ClampWorldPoint(Vector2{X: start.X + dx*rRange, Y: start.Y + dy*rRange})
	}
	entity.Stats.MP -= manaCost
	entity.Intent = IntentState{}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
	landingRadius := skillMetaRange(skill, "landingRadius", 250)
	dashSpeed := skillMetaRange(skill, "dashSpeed", 1600)
	if dashSpeed <= 0 {
		dashSpeed = rRange
	}
	travelTicks := uint64(math.Ceil(distance(start, targetPoint) / dashSpeed * float64(tickRate)))
	if travelTicks < 1 {
		travelTicks = 1
	}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = start
	entity.Control.DashEnd = targetPoint
	entity.Control.DashUntilTick = tick + travelTicks
	entity.Control.ActionLockedUntilTick = entity.Control.DashUntilTick
	entity.Tank.UnstoppableImpactPending = true
	entity.Tank.UnstoppableImpactTick = entity.Control.DashUntilTick
	entity.Tank.UnstoppableImpactLevel = state.Level
	entity.Tank.UnstoppableImpactRadius = landingRadius
	entity.Tank.UnstoppableKnockupTicks = secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1.5), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{130000, 105000, 80000}), tickRate)
	entity.Skills[tankRSkillID] = state
}

func (w *World) releasePreparedTankR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	state := entity.Skills[tankRSkillID]
	if state.Level <= 0 {
		w.cancelTankRPreparedCast(entity)
		return
	}
	if state.CooldownUntilTick > tick {
		w.cancelTankRPreparedCast(entity)
		return
	}
	state.Level = entity.Tank.UnstoppableCastLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	w.startTankRDash(entity, entity.Tank.UnstoppableCastTarget, state, w.SkillConfig(tankRSkillID), tick, tickRate)
}

func (w *World) cancelTankRPreparedCast(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
}

func (w *World) resolveTankRImpact(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableImpactPending {
		return
	}
	if tick < entity.Tank.UnstoppableImpactTick {
		return
	}
	skill := w.SkillConfig(tankRSkillID)
	level := entity.Tank.UnstoppableImpactLevel
	if level <= 0 {
		level = 1
	}
	radius := entity.Tank.UnstoppableImpactRadius
	if radius <= 0 {
		radius = skillMetaRange(skill, "landingRadius", 250)
	}
	knockupTicks := entity.Tank.UnstoppableKnockupTicks
	if knockupTicks == 0 {
		knockupTicks = secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1.5), tickRate)
	}
	damage := tankRDamage(entity, skill, level)
	w.addTankRImpactEffect(entity, entity.Position, radius, tick, tickRate)
	for _, target := range w.targetsInRadius(entity, entity.Position, radius) {
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, magicDamageAfterResistance(entity, target, damage, tick), "magic", tickRate)
			target.Control.AirborneUntilTick = tick + controlTicksAfterTenacity(target, knockupTicks, tick)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			target.Control.AirborneUntilTick = tick + knockupTicks
		}
	}
	entity.Tank.UnstoppableImpactPending = false
	entity.Tank.UnstoppableImpactTick = 0
	entity.Tank.UnstoppableImpactLevel = 0
	entity.Tank.UnstoppableImpactRadius = 0
	entity.Tank.UnstoppableKnockupTicks = 0
}

func tankRDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{200, 300, 400}) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.8)
}

func (w *World) addTankRImpactEffect(entity *Entity, center Vector2, radius float64, tick uint64, tickRate int) {
	if entity == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.35))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	id := w.NextEffectID("effect:tank_r_impact:")
	w.PutSkillEffect(SkillEffect{
		ID:        id,
		Kind:      "tank_r_impact",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	})
}
