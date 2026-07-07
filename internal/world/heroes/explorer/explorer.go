package explorer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"l-battle/internal/world/formula"
	"math"
	"strconv"
)

const (
	heroID    = "explorer"
	passiveID = "explorer_passive"
	qID       = "explorer_q"
	wID       = "explorer_w"
	eID       = "explorer_e"
	rID       = "explorer_r"
)

func init() {
	world.RegisterHeroHooks(heroID, world.HeroHooks{
		Cast: map[string]world.HeroCastHandler{
			qID: CastQ,
			wID: CastW,
			eID: CastE,
			rID: CastR,
		},
		Tick:            Tick,
		OnBasicHit:      OnBasicHit,
		OnSkillHit:      OnSkillHit,
		ActiveBuffs:     ActiveBuffs,
		ApplyStats:      ApplyStats,
		ExplorerQDamage: QDamage,
		ExplorerQHit:    QHit,
		ExplorerWAttach: WAttach,
		ExplorerEDamage: EDamage,
		ExplorerEHit:    EHit,
		ExplorerRDamage: RDamage,
		ExplorerRHit:    RHit,
	})
}

func CastQ(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.ExplorerQPending || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{28, 31, 34, 37, 40})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{5500, 5250, 5000, 4750, 4500})), tickRate)
	entity.Skills[qID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.ExplorerQPending = true
	entity.Passive.ExplorerQRelease = tick + windupTicks
	entity.Passive.ExplorerQTarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Passive.ExplorerQLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.ExplorerQRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseQ(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || !entity.Passive.ExplorerQPending || tick < entity.Passive.ExplorerQRelease {
		return
	}
	target := entity.Passive.ExplorerQTarget
	level := entity.Passive.ExplorerQLevel
	clearQ(entity)
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	entity.Combat.NextAttackTick = tick
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(qID)
	qRange := skillRange(skill, 1200)
	speed := skillMeta(skill, "projectileSpeed", 2000)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:explorer_q:"),
		Kind:         "explorer_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      qID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        qRange,
		Radius:       skillMeta(skill, "projectileRadius", 30),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(qRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func CastW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID || state.Level <= 0 || state.Stacks > 0 || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{50, 50, 50, 50, 50})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000})), tickRate)
	state.Stacks = 1
	state.StacksExpireTick = tick + secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	entity.Skills[wID] = state
	entity.Passive.ExplorerWTarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Passive.ExplorerWLevel = state.Level
	entity.Control.ActionLockedUntilTick = state.StacksExpireTick
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastE(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.ExplorerEPending || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{70, 70, 70, 70, 70})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{26000, 23000, 20000, 17000, 14000})), tickRate)
	entity.Skills[eID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 0.25), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.ExplorerEPending = true
	entity.Passive.ExplorerERelease = tick + windupTicks
	entity.Passive.ExplorerETarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Passive.ExplorerELevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.ExplorerERelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func CastR(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || state.Level <= 0 || entity.Passive.ExplorerRPending || tickRate <= 0 {
		return
	}
	cost := skillList(skill, "manaCost", state.Level, []float64{100, 100, 100})
	if entity.Stats.MP < cost {
		return
	}
	entity.Stats.MP -= cost
	state.CooldownUntilTick = tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{120000, 105000, 90000})), tickRate)
	entity.Skills[rID] = state
	windupTicks := secondsToTicks(skillMeta(skill, "castWindupSeconds", 1), tickRate)
	if windupTicks < 1 {
		windupTicks = 1
	}
	entity.Passive.ExplorerRPending = true
	entity.Passive.ExplorerRRelease = tick + windupTicks
	entity.Passive.ExplorerRTarget = world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	entity.Passive.ExplorerRLevel = state.Level
	entity.Control.ActionLockedUntilTick = entity.Passive.ExplorerRRelease
	w.LockAttackAfterCast(entity, tick, tickRate)
}

