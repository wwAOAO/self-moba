package berserker

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"strconv"
)

const (
	heroID    = "berserker"
	passiveID = "berserker_passive"
	qID       = "berserker_q"
	wID       = "berserker_w"
	eID       = "berserker_e"
	rID       = "berserker_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		TickEntity:            Tick,
		OnBasicHit:            ApplyWOnBasicHit,
		OnDamage:              ApplyBleedOnDamage,
		OnKill:                OnKill,
		SpecialRecast:         SpecialRecast,
		ReleasePreparedR:      ReleasePreparedR,
		CancelPreparedR:       CancelPreparedR,
		ActiveBuffs:           ActiveBuffs,
		ApplyStats:            ApplyStats,
		BasicAttackMultiplier: WAttackMultiplier,
	})
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || state.Stacks > 0 || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{30, 35, 40, 45, 50})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.75), tickRate)
	entity.Skills[qID] = state
	showQRange(w, entity, skill, tick, state.StacksExpireTick)
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func showQRange(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, expiresAt uint64) {
	id := w.NextEffectID("effect:berserker_q:")
	w.PutSkillEffect(world.SkillEffect{
		ID:           id,
		Kind:         "berserker_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Range:        skillRange(skill, 425),
		Radius:       skillMeta(skill, "innerRadius", 300),
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func releaseQ(w *world.World, entity *world.Entity, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	outerRadius := skillRange(skill, 425)
	innerRadius := skillMeta(skill, "innerRadius", 300)
	healHits := 0
	for _, target := range w.TargetsInRadius(entity, entity.Position, outerRadius) {
		targetDistance := distance(entity.Position, target.Position)
		if targetDistance > outerRadius {
			continue
		}
		outer := targetDistance > innerRadius
		damage := w.PhysicalDamageAfterResistance(entity, target, qDamage(entity, skill, level, outer), tick)
		target.Combat.LastHitTick = tick
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
			continue
		}
		wasAlive := target.Stats.HP > 0
		if outer {
			w.ApplyAOEDamage(entity, target, damage, "physical", tickRate)
		} else {
			w.ApplyAOEDamageWithoutBerserkerBleed(entity, target, damage, "physical", tickRate)
		}
		if outer && (world.IsHeroUnit(target) || world.IsMonster(target)) {
			healHits++
		}
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	healQ(w, entity, skill, healHits)
}

func qDamage(entity *world.Entity, skill config.SkillConfig, level int, outer bool) float64 {
	if outer {
		return skillList(skill, "outerBaseDamage", level, []float64{50, 80, 110, 140, 170}) +
			entity.Stats.Attack*skillList(skill, "outerTotalAdRatio", level, []float64{1, 1.1, 1.2, 1.3, 1.4})
	}
	return skillList(skill, "innerBaseDamage", level, []float64{17.5, 28, 38.5, 49, 59.5}) +
		entity.Stats.Attack*skillList(skill, "innerTotalAdRatio", level, []float64{0.35, 0.385, 0.42, 0.455, 0.49})
}

func healQ(w *world.World, entity *world.Entity, skill config.SkillConfig, hits int) {
	if entity == nil || hits <= 0 || entity.Stats.HP <= 0 || entity.Stats.HP >= entity.Stats.MaxHP {
		return
	}
	ratio := math.Min(skillMeta(skill, "outerHealMissingHPCap", 0.48), float64(hits)*skillMeta(skill, "outerHealMissingHPRatio", 0.12))
	if ratio <= 0 {
		return
	}
	beforeHP := entity.Stats.HP
	entity.Stats.HP += int(math.Floor(float64(entity.Stats.MaxHP-entity.Stats.HP) * ratio))
	if entity.Stats.HP > entity.Stats.MaxHP {
		entity.Stats.HP = entity.Stats.MaxHP
	}
	w.RefreshStatsAfterHPChange(entity, beforeHP)
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || state.Stacks > 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 30)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.Stacks = 1
	entity.Combat.NextAttackTick = tick
	entity.Skills[wID] = state
}

func WAttackMultiplier(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != heroID {
		return 1
	}
	state := attacker.Skills[wID]
	if state.Stacks <= 0 || state.Level <= 0 {
		return 1
	}
	return skillList(w.SkillConfig(wID), "damageMultiplier", state.Level, []float64{1.2, 1.4, 1.6, 1.8, 2})
}

func ApplyWOnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.HeroID != heroID {
		return
	}
	state := attacker.Skills[wID]
	if state.Stacks <= 0 || state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(wID)
	state.Stacks = 0
	state.StacksExpireTick = 0
	cooldownMS := skillList(skill, "cooldownMs", state.Level, []float64{9000, 8000, 7000, 6000, 5000})
	cooldownMS -= float64(bleedStacks(target, attacker.ID)) * skillMeta(skill, "cooldownReductionPerBleedStackSeconds", 1) * 1000
	if cooldownMS < 0 {
		cooldownMS = 0
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(attacker, int(cooldownMS), tickRate)
	attacker.Skills[wID] = state
	until := tick + secondsToTicks(skillMeta(skill, "slowSeconds", 2), tickRate)
	slow := skillList(skill, "slow", state.Level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
	w.ApplyMoveSpeedSlow(target, slow, until)
	w.ApplyAttackSpeedSlow(target, slow, until)
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || state.Stacks > 0 || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 45)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	dx, dy := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	entity.Berserker.ApprehendDir = world.Vector2{X: dx, Y: dy}
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	entity.Control.ActionLockedUntilTick = state.StacksExpireTick
	entity.Skills[eID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseE(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	dir := entity.Berserker.ApprehendDir
	if dir.X == 0 && dir.Y == 0 {
		dir.X = 1
	}
	until := tick + secondsToTicks(skillMeta(skill, "slowSeconds", 1), tickRate)
	for _, target := range w.TargetsInCone(entity, dir, skillRange(skill, 535), skillMeta(skill, "coneAngleDegrees", 50)) {
		if !canPullE(target) {
			continue
		}
		target.Position = w.ClampWorldPoint(world.Vector2{
			X: entity.Position.X + dir.X*(entity.Radius+target.Radius+1),
			Y: entity.Position.Y + dir.Y*(entity.Radius+target.Radius+1),
		})
		target.Control.AirborneUntilTick = tick + world.ControlTicksAfterTenacity(target, secondsToTicks(skillMeta(skill, "knockupSeconds", 0.25), tickRate), tick)
		w.ApplyMoveSpeedSlow(target, skillMeta(skill, "slow", 0.4), until)
	}
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || state.Stacks > 0 || tickRate <= 0 {
		return
	}
	state = expireRRecast(entity, state, tick)
	freeRecast := entity.Berserker.NoxianGuillotineRecast > tick
	if tick < state.CooldownUntilTick && !freeRecast {
		entity.Skills[rID] = state
		return
	}
	cost := rManaCost(skill, state.Level)
	if !freeRecast && entity.Stats.MP < cost {
		return
	}
	target := rTarget(w, entity, cast, skill)
	if target == nil {
		return
	}
	if !rInRange(entity, target, skill) {
		prepareR(w, entity, target, state, skill)
		entity.Skills[rID] = state
		return
	}
	startRWindup(w, entity, target, state, skill, freeRecast, cost, tick, tickRate)
}

func prepareR(w *world.World, entity *world.Entity, target *world.Entity, state world.SkillState, skill config.SkillConfig) {
	if entity == nil || target == nil {
		return
	}
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		return
	}
	castPosition := w.ClampWorldPoint(world.Vector2{
		X: target.Position.X - dx*skillRange(skill, 460),
		Y: target.Position.Y - dy*skillRange(skill, 460),
	})
	entity.Berserker.NoxianGuillotineCastPending = true
	entity.Berserker.NoxianGuillotineCastTarget = target.ID
	entity.Berserker.NoxianGuillotineLevel = state.Level
	entity.Intent.MoveTarget = &castPosition
	entity.Intent.AttackTargetID = ""
	entity.Intent.AttackPausedTill = 0
}

func startRWindup(w *world.World, entity *world.Entity, target *world.Entity, state world.SkillState, skill config.SkillConfig, freeRecast bool, cost float64, tick uint64, tickRate int) {
	if entity == nil || target == nil {
		return
	}
	if !freeRecast {
		if entity.Stats.MP < cost {
			return
		}
		entity.Stats.MP -= cost
	}
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.5), tickRate)
	if freeRecast {
		entity.Berserker.NoxianGuillotineRecast = 0
	}
	entity.Berserker.NoxianGuillotineCastPending = false
	entity.Berserker.NoxianGuillotineCastTarget = ""
	entity.Berserker.NoxianGuillotineTarget = target.ID
	entity.Berserker.NoxianGuillotineLevel = state.Level
	entity.Control.ActionLockedUntilTick = state.StacksExpireTick
	entity.Skills[rID] = state
	showRRange(w, entity, target, skill, tick, state.StacksExpireTick)
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func SpecialRecast(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || cast.SkillID != rID || entity.Berserker.NoxianGuillotineRecast <= tick {
		return false
	}
	CastR(w, entity, cast, state, skill, tick, tickRate)
	return true
}

func ReleasePreparedR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Berserker.NoxianGuillotineCastPending {
		return
	}
	target := w.EntityByID(entity.Berserker.NoxianGuillotineCastTarget)
	state := expireRRecast(entity, entity.Skills[rID], tick)
	if state.Level <= 0 || !validRTarget(entity, target) {
		CancelPreparedR(entity)
		entity.Skills[rID] = state
		return
	}
	state.Level = entity.Berserker.NoxianGuillotineLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	skill := w.SkillConfig(rID)
	if !rInRange(entity, target, skill) {
		prepareR(w, entity, target, state, skill)
		entity.Skills[rID] = state
		return
	}
	freeRecast := entity.Berserker.NoxianGuillotineRecast > tick
	if tick < state.CooldownUntilTick && !freeRecast {
		CancelPreparedR(entity)
		entity.Skills[rID] = state
		return
	}
	startRWindup(w, entity, target, state, skill, freeRecast, rManaCost(skill, state.Level), tick, tickRate)
}

func CancelPreparedR(entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID || !entity.Berserker.NoxianGuillotineCastPending {
		return
	}
	entity.Berserker.NoxianGuillotineCastPending = false
	entity.Berserker.NoxianGuillotineCastTarget = ""
	entity.Berserker.NoxianGuillotineLevel = 0
}

func releaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	state := expireRRecast(entity, entity.Skills[rID], tick)
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		entity.Skills[rID] = state
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	target := w.EntityByID(entity.Berserker.NoxianGuillotineTarget)
	level := entity.Berserker.NoxianGuillotineLevel
	entity.Berserker.NoxianGuillotineTarget = ""
	entity.Berserker.NoxianGuillotineLevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead || !canAttackTarget(entity, target) {
		entity.Skills[rID] = state
		return
	}
	skill := w.SkillConfig(rID)
	normalCooldown := tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{120000, 100000, 80000})), tickRate)
	state.CooldownUntilTick = normalCooldown
	jumpToRTarget(w, entity, target)
	target.Combat.LastHitTick = tick
	wasAlive := target.Stats.HP > 0
	damage := rDamage(entity, target, skill, level)
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = w.TrueDamageAfterReduction(target, damage, tick)
		target.Combat.LastDamageType = "true"
	} else {
		w.ApplyTrueDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			refreshROnHeroKill(w, entity, target, skill, level, normalCooldown, tick, tickRate)
			state = entity.Skills[rID]
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	addRVision(w, entity, target.Position, skill, tick, tickRate)
	entity.Skills[rID] = state
}

