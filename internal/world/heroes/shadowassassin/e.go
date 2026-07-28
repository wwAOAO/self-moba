package shadowassassin

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	target := w.EntityByID(cast.TargetID)
	if !validETarget(entity, target, tick) || distance(entity.Position, target.Position) > skillRange(skill, 700)+target.Radius {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{35, 40, 45, 50, 55})
	if entity.Stats.MP < cost {
		return
	}

	start := entity.Position
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	behindDistance := entity.Radius + target.Radius + skillMeta(skill, "landingPadding", 8)
	destination := w.ClampWorldPoint(world.Vector2{
		X: target.Position.X + dx*behindDistance,
		Y: target.Position.Y + dy*behindDistance,
	})

	startRRecall(w, entity)
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillList(skill, "cooldownMs", state.Level, []float64{18000, 16000, 14000, 12000, 10000}), tickRate)
	entity.Skills[eID] = state
	entity.Position = destination
	entity.Intent = world.IntentState{}
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0

	rootUntil := tick + secondsToTicks(skillMeta(skill, "rootSeconds", 1), tickRate)
	if target.Control.RootedUntilTick > rootUntil {
		rootUntil = target.Control.RootedUntilTick
	}
	w.ApplyRoot(target, rootUntil, tick, tickRate)
	entity.ShadowAssassin.ETargetID = target.ID
	entity.ShadowAssassin.EDamageAmp = skillList(skill, "damageAmp", state.Level, []float64{0.03, 0.06, 0.09, 0.12, 0.15})
	entity.ShadowAssassin.EDamageUntil = tick + secondsToTicks(skillMeta(skill, "damageAmpSeconds", 3), tickRate)

	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:shadow_assassin_e:"),
		Kind:         "shadow_assassin_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		TargetID:     target.ID,
		Start:        start,
		End:          target.Position,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "effectSeconds", 0.3), tickRate),
	})
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:shadow_assassin_e_mark:"),
		Kind:         "shadow_assassin_e_mark",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		TargetID:     target.ID,
		Start:        target.Position,
		End:          target.Position,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    entity.ShadowAssassin.EDamageUntil,
	})
}

func EDamageMultiplier(_ *world.World, attacker *world.Entity, target *world.Entity, _ string, tick uint64) float64 {
	if attacker == nil || target == nil || attacker.HeroID != heroID || attacker.ShadowAssassin.ETargetID != target.ID || tick >= attacker.ShadowAssassin.EDamageUntil {
		return 1
	}
	return 1 + math.Max(attacker.ShadowAssassin.EDamageAmp, 0)
}

func validETarget(source *world.Entity, target *world.Entity, tick uint64) bool {
	return source != nil && target != nil && world.IsHeroUnit(target) && source.Team != target.Team &&
		target.Stats.HP > 0 && !target.Death.Dead && tick >= target.Control.UntargetableUntilTick
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}
