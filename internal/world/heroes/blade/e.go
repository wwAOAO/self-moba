package blade

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 {
		return
	}
	start := entity.Position
	dx, dy := normalize(cast.TargetX-start.X, cast.TargetY-start.Y)
	if dx == 0 && dy == 0 {
		return
	}
	dashRange := skill.Range
	if dashRange <= 0 {
		dashRange = 650
	}
	end := w.ClampWorldPoint(world.Vector2{X: start.X + dx*dashRange, Y: start.Y + dy*dashRange})
	hits := bladeETargets(w, entity, start, world.Vector2{X: dx, Y: dy}, distance(start, end))
	for _, target := range hits {
		damage := bladeEDamage(w, entity, target, skill, state.Level, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
		}
		gainRage(entity, skillMeta(skill, "ragePerHit", 2), tick)
	}
	speed := skillMeta(skill, "dashSpeed", 1800)
	if speed <= 0 {
		speed = dashRange
	}
	travelTicks := uint64(math.Ceil(distance(start, end) / speed * float64(tickRate)))
	if travelTicks < 1 {
		travelTicks = 1
	}
	entity.Intent = world.IntentState{}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = start
	entity.Control.DashEnd = end
	entity.Control.DashUntilTick = tick + travelTicks
	entity.Control.ActionLockedUntilTick = entity.Control.DashUntilTick
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:blade_e_whirlwind:"),
		Kind:         "blade_e_whirlwind",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        start,
		End:          end,
		Dir:          world.Vector2{X: dx, Y: dy},
		Radius:       skillMeta(skill, "effectRadius", 70),
		Speed:        distance(start, end) / float64(travelTicks),
		CreatedAt:    tick,
		ExpiresAt:    entity.Control.DashUntilTick,
	})
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{13000, 12000, 11000, 10000, 9000})), tickRate)
	entity.Skills[eID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func RefundEOnCrit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if w == nil || attacker == nil || target == nil || attacker.HeroID != heroID || tickRate <= 0 {
		return
	}
	if !w.AttackCrits(attacker, target, tick) {
		return
	}
	state := attacker.Skills[eID]
	if state.CooldownUntilTick <= tick {
		return
	}
	skill := w.SkillConfig(eID)
	refundSeconds := skillMeta(skill, "minionCritRefundSeconds", 0.75)
	if target.Kind == world.EntityKindPlayer || target.Kind == world.EntityKindEnemyHero {
		refundSeconds = skillMeta(skill, "heroCritRefundSeconds", 1.5)
	}
	refundTicks := secondsToTicks(refundSeconds, tickRate)
	if state.CooldownUntilTick <= tick+refundTicks {
		state.CooldownUntilTick = tick
	} else {
		state.CooldownUntilTick -= refundTicks
	}
	attacker.Skills[eID] = state
}

func bladeEDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	rawDamage := skillList(skill, "baseDamage", level, []float64{80, 110, 140, 170, 200}) +
		attacker.Stats.BonusAttack*skillMeta(skill, "bonusAdRatio", 1.3)
	return w.PhysicalDamageAfterResistance(attacker, target, rawDamage, tick)
}

func bladeETargets(w *world.World, entity *world.Entity, start world.Vector2, dir world.Vector2, dashRange float64) []*world.Entity {
	targets := make([]*world.Entity, 0)
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		along, perpendicular := projectPoint(start, dir, target.Position)
		if along < -target.Radius || along > dashRange+target.Radius {
			return
		}
		if perpendicular <= entity.Radius+target.Radius {
			targets = append(targets, target)
		}
	})
	return targets
}

func canAttackTarget(attacker *world.Entity, target *world.Entity) bool {
	return attacker != nil && target != nil && attacker.Team != target.Team && target.Stats.HP > 0
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func projectPoint(origin world.Vector2, dir world.Vector2, point world.Vector2) (float64, float64) {
	vx := point.X - origin.X
	vy := point.Y - origin.Y
	along := vx*dir.X + vy*dir.Y
	px := origin.X + dir.X*along
	py := origin.Y + dir.Y*along
	return along, math.Hypot(point.X-px, point.Y-py)
}
