package shadowassassin

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func CastQ(w *world.World, entity *world.Entity, _ protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{40, 45, 50, 55, 60})
	if entity.Stats.MP < cost {
		return
	}

	startRRecall(w, entity)
	entity.Stats.MP -= cost
	entity.ShadowAssassin.QEmpowered = true
	entity.ShadowAssassin.QLevel = state.Level
	entity.Combat.NextAttackTick = tick
	entity.Combat.PendingAttackTargetID = ""
	entity.Combat.AttackReleaseTick = 0
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, skillList(skill, "cooldownMs", state.Level, []float64{8000, 7000, 6000, 5000, 4000}), tickRate)
	entity.Skills[qID] = state
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:shadow_assassin_q_cast:"),
		Kind:         "shadow_assassin_q_cast",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		TargetID:     entity.ID,
		Start:        entity.Position,
		End:          entity.Position,
		Radius:       entity.Radius,
		Count:        state.Level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "castEffectSeconds", 0.45), tickRate),
	})
}

func ActiveBuffs(_ *world.World, entity *world.Entity, _ uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID || !entity.ShadowAssassin.QEmpowered {
		return nil
	}
	return []world.BuffState{{ID: "shadow_assassin_q_ready", Name: "刺客诡道"}}
}

func QBonusPhysicalDamage(w *world.World, attacker *world.Entity, _ *world.Entity, _ uint64, _ int) int {
	if w == nil || attacker == nil || attacker.HeroID != heroID || !attacker.ShadowAssassin.QEmpowered || attacker.ShadowAssassin.QLevel <= 0 {
		return 0
	}
	skill := w.SkillConfig(qID)
	level := attacker.ShadowAssassin.QLevel
	raw := skillList(skill, "bonusDamage", level, []float64{30, 60, 90, 120, 150})
	raw += attacker.Stats.BonusAttack * skillMeta(skill, "bonusAdRatio", 0.3)
	return int(math.Round(raw))
}

func OnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if w == nil || attacker == nil || target == nil || attacker.HeroID != heroID || !attacker.ShadowAssassin.QEmpowered {
		return
	}
	level := attacker.ShadowAssassin.QLevel
	attacker.ShadowAssassin.QEmpowered = false
	attacker.ShadowAssassin.QLevel = 0
	if level <= 0 || tickRate <= 0 || attacker.Team == target.Team {
		return
	}
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:shadow_assassin_q_hit:"),
		Kind:         "shadow_assassin_q_hit",
		Team:         attacker.Team,
		SourceID:     attacker.ID,
		SourceHeroID: attacker.HeroID,
		TargetID:     target.ID,
		Start:        attacker.Position,
		End:          target.Position,
		Radius:       target.Radius,
		Count:        level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(w.SkillConfig(qID), "hitEffectSeconds", 0.35), tickRate),
	})
	if !world.IsHeroUnit(target) || target.Stats.HP <= 0 {
		return
	}
	applyBleed(w, attacker, target, level, tick, tickRate)
}

func applyBleed(w *world.World, source *world.Entity, target *world.Entity, level int, tick uint64, tickRate int) {
	skill := w.SkillConfig(qID)
	durationTicks := secondsToTicks(skillMeta(skill, "bleedDurationSeconds", 6), tickRate)
	tickTicks := secondsToTicks(skillMeta(skill, "bleedTickSeconds", 1), tickRate)
	if durationTicks == 0 || tickTicks == 0 {
		return
	}
	if target.ShadowAssassin.Bleeds == nil {
		target.ShadowAssassin.Bleeds = make(map[string]world.ShadowAssassinBleedState)
	}
	effectID := ""
	if previous, ok := target.ShadowAssassin.Bleeds[source.ID]; ok {
		effectID = previous.EffectID
	}
	if effectID == "" {
		effectID = w.NextEffectID("effect:shadow_assassin_q_reveal:")
	}
	expiresAt := tick + durationTicks
	target.ShadowAssassin.Bleeds[source.ID] = world.ShadowAssassinBleedState{
		Level:         level,
		ExpiresAtTick: expiresAt,
		NextTick:      tick + tickTicks,
		EffectID:      effectID,
	}
	w.ApplyMoveSpeedSlow(target, skillMeta(skill, "bleedSlow", 0.1), expiresAt)
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         "shadow_assassin_q_reveal",
		Team:         source.Team,
		SourceID:     source.ID,
		SourceHeroID: source.HeroID,
		TargetID:     target.ID,
		Start:        target.Position,
		End:          target.Position,
		Radius:       target.Radius,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || tickRate <= 0 {
		return
	}
	if entity.HeroID == heroID && entity.ShadowAssassin.RActive && tick >= entity.ShadowAssassin.RUntil {
		startRRecall(w, entity)
	}
	if len(entity.ShadowAssassin.Bleeds) == 0 {
		return
	}
	for sourceID, bleed := range entity.ShadowAssassin.Bleeds {
		source := w.EntityByID(sourceID)
		if source == nil || entity.Stats.HP <= 0 || tick > bleed.ExpiresAtTick {
			delete(entity.ShadowAssassin.Bleeds, sourceID)
			w.RemoveSkillEffect(bleed.EffectID)
			continue
		}
		if tick < bleed.NextTick {
			continue
		}
		skill := w.SkillConfig(qID)
		duration := skillMeta(skill, "bleedDurationSeconds", 6)
		interval := skillMeta(skill, "bleedTickSeconds", 1)
		total := skillList(skill, "bleedDamage", bleed.Level, []float64{10, 20, 30, 40, 50})
		total += source.Stats.BonusAttack * skillMeta(skill, "bleedBonusAdRatio", 1.2)
		value := bleed.Remainder + total*interval/duration
		rawDamage := int(math.Floor(value + 0.000000001))
		bleed.Remainder = value - float64(rawDamage)
		bleed.NextTick += secondsToTicks(interval, tickRate)
		entity.ShadowAssassin.Bleeds[sourceID] = bleed
		if rawDamage > 0 {
			entity.Combat.LastHitTick = tick
			wasAlive := entity.Stats.HP > 0
			damage := w.PhysicalDamageAfterResistance(source, entity, float64(rawDamage), tick)
			w.ApplyPetDamage(source, entity, damage, "physical", tickRate)
			if wasAlive && entity.Stats.HP == 0 {
				w.ApplyKillReward(source, entity)
				w.KillPlayer(entity, tick, tickRate)
				w.RemoveDeadUnit(entity)
			}
		}
		if tick >= bleed.ExpiresAtTick {
			delete(entity.ShadowAssassin.Bleeds, sourceID)
		}
	}
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if configured := skill.MetaLists[key]; len(configured) > 0 {
		values = configured
	}
	if level < 1 {
		level = 1
	}
	if level > len(values) {
		level = len(values)
	}
	return values[level-1]
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func cooldownTicksFor(entity *world.Entity, cooldownMS float64, tickRate int) uint64 {
	if cooldownMS <= 0 || tickRate <= 0 {
		return 0
	}
	haste := math.Max(entity.Stats.AbilityHaste, 0)
	return uint64(math.Ceil(cooldownMS / 1000 / (1 + haste/100) * float64(tickRate)))
}
