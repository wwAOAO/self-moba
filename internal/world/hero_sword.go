package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
	"strconv"
)

func (w *World) lockAttackAfterCast(entity *Entity, tick uint64, tickRate int) {
	nextAttackTick := tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(entity, tick), tickRate)
	if entity.Combat.NextAttackTick < nextAttackTick {
		entity.Combat.NextAttackTick = nextAttackTick
	}
}

func (w *World) applySwordQ(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.QPending {
		return
	}
	if tick > state.StacksExpireTick {
		state.Stacks = 0
	}
	hasWhirlwindStack := state.Stacks >= 2
	form := "line"
	qRange := skillRange(skill, 475)
	if swordEQWindowActive(entity, skill, tick, tickRate) {
		form = "circle"
		qRange = skillMetaRange(skill, "eqRadius", 375)
	} else if hasWhirlwindStack {
		form = "whirlwind"
		qRange = skillMetaRange(skill, "whirlwindRange", 900)
	}
	windupTicks := w.swordQWindupTicks(entity, skill, tickRate)
	entity.Sword.QPending = true
	entity.Sword.QReleaseTick = tick + windupTicks
	entity.Sword.QTarget = Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Sword.QForm = form
	entity.Sword.QRange = qRange
	entity.Control.ActionLockedUntilTick = tick + windupTicks
	entity.Skills[swordQSkillID] = state
}

func (w *World) releaseSwordQ(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != swordHeroID || !entity.Sword.QPending || tick < entity.Sword.QReleaseTick {
		return
	}
	skill := w.skillConfig(swordQSkillID)
	state := entity.Skills[swordQSkillID]
	form := entity.Sword.QForm
	qRange := entity.Sword.QRange
	targetPoint := entity.Sword.QTarget
	entity.Sword.QPending = false
	entity.Sword.QReleaseTick = 0
	entity.Sword.QTarget = Vector2{}
	entity.Sword.QForm = ""
	entity.Sword.QRange = 0
	hasWhirlwindStack := state.Stacks >= 2
	state.CooldownUntilTick = tick + w.swordQCooldownTicks(entity, skill, state.Level, tickRate)
	if form == "whirlwind" {
		w.spawnSwordWhirlwind(entity, targetPoint, qRange, skill, tick, tickRate)
		w.lockAttackAfterCast(entity, tick, tickRate)
		state.Stacks = 0
		state.StacksExpireTick = 0
		entity.Skills[swordQSkillID] = state
		return
	}
	targets := w.swordQTargets(entity, targetPoint, qRange, form, skill)
	for _, target := range targets {
		damage := w.swordQDamage(entity, target, skill, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			if form == "circle" {
				w.applyAOEDamage(entity, target, damage, "physical", tickRate)
			} else {
				w.applyDamage(entity, target, damage, tickRate)
			}
			if form == "circle" && hasWhirlwindStack {
				target.Control.AirborneUntilTick = tick + secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1), tickRate)
			}
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
		}
	}
	if len(targets) > 0 {
		if form == "circle" && hasWhirlwindStack {
			state.Stacks = 0
			state.StacksExpireTick = 0
		} else {
			state.Stacks++
			if state.Stacks > 2 {
				state.Stacks = 2
			}
			state.StacksExpireTick = tick + secondsToTicks(skillMetaRange(skill, "stackDurationSeconds", swordQStackTicks), tickRate)
		}
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordQSkillID] = state
}

