package butcher

import (
	"fmt"
	"math"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"l-battle/internal/world/geom"
)

const (
	heroID    = "butcher"
	passiveID = "butcher_passive"
	qID       = "butcher_q"
	wID       = "butcher_w"
	eID       = "butcher_e"
	rID       = "butcher_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		OnKill:            OnKill,
		ActiveBuffs:       ActiveBuffs,
		ApplyStats:        ApplyStats,
		DamageBlock:       DamageBlock,
		Tick:              Tick,
		OnMoveInput:       stopR,
		SpecialRecast:     SpecialRecast,
		ResolveProjectile: ResolveProjectile,
	})
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 || entity.Passive.ButcherRUntil > tick || dismemberInterrupted(entity, tick) {
		return
	}
	target := w.EntityByID(cast.TargetID)
	if !canDismemberTarget(entity, target) || geom.Distance(entity.Position, target.Position) > skill.Range+entity.Radius+target.Radius {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{100, 130, 170})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{30000, 30000, 30000})), tickRate)
	entity.Skills[rID] = state
	duration := skillMeta(skill, "unitDurationSeconds", 5.9)
	if world.IsHeroUnit(target) {
		duration = skillMeta(skill, "heroDurationSeconds", 2.95)
	}
	until := tick + secondsToTicks(duration, tickRate)
	previousStun := target.Control.StunnedUntilTick
	appliedStun := maxUint64(previousStun, until)
	if !w.ApplyStun(target, appliedStun, tick, tickRate) {
		return
	}
	entity.Passive.ButcherRTargetID = target.ID
	entity.Passive.ButcherRStartPosition = entity.Position
	entity.Passive.ButcherRUntil = until
	entity.Passive.ButcherRNextTick = tick
	entity.Passive.ButcherRLevel = state.Level
	entity.Passive.ButcherRPreviousStunUntil = previousStun
	entity.Passive.ButcherRAppliedStunUntil = appliedStun
	entity.Passive.ButcherREffectID = w.NextEffectID("effect:butcher_r:")
	entity.Intent = world.IntentState{}
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0
	w.PutSkillEffect(world.SkillEffect{ID: entity.Passive.ButcherREffectID, Kind: "butcher_r", Team: entity.Team, SourceID: entity.ID, SourceHeroID: entity.HeroID, TargetID: target.ID, Start: entity.Position, End: target.Position, Radius: target.Radius, CreatedAt: tick, ExpiresAt: until})
	tickR(w, entity, tick, tickRate)
}

func CastE(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 || entity.Passive.ButcherRUntil > tick {
		return
	}
	cost := skillMeta(skill, "manaCost", 35)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	duration := skillList(skill, "durationSeconds", state.Level, []float64{4, 5, 6, 7, 8})
	entity.Passive.ButcherEUntil = tick + secondsToTicks(duration, tickRate)
	entity.Passive.ButcherELevel = state.Level
	entity.Passive.ButcherEEffectID = w.NextEffectID("effect:butcher_e:")
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{20000, 19000, 18000, 17000, 16000})), tickRate)
	entity.Skills[eID] = state
	w.PutSkillEffect(world.SkillEffect{ID: entity.Passive.ButcherEEffectID, Kind: "butcher_e", Team: entity.Team, SourceID: entity.ID, SourceHeroID: entity.HeroID, Start: entity.Position, Radius: entity.Radius, CreatedAt: tick, ExpiresAt: entity.Passive.ButcherEUntil})
}

