package warrior

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const (
	heroID = "warrior"
	qID    = "slash"
	wID    = "dash"
	eID    = "judgment"
	rID    = "justice"
)

func init() {
	world.RegisterHeroCastHandlers(heroID, map[string]world.HeroCastHandler{
		qID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			ApplyQ(w, entity, state, skill, tick, tickRate)
		},
		wID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			ApplyW(w, entity, state, skill, tick, tickRate)
		},
		eID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			ApplyE(w, entity, state, skill, tick, tickRate)
		},
		rID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
			ApplyR(w, entity, cast, state, skill, tick, tickRate)
		},
	})
}

func ApplyQ(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.DecisiveStrikeUntilTick = tick + secondsToTicks(skillMeta(skill, "empowerDurationSeconds", 4.5), tickRate)
	entity.Warrior.DecisiveStrikeSpeedUntilTick = tick + secondsToTicks(skillList(skill, "moveSpeedDurationSeconds", state.Level, []float64{1.5, 2, 2.5, 3, 3.5}), tickRate)
	entity.Warrior.DecisiveStrikeLevel = state.Level
	entity.Warrior.DecisiveStrikeMoveSpeedBonus = skillMeta(skill, "moveSpeedBonus", 0.3)
	entity.Combat.NextAttackTick = tick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(math.Round(skillList(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000}))), tickRate)
	entity.Skills[qID] = state
}

func ApplyW(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.CourageUntilTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate)
	entity.Warrior.CourageFrontUntilTick = tick + secondsToTicks(skillMeta(skill, "frontDurationSeconds", 0.75), tickRate)
	entity.Warrior.CourageFrontDamageReduce = skillMeta(skill, "frontDamageReduce", 0.6)
	entity.Warrior.CourageFrontTenacity = skillMeta(skill, "frontTenacity", 0.6)
	entity.Warrior.CourageBackDamageReduce = skillMeta(skill, "backDamageReduce", 0.3)
	entity.Control.TenacityUntilTick = entity.Warrior.CourageFrontUntilTick
	entity.Passive.MaxShield = warriorWShieldValue(entity, skill, state.Level)
	entity.Passive.Shield = entity.Passive.MaxShield
	entity.Passive.ShieldExpireTick = entity.Warrior.CourageUntilTick
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(math.Round(skillList(skill, "cooldownMs", state.Level, []float64{24000, 22000, 20000, 18000, 16000}))), tickRate)
	entity.Skills[wID] = state
}

func warriorWShieldValue(entity *world.Entity, skill config.SkillConfig, skillLevel int) int {
	baseShield := skillList(skill, "shieldValue", skillLevel, []float64{70, 95, 120, 145, 170})
	return int(math.Round(baseShield + float64(entity.Stats.BonusHP)*skillMeta(skill, "bonusHealthRatio", 0.2)))
}

func ApplyE(w *world.World, entity *world.Entity, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	durationTicks := secondsToTicks(skillMeta(skill, "durationSeconds", 3), tickRate)
	spins := warriorESpinCount(entity, skill)
	if durationTicks == 0 || spins <= 0 {
		return
	}
	intervalTicks := durationTicks / uint64(spins)
	if intervalTicks < 1 {
		intervalTicks = 1
	}
	entity.Warrior.JudgmentUntilTick = tick + durationTicks
	entity.Warrior.JudgmentNextSpinTick = tick
	entity.Warrior.JudgmentSpinIntervalTicks = intervalTicks
	entity.Warrior.JudgmentSpinsRemaining = spins
	entity.Warrior.JudgmentLevel = state.Level
	entity.Warrior.JudgmentHits = make(map[string]int)
	state.CooldownUntilTick = 0
	entity.Skills[eID] = state
}

func warriorESpinCount(entity *world.Entity, skill config.SkillConfig) int {
	baseSpins := int(skillMeta(skill, "baseSpins", 7))
	if entity == nil {
		return baseSpins
	}
	bonusPerSpin := skillMeta(skill, "attackSpeedBonusPerExtraSpin", 0.25)
	if bonusPerSpin <= 0 {
		return baseSpins
	}
	attackSpeedBonus := entity.Stats.AttackSpeedBonus
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	return baseSpins + int(math.Floor(attackSpeedBonus/bonusPerSpin))
}

func ApplyR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || entity.Warrior.JusticePending {
		return
	}
	target := warriorRTarget(w, entity, world.Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.435), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Warrior.JusticePending = true
	entity.Warrior.JusticeReleaseTick = tick + windupTicks
	entity.Warrior.JusticeTargetID = target.ID
	entity.Warrior.JusticeLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Warrior.JusticeReleaseTick
	entity.Skills[rID] = state
}

func warriorRTarget(w *world.World, entity *world.Entity, targetPoint world.Vector2, skill config.SkillConfig) *world.Entity {
	var best *world.Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 400)
	pickPadding := skillMeta(skill, "targetPickPadding", 80)
	w.ForEachEntity(func(target *world.Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		if distance(entity.Position, target.Position) > castRange+target.Radius {
			return
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+pickPadding {
			return
		}
		if distToPoint < bestDistance {
			bestDistance = distToPoint
			best = target
		}
	})
	return best
}

func canAttackTarget(attacker *world.Entity, target *world.Entity) bool {
	return attacker != nil && target != nil && attacker.Team != target.Team && target.Stats.HP > 0
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func distance(a world.Vector2, b world.Vector2) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(dx, dy)
}

func skillMeta(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	if value, ok := skill.Meta[key]; ok {
		return value
	}
	return fallback
}

func skillList(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if skill.MetaLists != nil && len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
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
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}