func (w *World) expireSwordQStacks(entity *Entity, tick uint64) {
	if entity == nil || entity.HeroID != swordHeroID {
		return
	}
	state := entity.Skills[swordQSkillID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[swordQSkillID] = state
}

func (w *World) swordQWindupTicks(entity *Entity, skill config.SkillConfig, tickRate int) uint64 {
	seconds := swordQWindupSeconds(entity, skill)
	ticks := secondsToTicks(seconds, tickRate)
	if ticks < 1 {
		return 1
	}
	return ticks
}

func swordQWindupSeconds(entity *Entity, skill config.SkillConfig) float64 {
	base := skillMetaRange(skill, "castWindupSeconds", 0.328)
	minimum := skillMetaRange(skill, "minCastWindupSeconds", 0.09)
	bonus := 0.0
	if entity != nil {
		bonus = entity.Stats.AttackSpeedBonus
	}
	bonus = clamp(bonus, 0, math.MaxFloat64)
	seconds := base / (1 + bonus)
	if seconds < minimum {
		return minimum
	}
	return seconds
}

func swordEQWindowActive(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || tick >= entity.Control.DashUntilTick {
		return false
	}
	windowTicks := secondsToTicks(skillMetaRange(skill, "eqWindowSeconds", 0.15), tickRate)
	if windowTicks == 0 {
		return false
	}
	return entity.Control.DashUntilTick-tick <= windowTicks
}

func (w *World) applySwordW(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	w.nextWallID++
	id := "wind_wall:" + strconv.Itoa(w.nextWallID)
	width := skillMetaListByLevel(skill, "width", state.Level, []float64{300, 350, 400, 450, 500})
	placeDistance := skillMetaRange(skill, "placeDistance", 180)
	center := Vector2{
		X: clamp(entity.Position.X+dx*placeDistance, 0, w.width),
		Y: clamp(entity.Position.Y+dy*placeDistance, 0, w.height),
	}
	w.windWalls[id] = WindWall{
		ID:        id,
		Team:      entity.Team,
		Center:    center,
		Dir:       Vector2{X: -dy, Y: dx},
		Width:     width,
		ExpiresAt: tick + secondsToTicks(skillMetaRange(skill, "durationSeconds", windWallDuration), tickRate),
	}
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{26000, 24000, 22000, 20000, 18000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordWSkillID] = state
}

func (w *World) spawnSwordWhirlwind(entity *Entity, targetPoint Vector2, qRange float64, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	radius := skillMetaRange(skill, "whirlwindRadius", 70)
	speedPerSecond := skillMetaRange(skill, "whirlwindSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:sword_whirlwind:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "sword_whirlwind",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      swordQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       radius,
		Damage:       w.swordQDamage(entity, &Entity{ID: id}, skill, tick),
		KnockupTicks: secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
}

func (w *World) applySwordE(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.SweepingBladeTargetUntil == nil {
		entity.Sword.SweepingBladeTargetUntil = make(map[string]uint64)
	}
	target := w.swordETarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	if tick < entity.Sword.SweepingBladeTargetUntil[target.ID] {
		return
	}
	damage := swordEDamage(entity, target, skill, state.Level, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyMagicDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(entity, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	}
	entity.Sword.SweepingBladeStacks++
	maxStacks := int(skillMetaRange(skill, "maxStacks", 4))
	if entity.Sword.SweepingBladeStacks > maxStacks {
		entity.Sword.SweepingBladeStacks = maxStacks
	}
	targetCooldownMS := skillMetaListByLevelMS(skill, "targetCooldownMs", state.Level, []float64{10000, 9000, 8000, 7000, 6000})
	entity.Sword.SweepingBladeTargetUntil[target.ID] = tick + cooldownTicks(targetCooldownMS, tickRate)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	dashEnd := Vector2{
		X: clamp(target.Position.X+dx*(target.Radius+entity.Radius+skillMetaRange(skill, "dashThroughDistance", 34)), 0, w.width),
		Y: clamp(target.Position.Y+dy*(target.Radius+entity.Radius+skillMetaRange(skill, "dashThroughDistance", 34)), 0, w.height),
	}
	entity.Intent = IntentState{}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = dashEnd
	entity.Control.DashUntilTick = tick + secondsToTicks(skillMetaRange(skill, "dashDurationSeconds", 0.35), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{500, 400, 300, 200, 100}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordESkillID] = state
}

func (w *World) applySwordR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	target := w.swordRTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill, tick)
	if target == nil {
		return
	}
	entity.Position = Vector2{
		X: clamp(target.Position.X-entity.Radius-target.Radius-18, 0, w.width),
		Y: target.Position.Y,
	}
	entity.Intent = IntentState{}
	hits := w.swordRTargets(entity, target.Position, skill, tick)
	for _, hit := range hits {
		damage := swordRDamage(entity, hit, skill, state.Level, tick)
		hit.Combat.LastHitTick = tick
		if hit.Kind != EntityKindDummy {
			wasAlive := hit.Stats.HP > 0
			w.applyAOEDamage(entity, hit, damage, "physical", tickRate)
			hit.Control.AirborneUntilTick += secondsToTicks(skillMetaRange(skill, "airborneExtendSeconds", 1), tickRate)
			if wasAlive && hit.Stats.HP == 0 {
				w.applyKillReward(entity, hit)
				w.killPlayer(hit, tick, tickRate)
				w.removeDeadUnit(hit)
			}
		} else {
			hit.Combat.LastDamage = damage
			hit.Combat.LastDamageType = "physical"
		}
	}
	entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	entity.Passive.MaxShield = w.swordShieldValue(entity)
	entity.Passive.Shield = entity.Passive.MaxShield
	qState := entity.Skills[swordQSkillID]
	qState.Stacks = 0
	qState.StacksExpireTick = 0
	entity.Skills[swordQSkillID] = qState
	entity.Sword.LastBreathUntilTick = tick + secondsToTicks(skillMetaRange(skill, "lastBreathDurationSeconds", 15), tickRate)
	entity.Control.ActionLockedUntilTick = tick + secondsToTicks(skillMetaRange(skill, "selfActionLockSeconds", 1), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{80000, 55000, 30000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordRSkillID] = state
}

func (w *World) swordRTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig, tick uint64) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	pickPadding := skillMetaRange(skill, "targetPickPadding", 80)
	for _, target := range w.entities {
		if !isAirborneEnemyHero(entity, target, tick) {
			continue
		}
		dist := distance(targetPoint, target.Position)
		if dist > target.Radius+pickPadding {
			continue
		}
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	}
	if best != nil {
		return best
	}
	castRange := skillRange(skill, 1200)
	for _, target := range w.entities {
		if !isAirborneEnemyHero(entity, target, tick) {
			continue
		}
		dist := distance(entity.Position, target.Position)
		if dist > castRange+target.Radius {
			continue
		}
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	}
	return best
}

