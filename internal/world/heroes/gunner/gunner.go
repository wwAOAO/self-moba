package gunner

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

const heroID = "gunner"

const (
	qID = "gunner_q"
	wID = "gunner_w"
	eID = "gunner_e"
	rID = "gunner_r"
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
		TickEntity:                     Tick,
		OnBasicHit:                     OnBasicHit,
		OnSkillUpgrade:                 RefreshW,
		ActiveBuffs:                    ActiveBuffs,
		ApplyStats:                     ApplyStats,
		BasicAttackBonusPhysicalDamage: PassiveDamage,
	})
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || tickRate <= 0 {
		return
	}
	target := qTarget(w, entity, cast, skill)
	if target == nil {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{40, 40, 40, 40, 40})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	fireQ(w, entity, target, state.Level, skill, tick, tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skill, state.Level, tickRate)
	entity.Skills[qID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func fireQ(w *world.World, entity *world.Entity, target *world.Entity, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	dir := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dir.X == 0 && dir.Y == 0 {
		dir.X = 1
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:gunner_q:"),
		Kind:         "gunner_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          dir,
		SpeedPerTick: skillMeta(skill, "projectileSpeed", 1800) / float64(tickRate),
		Range:        skillRange(skill, 650) + target.Radius,
		Radius:       skillMeta(skill, "projectileRadius", 16),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func qTarget(w *world.World, entity *world.Entity, cast protocol.CastInput, skill config.SkillConfig) *world.Entity {
	if target := w.EntityByID(cast.TargetID); world.CanAttackTarget(entity, target) && distance(entity.Position, target.Position) <= skillRange(skill, 650)+target.Radius {
		return target
	}
	point := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	var best *world.Entity
	bestDistance := math.MaxFloat64
	w.ForEachEntity(func(target *world.Entity) {
		if !world.CanAttackTarget(entity, target) || distance(entity.Position, target.Position) > skillRange(skill, 650)+target.Radius {
			return
		}
		d := distance(point, target.Position)
		if d <= target.Radius+skillMeta(skill, "targetPickPadding", 90) && d < bestDistance {
			bestDistance = d
			best = target
		}
	})
	return best
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{45, 45, 45, 45, 45})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	entity.Passive.GunnerWActiveUntil = tick + secondsToTicks(skillMeta(skill, "activeDurationSeconds", 4), tickRate)
	entity.Passive.GunnerWAttackSpeed = skillList(skill, "attackSpeedBonus", state.Level, []float64{0.4, 0.55, 0.7, 0.85, 1})
	entity.Passive.GunnerWMoveSpeed = skillList(skill, "enhancedMoveSpeed", state.Level, []float64{60, 70, 80, 90, 100})
	state.CooldownUntilTick = tick + cooldownTicks(skill, state.Level, tickRate)
	entity.Skills[wID] = state
	w.RefreshPlayerStats(entity)
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{80, 80, 80, 80, 80})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	center := w.ClampWorldPoint(world.Vector2{X: cast.TargetX, Y: cast.TargetY})
	if distance(entity.Position, center) > skillRange(skill, 1000) {
		dir := normalize(center.X-entity.Position.X, center.Y-entity.Position.Y)
		center = w.ClampWorldPoint(world.Vector2{
			X: entity.Position.X + dir.X*skillRange(skill, 1000),
			Y: entity.Position.Y + dir.Y*skillRange(skill, 1000),
		})
	}
	entity.Passive.GunnerECenter = center
	entity.Passive.GunnerEExpireTick = tick + secondsToTicks(skillMeta(skill, "durationSeconds", 2), tickRate)
	entity.Passive.GunnerENextTick = tick
	entity.Passive.GunnerELevel = state.Level
	entity.Passive.GunnerEEffectID = addEEffect(w, entity, center, skillMeta(skill, "radius", 300), tick, entity.Passive.GunnerEExpireTick)
	state.CooldownUntilTick = tick + cooldownTicks(skill, state.Level, tickRate)
	entity.Skills[eID] = state
	w.LockAttackAfterCast(entity, tick, tickRate)
	tickE(w, entity, tick, tickRate)
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	before := entity.Passive.GunnerWMoveSpeed
	if tick >= entity.Passive.GunnerWActiveUntil {
		entity.Passive.GunnerWActiveUntil = 0
		entity.Passive.GunnerWAttackSpeed = 0
		entity.Passive.GunnerWMoveSpeed = passiveWMoveSpeed(entity, w.SkillConfig(wID), tick, tickRate)
	}
	if entity.Passive.GunnerWMoveSpeed != before {
		w.RefreshPlayerStats(entity)
	}
	tickE(w, entity, tick, tickRate)
	tickR(w, entity, tick, tickRate)
}

func RefreshW(w *world.World, entity *world.Entity, skillID string) {
	if entity == nil || entity.HeroID != heroID || skillID != wID {
		return
	}
	entity.Passive.GunnerWMoveSpeed = passiveWMoveSpeed(entity, w.SkillConfig(wID), 0, 20)
	w.RefreshPlayerStats(entity)
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || entity.HeroID != heroID || stats == nil {
		return
	}
	stats.MoveSpeed += entity.Passive.GunnerWMoveSpeed
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || entity.HeroID != heroID || entity.Passive.GunnerWMoveSpeed <= 0 {
		return nil
	}
	if entity.Passive.GunnerWActiveUntil > tick {
		return []world.BuffState{{
			ID:            "gunner_w_active",
			Name:          "W Active",
			ExpiresAtTick: entity.Passive.GunnerWActiveUntil,
		}}
	}
	return []world.BuffState{{
		ID:   "gunner_w_passive",
		Name: "W Passive",
	}}
}

func passiveWMoveSpeed(entity *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) float64 {
	state := entity.Skills[wID]
	if state.Level <= 0 || tickRate <= 0 || tick < entity.Combat.LastHitTick+secondsToTicks(4, tickRate) {
		return 0
	}
	if tick >= entity.Combat.LastHitTick+secondsToTicks(7, tickRate) {
		return skillList(skill, "enhancedMoveSpeed", state.Level, []float64{60, 70, 80, 90, 100})
	}
	return skillList(skill, "moveSpeed", state.Level, []float64{30, 35, 40, 45, 50})
}

func tickE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.GunnerEExpireTick == 0 {
		return
	}
	skill := w.SkillConfig(eID)
	if tick >= entity.Passive.GunnerEExpireTick {
		w.RemoveSkillEffect(entity.Passive.GunnerEEffectID)
		entity.Passive.GunnerEExpireTick = 0
		entity.Passive.GunnerENextTick = 0
		entity.Passive.GunnerELevel = 0
		entity.Passive.GunnerEEffectID = ""
		return
	}
	slow := skillMeta(skill, "slow", 0.4) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "slowAPRatio", 0.0006)
	for _, target := range w.TargetsInRadius(entity, entity.Passive.GunnerECenter, skillMeta(skill, "radius", 300)) {
		w.ApplyMoveSpeedSlow(target, slow, tick+2)
	}
	if tick < entity.Passive.GunnerENextTick {
		return
	}
	dealEDamage(w, entity, skill, entity.Passive.GunnerELevel, tick, tickRate)
	entity.Passive.GunnerENextTick = tick + secondsToTicks(skillMeta(skill, "tickSeconds", 0.25), tickRate)
}

