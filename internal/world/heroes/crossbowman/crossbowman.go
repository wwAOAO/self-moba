package crossbowman

import (
	"fmt"
	"math"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

const (
	heroID    = "crossbowman"
	passiveID = "crossbowman_night_hunter"
	qID       = "crossbowman_tumble"
	wID       = "crossbowman_silver_bolts"
	eID       = "crossbowman_condemn"
	rID       = "crossbowman_final_hour"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			eID: CastE,
			rID: CastR,
		},
		Tick:                           Tick,
		OnBasicHit:                     OnBasicHit,
		OnDamage:                       OnDamage,
		OnSkillHit:                     ApplySilverBoltsHit,
		OnKill:                         OnKill,
		ApplyStats:                     ApplyStats,
		BasicAttackBonusPhysicalDamage: QBonusPhysicalDamage,
		MoveSpeedBonus:                 MoveSpeedBonus,
		ActiveBuffs:                    ActiveBuffs,
		ResolveProjectile:              ResolveProjectile,
	})
}

func CastR(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 80)
	if entity.Stats.MP < cost {
		return
	}

	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{100000, 85000, 70000})), tickRate)
	entity.Skills[rID] = state
	entity.Crossbowman.UltimateLevel = state.Level
	entity.Crossbowman.UltimateTickRate = tickRate
	entity.Crossbowman.UltimateStartedAtTick = tick
	entity.Crossbowman.UltimateUntilTick = tick + secondsToTicks(skillList(skill, "durationSeconds", state.Level, []float64{8, 10, 12}), tickRate)
	entity.Crossbowman.UltimateDamagedHeroes = make(map[string]uint64)
	entity.Crossbowman.UltimateEffectID = w.NextEffectID("effect:crossbowman_final_hour:")
	putUltimateEffect(w, entity)
	w.RefreshPlayerStats(entity)
}

func OnDamage(w *world.World, source *world.Entity, target *world.Entity, _ bool, _ bool, _ bool, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || source.HeroID != heroID || tickRate <= 0 ||
		tick >= source.Crossbowman.UltimateUntilTick || target.ID == "" || !world.IsHeroUnit(target) ||
		target.Team == source.Team || target.Team == world.TeamNeutral {
		return
	}
	if source.Crossbowman.UltimateDamagedHeroes == nil {
		source.Crossbowman.UltimateDamagedHeroes = make(map[string]uint64)
	}
	markSeconds := skillMeta(w.SkillConfig(rID), "damageMarkSeconds", 3)
	source.Crossbowman.UltimateDamagedHeroes[target.ID] = tick + secondsToTicks(markSeconds, tickRate)
}

func OnKill(w *world.World, _ *world.Entity, target *world.Entity) {
	if w == nil || target == nil || !world.IsHeroUnit(target) {
		return
	}
	deathTick := target.Combat.LastHitTick
	w.ForEachEntity(func(entity *world.Entity) {
		if entity == nil || entity.Kind != world.EntityKindPlayer || entity.HeroID != heroID || entity.Team == target.Team ||
			len(entity.Crossbowman.UltimateDamagedHeroes) == 0 {
			return
		}
		markUntil, marked := entity.Crossbowman.UltimateDamagedHeroes[target.ID]
		delete(entity.Crossbowman.UltimateDamagedHeroes, target.ID)
		if !marked || deathTick >= markUntil || deathTick >= entity.Crossbowman.UltimateUntilTick {
			return
		}
		tickRate := entity.Crossbowman.UltimateTickRate
		if tickRate <= 0 {
			tickRate = 20
		}
		entity.Crossbowman.UltimateUntilTick += secondsToTicks(skillMeta(w.SkillConfig(rID), "extensionSeconds", 4), tickRate)
		putUltimateEffect(w, entity)
	})
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || entity.HeroID != heroID || stats == nil || entity.Crossbowman.UltimateUntilTick == 0 || entity.Crossbowman.UltimateLevel <= 0 {
		return
	}
	bonus := skillList(w.SkillConfig(rID), "bonusAttack", entity.Crossbowman.UltimateLevel, []float64{35, 50, 65})
	stats.Attack += bonus
	stats.BonusAttack += bonus
}