func releaseR(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || !entity.Passive.ExplorerRPending || tick < entity.Passive.ExplorerRRelease {
		return
	}
	target := entity.Passive.ExplorerRTarget
	level := entity.Passive.ExplorerRLevel
	clearR(entity)
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	skill := w.SkillConfig(rID)
	rRange := w.MapExitRange(entity.Position, world.Vector2{X: dx, Y: dy})
	speed := skillMeta(skill, "projectileSpeed", 2000)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:explorer_r:"),
		Kind:         "explorer_r",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      rID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        rRange,
		Radius:       skillMeta(skill, "projectileRadius", 160),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(rRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func releaseE(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if w == nil || entity == nil || entity.HeroID != heroID || !entity.Passive.ExplorerEPending {
		return
	}
	if explorerEInterrupted(entity, tick) {
		clearE(entity)
		return
	}
	if tick < entity.Passive.ExplorerERelease {
		return
	}
	targetPoint := entity.Passive.ExplorerETarget
	level := entity.Passive.ExplorerELevel
	clearE(entity)
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		return
	}
	skill := w.SkillConfig(eID)
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx != 0 || dy != 0 {
		eRange := skillRange(skill, 475)
		dist := math.Hypot(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
		if dist < eRange {
			eRange = dist
		}
		entity.Position = w.ClampWorldPoint(world.Vector2{X: entity.Position.X + dx*eRange, Y: entity.Position.Y + dy*eRange})
	}

	target := nearestETarget(w, entity, skillMeta(skill, "targetRange", 750), tick)
	if target == nil {
		return
	}
	speed := skillMeta(skill, "projectileSpeed", 2000)
	targetRange := skillMeta(skill, "targetRange", 750)
	tx, ty := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if tx == 0 && ty == 0 {
		tx = 1
	}
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:explorer_e:"),
		Kind:         "explorer_e",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      eID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: tx, Y: ty},
		SpeedPerTick: speed / float64(tickRate),
		Range:        targetRange,
		Radius:       skillMeta(skill, "projectileRadius", 60),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(targetRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func releaseW(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	state := entity.Skills[wID]
	if state.Stacks <= 0 || state.StacksExpireTick == 0 || tick < state.StacksExpireTick {
		return
	}
	state.Stacks = 0
	state.StacksExpireTick = 0
	entity.Skills[wID] = state
	if entity.Stats.HP <= 0 || entity.Death.Dead {
		entity.Passive.ExplorerWTarget = world.Vector2{}
		entity.Passive.ExplorerWLevel = 0
		return
	}
	skill := w.SkillConfig(wID)
	target := entity.Passive.ExplorerWTarget
	level := entity.Passive.ExplorerWLevel
	entity.Passive.ExplorerWTarget = world.Vector2{}
	entity.Passive.ExplorerWLevel = 0
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	wRange := skillRange(skill, 1200)
	speed := skillMeta(skill, "projectileSpeed", 1700)
	w.PutProjectile(&world.Projectile{
		ID:           w.NextProjectileID("projectile:explorer_w:"),
		Kind:         "explorer_w",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      wID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          world.Vector2{X: dx, Y: dy},
		SpeedPerTick: speed / float64(tickRate),
		Range:        wRange,
		Radius:       skillMeta(skill, "projectileRadius", 80),
		Damage:       level,
		CreatedAt:    tick,
		ExpiresAt:    tick + secondsToTicks(wRange/speed+0.2, tickRate),
		HitIDs:       make(map[string]bool),
	})
}

func OnBasicHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	detonateW(w, source, target, false, 0, tick, tickRate)
}

func OnSkillHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID || source.ID == target.ID || target.Stats.HP <= 0 {
		return
	}
	if source.Team == target.Team && target.Team != world.TeamNeutral {
		return
	}
	if tickRate <= 0 {
		tickRate = 20
	}
	skill := passiveSkill(w, source)
	maxStacks := int(skillMeta(skill, "maxStacks", 5))
	durationTicks := secondsToTicks(skillMeta(skill, "durationSeconds", 6), tickRate)
	if maxStacks <= 0 || durationTicks == 0 {
		return
	}
	cleanupExpired(source, tick)
	expiresAt := tick + durationTicks
	if len(source.Passive.ExplorerSpellForce) < maxStacks {
		source.Passive.ExplorerSpellForce = append(source.Passive.ExplorerSpellForce, expiresAt)
	} else {
		oldest := 0
		for i := 1; i < len(source.Passive.ExplorerSpellForce); i++ {
			if source.Passive.ExplorerSpellForce[i] < source.Passive.ExplorerSpellForce[oldest] {
				oldest = i
			}
		}
		source.Passive.ExplorerSpellForce[oldest] = expiresAt
	}
	if w != nil {
		w.RefreshPlayerStats(source)
	}
}

func Tick(w *world.World, entity *world.Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	releaseQ(w, entity, tick, tickRate)
	releaseE(w, entity, tick, tickRate)
	releaseR(w, entity, tick, tickRate)
	releaseW(w, entity, tick, tickRate)
	if cleanupExpired(entity, tick) && w != nil {
		w.RefreshPlayerStats(entity)
	}
}

func ApplyStats(w *world.World, entity *world.Entity, stats *world.Stats) {
	if entity == nil || entity.HeroID != heroID || stats == nil {
		return
	}
	stacks := len(entity.Passive.ExplorerSpellForce)
	if stacks == 0 {
		return
	}
	stats.AttackSpeedBonus += float64(stacks) * skillMeta(passiveSkill(w, entity), "attackSpeedPerStack", 0.1)
	stats.AttackSpeed = formula.FinalAttackSpeed(stats.BaseAttackSpeed, stats.AttackSpeedBonus, stats.AttackSpeedRatio, stats.AttackSpeedSlow)
}

func QDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	if w == nil || attacker == nil || target == nil {
		return 0
	}
	raw := skillList(skill, "baseDamage", skillLevel, []float64{20, 45, 70, 95, 120})
	raw += attacker.Stats.Attack * skillMeta(skill, "totalAdRatio", 1.3)
	raw += float64(attacker.Stats.AbilityPower) * skillMeta(skill, "apRatio", 0.4)
	damage := w.PhysicalDamageAfterResistance(attacker, target, raw, tick)
	damage += w.PhysicalDamageAfterResistance(attacker, target, w.EquipmentBasicAttackBonus(attacker, "physical"), tick)
	damage += w.MagicDamageAfterResistance(attacker, target, w.EquipmentBasicAttackBonus(attacker, "magic"), tick)
	if world.IsMinion(target) {
		damage += w.PhysicalDamageAfterResistance(attacker, target, w.EquipmentMinionBasicAttackBonus(attacker, "physical"), tick)
		damage += w.MagicDamageAfterResistance(attacker, target, w.EquipmentMinionBasicAttackBonus(attacker, "magic"), tick)
	}
	return damage
}

