package world

import (
	"math"
	"strconv"
)

type basicAttackDamagePart struct {
	damage     int
	damageType string
}

const (
	basicArrowProjectileKind      = "basic_arrow"
	siegeCannonballProjectileKind = "siege_cannonball"
	siegeMinionSplashRadius       = 300
	siegeMinionSplashSeconds      = 5
	siegeMinionSplashMaxHPRatio   = 0.01
	siegeMinionSplashTickSeconds  = 1
)

func (w *World) applyAttack(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if !canBasicAttack(attacker) || tick < attacker.Combat.NextAttackTick {
		return
	}
	if attacker.Combat.PendingAttackTargetID != "" {
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

	w.startAttackWindup(attacker, target, tick, tickRate)
}

func (w *World) startAttackWindup(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil {
		return
	}
	attacker.Combat.PendingAttackTargetID = target.ID
	attacker.Combat.AttackReleaseTick = tick + attackWindupTicks(attacker, tickRate)
	attacker.Combat.NextAttackTick = tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(attacker, tick), tickRate)
}

func attackWindupTicks(attacker *Entity, tickRate int) uint64 {
	if tickRate <= 0 {
		return 1
	}
	bonus := 0.0
	if attacker != nil {
		bonus = attacker.Stats.AttackSpeedBonus
	}
	if bonus < 0 {
		bonus = 0
	}
	baseWindup := 0.25
	if attacker != nil && attacker.Stats.AttackWindupSeconds > 0 {
		baseWindup = attacker.Stats.AttackWindupSeconds
	}
	ticks := math.Ceil((baseWindup / (1 + bonus)) * float64(tickRate))
	if ticks < 1 {
		return 1
	}
	return uint64(ticks)
}

func (w *World) releasePendingAttack(attacker *Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.Combat.PendingAttackTargetID == "" || tick < attacker.Combat.AttackReleaseTick {
		return
	}
	target := w.entities[attacker.Combat.PendingAttackTargetID]
	attacker.Combat.PendingAttackTargetID = ""
	attacker.Combat.AttackReleaseTick = 0
	if attacker.Death.Dead || attacker.Stats.HP <= 0 || target == nil || !canAttackTarget(attacker, target) {
		return
	}
	if attacker.HeroID == warriorHeroID && tick < attacker.Warrior.JudgmentUntilTick {
		return
	}
	w.resolveBasicAttack(attacker, target, tick, tickRate)
}

func (w *World) resolveBasicAttack(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if isRangedBasicAttacker(attacker) {
		w.fireBasicAttackProjectile(attacker, target, tick, tickRate)
		return
	}

	damage := w.attackDamage(attacker, target, tick, tickRate)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		if !w.applyMinionBasicAttackDamage(attacker, target, tick, tickRate) {
			w.applyBasicAttackDamage(attacker, target, damage, tickRate)
		}
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(attacker, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		w.recordDummyBasicAttackDamage(attacker, target, damage, tick)
	}
	w.onHeroBasicHit(attacker, target, tick, tickRate)
}

func isRangedBasicAttacker(attacker *Entity) bool {
	return attacker != nil && (attacker.HeroID == archerHeroID || attacker.HeroID == mageHeroID || attacker.HeroID == gunnerHeroID || attacker.HeroID == explorerHeroID || attacker.HeroID == frostmageHeroID || attacker.HeroID == fireMageHeroID || attacker.Kind == EntityKindRangedMinion || attacker.Kind == EntityKindSiegeMinion)
}

func canBasicAttack(entity *Entity) bool {
	return entity != nil && (entity.Kind == EntityKindPlayer || isMinion(entity))
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
	kind := basicArrowProjectileKind
	if attacker.Kind == EntityKindSiegeMinion && tick >= attacker.Combat.NextSiegeSplashTick {
		kind = siegeCannonballProjectileKind
	}
	w.nextProjectileID++
	id := "projectile:" + kind + ":" + strconv.Itoa(w.nextProjectileID)
	displayCount := 1
	if attacker.HeroID == archerHeroID && tick < attacker.Archer.FocusActiveUntil {
		displayCount = 3
	}
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         kind,
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
		DisplayCount: displayCount,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	}
}