func putUltimateEffect(w *world.World, entity *world.Entity) {
	if w == nil || entity == nil || entity.Crossbowman.UltimateEffectID == "" {
		return
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           entity.Crossbowman.UltimateEffectID,
		Kind:         "crossbowman_final_hour",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		End:          entity.Position,
		Radius:       entity.Radius,
		Count:        entity.Crossbowman.UltimateLevel,
		CreatedAt:    entity.Crossbowman.UltimateStartedAtTick,
		ExpiresAt:    entity.Crossbowman.UltimateUntilTick,
	})
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 ||
		entity.Crossbowman.TumbleEmpowerUntilTick > tick || entity.Crossbowman.TumblePendingCooldownTicks > 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 30)
	if entity.Stats.MP < cost {
		return
	}

	dx := cast.TargetX - entity.Position.X
	dy := cast.TargetY - entity.Position.Y
	distance := math.Hypot(dx, dy)
	if distance == 0 {
		dx = 1
		distance = 1
	}
	start := entity.Position
	rangeValue := skill.Range
	if rangeValue <= 0 {
		rangeValue = 300
	}
	end := w.ResolveCollisionPosition(entity, world.Vector2{
		X: start.X + dx/distance*rangeValue,
		Y: start.Y + dy/distance*rangeValue,
	})

	entity.Stats.MP -= cost
	entity.Position = end
	entity.Intent.MoveTarget = nil
	entity.Crossbowman.TumbleEmpowerUntilTick = tick + secondsToTicks(skillMeta(skill, "empowerDurationSeconds", 6), tickRate)
	entity.Crossbowman.TumbleLevel = state.Level
	entity.Combat.NextAttackTick = tick
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0

	cooldownMS := skillList(skill, "cooldownMs", state.Level, []float64{6000, 5000, 4000, 3000, 2000})
	if tick < entity.Crossbowman.UltimateUntilTick {
		ultimateLevel := entity.Crossbowman.UltimateLevel
		if ultimateLevel <= 0 {
			ultimateLevel = entity.Skills[rID].Level
		}
		if ultimateLevel > 0 {
			reduction := skillList(skill, "ultimateCooldownReduction", ultimateLevel, []float64{0.3, 0.4, 0.5})
			cooldownMS *= 1 - math.Min(reduction, 1)
		}
		entity.Control.InvisibleUntilTick = tick + secondsToTicks(skillMeta(skill, "invisibilitySeconds", 1), tickRate)
	}
	entity.Crossbowman.TumblePendingCooldownTicks = cooldownTicksFor(entity, int(math.Round(cooldownMS)), tickRate)
	state.CooldownUntilTick = 0
	entity.Skills[qID] = state

	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:crossbowman_tumble:"),
		Kind:         "crossbowman_tumble",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        start,
		End:          end,
		Radius:       entity.Radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "effectSeconds", 0.35), tickRate),
	})
}

func QBonusPhysicalDamage(w *world.World, attacker *world.Entity, _ *world.Entity, tick uint64, _ int) int {
	if w == nil || attacker == nil || attacker.HeroID != heroID || tick >= attacker.Crossbowman.TumbleEmpowerUntilTick {
		return 0
	}
	level := attacker.Crossbowman.TumbleLevel
	if level <= 0 {
		return 0
	}
	ratio := skillList(w.SkillConfig(qID), "totalAdRatio", level, []float64{0.75, 0.85, 0.95, 1.05, 1.15})
	return int(math.Round(attacker.Stats.Attack * ratio))
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Crossbowman.CondemnPending || tickRate <= 0 {
		return
	}
	target := condemnTarget(w, entity, cast, skill.Range)
	cost := skillMeta(skill, "manaCost", 90)
	if target == nil || entity.Stats.MP < cost {
		return
	}

	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{20000, 18000, 16000, 14000, 12000})), tickRate)
	entity.Skills[eID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.2), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Crossbowman.CondemnPending = true
	entity.Crossbowman.CondemnReleaseTick = tick + windupTicks
	entity.Crossbowman.CondemnTargetID = target.ID
	entity.Crossbowman.CondemnLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Crossbowman.CondemnReleaseTick
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func ResolveProjectile(w *world.World, source *world.Entity, projectile *world.Projectile, previousPosition world.Vector2, tick uint64, tickRate int) bool {
	if projectile == nil || projectile.SkillID != eID {
		return false
	}
	target := w.EntityByID(projectile.TargetID)
	if source == nil || target == nil || target.Stats.HP <= 0 || target.Death.Dead {
		w.RemoveProjectile(projectile.ID)
		return true
	}
	if !condemnProjectileIntersects(projectile, previousPosition, target) {
		return true
	}

	w.RemoveProjectile(projectile.ID)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	damage := condemnDamage(w, source, target, w.SkillConfig(eID), projectile.Damage, tick)
	wasAlive := target.Stats.HP > 0
	applyCondemnDamage(w, source, target, damage, tickRate)
	ApplySilverBoltsHit(w, source, target, tick, tickRate)
	if target.Kind != world.EntityKindDummy && wasAlive && target.Stats.HP == 0 {
		finishCondemnKill(w, source, target, tick, tickRate)
		return true
	}
	startCondemnKnockback(w, source, target, projectile.Dir, damage, tick, tickRate)
	return true
}

