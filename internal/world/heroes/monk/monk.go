package monk

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID    = "monk"
	passiveID = "monk_passive"
	qID       = "monk_q"
	wID       = "monk_w"
	eID       = "monk_e"
	rID       = "monk_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		OnBasicAttackRelease:  OnBasicAttackRelease,
		AttackSpeedMultiplier: FlurryAttackSpeedMultiplier,
		ActiveBuffs:           ActiveBuffs,
		Tick:                  Tick,
		TickEntity:            TickEntity,
		SpecialRecast:         SpecialRecast,
		ApplyStats:            ApplyStats,
		MonkQDamage:           QDamage,
		MonkQHit:              QHit,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID {
		return
	}
	resolveEchoStrike(w, entity, tick, tickRate)
	releaseQ(w, entity, tick, tickRate)
	releaseE(w, entity, tick, tickRate)
	releaseR(w, entity, tick, tickRate)
	tickESlows(w, entity, tick)
	if entity.Passive.MonkQMarkUntil > 0 && tick >= entity.Passive.MonkQMarkUntil {
		clearQMark(w, entity)
	}
	if entity.Passive.MonkWRecastUntil > 0 && tick >= entity.Passive.MonkWRecastUntil {
		entity.Passive.MonkWRecastUntil = 0
	}
	if entity.Passive.MonkWIronWillUntil > 0 && tick >= entity.Passive.MonkWIronWillUntil {
		entity.Passive.MonkWIronWillUntil = 0
		entity.Passive.MonkWIronWillLevel = 0
		w.RefreshPlayerStats(entity)
	}
	if entity.Passive.MonkERecastUntil > 0 && tick >= entity.Passive.MonkERecastUntil {
		entity.Passive.MonkERecastUntil = 0
		entity.Passive.MonkEHitIDs = nil
	}
}

func TickEntity(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || entity.Control.DashUntilTick <= tick || tickRate <= 0 {
		return
	}
	var echo *world.SkillEffect
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "monk_q_echo" && effect.SourceID == entity.ID && effect.ExpiresAt == entity.Control.DashUntilTick+1 {
			effectCopy := effect
			echo = &effectCopy
			break
		}
	}
	if echo == nil {
		return
	}
	target := w.EntityByID(echo.TargetID)
	if !world.CanAttackTarget(entity, target) {
		return
	}
	end := echoEndPosition(entity, target)
	dashDistance := distance(entity.Position, end)
	dashSpeed := skillMeta(w.SkillConfig(qID), "echoDashSpeed", 1400)
	if dashSpeed <= 0 {
		dashSpeed = 1400
	}
	dashTicks := secondsToTicks(dashDistance/dashSpeed, tickRate)
	if dashTicks < 1 {
		dashTicks = 1
	}
	startTick := tick
	if tick > 0 {
		startTick = tick - 1
	}
	entity.Control.DashStartTick = startTick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = end
	entity.Control.DashUntilTick = tick + dashTicks - 1
	entity.Control.ActionLockedUntilTick = entity.Control.DashUntilTick
	echo.End = end
	echo.ExpiresAt = entity.Control.DashUntilTick + 1
	echo.Speed = dashDistance / float64(dashTicks)
	w.PutSkillEffect(*echo)
}

