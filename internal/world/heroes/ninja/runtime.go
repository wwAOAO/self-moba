package ninja

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID = "ninja"
	qID    = "ninja_q"
	wID    = "ninja_w"
	eID    = "ninja_e"
	rID    = "ninja_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		Tick:                        Tick,
		OnDamage:                    MarkDamage,
		SpecialRecast:               SpecialRecast,
		ReleasePreparedR:            ReleasePreparedR,
		CancelPreparedR:             CancelPreparedR,
		BasicAttackBonusMagicDamage: PassiveDamage,
		NinjaQDamage:                QDamage,
		NinjaSkillHit:               SkillHit,
	})
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 {
		return
	}
	target := rTarget(w, entity, cast, skill)
	if target == nil {
		return
	}
	castPoint := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	if !rInRange(entity, target, skill) {
		prepareR(w, entity, target, state, skill, castPoint)
		entity.Skills[rID] = state
		return
	}
	startRWindup(entity, target, state, skill, castPoint, tick, tickRate)
}

func startRWindup(entity *world.Entity, target *world.Entity, state world.SkillState, skill config.SkillConfig, castPoint world.Vector2, tick uint64, tickRate int) {
	if entity == nil || target == nil {
		return
	}
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.6), tickRate)
	dashTicks := secondsToTicks(skillMeta(skill, "dashSeconds", 0.35), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	if dashTicks < 1 {
		dashTicks = 1
	}
	entity.Ninja.RCastPending = false
	entity.Ninja.RCastTargetID = ""
	entity.Ninja.RCastLevel = 0
	entity.Ninja.RCastPoint = castPoint
	entity.Ninja.RCastPointSet = true
	entity.Ninja.RPending = true
	entity.Ninja.RReleaseTick = tick + windupTicks
	entity.Ninja.RDashEndTick = entity.Ninja.RReleaseTick + dashTicks
	entity.Ninja.RTargetID = target.ID
	entity.Ninja.RLevel = state.Level
	entity.Intent.MoveTarget = nil
	entity.Control.ActionLockedUntilTick = entity.Ninja.RDashEndTick
	entity.Control.UntargetableUntilTick = entity.Ninja.RDashEndTick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{120000, 110000, 100000})), tickRate)
	entity.Skills[rID] = state
}

func rTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) *world.Entity {
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	pickPadding := skillMeta(skill, "targetPickPadding", 80)
	if target := w.EntityByID(cast.TargetID); validRTarget(entity, target) && distance(point, target.Position) <= target.Radius+pickPadding {
		return target
	}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !validRTarget(entity, target) {
			return
		}
		distToPoint := distance(point, target.Position)
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

func validRTarget(entity *world.Entity, target *world.Entity) bool {
	return entity != nil && target != nil && world.IsHeroUnit(target) && entity.Team != target.Team && target.Stats.HP > 0 && target.Control.UntargetableUntilTick == 0
}

func rInRange(entity *world.Entity, target *world.Entity, skill config.SkillConfig) bool {
	return entity != nil && target != nil && distance(entity.Position, target.Position) <= skillRange(skill, 625)+target.Radius
}

func prepareR(w *world.World, entity *world.Entity, target *world.Entity, state world.SkillState, skill config.SkillConfig, castPoint world.Vector2) {
	if entity == nil || target == nil {
		return
	}
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		return
	}
	castPosition := w.ClampWorldPoint(world.Vector2{
		X: target.Position.X - dx*skillRange(skill, 625),
		Y: target.Position.Y - dy*skillRange(skill, 625),
	})
	entity.Ninja.RCastPending = true
	entity.Ninja.RCastTargetID = target.ID
	entity.Ninja.RCastLevel = state.Level
	entity.Ninja.RCastPoint = castPoint
	entity.Ninja.RCastPointSet = true
	entity.Intent.MoveTarget = &castPosition
	entity.Intent.AttackTargetID = ""
	entity.Intent.AttackPausedTill = 0
}

