package killer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"sort"
)

const (
	heroID    = "killer"
	passiveID = "killer_passive"
	qID       = "killer_q"
	wID       = "killer_w"
	eID       = "killer_e"
	rID       = "killer_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		Tick:                Tick,
		OnDamage:            markVoracityTarget,
		OnKill:              triggerVoracity,
		ActiveBuffs:         ActiveBuffs,
		MoveSpeedMultiplier: MoveSpeedMultiplier,
		DamageReduction:     DamageReduction,
		KillerQDamage:       QDamage,
		KillerQHit:          QHit,
		KillerRDamage:       RDamage,
		KillerRHit:          RHit,
	})
}

type eDestination struct {
	Entity   *world.Entity
	DaggerID string
	Position world.Vector2
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	destination := findEDestination(w, entity, cast, skill, tick)
	if destination == nil {
		return
	}
	baseCooldownMS := skillList(skill, "cooldownMs", state.Level, []float64{14000, 12500, 11000, 9500, 8000})
	cooldownMS := baseCooldownMS
	if destination.DaggerID != "" {
		cooldownMS *= skillMeta(skill, "daggerCooldownRemainingRatio", 0.22)
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(math.Round(cooldownMS)), tickRate)
	entity.Skills[eID] = state
	start := entity.Position
	entity.Position = w.ClampWorldPoint(destination.Position)
	entity.Intent = world.IntentState{}
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0
	entity.Passive.KillerEDamageReduceUntil = tick + secondsToTicks(skillMeta(skill, "damageReductionSeconds", 3), tickRate)
	entity.Passive.KillerEDamageReduction = skillMeta(skill, "damageReduction", 0.2)
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:killer_e:"),
		Kind:         "killer_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        start,
		End:          entity.Position,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "blinkEffectSeconds", 0.2), tickRate),
	})
	if target := destination.Entity; target != nil && target.Kind != world.EntityKindFruit && world.CanAttackTarget(entity, target) {
		applyEDamage(w, entity, target, skill, state.Level, tick, tickRate)
	}
	if destination.DaggerID != "" {
		pickupDaggerByID(w, entity, destination.DaggerID, tick, tickRate)
	}
}

func CastR(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 || channelingR(entity, tick) || rInterrupted(entity, tick) {
		return
	}
	segments := int(skillMeta(skill, "damageSegments", 10))
	if segments <= 0 {
		return
	}
	durationTicks := secondsToTicks(skillMeta(skill, "channelSeconds", 2.5), tickRate)
	if durationTicks < uint64(segments) {
		durationTicks = uint64(segments)
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{90000, 60000, 45000})), tickRate)
	entity.Skills[rID] = state
	entity.Passive.KillerRStartTick = tick
	entity.Passive.KillerRExpireTick = tick + durationTicks
	entity.Passive.KillerRNextTick = tick
	entity.Passive.KillerRLevel = state.Level
	entity.Passive.KillerRSegmentsFired = 0
	entity.Passive.KillerRMoveSpeedMultiplier = skillMeta(skill, "moveSpeedMultiplier", 0.2)
	entity.Intent.AttackTargetID = ""
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0
	entity.Passive.KillerREffectID = w.NextEffectID("effect:killer_r_channel:")
	w.PutSkillEffect(world.SkillEffect{
		ID:           entity.Passive.KillerREffectID,
		Kind:         "killer_r_channel",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       skill.Range,
		Count:        segments,
		CreatedAt:    tick,
		ExpiresAt:    entity.Passive.KillerRExpireTick,
	})
	tickR(w, entity, tick, tickRate)
}

func tickR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || entity.Passive.KillerRExpireTick == 0 {
		return
	}
	if entity.Stats.HP <= 0 || entity.Death.Dead || tick >= entity.Passive.KillerRExpireTick || rInterrupted(entity, tick) {
		stopR(w, entity)
		return
	}
	entity.Intent.AttackTargetID = ""
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0
	if tick < entity.Passive.KillerRNextTick {
		return
	}
	skill := w.SkillConfig(rID)
	segments := int(skillMeta(skill, "damageSegments", 10))
	if entity.Passive.KillerRSegmentsFired >= segments {
		return
	}
	fireRSegment(w, entity, skill, entity.Passive.KillerRLevel, tick, tickRate)
	entity.Passive.KillerRSegmentsFired++
	durationTicks := entity.Passive.KillerRExpireTick - entity.Passive.KillerRStartTick
	entity.Passive.KillerRNextTick = entity.Passive.KillerRStartTick + uint64(entity.Passive.KillerRSegmentsFired)*durationTicks/uint64(segments)
}

