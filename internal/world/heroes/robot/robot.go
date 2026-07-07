package robot

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"l-battle/internal/world/formula"
	"math"
)

const (
	heroID = "robot"
	qID    = "robot_q"
	wID    = "robot_w"
	eID    = "robot_e"
	rID    = "robot_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		Tick:                           Tick,
		TickEntity:                     TickArcMarks,
		OnBasicHit:                     OnBasicHit,
		OnDamaged:                      TriggerManaBarrier,
		ActiveBuffs:                    ActiveBuffs,
		ApplyStats:                     ApplyStats,
		BasicAttackMultiplier:          EAttackMultiplier,
		BasicAttackBonusPhysicalDamage: EBonusPhysicalDamage,
		BasicAttackBonusMagicDamage:    WBonusMagicDamage,
		RobotQDamage:                   QDamage,
	})
}

func TriggerManaBarrier(w *world.World, source *world.Entity, entity *world.Entity, basicAttack bool, pet bool, skipBleed bool, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Stats.HP <= 0 || entity.Stats.MaxHP <= 0 || tick < entity.Passive.RobotShieldCDUntil || entity.Passive.RobotShieldMana > 0 {
		return
	}
	skill := w.HeroPassiveSkill(entity)
	if entity.Stats.HP/entity.Stats.MaxHP >= skillMeta(skill, "threshold", 0.2) {
		return
	}
	shield := int(math.Floor(entity.Stats.MaxMP * skillMeta(skill, "shieldMaxManaRatio", 0.3)))
	if shield <= 0 {
		return
	}
	if entity.Stats.MP < float64(shield) {
		shield = int(math.Floor(entity.Stats.MP))
	}
	if shield <= 0 {
		return
	}
	entity.Stats.MP -= float64(shield)
	entity.Passive.Shield += shield
	entity.Passive.MaxShield += shield
	entity.Passive.RobotShieldMana = shield
	entity.Passive.RobotShieldUntil = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 10), tickRate)
	entity.Passive.RobotShieldCDUntil = tick + secondsToTicks(skillMeta(skill, "cooldownSeconds", 90), tickRate)
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	releaseQ(w, entity, tick, tickRate)
	releaseR(w, entity, tick, tickRate)
	tickW(w, entity, tick, tickRate)
	tickE(w, entity, tick)
	tickManaBarrier(entity, tick)
}

func tickManaBarrier(entity *world.Entity, tick uint64) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.RobotShieldMana <= 0 || tick < entity.Passive.RobotShieldUntil {
		return
	}
	refund := entity.Passive.RobotShieldMana
	if refund > entity.Passive.Shield {
		refund = entity.Passive.Shield
	}
	entity.Passive.Shield -= refund
	entity.Passive.MaxShield -= refund
	if entity.Passive.MaxShield > entity.Passive.Shield {
		entity.Passive.MaxShield = entity.Passive.Shield
	}
	if entity.Passive.MaxShield < 0 {
		entity.Passive.MaxShield = 0
	}
	entity.Stats.MP += float64(refund)
	if entity.Stats.MP > entity.Stats.MaxMP {
		entity.Stats.MP = entity.Stats.MaxMP
	}
	entity.Passive.RobotShieldMana = 0
	entity.Passive.RobotShieldUntil = 0
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.RobotQPending {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{100, 100, 100, 100, 100})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.RobotQPending = true
	entity.Passive.RobotQReleaseTick = tick + windupTicks
	entity.Passive.RobotQTarget = w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	entity.Passive.RobotQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.RobotQReleaseTick
	entity.Skills[qID] = state
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 75)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	entity.Passive.RobotWStartTick = tick
	entity.Passive.RobotWUntil = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 5), tickRate)
	entity.Passive.RobotWLevel = state.Level
	entity.Passive.RobotWMoveSpeed = skillList(skill, "moveSpeedBonus", state.Level, []float64{0.7, 0.75, 0.8, 0.85, 0.9})
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{15000, 15000, 15000, 15000, 15000})), tickRate)
	entity.Skills[wID] = state
	w.RefreshPlayerStats(entity)
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 {
		return
	}
	if state.Stacks > 0 && state.StacksExpireTick > 0 && tick >= state.StacksExpireTick {
		state.Stacks = 0
		state.StacksExpireTick = 0
	}
	if state.Stacks > 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 25)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 5), tickRate)
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{9000, 8000, 7000, 6000, 5000})), tickRate)
	entity.Combat.NextAttackTick = tick
	entity.Skills[eID] = state
	w.RefreshPlayerStats(entity)
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || state.Stacks > 0 {
		return
	}
	cost := skillMeta(skill, "manaCost", 100)
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.2), tickRate)
	entity.Control.ActionLockedUntilTick = state.StacksExpireTick
	entity.Skills[rID] = state
}

