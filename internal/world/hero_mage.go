package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
	"strconv"
)

func (w *World) applyMageQ(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Mage.LightBindingPending {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMetaRange(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Mage.LightBindingPending = true
	entity.Mage.LightBindingReleaseTick = tick + windupTicks
	entity.Mage.LightBindingTarget = Vector2{
		X: clamp(cast.TargetX, 0, w.width),
		Y: clamp(cast.TargetY, 0, w.height),
	}
	entity.Mage.LightBindingLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Mage.LightBindingReleaseTick
	entity.Skills[mageQSkillID] = state
}

func (w *World) releaseMageQ(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != mageHeroID || !entity.Mage.LightBindingPending || tick < entity.Mage.LightBindingReleaseTick {
		return
	}
	targetPoint := entity.Mage.LightBindingTarget
	level := entity.Mage.LightBindingLevel
	entity.Mage.LightBindingPending = false
	entity.Mage.LightBindingReleaseTick = 0
	entity.Mage.LightBindingTarget = Vector2{}
	entity.Mage.LightBindingLevel = 0
	if level <= 0 {
		level = 1
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.skillConfig(mageQSkillID)
	state := entity.Skills[mageQSkillID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", level, []float64{15000, 14000, 13000, 12000, 11000}), tickRate)
	entity.Skills[mageQSkillID] = state
	qRange := skillRange(skill, 1175)
	speedPerSecond := skillMetaRange(skill, "projectileSpeed", 1400)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:mage_q:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "mage_light_binding",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      mageQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMetaRange(skill, "projectileRadius", 45),
		Damage:       level,
		EffectTicks:  secondsToTicks(skillMetaRange(skill, "rootSeconds", 2), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
}

func (w *World) applyMageIlluminationOnSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !canApplyMageIllumination(source, target) {
		return
	}
	skill := w.heroPassiveSkill(source)
	target.Control.MageIlluminationBy = source.ID
	target.Control.MageIlluminationUntil = tick + secondsToTicks(skillMetaRange(skill, "debuffSeconds", 6), tickRate)
}

func (w *World) applyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if !canApplyMageIllumination(source, target) {
		return
	}
	w.detonateMageIllumination(source, target, tick, tickRate)
	w.applyMageIlluminationOnSkillHit(source, target, tick, tickRate)
}

func (w *World) triggerMageIlluminationOnBasicAttack(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != mageHeroID {
		return
	}
	w.detonateMageIllumination(source, target, tick, tickRate)
}

func canApplyMageIllumination(source *Entity, target *Entity) bool {
	if source == nil || target == nil || source.HeroID != mageHeroID {
		return false
	}
	if source.ID == target.ID || target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
		return false
	}
	return target.Team != source.Team || target.Team == TeamNeutral
}

func (w *World) detonateMageIllumination(source *Entity, target *Entity, tick uint64, tickRate int) bool {
	if !mageIlluminationActive(source, target, tick) {
		return false
	}
	target.Control.MageIlluminationBy = ""
	target.Control.MageIlluminationUntil = 0
	damage := w.mageIlluminationDamage(source, target, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyMagicDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(source, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	}
	return true
}

func mageIlluminationActive(source *Entity, target *Entity, tick uint64) bool {
	return source != nil &&
		target != nil &&
		target.Control.MageIlluminationBy == source.ID &&
		target.Control.MageIlluminationUntil > tick
}

func (w *World) mageIlluminationDamage(source *Entity, target *Entity, tick uint64) int {
	skill := w.heroPassiveSkill(source)
	baseDamage := skillMetaCurveByLevel(skill, "detonateDamage", "detonateDamageLevels", source.Level, 20)
	rawDamage := baseDamage + float64(source.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.3)
	return magicDamageAfterResistance(source, target, rawDamage, tick)
}

func mageQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{50, 100, 150, 200, 250})
	rawDamage := (baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.7)) * multiplier
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}
