package world

import (
	"l-battle/internal/config"
	"math"
)

func (w *World) releaseMageR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.FinalSparkPending || tick < entity.Mage.FinalSparkReleaseTick {
		return
	}
	targetPoint := entity.Mage.FinalSparkTarget
	level := entity.Mage.FinalSparkLevel
	entity.Mage.FinalSparkPending = false
	entity.Mage.FinalSparkReleaseTick = 0
	entity.Mage.FinalSparkTarget = Vector2{}
	entity.Mage.FinalSparkLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(mageRSkillID)
	state := entity.Skills[mageRSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{80000, 65000, 50000}), tickRate)
	entity.Skills[mageRSkillID] = state
	w.addMageREffect(entity, Vector2{X: dx, Y: dy}, skillRange(skill, 3400), skillMetaRange(skill, "beamWidth", 200), tick, tickRate)
	for _, target := range w.mageRTargets(entity, Vector2{X: dx, Y: dy}, skill, tick) {
		damage := mageRDamage(entity, target, skill, level, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, damage, "magic", tickRate)
			w.applyMageIlluminationOnUltimateHit(entity, target, tick, tickRate)
			if target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero {
				target.Control.MageFinalSparkBy = entity.ID
				target.Control.MageFinalSparkUntil = tick + secondsToTicks(skillMetaRange(skill, "refundWindowSeconds", 1.75), tickRate)
				target.Control.MageFinalSparkRefund = skillMetaListByLevel(skill, "cooldownRefund", level, []float64{0.3, 0.4, 0.5})
			}
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			w.applyMageIlluminationOnUltimateHit(entity, target, tick, tickRate)
		}
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) mageRTargets(entity *Entity, direction Vector2, skill config.SkillConfig, tick uint64) []*Entity {
	hits := make([]*Entity, 0)
	castRange := skillRange(skill, 3400)
	halfWidth := skillMetaRange(skill, "beamWidth", 200) / 2
	w.ForEachEntity(func(target *Entity) {
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

func mageRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{300, 400, 500})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.75)
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func (w *World) addMageREffect(entity *Entity, direction Vector2, beamRange float64, beamWidth float64, tick uint64, tickRate int) {
	id := w.NextEffectID("effect:mage_r:")
	lifeTicks := secondsToTicks(0.25, tickRate)
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.PutSkillEffect(SkillEffect{
		ID:        id,
		Kind:      "mage_final_spark",
		Team:      entity.Team,
		Start:     entity.Position,
		End:       w.ClampWorldPoint(Vector2{X: entity.Position.X + direction.X*beamRange, Y: entity.Position.Y + direction.Y*beamRange}),
		Dir:       direction,
		Range:     beamRange,
		Width:     beamWidth,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	})
}

func applyMageFinalSparkRefund(w *World, target *Entity) {
	if target == nil || target.Control.MageFinalSparkBy == "" || target.Combat.LastHitTick > target.Control.MageFinalSparkUntil {
		return
	}
	caster := w.EntityByID(target.Control.MageFinalSparkBy)
	if caster == nil {
		return
	}
	state := caster.Skills[mageRSkillID]
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
	caster.Skills[mageRSkillID] = state
	target.Control.MageFinalSparkBy = ""
	target.Control.MageFinalSparkUntil = 0
	target.Control.MageFinalSparkRefund = 0
}