func jumpToRTarget(w *world.World, entity *world.Entity, target *world.Entity) {
	if entity == nil || target == nil {
		return
	}
	dx, dy := normalize(entity.Position.X-target.Position.X, entity.Position.Y-target.Position.Y)
	if dx == 0 && dy == 0 {
		dx = -1
	}
	distance := entity.Stats.AttackRange
	if distance <= 0 {
		distance = 175
	}
	entity.Position = w.ClampWorldPoint(world.Vector2{
		X: target.Position.X + dx*distance,
		Y: target.Position.Y + dy*distance,
	})
}

func rManaCost(skill config.SkillConfig, level int) float64 {
	return skillList(skill, "manaCost", level, []float64{100, 100, 0})
}

func expireRRecast(entity *world.Entity, state world.SkillState, tick uint64) world.SkillState {
	if entity == nil || entity.Berserker.NoxianGuillotineRecast == 0 || tick <= entity.Berserker.NoxianGuillotineRecast {
		return state
	}
	entity.Berserker.NoxianGuillotineRecast = 0
	entity.Berserker.NoxianGuillotineRestore = 0
	return state
}

func rDamage(source *world.Entity, target *world.Entity, skill config.SkillConfig, level int) float64 {
	base := skillList(skill, "baseDamage", level, []float64{125, 250, 375})
	bonusAD := 0.0
	if source != nil {
		bonusAD = source.Stats.BonusAttack
	}
	multiplier := 1 + float64(bleedStacks(target, source.ID))*skillMeta(skill, "damagePerBleedStack", 0.2)
	return (base + bonusAD*skillMeta(skill, "bonusAdRatio", 0.75)) * multiplier
}

