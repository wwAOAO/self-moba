package warrior

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Warrior.JusticePending {
		return
	}
	target := rTarget(w, entity, world.Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.435), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Warrior.JusticePending = true
	entity.Warrior.JusticeReleaseTick = tick + windupTicks
	entity.Warrior.JusticeTargetID = target.ID
	entity.Warrior.JusticeLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Warrior.JusticeReleaseTick
	entity.Skills[rID] = state
}

func ReleaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Warrior.JusticePending || tick < entity.Warrior.JusticeReleaseTick {
		return
	}
	target := w.EntityByID(entity.Warrior.JusticeTargetID)
	level := entity.Warrior.JusticeLevel
	entity.Warrior.JusticePending = false
	entity.Warrior.JusticeReleaseTick = 0
	entity.Warrior.JusticeTargetID = ""
	entity.Warrior.JusticeLevel = 0
	if !canAttackTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(rID)
	state := entity.Skills[rID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{120000, 100000, 80000})), tickRate)
	entity.Skills[rID] = state
	damage := RDamage(target, skill, level)
	target.Combat.LastHitTick = tick
	if target.Kind != world.EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.ApplyTrueDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = w.WarriorTrueDamageAfterReduction(target, damage, tick)
		target.Combat.LastDamageType = "true"
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func RDamage(target *world.Entity, skill config.SkillConfig, level int) float64 {
	if target == nil {
		return 0
	}
	baseDamage := skillList(skill, "baseDamage", level, []float64{150, 250, 350})
	missingHPRatio := skillList(skill, "missingHPRatio", level, []float64{0.25, 0.3, 0.35})
	missingHP := target.Stats.MaxHP - target.Stats.HP
	if missingHP < 0 {
		missingHP = 0
	}
	return baseDamage + float64(missingHP)*missingHPRatio
}

func rTarget(w *world.World, entity *world.Entity, targetPoint world.Vector2, skill config.SkillConfig) *world.Entity {
	var best *world.Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 400)
	pickPadding := skillMeta(skill, "targetPickPadding", 80)
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		if distance(entity.Position, target.Position) > castRange+target.Radius {
			return
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+pickPadding {
			return
		}
		if distToPoint < bestDistance {
			bestDistance = distToPoint
			best = target
		}
	})
	return best
}
