package tank

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{30, 35, 40, 45, 50})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Tank.ThunderclapEmpowerUntil = tick + secondsToTicks(skillMeta(skill, "aftershockDurationSeconds", 5), tickRate)
	entity.Tank.ThunderclapAftershockUntil = entity.Tank.ThunderclapEmpowerUntil
	entity.Tank.ThunderclapLevel = state.Level
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{10000, 9500, 9000, 8500, 8000})), tickRate)
	entity.Skills[wID] = state
}

func RefreshWPassive(w *world.World, entity *world.Entity) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	if entity.Tank.ThunderclapArmorBonus != 0 {
		entity.Stats.PhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Stats.BonusPhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Tank.ThunderclapArmorBonus = 0
	}
	state, ok := entity.Skills[wID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.SkillConfig(wID)
	ratio := skillList(skill, "passiveArmorRatio", state.Level, []float64{0.1, 0.15, 0.2, 0.25, 0.3})
	if entity.Passive.Shield > 0 {
		ratio = skillMeta(skill, "shieldArmorRatio", 0.3)
	}
	baseArmor := entity.Stats.PhysicalDefense - entity.Stats.BonusPhysicalDefense
	bonus := baseArmor * ratio
	entity.Tank.ThunderclapArmorBonus = bonus
	entity.Stats.PhysicalDefense += bonus
	entity.Stats.BonusPhysicalDefense += bonus
}

func WBonusDamage(w *world.World, attacker *world.Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != heroID || tick >= attacker.Tank.ThunderclapEmpowerUntil {
		return 0
	}
	skill := w.SkillConfig(wID)
	level := attacker.Tank.ThunderclapLevel
	if level <= 0 {
		level = 1
	}
	return skillList(skill, "bonusDamage", level, []float64{30, 40, 50, 60, 70}) +
		float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.2) +
		attacker.Stats.PhysicalDefense*skillMeta(skill, "armorRatio", 0.15)
}

func ApplyWAftershock(w *world.World, attacker *world.Entity, primary *world.Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.HeroID != heroID || tick >= attacker.Tank.ThunderclapAftershockUntil {
		return
	}
	skill := w.SkillConfig(wID)
	level := attacker.Tank.ThunderclapLevel
	if level <= 0 {
		level = 1
	}
	damage := skillList(skill, "aftershockDamage", level, []float64{15, 25, 35, 45, 55}) +
		float64(attacker.Stats.AbilityPower)*skillMeta(skill, "aftershockAPRatio", 0.3) +
		attacker.Stats.PhysicalDefense*skillMeta(skill, "aftershockArmorRatio", 0.15)
	direction := world.Vector2{X: 1, Y: 0}
	if primary != nil {
		dx, dy := normalize(primary.Position.X-attacker.Position.X, primary.Position.Y-attacker.Position.Y)
		if dx != 0 || dy != 0 {
			direction = world.Vector2{X: dx, Y: dy}
		}
	}
	coneRange := skillMeta(skill, "aftershockConeRange", 300)
	coneAngle := skillMeta(skill, "aftershockConeAngleDegrees", 70)
	addWAftershockEffect(w, attacker, direction, coneRange, coneAngle, tick, tickRate)
	for _, target := range w.TankTargetsInCone(attacker, direction, coneRange, coneAngle) {
		target.Combat.LastHitTick = tick
		previousDamage := 0
		if primary != nil && target.ID == primary.ID {
			previousDamage = target.Combat.LastDamage
		}
		aftershockDamage := damage
		if isMonster(target) {
			aftershockDamage *= skillMeta(skill, "monsterDamageMultiplier", 1.8)
		}
		if target.Kind != world.EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.ApplyAOEDamage(attacker, target, w.TankPhysicalDamageAfterResistance(attacker, target, aftershockDamage, tick), "physical", tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(attacker, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = w.TankPhysicalDamageAfterResistance(attacker, target, aftershockDamage, tick)
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

func addWAftershockEffect(w *world.World, attacker *world.Entity, direction world.Vector2, coneRange float64, coneAngle float64, tick uint64, tickRate int) {
	if attacker == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.25))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	id := w.NextEffectID("effect:tank_w_aftershock:")
	w.PutSkillEffect(world.SkillEffect{
		ID:        id,
		Kind:      "tank_w_aftershock",
		Team:      attacker.Team,
		Start:     attacker.Position,
		Dir:       direction,
		Range:     coneRange,
		Radius:    coneAngle,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	})
}

func isMonster(entity *world.Entity) bool {
	if entity == nil {
		return false
	}
	switch entity.Kind {
	case world.EntityKindBlueBuff, world.EntityKindRedBuff, world.EntityKindGromp, world.EntityKindRaptor, world.EntityKindMurkWolf, world.EntityKindKrugCamp, world.EntityKindBaronNashor:
		return true
	default:
		return false
	}
}