func refreshROnHeroKill(w *world.World, entity *world.Entity, target *world.Entity, skill config.SkillConfig, level int, normalCooldown uint64, tick uint64, tickRate int) {
	if entity == nil || target == nil || !world.IsHeroUnit(target) {
		return
	}
	activateBloodRage(w, entity, tick, tickRate)
	state := entity.Skills[rID]
	state.Stacks = 0
	state.StacksExpireTick = 0
	if level >= 3 {
		state.CooldownUntilTick = tick
		entity.Berserker.NoxianGuillotineRecast = 0
		entity.Berserker.NoxianGuillotineRestore = 0
		entity.Skills[rID] = state
		return
	}
	entity.Berserker.NoxianGuillotineRecast = tick + secondsToTicks(skillMeta(skill, "recastSeconds", 20), tickRate)
	entity.Berserker.NoxianGuillotineRestore = 0
	state.CooldownUntilTick = normalCooldown
	entity.Skills[rID] = state
}

func rTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) *world.Entity {
	targetPoint := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	pickPadding := skillMeta(skill, "targetPickPadding", 80)
	if target := w.EntityByID(cast.TargetID); validRTarget(entity, target) && distance(targetPoint, target.Position) <= target.Radius+pickPadding {
		return target
	}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !validRTarget(entity, target) {
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

func validRTarget(entity *world.Entity, target *world.Entity) bool {
	return canAttackTarget(entity, target) && world.IsHeroUnit(target)
}

func rInRange(entity *world.Entity, target *world.Entity, skill config.SkillConfig) bool {
	return entity != nil && target != nil && distance(entity.Position, target.Position) <= skillRange(skill, 460)+target.Radius
}

func showRRange(w *world.World, entity *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, expiresAt uint64) {
	if entity == nil || target == nil {
		return
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:berserker_r:"),
		Kind:         "berserker_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		End:          target.Position,
		Range:        skillRange(skill, 460),
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func addRVision(w *world.World, entity *world.Entity, center world.Vector2, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil {
		return
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:berserker_r_vision:"),
		Kind:         "berserker_r_vision",
		Team:         entity.Team,
		SourceHeroID: entity.HeroID,
		Start:        center,
		Radius:       skillMeta(skill, "visionRadius", 450),
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "visionSeconds", 2.5), tickRate),
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil {
		return
	}
	if entity.HeroID == heroID {
		releaseQIfReady(w, entity, tick, tickRate)
		releaseEIfReady(w, entity, tick, tickRate)
		releaseR(w, entity, tick, tickRate)
		if entity.Berserker.BloodRageUntil > 0 && tick >= entity.Berserker.BloodRageUntil {
			entity.Berserker.BloodRageUntil = 0
			w.RefreshPlayerStats(entity)
		}
	}
	tickBleeds(w, entity, tick, tickRate)
}

func releaseQIfReady(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	state := entity.Skills[qID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		entity.Skills[qID] = state
		return
	}
	skill := w.SkillConfig(qID)
	releaseQ(w, entity, state.Level, skill, tick, tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{9000, 8000, 7000, 6000, 5000})), tickRate)
	entity.Skills[qID] = state
}

func releaseEIfReady(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	state := entity.Skills[eID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		entity.Skills[eID] = state
		return
	}
	skill := w.SkillConfig(eID)
	releaseE(w, entity, skill, tick, tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{24000, 21000, 18000, 15000, 12000})), tickRate)
	entity.Skills[eID] = state
}

