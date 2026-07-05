package tank

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID = "tank"
	qID    = "slam"
	rID    = "earthquake"
)

func init() {
	world.RegisterHeroCastHandlers(heroID, map[string]world.HeroCastHandler{
		qID: ApplyQ,
		rID: ApplyR,
	})
}

func ApplyQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Tank.SeismicShardPending {
		return
	}
	target := w.EntityByID(cast.TargetID)
	if target == nil {
		target = tankQTarget(w, entity, world.Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	}
	if !canAttackTarget(entity, target) || distance(entity.Position, target.Position) > skillRange(skill, 625)+target.Radius {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{70, 75, 80, 85, 90})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Tank.SeismicShardPending = true
	entity.Tank.SeismicShardReleaseTick = tick + windupTicks
	entity.Tank.SeismicShardTargetID = target.ID
	entity.Tank.SeismicShardLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Tank.SeismicShardReleaseTick
	entity.Skills[qID] = state
}

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	start := entity.Position
	rRange := skillRange(skill, 1000)
	targetPoint := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	if distance(start, targetPoint) > rRange {
		dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
		if dx == 0 && dy == 0 {
			return
		}
		castPosition := w.ClampWorldPoint(world.Vector2{X: targetPoint.X - dx*rRange, Y: targetPoint.Y - dy*rRange})
		entity.Tank.UnstoppableCastPending = true
		entity.Tank.UnstoppableCastTarget = targetPoint
		entity.Tank.UnstoppableCastLevel = state.Level
		entity.Intent.MoveTarget = &castPosition
		entity.Intent.AttackTargetID = ""
		entity.Intent.AttackPausedTill = 0
		return
	}
	w.StartTankRDash(entity, targetPoint, state, skill, tick, tickRate)
}

func tankQTarget(w *world.World, entity *world.Entity, targetPoint world.Vector2, skill config.SkillConfig) *world.Entity {
	var best *world.Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 625)
	pickPadding := skillMeta(skill, "targetPickPadding", 90)
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) || distance(entity.Position, target.Position) > castRange+target.Radius {
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

func canAttackTarget(attacker *world.Entity, target *world.Entity) bool {
	return attacker != nil && target != nil && attacker.Team != target.Team && target.Stats.HP > 0
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if skill.MetaLists != nil && len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
	}
	if level < 1 {
		level = 1
	}
	if level > len(values) {
		level = len(values)
	}
	return values[level-1]
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
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
