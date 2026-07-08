package firemage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"strconv"
)

const (
	heroID    = "fire_mage"
	passiveID = "fire_mage_passive"
	qID       = "fire_mage_q"
	wID       = "fire_mage_w"
	eID       = "fire_mage_e"
	rID       = "fire_mage_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		TickEntity:       Tick,
		OnDamage:         OnDamage,
		OnKill:           OnKill,
		ActiveBuffs:      ActiveBuffs,
		ReleasePreparedR: ReleasePreparedW,
		CancelPreparedR:  CancelPreparedW,
	})
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.FireQPending || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 50)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{8000, 7500, 7000, 6500, 6000})), tickRate)
	entity.Skills[qID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.FireQPending = true
	entity.Passive.FireQReleaseTick = tick + windupTicks
	entity.Passive.FireQTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Passive.FireQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.FireQReleaseTick
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FireQPending || tick < entity.Passive.FireQReleaseTick {
		return
	}
	target := entity.Passive.FireQTarget
	level := entity.Passive.FireQLevel
	entity.Passive.FireQPending = false
	entity.Passive.FireQReleaseTick = 0
	entity.Passive.FireQTarget = world.Vector2{}
	entity.Passive.FireQLevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	qRange := skill.Range
	if qRange <= 0 {
		qRange = 1050
	}
	speed := skillMeta(skill, "projectileSpeed", 3000)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:fire_mage_q:"),
		Kind:         "fire_mage_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 28),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.FireRPending || tickRate <= 0 {
		return
	}
	target := targetedEnemy(w, entity, cast, skill, 750)
	if target == nil {
		return
	}
	cost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{105000, 90000, 75000})), tickRate)
	entity.Skills[rID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.2), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.FireRPending = true
	entity.Passive.FireRReleaseTick = tick + windupTicks
	entity.Passive.FireRTargetID = target.ID
	entity.Passive.FireRLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.FireRReleaseTick
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FireRPending || tick < entity.Passive.FireRReleaseTick {
		return
	}
	targetID := entity.Passive.FireRTargetID
	level := entity.Passive.FireRLevel
	entity.Passive.FireRPending = false
	entity.Passive.FireRReleaseTick = 0
	entity.Passive.FireRTargetID = ""
	entity.Passive.FireRLevel = 0
	target := w.EntityByID(targetID)
	if entity.Stats.HP <= 0 || entity.Death.Dead || !world.CanAttackTarget(entity, target) {
		return
	}
	skill := w.SkillConfig(rID)
	speed := skillMeta(skill, "projectileSpeed", 1800)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:fire_mage_r:"),
		Kind:         "fire_mage_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      rID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        distance(entity.Position, target.Position) + target.Radius,
		Radius:       skillMeta(skill, "projectileRadius", 36),
		Damage:       level,
		MagicDamage:  int(skillMeta(skill, "maxBounces", 4)),
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.FireWPending || entity.Passive.FireWCastPending || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{70, 80, 90, 100, 110})
	if entity.Stats.MP < cost {
		return
	}
	center := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	if skill.Range > 0 && distance(entity.Position, center) > skill.Range {
		dx, dy := normalize(center.X-entity.Position.X, center.Y-entity.Position.Y)
		if dx == 0 && dy == 0 {
			return
		}
		castPosition := w.ClampWorldPoint(world.Vector2{X: center.X - dx*skill.Range, Y: center.Y - dy*skill.Range})
		entity.Passive.FireWCastPending = true
		entity.Passive.FireWCastTarget = center
		entity.Passive.FireWCastLevel = state.Level
		entity.Intent.MoveTarget = &castPosition
		entity.Intent.AttackTargetID = ""
		entity.Intent.AttackPausedTill = 0
		return
	}
	startW(w, entity, center, state, skill, tick, tickRate)
}

