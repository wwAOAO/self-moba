package doctor

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID                  = "doctor"
	passiveID               = "doctor_passive"
	qID                     = "doctor_q"
	wID                     = "doctor_w"
	eID                     = "doctor_e"
	rID                     = "doctor_r"
	defaultHealMaxHPRatio   = 0.04
	defaultCooldownRefundS  = 15
	defaultExtraRegen5Ratio = 0.002
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		Tick:                           Tick,
		OnBasicHit:                     OnBasicHit,
		OnDamaged:                      OnDamaged,
		SpecialRecast:                  SpecialRecast,
		ActiveBuffs:                    ActiveBuffs,
		ApplyStats:                     ApplyStats,
		BasicAttackBonusPhysicalDamage: EBonusPhysicalDamage,
		DoctorQDamage:                  QDamage,
		DoctorQHit:                     QHit,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID {
		return
	}
	releaseQ(w, entity, tick, tickRate)
	tickW(w, entity, tick, tickRate)
	tickR(w, entity, tick, tickRate)
	if entity.Passive.DoctorCanisterEffectID == "" {
		return
	}
	if tick >= entity.Passive.DoctorCanisterExpiresAt {
		w.RemoveDoctorCanister(entity)
		return
	}
	center := entity.Passive.DoctorCanisterPosition
	radius := entity.Passive.DoctorCanisterRadius
	if radius <= 0 {
		w.RemoveDoctorCanister(entity)
		return
	}
	if distance(entity.Position, center) <= radius+entity.Radius {
		pickupCanister(w, entity, tick, tickRate)
		return
	}
	w.ForEachEntity(func(other *world.Entity) {
		if entity.Passive.DoctorCanisterEffectID == "" || other == nil || other.ID == entity.ID || !world.IsHeroUnit(other) || other.Team == entity.Team || other.Stats.HP <= 0 {
			return
		}
		if distance(other.Position, center) <= radius+other.Radius {
			w.RemoveDoctorCanister(entity)
		}
	})
}