func CastW(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.ButcherWActive || tickRate <= 0 || entity.Passive.ButcherRUntil > tick {
		return
	}
	interval := maxUint64(1, secondsToTicks(skillMeta(skill, "damageIntervalSeconds", 0.5), tickRate))
	entity.Passive.ButcherWActive = true
	entity.Passive.ButcherWNextTick = tick + interval
	entity.Passive.ButcherWLevel = state.Level
	entity.Passive.ButcherWEffectID = w.NextEffectID("effect:butcher_w:")
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, 500, tickRate)
	state.Stacks = 1
	entity.Skills[wID] = state
	w.PutSkillEffect(world.SkillEffect{ID: entity.Passive.ButcherWEffectID, Kind: "butcher_w", Team: entity.Team, SourceID: entity.ID, SourceHeroID: entity.HeroID, Start: entity.Position, Radius: skill.Range, CreatedAt: tick, ExpiresAt: ^uint64(0)})
}

func SpecialRecast(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, _ config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || entity.HeroID != heroID || cast.SkillID != wID || !entity.Passive.ButcherWActive {
		return false
	}
	if tick >= state.CooldownUntilTick {
		deactivateW(w, entity, tick, tickRate, true)
	}
	return true
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.ButcherQPending || tickRate <= 0 || entity.Passive.ButcherRUntil > tick {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{125, 130, 135, 140, 155})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{21000, 19000, 17000, 15000, 13000})), tickRate)
	entity.Skills[qID] = state
	target := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	if cast.TargetID != "" {
		if selected := w.EntityByID(cast.TargetID); selected != nil {
			target = selected.Position
		}
	}
	entity.Passive.ButcherQPending = true
	entity.Passive.ButcherQRelease = tick + maxUint64(1, secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.3), tickRate))
	entity.Passive.ButcherQTarget = w.ClampWorldPoint(target)
	entity.Passive.ButcherQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.ButcherQRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	tickR(w, entity, tick, tickRate)
	tickW(w, entity, tick, tickRate)
	if entity.Passive.ButcherEUntil > 0 && (entity.Death.Dead || entity.Stats.HP <= 0 || tick >= entity.Passive.ButcherEUntil) {
		deactivateE(w, entity)
	}
	if !entity.Passive.ButcherQPending || tick < entity.Passive.ButcherQRelease {
		return
	}
	target := entity.Passive.ButcherQTarget
	level := entity.Passive.ButcherQLevel
	entity.Passive.ButcherQPending = false
	entity.Passive.ButcherQRelease = 0
	entity.Passive.ButcherQTarget = world.Vector2{}
	entity.Passive.ButcherQLevel = 0
	if entity.Death.Dead || entity.Stats.HP <= 0 || tickRate <= 0 {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	rangeValue := skill.Range
	if rangeValue <= 0 {
		rangeValue = 1300
	}
	speed := skillMeta(skill, "projectileSpeed", 1450)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:butcher_q:"),
		Kind:         "butcher_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        rangeValue,
		Radius:       skillMeta(skill, "projectileRadius", 100),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(rangeValue/speed+0.1, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func tickR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.Passive.ButcherRUntil == 0 {
		return
	}
	target := w.EntityByID(entity.Passive.ButcherRTargetID)
	if target == nil || target.Stats.HP <= 0 || target.Death.Dead || entity.Stats.HP <= 0 || entity.Death.Dead || tick >= entity.Passive.ButcherRUntil || dismemberInterrupted(entity, tick) || geom.Distance(entity.Position, entity.Passive.ButcherRStartPosition) > 0.001 {
		stopR(w, entity, tick)
		return
	}
	entity.Intent = world.IntentState{}
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0
	target.Control.StunnedUntilTick = maxUint64(target.Control.StunnedUntilTick, entity.Passive.ButcherRAppliedStunUntil)
	if tick < entity.Passive.ButcherRNextTick {
		return
	}
	applyRDamage(w, entity, target, tick, tickRate)
	entity.Passive.ButcherRNextTick += maxUint64(1, secondsToTicks(skillMeta(w.SkillConfig(rID), "damageIntervalSeconds", 1), tickRate))
}

func applyRDamage(w *world.World, entity *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	skill := w.SkillConfig(rID)
	raw := skillList(skill, "baseDamagePerSecond", entity.Passive.ButcherRLevel, []float64{75, 125, 175}) + entity.Stats.BonusAttack*skillMeta(skill, "bonusAdRatioPerSecond", 0.75)
	damage := w.MagicDamageAfterResistance(entity, target, raw, tick)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	wasAlive := target.Stats.HP > 0
	w.ApplyMagicDamage(entity, target, damage, tickRate)
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(entity, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}

func stopR(w *world.World, entity *world.Entity, tick uint64) {
	if w == nil || entity == nil {
		return
	}
	target := w.EntityByID(entity.Passive.ButcherRTargetID)
	if target != nil && target.Control.StunnedUntilTick == entity.Passive.ButcherRAppliedStunUntil {
		target.Control.StunnedUntilTick = maxUint64(entity.Passive.ButcherRPreviousStunUntil, tick)
	}
	w.RemoveSkillEffect(entity.Passive.ButcherREffectID)
	entity.Passive.ButcherRTargetID = ""
	entity.Passive.ButcherRStartPosition = world.Vector2{}
	entity.Passive.ButcherRUntil = 0
	entity.Passive.ButcherRNextTick = 0
	entity.Passive.ButcherRLevel = 0
	entity.Passive.ButcherREffectID = ""
	entity.Passive.ButcherRPreviousStunUntil = 0
	entity.Passive.ButcherRAppliedStunUntil = 0
}

func canDismemberTarget(source *world.Entity, target *world.Entity) bool {
	return world.CanAttackTarget(source, target) && (world.IsHeroUnit(target) || world.IsMinion(target) || world.IsMonster(target) || target.Kind == world.EntityKindDummy)
}

func dismemberInterrupted(entity *world.Entity, tick uint64) bool {
	return tick < entity.Control.StunnedUntilTick || tick < entity.Control.AirborneUntilTick || tick < entity.Control.RootedUntilTick || tick < entity.Control.TauntedUntilTick || tick < entity.Control.SuppressedUntilTick || tick < entity.Control.SilencedUntilTick
}

func deactivateE(w *world.World, entity *world.Entity) {
	if entity == nil {
		return
	}
	if w != nil {
		w.RemoveSkillEffect(entity.Passive.ButcherEEffectID)
	}
	entity.Passive.ButcherEUntil = 0
	entity.Passive.ButcherELevel = 0
	entity.Passive.ButcherEEffectID = ""
}

func tickW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || !entity.Passive.ButcherWActive || tickRate <= 0 {
		return
	}
	if entity.Death.Dead || entity.Stats.HP <= 0 {
		deactivateW(w, entity, tick, tickRate, false)
		return
	}
	skill := w.SkillConfig(wID)
	interval := maxUint64(1, secondsToTicks(skillMeta(skill, "damageIntervalSeconds", 0.5), tickRate))
	intervalSeconds := float64(interval) / float64(tickRate)
	manaCost := skillMeta(skill, "manaPerSecond", 8) * intervalSeconds
	for entity.Passive.ButcherWNextTick > 0 && tick >= entity.Passive.ButcherWNextTick {
		if entity.Stats.MP < manaCost {
			deactivateW(w, entity, tick, tickRate, true)
			return
		}
		entity.Stats.MP -= manaCost
		applyWPeriodicDamage(w, entity, skill, entity.Passive.ButcherWLevel, entity.Passive.ButcherWNextTick, tickRate, intervalSeconds)
		entity.Passive.ButcherWNextTick += interval
	}
}

func applyWPeriodicDamage(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int, intervalSeconds float64) {
	raw := (skillList(skill, "damagePerSecond", level, []float64{30, 60, 90, 120, 150}) + entity.Stats.BonusAttack*skillMeta(skill, "bonusAdRatioPerSecond", 0.05)) * intervalSeconds
	duration := maxUint64(1, secondsToTicks(skillMeta(skill, "auraLingerSeconds", 0.5), tickRate))
	until := tick + duration
	slow := skillList(skill, "slow", level, []float64{0.14, 0.18, 0.22, 0.26, 0.3})
	for _, target := range w.TargetsInRadius(entity, entity.Position, skill.Range) {
		w.ApplyMoveSpeedSlow(target, slow, until)
		w.ApplyGrievousWounds(target, skillMeta(skill, "grievousWounds", 0.4), until)
		damage := w.MagicDamageAfterResistance(entity, target, raw, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyAOEDamage(entity, target, damage, "magic", tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	entity.Combat.LastHitTick = tick
	entity.Combat.DamageEvents = nil
	w.ApplyNonlethalMagicDamage(entity, raw, tick, tickRate)
}

func deactivateW(w *world.World, entity *world.Entity, tick uint64, tickRate int, startCooldown bool) {
	if entity == nil {
		return
	}
	if w != nil {
		w.RemoveSkillEffect(entity.Passive.ButcherWEffectID)
	}
	entity.Passive.ButcherWActive = false
	entity.Passive.ButcherWNextTick = 0
	entity.Passive.ButcherWLevel = 0
	entity.Passive.ButcherWEffectID = ""
	state := entity.Skills[wID]
	state.Stacks = 0
	if startCooldown {
		state.CooldownUntilTick = tick + cooldownTicksFor(entity, 500, tickRate)
	}
	entity.Skills[wID] = state
}

func ResolveProjectile(w *world.World, source *world.Entity, projectile *world.Projectile, previousPosition world.Vector2, tick uint64, tickRate int) bool {
	if projectile == nil || projectile.SkillID != qID {
		return false
	}
	hit := firstHookTarget(w, source, projectile, previousPosition)
	if hit == nil {
		return true
	}
	w.RemoveProjectile(projectile.ID)
	resolveHookHit(w, source, hit, projectile.Damage, tick, tickRate)
	return true
}

func firstHookTarget(w *world.World, source *world.Entity, projectile *world.Projectile, previousPosition world.Vector2) *world.Entity {
	if w == nil || source == nil || projectile == nil {
		return nil
	}
	var hit *world.Entity
	bestAlong := math.MaxFloat64
	step := geom.Distance(previousPosition, projectile.Position)
	w.ForEachEntity(func(target *world.Entity) {
		if !canHookTarget(source, target) {
			return
		}
		along, perpendicular := geom.ProjectPoint(previousPosition, projectile.Dir, target.Position)
		if along < -target.Radius || along > step+target.Radius || perpendicular > projectile.Radius+target.Radius {
			return
		}
		if along < bestAlong {
			bestAlong = along
			hit = target
		}
	})
	return hit
}

func canHookTarget(source *world.Entity, target *world.Entity) bool {
	if source == nil || target == nil || target.ID == source.ID || target.Stats.HP <= 0 || target.Death.Dead || target.Control.UntargetableUntilTick > 0 {
		return false
	}
	return world.IsHeroUnit(target) || world.IsMinion(target) || world.IsMonster(target) || target.Kind == world.EntityKindDummy
}

func resolveHookHit(w *world.World, source *world.Entity, target *world.Entity, level int, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || tickRate <= 0 {
		return
	}
	ally := target.Team == source.Team
	end, pullDistance := hookEnd(source, target, skillMeta(w.SkillConfig(qID), "frontPadding", 1))
	if !ally {
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		wasAlive := target.Stats.HP > 0
		rawDamage := hookDamage(w, source, target, level, pullDistance)
		w.ApplyTrueDamage(source, target, rawDamage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
			return
		}
	} else {
		reduceAllyCooldown(source, w.SkillConfig(qID), level, tick, tickRate)
	}
	pullHookTarget(w, source, target, end, pullDistance, tick, tickRate)
}

func hookDamage(w *world.World, source *world.Entity, target *world.Entity, level int, pullDistance float64) float64 {
	skill := w.SkillConfig(qID)
	if world.IsMinion(target) {
		return target.Stats.HP + target.Stats.MaxHP
	}
	return skillList(skill, "baseDamage", level, []float64{150, 220, 290, 360, 420}) + source.Stats.Attack*skillMeta(skill, "totalAdRatio", 0.3) + pullDistance*skillMeta(skill, "pullDistanceDamageRatio", 0.2)
}

func hookEnd(source *world.Entity, target *world.Entity, padding float64) (world.Vector2, float64) {
	dx, dy := normalize(target.Position.X-source.Position.X, target.Position.Y-source.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	end := world.Vector2{X: source.Position.X + dx*(source.Radius+target.Radius+padding), Y: source.Position.Y + dy*(source.Radius+target.Radius+padding)}
	return end, geom.Distance(target.Position, end)
}

func pullHookTarget(w *world.World, source *world.Entity, target *world.Entity, end world.Vector2, pullDistance float64, tick uint64, tickRate int) {
	speed := skillMeta(w.SkillConfig(qID), "pullSpeed", 1450)
	pullTicks := maxUint64(1, uint64(math.Ceil(pullDistance/speed*float64(tickRate))))
	until := tick + pullTicks
	target.Intent = world.IntentState{}
	target.Combat.PendingAttackTargetID = ""
	target.Combat.AttackReleaseTick = 0
	target.Control.DashStartTick = tick
	target.Control.DashStart = target.Position
	target.Control.DashEnd = w.ClampWorldPoint(end)
	target.Control.DashUntilTick = until
	target.Control.ActionLockedUntilTick = maxUint64(target.Control.ActionLockedUntilTick, until)
	target.Control.AirborneUntilTick = maxUint64(target.Control.AirborneUntilTick, until)
	w.PutSkillEffect(world.SkillEffect{ID: w.NextEffectID("effect:butcher_q_pull:"), Kind: "butcher_q_pull", Team: source.Team, SourceID: source.ID, SourceHeroID: source.HeroID, TargetID: target.ID, Start: source.Position, End: target.Position, Radius: target.Radius, CreatedAt: tick, ExpiresAt: until})
}

func reduceAllyCooldown(entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	state := entity.Skills[qID]
	fullTicks := cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{21000, 19000, 17000, 15000, 13000})), tickRate)
	reduction := uint64(math.Round(float64(fullTicks) * (1 - skillMeta(skill, "allyCooldownMultiplier", 0.5))))
	if state.CooldownUntilTick > reduction {
		state.CooldownUntilTick -= reduction
	}
	if state.CooldownUntilTick < tick {
		state.CooldownUntilTick = tick
	}
	entity.Skills[qID] = state
}

func DamageBlock(w *world.World, entity *world.Entity) float64 {
	if entity == nil || entity.HeroID != heroID {
		return 0
	}
	block := levelCurve(entity.Level, passiveMeta(w, "damageBlockMin", 8), passiveMeta(w, "damageBlockMax", 26))
	if entity.Passive.ButcherEUntil > 0 {
		skill := w.SkillConfig(eID)
		block += skillList(skill, "damageBlock", entity.Passive.ButcherELevel, []float64{35, 50, 65, 80, 95})
		block += entity.Stats.PhysicalDefense*skillMeta(skill, "armorRatio", 0.05) + entity.Stats.MagicDefense*skillMeta(skill, "magicResistanceRatio", 0.05)
	}
	return block
}

func MagicDamageReduction(w *world.World, entity *world.Entity, _ uint64) float64 {
	if entity == nil || entity.HeroID != heroID {
		return 0
	}
	return levelCurve(entity.Level, passiveMeta(w, "magicDamageReductionMin", 0.04), passiveMeta(w, "magicDamageReductionMax", 0.16))
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	if w == nil || target == nil || !world.IsHeroUnit(target) {
		return
	}
	radius := passiveMeta(w, "fleshRange", 450)
	w.ForEachEntity(func(entity *world.Entity) {
		if entity.Kind != world.EntityKindPlayer || entity.HeroID != heroID || entity.Team == target.Team {
			return
		}
		nearby := !entity.Death.Dead && entity.Stats.HP > 0 && geom.Distance(entity.Position, target.Position) <= radius
		if entity != killer && !nearby {
			return
		}
		entity.Passive.ButcherFlesh += fleshGain(w.SkillConfig(passiveID), entity.Level)
		w.RefreshPlayerStats(entity)
	})
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || entity.HeroID != heroID || stats == nil {
		return
	}
	magicReduction := MagicDamageReduction(w, entity, 0)
	stats.MagicDamageReduce = 1 - (1-stats.MagicDamageReduce)*(1-magicReduction)
	if entity.Passive.ButcherFlesh <= 0 {
		return
	}
	stacks := float64(entity.Passive.ButcherFlesh)
	hp := stacks * passiveMeta(w, "maxHPPerStack", 0.22)
	attack := stacks * passiveMeta(w, "attackPerStack", 0.01)
	stats.HP += hp
	stats.MaxHP += hp
	stats.BonusHP += hp
	stats.HPRegen5 += stacks * passiveMeta(w, "hpRegenPerSecondPerStack", 0.001) * 5
	stats.Attack += attack
	stats.BonusAttack += attack
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID {
		return nil
	}
	buffs := make([]world.BuffState, 0, 4)
	if entity.Passive.ButcherWActive {
		buffs = append(buffs, world.BuffState{ID: wID, Name: "腐烂"})
	}
	if entity.Passive.ButcherEUntil > tick {
		buffs = append(buffs, world.BuffState{ID: eID, Name: "肉盾", ExpiresAtTick: entity.Passive.ButcherEUntil})
	}
	if entity.Passive.ButcherRUntil > tick {
		buffs = append(buffs, world.BuffState{ID: rID, Name: "肢解", ExpiresAtTick: entity.Passive.ButcherRUntil})
	}
	if entity.Passive.ButcherFlesh <= 0 {
		return buffs
	}
	stacks := float64(entity.Passive.ButcherFlesh)
	return append(buffs, world.BuffState{
		ID:      "butcher_flesh",
		Name:    "腐血",
		Stacks:  entity.Passive.ButcherFlesh,
		Tooltip: fmt.Sprintf("最大生命 +%.2f\n生命回复 +%.3f/秒\n攻击力 +%.2f", stacks*passiveMeta(w, "maxHPPerStack", 0.22), stacks*passiveMeta(w, "hpRegenPerSecondPerStack", 0.001), stacks*passiveMeta(w, "attackPerStack", 0.01)),
	})
}

func passiveMeta(w *world.World, key string, fallback float64) float64 {
	if w == nil {
		return fallback
	}
	if value, ok := w.SkillConfig(passiveID).Meta[key]; ok {
		return value
	}
	return fallback
}

func fleshGain(skill config.SkillConfig, level int) int {
	values := skill.MetaLists["fleshPerHeroDeath"]
	tiers := skill.MetaLists["fleshTierLevels"]
	gain := 3
	for i, tier := range tiers {
		if level < int(tier) || i >= len(values) {
			break
		}
		gain = int(values[i])
	}
	return gain
}

func levelCurve(level int, start float64, end float64) float64 {
	if level < world.MinHeroLevel {
		level = world.MinHeroLevel
	}
	if level > world.MaxHeroLevel {
		level = world.MaxHeroLevel
	}
	return start + (end-start)*float64(level-world.MinHeroLevel)/float64(world.MaxHeroLevel-world.MinHeroLevel)
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := skill.MetaLists[key]
	if len(values) == 0 {
		values = fallback
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

func normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func maxUint64(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