func (w *World) attackDamage(attacker *Entity, target *Entity, tick uint64, tickRate int) int {
	if parts := w.minionBasicAttackDamageParts(attacker, target, tick); len(parts) > 0 {
		total := 0
		for _, part := range parts {
			total += part.damage
		}
		return total
	}
	attack := attacker.Stats.Attack
	crit := false
	if attacker.HeroID == archerHeroID {
		attack *= w.archerBasicAttackMultiplier(attacker, target, tick)
	} else if w.attackCrits(attacker, target, tick) {
		crit = true
		attack *= w.critDamageMultiplier(attacker)
	}
	rawPhysical := attack + w.warriorQBonusDamage(attacker, tick) + w.tankWBonusDamage(attacker, tick)
	rawPhysical *= w.heroBasicAttackMultiplier(attacker, target, tick)
	rawPhysical = minionBasicAttackRawDamage(attacker, target, rawPhysical)
	if isMinion(target) {
		rawPhysical += w.equipmentMinionBasicAttackBonus(attacker, "physical")
	}
	rawPhysical += float64(w.heroBasicAttackBonusPhysicalDamage(attacker, target, tick, tickRate))
	damage := reduceCritDamage(target, w.applyCritFinalDamageMultiplier(attacker, physicalDamageAfterResistance(attacker, target, rawPhysical, tick), crit), crit)
	damage += magicDamageAfterResistance(attacker, target, w.equipmentBasicAttackBonus(attacker, "magic"), tick)
	damage += physicalDamageAfterResistance(attacker, target, w.equipmentBasicAttackBonus(attacker, "physical"), tick)
	if isMinion(target) {
		damage += magicDamageAfterResistance(attacker, target, w.equipmentMinionBasicAttackBonus(attacker, "magic"), tick)
	}
	damage += w.heroBasicAttackBonusMagicDamage(attacker, target, tick, tickRate)
	return damage
}

func (w *World) applyMinionBasicAttackDamage(attacker *Entity, target *Entity, tick uint64, tickRate int) bool {
	parts := w.minionBasicAttackDamageParts(attacker, target, tick)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if target.Stats.HP <= 0 {
			break
		}
		w.applyResolvedDamage(attacker, target, part.damage, part.damageType, sustainBasicAttack, tickRate)
	}
	w.applySiegeMinionSplash(attacker, target, tick, tickRate)
	return true
}

func (w *World) minionBasicAttackDamageParts(attacker *Entity, target *Entity, tick uint64) []basicAttackDamagePart {
	if attacker == nil || target == nil {
		return nil
	}
	rawDamage := minionBasicAttackRawDamage(attacker, target, attacker.Stats.Attack)
	switch attacker.Kind {
	case EntityKindRangedMinion:
		physicalDamage := physicalDamageAfterResistance(attacker, target, rawDamage, tick)
		parts := []basicAttackDamagePart{{damage: physicalDamage, damageType: "physical"}}
		if !IsHeroUnit(target) {
			parts = append(parts, basicAttackDamagePart{damage: magicDamageAfterResistance(attacker, target, float64(physicalDamage)*0.2, tick), damageType: "magic"})
		}
		return parts
	case EntityKindSiegeMinion:
		return []basicAttackDamagePart{{damage: physicalDamageAfterResistance(attacker, target, rawDamage, tick), damageType: "physical"}}
	default:
		return nil
	}
}