func startW(w *world.World, entity *world.Entity, center world.Vector2, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	cost := skillList(skill, "manaCost", state.Level, []float64{70, 80, 90, 100, 110})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{10000, 9500, 9000, 8500, 8000})), tickRate)
	entity.Skills[wID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	triggerTicks := windupTicks + secondsToTicks(skillMeta(skill, "landingDelaySeconds", 0.5), tickRate)
	if triggerTicks < 1 {
		triggerTicks = 1
	}
	entity.Passive.FireWPending = true
	entity.Passive.FireWTriggerTick = tick + triggerTicks
	entity.Passive.FireWCenter = center
	entity.Passive.FireWLevel = state.Level
	entity.Control.ActionLockedUntilTick = tick + windupTicks
	w.LockAttackAfterCast(entity, tick, tickRate)
	showWEffect(w, entity, center, skill, tick, entity.Passive.FireWTriggerTick+2)
}

func ReleasePreparedW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FireWCastPending {
		return
	}
	state := entity.Skills[wID]
	if state.Level <= 0 || state.CooldownUntilTick > tick {
		CancelPreparedW(entity)
		return
	}
	state.Level = entity.Passive.FireWCastLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	target := entity.Passive.FireWCastTarget
	CancelPreparedW(entity)
	startW(w, entity, target, state, w.SkillConfig(wID), tick, tickRate)
}

func CancelPreparedW(entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FireWCastPending {
		return
	}
	entity.Passive.FireWCastPending = false
	entity.Passive.FireWCastTarget = world.Vector2{}
	entity.Passive.FireWCastLevel = 0
}

func triggerW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.FireWPending || tick < entity.Passive.FireWTriggerTick {
		return
	}
	center := entity.Passive.FireWCenter
	level := entity.Passive.FireWLevel
	entity.Passive.FireWPending = false
	entity.Passive.FireWTriggerTick = 0
	entity.Passive.FireWCenter = world.Vector2{}
	entity.Passive.FireWLevel = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	skill := w.SkillConfig(wID)
	rawDamage := skillList(skill, "baseDamage", level, []float64{75, 120, 165, 210, 255}) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.6)
	for _, target := range w.TargetsInRadius(entity, center, skillMeta(skill, "landingRadius", 260)) {
		damageRaw := rawDamage
		if burningFrom(target, entity.ID, tick) {
			damageRaw *= skillMeta(skill, "burningDamageMultiplier", 1.25)
		}
		damage := w.MagicDamageAfterResistance(entity, target, damageRaw, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			applyBurn(w, entity, target, tick, tickRate)
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
}

func showWEffect(w *world.World, entity *world.Entity, center world.Vector2, skill config.SkillConfig, tick uint64, expiresAt uint64) {
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:fire_mage_w:"),
		Kind:         "fire_mage_w",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        center,
		Radius:       skillMeta(skill, "landingRadius", 260),
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	target := eTarget(w, entity, cast, skill)
	if target == nil {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{70, 75, 80, 85, 90})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{12000, 11000, 10000, 9000, 8000})), tickRate)
	entity.Skills[eID] = state
	wasBurning := burningFrom(target, entity.ID, tick)
	damage := w.MagicDamageAfterResistance(entity, target, eRawDamage(entity, skill, state.Level), tick)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
		applyBurn(w, entity, target, tick, tickRate)
	} else {
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	if wasBurning {
		spreadBurn(w, entity, target, skill, tick, tickRate)
	}
	showEEffect(w, entity, target.Position, skill, tick, tickRate)
}

func eTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) *world.Entity {
	return targetedEnemy(w, entity, cast, skill, 625)
}

func targetedEnemy(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig, fallbackRange float64) *world.Entity {
	castRange := skill.Range
	if castRange <= 0 {
		castRange = fallbackRange
	}
	if target := w.EntityByID(cast.TargetID); validETarget(entity, target, castRange) {
		return target
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	padding := skillMeta(skill, "targetPickPadding", 80)
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !validETarget(entity, target, castRange) {
			return
		}
		pointDistance := distance(point, target.Position)
		if pointDistance > target.Radius+padding || pointDistance >= bestDistance {
			return
		}
		best = target
		bestDistance = pointDistance
	})
	return best
}

func validETarget(entity *world.Entity, target *world.Entity, castRange float64) bool {
	return world.CanAttackTarget(entity, target) && distance(entity.Position, target.Position) <= castRange+target.Radius
}

