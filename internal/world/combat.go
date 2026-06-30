package world

import (
	"l-battle/internal/config"
	"math"
	"strconv"
)

func (w *World) swordQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, tick uint64) int {
	state := attacker.Skills[swordQSkillID]
	baseDamage := skillMetaListByLevel(skill, "baseDamage", state.Level, []float64{20, 45, 70, 95, 120})
	attack := baseDamage + attacker.Stats.Attack*skillMetaRange(skill, "adRatio", 1)
	if w.attackCrits(attacker, target, tick) {
		attack *= w.critDamageMultiplier(attacker)
	}
	return physicalDamageAfterResistance(attacker, target, attack, tick)
}

func (w *World) swordQCooldownTicks(entity *Entity, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	attackSpeedBonus := 0.0
	if entity != nil {
		attackSpeedBonus = entity.Stats.AttackSpeedBonus
	}
	return swordQCooldownTicksByBonus(attackSpeedBonus, skill, skillLevel, tickRate)
}

func swordQCooldownTicksByBonus(attackSpeedBonus float64, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	baseCooldownMS := skillMetaListByLevelMS(skill, "cooldownMs", skillLevel, []float64{6000, 5500, 5000, 4500, 4000})
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	seconds := float64(baseCooldownMS) / 1000 * (1 - attackSpeedBonus*0.6)
	minSeconds := skillMetaRange(skill, "minCooldownSeconds", 1.33)
	if seconds < minSeconds {
		seconds = minSeconds
	}
	return uint64(math.Ceil(seconds*float64(tickRate) - 1e-6))
}

func (w *World) applyAttack(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker.Kind != EntityKindPlayer || tick < attacker.Combat.NextAttackTick {
		return
	}
	if attacker.HeroID == warriorHeroID && tick < attacker.Warrior.JudgmentUntilTick {
		return
	}
	if !canAttackTarget(attacker, target) {
		return
	}
	if distance(attacker.Position, target.Position) > w.attackReachAtTick(attacker, target, tick) {
		return
	}

	if isRangedBasicAttacker(attacker) {
		w.fireBasicAttackProjectile(attacker, target, tick, tickRate)
		attacker.Combat.NextAttackTick = tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(attacker, tick), tickRate)
		return
	}

	damage := w.attackDamage(attacker, target, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyDamage(attacker, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(attacker, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
	}
	attacker.Combat.NextAttackTick = tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(attacker, tick), tickRate)
	w.applyTankWAftershock(attacker, target, tick, tickRate)
	w.consumeWarriorQ(attacker, target, tick, tickRate)
}

func isRangedBasicAttacker(attacker *Entity) bool {
	return attacker != nil && attacker.HeroID == archerHeroID
}

func (w *World) fireBasicAttackProjectile(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil {
		return
	}
	dx, dy := normalize(target.Position.X-attacker.Position.X, target.Position.Y-attacker.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	speedPerSecond := 1400.0
	if tickRate > 0 {
		speedPerSecond /= float64(tickRate)
	}
	w.nextProjectileID++
	id := "projectile:basic_arrow:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "basic_arrow",
		Team:         attacker.Team,
		SourceID:     attacker.ID,
		TargetID:     target.ID,
		Position:     attacker.Position,
		Start:        attacker.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerSecond,
		Range:        w.attackReachAtTick(attacker, target, tick) + 220,
		Radius:       10,
		DisplayRange: attacker.Stats.AttackRange,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	}
}

func (w *World) attackDamage(attacker *Entity, target *Entity, tick uint64) int {
	attack := attacker.Stats.Attack
	if attacker.HeroID == archerHeroID {
		attack *= w.archerBasicAttackMultiplier(attacker, target, tick)
	} else if w.attackCrits(attacker, target, tick) {
		attack *= w.critDamageMultiplier(attacker)
	}
	return physicalDamageAfterResistance(attacker, target, attack+w.warriorQBonusDamage(attacker, tick)+w.tankWBonusDamage(attacker, tick), tick)
}

func (w *World) archerBasicAttackMultiplier(attacker *Entity, target *Entity, tick uint64) float64 {
	if attacker == nil || target == nil || attacker.HeroID != archerHeroID {
		return 1
	}
	if target.Control.MoveSpeedSlow <= 0 || target.Control.MoveSpeedSlowUntil == 0 || tick >= target.Control.MoveSpeedSlowUntil {
		return 1
	}
	skill := w.heroPassiveSkill(attacker)
	multiplier := skillMetaRange(skill, "slowedTargetDamageMultiplier", 1.1)
	critRatio := skillMetaRange(skill, "critChanceDamageRatio", 1)
	if critRatio > 0 {
		multiplier += w.critChance(attacker) * critRatio
	}
	if multiplier < 1 {
		return 1
	}
	return multiplier
}