func (w *World) applySiegeMinionSplash(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.Kind != EntityKindSiegeMinion || tick < attacker.Combat.NextSiegeSplashTick {
		return
	}
	attacker.Combat.NextSiegeSplashTick = tick + secondsToTicks(siegeMinionSplashSeconds, tickRate)
	for _, hit := range w.targetsInRadius(attacker, target.Position, siegeMinionSplashRadius) {
		if hit.ID == target.ID {
			continue
		}
		key := attacker.ID + "->" + hit.ID
		w.siegeSplashBurns[key] = EquipmentBurn{
			SourceID:       attacker.ID,
			TargetID:       hit.ID,
			NextTick:       tick + secondsToTicks(siegeMinionSplashTickSeconds, tickRate),
			ExpiresAt:      tick + secondsToTicks(siegeMinionSplashSeconds, tickRate),
			FlatDamage:     attacker.Stats.Attack,
			BaseMaxHPRatio: siegeMinionSplashMaxHPRatio,
		}
	}
}

func (w *World) tickSiegeMinionSplashBurns(tick uint64, tickRate int) {
	for key, burn := range w.siegeSplashBurns {
		source := w.entities[burn.SourceID]
		target := w.entities[burn.TargetID]
		if tick > burn.ExpiresAt || source == nil || target == nil || target.Stats.HP <= 0 {
			delete(w.siegeSplashBurns, key)
			continue
		}
		if tick < burn.NextTick {
			continue
		}
		damage := magicDamageAfterResistance(source, target, burn.FlatDamage+target.Stats.MaxHP*burn.BaseMaxHPRatio, tick)
		target.Combat.LastHitTick = tick
		wasAlive := target.Stats.HP > 0
		w.applyResolvedDamage(source, target, damage, "magic", sustainAOESkill, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(source, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
		burn.NextTick += secondsToTicks(siegeMinionSplashTickSeconds, tickRate)
		if burn.NextTick > burn.ExpiresAt {
			delete(w.siegeSplashBurns, key)
			continue
		}
		w.siegeSplashBurns[key] = burn
	}
}

func (w *World) recordDummyBasicAttackDamage(attacker *Entity, target *Entity, damage int, tick uint64) {
	target.Combat.LastDamage = damage
	target.Combat.LastDamageType = "physical"
	parts := w.minionBasicAttackDamageParts(attacker, target, tick)
	if len(parts) == 0 {
		return
	}
	target.Combat.LastDamage = 0
	for _, part := range parts {
		target.Combat.LastDamage += part.damage
		target.Combat.LastDamageType = part.damageType
	}
}

func minionBasicAttackRawDamage(attacker *Entity, target *Entity, rawPhysical float64) float64 {
	if !isMinion(attacker) || target == nil {
		return rawPhysical
	}
	if target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero {
		rawPhysical *= 0.6
	}
	if target.Kind == EntityKindTower {
		if attacker.Kind == EntityKindSiegeMinion {
			rawPhysical *= 0.84
		} else {
			rawPhysical *= 0.6
		}
	}
	if isMinion(target) {
		ratio := 0.02
		if attacker.Kind == EntityKindRangedMinion {
			ratio = 0.04
		} else if attacker.Kind == EntityKindSiegeMinion {
			ratio = 0.05
		}
		rawPhysical += target.Stats.HP * ratio
	}
	return rawPhysical
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
	if heroHooksFor(warriorHeroID).QBonusDamage != nil {
		return heroHooksFor(warriorHeroID).QBonusDamage(w, attacker, tick)
	}
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

func (w *World) WarriorControlTicksAfterTenacity(target *Entity, ticks uint64, tick uint64) uint64 {
	return controlTicksAfterTenacity(target, ticks, tick)
}

func (w *World) tankWBonusDamage(attacker *Entity, tick uint64) float64 {
	if heroHooksFor(tankHeroID).WBonusDamage != nil {
		return heroHooksFor(tankHeroID).WBonusDamage(w, attacker, tick)
	}
	return 0
}

func isBasicAttackProjectileKind(kind string) bool {
	return kind == basicArrowProjectileKind || kind == siegeCannonballProjectileKind
}