func SpecialRecast(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || entity.HeroID != heroID {
		return false
	}
	if cast.SkillID == qID && qMarkActive(entity, tick) {
		castEcho(w, entity, tick, tickRate)
		return true
	}
	if cast.SkillID == wID && entity.Passive.MonkWRecastUntil > 0 && tick < entity.Passive.MonkWRecastUntil {
		castIronWill(w, entity, state, skill, tick, tickRate)
		return true
	}
	if cast.SkillID == eID && entity.Passive.MonkERecastUntil > 0 && tick < entity.Passive.MonkERecastUntil {
		castCripple(w, entity, state, skill, tick, tickRate)
		return true
	}
	return false
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.MonkQPending || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "energyCost", 50)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{10000, 9000, 8000, 7000, 6000})), tickRate)
	entity.Skills[qID] = state
	activateFlurry(w, entity, tick, tickRate)

	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.MonkQPending = true
	entity.Passive.MonkQRelease = tick + windupTicks
	entity.Passive.MonkQTarget = qTargetPoint(entity, cast, skill)
	entity.Passive.MonkQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.MonkQRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastSkill(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	if state.SkillID == "" {
		state.SkillID = skill.SkillID
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skill.CooldownMS, tickRate)
	entity.Skills[state.SkillID] = state
	activateFlurry(w, entity, tick, tickRate)
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	target := safeguardTarget(w, entity, cast, skillRange(skill, 700))
	if target == nil {
		return
	}
	cost := skillMeta(skill, "energyCost", 50)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	start := entity.Position
	end := start
	if target.ID != entity.ID {
		end = safeguardEndPosition(entity, target)
	}
	dashDistance := distance(start, end)
	dashTicks := uint64(0)
	if dashDistance > 0 {
		dashSpeed := skillMeta(skill, "safeguardDashSpeed", 1400)
		if dashSpeed <= 0 {
			dashSpeed = 1400
		}
		dashTicks = secondsToTicks(dashDistance/dashSpeed, tickRate)
		if dashTicks < 1 {
			dashTicks = 1
		}
		entity.Control.DashStartTick = tick
		entity.Control.DashStart = start
		entity.Control.DashEnd = end
		entity.Control.DashUntilTick = tick + dashTicks
	}
	entity.Intent = world.IntentState{}
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0

	shield := shieldValue(entity, skill, state.Level)
	shieldUntil := tick + secondsToTicks(skillMeta(skill, "shieldSeconds", 2), tickRate)
	w.AddMageShieldLayer(entity, shield, shieldUntil)
	cooldownMs := int(skillList(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000}))
	if target.Kind == world.EntityKindPlayer && target.Team == entity.Team {
		if target.ID != entity.ID {
			w.AddMageShieldLayer(target, shield, shieldUntil)
		}
		cooldownMs /= 2
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, cooldownMs, tickRate)
	entity.Passive.MonkWRecastUntil = tick + secondsToTicks(skillMeta(skill, "recastSeconds", 3), tickRate)
	entity.Skills[wID] = state
	effectTicks := dashTicks
	if effectTicks < 1 {
		effectTicks = secondsToTicks(skillMeta(skill, "safeguardEffectSeconds", 0.45), tickRate)
	}
	effectSpeed := float64(0)
	if dashTicks > 0 {
		effectSpeed = dashDistance / float64(dashTicks)
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:monk_w_safeguard:"),
		Kind:         "monk_w_safeguard",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		TargetID:     target.ID,
		Start:        start,
		End:          end,
		Radius:       skillMeta(skill, "safeguardEffectRadius", 85),
		Speed:        effectSpeed,
		CreatedAt:    tick,
		ExpiresAt:    tick + effectTicks,
	})
	activateFlurry(w, entity, tick, tickRate)
}