func releaseCondemn(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || !entity.Crossbowman.CondemnPending || tick < entity.Crossbowman.CondemnReleaseTick || tickRate <= 0 {
		return
	}
	targetID := entity.Crossbowman.CondemnTargetID
	level := entity.Crossbowman.CondemnLevel
	entity.Crossbowman.CondemnPending = false
	entity.Crossbowman.CondemnReleaseTick = 0
	entity.Crossbowman.CondemnTargetID = ""
	entity.Crossbowman.CondemnLevel = 0
	target := w.EntityByID(targetID)
	if entity.Stats.HP <= 0 || entity.Death.Dead || !world.CanAttackTarget(entity, target) {
		return
	}

	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(eID)
	speed := skillMeta(skill, "projectileSpeed", 2200)
	lifetime := skillMeta(skill, "projectileLifetimeSeconds", 2)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:crossbowman_condemn:"),
		Kind:         "crossbowman_condemn",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      eID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        speed * lifetime,
		Radius:       skillMeta(skill, "projectileRadius", 18),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(lifetime, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func startCondemnKnockback(w *world.World, source *world.Entity, target *world.Entity, fallbackDir world.Vector2, damage int, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || tickRate <= 0 {
		return
	}
	dx, dy := normalize(target.Position.X-source.Position.X, target.Position.Y-source.Position.Y)
	if dx == 0 && dy == 0 {
		dx, dy = fallbackDir.X, fallbackDir.Y
	}
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(eID)
	knockbackDistance := skillMeta(skill, "knockbackDistance", 470)
	start := target.Position
	desiredEnd := world.Vector2{X: start.X + dx*knockbackDistance, Y: start.Y + dy*knockbackDistance}
	end := w.ClampWorldPoint(desiredEnd)
	collided := distance(desiredEnd, end) > 0.000001
	actualDistance := distance(start, end)
	travelSeconds := skillMeta(skill, "knockbackSeconds", 0.25)
	if knockbackDistance > 0 {
		travelSeconds *= actualDistance / knockbackDistance
	}
	travelTicks := secondsToTicks(travelSeconds, tickRate)
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
	if target.Control.ActionLockedUntilTick < until {
		target.Control.ActionLockedUntilTick = until
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:crossbowman_condemn_knockback:"),
		Kind:         "crossbowman_condemn_knockback",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		Start:        start,
		End:          end,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    until,
	})
	if collided {
		source.Crossbowman.CondemnImpactTick = until
		source.Crossbowman.CondemnImpactTargetID = target.ID
		source.Crossbowman.CondemnImpactDamage = damage
	}
}

func resolveCondemnImpact(w *world.World, source *world.Entity, tick uint64, tickRate int) {
	if source == nil || source.Crossbowman.CondemnImpactTick == 0 || tick < source.Crossbowman.CondemnImpactTick || tickRate <= 0 {
		return
	}
	targetID := source.Crossbowman.CondemnImpactTargetID
	damage := source.Crossbowman.CondemnImpactDamage
	source.Crossbowman.CondemnImpactTick = 0
	source.Crossbowman.CondemnImpactTargetID = ""
	source.Crossbowman.CondemnImpactDamage = 0
	target := w.EntityByID(targetID)
	if target == nil || target.Stats.HP <= 0 || target.Death.Dead {
		return
	}

	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	wasAlive := target.Stats.HP > 0
	applyCondemnDamage(w, source, target, damage, tickRate)
	if target.Kind != world.EntityKindDummy && wasAlive && target.Stats.HP == 0 {
		finishCondemnKill(w, source, target, tick, tickRate)
		return
	}
	stunTicks := world.ControlTicksAfterTenacity(target, secondsToTicks(skillMeta(w.SkillConfig(eID), "stunSeconds", 1.5), tickRate), tick)
	stunUntil := tick + stunTicks
	if target.Control.StunnedUntilTick > stunUntil {
		stunUntil = target.Control.StunnedUntilTick
	}
	w.ApplyStun(target, stunUntil, tick, tickRate)
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:crossbowman_condemn_impact:"),
		Kind:         "crossbowman_condemn_impact",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		TargetID:     target.ID,
		Start:        target.Position,
		End:          target.Position,
		Radius:       target.Radius,
		Count:        damage,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(w.SkillConfig(eID), "impactEffectSeconds", 0.5), tickRate),
	})
}