func (w *World) swordRTargets(entity *Entity, center Vector2, skill config.SkillConfig, tick uint64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !isAirborneEnemyHero(entity, target, tick) {
			continue
		}
		if distance(center, target.Position) <= skillMetaRange(skill, "hitRadius", 450)+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func isAirborneEnemyHero(attacker *Entity, target *Entity, tick uint64) bool {
	if !canAttackTarget(attacker, target) {
		return false
	}
	if target.Control.AirborneUntilTick <= tick {
		return false
	}
	return target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero
}

func (w *World) swordETarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) > skillRange(skill, 475)+target.Radius {
			continue
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+skillMetaRange(skill, "targetPickPadding", 48) {
			continue
		}
		if distToPoint < bestDistance {
			best = target
			bestDistance = distToPoint
		}
	}
	return best
}

func swordEDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{60, 70, 80, 90, 100})
	damageValue := baseDamage + attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 0.2) + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	damageValue *= 1 + float64(attacker.Sword.SweepingBladeStacks)*skillMetaRange(skill, "stackDamageBonus", 0.25)
	return magicDamageAfterResistance(attacker, target, damageValue, tick)
}

func swordRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{200, 300, 400})
	return physicalDamageAfterResistance(attacker, target, baseDamage+attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 1.5), tick)
}

func (w *World) swordQTargets(entity *Entity, targetPoint Vector2, qRange float64, form string, skill config.SkillConfig) []*Entity {
	if form == "circle" {
		return w.targetsInRadius(entity, entity.Position, qRange)
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	if form == "whirlwind" {
		return w.targetsAlongMovingCircle(entity, entity.Position, Vector2{X: dx, Y: dy}, qRange, skillMetaRange(skill, "whirlwindRadius", 70))
	}
	width := skillMetaRange(skill, "lineWidth", 55)
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(entity.Position, Vector2{X: dx, Y: dy}, target.Position)
		if along < 0 || along > qRange+target.Radius {
			continue
		}
		if perpendicular > width+target.Radius {
			continue
		}
		hits = append(hits, target)
	}
	return hits
}

func (w *World) targetsAlongMovingCircle(entity *Entity, origin Vector2, direction Vector2, travelRange float64, radius float64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(origin, direction, target.Position)
		if along < -target.Radius || along > travelRange+target.Radius {
			continue
		}
		if perpendicular <= radius+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func (w *World) targetsInRadius(entity *Entity, center Vector2, radius float64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(center, target.Position) <= radius+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func (w *World) targetsInCone(entity *Entity, direction Vector2, coneRange float64, angleDegrees float64) []*Entity {
	hits := make([]*Entity, 0)
	if direction.X == 0 && direction.Y == 0 {
		direction = Vector2{X: 1, Y: 0}
	}
	cosLimit := math.Cos((angleDegrees / 2) * math.Pi / 180)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		toTarget := Vector2{X: target.Position.X - entity.Position.X, Y: target.Position.Y - entity.Position.Y}
		dist := math.Hypot(toTarget.X, toTarget.Y)
		if dist > coneRange+target.Radius || dist == 0 {
			continue
		}
		dot := (toTarget.X*direction.X + toTarget.Y*direction.Y) / dist
		if dot >= cosLimit {
			hits = append(hits, target)
		}
	}
	return hits
}

func isMonster(entity *Entity) bool {
	if entity == nil {
		return false
	}
	switch entity.Kind {
	case EntityKindPlayer, EntityKindEnemyHero, EntityKindTower, EntityKindBarracks, EntityKindCrystal, EntityKindDummy:
		return false
	default:
		return entity.Team == TeamNeutral
	}
}

func isMinion(entity *Entity) bool {
	if entity == nil {
		return false
	}
	switch entity.Kind {
	case EntityKindMeleeMinion, EntityKindRangedMinion, EntityKindSiegeMinion, EntityKindSuperMinion:
		return true
	default:
		return false
	}
}