func SpecialRecast(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || entity.HeroID != heroID || cast.SkillID != wID || entity.Passive.DoctorWActiveUntil == 0 {
		return false
	}
	detonateW(w, entity, skill, tick, tickRate)
	return true
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.DoctorQPending || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "healthCost", state.Level, []float64{50, 60, 70, 80, 90})
	if entity.Stats.HP <= cost {
		return
	}
	beforeHP := entity.Stats.HP
	entity.Stats.HP -= cost
	w.RefreshStatsAfterHPChange(entity, beforeHP)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{4000, 4000, 4000, 4000, 4000})), tickRate)
	entity.Skills[qID] = state

	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.DoctorQPending = true
	entity.Passive.DoctorQRelease = tick + windupTicks
	entity.Passive.DoctorQTarget = qTargetPoint(w, entity, cast, skill)
	entity.Passive.DoctorQLevel = state.Level
	entity.Passive.DoctorQHealthCost = cost
	entity.Control.ActionLockedUntilTick = entity.Passive.DoctorQRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	if entity.Passive.DoctorWActiveUntil != 0 {
		detonateW(w, entity, skill, tick, tickRate)
		return
	}
	cost := entity.Stats.HP * skillMeta(skill, "healthCostCurrentRatio", 0.05)
	if cost <= 0 || entity.Stats.HP <= cost {
		return
	}
	beforeHP := entity.Stats.HP
	entity.Stats.HP -= cost
	w.RefreshStatsAfterHPChange(entity, beforeHP)

	durationTicks := secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate)
	if durationTicks < 1 {
		durationTicks = 1
	}
	intervalTicks := secondsToTicks(skillMeta(skill, "damageIntervalSeconds", 1), tickRate)
	if intervalTicks < 1 {
		intervalTicks = 1
	}
	activeUntil := tick + durationTicks
	entity.Passive.DoctorWActiveUntil = activeUntil
	entity.Passive.DoctorWNextDamageTick = tick + intervalTicks
	entity.Passive.DoctorWLevel = state.Level
	entity.Passive.DoctorWGrayHealth = 0
	entity.Passive.DoctorWEffectID = w.NextEffectID("effect:doctor_w:")
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{17000, 16500, 16000, 15500, 15000})), tickRate)
	state.Stacks = 1
	state.StacksExpireTick = activeUntil
	entity.Skills[wID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
	w.PutSkillEffect(world.SkillEffect{
		ID:           entity.Passive.DoctorWEffectID,
		Kind:         "doctor_w",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       skillMeta(skill, "radius", 325),
		CreatedAt:    tick,
		ExpiresAt:    activeUntil,
	})
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	if state.Stacks > 0 && state.StacksExpireTick > 0 && tick >= state.StacksExpireTick {
		state.Stacks = 0
		state.StacksExpireTick = 0
	}
	if state.Stacks > 0 {
		return
	}
	cost := skillList(skill, "healthCost", state.Level, []float64{10, 25, 40, 55, 70})
	if entity.Stats.HP <= cost {
		return
	}
	beforeHP := entity.Stats.HP
	entity.Stats.HP -= cost
	w.RefreshStatsAfterHPChange(entity, beforeHP)
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000})), tickRate)
	entity.Combat.NextAttackTick = tick
	entity.Skills[eID] = state
	w.RefreshPlayerStats(entity)
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := entity.Stats.HP * skillMeta(skill, "healthCostCurrentRatio", 0.2)
	if cost <= 0 || entity.Stats.HP <= cost {
		return
	}
	beforeHP := entity.Stats.HP
	entity.Stats.HP -= cost
	w.RefreshStatsAfterHPChange(entity, beforeHP)
	durationTicks := secondsToTicks(skillMeta(skill, "durationSeconds", 12), tickRate)
	if durationTicks < 1 {
		durationTicks = 1
	}
	healIntervalTicks := secondsToTicks(skillMeta(skill, "healIntervalSeconds", 1), tickRate)
	if healIntervalTicks < 1 {
		healIntervalTicks = 1
	}
	entity.Passive.DoctorRUntil = tick + durationTicks
	entity.Passive.DoctorRNextHealTick = tick + healIntervalTicks
	entity.Passive.DoctorRLevel = state.Level
	entity.Passive.DoctorREffectID = w.NextEffectID("effect:doctor_r:")
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{110000, 100000, 90000})), tickRate)
	state.Stacks = 1
	state.StacksExpireTick = entity.Passive.DoctorRUntil
	entity.Skills[rID] = state
	w.RefreshPlayerStats(entity)
	w.PutSkillEffect(world.SkillEffect{
		ID:           entity.Passive.DoctorREffectID,
		Kind:         "doctor_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       entity.Radius,
		CreatedAt:    tick,
		ExpiresAt:    entity.Passive.DoctorRUntil,
	})
}

func tickW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.DoctorWActiveUntil == 0 || tickRate <= 0 {
		return
	}
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		clearW(w, entity)
		return
	}
	skill := w.SkillConfig(wID)
	intervalTicks := secondsToTicks(skillMeta(skill, "damageIntervalSeconds", 1), tickRate)
	if intervalTicks < 1 {
		intervalTicks = 1
	}
	for entity.Passive.DoctorWNextDamageTick > 0 && tick >= entity.Passive.DoctorWNextDamageTick && entity.Passive.DoctorWNextDamageTick <= entity.Passive.DoctorWActiveUntil {
		applyWPeriodicDamage(w, entity, skill, entity.Passive.DoctorWLevel, entity.Passive.DoctorWNextDamageTick, tickRate)
		entity.Passive.DoctorWNextDamageTick += intervalTicks
	}
	if tick >= entity.Passive.DoctorWActiveUntil {
		detonateW(w, entity, skill, tick, tickRate)
	}
}