func condemnDamage(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	base := skillList(skill, "baseDamage", level, []float64{50, 85, 120, 155, 190})
	raw := base
	if source != nil {
		raw += source.Stats.BonusAttack * skillMeta(skill, "bonusAdRatio", 0.5)
	}
	return w.PhysicalDamageAfterResistance(source, target, raw, tick)
}

func applyCondemnDamage(w *world.World, source *world.Entity, target *world.Entity, damage int, tickRate int) {
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
		return
	}
	w.ApplyDamage(source, target, damage, tickRate)
}

func finishCondemnKill(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	w.ApplyKillReward(source, target)
	w.KillPlayer(target, tick, tickRate)
	w.RemoveDeadUnit(target)
}

func condemnProjectileIntersects(projectile *world.Projectile, previousPosition world.Vector2, target *world.Entity) bool {
	dx := projectile.Position.X - previousPosition.X
	dy := projectile.Position.Y - previousPosition.Y
	lengthSquared := dx*dx + dy*dy
	t := 0.0
	if lengthSquared > 0 {
		t = ((target.Position.X-previousPosition.X)*dx + (target.Position.Y-previousPosition.Y)*dy) / lengthSquared
		t = math.Max(0, math.Min(1, t))
	}
	closestX := previousPosition.X + dx*t
	closestY := previousPosition.Y + dy*t
	return math.Hypot(target.Position.X-closestX, target.Position.Y-closestY) <= projectile.Radius+target.Radius
}

func condemnTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, castRange float64) *world.Entity {
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

func ConsumeQ(_ *world.World, attacker *world.Entity, _ *world.Entity, tick uint64, _ int) {
	if attacker == nil || attacker.HeroID != heroID || tick >= attacker.Crossbowman.TumbleEmpowerUntilTick {
		return
	}
	attacker.Crossbowman.TumbleEmpowerUntilTick = 0
	attacker.Crossbowman.TumbleLevel = 0
	startTumbleCooldown(attacker, tick)
}

func startTumbleCooldown(entity *world.Entity, startTick uint64) {
	if entity == nil || entity.Crossbowman.TumblePendingCooldownTicks == 0 {
		return
	}
	state := entity.Skills[qID]
	state.CooldownUntilTick = startTick + entity.Crossbowman.TumblePendingCooldownTicks
	entity.Skills[qID] = state
	entity.Crossbowman.TumblePendingCooldownTicks = 0
}

func OnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	ConsumeQ(w, attacker, target, tick, tickRate)
	ApplySilverBoltsHit(w, attacker, target, tick, tickRate)
}

func ApplySilverBoltsHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || source.HeroID != heroID || target.Stats.HP <= 0 ||
		!world.CanAttackTarget(source, target) {
		return
	}
	state := source.Skills[wID]
	if state.Level <= 0 {
		return
	}
	if source.Crossbowman.SilverBoltsTargetID != target.ID {
		clearSilverBolts(w, source)
		source.Crossbowman.SilverBoltsTargetID = target.ID
		source.Crossbowman.SilverBoltsEffectID = w.NextEffectID("effect:crossbowman_silver_bolts:")
	}
	source.Crossbowman.SilverBoltsStacks++
	source.Crossbowman.SilverBoltsUntilTick = tick + secondsToTicks(skillMeta(w.SkillConfig(wID), "stackDurationSeconds", 5), tickRate)
	if source.Crossbowman.SilverBoltsStacks < 3 {
		showSilverBoltsEffect(w, source, target, source.Crossbowman.SilverBoltsStacks, false, 0, tick, tickRate)
		return
	}

	damage := silverBoltsDamage(target, w.SkillConfig(wID), state.Level)
	clearSilverBolts(w, source)
	target.Combat.LastHitTick = tick
	wasAlive := target.Stats.HP > 0
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = w.TrueDamageAfterReduction(target, damage, tick)
		target.Combat.LastDamageType = "true"
	} else {
		w.ApplyTrueDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	showSilverBoltsEffect(w, source, target, 3, true, int(math.Round(damage)), tick, tickRate)
}