func (w *World) warriorQBonusDamage(attacker *Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != warriorHeroID || tick >= attacker.Warrior.DecisiveStrikeUntilTick {
		return 0
	}
	skill := w.skillConfig(warriorQSkillID)
	level := attacker.Warrior.DecisiveStrikeLevel
	if level <= 0 {
		level = 1
	}
	baseDamage := skillMetaListByLevel(skill, "bonusDamage", level, []float64{30, 60, 90, 120, 150})
	return baseDamage + attacker.Stats.Attack*skillMetaRange(skill, "totalAdRatio", 1.4)
}

func (w *World) consumeWarriorQ(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.HeroID != warriorHeroID || tick >= attacker.Warrior.DecisiveStrikeUntilTick {
		return
	}
	skill := w.skillConfig(warriorQSkillID)
	if target != nil {
		silenceTicks := secondsToTicks(skillMetaRange(skill, "silenceSeconds", 1.5), tickRate)
		target.Control.SilencedUntilTick = tick + controlTicksAfterTenacity(target, silenceTicks, tick)
	}
	attacker.Warrior.DecisiveStrikeUntilTick = 0
	attacker.Warrior.DecisiveStrikeLevel = 0
}

func (w *World) tankWBonusDamage(attacker *Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != tankHeroID || tick >= attacker.Tank.ThunderclapEmpowerUntil {
		return 0
	}
	skill := w.skillConfig(tankWSkillID)
	level := attacker.Tank.ThunderclapLevel
	if level <= 0 {
		level = 1
	}
	return skillMetaListByLevel(skill, "bonusDamage", level, []float64{30, 40, 50, 60, 70}) +
		float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.2) +
		attacker.Stats.PhysicalDefense*skillMetaRange(skill, "armorRatio", 0.15)
}

func (w *World) applyTankWAftershock(attacker *Entity, primary *Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.HeroID != tankHeroID || tick >= attacker.Tank.ThunderclapAftershockUntil {
		return
	}
	skill := w.skillConfig(tankWSkillID)
	level := attacker.Tank.ThunderclapLevel
	if level <= 0 {
		level = 1
	}
	damage := skillMetaListByLevel(skill, "aftershockDamage", level, []float64{15, 25, 35, 45, 55}) +
		float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "aftershockAPRatio", 0.3) +
		attacker.Stats.PhysicalDefense*skillMetaRange(skill, "aftershockArmorRatio", 0.15)
	direction := Vector2{X: 1, Y: 0}
	if primary != nil {
		dx, dy := normalize(primary.Position.X-attacker.Position.X, primary.Position.Y-attacker.Position.Y)
		if dx != 0 || dy != 0 {
			direction = Vector2{X: dx, Y: dy}
		}
	}
	coneRange := skillMetaRange(skill, "aftershockConeRange", 300)
	coneAngle := skillMetaRange(skill, "aftershockConeAngleDegrees", 70)
	w.addTankWAftershockEffect(attacker, direction, coneRange, coneAngle, tick, tickRate)
	for _, target := range w.targetsInCone(attacker, direction, coneRange, coneAngle) {
		target.Combat.LastHitTick = tick
		previousDamage := 0
		if primary != nil && target.ID == primary.ID {
			previousDamage = target.Combat.LastDamage
		}
		aftershockDamage := damage
		if isMonster(target) {
			aftershockDamage *= skillMetaRange(skill, "monsterDamageMultiplier", 1.8)
		}
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyDamage(attacker, target, physicalDamageAfterResistance(attacker, target, aftershockDamage, tick), tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(attacker, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = physicalDamageAfterResistance(attacker, target, aftershockDamage, tick)
			target.Combat.LastDamageType = "physical"
		}
		if previousDamage > 0 {
			target.Combat.LastDamage += previousDamage
		}
	}
	if tick < attacker.Tank.ThunderclapEmpowerUntil {
		attacker.Tank.ThunderclapEmpowerUntil = 0
	}
}