func QHit(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if source == nil || source.HeroID != heroID || tickRate <= 0 {
		return
	}
	refund := secondsToTicks(skillMeta(skill, "cooldownRefundSeconds", 1.5), tickRate)
	for skillID, state := range source.Skills {
		if state.CooldownUntilTick <= tick {
			continue
		}
		if state.CooldownUntilTick <= tick+refund {
			state.CooldownUntilTick = tick
		} else {
			state.CooldownUntilTick -= refund
		}
		source.Skills[skillID] = state
	}
	level := source.Skills[qID].Level
	detonateW(w, source, target, true, skillList(skill, "manaCost", level, []float64{28, 31, 34, 37, 40}), tick, tickRate)
}

func EDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	if w == nil || attacker == nil || target == nil {
		return 0
	}
	raw := skillList(skill, "baseDamage", skillLevel, []float64{80, 130, 180, 230, 280})
	raw += attacker.Stats.BonusAttack * skillMeta(skill, "bonusAdRatio", 0.6)
	raw += float64(attacker.Stats.AbilityPower) * skillMeta(skill, "apRatio", 0.75)
	return w.MagicDamageAfterResistance(attacker, target, raw, tick)
}

func EHit(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if source == nil || source.HeroID != heroID {
		return
	}
	level := source.Skills[eID].Level
	detonateW(w, source, target, true, skillList(skill, "manaCost", level, []float64{70, 70, 70, 70, 70}), tick, tickRate)
}