func CastE(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.MonkEPending || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "energyCost", 50)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{10000, 10000, 10000, 10000, 10000})), tickRate)
	entity.Skills[eID] = state
	activateFlurry(w, entity, tick, tickRate)

	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.MonkEPending = true
	entity.Passive.MonkERelease = tick + windupTicks
	entity.Passive.MonkELevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.MonkERelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.MonkRPending || tickRate <= 0 {
		return
	}
	target := dragonKickTarget(w, entity, cast, skillRange(skill, 375))
	if target == nil {
		return
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{110000, 85000, 60000})), tickRate)
	entity.Skills[rID] = state
	activateFlurry(w, entity, tick, tickRate)

	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.MonkRPending = true
	entity.Passive.MonkRRelease = tick + windupTicks
	entity.Passive.MonkRTargetID = target.ID
	entity.Passive.MonkRLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.MonkRRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || !entity.Passive.MonkQPending || tick < entity.Passive.MonkQRelease || tickRate <= 0 {
		return
	}
	target := entity.Passive.MonkQTarget
	level := entity.Passive.MonkQLevel
	entity.Passive.MonkQPending = false
	entity.Passive.MonkQRelease = 0
	entity.Passive.MonkQTarget = world.Vector2{}
	entity.Passive.MonkQLevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	qRange := skillRange(skill, 1100)
	speed := skillMeta(skill, "projectileSpeed", 1400)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:monk_q:"),
		Kind:         "monk_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileWidth", 70) / 2,
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func releaseE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || !entity.Passive.MonkEPending || tick < entity.Passive.MonkERelease || tickRate <= 0 {
		return
	}
	level := entity.Passive.MonkELevel
	entity.Passive.MonkEPending = false
	entity.Passive.MonkERelease = 0
	entity.Passive.MonkELevel = 0
	entity.Passive.MonkEHitIDs = nil
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	skill := w.SkillConfig(eID)
	radius := skillMeta(skill, "radius", skillRange(skill, 350))
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:monk_e_tempest:"),
		Kind:         "monk_e_tempest",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "tempestEffectSeconds", 0.5), tickRate),
	})
	hits := w.TargetsInRadius(entity, entity.Position, radius)
	if len(hits) == 0 {
		return
	}
	entity.Passive.MonkEHitIDs = make(map[string]bool, len(hits))
	entity.Passive.MonkERecastUntil = tick + secondsToTicks(skillMeta(skill, "recastSeconds", 3), tickRate)
	revealUntil := tick + secondsToTicks(skillMeta(skill, "revealSeconds", 4), tickRate)
	for _, target := range hits {
		entity.Passive.MonkEHitIDs[target.ID] = true
		damage := eDamage(w, entity, target, skill, level, tick)
		wasAlive := target.Stats.HP > 0
		target.Combat.LastHitTick = tick
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
		} else {
			w.ApplyMagicDamage(entity, target, damage, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		}
		w.PutSkillEffect(world.SkillEffect{
			ID:           w.NextEffectID("effect:monk_e_reveal:"),
			Kind:         "monk_e_reveal",
			Team:         entity.Team,
			SourceID:     entity.ID,
			SourceHeroID: entity.HeroID,
			Start:        target.Position,
			Radius:       target.Radius,
			CreatedAt:    tick,
			ExpiresAt:    revealUntil,
		})
	}
}

func eDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	raw := skillList(skill, "baseDamage", level, []float64{60, 95, 130, 165, 200}) + attacker.Stats.BonusAttack*skillMeta(skill, "bonusAdRatio", 1)
	return w.MagicDamageAfterResistance(attacker, target, raw, tick)
}

func releaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || !entity.Passive.MonkRPending || tick < entity.Passive.MonkRRelease || tickRate <= 0 {
		return
	}
	targetID := entity.Passive.MonkRTargetID
	level := entity.Passive.MonkRLevel
	entity.Passive.MonkRPending = false
	entity.Passive.MonkRRelease = 0
	entity.Passive.MonkRTargetID = ""
	entity.Passive.MonkRLevel = 0
	target := w.EntityByID(targetID)
	if entity.Stats.HP <= 0 || entity.Death.Dead || !world.CanAttackTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(rID)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	start := target.Position
	distance := skillMeta(skill, "knockbackDistance", 1200)
	end := w.ClampWorldPoint(world.Vector2{X: start.X + dx*distance, Y: start.Y + dy*distance})
	travelTicks := secondsToTicks(skillMeta(skill, "knockbackSeconds", 1), tickRate)
	if travelTicks < 1 {
		travelTicks = 1
	}
	until := tick + travelTicks
	target.Intent = world.IntentState{}
	target.Combat.PendingAttackTargetID = ""
	target.Combat.AttackReleaseTick = 0
	target.Control.DashStartTick = tick
	target.Control.DashStart = start
	target.Control.DashEnd = end
	target.Control.DashUntilTick = until
	target.Control.ActionLockedUntilTick = until
	w.ApplyAirborne(target, until, tick, tickRate)
	applyRDamage(w, entity, target, skill, level, 0, tick, tickRate)

	extraRaw := target.Stats.BonusHP * skillMeta(skill, "collisionBonusHPRatio", 0.12)
	for _, hit := range targetsAlongKickPath(w, entity, target, start, world.Vector2{X: dx, Y: dy}, distance) {
		applyRDamage(w, entity, hit, skill, level, extraRaw, tick, tickRate)
		w.ApplyAirborne(hit, tick+secondsToTicks(skillMeta(skill, "collisionAirborneSeconds", 1), tickRate), tick, tickRate)
	}
}

