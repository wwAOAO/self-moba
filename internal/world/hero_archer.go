package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
	"strconv"
)

func (w *World) applyArcherQ(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	maxStacks := archerFocusMaxStacks(skill)
	if entity.Archer.FocusStacks < maxStacks {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 50)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Archer.FocusStacks = 0
	entity.Archer.FocusExpireTick = 0
	entity.Archer.FocusActiveLevel = state.Level
	entity.Archer.FocusAttackSpeed = skillMetaListByLevel(skill, "attackSpeedBonus", state.Level, []float64{0.2, 0.25, 0.3, 0.35, 0.4})
	entity.Archer.FocusBonusADRatio = skillMetaListByLevel(skill, "bonusAdDamageRatio", state.Level, []float64{1.05, 1.1, 1.15, 1.2, 1.25})
	entity.Archer.FocusActiveUntil = tick + secondsToTicks(skillMetaRange(skill, "activeDurationSeconds", 5), tickRate)
	entity.Skills[archerQSkillID] = state
}

func (w *World) applyArcherW(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{70, 65, 60, 55, 50})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Skills[archerWSkillID] = SkillState{
		SkillID:           state.SkillID,
		Level:             state.Level,
		CooldownUntilTick: tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{14000, 11500, 9000, 6500, 4000}), tickRate),
		Stacks:            state.Stacks,
		StacksExpireTick:  state.StacksExpireTick,
	}
	w.lockAttackAfterCast(entity, tick, tickRate)

	target := Vector2{X: cast.TargetX, Y: cast.TargetY}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	centerAngle := math.Atan2(dy, dx)
	arrowCount := int(math.Round(skillMetaListByLevel(skill, "arrowCount", state.Level, []float64{7, 8, 9, 10, 11})))
	if arrowCount < 1 {
		arrowCount = 1
	}
	coneAngle := skillMetaRange(skill, "coneAngleDegrees", 48) * math.Pi / 180
	startAngle := centerAngle - coneAngle/2
	step := 0.0
	if arrowCount > 1 {
		step = coneAngle / float64(arrowCount-1)
	}
	speedPerTick := skillMetaRange(skill, "projectileSpeed", 1500)
	if tickRate > 0 {
		speedPerTick /= float64(tickRate)
	}
	rangeValue := skillRange(skill, 1200)
	radius := skillMetaRange(skill, "projectileRadius", 16)
	w.nextProjectileID++
	groupID := "projectile:archer_w_group:" + strconv.Itoa(w.nextProjectileID)
	for i := 0; i < arrowCount; i++ {
		angle := startAngle + step*float64(i)
		w.nextProjectileID++
		id := "projectile:archer_w:" + strconv.Itoa(w.nextProjectileID)
		w.projectiles[id] = &Projectile{
			ID:           id,
			Kind:         "archer_volley_arrow",
			Team:         entity.Team,
			SourceID:     entity.ID,
			SkillID:      archerWSkillID,
			GroupID:      groupID,
			Position:     entity.Position,
			Start:        entity.Position,
			Dir:          Vector2{X: math.Cos(angle), Y: math.Sin(angle)},
			SpeedPerTick: speedPerTick,
			Range:        rangeValue,
			Radius:       radius,
			Damage:       state.Level,
			CreatedAt:    tick,
			ExpiresAt:    tick + secondsToTicks(rangeValue/skillMetaRange(skill, "projectileSpeed", 1500)+0.2, tickRate),
			HitIDs:       make(map[string]bool),
		}
	}
}

func (w *World) applyArcherE(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	maxCharges := archerHawkMaxCharges(skill)
	if state.Stacks <= 0 {
		return
	}
	state.Stacks--
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skill.CooldownMS, tickRate)
	if state.Stacks < maxCharges && state.StacksExpireTick == 0 {
		state.StacksExpireTick = tick + archerHawkRechargeTicks(entity, skill, state.Level, tickRate)
	}
	entity.Skills[archerESkillID] = state
	w.lockAttackAfterCast(entity, tick, tickRate)

	target := Vector2{
		X: clamp(cast.TargetX, 0, w.width),
		Y: clamp(cast.TargetY, 0, w.height),
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	speed := skillMetaRange(skill, "projectileSpeed", 1800)
	travelSeconds := distance(entity.Position, target) / speed
	arriveTick := tick + secondsToTicks(travelSeconds, tickRate)
	expiresAt := arriveTick + secondsToTicks(skillMetaRange(skill, "lingerSeconds", 5), tickRate)
	effectSpeed := speed
	if tickRate > 0 {
		effectSpeed /= float64(tickRate)
	}
	w.nextEffectID++
	id := "effect:archer_hawk:" + strconv.Itoa(w.nextEffectID)
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "archer_hawk",
		Team:      entity.Team,
		Start:     entity.Position,
		End:       target,
		Dir:       Vector2{X: dx, Y: dy},
		Speed:     effectSpeed,
		Height:    float64(arriveTick),
		Radius:    80,
		CreatedAt: tick,
		ExpiresAt: expiresAt,
	}
}