func ReleasePreparedR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Ninja.RCastPending {
		return
	}
	target := w.EntityByID(entity.Ninja.RCastTargetID)
	state := entity.Skills[rID]
	if state.Level <= 0 || !validRTarget(entity, target) {
		CancelPreparedR(entity)
		entity.Skills[rID] = state
		return
	}
	state.Level = entity.Ninja.RCastLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	skill := w.SkillConfig(rID)
	if !rInRange(entity, target, skill) {
		prepareR(w, entity, target, state, skill, entity.Ninja.RCastPoint)
		entity.Skills[rID] = state
		return
	}
	if tick < state.CooldownUntilTick {
		CancelPreparedR(entity)
		entity.Skills[rID] = state
		return
	}
	startRWindup(entity, target, state, skill, entity.Ninja.RCastPoint, tick, tickRate)
}

func CancelPreparedR(entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID || !entity.Ninja.RCastPending {
		return
	}
	entity.Ninja.RCastPending = false
	entity.Ninja.RCastTargetID = ""
	entity.Ninja.RCastLevel = 0
	entity.Ninja.RCastPoint = world.Vector2{}
	entity.Ninja.RCastPointSet = false
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 40)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{5000, 4500, 4000, 3500, 3000})), tickRate)
	entity.Skills[eID] = state
	showERange(w, entity, skill, tick, tickRate)

	groupID := w.NextEffectID("effect:ninja_e_group:")
	hits := eHits(w, entity, skillRange(skill, 290), tick)
	heroHits, hitIDs := applyEHits(w, entity, skill, state.Level, groupID, hits, nil, tick, tickRate)
	if wShadowMoving(entity, tick) {
		entity.Ninja.PendingShadowELevel = state.Level
		entity.Ninja.PendingShadowEGroup = groupID
		entity.Ninja.PendingShadowEHitIDs = hitIDs
	}
	reduceWCooldown(entity, heroHits, tick, secondsToTicks(skillMeta(skill, "wCooldownRefundSeconds", 3), tickRate))
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func applyEHits(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, groupID string, hits map[string]uint8, alreadyHit map[string]bool, tick uint64, tickRate int) (int, map[string]bool) {
	heroHits := 0
	hitIDs := map[string]bool{}
	for targetID, mark := range hits {
		target := w.EntityByID(targetID)
		if target == nil {
			continue
		}
		previouslyHit := alreadyHit[targetID]
		hitIDs[targetID] = true
		target.Combat.LastHitTick = tick
		wasAlive := target.Stats.HP > 0
		if !previouslyHit {
			damage := w.PhysicalDamageAfterResistance(entity, target, eRawDamage(entity, skill, level), tick)
			if target.Kind == world.EntityKindDummy {
				target.Combat.LastDamage = damage
				target.Combat.LastDamageType = "physical"
			} else {
				w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
			}
		}
		if world.IsHeroUnit(target) {
			heroHits++
			if mark&1 != 0 {
				SkillHit(w, entity, target, eID, groupID, false, tick, tickRate)
			}
			if mark&2 != 0 {
				SkillHit(w, entity, target, eID, groupID, true, tick, tickRate)
			}
		}
		if mark&2 != 0 {
			slow := skillList(skill, "shadowSlow", level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
			if previouslyHit || mark&1 != 0 {
				slow = skillList(skill, "doubleSlow", level, []float64{0.3, 0.375, 0.45, 0.45, 0.6})
			}
			w.ApplyMoveSpeedSlow(target, slow, tick+secondsToTicks(skillMeta(skill, "slowSeconds", 1.5), tickRate))
		}
		if !previouslyHit && wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	return heroHits, hitIDs
}

func showERange(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	radius := skillRange(skill, 290)
	expiresAt := tick + secondsToTicks(skillMeta(skill, "rangeDisplaySeconds", 0.35), tickRate)
	putERangeEffect(w, entity, entity.Position, radius, tick, expiresAt)
	for _, position := range readyShadowPositions(entity, tick) {
		putERangeEffect(w, entity, position, radius, tick, expiresAt)
	}
}

func putERangeEffect(w *world.World, entity *world.Entity, center world.Vector2, radius float64, tick uint64, expiresAt uint64) {
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:ninja_e:"),
		Kind:         "ninja_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        center,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func eHits(w *world.World, entity *world.Entity, radius float64, tick uint64) map[string]uint8 {
	hits := map[string]uint8{}
	for _, target := range w.TargetsInRadius(entity, entity.Position, radius) {
		hits[target.ID] |= 1
	}
	for _, position := range readyShadowPositions(entity, tick) {
		for _, target := range w.TargetsInRadius(entity, position, radius) {
			hits[target.ID] |= 2
		}
	}
	return hits
}

func eRawDamage(attacker *world.Entity, skill config.SkillConfig, level int) float64 {
	bonusAD := 0.0
	if attacker != nil {
		bonusAD = attacker.Stats.BonusAttack
	}
	return skillList(skill, "baseDamage", level, []float64{70, 92.5, 115, 137.5, 160}) + bonusAD*skillMeta(skill, "bonusAdRatio", 0.7)
}

func reduceWCooldown(entity *world.Entity, hits int, tick uint64, refundTicks uint64) {
	if entity == nil || hits <= 0 || refundTicks == 0 {
		return
	}
	state := entity.Skills[wID]
	reduction := uint64(hits) * refundTicks
	if state.CooldownUntilTick <= tick+reduction {
		state.CooldownUntilTick = tick
	} else {
		state.CooldownUntilTick -= reduction
	}
	entity.Skills[wID] = state
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Ninja.QPending {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{55, 60, 65, 70, 75})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Ninja.QPending = true
	entity.Ninja.QReleaseTick = tick + windupTicks
	entity.Ninja.QTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Ninja.QLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Ninja.QReleaseTick
	entity.Skills[qID] = state
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Ninja.QPending || tick < entity.Ninja.QReleaseTick {
		return
	}
	target := entity.Ninja.QTarget
	level := entity.Ninja.QLevel
	entity.Ninja.QPending = false
	entity.Ninja.QReleaseTick = 0
	entity.Ninja.QTarget = world.Vector2{}
	entity.Ninja.QLevel = 0
	if level <= 0 {
		level = 1
	}
	skill := w.SkillConfig(qID)
	state := entity.Skills[qID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{6000, 6000, 6000, 6000, 6000})), tickRate)
	entity.Skills[qID] = state

	groupID := w.NextProjectileID("projectile:ninja_q_group:")
	entity.Ninja.SkillHitMarks = map[string]uint8{}
	entity.Ninja.SkillEnergyRefunded = map[string]bool{}
	fireShurikenAt(w, entity, entity.Position, target, skill, level, groupID, false, tick, tickRate)
	for _, position := range readyShadowPositions(entity, tick) {
		fireShurikenAt(w, entity, position, target, skill, level, groupID, true, tick, tickRate)
	}
	if wShadowMoving(entity, tick) {
		entity.Ninja.PendingShadowQTarget = target
		entity.Ninja.PendingShadowQLevel = level
		entity.Ninja.PendingShadowQGroup = groupID
	}
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func fireShurikenAt(w *world.World, entity *world.Entity, start world.Vector2, target world.Vector2, skill config.SkillConfig, level int, groupID string, fromShadow bool, tick uint64, tickRate int) {
	dx, dy := normalize(target.X-start.X, target.Y-start.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	fireShuriken(w, entity, start, world.Vector2{X: dx, Y: dy}, skill, level, groupID, fromShadow, tick, tickRate)
}

func fireShuriken(w *world.World, entity *world.Entity, start world.Vector2, dir world.Vector2, skill config.SkillConfig, level int, groupID string, fromShadow bool, tick uint64, tickRate int) {
	qRange := skillRange(skill, 900)
	speedPerSecond := skillMeta(skill, "projectileSpeed", 1700)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:ninja_q:"),
		Kind:         "ninja_shuriken",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		GroupID:      groupID,
		Position:     start,
		Start:        start,
		Dir:          dir,
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 87.5),
		Damage:       level,
		FromShadow:   fromShadow,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speedPerSecond+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{20, 25, 30, 35, 40})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{20000, 19000, 18000, 17000, 16000})), tickRate)
	entity.Skills[wID] = state

	dx, dy := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	dashRange := skillRange(skill, 650)
	target := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	distanceToTarget := distance(entity.Position, target)
	if distanceToTarget > 0 && distanceToTarget < dashRange {
		dashRange = distanceToTarget
	}
	entity.Ninja.ShadowPosition = w.ClampWorldPoint(world.Vector2{
		X: entity.Position.X + dx*dashRange,
		Y: entity.Position.Y + dy*dashRange,
	})
	entity.Ninja.ShadowExpiresAt = tick + secondsToTicks(skillMeta(skill, "shadowDurationSeconds", 5), tickRate)
	entity.Ninja.ShadowReadyTick = tick + secondsToTicks(skillMeta(skill, "shadowDashSeconds", 0.25), tickRate)
	entity.Ninja.ShadowRecastSkillID = wID
	entity.Ninja.ShadowRecastUntil = entity.Ninja.ShadowExpiresAt
	if entity.Ninja.ShadowEffectID == "" {
		entity.Ninja.ShadowEffectID = w.NextEffectID("effect:ninja_shadow:")
	}
	putShadowEffect(w, entity, entity.Ninja.ShadowEffectID, entity.Ninja.ShadowPosition, entity.Ninja.ShadowExpiresAt, tick, entity.Position, shadowSpeedPerTick(entity.Position, entity.Ninja.ShadowPosition, skillMeta(skill, "shadowDashSeconds", 0.25), tickRate))
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func SpecialRecast(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil {
		return false
	}
	oldPosition := entity.Position
	if cast.SkillID == wID && wShadowReady(entity, tick) && entity.Ninja.ShadowRecastSkillID == wID && tick < entity.Ninja.ShadowRecastUntil {
		entity.Position = entity.Ninja.ShadowPosition
		entity.Ninja.ShadowPosition = oldPosition
		entity.Ninja.ShadowRecastSkillID = ""
		entity.Ninja.ShadowRecastUntil = 0
		putShadowEffect(w, entity, entity.Ninja.ShadowEffectID, entity.Ninja.ShadowPosition, entity.Ninja.ShadowExpiresAt, tick, entity.Ninja.ShadowPosition, 0)
		return true
	}
	if cast.SkillID == rID && rShadowActive(entity, tick) && tick < entity.Ninja.RShadowRecastUntil {
		entity.Position = entity.Ninja.RShadowPosition
		entity.Ninja.RShadowPosition = oldPosition
		entity.Ninja.RShadowRecastUntil = 0
		putShadowEffect(w, entity, entity.Ninja.RShadowEffectID, entity.Ninja.RShadowPosition, entity.Ninja.RShadowExpiresAt, tick, entity.Ninja.RShadowPosition, 0)
		return true
	}
	return false
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	releaseQ(w, entity, tick, tickRate)
	releasePendingShadowCopies(w, entity, tick, tickRate)
	tickR(w, entity, tick, tickRate)
	if entity.Ninja.ShadowExpiresAt > 0 && tick >= entity.Ninja.ShadowExpiresAt {
		entity.Ninja.ShadowExpiresAt = 0
		entity.Ninja.ShadowEffectID = ""
		entity.Ninja.ShadowReadyTick = 0
		entity.Ninja.ShadowRecastSkillID = ""
		entity.Ninja.ShadowRecastUntil = 0
		entity.Ninja.PendingShadowQLevel = 0
		entity.Ninja.PendingShadowQGroup = ""
		entity.Ninja.PendingShadowELevel = 0
		entity.Ninja.PendingShadowEGroup = ""
		entity.Ninja.PendingShadowEHitIDs = nil
	}
	if entity.Ninja.RShadowExpiresAt > 0 && tick >= entity.Ninja.RShadowExpiresAt {
		entity.Ninja.RShadowExpiresAt = 0
		entity.Ninja.RShadowEffectID = ""
		entity.Ninja.RShadowRecastUntil = 0
	}
}