func tickR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.DoctorRUntil == 0 || tickRate <= 0 {
		return
	}
	if entity.Stats.HP <= 0 || entity.Death.Dead || tick >= entity.Passive.DoctorRUntil {
		clearR(w, entity)
		return
	}
	skill := w.SkillConfig(rID)
	intervalTicks := secondsToTicks(skillMeta(skill, "healIntervalSeconds", 1), tickRate)
	if intervalTicks < 1 {
		intervalTicks = 1
	}
	for entity.Passive.DoctorRNextHealTick > 0 && tick >= entity.Passive.DoctorRNextHealTick && entity.Passive.DoctorRNextHealTick < entity.Passive.DoctorRUntil {
		beforeHP := entity.Stats.HP
		heal := entity.Stats.MaxHP * skillList(skill, "maxHPHealPerSecond", entity.Passive.DoctorRLevel, []float64{0.08, 0.11, 0.14})
		entity.Stats.HP = math.Min(entity.Stats.MaxHP, entity.Stats.HP+heal)
		w.RefreshStatsAfterHPChange(entity, beforeHP)
		entity.Passive.DoctorRNextHealTick += intervalTicks
	}
}

func OnDamaged(w *world.World, source *world.Entity, target *world.Entity, basicAttack bool, pet bool, skipBleed bool, tick uint64, tickRate int) {
	if target == nil || target.HeroID != heroID || target.Passive.DoctorWActiveUntil == 0 || tick > target.Passive.DoctorWActiveUntil || target.Combat.LastDamage <= 0 {
		return
	}
	skill := w.SkillConfig(wID)
	ratio := skillList(skill, "grayHealthRatio", target.Passive.DoctorWLevel, []float64{0.25, 0.3, 0.35, 0.4, 0.45})
	target.Passive.DoctorWGrayHealth += float64(target.Combat.LastDamage) * ratio
}

func applyWPeriodicDamage(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	raw := skillList(skill, "damagePerSecond", level, []float64{20, 35, 50, 65, 80})
	applyWAreaDamage(w, entity, skillMeta(skill, "radius", 325), raw, tick, tickRate)
}

func detonateW(w *world.World, entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.DoctorWActiveUntil == 0 {
		return
	}
	level := entity.Passive.DoctorWLevel
	grayHealth := entity.Passive.DoctorWGrayHealth
	clearW(w, entity)
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	raw := skillList(skill, "burstDamage", level, []float64{20, 35, 50, 65, 80}) + entity.Stats.BonusHP*skillMeta(skill, "bonusHPRatio", 0.07)
	hitHero, hitAny := applyWAreaDamage(w, entity, skillMeta(skill, "radius", 325), raw, tick, tickRate)
	if grayHealth <= 0 || !hitAny {
		return
	}
	healRatio := skillMeta(skill, "heroGrayHealRatio", 1)
	if !hitHero {
		healRatio = skillMeta(skill, "nonHeroGrayHealRatio", 0.5)
	}
	beforeHP := entity.Stats.HP
	entity.Stats.HP = math.Min(entity.Stats.MaxHP, entity.Stats.HP+grayHealth*healRatio)
	w.RefreshStatsAfterHPChange(entity, beforeHP)
}

func applyWAreaDamage(w *world.World, entity *world.Entity, radius float64, raw float64, tick uint64, tickRate int) (bool, bool) {
	hitHero := false
	hitAny := false
	for _, target := range w.TargetsInRadius(entity, entity.Position, radius) {
		damage := w.MagicDamageAfterResistance(entity, target, raw, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		hitAny = true
		if world.IsHeroUnit(target) {
			hitHero = true
		}
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	return hitHero, hitAny
}

func clearW(w *world.World, entity *world.Entity) {
	if entity == nil {
		return
	}
	if entity.Passive.DoctorWEffectID != "" {
		w.RemoveSkillEffect(entity.Passive.DoctorWEffectID)
	}
	state := entity.Skills[wID]
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[wID] = state
	entity.Passive.DoctorWActiveUntil = 0
	entity.Passive.DoctorWNextDamageTick = 0
	entity.Passive.DoctorWLevel = 0
	entity.Passive.DoctorWGrayHealth = 0
	entity.Passive.DoctorWEffectID = ""
}

func clearR(w *world.World, entity *world.Entity) {
	if entity == nil {
		return
	}
	if entity.Passive.DoctorREffectID != "" {
		w.RemoveSkillEffect(entity.Passive.DoctorREffectID)
	}
	state := entity.Skills[rID]
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[rID] = state
	entity.Passive.DoctorRUntil = 0
	entity.Passive.DoctorRNextHealTick = 0
	entity.Passive.DoctorRLevel = 0
	entity.Passive.DoctorREffectID = ""
	w.RefreshPlayerStats(entity)
}

func EBonusPhysicalDamage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) int {
	if !eActive(attacker, tick) {
		return 0
	}
	skill := w.SkillConfig(eID)
	missingRatio := 0.0
	if attacker.Stats.MaxHP > 0 {
		missingRatio = (attacker.Stats.MaxHP - attacker.Stats.HP) / attacker.Stats.MaxHP
	}
	capRatio := skillMeta(skill, "missingHPCapRatio", 0.7)
	bonusCap := skillMeta(skill, "missingHPDamageBonusCap", 0.4)
	bonus := 0.0
	if capRatio > 0 {
		bonus = math.Min(math.Max(missingRatio, 0), capRatio) / capRatio * bonusCap
	}
	return int(math.Round(attacker.Stats.Attack * skillMeta(skill, "totalADRatio", 1) * (1 + bonus)))
}

func OnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if !eActive(attacker, tick) || target == nil {
		return
	}
	state := attacker.Skills[eID]
	state.Stacks = 0
	state.StacksExpireTick = 0
	attacker.Skills[eID] = state
	w.RefreshPlayerStats(attacker)
}