func fireRSegment(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	targets := nearestRTargets(w, entity, skill.Range, int(skillMeta(skill, "maxTargets", 3)))
	for _, target := range targets {
		dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
		if dx == 0 && dy == 0 {
			dx = 1
		}
		w.PutProjectile(&world.Projectile{
			ID:           w.NextProjectileID("projectile:killer_r:"),
			Kind:         "killer_r",
			Team:         entity.Team,
			SourceID:     entity.ID,
			TargetID:     target.ID,
			SkillID:      rID,
			Position:     entity.Position,
			Start:        entity.Position,
			Dir:          world.Vector2{X: dx, Y: dy},
			SpeedPerTick: skillMeta(skill, "projectileSpeed", 2000) / float64(tickRate),
			Range:        distance(entity.Position, target.Position) + target.Radius,
			Radius:       skillMeta(skill, "projectileRadius", 18),
			Damage:       level,
			CreatedAt:    tick,
			ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "projectileLifetimeSeconds", 1), tickRate),
			HitIDs:       make(map[string]bool),
		})
	}
}

func nearestRTargets(w *world.World, source *world.Entity, radius float64, limit int) []*world.Entity {
	if w == nil || source == nil || limit <= 0 {
		return nil
	}
	if radius <= 0 {
		radius = 550
	}
	targets := make([]*world.Entity, 0, limit)
	w.ForEachEntity(func(target *world.Entity) {
		if !world.IsHeroUnit(target) || !world.CanAttackTarget(source, target) || distance(source.Position, target.Position) > radius+target.Radius {
			return
		}
		targets = append(targets, target)
	})
	sort.Slice(targets, func(i int, j int) bool {
		left := distance(source.Position, targets[i].Position)
		right := distance(source.Position, targets[j].Position)
		if left == right {
			return targets[i].ID < targets[j].ID
		}
		return left < right
	})
	if len(targets) > limit {
		targets = targets[:limit]
	}
	return targets
}

func RDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	rawDamage := skillList(skill, "baseDamage", level, []float64{40, 65, 90}) +
		float64(attacker.Stats.AbilityPower)*skillList(skill, "apRatio", level, []float64{0.165, 0.264, 0.363}) +
		attacker.Stats.BonusAttack*skillList(skill, "bonusAdRatio", level, []float64{0.22, 0.352, 0.484})
	return w.MagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func RHit(w *world.World, _ *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if w == nil || target == nil || tickRate <= 0 {
		return
	}
	skill := w.SkillConfig(rID)
	w.ApplyGrievousWounds(target, skillMeta(skill, "grievousWounds", 0.4), tick+secondsToTicks(skillMeta(skill, "grievousWoundsSeconds", 3), tickRate))
}

func channelingR(entity *world.Entity, tick uint64) bool {
	return entity != nil && entity.Passive.KillerRExpireTick > tick
}

func rInterrupted(entity *world.Entity, tick uint64) bool {
	if entity == nil {
		return true
	}
	return tick < entity.Control.StunnedUntilTick || tick < entity.Control.AirborneUntilTick || tick < entity.Control.RootedUntilTick ||
		tick < entity.Control.TauntedUntilTick || tick < entity.Control.SuppressedUntilTick
}

func stopR(w *world.World, entity *world.Entity) {
	if w == nil || entity == nil {
		return
	}
	w.RemoveSkillEffect(entity.Passive.KillerREffectID)
	entity.Passive.KillerRStartTick = 0
	entity.Passive.KillerRExpireTick = 0
	entity.Passive.KillerRNextTick = 0
	entity.Passive.KillerRLevel = 0
	entity.Passive.KillerRSegmentsFired = 0
	entity.Passive.KillerREffectID = ""
	entity.Passive.KillerRMoveSpeedMultiplier = 0
}