func RDamage(w *world.World, attacker *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	if w == nil || attacker == nil || target == nil {
		return 0
	}
	base := skillList(skill, "baseDamage", skillLevel, []float64{350, 550, 750})
	if world.IsMinion(target) || world.IsMonster(target) && target.Kind != world.EntityKindBaronNashor {
		base = skillList(skill, "reducedBaseDamage", skillLevel, []float64{150, 225, 300})
	}
	raw := base + attacker.Stats.BonusAttack*skillMeta(skill, "bonusAdRatio", 1) + float64(attacker.Stats.AbilityPower)*skillMeta(skill, "apRatio", 1.1)
	return w.MagicDamageAfterResistance(attacker, target, raw, tick)
}

func RHit(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if source == nil || source.HeroID != heroID {
		return
	}
	level := source.Skills[rID].Level
	detonateW(w, source, target, true, skillList(skill, "manaCost", level, []float64{100, 100, 100}), tick, tickRate)
}

func WAttach(w *world.World, source *world.Entity, target *world.Entity, skill config.SkillConfig, skillLevel int, tick uint64, tickRate int) {
	if source == nil || target == nil || source.HeroID != heroID || tickRate <= 0 {
		return
	}
	if target.Passive.ExplorerFluxMarks == nil {
		target.Passive.ExplorerFluxMarks = map[string]world.ExplorerFluxState{}
	}
	target.Passive.ExplorerFluxMarks[source.ID] = world.ExplorerFluxState{
		Level:     skillLevel,
		ExpiresAt: tick + secondsToTicks(skillMeta(skill, "durationSeconds", 4), tickRate),
	}
}

func detonateW(w *world.World, source *world.Entity, target *world.Entity, skillDetonation bool, manaCost float64, tick uint64, tickRate int) {
	if w == nil || source == nil || target == nil || source.HeroID != heroID || target.Passive.ExplorerFluxMarks == nil {
		return
	}
	mark := target.Passive.ExplorerFluxMarks[source.ID]
	if mark.Level <= 0 || tick >= mark.ExpiresAt || target.Stats.HP <= 0 {
		delete(target.Passive.ExplorerFluxMarks, source.ID)
		return
	}
	delete(target.Passive.ExplorerFluxMarks, source.ID)
	skill := w.SkillConfig(wID)
	raw := skillList(skill, "baseDamage", mark.Level, []float64{80, 135, 190, 245, 300})
	raw += source.Stats.BonusAttack * skillMeta(skill, "bonusAdRatio", 1)
	raw += float64(source.Stats.AbilityPower) * skillMeta(skill, "apRatio", 0.9)
	target.Combat.LastHitTick = tick
	damage := w.MagicDamageAfterResistance(source, target, raw, tick)
	if target.Kind == world.EntityKindDummy {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	} else {
		wasAlive := target.Stats.HP > 0
		w.ApplyMagicDamage(source, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.ApplyKillReward(source, target)
			w.KillPlayer(target, tick, tickRate)
			w.RemoveDeadUnit(target)
		}
	}
	if skillDetonation {
		source.Stats.MP += manaCost + skillMeta(skill, "manaRefundBonus", 60)
		if source.Stats.MP > source.Stats.MaxMP {
			source.Stats.MP = source.Stats.MaxMP
		}
	}
}

func ActiveBuffs(w *world.World, entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil {
		return nil
	}
	buffs := activeExplorerFluxBuffs(entity, tick)
	if entity.HeroID != heroID {
		return buffs
	}
	stacks := 0
	expiresAt := uint64(0)
	for _, expiry := range entity.Passive.ExplorerSpellForce {
		if tick >= expiry {
			continue
		}
		stacks++
		if expiresAt == 0 || expiry < expiresAt {
			expiresAt = expiry
		}
	}
	if stacks == 0 {
		return buffs
	}
	buffs = append(buffs, world.BuffState{
		ID:            "explorer_rising_spell_force",
		Name:          "Rising Spell Force " + strconv.Itoa(stacks),
		ExpiresAtTick: expiresAt,
	})
	return buffs
}