func eActive(entity *world.Entity, tick uint64) bool {
	if entity == nil || entity.HeroID != heroID {
		return false
	}
	state := entity.Skills[eID]
	return state.Stacks > 0 && state.Level > 0 && (tick == 0 || tick < state.StacksExpireTick)
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.DoctorQPending || tick < entity.Passive.DoctorQRelease || tickRate <= 0 {
		return
	}
	target := entity.Passive.DoctorQTarget
	level := entity.Passive.DoctorQLevel
	cost := entity.Passive.DoctorQHealthCost
	entity.Passive.DoctorQPending = false
	entity.Passive.DoctorQRelease = 0
	entity.Passive.DoctorQTarget = world.Vector2{}
	entity.Passive.DoctorQLevel = 0
	entity.Passive.DoctorQHealthCost = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	speed := skillMeta(skill, "projectileSpeed", 2000)
	qRange := skillRange(skill, 975)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:doctor_q:"),
		Kind:         "doctor_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileWidth", 120) / 2,
		Damage:       level,
		EffectRatio:  cost,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	if target == nil {
		return 0
	}
	ratio := skillList(skill, "currentHPRatio", skillLevel, []float64{0.2, 0.225, 0.25, 0.275, 0.3})
	raw := math.Max(target.Stats.HP*ratio, skillList(skill, "minimumDamage", skillLevel, []float64{80, 130, 180, 230, 280}))
	if world.IsMonster(target) {
		raw = math.Min(raw, skillList(skill, "monsterMaxDamage", skillLevel, []float64{300, 375, 450, 525, 600}))
	}
	return w.MagicDamageAfterResistance(attacker, target, raw, tick)
}

