package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ReleaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Mage.FinalSparkPending || tick < entity.Mage.FinalSparkReleaseTick {
		return
	}
	targetPoint := entity.Mage.FinalSparkTarget
	level := entity.Mage.FinalSparkLevel
	entity.Mage.FinalSparkPending = false
	entity.Mage.FinalSparkReleaseTick = 0
	entity.Mage.FinalSparkTarget = world.Vector2{}
	entity.Mage.FinalSparkLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(rID)
	state := entity.Skills[rID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{80000, 65000, 50000})), tickRate)
	entity.Skills[rID] = state
	dir := world.Vector2{X: dx, Y: dy}
	addREffect(w, entity, dir, skillRange(skill, 3400), skillMeta(skill, "beamWidth", 200), tick, tickRate)
	for _, target := range rTargets(w, entity, dir, skill) {
		damage := rDamage(w, entity, target, skill, level, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, damage, "magic", tickRate)
			w.ApplyMageIlluminationOnUltimateHit(entity, target, tick, tickRate)
			if target.Kind == world.EntityKindPlayer || target.Kind == world.EntityKindEnemyHero {
				target.Control.MageFinalSparkBy = entity.ID
				target.Control.MageFinalSparkUntil = tick + secondsToTicks(skillMeta(skill, "refundWindowSeconds", 1.75), tickRate)
				target.Control.MageFinalSparkRefund = skillList(skill, "cooldownRefund", level, []float64{0.3, 0.4, 0.5})
			}
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			w.ApplyMageIlluminationOnUltimateHit(entity, target, tick, tickRate)
		}
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func rTargets(w *world.World, entity *world.Entity, direction world.Vector2, skill config.SkillConfig) []*world.Entity {
	hits := make([]*world.Entity, 0)
	castRange := skillRange(skill, 3400)
	halfWidth := skillMeta(skill, "beamWidth", 200) / 2
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		along, perpendicular := projectPoint(entity.Position, direction, target.Position)
		if along < -target.Radius || along > castRange+target.Radius {
			return
		}
		if perpendicular <= halfWidth+target.Radius {
			hits = append(hits, target)
		}
	})
	return hits
}

func rDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	baseDamage := skillList(skill, "baseDamage", level, []float64{300, 400, 500})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.75)
	return w.MageMagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func addREffect(w *world.World, entity *world.Entity, direction world.Vector2, beamRange float64, beamWidth float64, tick uint64, tickRate int) {
	lifeTicks := secondsToTicks(0.25, tickRate)
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:        w.NextEffectID("effect:mage_r:"),
		Kind:      "mage_final_spark",
		Team:      entity.Team,
		Start:     entity.Position,
		End:       w.ClampWorldPoint(world.Vector2{X: entity.Position.X + direction.X*beamRange, Y: entity.Position.Y + direction.Y*beamRange}),
		Dir:       direction,
		Range:     beamRange,
		Width:     beamWidth,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	})
}

func ApplyFinalSparkRefund(w *world.World, target *world.Entity) {
	if target == nil || target.Control.MageFinalSparkBy == "" || target.Combat.LastHitTick > target.Control.MageFinalSparkUntil {
		return
	}
	caster := w.EntityByID(target.Control.MageFinalSparkBy)
	if caster == nil {
		return
	}
	state := caster.Skills[rID]
	remaining := int64(state.CooldownUntilTick) - int64(target.Combat.LastHitTick)
	if remaining <= 0 {
		return
	}
	refund := uint64(math.Round(float64(remaining) * clamp(target.Control.MageFinalSparkRefund, 0, 1)))
	if refund >= uint64(remaining) {
		state.CooldownUntilTick = target.Combat.LastHitTick
	} else {
		state.CooldownUntilTick -= refund
	}
	caster.Skills[rID] = state
	target.Control.MageFinalSparkBy = ""
	target.Control.MageFinalSparkUntil = 0
	target.Control.MageFinalSparkRefund = 0
}

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.FinalSparkPending {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.5), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.FinalSparkPending = true
	entity.Mage.FinalSparkReleaseTick = tick + windupTicks
	entity.Mage.FinalSparkTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Mage.FinalSparkLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.FinalSparkReleaseTick
	entity.Skills[rID] = state
}
