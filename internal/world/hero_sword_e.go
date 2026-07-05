package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

func applySwordE(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.SweepingBladeTargetUntil == nil {
		entity.Sword.SweepingBladeTargetUntil = make(map[string]uint64)
	}
	target := w.swordETarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	if tick < entity.Sword.SweepingBladeTargetUntil[target.ID] {
		return
	}
	damage := swordEDamage(entity, target, skill, state.Level, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	}
	entity.Sword.SweepingBladeStacks++
	maxStacks := int(skillMetaRange(skill, "maxStacks", 4))
	if entity.Sword.SweepingBladeStacks > maxStacks {
		entity.Sword.SweepingBladeStacks = maxStacks
	}
	targetCooldownMS := skillMetaListByLevelMS(skill, "targetCooldownMs", state.Level, []float64{10000, 9000, 8000, 7000, 6000})
	entity.Sword.SweepingBladeTargetUntil[target.ID] = tick + cooldownTicks(targetCooldownMS, tickRate)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	dashThroughDistance := target.Radius + entity.Radius + skillMetaRange(skill, "dashThroughDistance", 34)
	dashEnd := w.ClampWorldPoint(Vector2{
		X: target.Position.X + dx*dashThroughDistance,
		Y: target.Position.Y + dy*dashThroughDistance,
	})
	entity.Intent = IntentState{}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = dashEnd
	entity.Control.DashUntilTick = tick + secondsToTicks(skillMetaRange(skill, "dashDurationSeconds", 0.35), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{500, 400, 300, 200, 100}), tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordESkillID] = state
}

func (w *World) swordETarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		if distance(entity.Position, target.Position) > skillRange(skill, 475)+target.Radius {
			return
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+skillMetaRange(skill, "targetPickPadding", 48) {
			return
		}
		if distToPoint < bestDistance {
			best = target
			bestDistance = distToPoint
		}
	})
	return best
}

func swordEDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{60, 70, 80, 90, 100})
	damageValue := baseDamage + attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 0.2) + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	damageValue *= 1 + float64(attacker.Sword.SweepingBladeStacks)*skillMetaRange(skill, "stackDamageBonus", 0.25)
	return magicDamageAfterResistance(attacker, target, damageValue, tick)
}