func eRawDamage(entity *world.Entity, skill config.SkillConfig, level int) float64 {
	return skillList(skill, "baseDamage", level, []float64{70, 95, 120, 145, 170}) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.5)
}

func spreadBurn(w *world.World, entity *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	for _, spreadTarget := range w.TargetsInRadius(entity, target.Position, skillMeta(skill, "spreadRadius", 600)) {
		if spreadTarget.ID == target.ID {
			continue
		}
		applyBurn(w, entity, spreadTarget, tick, tickRate)
	}
}

func showEEffect(w *world.World, entity *world.Entity, center world.Vector2, skill config.SkillConfig, tick uint64, tickRate int) {
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:fire_mage_e:"),
		Kind:         "fire_mage_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        center,
		Radius:       skillMeta(skill, "spreadRadius", 600),
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(0.25, tickRate),
	})
}

func OnDamage(w *world.World, source *world.Entity, target *world.Entity, basicAttack bool, pet bool, skipBleed bool, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID || basicAttack || pet || !world.CanAttackTarget(source, target) || tickRate <= 0 {
		return
	}
	applyBurn(w, source, target, tick, tickRate)
}

func applyBurn(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	skill := passiveSkill(w, source)
	maxStacks := int(skillMeta(skill, "maxStacks", 3))
	if maxStacks <= 0 {
		return
	}
	if target.Passive.FireBurns == nil {
		target.Passive.FireBurns = map[string]world.FireBurnState{}
	}
	burn := target.Passive.FireBurns[source.ID]
	if burn.Stacks < maxStacks {
		burn.Stacks++
	}
	burn.ExpiresAtTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate)
	if burn.NextTick == 0 {
		burn.NextTick = tick + secondsToTicks(skillMeta(skill, "tickSeconds", 1), tickRate)
	}
	// ponytail: no large-monster flag yet; narrow this when monster size exists.
	if burn.Stacks >= maxStacks && burn.ExplosionAtTick == 0 && (world.IsHeroUnit(target) || world.IsMonster(target)) {
		burn.ExplosionAtTick = tick + secondsToTicks(skillMeta(skill, "explosionDelaySeconds", 2), tickRate)
	}
	target.Passive.FireBurns[source.ID] = burn
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || tickRate <= 0 {
		return
	}
	if entity.HeroID == heroID {
		releaseQ(w, entity, tick, tickRate)
		releaseR(w, entity, tick, tickRate)
		triggerW(w, entity, tick, tickRate)
		tickManaRestore(w, entity, tick, tickRate)
	}
	tickBurns(w, entity, tick, tickRate)
}

func tickBurns(w *world.World, target *world.Entity, tick uint64, tickRate int) {
	if len(target.Passive.FireBurns) == 0 {
		return
	}
	for sourceID, burn := range target.Passive.FireBurns {
		source := w.EntityByID(sourceID)
		if source == nil || target.Stats.HP <= 0 || tick >= burn.ExpiresAtTick {
			delete(target.Passive.FireBurns, sourceID)
			continue
		}
		if burn.ExplosionAtTick > 0 && tick >= burn.ExplosionAtTick {
			explode(w, source, target, tick, tickRate)
			delete(target.Passive.FireBurns, sourceID)
			continue
		}
		if tick < burn.NextTick || burn.Stacks <= 0 {
			continue
		}
		skill := passiveSkill(w, source)
		tickSeconds := skillMeta(skill, "tickSeconds", 1)
		raw := target.Stats.MaxHP * skillMeta(skill, "burnMaxHPRatioPerSecond", 0.02) * float64(burn.Stacks) * tickSeconds
		damage := w.MagicDamageAfterResistance(source, target, raw, tick)
		burn.NextTick += secondsToTicks(tickSeconds, tickRate)
		target.Passive.FireBurns[sourceID] = burn
		applyPassiveDamage(w, source, target, damage, tick, tickRate)
	}
}