func (w *World) applyArcherR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{100000, 90000, 80000}), tickRate)
	entity.Skills[archerRSkillID] = state
	delayTicks := secondsToTicks(skillMetaRange(skill, "castDelaySeconds", 0.25), tickRate)
	entity.Control.ActionLockedUntilTick = tick + delayTicks
	entity.Archer.CrystalArrowPending = true
	entity.Archer.CrystalArrowReleaseTick = tick + delayTicks
	entity.Archer.CrystalArrowTarget = Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Archer.CrystalArrowLevel = state.Level
}

func (w *World) releaseArcherCrystalArrow(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID || !entity.Archer.CrystalArrowPending || tick < entity.Archer.CrystalArrowReleaseTick {
		return
	}
	skill := w.skillConfig(archerRSkillID)
	target := entity.Archer.CrystalArrowTarget
	level := entity.Archer.CrystalArrowLevel
	entity.Archer.CrystalArrowPending = false
	entity.Archer.CrystalArrowReleaseTick = 0
	entity.Archer.CrystalArrowTarget = Vector2{}
	entity.Archer.CrystalArrowLevel = 0

	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	minSpeed := skillMetaRange(skill, "projectileMinSpeed", 1500)
	maxSpeed := skillMetaRange(skill, "projectileMaxSpeed", 2100)
	speedPerTick := minSpeed
	if tickRate > 0 {
		speedPerTick /= float64(tickRate)
	}
	w.nextProjectileID++
	id := "projectile:archer_r:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "archer_crystal_arrow",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      archerRSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		SpeedMin:     minSpeed,
		SpeedMax:     maxSpeed,
		Range:        skillRange(skill, DefaultMapWidth),
		Radius:       skillMetaRange(skill, "projectileRadius", 28),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillRange(skill, DefaultMapWidth)/minSpeed+0.5, tickRate),
		HitIDs:       make(map[string]bool),
	}
}

func archerRDamage(entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64, multiplier float64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{200, 400, 600})
	rawDamage := (baseDamage + float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 1)) * multiplier
	return magicDamageAfterResistance(entity, target, rawDamage, tick)
}

func archerRStunTicks(projectile *Projectile, skill config.SkillConfig, tickRate int) uint64 {
	if projectile == nil {
		return secondsToTicks(skillMetaRange(skill, "minStunSeconds", 1), tickRate)
	}
	minSeconds := skillMetaRange(skill, "minStunSeconds", 1)
	maxSeconds := skillMetaRange(skill, "maxStunSeconds", 3.5)
	maxDistance := skillMetaRange(skill, "maxStunDistance", 1400)
	progress := 1.0
	if maxDistance > 0 {
		progress = clamp(projectile.Traveled/maxDistance, 0, 1)
	}
	return secondsToTicks(minSeconds+(maxSeconds-minSeconds)*progress, tickRate)
}

func (w *World) applyArcherRSplash(source *Entity, primary *Entity, projectile *Projectile, skill config.SkillConfig, tick uint64, tickRate int) {
	if source == nil || primary == nil || projectile == nil {
		return
	}
	radius := skillMetaRange(skill, "splashRadius", 260)
	multiplier := skillMetaRange(skill, "splashDamageMultiplier", 0.5)
	for _, target := range w.entities {
		if target == nil || target.ID == primary.ID || !canAttackTarget(source, target) {
			continue
		}
		if distance(target.Position, primary.Position) > radius+target.Radius {
			continue
		}
		damage := archerRDamage(source, target, skill, projectile.Damage, tick, multiplier)
		target.Combat.LastHitTick = tick
		wasAlive := target.Stats.HP > 0
		w.applyMagicDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(source, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	}
}

func (w *World) refreshArcherSkillOnUpgrade(entity *Entity, skillID string) {
	if entity == nil || entity.HeroID != archerHeroID || skillID != archerESkillID {
		return
	}
	state := entity.Skills[archerESkillID]
	if state.Level <= 0 {
		return
	}
	skill := w.skillConfig(archerESkillID)
	maxCharges := archerHawkMaxCharges(skill)
	if state.Stacks <= 0 {
		state.Stacks = maxCharges
		state.StacksExpireTick = 0
		entity.Skills[archerESkillID] = state
	}
}

func (w *World) tickArcherHawkCharges(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	state, ok := entity.Skills[archerESkillID]
	if !ok || state.Level <= 0 {
		return
	}
	maxCharges := archerHawkMaxCharges(w.skillConfig(archerESkillID))
	if state.Stacks >= maxCharges {
		state.Stacks = maxCharges
		state.StacksExpireTick = 0
		entity.Skills[archerESkillID] = state
		return
	}
	if state.StacksExpireTick == 0 {
		state.StacksExpireTick = tick + archerHawkRechargeTicks(entity, w.skillConfig(archerESkillID), state.Level, tickRate)
		entity.Skills[archerESkillID] = state
		return
	}
	for state.Stacks < maxCharges && state.StacksExpireTick > 0 && tick >= state.StacksExpireTick {
		state.Stacks++
		if state.Stacks >= maxCharges {
			state.StacksExpireTick = 0
			break
		}
		state.StacksExpireTick += archerHawkRechargeTicks(entity, w.skillConfig(archerESkillID), state.Level, tickRate)
	}
	entity.Skills[archerESkillID] = state
}

func archerHawkMaxCharges(skill config.SkillConfig) int {
	maxCharges := int(math.Round(skillMetaRange(skill, "maxCharges", 2)))
	if maxCharges < 1 {
		return 1
	}
	return maxCharges
}

func archerHawkRechargeTicks(entity *Entity, skill config.SkillConfig, level int, tickRate int) uint64 {
	return cooldownTicksFor(entity, skillMetaListByLevelMS(skill, "rechargeMs", level, []float64{90000, 80000, 70000, 60000, 50000}), tickRate)
}

func archerWDamage(entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{20, 35, 50, 65, 80})
	rawDamage := baseDamage + entity.Stats.Attack*skillMetaRange(skill, "adRatio", 1)
	return physicalDamageAfterResistance(entity, target, rawDamage, tick)
}