func DamageReduction(entity *world.Entity, tick uint64) float64 {
	if entity == nil || entity.HeroID != heroID || tick >= entity.Passive.KillerEDamageReduceUntil {
		return 0
	}
	return entity.Passive.KillerEDamageReduction
}

func ActiveBuffs(_ *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID {
		return nil
	}
	buffs := make([]world.BuffState, 0, 3)
	if entity.Passive.KillerWMoveSpeedUntilTick > tick {
		buffs = append(buffs, world.BuffState{ID: "killer_preparation_speed", Name: "伺机待发", ExpiresAtTick: entity.Passive.KillerWMoveSpeedUntilTick})
	}
	if entity.Passive.KillerEDamageReduceUntil > tick {
		buffs = append(buffs, world.BuffState{ID: "killer_shunpo_reduction", Name: "瞬步减伤", ExpiresAtTick: entity.Passive.KillerEDamageReduceUntil})
	}
	if entity.Passive.KillerRExpireTick > tick {
		buffs = append(buffs, world.BuffState{ID: "killer_death_lotus", Name: "死亡莲华", ExpiresAtTick: entity.Passive.KillerRExpireTick})
	}
	return buffs
}

func applyEDamage(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil {
		return
	}
	rawDamage := skillList(skill, "baseDamage", level, []float64{30, 45, 60, 75, 90}) +
		float64(source.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.4) +
		source.Stats.BonusAttack*skillMeta(skill, "bonusAdRatio", 0.65)
	damage := w.MagicDamageAfterResistance(source, target, rawDamage, tick)
	wasAlive := target.Stats.HP > 0
	target.Combat.LastHitTick = tick
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
		return
	}
	w.ApplyMagicDamage(source, target, damage, tickRate)
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(source, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}

func findEDestination(w *world.World, source *world.Entity, cast protocol.CastInput, skill config.SkillConfig, tick uint64) *eDestination {
	blinkRange := skill.Range
	if blinkRange <= 0 {
		blinkRange = 700
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	var best *eDestination
	bestDistance := math.MaxFloat64
	for _, dagger := range source.Passive.KillerDaggers {
		if tick >= dagger.ExpiresAt || distance(source.Position, dagger.Position) > blinkRange {
			continue
		}
		dist := distance(point, dagger.Position)
		if dist < bestDistance {
			best = &eDestination{DaggerID: dagger.EffectID, Position: dagger.Position}
			bestDistance = dist
		}
	}
	w.ForEachEntity(func(target *world.Entity) {
		if !validEEntityTarget(source, target, blinkRange, tick) {
			return
		}
		dist := distance(point, target.Position)
		if dist >= bestDistance {
			return
		}
		best = &eDestination{Entity: target, Position: target.Position}
		bestDistance = dist
	})
	return best
}

func validEEntityTarget(source *world.Entity, target *world.Entity, blinkRange float64, tick uint64) bool {
	if source == nil || target == nil || target.ID == source.ID || target.Stats.HP <= 0 || distance(source.Position, target.Position) > blinkRange+target.Radius {
		return false
	}
	if target.Kind == world.EntityKindPlayer && target.Death.Dead {
		return false
	}
	if target.Control.UntargetableUntilTick > tick {
		return false
	}
	switch target.Kind {
	case world.EntityKindPlayer, world.EntityKindEnemyHero,
		world.EntityKindMeleeMinion, world.EntityKindRangedMinion, world.EntityKindSiegeMinion, world.EntityKindSuperMinion,
		world.EntityKindFruit, world.EntityKindDummy:
		return true
	case world.EntityKindTower, world.EntityKindBarracks, world.EntityKindCrystal, world.EntityKindFountain, world.EntityKindWard:
		return false
	default:
		return world.IsMonster(target)
	}
}

func pickupDaggerByID(w *world.World, entity *world.Entity, effectID string, tick uint64, tickRate int) bool {
	if w == nil || entity == nil || effectID == "" {
		return false
	}
	for index, dagger := range entity.Passive.KillerDaggers {
		if dagger.EffectID != effectID || tick >= dagger.ExpiresAt {
			continue
		}
		w.RemoveSkillEffect(dagger.EffectID)
		entity.Passive.KillerDaggers = append(entity.Passive.KillerDaggers[:index], entity.Passive.KillerDaggers[index+1:]...)
		triggerDaggerSlash(w, entity, tick, tickRate)
		return true
	}
	return false
}

func CastW(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 || channelingR(entity, tick) {
		return
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{15000, 14000, 13000, 12000, 11000})), tickRate)
	entity.Skills[wID] = state
	landsAt := tick + secondsToTicks(skillMeta(skill, "landingDelaySeconds", 1), tickRate)
	effectID := w.NextEffectID("effect:killer_w_airborne_dagger:")
	entity.Passive.KillerAirborneDaggers = append(entity.Passive.KillerAirborneDaggers, world.KillerAirborneDaggerState{
		EffectID:           effectID,
		Position:           entity.Position,
		Direction:          world.Vector2{X: 0, Y: -1},
		LandsAt:            landsAt,
		GroundEffectKind:   "killer_w_dagger",
		GroundEffectPrefix: "effect:killer_w_dagger:",
	})
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         "killer_w_dagger_airborne",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       skillMeta(skill, "daggerRadius", 32),
		CreatedAt:    tick,
		ExpiresAt:    landsAt,
	})
	durationTicks := secondsToTicks(skillMeta(skill, "moveSpeedDurationSeconds", 1.25), tickRate)
	entity.Passive.KillerWMoveSpeedStartTick = tick
	entity.Passive.KillerWMoveSpeedUntilTick = tick + durationTicks
	entity.Passive.KillerWMoveSpeedBonus = skillList(skill, "moveSpeedBonus", state.Level, []float64{0.5, 0.6, 0.7, 0.8, 0.9})
}