func tickW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.RobotWUntil == 0 {
		return
	}
	if tick >= entity.Passive.RobotWUntil {
		skill := w.SkillConfig(wID)
		entity.Passive.RobotWStartTick = 0
		entity.Passive.RobotWUntil = 0
		entity.Passive.RobotWLevel = 0
		entity.Passive.RobotWMoveSpeed = 0
		w.ApplyMoveSpeedSlow(entity, skillMeta(skill, "slow", 0.3), tick+secondsToTicks(skillMeta(skill, "slowSeconds", 1.5), tickRate))
		w.RefreshPlayerStats(entity)
		return
	}
	next := wMoveSpeed(w.SkillConfig(wID), entity.Passive.RobotWLevel, tick-entity.Passive.RobotWStartTick, tickRate)
	if math.Abs(next-entity.Passive.RobotWMoveSpeed) > 0.000001 {
		entity.Passive.RobotWMoveSpeed = next
		w.RefreshPlayerStats(entity)
	}
}

func releaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	state := entity.Skills[rID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		entity.Skills[rID] = state
		return
	}
	skill := w.SkillConfig(rID)
	level := state.Level
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{60000, 40000, 20000})), tickRate)
	entity.Skills[rID] = state
	radius := skillRange(skill, 600)
	w.PutSkillEffect(world.SkillEffect{
		ID:           w.NextEffectID("effect:robot_r:"),
		Kind:         "robot_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(skillMeta(skill, "rangeDisplaySeconds", 0.45), tickRate),
	})
	damageRaw := skillList(skill, "activeDamage", level, []float64{250, 375, 500}) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "activeAPRatio", 1)
	silenceTicks := secondsToTicks(skillMeta(skill, "silenceSeconds", 0.5), tickRate)
	for _, target := range w.TargetsInRadius(entity, entity.Position, radius) {
		if !world.CanAttackTarget(entity, target) {
			continue
		}
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		wasAlive := target.Stats.HP > 0
		w.RemoveAllShields(target)
		damage := w.MagicDamageAfterResistance(entity, target, damageRaw, tick)
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
		} else {
			w.ApplyMagicDamage(entity, target, damage, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.ApplyKillReward(entity, target)
				w.KillPlayer(target, tick, tickRate)
				w.RemoveDeadUnit(target)
			}
		}
		target.Control.SilencedUntilTick = tick + world.ControlTicksAfterTenacity(target, silenceTicks, tick)
	}
}

