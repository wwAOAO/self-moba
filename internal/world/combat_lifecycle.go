package world

func (w *World) killPlayer(target *Entity, tick uint64, tickRate int) {
	if target.Kind != EntityKindPlayer || target.Death.Dead {
		return
	}
	w.removeProjectilesTargeting(target)
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
	target.Passive.Bleeds = nil
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
	if heroHooksFor(swordHeroID).ApplyShield != nil {
		heroHooksFor(swordHeroID).ApplyShield(w, source, target, tickRate)
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

func (w *World) swordShieldValue(entity *Entity) int {
	if heroHooksFor(swordHeroID).ShieldValue == nil {
		return 0
	}
	return heroHooksFor(swordHeroID).ShieldValue(w, entity)
}

func (w *World) removeDeadUnit(target *Entity) {
	if target.Kind == EntityKindPlayer || target.Kind == EntityKindDummy {
		return
	}
	w.removeProjectilesTargeting(target)
	delete(w.entities, target.ID)
	for _, entity := range w.entities {
		if entity.Intent.AttackTargetID == target.ID {
			entity.Intent.AttackTargetID = ""
		}
	}
}

func (w *World) removeProjectilesTargeting(target *Entity) {
	if target == nil {
		return
	}
	for id, projectile := range w.projectiles {
		if projectile.TargetID != target.ID {
			continue
		}
		projectile.Position = target.Position
		delete(w.projectiles, id)
		w.cleanupProjectileGroup(projectile)
	}
}