func MoveSpeedMultiplier(entity *world.Entity, tick uint64) float64 {
	if entity == nil || entity.HeroID != heroID {
		return 1
	}
	multiplier := 1.0
	if entity.Passive.KillerWMoveSpeedUntilTick > entity.Passive.KillerWMoveSpeedStartTick &&
		tick >= entity.Passive.KillerWMoveSpeedStartTick && tick < entity.Passive.KillerWMoveSpeedUntilTick {
		duration := float64(entity.Passive.KillerWMoveSpeedUntilTick - entity.Passive.KillerWMoveSpeedStartTick)
		remaining := float64(entity.Passive.KillerWMoveSpeedUntilTick - tick)
		multiplier *= 1 + entity.Passive.KillerWMoveSpeedBonus*remaining/duration
	}
	if channelingR(entity, tick) {
		channelMultiplier := entity.Passive.KillerRMoveSpeedMultiplier
		if channelMultiplier <= 0 {
			channelMultiplier = 0.2
		}
		multiplier *= channelMultiplier
	}
	return multiplier
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.KillerQPending || tickRate <= 0 || channelingR(entity, tick) {
		return
	}
	target := targetedEnemy(w, entity, cast, skill)
	if target == nil {
		return
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{11000, 10000, 9000, 8000, 7000})), tickRate)
	entity.Skills[qID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.KillerQPending = true
	entity.Passive.KillerQReleaseTick = tick + windupTicks
	entity.Passive.KillerQTargetID = target.ID
	entity.Passive.KillerQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.KillerQReleaseTick
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:killer_q_cast_range:"),
		Kind:         "killer_q_cast_range",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		End:          target.Position,
		Range:        skill.Range,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    entity.Passive.KillerQReleaseTick,
	})
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	tickR(w, entity, tick, tickRate)
	releaseQ(w, entity, tick, tickRate)
	landAirborneDaggers(w, entity, tick, tickRate)
	if len(entity.Passive.KillerDaggers) == 0 {
		return
	}
	daggers := entity.Passive.KillerDaggers[:0]
	for _, dagger := range entity.Passive.KillerDaggers {
		if tick >= dagger.ExpiresAt {
			w.RemoveSkillEffect(dagger.EffectID)
			continue
		}
		if entity.Stats.HP > 0 && !entity.Death.Dead && daggerInPickupRange(w, entity, dagger) {
			w.RemoveSkillEffect(dagger.EffectID)
			triggerDaggerSlash(w, entity, tick, tickRate)
			continue
		}
		daggers = append(daggers, dagger)
	}
	entity.Passive.KillerDaggers = daggers
}