func activeExplorerFluxBuffs(entity *world.Entity, tick uint64) []world.BuffState {
	if entity == nil || len(entity.Passive.ExplorerFluxMarks) == 0 {
		return nil
	}
	buffs := make([]world.BuffState, 0, len(entity.Passive.ExplorerFluxMarks))
	for sourceID, mark := range entity.Passive.ExplorerFluxMarks {
		if tick >= mark.ExpiresAt {
			continue
		}
		buffs = append(buffs, world.BuffState{
			ID:            "explorer_essence_flux:" + sourceID,
			Name:          "Essence Flux",
			ExpiresAtTick: mark.ExpiresAt,
			Negative:      true,
		})
	}
	return buffs
}

func nearestETarget(w *world.World, source *world.Entity, radius float64, tick uint64) *world.Entity {
	var best *world.Entity
	bestDist := math.MaxFloat64
	bestMarked := false
	w.ForEachEntity(func(target *world.Entity) {
		if !world.CanAttackTarget(source, target) {
			return
		}
		dist := math.Hypot(target.Position.X-source.Position.X, target.Position.Y-source.Position.Y)
		if dist > radius+target.Radius {
			return
		}
		marked := target.Passive.ExplorerFluxMarks != nil && tick < target.Passive.ExplorerFluxMarks[source.ID].ExpiresAt
		if best != nil && (bestMarked && !marked || bestMarked == marked && dist >= bestDist) {
			return
		}
		best = target
		bestDist = dist
		bestMarked = marked
	})
	return best
}

func explorerEInterrupted(entity *world.Entity, tick uint64) bool {
	return tick < entity.Control.StunnedUntilTick || tick < entity.Control.AirborneUntilTick || tick < entity.Control.RootedUntilTick
}

func clearQ(entity *world.Entity) {
	entity.Passive.ExplorerQPending = false
	entity.Passive.ExplorerQRelease = 0
	entity.Passive.ExplorerQTarget = world.Vector2{}
	entity.Passive.ExplorerQLevel = 0
}

func clearE(entity *world.Entity) {
	entity.Passive.ExplorerEPending = false
	entity.Passive.ExplorerERelease = 0
	entity.Passive.ExplorerETarget = world.Vector2{}
	entity.Passive.ExplorerELevel = 0
}

func clearR(entity *world.Entity) {
	entity.Passive.ExplorerRPending = false
	entity.Passive.ExplorerRRelease = 0
	entity.Passive.ExplorerRTarget = world.Vector2{}
	entity.Passive.ExplorerRLevel = 0
}

func cleanupExpired(entity *world.Entity, tick uint64) bool {
	if entity == nil || len(entity.Passive.ExplorerSpellForce) == 0 {
		return false
	}
	stacks := entity.Passive.ExplorerSpellForce
	active := stacks[:0]
	for _, expiry := range stacks {
		if tick < expiry {
			active = append(active, expiry)
		}
	}
	entity.Passive.ExplorerSpellForce = active
	return len(active) != len(stacks)
}

func passiveSkill(w *world.World, entity *world.Entity) config.SkillConfig {
	if w == nil || entity == nil {
		return config.SkillConfig{}
	}
	return w.HeroPassiveSkill(entity)
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

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
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

func cooldownTicksFor(entity *world.Entity, cooldownMS int, tickRate int) uint64 {
	seconds := float64(cooldownMS) / 1000
	if entity != nil && entity.Stats.AbilityHaste > 0 {
		seconds /= 1 + entity.Stats.AbilityHaste/100
	}
	return secondsToTicks(seconds, tickRate)
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 || tickRate <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func normalize(dx float64, dy float64) (float64, float64) {
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}
