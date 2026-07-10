package sword

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	target := rTarget(w, entity, world.Vector2{X: cast.TargetX, Y: cast.TargetY}, skill, tick)
	if target == nil {
		return
	}
	entity.Position = w.ClampWorldPoint(world.Vector2{X: target.Position.X - entity.Radius - target.Radius - 18, Y: target.Position.Y})
	entity.Intent = world.IntentState{}
	hits := rTargets(w, entity, target.Position, skill, tick)
	for _, hit := range hits {
		damage := w.SwordRDamage(entity, hit, skill, state.Level, tick)
		hit.Combat.LastHitTick = tick
		if hit.Kind != world.EntityKindDummy {
			wasAlive := hit.Stats.HP > 0
			w.ApplyAOEDamage(entity, hit, damage, "physical", tickRate)
			w.ExtendAirborne(hit, secondsToTicks(skillMeta(skill, "airborneExtendSeconds", 1), tickRate), tick, tickRate)
			if wasAlive && hit.Stats.HP == 0 {
				w.ApplyKillReward(entity, hit)
				w.KillPlayer(hit, tick, tickRate)
				w.RemoveDeadUnit(hit)
			}
		} else {
			hit.Combat.LastDamage = damage
			hit.Combat.LastDamageType = "physical"
		}
	}
	entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	entity.Passive.MaxShield = w.SwordShieldValue(entity)
	entity.Passive.Shield = entity.Passive.MaxShield
	qState := entity.Skills[qID]
	qState.Stacks = 0
	qState.StacksExpireTick = 0
	entity.Skills[qID] = qState
	entity.Sword.LastBreathUntilTick = tick + secondsToTicks(skillMeta(skill, "lastBreathDurationSeconds", 15), tickRate)
	entity.Control.ActionLockedUntilTick = tick + secondsToTicks(skillMeta(skill, "selfActionLockSeconds", 1), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListMS(skill, "cooldownMs", state.Level, []float64{80000, 55000, 30000}), tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[rID] = state
}

func rTarget(w *world.World, entity *world.Entity, targetPoint world.Vector2, skill config.SkillConfig, tick uint64) *world.Entity {
	var best *world.Entity
	bestDistance := math.MaxFloat64
	pickPadding := skillMeta(skill, "targetPickPadding", 80)
	w.ForEachEntity(func(target *world.Entity) {
		if !isAirborneEnemyHero(entity, target, tick) {
			return
		}
		dist := distance(targetPoint, target.Position)
		if dist > target.Radius+pickPadding {
			return
		}
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	})
	if best != nil {
		return best
	}
	castRange := skillRange(skill, 1200)
	w.ForEachEntity(func(target *world.Entity) {
		if !isAirborneEnemyHero(entity, target, tick) {
			return
		}
		dist := distance(entity.Position, target.Position)
		if dist > castRange+target.Radius {
			return
		}
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	})
	return best
}

func rTargets(w *world.World, entity *world.Entity, center world.Vector2, skill config.SkillConfig, tick uint64) []*world.Entity {
	hits := make([]*world.Entity, 0)
	w.ForEachEntity(func(target *world.Entity) {
		if !isAirborneEnemyHero(entity, target, tick) {
			return
		}
		if distance(center, target.Position) <= skillMeta(skill, "hitRadius", 450)+target.Radius {
			hits = append(hits, target)
		}
	})
	return hits
}

func isAirborneEnemyHero(attacker *world.Entity, target *world.Entity, tick uint64) bool {
	if !canAttackTarget(attacker, target) {
		return false
	}
	if target.Control.AirborneUntilTick <= tick {
		return false
	}
	return target.Kind == world.EntityKindPlayer || target.Kind == world.EntityKindEnemyHero
}