func explode(w *world.World, source *world.Entity, center *world.Entity, tick uint64, tickRate int) {
	skill := passiveSkill(w, source)
	damage := skillCurve(skill, "explosionDamage", "explosionDamageLevels", source.Level, 120)
	for _, target := range w.TargetsInRadius(source, center.Position, skillMeta(skill, "explosionRadius", 400)) {
		applyPassiveDamage(w, source, target, w.MagicDamageAfterResistance(source, target, damage, tick), tick, tickRate)
	}
}

func applyPassiveDamage(w *world.World, source *world.Entity, target *world.Entity, damage int, tick uint64, tickRate int) {
	if damage <= 0 || target == nil {
		return
	}
	target.Combat.LastHitTick = tick
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
		return
	}
	wasAlive := target.Stats.HP > 0
	w.ApplyPetDamage(source, target, damage, "magic", tickRate)
	if target.Combat.LastDamage > 0 {
		w.ApplyEquipmentSkillBurn(source, target, tick, tickRate)
	}
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(source, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}

func OnKill(w *world.World, killer *world.Entity, target *world.Entity) {
	if killer == nil || target == nil || killer.HeroID != heroID || !burningFrom(target, killer.ID, target.Combat.LastHitTick) {
		return
	}
	skill := passiveSkill(w, killer)
	tickRate := target.Death.RespawnTickRate
	if tickRate <= 0 {
		tickRate = 20
	}
	killer.Passive.FireManaUntil = target.Combat.LastHitTick + secondsToTicks(skillMeta(skill, "manaRestoreSeconds", 10), tickRate)
	killer.Passive.FireManaNextTick = target.Combat.LastHitTick + secondsToTicks(skillMeta(skill, "manaRestoreTickSeconds", 5), tickRate)
}

func tickManaRestore(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity.Passive.FireManaUntil == 0 || tick < entity.Passive.FireManaNextTick {
		return
	}
	skill := passiveSkill(w, entity)
	for entity.Passive.FireManaNextTick <= tick && entity.Passive.FireManaNextTick <= entity.Passive.FireManaUntil {
		entity.Stats.MP += skillCurve(skill, "manaRestore", "manaRestoreLevels", entity.Level, 30)
		if entity.Stats.MP > entity.Stats.MaxMP {
			entity.Stats.MP = entity.Stats.MaxMP
		}
		entity.Passive.FireManaNextTick += secondsToTicks(skillMeta(skill, "manaRestoreTickSeconds", 5), tickRate)
	}
	if tick >= entity.Passive.FireManaUntil {
		entity.Passive.FireManaUntil = 0
		entity.Passive.FireManaNextTick = 0
	}
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil {
		return nil
	}
	buffs := make([]world.BuffState, 0, len(entity.Passive.FireBurns)+1)
	if entity.HeroID == heroID && entity.Passive.FireManaUntil > tick {
		buffs = append(buffs, world.BuffState{ID: "fire_mage_mana_restore", Name: "Blaze Mana", ExpiresAtTick: entity.Passive.FireManaUntil})
	}
	for sourceID, burn := range entity.Passive.FireBurns {
		if burn.Stacks <= 0 || tick >= burn.ExpiresAtTick {
			continue
		}
		buffs = append(buffs, world.BuffState{
			ID:              "fire_mage_blaze:" + sourceID,
			Name:            "Blaze " + strconv.Itoa(burn.Stacks),
			ExpiresAtTick:   burn.ExpiresAtTick,
			ExplosionAtTick: burn.ExplosionAtTick,
			Negative:        true,
		})
	}
	return buffs
}

func burningFrom(target *world.Entity, sourceID string, tick uint64) bool {
	if target == nil || target.Passive.FireBurns == nil {
		return false
	}
	burn := target.Passive.FireBurns[sourceID]
	return burn.Stacks > 0 && tick < burn.ExpiresAtTick
}

func passiveSkill(w *world.World, entity *world.Entity) config.SkillConfig {
	if w == nil || entity == nil {
		return config.SkillConfig{}
	}
	return w.SkillConfig(passiveID)
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
	currentLevel := float64(level)
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
		return values[i-1] + (values[i]-values[i-1])*(currentLevel-fromLevel)/(toLevel-fromLevel)
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