func landAirborneDaggers(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || len(entity.Passive.KillerAirborneDaggers) == 0 {
		return
	}
	pending := entity.Passive.KillerAirborneDaggers[:0]
	for _, dagger := range entity.Passive.KillerAirborneDaggers {
		if tick < dagger.LandsAt {
			pending = append(pending, dagger)
			continue
		}
		w.RemoveSkillEffect(dagger.EffectID)
		effectKind := dagger.GroundEffectKind
		effectPrefix := dagger.GroundEffectPrefix
		if effectKind == "" {
			effectKind = "killer_w_dagger"
		}
		if effectPrefix == "" {
			effectPrefix = "effect:killer_w_dagger:"
		}
		putGroundDagger(w, entity, dagger.Position, dagger.Direction, effectKind, effectPrefix, tick, tickRate)
	}
	entity.Passive.KillerAirborneDaggers = pending
}

func daggerInPickupRange(w *world.World, entity *world.Entity, dagger world.KillerDaggerState) bool {
	if w == nil || entity == nil {
		return false
	}
	pickupRadius := skillMeta(w.SkillConfig(passiveID), "daggerPickupRadius", 55)
	return distance(entity.Position, dagger.Position) <= pickupRadius+entity.Radius
}

func triggerDaggerSlash(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || tickRate <= 0 {
		return
	}
	skill := w.SkillConfig(passiveID)
	radius := skillMeta(skill, "daggerSlashRadius", 340)
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:killer_dagger_slash:"),
		Kind:         "killer_dagger_slash",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "daggerSlashEffectSeconds", 0.25), tickRate),
	})
	for _, target := range w.TargetsInRadius(entity, entity.Position, radius) {
		wasAlive := target.Stats.HP > 0
		target.Combat.LastHitTick = tick
		damage := w.MagicDamageAfterResistance(entity, target, daggerSlashRawDamage(entity, skill), tick)
		w.ApplyMagicDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
}