func dealEDamage(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	ticksPerSecond := 1 / skillMeta(skill, "tickSeconds", 0.25)
	rawDamage := (skillList(skill, "damagePerSecond", level, []float64{35, 50, 65, 80, 95}) + float64(entity.Stats.AbilityPower)*skillMeta(skill, "apRatio", 0.6)) / ticksPerSecond
	for _, target := range w.TargetsInRadius(entity, entity.Passive.GunnerECenter, skillMeta(skill, "radius", 300)) {
		target.Combat.LastHitTick = tick
		target.Combat.DamageEvents = nil
		damage := w.MagicDamageAfterResistance(entity, target, rawDamage, tick)
		if target.Kind == world.EntityKindDummy {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "magic"
			continue
		}
		wasAlive := target.Stats.HP > 0
		w.ApplyAOEDamage(entity, target, damage, "magic", tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(entity, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
}

func addEEffect(w *world.World, entity *world.Entity, center world.Vector2, radius float64, tick uint64, expiresAt uint64) string {
	id := w.NextEffectID("effect:gunner_e:")
	w.PutSkillEffect(world.SkillEffect{
		ID:           id,
		Kind:         "gunner_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        center,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
	return id
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 || tickRate <= 0 || entity.Passive.GunnerRExpireTick > tick {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{100, 100, 100})
	if entity.Stats.MP < cost {
		return
	}
	dir := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dir.X == 0 && dir.Y == 0 {
		dir.X = 1
	}
	entity.Stats.MP -= cost
	durationTicks := secondsToTicks(skillMeta(skill, "durationSeconds", 3), tickRate)
	entity.Passive.GunnerRDir = dir
	entity.Passive.GunnerRStartTick = tick
	entity.Passive.GunnerRExpireTick = tick + durationTicks
	entity.Passive.GunnerRNextTick = rWaveTick(tick, 0, int(skillList(skill, "waves", state.Level, []float64{14, 16, 18})), skill, tickRate)
	entity.Passive.GunnerRLevel = state.Level
	entity.Passive.GunnerRWaves = int(skillList(skill, "waves", state.Level, []float64{14, 16, 18}))
	entity.Passive.GunnerRWaveCount = 0
	entity.Passive.GunnerREffectID = addREffect(w, entity, dir, skill, tick, entity.Passive.GunnerRExpireTick)
	lastWaveTick := rWaveTick(tick, entity.Passive.GunnerRWaves-1, entity.Passive.GunnerRWaves, skill, tickRate)
	entity.Control.ActionLockedUntilTick = min(lastWaveTick+1, entity.Passive.GunnerRExpireTick)
	state.CooldownUntilTick = tick + cooldownTicks(skill, state.Level, tickRate)
	entity.Skills[rID] = state
	tickR(w, entity, tick, tickRate)
}

func tickR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || entity.Passive.GunnerRExpireTick == 0 {
		return
	}
	if tick >= entity.Passive.GunnerRExpireTick || entity.Passive.GunnerRWaveCount >= entity.Passive.GunnerRWaves {
		w.RemoveSkillEffect(entity.Passive.GunnerREffectID)
		entity.Passive.GunnerRExpireTick = 0
		entity.Passive.GunnerRStartTick = 0
		entity.Passive.GunnerRNextTick = 0
		entity.Passive.GunnerRLevel = 0
		entity.Passive.GunnerRWaves = 0
		entity.Passive.GunnerRWaveCount = 0
		entity.Passive.GunnerREffectID = ""
		return
	}
	if tick < entity.Passive.GunnerRNextTick {
		return
	}
	skill := w.SkillConfig(rID)
	fireRWave(w, entity, skill, entity.Passive.GunnerRLevel, tick, tickRate)
	entity.Passive.GunnerRWaveCount++
	entity.Passive.GunnerRNextTick = rWaveTick(entity.Passive.GunnerRStartTick, entity.Passive.GunnerRWaveCount, entity.Passive.GunnerRWaves, skill, tickRate)
}

func fireRWave(w *world.World, entity *world.Entity, skill config.SkillConfig, level int, tick uint64, tickRate int) {
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:gunner_r:"),
		Kind:         "gunner_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      rID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          entity.Passive.GunnerRDir,
		SpeedPerTick: skillMeta(skill, "projectileSpeed", 2800) / float64(tickRate),
		Range:        skillRange(skill, 1400),
		Radius:       skillMeta(skill, "projectileRadius", 18),
		Damage:       level,
		EffectRatio:  skillMeta(skill, "coneAngleDegrees", 45),
		DisplayRange: skillMeta(skill, "coneAngleDegrees", 45) * 0.75,
		DisplayCount: 7,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(1, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func rWaveTick(startTick uint64, waveIndex int, waves int, skill config.SkillConfig, tickRate int) uint64 {
	if waveIndex >= waves {
		return startTick + secondsToTicks(skillMeta(skill, "durationSeconds", 3), tickRate)
	}
	first := skillMeta(skill, "firstWaveSeconds", 0.066)
	last := skillMeta(skill, "lastWaveSeconds", 2.904)
	offset := first
	if waves > 1 {
		offset += (last - first) * float64(waveIndex) / float64(waves-1)
	}
	return startTick + secondsToTicks(offset, tickRate)
}

func addREffect(w *world.World, entity *world.Entity, dir world.Vector2, skill config.SkillConfig, tick uint64, expiresAt uint64) string {
	id := w.NextEffectID("effect:gunner_r:")
	w.PutSkillEffect(world.SkillEffect{
		ID:           id,
		Kind:         "gunner_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SourceHeroID: entity.HeroID,
		Start:        entity.Position,
		Dir:          dir,
		Range:        skillRange(skill, 1400),
		Width:        skillMeta(skill, "coneAngleDegrees", 45),
		CreatedAt:    tick,
		ExpiresAt:    expiresAt,
	})
	return id
}

func PassiveDamage(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) int {
	if attacker == nil || target == nil || attacker.HeroID != heroID || attacker.Passive.GunnerTargetID == target.ID {
		return 0
	}
	ratio := levelRatio(attacker.Level, 0.5, 1)
	if world.IsMinion(target) {
		ratio = levelRatio(attacker.Level, 0.25, 0.5)
	}
	return int(attacker.Stats.Attack * ratio)
}

func OnBasicHit(w *world.World, attacker *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if attacker == nil || target == nil || attacker.HeroID != heroID || attacker.Passive.GunnerTargetID == target.ID {
		return
	}
	attacker.Passive.GunnerTargetID = target.ID
	state := attacker.Skills[wID]
	reduction := uint64(tickRate * 2)
	if reduction == 0 || state.CooldownUntilTick <= tick {
		return
	}
	if state.CooldownUntilTick <= tick+reduction {
		state.CooldownUntilTick = tick
	} else {
		state.CooldownUntilTick -= reduction
	}
	attacker.Skills[wID] = state
}

func levelRatio(level int, min float64, max float64) float64 {
	if level < world.MinHeroLevel {
		level = world.MinHeroLevel
	}
	if level > world.MaxHeroLevel {
		level = world.MaxHeroLevel
	}
	return min + (max-min)*float64(level-world.MinHeroLevel)/float64(world.MaxHeroLevel-world.MinHeroLevel)
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
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

func cooldownTicks(skill config.SkillConfig, level int, tickRate int) uint64 {
	return uint64(math.Ceil(skillList(skill, "cooldownMs", level, []float64{7000, 6000, 5000, 4000, 3000}) / 1000 * float64(tickRate)))
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func distance(a world.Vector2, b world.Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func normalize(dx float64, dy float64) world.Vector2 {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return world.Vector2{}
	}
	return world.Vector2{X: dx / length, Y: dy / length}
}