func (w *World) addTankWAftershockEffect(attacker *Entity, direction Vector2, coneRange float64, coneAngle float64, tick uint64, tickRate int) {
	if attacker == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.25))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.nextEffectID++
	id := "effect:tank_w_aftershock:" + strconv.Itoa(w.nextEffectID)
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "tank_w_aftershock",
		Team:      attacker.Team,
		Start:     attacker.Position,
		Dir:       direction,
		Range:     coneRange,
		Radius:    coneAngle,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	}
}

func physicalDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	damageReduce := target.damageReductionForType("physical", tick)
	return damageAfterResistance(rawDamage, effectiveResistance(target.Stats.PhysicalDefense, attacker.Stats.PhysicalPenPercent, attacker.Stats.PhysicalPenFlat), damageReduce)
}

func magicDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	damageReduce := target.damageReductionForType("magic", tick)
	return damageAfterResistance(rawDamage, effectiveResistance(target.Stats.MagicDefense, attacker.Stats.MagicPenPercent, attacker.Stats.MagicPenFlat), damageReduce)
}

func tankQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{70, 120, 170, 220, 270})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func applyTankQMoveSpeedSteal(source *Entity, target *Entity, ratio float64, until uint64) {
	if source == nil || target == nil || ratio <= 0 || until == 0 {
		return
	}
	ratio = clamp(ratio, 0, 1)
	stolen := EffectiveMoveSpeedAtTick(target, 0) * ratio
	source.Control.MoveSpeedBonusFlat = stolen
	source.Control.MoveSpeedBonusUntil = until
	target.Control.MoveSpeedSlow = ratio
	target.Control.MoveSpeedSlowUntil = until
}

func trueDamageAfterReduction(target *Entity, rawDamage float64, tick uint64) int {
	return damageAfterResistance(rawDamage, 0, target.damageReductionForType("true", tick))
}

func (entity *Entity) damageReductionForType(damageType string, tick uint64) float64 {
	reductions := []float64{entity.Stats.DamageReduce}
	switch damageType {
	case "physical":
		reductions = append(reductions, entity.Stats.PhysicalDamageReduce)
	case "magic":
		reductions = append(reductions, entity.Stats.MagicDamageReduce)
	}
	if entity.HeroID == warriorHeroID && entity.Warrior.CourageUntilTick > 0 {
		reductions = append(reductions, entity.Warrior.courageDamageReductionAtTick(tick))
	}
	return stackDamageReduction(reductions...)
}

func (entity *Entity) tenacityAtTick(tick uint64) float64 {
	tenacity := []float64{}
	if entity.HeroID == warriorHeroID && entity.Warrior.CourageFrontUntilTick > 0 && (tick == 0 || tick < entity.Warrior.CourageFrontUntilTick) {
		tenacity = append(tenacity, entity.Warrior.CourageFrontTenacity)
	}
	return stackTenacity(tenacity...)
}

func controlTicksAfterTenacity(target *Entity, ticks uint64, tick uint64) uint64 {
	if target == nil || ticks == 0 {
		return ticks
	}
	remainingRatio := 1 - target.tenacityAtTick(tick)
	adjusted := uint64(math.Ceil(float64(ticks) * remainingRatio))
	if adjusted < 1 {
		return 1
	}
	return adjusted
}

func (state WarriorState) courageDamageReductionAtTick(tick uint64) float64 {
	if state.CourageUntilTick == 0 {
		return 0
	}
	if tick > 0 && tick >= state.CourageUntilTick {
		return 0
	}
	if tick == 0 || tick < state.CourageFrontUntilTick {
		return state.CourageFrontDamageReduce
	}
	return state.CourageBackDamageReduce
}

func warriorWShieldValue(entity *Entity, skill config.SkillConfig, skillLevel int) int {
	baseShield := skillMetaListByLevel(skill, "shieldValue", skillLevel, []float64{70, 95, 120, 145, 170})
	return int(math.Round(baseShield + float64(entity.Stats.BonusHP)*skillMetaRange(skill, "bonusHealthRatio", 0.2)))
}

func effectiveResistance(resistance float64, percentPen float64, flatPen float64) float64 {
	if resistance < 0 {
		return resistance
	}
	if percentPen < 0 {
		percentPen = 0
	}
	if percentPen > 1 {
		percentPen = 1
	}
	if flatPen < 0 {
		flatPen = 0
	}
	effective := resistance*(1-percentPen) - flatPen
	if effective < 0 {
		return 0
	}
	return effective
}