func daggerSlashRawDamage(entity *world.Entity, skill config.SkillConfig) float64 {
	if entity == nil {
		return 0
	}
	baseDamage := skillCurve(skill, "daggerBaseDamage", "daggerBaseDamageLevels", entity.Level, 68)
	apRatio := steppedValue(skill, "daggerAPRatio", "daggerAPRatioLevels", entity.Level, 0.7)
	return baseDamage + entity.Stats.BonusAttack*skillMeta(skill, "daggerBonusAdRatio", 0.65) + float64(entity.Stats.AbilityPower)*apRatio
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || !entity.Passive.KillerQPending || tick < entity.Passive.KillerQReleaseTick {
		return
	}
	targetID := entity.Passive.KillerQTargetID
	level := entity.Passive.KillerQLevel
	entity.Passive.KillerQPending = false
	entity.Passive.KillerQReleaseTick = 0
	entity.Passive.KillerQTargetID = ""
	entity.Passive.KillerQLevel = 0
	target := w.EntityByID(targetID)
	if entity.Stats.HP <= 0 || entity.Death.Dead || !world.CanAttackTarget(entity, target) || tickRate <= 0 {
		return
	}
	skill := w.SkillConfig(qID)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	speed := skillMeta(skill, "projectileSpeed", 1600)
	qRange := skill.Range
	if qRange <= 0 {
		qRange = 625
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:killer_q:"),
		Kind:         "killer_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 24),
		Damage:       level,
		MagicDamage:  int(skillMeta(skill, "maxBounces", 2)),
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	baseDamage := skillList(skill, "baseDamage", level, []float64{75, 105, 135, 165, 195})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.3)
	return w.MagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func QHit(w *world.World, source *world.Entity, target *world.Entity, projectile *world.Projectile, tick uint64, tickRate int) bool {
	if w == nil || source == nil || target == nil || projectile == nil {
		return false
	}
	if projectile.MagicDamage <= 0 {
		spawnQDagger(w, source, target, projectile, tick, tickRate)
		return false
	}
	next := nextQBounceTarget(w, source, target, projectile, skillMeta(w.SkillConfig(qID), "bounceRange", 450))
	if next == nil {
		spawnQDagger(w, source, target, projectile, tick, tickRate)
		return false
	}
	dx, dy := normalize(next.Position.X-target.Position.X, next.Position.Y-target.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	projectile.TargetID = next.ID
	projectile.Position = target.Position
	projectile.Start = target.Position
	projectile.Dir = world.Vector2{X: dx, Y: dy}
	projectile.Range = distance(target.Position, next.Position) + next.Radius
	projectile.Traveled = 0
	projectile.MagicDamage--
	projectile.CreatedAt = tick
	projectile.ExpiresAt = tick + secondsToTicks(2, tickRate)
	return true
}

func spawnQDagger(w *world.World, source *world.Entity, target *world.Entity, projectile *world.Projectile, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || projectile == nil || tickRate <= 0 {
		return
	}
	skill := w.SkillConfig(qID)
	dx, dy := normalize(target.Position.X-projectile.Start.X, target.Position.Y-projectile.Start.Y)
	if dx == 0 && dy == 0 {
		dx, dy = projectile.Dir.X, projectile.Dir.Y
	}
	landingDistance := skillMeta(skill, "daggerLandingDistance", 350)
	position := w.ClampWorldPoint(world.Vector2{X: target.Position.X + dx*landingDistance, Y: target.Position.Y + dy*landingDistance})
	travelDistance := distance(target.Position, position)
	landingSpeed := skillMeta(skill, "daggerLandingSpeed", 1000)
	travelSeconds := 0.35
	if landingSpeed > 0 {
		travelSeconds = travelDistance / landingSpeed
	}
	landsAt := tick + secondsToTicks(travelSeconds, tickRate)
	if landsAt <= tick {
		landsAt = tick + 1
	}
	effectID := w.NextEffectID("effect:killer_q_dagger_airborne:")
	source.Passive.KillerAirborneDaggers = append(source.Passive.KillerAirborneDaggers, world.KillerAirborneDaggerState{
		EffectID:           effectID,
		Position:           position,
		Direction:          world.Vector2{X: dx, Y: dy},
		LandsAt:            landsAt,
		GroundEffectKind:   "killer_q_dagger",
		GroundEffectPrefix: "effect:killer_q_dagger:",
	})
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         "killer_q_dagger_airborne",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		Start:        target.Position,
		End:          position,
		Dir:          world.Vector2{X: dx, Y: dy},
		Radius:       skillMeta(skill, "daggerRadius", 32),
		CreatedAt:    tick,
		ExpiresAt:    landsAt,
	})
}

func putGroundDagger(w *world.World, source *world.Entity, position world.Vector2, direction world.Vector2, effectKind string, effectPrefix string, tick uint64, tickRate int) {
	if w == nil || source == nil || tickRate <= 0 {
		return
	}
	passive := w.SkillConfig(passiveID)
	expiresAt := tick + secondsToTicks(skillMeta(passive, "daggerDurationSeconds", 4), tickRate)
	effectID := w.NextEffectID(effectPrefix)
	source.Passive.KillerDaggers = append(source.Passive.KillerDaggers, world.KillerDaggerState{
		EffectID:  effectID,
		Position:  position,
		ExpiresAt: expiresAt,
	})
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         effectKind,
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		Start:        position,
		Dir:          direction,
		Radius:       skillMeta(w.SkillConfig(qID), "daggerRadius", 32),
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func nextQBounceTarget(w *world.World, source *world.Entity, current *world.Entity, projectile *world.Projectile, bounceRange float64) *world.Entity {
	var best *world.Entity
	bestDistance := math.MaxFloat64
	for _, target := range w.TargetsInRadius(source, current.Position, bounceRange) {
		if target == nil || projectile.HitIDs[target.ID] {
			continue
		}
		dist := distance(current.Position, target.Position)
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	}
	return best
}

func targetedEnemy(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) *world.Entity {
	castRange := skill.Range
	if castRange <= 0 {
		castRange = 625
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !validQTarget(entity, target, castRange) {
			return
		}
		dist := distance(point, target.Position)
		if dist >= bestDistance {
			return
		}
		best = target
		bestDistance = dist
	})
	return best
}

