package sword

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.SweepingBladeTargetUntil == nil {
		entity.Sword.SweepingBladeTargetUntil = make(map[string]uint64)
	}
	target := eTarget(w, entity, world.Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	if tick < entity.Sword.SweepingBladeTargetUntil[target.ID] {
		return
	}
	damage := w.SwordEDamage(entity, target, skill, state.Level, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != world.EntityKindDummy {
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
	maxStacks := int(skillMeta(skill, "maxStacks", 4))
	if entity.Sword.SweepingBladeStacks > maxStacks {
		entity.Sword.SweepingBladeStacks = maxStacks
	}
	targetCooldownMS := skillMetaListMS(skill, "targetCooldownMs", state.Level, []float64{10000, 9000, 8000, 7000, 6000})
	entity.Sword.SweepingBladeTargetUntil[target.ID] = tick + cooldownTicks(targetCooldownMS, tickRate)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	dashThroughDistance := target.Radius + entity.Radius + skillMeta(skill, "dashThroughDistance", 34)
	dashEnd := w.ClampWorldPoint(world.Vector2{
		X: target.Position.X + dx*dashThroughDistance,
		Y: target.Position.Y + dy*dashThroughDistance,
	})
	entity.Intent = world.IntentState{}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = dashEnd
	entity.Control.DashUntilTick = tick + secondsToTicks(skillMeta(skill, "dashDurationSeconds", 0.35), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListMS(skill, "cooldownMs", state.Level, []float64{500, 400, 300, 200, 100}), tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[eID] = state
}

func eTarget(w *world.World, entity *world.Entity, targetPoint world.Vector2, skill config.SkillConfig) *world.Entity {
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		if distance(entity.Position, target.Position) > skillRange(skill, 475)+target.Radius {
			return
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+skillMeta(skill, "targetPickPadding", 48) {
			return
		}
		if distToPoint < bestDistance {
			best = target
			bestDistance = distToPoint
		}
	})
	return best
}