func applyRDamage(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, level int, extraRaw float64, tick uint64, tickRate int) {
	raw := skillList(skill, "baseDamage", level, []float64{175, 400, 625}) + source.Stats.BonusAttack*skillMeta(skill, "bonusAdRatio", 2) + extraRaw
	damage := w.PhysicalDamageAfterResistance(source, target, raw, tick)
	wasAlive := target.Stats.HP > 0
	target.Combat.LastHitTick = tick
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
		return
	}
	w.ApplyDamage(source, target, damage, tickRate)
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(source, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}

func targetsAlongKickPath(w *world.World, source *world.Entity, kicked *world.Entity, start world.Vector2, dir world.Vector2, kickDistance float64) []*world.Entity {
	hits := make([]*world.Entity, 0)
	w.ForEachEntity(func(target *world.Entity) {
		if target == nil || target.ID == kicked.ID || !world.CanAttackTarget(source, target) {
			return
		}
		along, perpendicular := projectPointOnDir(start, dir, target.Position)
		if along < 0 || along > kickDistance+target.Radius {
			return
		}
		if perpendicular <= kicked.Radius+target.Radius {
			hits = append(hits, target)
		}
	})
	return hits
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, echo bool, tick uint64) int {
	if attacker == nil || target == nil {
		return 0
	}
	raw := skillList(skill, "baseDamage", skillLevel, []float64{65, 95, 125, 155, 185}) + attacker.Stats.BonusAttack*skillMeta(skill, "bonusAdRatio", 0.95)
	if echo {
		missing := math.Max(0, target.Stats.MaxHP-target.Stats.HP)
		execute := missing * skillMeta(skill, "missingHPRatio", 0.08)
		if world.IsMonster(target) {
			execute = math.Min(execute, skillMeta(skill, "monsterExecuteDamageCap", 400))
		}
		raw += execute
	}
	return w.PhysicalDamageAfterResistance(attacker, target, raw, tick)
}

func QHit(w *world.World, source *world.Entity, target *world.Entity, projectile *world.Projectile, damage int, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil {
		return
	}
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
	} else {
		w.ApplyDamage(source, target, damage, tickRate)
	}
	markQTarget(w, source, target, projectile.Damage, tick, tickRate)
}

func markQTarget(w *world.World, source *world.Entity, target *world.Entity, level int, tick uint64, tickRate int) {
	clearQMark(w, source)
	expiresAt := tick + secondsToTicks(skillMeta(w.SkillConfig(qID), "markDurationSeconds", 3), tickRate)
	effectID := w.NextEffectID("effect:monk_q_mark:")
	source.Passive.MonkQMarkTargetID = target.ID
	source.Passive.MonkQMarkUntil = expiresAt
	source.Passive.MonkQMarkLevel = level
	source.Passive.MonkQMarkEffectID = effectID
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         "monk_q_mark",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		TargetID:     target.ID,
		Start:        target.Position,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func castIronWill(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "recastEnergyCost", 30)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	entity.Passive.MonkWRecastUntil = 0
	entity.Passive.MonkWIronWillUntil = tick + secondsToTicks(skillMeta(skill, "ironWillSeconds", 4), tickRate)
	entity.Passive.MonkWIronWillLevel = state.Level
	entity.Skills[wID] = state
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:monk_w_iron_will:"),
		Kind:         "monk_w_iron_will",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       skillMeta(skill, "ironWillEffectRadius", 75),
		CreatedAt:    tick,
		ExpiresAt:    entity.Passive.MonkWIronWillUntil,
	})
	activateFlurry(w, entity, tick, tickRate)
	w.RefreshPlayerStats(entity)
}