func validQTarget(source *world.Entity, target *world.Entity, castRange float64) bool {
	return world.CanAttackTarget(source, target) && distance(source.Position, target.Position) <= castRange+target.Radius
}

func markVoracityTarget(w *world.World, source *world.Entity, target *world.Entity, _ bool, _ bool, _ bool, tick uint64, tickRate int) {
	if w == nil || source == nil || source.HeroID != heroID || target == nil ||
		!world.IsHeroUnit(target) || source.Team == target.Team || target.Combat.LastDamage <= 0 || tickRate <= 0 {
		return
	}
	if source.Passive.KillerVoracityMarks == nil {
		source.Passive.KillerVoracityMarks = make(map[string]world.KillerVoracityMark)
	}
	windowSeconds := skillMeta(w.SkillConfig(passiveID), "damageWindowSeconds", 3)
	source.Passive.KillerVoracityMarks[target.ID] = world.KillerVoracityMark{
		DamagedAt: tick,
		ExpiresAt: tick + secondsToTicks(windowSeconds, tickRate),
		TickRate:  tickRate,
	}
}

func triggerVoracity(w *world.World, _ *world.Entity, target *world.Entity) {
	if w == nil || target == nil || !world.IsHeroUnit(target) {
		return
	}
	deathTick := target.Combat.LastHitTick
	w.ForEachEntity(func(entity *world.Entity) {
		if entity == nil || entity.HeroID != heroID || entity.Team == target.Team || entity.Passive.KillerVoracityMarks == nil {
			return
		}
		mark, ok := entity.Passive.KillerVoracityMarks[target.ID]
		delete(entity.Passive.KillerVoracityMarks, target.ID)
		if !ok || deathTick < mark.DamagedAt || deathTick > mark.ExpiresAt || mark.TickRate <= 0 {
			return
		}

		for _, skillID := range []string{qID, wID, eID} {
			state, exists := entity.Skills[skillID]
			if !exists {
				continue
			}
			state.CooldownUntilTick = deathTick
			entity.Skills[skillID] = state
		}

		state, exists := entity.Skills[rID]
		if !exists || state.CooldownUntilTick <= deathTick {
			return
		}
		refundSeconds := skillMeta(w.SkillConfig(passiveID), "ultimateCooldownRefundSeconds", 15)
		refundTicks := secondsToTicks(refundSeconds, mark.TickRate)
		if state.CooldownUntilTick <= deathTick+refundTicks {
			state.CooldownUntilTick = deathTick
		} else {
			state.CooldownUntilTick -= refundTicks
		}
		entity.Skills[rID] = state
	})
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if value, ok := skill.Meta[key]; ok {
		return value
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

func skillCurve(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	current := float64(level)
	if current <= levels[0] {
		return values[0]
	}
	last := len(values) - 1
	if current >= levels[last] {
		return values[last]
	}
	for index := 1; index < len(levels); index++ {
		if current > levels[index] {
			continue
		}
		span := levels[index] - levels[index-1]
		if span <= 0 {
			return values[index]
		}
		progress := (current - levels[index-1]) / span
		return values[index-1] + (values[index]-values[index-1])*progress
	}
	return values[last]
}

func steppedValue(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	value := values[0]
	for index := 1; index < len(levels); index++ {
		if float64(level) < levels[index] {
			break
		}
		value = values[index]
	}
	return value
}

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 || tickRate <= 0 {
		return 0
	}
	haste := 0.0
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		haste = entity.Stats.AbilityHaste
	}
	seconds := float64(cooldownMS) / 1000 / (1 + haste/100)
	return secondsToTicks(seconds, tickRate)
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

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}