func releasePendingShadowCopies(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if !wShadowReady(entity, tick) {
		return
	}
	if entity.Ninja.PendingShadowQLevel > 0 && entity.Ninja.PendingShadowQGroup != "" {
		fireShurikenAt(w, entity, entity.Ninja.ShadowPosition, entity.Ninja.PendingShadowQTarget, w.SkillConfig(qID), entity.Ninja.PendingShadowQLevel, entity.Ninja.PendingShadowQGroup, true, tick, tickRate)
		entity.Ninja.PendingShadowQLevel = 0
		entity.Ninja.PendingShadowQGroup = ""
	}
	if entity.Ninja.PendingShadowELevel > 0 && entity.Ninja.PendingShadowEGroup != "" {
		skill := w.SkillConfig(eID)
		hits := map[string]uint8{}
		for _, target := range w.TargetsInRadius(entity, entity.Ninja.ShadowPosition, skillRange(skill, 290)) {
			hits[target.ID] = 2
		}
		heroHits, _ := applyEHits(w, entity, skill, entity.Ninja.PendingShadowELevel, entity.Ninja.PendingShadowEGroup, hits, entity.Ninja.PendingShadowEHitIDs, tick, tickRate)
		reduceWCooldown(entity, heroHits, tick, secondsToTicks(skillMeta(skill, "wCooldownRefundSeconds", 3), tickRate))
		entity.Ninja.PendingShadowELevel = 0
		entity.Ninja.PendingShadowEGroup = ""
		entity.Ninja.PendingShadowEHitIDs = nil
	}
}