func QHit(w *world.World, source *world.Entity, target *world.Entity, projectile *world.Projectile, damage int, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || projectile == nil {
		return
	}
	wasAlive := target.Stats.HP > 0
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	} else {
		w.ApplyMagicDamage(source, target, damage, tickRate)
	}
	skill := w.SkillConfig(qID)
	w.ApplyMoveSpeedSlow(target, skillMeta(skill, "slow", 0.4), tick+secondsToTicks(skillMeta(skill, "slowSeconds", 2), tickRate))
	refund := projectile.EffectRatio * 0.5
	if target.Kind != world.EntityKindDummy && wasAlive && target.Stats.HP == 0 {
		refund = projectile.EffectRatio
	}
	if refund <= 0 || source.Stats.HP <= 0 || source.Death.Dead {
		return
	}
	beforeHP := source.Stats.HP
	source.Stats.HP = math.Min(source.Stats.MaxHP, source.Stats.HP+refund)
	w.RefreshStatsAfterHPChange(source, beforeHP)
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if w == nil || entity == nil || stats == nil || entity.HeroID != heroID {
		return
	}
	skill := w.SkillConfig(passiveID)
	stats.HPRegen5 += stats.MaxHP * skillMeta(skill, "extraHPRegen5MaxHPRatio", defaultExtraRegen5Ratio)
	if state := entity.Skills[eID]; state.Level > 0 {
		bonusAttack := stats.MaxHP * skillList(w.SkillConfig(eID), "maxHPAttackRatio", state.Level, []float64{0.02, 0.0225, 0.025, 0.0275, 0.03})
		stats.Attack += bonusAttack
		stats.BonusAttack += bonusAttack
	}
	if eActive(entity, 0) {
		stats.AttackRange += skillMeta(w.SkillConfig(eID), "attackRangeBonus", 50)
	}
	if entity.Passive.DoctorRUntil > 0 {
		rSkill := w.SkillConfig(rID)
		stats.MoveSpeed *= 1 + skillList(rSkill, "moveSpeedBonus", entity.Passive.DoctorRLevel, []float64{0.3, 0.5, 0.7})
		bonusAttack := stats.Attack * skillList(rSkill, "attackBonusRatio", entity.Passive.DoctorRLevel, []float64{0.15, 0.25, 0.35})
		stats.Attack += bonusAttack
		stats.BonusAttack += bonusAttack
	}
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID {
		return nil
	}
	buffs := make([]world.BuffState, 0, 2)
	if tick < entity.Passive.DoctorPassiveCooldownUntil {
		buffs = append(buffs, world.BuffState{
			ID:            passiveID,
			Name:          "自由之足",
			ExpiresAtTick: entity.Passive.DoctorPassiveCooldownUntil,
		})
	} else {
		buffs = append(buffs, world.BuffState{ID: passiveID, Name: "自由之足"})
	}
	if entity.Passive.DoctorWActiveUntil > tick {
		buffs = append(buffs, world.BuffState{
			ID:            wID,
			Name:          "电击疗法",
			ExpiresAtTick: entity.Passive.DoctorWActiveUntil,
		})
	}
	if state := entity.Skills[eID]; state.Stacks > 0 && tick < state.StacksExpireTick {
		buffs = append(buffs, world.BuffState{
			ID:            eID,
			Name:          "大力行医",
			ExpiresAtTick: state.StacksExpireTick,
		})
	}
	if entity.Passive.DoctorRUntil > tick {
		buffs = append(buffs, world.BuffState{
			ID:            rID,
			Name:          "极限剂量",
			ExpiresAtTick: entity.Passive.DoctorRUntil,
		})
	}
	return buffs
}

func pickupCanister(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	skill := w.SkillConfig(passiveID)
	healRatio := skillMeta(skill, "canisterHealMaxHPRatio", defaultHealMaxHPRatio)
	refundTicks := secondsToTicks(skillMeta(skill, "canisterCooldownRefundSeconds", defaultCooldownRefundS), tickRate)
	beforeHP := entity.Stats.HP
	entity.Stats.HP = math.Min(entity.Stats.MaxHP, entity.Stats.HP+entity.Stats.MaxHP*healRatio)
	w.RefreshStatsAfterHPChange(entity, beforeHP)

	if refundTicks >= entity.Passive.DoctorPassiveCooldownUntil-tick {
		entity.Passive.DoctorPassiveCooldownUntil = tick
	} else {
		entity.Passive.DoctorPassiveCooldownUntil -= refundTicks
	}
	if state, ok := entity.Skills[passiveID]; ok {
		state.CooldownUntilTick = entity.Passive.DoctorPassiveCooldownUntil
		entity.Skills[passiveID] = state
	}
	w.RemoveDoctorCanister(entity)
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

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
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

func qTargetPoint(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) world.Vector2 {
	target := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	if cast.TargetID != "" {
		if targetEntity := w.EntityByID(cast.TargetID); targetEntity != nil {
			target = targetEntity.Position
		}
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	qRange := skillRange(skill, 975)
	return world.Vector2{X: entity.Position.X + dx*qRange, Y: entity.Position.Y + dy*qRange}
}

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func distance(a world.Vector2, b world.Vector2) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(dx, dy)
}