func ApplyBleedOnDamage(w *world.World, source *world.Entity, target *world.Entity, basicAttack bool, pet bool, skipBleed bool, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID || pet || skipBleed || !actualEnemy(source, target) || tickRate <= 0 {
		return
	}
	applyBleedStack(w, source, target, tick, tickRate, source.Berserker.BloodRageUntil > tick)
}

func applyBleedStack(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int, fullStacks bool) {
	if source == nil || target == nil || source.HeroID != heroID || !actualEnemy(source, target) || tickRate <= 0 {
		return
	}
	skill := w.HeroPassiveSkill(source)
	maxStacks := int(skillMeta(skill, "maxBleedStacks", 5))
	if maxStacks <= 0 {
		return
	}
	duration, tickSeconds := bleedTiming(skill)
	if target.Passive.Bleeds == nil {
		target.Passive.Bleeds = map[string]world.BleedState{}
	}
	bleed := target.Passive.Bleeds[source.ID]
	oldStacks := bleed.Stacks
	if fullStacks {
		bleed.Stacks = maxStacks
	} else if bleed.Stacks < maxStacks {
		bleed.Stacks++
	}
	bleed.ExpiresAtTick = tick + secondsToTicks(duration, tickRate)
	if bleed.NextTick == 0 {
		bleed.NextTick = tick + secondsToTicks(tickSeconds, tickRate)
	}
	target.Passive.Bleeds[source.ID] = bleed
	if oldStacks < maxStacks && bleed.Stacks >= maxStacks && world.IsHeroUnit(target) {
		activateBloodRage(w, source, tick, tickRate)
	}
}