func damageAfterResistance(rawDamage float64, resistance float64, damageReduce float64) int {
	if rawDamage <= 0 {
		return 0
	}
	multiplier := 100 / (resistance + 100)
	if resistance < 0 {
		denominator := 100 + resistance
		if denominator < 1 {
			denominator = 1
		}
		multiplier = 100 / denominator
	}
	damageReduce = clamp(damageReduce, 0, 1)
	damage := int(math.Round(rawDamage * multiplier * (1 - damageReduce)))
	if damage < 1 {
		return 1
	}
	return damage
}

func stackDamageReduction(reductions ...float64) float64 {
	multiplier := 1.0
	for _, reduction := range reductions {
		reduction = clamp(reduction, 0, 1)
		multiplier *= 1 - reduction
	}
	return 1 - multiplier
}

func stackTenacity(tenacityValues ...float64) float64 {
	multiplier := 1.0
	for _, tenacity := range tenacityValues {
		tenacity = clamp(tenacity, 0, 1)
		multiplier *= 1 - tenacity
	}
	return 1 - multiplier
}

func (w *World) attackCrits(attacker *Entity, target *Entity, tick uint64) bool {
	chance := w.critChance(attacker)
	if chance <= 0 {
		return false
	}
	if chance >= 1 {
		return true
	}
	roll := deterministicCritRoll(attacker.ID, target.ID, tick)
	return roll < chance
}

func (w *World) critChance(attacker *Entity) float64 {
	chance := attacker.Stats.CritChance
	if attacker.HeroID == swordHeroID {
		chance *= skillMetaRange(w.heroPassiveSkill(attacker), "critChanceMultiplier", 2)
	}
	if chance > 1 {
		return 1
	}
	if chance < 0 {
		return 0
	}
	return chance
}

func (w *World) critDamageMultiplier(attacker *Entity) float64 {
	if attacker.HeroID == swordHeroID {
		return skillMetaRange(w.heroPassiveSkill(attacker), "critDamageMultiplier", 1.9)
	}
	return 2
}

func deterministicCritRoll(attackerID string, targetID string, tick uint64) float64 {
	hash := uint64(1469598103934665603)
	for _, value := range []string{attackerID, targetID, strconv.FormatUint(tick, 10)} {
		for i := 0; i < len(value); i++ {
			hash ^= uint64(value[i])
			hash *= 1099511628211
		}
	}
	return float64(hash%critRollModulo) / critRollModulo
}

func (w *World) applyDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "physical", tickRate)
}

func (w *World) applyMagicDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "magic", tickRate)
}

func (w *World) applyTrueDamage(source *Entity, target *Entity, rawDamage float64, tickRate int) {
	w.applyResolvedDamage(source, target, trueDamageAfterReduction(target, rawDamage, target.Combat.LastHitTick), "true", tickRate)
}

func (w *World) applyResolvedDamage(source *Entity, target *Entity, damage int, damageType string, tickRate int) {
	if damage <= 0 {
		target.Combat.LastDamage = 0
		target.Combat.LastDamageType = ""
		return
	}
	damage = w.applyShield(source, target, damage, tickRate)
	target.Combat.LastDamage = damage
	target.Combat.LastDamageType = damageType
	w.applyArcherFrostShot(source, target, target.Combat.LastHitTick, tickRate)
	w.breakTankGraniteShield(target, target.Combat.LastHitTick)
	if damage <= 0 {
		return
	}
	target.Stats.HP -= damage
	if target.Stats.HP < 0 {
		target.Stats.HP = 0
	}
	w.breakWarriorToughness(source, target, target.Combat.LastHitTick)
}

func (w *World) applyArcherFrostShot(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != archerHeroID {
		return
	}
	skill := w.heroPassiveSkill(source)
	slow := archerFrostSlowRatio(source.Level, skill)
	if w.attackCrits(source, target, tick) {
		slow *= skillMetaRange(skill, "critSlowMultiplier", 2)
	}
	duration := secondsToTicks(skillMetaRange(skill, "slowSeconds", 2), tickRate)
	applyMoveSpeedSlow(target, slow, tick+duration)
}

func archerFrostSlowRatio(level int, skill config.SkillConfig) float64 {
	minSlow := skillMetaRange(skill, "slowMin", 0.2)
	maxSlow := skillMetaRange(skill, "slowMax", 0.3)
	level = clampInt(level, MinHeroLevel, MaxHeroLevel)
	if MaxHeroLevel <= MinHeroLevel {
		return maxSlow
	}
	progress := float64(level-MinHeroLevel) / float64(MaxHeroLevel-MinHeroLevel)
	return minSlow + (maxSlow-minSlow)*progress
}