func tickE(w *world.World, entity *world.Entity, tick uint64) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	state := entity.Skills[eID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[eID] = state
	w.RefreshPlayerStats(entity)
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || !entity.Passive.RobotQPending || tick < entity.Passive.RobotQReleaseTick {
		return
	}
	target := entity.Passive.RobotQTarget
	level := entity.Passive.RobotQLevel
	entity.Passive.RobotQPending = false
	entity.Passive.RobotQReleaseTick = 0
	entity.Passive.RobotQTarget = world.Vector2{}
	entity.Passive.RobotQLevel = 0
	if level <= 0 {
		level = 1
	}
	skill := w.SkillConfig(qID)
	state := entity.Skills[qID]
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", level, []float64{20000, 19000, 18000, 17000, 16000})), tickRate)
	entity.Skills[qID] = state
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	qRange := skillRange(skill, 925)
	speedPerSecond := skillMeta(skill, "projectileSpeed", 1800)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:robot_q:"),
		Kind:         "robot_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 70),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speedPerSecond+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	rawDamage := skillList(skill, "baseDamage", skillLevel, []float64{80, 135, 190, 245, 300}) + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 1)
	return w.MagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || entity.HeroID != heroID || stats == nil {
		return
	}
	if entity.Passive.RobotWUntil > 0 {
		stats.MoveSpeed *= 1 + entity.Passive.RobotWMoveSpeed
		stats.AttackSpeedBonus += skillList(w.SkillConfig(wID), "attackSpeedBonus", entity.Passive.RobotWLevel, []float64{0.3, 0.38, 0.46, 0.54, 0.62})
		stats.AttackSpeed = formula.FinalAttackSpeed(stats.BaseAttackSpeed, stats.AttackSpeedBonus, stats.AttackSpeedRatio, stats.AttackSpeedSlow)
	}
	if eActive(entity, 0) {
		stats.AttackRange = math.Max(stats.AttackRange, skillRange(w.SkillConfig(eID), 300))
	}
}

func EAttackMultiplier(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64) float64 {
	if !eActive(attacker, tick) {
		return 1
	}
	return skillMeta(w.SkillConfig(eID), "attackMultiplier", 2)
}

func EBonusPhysicalDamage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) int {
	if !eActive(attacker, tick) {
		return 0
	}
	return int(math.Round(float64(attacker.Stats.AbilityPower) * skillMeta(w.SkillConfig(eID), "apRatio", 0.25)))
}

func OnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	ApplyEOnBasicHit(w, attacker, target, tick, tickRate)
	ApplyRMarkOnBasicHit(w, attacker, target, tick, tickRate)
}

func ApplyEOnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if !eActive(attacker, tick) || target == nil {
		return
	}
	state := attacker.Skills[eID]
	state.Stacks = 0
	state.StacksExpireTick = 0
	attacker.Skills[eID] = state
	w.RefreshPlayerStats(attacker)
	w.InterruptControl(target)
	knockupTicks := secondsToTicks(skillMeta(w.SkillConfig(eID), "knockupSeconds", 1), tickRate)
	target.Control.AirborneUntilTick = tick + world.ControlTicksAfterTenacity(target, knockupTicks, tick)
}

func eActive(entity *world.Entity, tick uint64) bool {
	if entity == nil || entity.HeroID != heroID {
		return false
	}
	state := entity.Skills[eID]
	return state.Stacks > 0 && state.Level > 0 && (tick == 0 || tick < state.StacksExpireTick)
}

func ApplyRMarkOnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.HeroID != heroID || tickRate <= 0 || !rPassiveEnabled(attacker, tick) {
		return
	}
	if target.Passive.RobotArcMarks == nil {
		target.Passive.RobotArcMarks = map[string]world.RobotArcState{}
	}
	mark := target.Passive.RobotArcMarks[attacker.ID]
	if mark.Stacks < 3 {
		mark.Stacks++
	}
	if mark.TriggerTick == 0 || tick >= mark.TriggerTick {
		mark.TriggerTick = tick + secondsToTicks(skillMeta(w.SkillConfig(rID), "passiveDelaySeconds", 1), tickRate)
	}
	target.Passive.RobotArcMarks[attacker.ID] = mark
}