func tickBleeds(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if len(entity.Passive.Bleeds) == 0 || tickRate <= 0 {
		return
	}
	for sourceID, bleed := range entity.Passive.Bleeds {
		source := w.EntityByID(sourceID)
		if source == nil || entity.Stats.HP <= 0 || tick >= bleed.ExpiresAtTick {
			delete(entity.Passive.Bleeds, sourceID)
			continue
		}
		if tick < bleed.NextTick || bleed.Stacks <= 0 {
			continue
		}
		skill := w.HeroPassiveSkill(source)
		duration, tickSeconds := bleedTiming(skill)
		rawDamage := bleedRawDamage(source, skill) * float64(bleed.Stacks)
		if world.IsMonster(entity) {
			rawDamage *= skillMeta(skill, "bleedMonsterDamageMultiplier", 2.5)
		}
		value := bleed.Remainder + rawDamage*tickSeconds/duration
		damage := int(math.Floor(value + 0.000000001))
		bleed.Remainder = value - float64(damage)
		bleed.NextTick += secondsToTicks(tickSeconds, tickRate)
		entity.Passive.Bleeds[sourceID] = bleed
		if damage <= 0 {
			continue
		}
		entity.Combat.LastHitTick = tick
		wasAlive := entity.Stats.HP > 0
		w.ApplyPetDamage(source, entity, w.PhysicalDamageAfterResistance(source, entity, float64(damage), tick), "physical", tickRate)
		if wasAlive && entity.Stats.HP == 0 {
			w.ApplyKillReward(source, entity)
			w.KillPlayer(entity, tick, tickRate)
			w.RemoveDeadUnit(entity)
		}
	}
}

