package world

import "math"

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
	consumeShieldLayers(target, absorbed)
	consumeEquipmentPhysicalDamageShield(target, absorbed)
	target.Passive.Shield -= absorbed
	if absorbed > 0 {
		w.triggerStoneplateCooldown(target, tickRate)
	}
	if target.Passive.Shield <= 0 {
		if deactivateStoneplateShield(target) {
			w.recalculatePlayerStats(target)
		}
	}
	return damage - absorbed
}

func consumeShieldLayers(target *Entity, absorbed int) {
	if target == nil || absorbed <= 0 || len(target.Passive.ShieldLayers) == 0 {
		return
	}
	remaining := absorbed
	for i := range target.Passive.ShieldLayers {
		if remaining <= 0 {
			break
		}
		if target.Passive.ShieldLayers[i].Amount <= remaining {
			remaining -= target.Passive.ShieldLayers[i].Amount
			target.Passive.ShieldLayers[i].Amount = 0
			continue
		}
		target.Passive.ShieldLayers[i].Amount -= remaining
		remaining = 0
	}
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
