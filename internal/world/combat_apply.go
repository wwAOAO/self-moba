package world

import "math"

func (w *World) applyDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "physical", sustainSingleTargetSkill, tickRate)
}

func (w *World) applyMagicDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "magic", sustainSingleTargetSkill, tickRate)
}

func (w *World) applyTrueDamage(source *Entity, target *Entity, rawDamage float64, tickRate int) {
	w.applyResolvedDamage(source, target, trueDamageAfterReduction(target, rawDamage, target.Combat.LastHitTick), "true", sustainSingleTargetSkill, tickRate)
}

func (w *World) applyBasicAttackDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "physical", sustainBasicAttack, tickRate)
}

func (w *World) applyAOEDamage(source *Entity, target *Entity, damage int, damageType string, tickRate int) {
	w.applyResolvedDamage(source, target, damage, damageType, sustainAOESkill, tickRate)
}

func (w *World) applyResolvedDamage(source *Entity, target *Entity, damage int, damageType string, context sustainContext, tickRate int) {
	if damage <= 0 {
		target.Combat.LastDamage = 0
		target.Combat.LastDamageType = ""
		target.Combat.DamageEvents = nil
		target.Combat.DamageEventsTick = target.Combat.LastHitTick
		return
	}
	if context.BasicAttack && target.Stats.BasicAttackBlock > 0 {
		damage = int(math.Round(float64(damage) * (1 - clamp(target.Stats.BasicAttackBlock, 0, 1))))
		if damage < 1 {
			damage = 1
		}
	}
	damage = w.applyShield(source, target, damage, tickRate)
	target.Combat.LastDamage = damage
	target.Combat.LastDamageType = damageType
	if target.Combat.DamageEventsTick != target.Combat.LastHitTick {
		target.Combat.DamageEvents = nil
		target.Combat.DamageEventsTick = target.Combat.LastHitTick
	}
	target.Combat.DamageEvents = append(target.Combat.DamageEvents, DamageEvent{Damage: damage, DamageType: damageType})
	w.applyArcherFrostShot(source, target, target.Combat.LastHitTick, tickRate)
	w.breakTankGraniteShield(target, target.Combat.LastHitTick)
	if damage <= 0 {
		return
	}
	beforeHP := target.Stats.HP
	target.Stats.HP -= damage
	if target.Stats.HP < 0 {
		target.Stats.HP = 0
	}
	w.triggerEquipmentLowHealthShield(target, tickRate)
	w.triggerEquipmentHeroDamageManaShield(source, target, tickRate)
	if context.BasicAttack {
		w.triggerEquipmentBasicAttackAttackerSlow(source, target, tickRate)
		w.triggerEquipmentBasicAttackStacks(source, target.Combat.LastHitTick, tickRate)
	}
	if damageType == "magic" {
		w.triggerEquipmentMagicHitStacks(target)
	}
	w.triggerStoneplateCooldown(target, tickRate)
	if !context.BasicAttack && !context.Pet {
		w.applyEquipmentSkillBurn(source, target, target.Combat.LastHitTick, tickRate)
	}
	w.applySustain(source, beforeHP-target.Stats.HP, context)
	if damageType == "physical" {
		w.triggerEquipmentPhysicalDamageEffects(source, target, beforeHP-target.Stats.HP, target.Combat.LastHitTick, tickRate)
	}
	w.triggerSunfireCombat(source, target.Combat.LastHitTick, tickRate)
	w.triggerSunfireCombat(target, target.Combat.LastHitTick, tickRate)
	w.triggerEquipmentHeroHitHeal(source, target)
	w.breakWarriorToughness(source, target, target.Combat.LastHitTick)
}