func castCripple(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || state.Level <= 0 || tickRate <= 0 || len(entity.Passive.MonkEHitIDs) == 0 {
		return
	}
	cost := skillMeta(skill, "recastEnergyCost", 30)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	entity.Passive.MonkERecastUntil = 0
	activateFlurry(w, entity, tick, tickRate)

	duration := secondsToTicks(skillMeta(skill, "slowSeconds", 4), tickRate)
	until := tick + duration
	slow := skillList(skill, "slow", state.Level, []float64{0.2, 0.3, 0.4, 0.5, 0.6})
	if entity.Passive.MonkESlows == nil {
		entity.Passive.MonkESlows = make(map[string]world.MonkESlowState)
	}
	for targetID := range entity.Passive.MonkEHitIDs {
		target := w.EntityByID(targetID)
		if !world.CanAttackTarget(entity, target) {
			continue
		}
		applied := slow * (1 - target.Stats.SlowResist)
		if applied < 0 {
			applied = 0
		}
		if applied > 1 {
			applied = 1
		}
		w.ApplyMoveSpeedSlow(target, slow, until)
		entity.Passive.MonkESlows[targetID] = world.MonkESlowState{StartTick: tick, Until: until, Slow: applied}
		w.PutSkillEffect(world.SkillEffect{
			ID:           w.NextEffectID("effect:monk_e_cripple:"),
			Kind:         "monk_e_cripple",
			Team:         entity.Team,
			SourceID:     entity.ID,
			SourceHeroID: entity.HeroID,
			TargetID:     target.ID,
			Start:        target.Position,
			End:          target.Position,
			Radius:       target.Radius + 55,
			CreatedAt:    tick,
			ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "crippleEffectSeconds", 0.65), tickRate),
		})
	}
	entity.Passive.MonkEHitIDs = nil
}

func tickESlows(w *world.World, entity *world.Entity, tick uint64) {
	if entity == nil || len(entity.Passive.MonkESlows) == 0 {
		return
	}
	for targetID, slow := range entity.Passive.MonkESlows {
		target := w.EntityByID(targetID)
		if target == nil || tick >= slow.Until || slow.Until <= slow.StartTick {
			if target != nil && target.Control.MoveSpeedSlowUntil == slow.Until {
				target.Control.MoveSpeedSlow = 0
				target.Control.MoveSpeedSlowUntil = 0
			}
			delete(entity.Passive.MonkESlows, targetID)
			continue
		}
		if target.Control.MoveSpeedSlowUntil != slow.Until {
			continue
		}
		remaining := float64(slow.Until-tick) / float64(slow.Until-slow.StartTick)
		target.Control.MoveSpeedSlow = slow.Slow * remaining
	}
}

func castEcho(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	target := w.EntityByID(entity.Passive.MonkQMarkTargetID)
	if !qMarkActive(entity, tick) || !world.CanAttackTarget(entity, target) || tickRate <= 0 {
		clearQMark(w, entity)
		return
	}
	level := entity.Passive.MonkQMarkLevel
	clearQMark(w, entity)
	activateFlurry(w, entity, tick, tickRate)

	end := echoEndPosition(entity, target)
	dashDistance := distance(entity.Position, end)
	dashSpeed := skillMeta(w.SkillConfig(qID), "echoDashSpeed", 1400)
	if dashSpeed <= 0 {
		dashSpeed = 1400
	}
	dashTicks := secondsToTicks(dashDistance/dashSpeed, tickRate)
	if dashTicks < 1 {
		dashTicks = 1
	}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = end
	entity.Control.DashUntilTick = tick + dashTicks
	entity.Control.ActionLockedUntilTick = entity.Control.DashUntilTick
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:monk_q_echo:"),
		Kind:         "monk_q_echo",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		TargetID:     target.ID,
		Start:        entity.Position,
		End:          end,
		Radius:       skillMeta(w.SkillConfig(qID), "echoEffectRadius", 55),
		Count:        level,
		Speed:        dashDistance / float64(dashTicks),
		CreatedAt:    tick,
		ExpiresAt:    entity.Control.DashUntilTick + 1,
	})
}

func resolveEchoStrike(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity.Control.DashUntilTick != 0 || tickRate <= 0 {
		return
	}
	for _, effect := range w.SkillEffects() {
		if effect.Kind != "monk_q_echo" || effect.SourceID != entity.ID || effect.ExpiresAt != tick+1 {
			continue
		}
		target := w.EntityByID(effect.TargetID)
		if !world.CanAttackTarget(entity, target) {
			return
		}
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		damage := QDamage(w, entity, target, w.SkillConfig(qID), effect.Count, true, tick)
		wasAlive := target.Stats.HP > 0
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
			return
		}
		w.ApplyDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
		return
	}
}

func qMarkActive(entity *world.Entity, tick uint64) bool {
	return entity != nil && entity.HeroID == heroID && entity.Passive.MonkQMarkTargetID != "" &&
		entity.Passive.MonkQMarkUntil > 0 && (tick == 0 || tick < entity.Passive.MonkQMarkUntil)
}