func TickArcMarks(w *world.World, target *world.Entity, tick uint64, tickRate int) {
	if target == nil || len(target.Passive.RobotArcMarks) == 0 {
		return
	}
	for sourceID, mark := range target.Passive.RobotArcMarks {
		if mark.Stacks <= 0 || mark.TriggerTick == 0 || tick < mark.TriggerTick {
			continue
		}
		delete(target.Passive.RobotArcMarks, sourceID)
		source := w.EntityByID(sourceID)
		if source == nil || source.HeroID != heroID || source.Stats.HP <= 0 || source.Death.Dead || !rPassiveEnabled(source, tick) || !world.CanAttackTarget(source, target) {
			continue
		}
		level := source.Skills[rID].Level
		if level <= 0 {
			continue
		}
		skill := w.SkillConfig(rID)
		rawDamage := float64(mark.Stacks) * (skillList(skill, "passiveDamage", level, []float64{50, 100, 150}) + float64(source.Stats.AbilityPower)*skillMeta(skill, "passiveAPRatio", 0.3))
		damage := w.MagicDamageAfterResistance(source, target, rawDamage, tick)
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
}

func rPassiveEnabled(entity *world.Entity, tick uint64) bool {
	if entity == nil || entity.HeroID != heroID {
		return false
	}
	state := entity.Skills[rID]
	return state.Level > 0 && state.Stacks <= 0 && tick >= state.CooldownUntilTick
}

func WBonusMagicDamage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) int {
	if attacker == nil || target == nil || attacker.HeroID != heroID || attacker.Passive.RobotWUntil == 0 || tick >= attacker.Passive.RobotWUntil || target.Stats.MaxHP <= 0 {
		return 0
	}
	rawDamage := target.Stats.MaxHP * skillMeta(w.SkillConfig(wID), "maxHPMagicDamageRatio", 0.01)
	if world.IsMinion(target) || world.IsMonster(target) {
		rawDamage = math.Min(rawDamage, skillCurve(w.SkillConfig(wID), "minionMonsterDamageCap", "minionMonsterDamageCapLevels", attacker.Level, 60))
	}
	return w.MagicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID {
		return nil
	}
	buffs := make([]world.BuffState, 0, 3)
	if entity.Passive.RobotShieldMana > 0 && tick < entity.Passive.RobotShieldUntil {
		buffs = append(buffs, world.BuffState{
			ID:            "robot_mana_barrier",
			Name:          "Mana Barrier",
			ExpiresAtTick: entity.Passive.RobotShieldUntil,
		})
	}
	if entity.Passive.RobotWUntil > tick {
		buffs = append(buffs, world.BuffState{
			ID:            "robot_overdrive",
			Name:          "Overdrive",
			ExpiresAtTick: entity.Passive.RobotWUntil,
		})
	}
	if state := entity.Skills[eID]; state.Stacks > 0 && tick < state.StacksExpireTick {
		buffs = append(buffs, world.BuffState{
			ID:            "robot_power_fist",
			Name:          "Power Fist",
			ExpiresAtTick: state.StacksExpireTick,
		})
	}
	return buffs
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

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillCurve(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	if skill.MetaLists == nil {
		return fallback
	}
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	currentLevel := float64(clampInt(level, world.MinHeroLevel, world.MaxHeroLevel))
	if currentLevel <= levels[0] {
		return values[0]
	}
	last := len(values) - 1
	if currentLevel >= levels[last] {
		return values[last]
	}
	for i := 1; i < len(values); i++ {
		if currentLevel > levels[i] {
			continue
		}
		return values[i-1] + (values[i]-values[i-1])*(currentLevel-levels[i-1])/(levels[i]-levels[i-1])
	}
	return values[last]
}

func wMoveSpeed(skill config.SkillConfig, level int, elapsedTicks uint64, tickRate int) float64 {
	start := skillList(skill, "moveSpeedBonus", level, []float64{0.7, 0.75, 0.8, 0.85, 0.9})
	minimum := skillMeta(skill, "minMoveSpeedBonus", 0.1)
	decayTicks := secondsToTicks(skillMeta(skill, "decaySeconds", 2.5), tickRate)
	if decayTicks == 0 || elapsedTicks >= decayTicks {
		return minimum
	}
	return start - (start-minimum)*float64(elapsedTicks)/float64(decayTicks)
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
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

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