func (w *World) addArcherFocusStack(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	state, ok := entity.Skills[archerQSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	if tick < entity.Archer.FocusActiveUntil {
		return
	}
	skill := w.skillConfig(archerQSkillID)
	maxStacks := archerFocusMaxStacks(skill)
	if entity.Archer.FocusStacks < maxStacks {
		entity.Archer.FocusStacks++
	}
	entity.Archer.FocusExpireTick = tick + secondsToTicks(skillMetaRange(skill, "stackDurationSeconds", 4), tickRate)
	entity.Skills[archerQSkillID] = state
}

func archerFocusMaxStacks(skill config.SkillConfig) int {
	maxStacks := int(math.Round(skillMetaRange(skill, "maxStacks", 4)))
	if maxStacks < 1 {
		return 1
	}
	return maxStacks
}

func (w *World) expireArcherFocus(entity *Entity, tick uint64) {
	if entity == nil || entity.HeroID != archerHeroID {
		return
	}
	if entity.Archer.FocusStacks > 0 && entity.Archer.FocusExpireTick > 0 && tick >= entity.Archer.FocusExpireTick {
		entity.Archer.FocusStacks = 0
		entity.Archer.FocusExpireTick = 0
	}
	if entity.Archer.FocusActiveUntil > 0 && tick >= entity.Archer.FocusActiveUntil {
		entity.Archer.FocusActiveUntil = 0
		entity.Archer.FocusActiveLevel = 0
		entity.Archer.FocusAttackSpeed = 0
		entity.Archer.FocusBonusADRatio = 0
	}
}

func (w *World) archerFocusBonusDamage(attacker *Entity, target *Entity, tick uint64) int {
	if attacker == nil || target == nil || attacker.HeroID != archerHeroID || tick >= attacker.Archer.FocusActiveUntil {
		return 0
	}
	ratio := attacker.Archer.FocusBonusADRatio
	if ratio <= 0 {
		skill := w.skillConfig(archerQSkillID)
		ratio = skillMetaListByLevel(skill, "bonusAdDamageRatio", attacker.Archer.FocusActiveLevel, []float64{1.05, 1.1, 1.15, 1.2, 1.25})
	}
	return physicalDamageAfterResistance(attacker, target, attacker.Stats.Attack*ratio, tick)
}

func (w *World) applyArcherFocusOnBasicHit(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.HeroID != archerHeroID {
		return
	}
	if tick < attacker.Archer.FocusActiveUntil {
		damage := w.archerFocusBonusDamage(attacker, target, tick)
		if damage > 0 {
			previousDamage := target.Combat.LastDamage
			target.Combat.LastHitTick = tick
			if target.Kind != EntityKindDummy {
				wasAlive := target.Stats.HP > 0
				w.applyDamage(attacker, target, damage, tickRate)
				if previousDamage > 0 {
					target.Combat.LastDamage += previousDamage
				}
				if wasAlive && target.Stats.HP == 0 {
					w.applyKillReward(attacker, target)
					w.killPlayer(target, tick, tickRate)
					w.removeDeadUnit(target)
				}
			} else {
				target.Combat.LastDamage = damage
				target.Combat.LastDamageType = "physical"
				if previousDamage > 0 {
					target.Combat.LastDamage += previousDamage
				}
			}
		}
		return
	}
	w.addArcherFocusStack(attacker, tick, tickRate)
}