func clearQMark(w *world.World, entity *world.Entity) {
	if entity == nil {
		return
	}
	if entity.Passive.MonkQMarkEffectID != "" {
		w.RemoveSkillEffect(entity.Passive.MonkQMarkEffectID)
	}
	entity.Passive.MonkQMarkTargetID = ""
	entity.Passive.MonkQMarkUntil = 0
	entity.Passive.MonkQMarkLevel = 0
	entity.Passive.MonkQMarkEffectID = ""
}

func activateFlurry(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	skill := w.SkillConfig(passiveID)
	entity.Passive.MonkFlurryUntil = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 3), tickRate)
	entity.Passive.MonkFlurryAttacks = int(skillMeta(skill, "attackCount", 2))
	entity.Passive.MonkFlurryHitIndex = 0
}

func OnBasicAttackRelease(w *world.World, source *world.Entity, _ *world.Entity, tick uint64, tickRate int) {
	if w == nil || source == nil || source.HeroID != heroID || !flurryActive(source, tick) || tickRate <= 0 {
		return
	}
	source.Passive.MonkFlurryAttacks--
	source.Passive.MonkFlurryHitIndex++
	restoreEnergy(source, energyRefund(source.Level, source.Passive.MonkFlurryHitIndex))
	reduceBasicCooldowns(source, tick, secondsToTicks(skillMeta(w.SkillConfig(passiveID), "cooldownRefundSeconds", 0.5), tickRate))
	if source.Passive.MonkFlurryAttacks <= 0 {
		source.Passive.MonkFlurryUntil = 0
		source.Passive.MonkFlurryHitIndex = 0
	}
}

func FlurryAttackSpeedMultiplier(entity *world.Entity, tick uint64) float64 {
	if !flurryActive(entity, tick) {
		return 1
	}
	return 1.4
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID {
		return nil
	}
	buffs := make([]world.BuffState, 0, 3)
	if flurryActive(entity, tick) {
		buffs = append(buffs, world.BuffState{
			ID:            passiveID,
			Name:          "疾风骤雨",
			ExpiresAtTick: entity.Passive.MonkFlurryUntil,
		})
	}
	if entity.Passive.MonkWRecastUntil > tick {
		buffs = append(buffs, world.BuffState{
			ID:            "monk_w_recast",
			Name:          "铁布衫再释放",
			ExpiresAtTick: entity.Passive.MonkWRecastUntil,
		})
	}
	if entity.Passive.MonkWIronWillUntil > tick {
		buffs = append(buffs, world.BuffState{
			ID:            wID,
			Name:          "铁布衫",
			ExpiresAtTick: entity.Passive.MonkWIronWillUntil,
		})
	}
	return buffs
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if w == nil || entity == nil || stats == nil || entity.HeroID != heroID || entity.Passive.MonkWIronWillUntil == 0 {
		return
	}
	ratio := skillList(w.SkillConfig(wID), "ironWillPercent", entity.Passive.MonkWIronWillLevel, []float64{0.05, 0.1, 0.15, 0.2, 0.25})
	baseArmor := stats.PhysicalDefense - stats.BonusPhysicalDefense
	if baseArmor < 0 {
		baseArmor = 0
	}
	stats.PhysicalDefense += baseArmor * ratio
	stats.Omnivamp += ratio
}

func flurryActive(entity *world.Entity, tick uint64) bool {
	return entity != nil && entity.HeroID == heroID && entity.Passive.MonkFlurryAttacks > 0 &&
		entity.Passive.MonkFlurryUntil > 0 && (tick == 0 || tick < entity.Passive.MonkFlurryUntil)
}

func energyRefund(level int, hitIndex int) float64 {
	first, second := 20.0, 10.0
	if level >= 13 {
		first, second = 40, 20
	} else if level >= 7 {
		first, second = 30, 15
	}
	if hitIndex == 1 {
		return first
	}
	return second
}

func restoreEnergy(entity *world.Entity, value float64) {
	if entity == nil || value <= 0 {
		return
	}
	entity.Stats.MP = math.Min(entity.Stats.MaxMP, entity.Stats.MP+value)
}