func tickR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity.Ninja.RPending && entity.Ninja.RReleaseTick > 0 && tick >= entity.Ninja.RReleaseTick && entity.Ninja.RDashEndTick > 0 && entity.Control.DashUntilTick == 0 {
		startRDash(w, entity, tick, tickRate)
	}
	if entity.Ninja.RPending && entity.Ninja.RDashEndTick > 0 && tick >= entity.Ninja.RDashEndTick {
		finishRDash(w, entity, tick, tickRate)
	}
	if entity.Ninja.RMarkTriggerTick > 0 && tick >= entity.Ninja.RMarkTriggerTick {
		triggerRMark(w, entity, tick, tickRate)
	}
}

func startRDash(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	target := w.EntityByID(entity.Ninja.RTargetID)
	if !validRMarkedTarget(entity, target) {
		entity.Ninja.RPending = false
		entity.Control.UntargetableUntilTick = 0
		return
	}
	entity.Ninja.RShadowPosition = rShadowSpawnPosition(entity)
	entity.Ninja.RCastPoint = world.Vector2{}
	entity.Ninja.RCastPointSet = false
	skill := w.SkillConfig(rID)
	entity.Ninja.RShadowExpiresAt = tick + secondsToTicks(skillMeta(skill, "shadowDurationSeconds", 7.5), tickRate)
	lingerTicks := secondsToTicks(skillMeta(skill, "shadowLingerSeconds", 1.5), tickRate)
	if lingerTicks >= entity.Ninja.RShadowExpiresAt-tick {
		entity.Ninja.RShadowRecastUntil = entity.Ninja.RShadowExpiresAt
	} else {
		entity.Ninja.RShadowRecastUntil = entity.Ninja.RShadowExpiresAt - lingerTicks
	}
	if entity.Ninja.RShadowEffectID == "" {
		entity.Ninja.RShadowEffectID = w.NextEffectID("effect:ninja_shadow:")
	}
	putShadowEffect(w, entity, entity.Ninja.RShadowEffectID, entity.Ninja.RShadowPosition, entity.Ninja.RShadowExpiresAt, tick, entity.Ninja.RShadowPosition, 0)
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = target.Position
	entity.Control.DashUntilTick = entity.Ninja.RDashEndTick
}

