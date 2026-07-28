package shadowassassin

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func CastR(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 || entity.ShadowAssassin.RActive {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{80, 90, 100})
	if entity.Stats.MP < cost {
		return
	}

	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillList(skill, "cooldownMs", state.Level, []float64{75000, 65000, 55000}), tickRate)
	entity.Skills[rID] = state
	entity.ShadowAssassin.RActive = true
	entity.ShadowAssassin.RUntil = tick + secondsToTicks(skillMeta(skill, "invisibilitySeconds", 2.5), tickRate)
	entity.ShadowAssassin.RLevel = state.Level
	entity.ShadowAssassin.RGroupID = w.NextEffectID("effect:shadow_assassin_r_group:")
	entity.ShadowAssassin.RMoveSpeedMultiplier = 1 + skillMeta(skill, "moveSpeedBonus", 0.4)
	entity.Control.InvisibleUntilTick = entity.ShadowAssassin.RUntil
	spawnRBlades(w, entity, state.Level, skill, tick, tickRate)
}

func spawnRBlades(w *world.World, entity *world.Entity, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	count := int(skillMeta(skill, "bladeCount", 24))
	if count < 1 {
		count = 1
	}
	rangeValue := skillMeta(skill, "projectileRange", 550)
	hitIDs := make(map[string]bool)
	for index := 0; index < count; index++ {
		angle := math.Pi * 2 * float64(index) / float64(count)
		w.PutProjectile(&world.Projectile{
			ID:           w.NextProjectileID("projectile:shadow_assassin_r:"),
			Kind:         "shadow_assassin_r",
			Team:         entity.Team,
			SourceID:     entity.ID,
			SkillID:      rID,
			GroupID:      entity.ShadowAssassin.RGroupID,
			Position:     entity.Position,
			Start:        entity.Position,
			Dir:          world.Vector2{X: math.Cos(angle), Y: math.Sin(angle)},
			SpeedPerTick: skillMeta(skill, "projectileSpeed", 1200) / float64(tickRate),
			Range:        rangeValue,
			Radius:       skillMeta(skill, "projectileRadius", 55),
			Damage:       level,
			Recallable:   true,
			CreatedAt:    tick,
			ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "returnLifetimeSeconds", 5), tickRate),
			HitIDs:       hitIDs,
			DisplayCount: index + 1,
		})
	}
}

func startRRecall(w *world.World, entity *world.Entity) {
	if w == nil || entity == nil || !entity.ShadowAssassin.RActive {
		return
	}
	entity.ShadowAssassin.RActive = false
	entity.ShadowAssassin.RUntil = 0
	entity.ShadowAssassin.RMoveSpeedMultiplier = 0
	entity.Control.InvisibleUntilTick = 0
	w.RecallProjectileGroup(entity.ShadowAssassin.RGroupID)
}

func BreakROnAttack(w *world.World, entity *world.Entity, _ *world.Entity, _ uint64, _ int) {
	startRRecall(w, entity)
}

func SpecialRecast(w *world.World, entity *world.Entity, cast protocol.CastInput, _ world.SkillState, _ config.SkillConfig, _ uint64, _ int) bool {
	if entity == nil || cast.SkillID != rID || !entity.ShadowAssassin.RActive {
		return false
	}
	startRRecall(w, entity)
	return true
}

func RMoveSpeedMultiplier(entity *world.Entity, tick uint64) float64 {
	if entity == nil || entity.HeroID != heroID || !entity.ShadowAssassin.RActive || tick >= entity.ShadowAssassin.RUntil {
		return 1
	}
	if entity.ShadowAssassin.RMoveSpeedMultiplier > 0 {
		return entity.ShadowAssassin.RMoveSpeedMultiplier
	}
	return 1.4
}

func ResolveProjectile(w *world.World, source *world.Entity, projectile *world.Projectile, previousPosition world.Vector2, tick uint64, tickRate int) bool {
	if projectile == nil {
		return false
	}
	if projectile.SkillID == wID {
		return ResolveWProjectile(w, source, projectile, previousPosition, tick, tickRate)
	}
	if projectile.SkillID != rID {
		return false
	}
	if w == nil || source == nil || source.HeroID != heroID || tickRate <= 0 {
		return true
	}
	phase := "out:"
	if projectile.Returning && (projectile.EffectTicks == 0 || tick > projectile.EffectTicks) {
		phase = "return:"
	}
	w.ForEachEntity(func(target *world.Entity) {
		if !world.CanAttackTarget(source, target) || !segmentHitsCircle(previousPosition, projectile.Position, target.Position, projectile.Radius+target.Radius) {
			return
		}
		key := phase + target.ID
		if projectile.HitIDs[key] {
			return
		}
		projectile.HitIDs[key] = true
		applyRHit(w, source, target, projectile.Damage, tick, tickRate)
	})
	return true
}

func applyRHit(w *world.World, source *world.Entity, target *world.Entity, level int, tick uint64, tickRate int) {
	skill := w.SkillConfig(rID)
	raw := skillList(skill, "damage", level, []float64{120, 190, 260})
	raw += source.Stats.BonusAttack * skillMeta(skill, "bonusAdRatio", 0.9)
	damage := w.PhysicalDamageAfterResistance(source, target, raw, tick)
	target.Combat.LastHitTick = tick
	wasAlive := target.Stats.HP > 0
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
	} else {
		w.ApplyAOEDamage(source, target, damage, "physical", tickRate)
	}
	if wasAlive && target.Stats.HP == 0 {
		w.ApplyKillReward(source, target)
		w.KillPlayer(target, tick, tickRate)
		w.RemoveDeadUnit(target)
	}
}
