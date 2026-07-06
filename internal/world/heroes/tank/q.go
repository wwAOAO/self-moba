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
	wID    = "guard"
	eID    = "taunt"
	rID    = "earthquake"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: ApplyQ,
			wID: ApplyW,
			eID: ApplyE,
			rID: ApplyR,
		},
		Tick:              Tick,
		OnBasicHit:        ApplyWAftershock,
		StartRDash:        StartRDash,
		ReleasePreparedR:  ReleasePreparedR,
		CancelPreparedR:   CancelPreparedR,
		ResolveRImpact:    ResolveRImpact,
		RefreshWPassive:   RefreshWPassive,
		PassiveState:      PassiveState,
		RefreshGranite:    RefreshGranite,
		RefreshGraniteMax: RefreshGraniteMax,
		WBonusDamage:      WBonusDamage,
		TankQDamage:       QDamage,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	TickGranite(w, entity, tick, tickRate)
	RefreshWPassive(w, entity)
	ReleaseQ(w, entity, tick, tickRate)
	ReleaseE(w, entity, tick, tickRate)
}

func ReleaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Tank.SeismicShardPending || tick < entity.Tank.SeismicShardReleaseTick {
		return
	}
	target := w.EntityByID(entity.Tank.SeismicShardTargetID)
	level := entity.Tank.SeismicShardLevel
	entity.Tank.SeismicShardPending = false
	entity.Tank.SeismicShardReleaseTick = 0
	entity.Tank.SeismicShardTargetID = ""
	entity.Tank.SeismicShardLevel = 0
	if !canAttackTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(qID)
	state := entity.Skills[qID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{8000, 8000, 8000, 8000, 8000})), tickRate)
	entity.Skills[qID] = state
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	qRange := skillRange(skill, 625)
	speedPerSecond := skillMeta(skill, "projectileSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	id := w.NextProjectileID("projectile:tank_q:")
	w.PutProjectile(&world.Projectile{
		ID:           id,
		Kind:         "tank_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 45),
		Damage:       level,
		EffectRatio:  skillList(skill, "moveSpeedSteal", level, []float64{0.2, 0.25, 0.3, 0.35, 0.4}),
		EffectTicks:  secondsToTicks(skillMeta(skill, "moveSpeedStealSeconds", 3), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	})
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillList(skill, "baseDamage", skillLevel, []float64{70, 120, 170, 220, 270})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.6)
	return w.TankMagicDamageAfterResistance(attacker, target, rawDamage, tick)
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

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
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