func silverBoltsDamage(target *world.Entity, skill config.SkillConfig, level int) float64 {
	minimum := skillList(skill, "minimumDamage", level, []float64{50, 65, 80, 95, 110})
	ratio := skillList(skill, "maxHealthRatio", level, []float64{0.06, 0.07, 0.08, 0.09, 0.1})
	damage := minimum
	if target != nil {
		damage = math.Max(damage, target.Stats.MaxHP*ratio)
		if world.IsMonster(target) {
			capValue := skillList(skill, "monsterDamageCap", level, []float64{140, 155, 170, 185, 200})
			damage = math.Min(damage, capValue)
		}
	}
	return damage
}

func showSilverBoltsEffect(w *world.World, source *world.Entity, target *world.Entity, stacks int, proc bool, damage int, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil {
		return
	}
	skill := w.SkillConfig(wID)
	kind := "crossbowman_silver_bolts"
	effectID := source.Crossbowman.SilverBoltsEffectID
	expiresAt := source.Crossbowman.SilverBoltsUntilTick
	count := stacks
	if proc {
		kind = "crossbowman_silver_bolts_proc"
		effectID = w.NextEffectID("effect:" + kind + ":")
		expiresAt = tick + secondsToTicks(skillMeta(skill, "procEffectSeconds", 0.45), tickRate)
		count = damage
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         kind,
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		TargetID:     target.ID,
		Start:        target.Position,
		End:          target.Position,
		Radius:       target.Radius,
		Count:        count,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func clearSilverBolts(w *world.World, source *world.Entity) {
	if source == nil {
		return
	}
	if w != nil {
		w.RemoveSkillEffect(source.Crossbowman.SilverBoltsEffectID)
	}
	source.Crossbowman.SilverBoltsTargetID = ""
	source.Crossbowman.SilverBoltsStacks = 0
	source.Crossbowman.SilverBoltsUntilTick = 0
	source.Crossbowman.SilverBoltsEffectID = ""
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || tickRate <= 0 {
		return
	}
	releaseCondemn(w, entity, tick, tickRate)
	resolveCondemnImpact(w, entity, tick, tickRate)
	if entity.Crossbowman.UltimateUntilTick > 0 && tick >= entity.Crossbowman.UltimateUntilTick {
		entity.Crossbowman.UltimateUntilTick = 0
		entity.Crossbowman.UltimateLevel = 0
		entity.Crossbowman.UltimateTickRate = 0
		entity.Crossbowman.UltimateStartedAtTick = 0
		entity.Crossbowman.UltimateEffectID = ""
		entity.Crossbowman.UltimateDamagedHeroes = nil
		w.RefreshPlayerStats(entity)
	}
	if entity.Crossbowman.TumbleEmpowerUntilTick > 0 && tick >= entity.Crossbowman.TumbleEmpowerUntilTick {
		expiredAt := entity.Crossbowman.TumbleEmpowerUntilTick
		entity.Crossbowman.TumbleEmpowerUntilTick = 0
		entity.Crossbowman.TumbleLevel = 0
		startTumbleCooldown(entity, expiredAt)
	}
	if entity.Crossbowman.SilverBoltsTargetID != "" {
		target := w.EntityByID(entity.Crossbowman.SilverBoltsTargetID)
		if tick >= entity.Crossbowman.SilverBoltsUntilTick || target == nil || target.Stats.HP <= 0 || target.Death.Dead {
			clearSilverBolts(w, entity)
		}
	}
	destination, moving := w.MovementDestination(entity, tick)
	if !moving {
		return
	}
	skill := w.SkillConfig(passiveID)
	if !hasEnemyHeroAhead(w, entity, destination, skill.Range) {
		return
	}
	entity.Crossbowman.NightHunterMoveSpeed = skillMeta(skill, "moveSpeedBonus", 30)
	entity.Crossbowman.NightHunterUltimateMoveSpeed = skillMeta(skill, "ultimateMoveSpeedBonus", 90)
	entity.Crossbowman.NightHunterUntil = tick + secondsToTicks(skillMeta(skill, "lingerSeconds", 2), tickRate)
}

func MoveSpeedBonus(entity *world.Entity, tick uint64) float64 {
	if entity == nil || entity.HeroID != heroID || tick >= entity.Crossbowman.NightHunterUntil {
		return 0
	}
	if tick < entity.Crossbowman.UltimateUntilTick {
		if entity.Crossbowman.NightHunterUltimateMoveSpeed > 0 {
			return entity.Crossbowman.NightHunterUltimateMoveSpeed
		}
		return 90
	}
	if entity.Crossbowman.NightHunterMoveSpeed > 0 {
		return entity.Crossbowman.NightHunterMoveSpeed
	}
	return 30
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil {
		return nil
	}
	buffs := make([]world.BuffState, 0, 5)
	if entity.HeroID == heroID && tick < entity.Crossbowman.UltimateUntilTick && entity.Crossbowman.UltimateLevel > 0 {
		bonus := skillList(w.SkillConfig(rID), "bonusAttack", entity.Crossbowman.UltimateLevel, []float64{35, 50, 65})
		buffs = append(buffs, world.BuffState{
			ID:            rID,
			Name:          "终极时刻",
			Tooltip:       fmt.Sprintf("+%.0f 额外攻击力", bonus),
			ExpiresAtTick: entity.Crossbowman.UltimateUntilTick,
		})
	}
	if entity.HeroID == heroID && tick < entity.Crossbowman.NightHunterUntil {
		buffs = append(buffs, world.BuffState{
			ID:            passiveID,
			Name:          "暗夜猎手",
			Tooltip:       "+" + formatMoveSpeed(MoveSpeedBonus(entity, tick)) + " 移动速度",
			ExpiresAtTick: entity.Crossbowman.NightHunterUntil,
		})
	}
	if entity.HeroID == heroID && tick < entity.Crossbowman.TumbleEmpowerUntilTick {
		buffs = append(buffs, world.BuffState{
			ID:            qID,
			Name:          "闪避突袭",
			Tooltip:       "下一次普攻造成额外物理伤害",
			ExpiresAtTick: entity.Crossbowman.TumbleEmpowerUntilTick,
		})
	}
	w.ForEachEntity(func(source *world.Entity) {
		if source == nil || source.HeroID != heroID || source.Crossbowman.SilverBoltsTargetID != entity.ID ||
			source.Crossbowman.SilverBoltsStacks <= 0 || tick >= source.Crossbowman.SilverBoltsUntilTick {
			return
		}
		buffs = append(buffs, world.BuffState{
			ID:            wID + ":" + source.ID,
			Name:          "圣银弩箭",
			Stacks:        source.Crossbowman.SilverBoltsStacks,
			Tooltip:       fmt.Sprintf("%d/3 层；第三层触发真实伤害", source.Crossbowman.SilverBoltsStacks),
			ExpiresAtTick: source.Crossbowman.SilverBoltsUntilTick,
			Negative:      true,
		})
	})
	return buffs
}

func hasEnemyHeroAhead(w *world.World, source *world.Entity, destination world.Vector2, radius float64) bool {
	if radius <= 0 {
		radius = 2000
	}
	moveX := destination.X - source.Position.X
	moveY := destination.Y - source.Position.Y
	moveLength := math.Hypot(moveX, moveY)
	if moveLength == 0 {
		return false
	}
	found := false
	w.ForEachEntity(func(target *world.Entity) {
		if found || target == nil || target.ID == source.ID || !world.IsHeroUnit(target) ||
			target.Team == source.Team || target.Team == world.TeamNeutral || target.Stats.HP <= 0 || target.Death.Dead {
			return
		}
		toTargetX := target.Position.X - source.Position.X
		toTargetY := target.Position.Y - source.Position.Y
		targetDistance := math.Hypot(toTargetX, toTargetY)
		if targetDistance == 0 || targetDistance > radius {
			return
		}
		found = moveX*toTargetX+moveY*toTargetY > 0
	})
	return found
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if configured := skill.MetaLists[key]; len(configured) > 0 {
		values = configured
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

func formatMoveSpeed(value float64) string {
	return fmt.Sprintf("%.0f", value)
}