func reduceBasicCooldowns(entity *world.Entity, tick uint64, refundTicks uint64) {
	if entity == nil || refundTicks == 0 {
		return
	}
	for _, skillID := range []string{qID, wID, eID} {
		state, ok := entity.Skills[skillID]
		if !ok || state.CooldownUntilTick <= tick {
			continue
		}
		if state.CooldownUntilTick <= tick+refundTicks {
			state.CooldownUntilTick = tick
		} else {
			state.CooldownUntilTick -= refundTicks
		}
		entity.Skills[skillID] = state
	}
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
	}
	if len(values) == 0 {
		return 0
	}
	if level < 1 {
		level = 1
	}
	if level > len(values) {
		level = len(values)
	}
	return values[level-1]
}

func safeguardTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, castRange float64) *world.Entity {
	if target := w.EntityByID(cast.TargetID); validSafeguardTarget(entity, target, castRange) {
		return target
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !validSafeguardTarget(entity, target, castRange) {
			return
		}
		dist := distance(point, target.Position)
		if dist > target.Radius+80 || dist >= bestDistance {
			return
		}
		best = target
		bestDistance = dist
	})
	return best
}

func dragonKickTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, castRange float64) *world.Entity {
	if target := w.EntityByID(cast.TargetID); world.CanAttackTarget(entity, target) && distance(entity.Position, target.Position) <= castRange+target.Radius {
		return target
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !world.CanAttackTarget(entity, target) || distance(entity.Position, target.Position) > castRange+target.Radius {
			return
		}
		dist := distance(point, target.Position)
		if dist > target.Radius+80 || dist >= bestDistance {
			return
		}
		best = target
		bestDistance = dist
	})
	return best
}

func validSafeguardTarget(entity *world.Entity, target *world.Entity, castRange float64) bool {
	if entity == nil || target == nil || target.Stats.HP <= 0 {
		return false
	}
	if distance(entity.Position, target.Position) > castRange+target.Radius {
		return false
	}
	switch target.Kind {
	case world.EntityKindPlayer:
		return target.Team == entity.Team && !target.Death.Dead
	case world.EntityKindMeleeMinion, world.EntityKindRangedMinion, world.EntityKindSiegeMinion, world.EntityKindSuperMinion, world.EntityKindWard:
		return target.Team == entity.Team
	case world.EntityKindFruit:
		return true
	default:
		return false
	}
}

func safeguardEndPosition(entity *world.Entity, target *world.Entity) world.Vector2 {
	dx, dy := normalize(entity.Position.X-target.Position.X, entity.Position.Y-target.Position.Y)
	if dx == 0 && dy == 0 {
		dx = -1
	}
	stopDistance := entity.Radius + target.Radius
	return world.Vector2{
		X: target.Position.X + dx*stopDistance,
		Y: target.Position.Y + dy*stopDistance,
	}
}

func shieldValue(entity *world.Entity, skill config.SkillConfig, level int) int {
	value := skillList(skill, "shieldValue", level, []float64{60, 100, 140, 180, 220})
	if entity != nil {
		value += float64(entity.Stats.AbilityPower) * skillMeta(skill, "apShieldRatio", 0.8)
	}
	return int(math.Round(value))
}

func qTargetPoint(entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) world.Vector2 {
	target := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	qRange := skillRange(skill, 1100)
	return world.Vector2{X: entity.Position.X + dx*qRange, Y: entity.Position.Y + dy*qRange}
}

func echoEndPosition(entity *world.Entity, target *world.Entity) world.Vector2 {
	dx, dy := normalize(entity.Position.X-target.Position.X, entity.Position.Y-target.Position.Y)
	if dx == 0 && dy == 0 {
		dx = -1
	}
	stopDistance := entity.Radius + target.Radius
	return world.Vector2{
		X: target.Position.X + dx*stopDistance,
		Y: target.Position.Y + dy*stopDistance,
	}
}

func normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func projectPointOnDir(origin world.Vector2, dir world.Vector2, point world.Vector2) (float64, float64) {
	vx := point.X - origin.X
	vy := point.Y - origin.Y
	along := vx*dir.X + vy*dir.Y
	px := origin.X + dir.X*along
	py := origin.Y + dir.Y*along
	return along, math.Hypot(point.X-px, point.Y-py)
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 || tickRate <= 0 {
		return 0
	}
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}