func bleedRawDamage(source *world.Entity, skill config.SkillConfig) float64 {
	level := world.MinHeroLevel
	bonusAD := 0.0
	if source != nil {
		level = source.Level
		bonusAD = source.Stats.BonusAttack
	}
	return skillCurve(skill, "bleedDamage", "bleedDamageLevels", level, 13) + bonusAD*skillMeta(skill, "bleedBonusAdRatio", 0.3)
}

func bleedTiming(skill config.SkillConfig) (float64, float64) {
	duration := skillMeta(skill, "bleedDurationSeconds", 5)
	if duration <= 0 {
		duration = 5
	}
	tickSeconds := skillMeta(skill, "bleedTickSeconds", 1)
	if tickSeconds <= 0 {
		tickSeconds = 1
	}
	return duration, tickSeconds
}

func activateBloodRage(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	if tickRate <= 0 {
		tickRate = 20
	}
	entity.Berserker.BloodRageUntil = tick + secondsToTicks(skillMeta(w.HeroPassiveSkill(entity), "bloodRageDurationSeconds", 5), tickRate)
	w.RefreshPlayerStats(entity)
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || stats == nil || entity.HeroID != heroID {
		return
	}
	if entity.Berserker.BloodRageUntil > 0 {
		bonus := skillCurve(w.HeroPassiveSkill(entity), "bloodRageAttack", "bloodRageAttackLevels", entity.Level, 30)
		stats.Attack += bonus
		stats.BonusAttack += bonus
	}
	state := entity.Skills[eID]
	if state.Level > 0 {
		stats.PhysicalPenPercent += skillList(w.SkillConfig(eID), "armorPenPercent", state.Level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
	}
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	if killer == nil || target == nil || killer.HeroID != heroID {
		return
	}
	if world.IsHeroUnit(target) {
		activateBloodRage(w, killer, target.Combat.LastHitTick, target.Death.RespawnTickRate)
	}
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil {
		return nil
	}
	buffs := make([]world.BuffState, 0, len(entity.Passive.Bleeds)+2)
	if entity.Berserker.BloodRageUntil > tick {
		buffs = append(buffs, world.BuffState{ID: "berserker_blood_rage", Name: "血怒", ExpiresAtTick: entity.Berserker.BloodRageUntil})
	}
	if entity.Berserker.NoxianGuillotineRecast > tick {
		buffs = append(buffs, world.BuffState{ID: "berserker_noxian_guillotine_recast", Name: "断头台再释放", ExpiresAtTick: entity.Berserker.NoxianGuillotineRecast})
	}
	for sourceID, bleed := range entity.Passive.Bleeds {
		if bleed.Stacks <= 0 || tick >= bleed.ExpiresAtTick {
			continue
		}
		buffs = append(buffs, world.BuffState{
			ID:            "berserker_bleed:" + sourceID,
			Name:          "出血" + strconv.Itoa(bleed.Stacks) + "层",
			ExpiresAtTick: bleed.ExpiresAtTick,
			Negative:      true,
		})
	}
	return buffs
}

func canPullE(target *world.Entity) bool {
	return target != nil && target.Kind != world.EntityKindBaronNashor && (world.IsHeroUnit(target) || world.IsMinion(target) || world.IsMonster(target) || target.Kind == world.EntityKindDummy)
}

func actualEnemy(source *world.Entity, target *world.Entity) bool {
	return source.Team != target.Team && target.Stats.HP > 0
}

func bleedStacks(target *world.Entity, sourceID string) int {
	if target == nil || target.Passive.Bleeds == nil {
		return 0
	}
	return target.Passive.Bleeds[sourceID].Stacks
}

func canAttackTarget(attacker *world.Entity, target *world.Entity) bool {
	if attacker == nil || target == nil || target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == world.EntityKindFountain {
		return false
	}
	if target.Kind == world.EntityKindPlayer && target.Death.Dead {
		return false
	}
	return target.ID != attacker.ID && target.Team != attacker.Team
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

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