func applyMoveSpeedSlow(target *Entity, slow float64, until uint64) {
	if target == nil || slow <= 0 || until == 0 {
		return
	}
	slow = clamp(slow, 0, 1)
	if until < target.Control.MoveSpeedSlowUntil && slow <= target.Control.MoveSpeedSlow {
		return
	}
	target.Control.MoveSpeedSlow = slow
	target.Control.MoveSpeedSlowUntil = until
}

func (w *World) skillConfig(skillID string) config.SkillConfig {
	if w == nil || w.skills == nil || skillID == "" {
		return config.SkillConfig{}
	}
	skill, _ := w.skills.Get(skillID)
	return skill
}

func (w *World) heroPassiveSkill(entity *Entity) config.SkillConfig {
	if entity == nil || w == nil || w.heroes == nil {
		return config.SkillConfig{}
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return config.SkillConfig{}
	}
	return w.skillConfig(hero.Skills.Passive)
}

func (w *World) breakWarriorToughness(source *Entity, target *Entity, tick uint64) {
	if target == nil || target.HeroID != warriorHeroID || !warriorToughnessBreaksRegen(source) {
		return
	}
	target.Passive.LastRegenBreakTick = tick
	target.Passive.NextRegenTick = 0
}

func (w *World) breakTankGraniteShield(target *Entity, tick uint64) {
	if target == nil || target.HeroID != tankHeroID {
		return
	}
	target.Passive.LastRegenBreakTick = tick
	target.Passive.NextRegenTick = 0
}

func warriorToughnessBreaksRegen(source *Entity) bool {
	if source == nil {
		return false
	}
	switch source.Kind {
	case EntityKindPlayer, EntityKindEnemyHero, EntityKindTower:
		return true
	default:
		return false
	}
}

func (w *World) killPlayer(target *Entity, tick uint64, tickRate int) {
	if target.Kind != EntityKindPlayer || target.Death.Dead {
		return
	}
	target.Death = DeathState{
		Dead:              true,
		RespawnTick:       tick + uint64(respawnSeconds*tickRate),
		RespawnTickRate:   tickRate,
		RespawnSeconds:    respawnSeconds,
		LastDeathPosition: target.Position,
	}
	target.Intent = IntentState{}
	target.Warrior = WarriorState{}
	target.Passive.Shield = 0
	target.Passive.MaxShield = 0
	target.Passive.ShieldExpireTick = 0
	for _, entity := range w.entities {
		if entity.Intent.AttackTargetID == target.ID {
			entity.Intent.AttackTargetID = ""
		}
	}
}

func (w *World) applyShield(source *Entity, target *Entity, damage int, tickRate int) int {
	if target == nil {
		return damage
	}
	if target.HeroID == swordHeroID && target.Passive.Shield <= 0 && target.Passive.SwordIntent >= target.Passive.MaxSwordIntent && swordShieldTriggers(source) {
		skill := w.heroPassiveSkill(target)
		target.Passive.MaxShield = w.swordShieldValue(target)
		target.Passive.Shield = target.Passive.MaxShield
		target.Passive.ShieldExpireTick = target.Combat.LastHitTick + secondsToTicks(skillMetaRange(skill, "shieldDurationSeconds", 1), tickRate)
		target.Passive.SwordIntent = 0
	}
	if target.Passive.Shield <= 0 {
		return damage
	}
	absorbed := damage
	if absorbed > target.Passive.Shield {
		absorbed = target.Passive.Shield
	}
	target.Passive.Shield -= absorbed
	return damage - absorbed
}

func swordShieldTriggers(source *Entity) bool {
	if source == nil {
		return false
	}
	return source.Kind == EntityKindPlayer || source.Kind == EntityKindEnemyHero
}

func (w *World) swordShieldValue(entity *Entity) int {
	level := clampInt(entity.Level, MinHeroLevel, MaxHeroLevel)
	skill := w.heroPassiveSkill(entity)
	return int(math.Round(skillMetaCurveByLevel(skill, "shieldValue", "shieldValueLevels", level, 125)))
}

func (w *World) removeDeadUnit(target *Entity) {
	if target.Kind == EntityKindPlayer || target.Kind == EntityKindDummy {
		return
	}
	delete(w.entities, target.ID)
	for _, entity := range w.entities {
		if entity.Intent.AttackTargetID == target.ID {
			entity.Intent.AttackTargetID = ""
		}
	}
}
