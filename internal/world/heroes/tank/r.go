package tank

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"math"
)

func StartRDash(w *world.World, entity *world.Entity, targetPoint world.Vector2, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	start := entity.Position
	dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
	if dx == 0 && dy == 0 {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	rRange := skillRange(skill, 1000)
	if distance(start, targetPoint) > rRange {
		targetPoint = w.ClampWorldPoint(world.Vector2{X: start.X + dx*rRange, Y: start.Y + dy*rRange})
	}
	entity.Stats.MP -= manaCost
	entity.Intent = world.IntentState{}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = world.Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
	landingRadius := skillMeta(skill, "landingRadius", 250)
	dashSpeed := skillMeta(skill, "dashSpeed", 1600)
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
	entity.Tank.UnstoppableKnockupTicks = secondsToTicks(skillMeta(skill, "knockupSeconds", 1.5), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{130000, 105000, 80000})), tickRate)
	entity.Skills[rID] = state
}

func ReleasePreparedR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	state := entity.Skills[rID]
	if state.Level <= 0 || state.CooldownUntilTick > tick {
		CancelPreparedR(entity)
		return
	}
	state.Level = entity.Tank.UnstoppableCastLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	StartRDash(w, entity, entity.Tank.UnstoppableCastTarget, state, w.SkillConfig(rID), tick, tickRate)
}

func CancelPreparedR(entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = world.Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
}

func ResolveRImpact(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Tank.UnstoppableImpactPending || tick < entity.Tank.UnstoppableImpactTick {
		return
	}
	skill := w.SkillConfig(rID)
	level := entity.Tank.UnstoppableImpactLevel
	if level <= 0 {
		level = 1
	}
	radius := entity.Tank.UnstoppableImpactRadius
	if radius <= 0 {
		radius = skillMeta(skill, "landingRadius", 250)
	}
	knockupTicks := entity.Tank.UnstoppableKnockupTicks
	if knockupTicks == 0 {
		knockupTicks = secondsToTicks(skillMeta(skill, "knockupSeconds", 1.5), tickRate)
	}
	damage := rDamage(entity, skill, level)
	addRImpactEffect(w, entity, entity.Position, radius, tick, tickRate)
	for _, target := range w.TankTargetsInRadius(entity, entity.Position, radius) {
		target.Combat.LastHitTick = tick
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, w.TankMagicDamageAfterResistance(entity, target, damage, tick), "magic", tickRate)
			target.Control.AirborneUntilTick = tick + world.TankControlTicksAfterTenacity(target, knockupTicks, tick)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = w.TankMagicDamageAfterResistance(entity, target, damage, tick)
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

func rDamage(entity *world.Entity, skill config.SkillConfig, level int) float64 {
	return skillList(skill, "baseDamage", level, []float64{200, 300, 400}) +
		float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.8)
}

func addRImpactEffect(w *world.World, entity *world.Entity, center world.Vector2, radius float64, tick uint64, tickRate int) {
	if entity == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.35))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	id := w.NextEffectID("effect:tank_r_impact:")
	w.PutSkillEffect(world.SkillEffect{
		ID:        id,
		Kind:      "tank_r_impact",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	})
}
