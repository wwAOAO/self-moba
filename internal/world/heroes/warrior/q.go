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
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
				ApplyQ(w, entity, state, skill, tick, tickRate)
			},
			wID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
				ApplyW(w, entity, state, skill, tick, tickRate)
			},
			eID: func(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
				ApplyE(w, entity, state, skill, tick, tickRate)
			},
			rID: ApplyR,
		},
		ReleaseR:            ReleaseR,
		StopE:               StopE,
		TickE:               TickE,
		QBonusDamage:        QBonusDamage,
		ConsumeQ:            ConsumeQ,
		ApplyWPassiveKill:   ApplyWPassiveKill,
		TickToughness:       TickToughness,
		ToughnessRegenRatio: ToughnessRegenRatio,
		WarriorRDamage:      RDamage,
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

func QBonusDamage(w *world.World, attacker *world.Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != heroID || tick >= attacker.Warrior.DecisiveStrikeUntilTick {
		return 0
	}
	skill := w.SkillConfig(qID)
	level := attacker.Warrior.DecisiveStrikeLevel
	if level <= 0 {
		level = 1
	}
	baseDamage := skillList(skill, "bonusDamage", level, []float64{30, 60, 90, 120, 150})
	return baseDamage + attacker.Stats.Attack*skillMeta(skill, "totalAdRatio", 1.4)
}

func ConsumeQ(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.HeroID != heroID || tick >= attacker.Warrior.DecisiveStrikeUntilTick {
		return
	}
	skill := w.SkillConfig(qID)
	if target != nil {
		silenceTicks := secondsToTicks(skillMeta(skill, "silenceSeconds", 1.5), tickRate)
		target.Control.SilencedUntilTick = tick + w.WarriorControlTicksAfterTenacity(target, silenceTicks, tick)
	}
	attacker.Warrior.DecisiveStrikeUntilTick = 0
	attacker.Warrior.DecisiveStrikeLevel = 0
}