func rShadowSpawnPosition(entity *world.Entity) world.Vector2 {
	if entity != nil && entity.Ninja.RCastPointSet {
		return entity.Ninja.RCastPoint
	}
	if entity != nil {
		return entity.Position
	}
	return world.Vector2{}
}

func finishRDash(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	target := w.EntityByID(entity.Ninja.RTargetID)
	entity.Position = entity.Control.DashEnd
	entity.Control.DashUntilTick = 0
	entity.Control.DashStartTick = 0
	if entity.Control.UntargetableUntilTick <= tick {
		entity.Control.UntargetableUntilTick = 0
	}
	entity.Ninja.RPending = false
	entity.Ninja.RReleaseTick = 0
	entity.Ninja.RDashEndTick = 0
	if !validRMarkedTarget(entity, target) {
		return
	}
	entity.Ninja.RMarkTargetID = target.ID
	entity.Ninja.RMarkTriggerTick = tick + secondsToTicks(skillMeta(w.SkillConfig(rID), "markDelaySeconds", 3), tickRate)
	entity.Ninja.RMarkDamage = 0
	entity.Ninja.RMarkLevel = entity.Ninja.RLevel
}

func triggerRMark(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	target := w.EntityByID(entity.Ninja.RMarkTargetID)
	level := entity.Ninja.RMarkLevel
	storedDamage := entity.Ninja.RMarkDamage
	entity.Ninja.RMarkTargetID = ""
	entity.Ninja.RMarkTriggerTick = 0
	entity.Ninja.RMarkDamage = 0
	entity.Ninja.RMarkLevel = 0
	if !validRMarkedTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(rID)
	ratio := skillList(skill, "storedDamageRatio", level, []float64{0.25, 0.4, 0.55})
	rawDamage := entity.Stats.Attack*skillMeta(skill, "attackRatio", 1) + storedDamage*ratio
	damage := w.PhysicalDamageAfterResistance(entity, target, rawDamage, tick)
	target.Combat.LastHitTick = tick
	wasAlive := target.Stats.HP > 0
	w.ApplyDamage(entity, target, damage, tickRate)
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(entity, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}

func MarkDamage(w *world.World, source *world.Entity, target *world.Entity, basicAttack bool, pet bool, skipBleed bool, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID || source.Ninja.RMarkTargetID != target.ID || source.Ninja.RMarkTriggerTick == 0 || tick >= source.Ninja.RMarkTriggerTick {
		return
	}
	source.Ninja.RMarkDamage += float64(target.Combat.LastDamage)
}

func validRMarkedTarget(entity *world.Entity, target *world.Entity) bool {
	return entity != nil && target != nil && world.IsHeroUnit(target) && entity.Team != target.Team && target.Stats.HP > 0
}

func putShadowEffect(w *world.World, entity *world.Entity, id string, end world.Vector2, expiresAt uint64, tick uint64, start world.Vector2, speed float64) {
	dx, dy := normalize(end.X-start.X, end.Y-start.Y)
	w.PutSkillEffect(world.SkillEffect{
		ID:           id,
		Kind:         "ninja_shadow",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        start,
		End:          end,
		Dir:          world.Vector2{X: dx, Y: dy},
		Radius:       entity.Radius,
		Speed:        speed,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func shadowSpeedPerTick(start world.Vector2, end world.Vector2, seconds float64, tickRate int) float64 {
	ticks := secondsToTicks(seconds, tickRate)
	if ticks == 0 {
		return 0
	}
	return distance(start, end) / float64(ticks)
}

func wShadowActive(entity *world.Entity, tick uint64) bool {
	return entity != nil && entity.Ninja.ShadowExpiresAt > tick
}

func wShadowReady(entity *world.Entity, tick uint64) bool {
	return wShadowActive(entity, tick) && (entity.Ninja.ShadowReadyTick == 0 || tick >= entity.Ninja.ShadowReadyTick)
}

func wShadowMoving(entity *world.Entity, tick uint64) bool {
	return wShadowActive(entity, tick) && entity.Ninja.ShadowReadyTick > tick
}

func rShadowActive(entity *world.Entity, tick uint64) bool {
	return entity != nil && entity.Ninja.RShadowExpiresAt > tick
}

func readyShadowPositions(entity *world.Entity, tick uint64) []world.Vector2 {
	positions := make([]world.Vector2, 0, 2)
	if wShadowReady(entity, tick) {
		positions = append(positions, entity.Ninja.ShadowPosition)
	}
	if rShadowActive(entity, tick) {
		positions = append(positions, entity.Ninja.RShadowPosition)
	}
	return positions
}

func SkillHit(w *world.World, source *world.Entity, target *world.Entity, skillID string, groupID string, fromShadow bool, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID || groupID == "" || !world.IsHeroUnit(target) || source.Team == target.Team {
		return
	}
	wLevel := source.Skills[wID].Level
	if wLevel <= 0 {
		return
	}
	if source.Ninja.SkillHitMarks == nil {
		source.Ninja.SkillHitMarks = map[string]uint8{}
	}
	if source.Ninja.SkillEnergyRefunded == nil {
		source.Ninja.SkillEnergyRefunded = map[string]bool{}
	}
	key := groupID + ":" + target.ID
	mark := uint8(1)
	if fromShadow {
		mark = 2
	}
	source.Ninja.SkillHitMarks[key] |= mark
	if source.Ninja.SkillHitMarks[key] != 3 || source.Ninja.SkillEnergyRefunded[groupID] {
		return
	}
	source.Ninja.SkillEnergyRefunded[groupID] = true
	refund := skillList(w.SkillConfig(wID), "energyRefund", wLevel, []float64{30, 35, 40, 45, 50})
	source.Stats.MP += refund
	if source.Stats.MP > source.Stats.MaxMP {
		source.Stats.MP = source.Stats.MaxMP
	}
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, hitNumber int, tick uint64) int {
	return w.PhysicalDamageAfterResistance(attacker, target, qRawDamage(attacker, skill, skillLevel, hitNumber), tick)
}

func qRawDamage(attacker *world.Entity, skill config.SkillConfig, level int, hitNumber int) float64 {
	baseKey := "firstBaseDamage"
	ratio := skillMeta(skill, "firstBonusAdRatio", 1)
	fallback := []float64{80, 120, 160, 200, 240}
	if hitNumber > 1 {
		baseKey = "laterBaseDamage"
		ratio = skillMeta(skill, "laterBonusAdRatio", 0.6)
		fallback = []float64{48, 72, 96, 120, 144}
	}
	bonusAD := 0.0
	if attacker != nil {
		bonusAD = attacker.Stats.BonusAttack
	}
	return skillList(skill, baseKey, level, fallback) + bonusAD*ratio
}

func PassiveDamage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) int {
	rawDamage := passiveRawDamage(attacker, target, w.HeroPassiveSkill(attacker), tick, tickRate)
	if rawDamage <= 0 {
		return 0
	}
	return w.MagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func passiveRawDamage(attacker *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) float64 {
	if attacker == nil || target == nil || attacker.HeroID != heroID || target.Team == attacker.Team || target.Stats.MaxHP <= 0 || target.Stats.HP*2 >= target.Stats.MaxHP {
		return 0
	}
	if target.Kind != world.EntityKindDummy && !world.IsHeroUnit(target) && !world.IsMinion(target) && !world.IsMonster(target) {
		return 0
	}
	if world.IsHeroUnit(target) {
		if attacker.Passive.NinjaSoulCooldowns == nil {
			attacker.Passive.NinjaSoulCooldowns = map[string]uint64{}
		}
		if attacker.Passive.NinjaSoulCooldowns[target.ID] > tick {
			return 0
		}
		attacker.Passive.NinjaSoulCooldowns[target.ID] = tick + secondsToTicks(skillMeta(skill, "heroCooldownSeconds", 10), tickRate)
	}
	rawDamage := target.Stats.MaxHP * skillCurve(skill, "maxHPRatio", "maxHPRatioLevels", attacker.Level, 0.05)
	if world.IsMonster(target) {
		rawDamage *= skillMeta(skill, "monsterDamageMultiplier", 0.75)
		if target.Kind == world.EntityKindBaronNashor {
			rawDamage = math.Min(rawDamage, skillMeta(skill, "epicMonsterDamageCap", 175))
		}
	}
	return rawDamage
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

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillCurve(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	if skill.MetaLists == nil {
		return fallback
	}
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	currentLevel := float64(clampInt(level, world.MinHeroLevel, world.MaxHeroLevel))
	if currentLevel <= levels[0] {
		return values[0]
	}
	last := len(values) - 1
	if currentLevel >= levels[last] {
		return values[last]
	}
	for i := 1; i < len(values); i++ {
		if currentLevel > levels[i] {
			continue
		}
		fromLevel := levels[i-1]
		toLevel := levels[i]
		if toLevel <= fromLevel {
			return values[i]
		}
		t := (currentLevel - fromLevel) / (toLevel - fromLevel)
		return values[i-1] + (values[i]-values[i-1])*t
	}
	return values[last]
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
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

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
