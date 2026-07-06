package world

import (
	"math"
	"strconv"
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
	if distance(attacker.Position, target.Position) > w.attackReachAtTick(attacker, target, tick) {
		return
	}
	w.resolveBasicAttack(attacker, target, tick, tickRate)
}

func (w *World) resolveBasicAttack(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if isRangedBasicAttacker(attacker) {
		w.fireBasicAttackProjectile(attacker, target, tick, tickRate)
		return
	}

	damage := w.attackDamage(attacker, target, tick)
	target.Combat.LastHitTick = tick
	target.Combat.DamageEvents = nil
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyBasicAttackDamage(attacker, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(attacker, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
	}
	w.onHeroBasicHit(attacker, target, tick, tickRate)
}

func isRangedBasicAttacker(attacker *Entity) bool {
	return attacker != nil && (attacker.HeroID == archerHeroID || attacker.HeroID == mageHeroID || attacker.HeroID == gunnerHeroID || attacker.Kind == EntityKindRangedMinion || attacker.Kind == EntityKindSiegeMinion)
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
	w.nextProjectileID++
	id := "projectile:basic_arrow:" + strconv.Itoa(w.nextProjectileID)
	displayCount := 1
	if attacker.HeroID == archerHeroID && tick < attacker.Archer.FocusActiveUntil {
		displayCount = 3
	}
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
		DisplayCount: displayCount,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	}
}

func (w *World) attackDamage(attacker *Entity, target *Entity, tick uint64) int {
	attack := attacker.Stats.Attack
	crit := false
	if attacker.HeroID == archerHeroID {
		attack *= w.archerBasicAttackMultiplier(attacker, target, tick)
	} else if w.attackCrits(attacker, target, tick) {
		crit = true
		attack *= w.critDamageMultiplier(attacker)
	}
	rawPhysical := attack + w.warriorQBonusDamage(attacker, tick) + w.tankWBonusDamage(attacker, tick)
	rawPhysical = minionBasicAttackRawDamage(attacker, target, rawPhysical)
	if isMinion(target) {
		rawPhysical += w.equipmentMinionBasicAttackBonus(attacker, "physical")
	}
	damage := reduceCritDamage(target, w.applyCritFinalDamageMultiplier(attacker, physicalDamageAfterResistance(attacker, target, rawPhysical, tick), crit), crit)
	damage += magicDamageAfterResistance(attacker, target, w.equipmentBasicAttackBonus(attacker, "magic"), tick)
	damage += physicalDamageAfterResistance(attacker, target, w.equipmentBasicAttackBonus(attacker, "physical"), tick)
	if isMinion(target) {
		damage += magicDamageAfterResistance(attacker, target, w.equipmentMinionBasicAttackBonus(attacker, "magic"), tick)
	}
	return damage
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
		rawPhysical += float64(target.Stats.HP) * ratio
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
