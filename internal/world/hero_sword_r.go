package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

func applySwordR(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	target := w.swordRTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill, tick)
	if target == nil {
		return
	}
	entity.Position = w.ClampWorldPoint(Vector2{X: target.Position.X - entity.Radius - target.Radius - 18, Y: target.Position.Y})
	entity.Intent = IntentState{}
	hits := w.swordRTargets(entity, target.Position, skill, tick)
	for _, hit := range hits {
		damage := swordRDamage(entity, hit, skill, state.Level, tick)
		hit.Combat.LastHitTick = tick
		if hit.Kind != EntityKindDummy {
			wasAlive := hit.Stats.HP > 0
			w.ApplyAOEDamage(entity, hit, damage, "physical", tickRate)
			hit.Control.AirborneUntilTick += secondsToTicks(skillMetaRange(skill, "airborneExtendSeconds", 1), tickRate)
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
	entity.Passive.MaxShield = w.swordShieldValue(entity)
	entity.Passive.Shield = entity.Passive.MaxShield
	qState := entity.Skills[swordQSkillID]
	qState.Stacks = 0
	qState.StacksExpireTick = 0
	entity.Skills[swordQSkillID] = qState
	entity.Sword.LastBreathUntilTick = tick + secondsToTicks(skillMetaRange(skill, "lastBreathDurationSeconds", 15), tickRate)
	entity.Control.ActionLockedUntilTick = tick + secondsToTicks(skillMetaRange(skill, "selfActionLockSeconds", 1), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{80000, 55000, 30000}), tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordRSkillID] = state
}

func (w *World) swordRTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig, tick uint64) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	pickPadding := skillMetaRange(skill, "targetPickPadding", 80)
	w.ForEachEntity(func(target *Entity) {
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
	w.ForEachEntity(func(target *Entity) {
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

func (w *World) swordRTargets(entity *Entity, center Vector2, skill config.SkillConfig, tick uint64) []*Entity {
	hits := make([]*Entity, 0)
	w.ForEachEntity(func(target *Entity) {
		if !isAirborneEnemyHero(entity, target, tick) {
			return
		}
		if distance(center, target.Position) <= skillMetaRange(skill, "hitRadius", 450)+target.Radius {
			hits = append(hits, target)
		}
	})
	return hits
}

func isAirborneEnemyHero(attacker *Entity, target *Entity, tick uint64) bool {
	if !canAttackTarget(attacker, target) {
		return false
	}
	if target.Control.AirborneUntilTick <= tick {
		return false
	}
	return target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero
}

func swordRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{200, 300, 400})
	return physicalDamageAfterResistance(attacker, target, baseDamage+attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 1.5), tick)
}
