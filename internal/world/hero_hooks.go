package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

type HeroHooks struct {
	Cast map[string]HeroCastHandler

	ApplyW                func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int)
	ReleaseQ              func(w *World, entity *Entity, tick uint64, tickRate int)
	ReleaseW              func(w *World, entity *Entity, tick uint64, tickRate int)
	ReleaseE              func(w *World, entity *Entity, tick uint64, tickRate int)
	ReleaseR              func(w *World, entity *Entity, tick uint64, tickRate int)
	RefreshSkillOnUpgrade func(w *World, entity *Entity, skillID string)
	TickHawkCharges       func(w *World, entity *Entity, tick uint64, tickRate int)
	AddFocusStack         func(w *World, entity *Entity, tick uint64, tickRate int)
	ExpireFocus           func(w *World, entity *Entity, tick uint64)
	FocusBonusDamage      func(w *World, attacker *Entity, target *Entity, tick uint64) int
	ApplyFocusOnBasicHit  func(w *World, attacker *Entity, target *Entity, tick uint64, tickRate int)
	ApplyFrostShot        func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)
	WDamage               func(w *World, entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64) int
	ArcherRDamage         func(w *World, entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64, multiplier float64) int
	RStunTicks            func(projectile *Projectile, skill config.SkillConfig, tickRate int) uint64
	ApplyRSplash          func(w *World, source *Entity, primary *Entity, projectile *Projectile, skill config.SkillConfig, tick uint64, tickRate int)

	GainBasicAttackRage func(w *World, attacker *Entity, target *Entity, tick uint64)
	GainKillRage        func(w *World, killer *Entity)
	GainSkillHitRage    func(w *World, source *Entity, tick uint64)
	TickRageDecay       func(w *World, entity *Entity, tick uint64, tickRate int)
	RageCritChance      func(entity *Entity, skill config.SkillConfig) float64

	ActivateEZone         func(w *World, entity *Entity, center Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int)
	TickE                 func(w *World, entity *Entity, tick uint64, tickRate int)
	DetonateE             func(w *World, entity *Entity, skill config.SkillConfig, tick uint64, tickRate int)
	ApplyFinalSparkRefund func(w *World, target *Entity)
	MageQDamage           func(w *World, attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int
	ApplyIllumination     func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)
	ApplyUltimateIllum    func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)
	TriggerIllumination   func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)
	DetonateIllumination  func(w *World, source *Entity, target *Entity, tick uint64, tickRate int) bool

	ExpireQStacks            func(entity *Entity, tick uint64)
	CritChanceMultiplier     func(w *World, entity *Entity) float64
	ApplyCritFinalMultiplier func(w *World, attacker *Entity, damage int, crit bool) int
	ApplyShield              func(w *World, source *Entity, target *Entity, tickRate int)
	ShieldValue              func(w *World, entity *Entity) int
	PassiveState             func(w *World, hero config.HeroConfig) PassiveState
	StateForHero             func(heroID string) SwordState
	ChargeIntent             func(w *World, entity *Entity, moved float64)
	TickShield               func(entity *Entity, tick uint64)
	ApplyCritOverflowStats   func(w *World, entity *Entity, stats *Stats)

	StartRDash        func(w *World, entity *Entity, targetPoint Vector2, state SkillState, skill config.SkillConfig, tick uint64, tickRate int)
	ReleasePreparedR  func(w *World, entity *Entity, tick uint64, tickRate int)
	CancelPreparedR   func(entity *Entity)
	ResolveRImpact    func(w *World, entity *Entity, tick uint64, tickRate int)
	RefreshWPassive   func(w *World, entity *Entity)
	TickGranite       func(w *World, entity *Entity, tick uint64, tickRate int)
	RefreshGranite    func(w *World, entity *Entity)
	RefreshGraniteMax func(w *World, entity *Entity)
	WBonusDamage      func(w *World, attacker *Entity, tick uint64) float64
	ApplyWAftershock  func(w *World, attacker *Entity, primary *Entity, tick uint64, tickRate int)
	TankQDamage       func(w *World, attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int

	StopE               func(w *World, entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int)
	QBonusDamage        func(w *World, attacker *Entity, tick uint64) float64
	ConsumeQ            func(w *World, attacker *Entity, target *Entity, tick uint64, tickRate int)
	ApplyWPassiveKill   func(w *World, killer *Entity, target *Entity)
	TickToughness       func(w *World, entity *Entity, tick uint64, tickRate int)
	ToughnessRegenRatio func(level int, skill config.SkillConfig) float64
	WarriorRDamage      func(target *Entity, skill config.SkillConfig, level int) float64
}

var heroHooks = map[string]HeroHooks{}

func RegisterHeroHooks(heroID string, hooks HeroHooks) {
	heroHooks[heroID] = hooks
	if len(hooks.Cast) > 0 {
		RegisterHeroCastHandlers(heroID, hooks.Cast)
	}
}

func heroHooksFor(heroID string) HeroHooks {
	return heroHooks[heroID]
}

func (w *World) refreshArcherSkillOnUpgrade(entity *Entity, skillID string) {
	if h := heroHooksFor(archerHeroID).RefreshSkillOnUpgrade; h != nil {
		h(w, entity, skillID)
	}
}

func (w *World) tickArcherHawkCharges(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(archerHeroID).TickHawkCharges; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) addArcherFocusStack(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(archerHeroID).AddFocusStack; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) expireArcherFocus(entity *Entity, tick uint64) {
	if h := heroHooksFor(archerHeroID).ExpireFocus; h != nil {
		h(w, entity, tick)
	}
}

func (w *World) archerFocusBonusDamage(attacker *Entity, target *Entity, tick uint64) int {
	if h := heroHooksFor(archerHeroID).FocusBonusDamage; h != nil {
		return h(w, attacker, target, tick)
	}
	return 0
}

func (w *World) applyArcherFocusOnBasicHit(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(archerHeroID).ApplyFocusOnBasicHit; h != nil {
		h(w, attacker, target, tick, tickRate)
	}
}

func (w *World) releaseArcherCrystalArrow(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(archerHeroID).ReleaseR; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func archerRDamage(entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64, multiplier float64) int {
	if h := heroHooksFor(archerHeroID).ArcherRDamage; h != nil {
		return h(nil, entity, target, skill, level, tick, multiplier)
	}
	return 0
}

func archerRStunTicks(projectile *Projectile, skill config.SkillConfig, tickRate int) uint64 {
	if h := heroHooksFor(archerHeroID).RStunTicks; h != nil {
		return h(projectile, skill, tickRate)
	}
	return 0
}

func applyArcherRSplash(w *World, source *Entity, primary *Entity, projectile *Projectile, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(archerHeroID).ApplyRSplash; h != nil {
		h(w, source, primary, projectile, skill, tick, tickRate)
	}
}

func (w *World) ArcherMagicDamageAfterResistance(source *Entity, target *Entity, raw float64, tick uint64) int {
	return magicDamageAfterResistance(source, target, raw, tick)
}

func applyArcherW(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(archerHeroID).ApplyW; h != nil {
		h(w, entity, cast, state, skill, tick, tickRate)
	}
}

func archerWDamage(entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64) int {
	if h := heroHooksFor(archerHeroID).WDamage; h != nil {
		return h(nil, entity, target, skill, level, tick)
	}
	return 0
}

func (w *World) ArcherPhysicalDamageAfterResistance(source *Entity, target *Entity, raw float64, tick uint64) int {
	return physicalDamageAfterResistance(source, target, raw, tick)
}

func (w *World) HeroPassiveSkill(entity *Entity) config.SkillConfig {
	return w.heroPassiveSkill(entity)
}

func (w *World) AttackCrits(attacker *Entity, target *Entity, tick uint64) bool {
	return w.attackCrits(attacker, target, tick)
}

func (w *World) gainBladeBasicAttackRage(attacker *Entity, target *Entity, tick uint64) {
	if h := heroHooksFor(bladeHeroID).GainBasicAttackRage; h != nil {
		h(w, attacker, target, tick)
	}
}

func (w *World) gainBladeKillRage(killer *Entity) {
	if h := heroHooksFor(bladeHeroID).GainKillRage; h != nil {
		h(w, killer)
	}
}

func (w *World) gainBladeSkillHitRage(source *Entity, tick uint64) {
	if h := heroHooksFor(bladeHeroID).GainSkillHitRage; h != nil {
		h(w, source, tick)
	}
}

func (w *World) tickBladeRageDecay(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(bladeHeroID).TickRageDecay; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func bladeRageCritChance(entity *Entity, skill config.SkillConfig) float64 {
	if h := heroHooksFor(bladeHeroID).RageCritChance; h != nil {
		return h(entity, skill)
	}
	return 0
}

func (w *World) releaseMageQ(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ReleaseQ; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) mageQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int {
	if h := heroHooksFor(mageHeroID).MageQDamage; h != nil {
		return h(w, attacker, target, skill, skillLevel, multiplier, tick)
	}
	return 0
}

func (w *World) MageMagicDamageAfterResistance(source *Entity, target *Entity, raw float64, tick uint64) int {
	return magicDamageAfterResistance(source, target, raw, tick)
}

func (w *World) releaseMageW(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ReleaseW; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func mageWShieldValue(entity *Entity, skill config.SkillConfig, level int) int {
	baseShield := skillMetaListByLevel(skill, "shieldValue", level, []float64{50, 65, 80, 95, 110})
	return int(math.Round(baseShield + float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.2)))
}

func (w *World) MageWShieldValue(entity *Entity, skill config.SkillConfig, level int) int {
	return mageWShieldValue(entity, skill, level)
}

func (w *World) AddMageShieldLayer(target *Entity, amount int, expiresAt uint64) {
	w.addMageShieldLayer(target, amount, expiresAt)
}

func (w *World) addMageShieldLayer(target *Entity, amount int, expiresAt uint64) {
	if target == nil || amount <= 0 || expiresAt == 0 {
		return
	}
	target.Passive.ShieldLayers = append(target.Passive.ShieldLayers, ShieldLayer{Amount: amount, ExpiresAt: expiresAt})
	target.Passive.Shield += amount
	target.Passive.MaxShield += amount
}

func (w *World) releaseMageE(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ReleaseE; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) activateMageEZone(entity *Entity, center Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ActivateEZone; h != nil {
		h(w, entity, center, level, skill, tick, tickRate)
	}
}

func (w *World) tickMageE(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).TickE; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) detonateMageE(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).DetonateE; h != nil {
		h(w, entity, skill, tick, tickRate)
	}
}

func (w *World) addMageEEffect(entity *Entity, center Vector2, radius float64, tick uint64, expiresAt uint64) string {
	id := w.NextEffectID("effect:mage_e:")
	w.PutSkillEffect(SkillEffect{
		ID:        id,
		Kind:      "mage_lucent_singularity",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: expiresAt,
	})
	return id
}

func (w *World) MageTargetsInRadius(entity *Entity, center Vector2, radius float64) []*Entity {
	return w.targetsInRadius(entity, center, radius)
}

func (w *World) ApplyMageMoveSpeedSlow(target *Entity, slow float64, until uint64) {
	applyMoveSpeedSlow(target, slow, until)
}

func (w *World) ApplyMageIlluminationOnSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	w.applyMageIlluminationOnSkillHit(source, target, tick, tickRate)
}

func (w *World) releaseMageR(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ReleaseR; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) mageRTargets(entity *Entity, direction Vector2, skill config.SkillConfig, tick uint64) []*Entity {
	hits := make([]*Entity, 0)
	castRange := skillRange(skill, 3400)
	halfWidth := skillMetaRange(skill, "beamWidth", 200) / 2
	w.ForEachEntity(func(target *Entity) {
		if !canAttackTarget(entity, target) {
			return
		}
		along, perpendicular := projectPoint(entity.Position, direction, target.Position)
		if along < -target.Radius || along > castRange+target.Radius {
			return
		}
		if perpendicular <= halfWidth+target.Radius {
			hits = append(hits, target)
		}
	})
	return hits
}

func (w *World) addMageREffect(entity *Entity, direction Vector2, beamRange float64, beamWidth float64, tick uint64, tickRate int) {
	id := w.NextEffectID("effect:mage_r:")
	lifeTicks := secondsToTicks(0.25, tickRate)
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.PutSkillEffect(SkillEffect{
		ID:        id,
		Kind:      "mage_final_spark",
		Team:      entity.Team,
		Start:     entity.Position,
		End:       w.ClampWorldPoint(Vector2{X: entity.Position.X + direction.X*beamRange, Y: entity.Position.Y + direction.Y*beamRange}),
		Dir:       direction,
		Range:     beamRange,
		Width:     beamWidth,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	})
}

func (w *World) ApplyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	w.applyMageIlluminationOnUltimateHit(source, target, tick, tickRate)
}

func applyMageFinalSparkRefund(w *World, target *Entity) {
	if h := heroHooksFor(mageHeroID).ApplyFinalSparkRefund; h != nil {
		h(w, target)
	}
}

func (w *World) applyMageIlluminationOnSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ApplyIllumination; h != nil {
		h(w, source, target, tick, tickRate)
	}
}

func (w *World) applyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ApplyUltimateIllum; h != nil {
		h(w, source, target, tick, tickRate)
	}
}

func (w *World) triggerMageIlluminationOnBasicAttack(source *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).TriggerIllumination; h != nil {
		h(w, source, target, tick, tickRate)
	}
}

func (w *World) detonateMageIllumination(source *Entity, target *Entity, tick uint64, tickRate int) bool {
	if h := heroHooksFor(mageHeroID).DetonateIllumination; h != nil {
		return h(w, source, target, tick, tickRate)
	}
	return false
}

func (w *World) MagePassiveSkill(entity *Entity) config.SkillConfig {
	return w.heroPassiveSkill(entity)
}

func MageSkillMetaCurveByLevel(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	return skillMetaCurveByLevel(skill, valueKey, levelKey, level, fallback)
}

func (w *World) releaseSwordQ(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(swordHeroID).ReleaseQ; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) expireSwordQStacks(entity *Entity, tick uint64) {
	if h := heroHooksFor(swordHeroID).ExpireQStacks; h != nil {
		h(entity, tick)
	}
}

func (w *World) SwordQTargets(entity *Entity, targetPoint Vector2, qRange float64, form string, skill config.SkillConfig) []*Entity {
	return w.swordQTargets(entity, targetPoint, qRange, form, skill)
}

func (w *World) SwordQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, tick uint64) int {
	return w.swordQDamage(attacker, target, skill, tick)
}

func (w *World) SwordQCooldownTicks(entity *Entity, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	return w.swordQCooldownTicks(entity, skill, skillLevel, tickRate)
}

func swordQWindupSeconds(entity *Entity, skill config.SkillConfig) float64 {
	base := skillMetaRange(skill, "castWindupSeconds", 0.328)
	minimum := skillMetaRange(skill, "minCastWindupSeconds", 0.09)
	bonus := 0.0
	if entity != nil {
		bonus = entity.Stats.AttackSpeedBonus
	}
	bonus = clamp(bonus, 0, math.MaxFloat64)
	seconds := base / (1 + bonus)
	if seconds < minimum {
		return minimum
	}
	return seconds
}

func (w *World) SwordEDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{60, 70, 80, 90, 100})
	damageValue := baseDamage + attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 0.2) + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	damageValue *= 1 + float64(attacker.Sword.SweepingBladeStacks)*skillMetaRange(skill, "stackDamageBonus", 0.25)
	return magicDamageAfterResistance(attacker, target, damageValue, tick)
}

func (w *World) SwordShieldValue(entity *Entity) int {
	return w.swordShieldValue(entity)
}

func (w *World) SwordRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{200, 300, 400})
	return physicalDamageAfterResistance(attacker, target, baseDamage+attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 1.5), tick)
}

func (w *World) releaseTankQ(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(tankHeroID).ReleaseQ; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) refreshTankWPassive(entity *Entity) {
	if h := heroHooksFor(tankHeroID).RefreshWPassive; h != nil {
		h(w, entity)
	}
}

func (w *World) releaseTankE(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(tankHeroID).ReleaseE; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) TankTargetsInRadius(entity *Entity, center Vector2, radius float64) []*Entity {
	return w.targetsInRadius(entity, center, radius)
}

func (w *World) TankMagicDamageAfterResistance(source *Entity, target *Entity, raw float64, tick uint64) int {
	return magicDamageAfterResistance(source, target, raw, tick)
}

func (w *World) TankTargetsInCone(entity *Entity, direction Vector2, coneRange float64, angleDegrees float64) []*Entity {
	return w.targetsInCone(entity, direction, coneRange, angleDegrees)
}

func (w *World) TankPhysicalDamageAfterResistance(source *Entity, target *Entity, raw float64, tick uint64) int {
	return physicalDamageAfterResistance(source, target, raw, tick)
}

func (w *World) startTankRDash(entity *Entity, targetPoint Vector2, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(tankHeroID).StartRDash; h != nil {
		h(w, entity, targetPoint, state, skill, tick, tickRate)
	}
}

func (w *World) releasePreparedTankR(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(tankHeroID).ReleasePreparedR; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) cancelTankRPreparedCast(entity *Entity) {
	if h := heroHooksFor(tankHeroID).CancelPreparedR; h != nil {
		h(entity)
	}
}

func (w *World) resolveTankRImpact(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(tankHeroID).ResolveRImpact; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func TankControlTicksAfterTenacity(target *Entity, ticks uint64, now uint64) uint64 {
	return controlTicksAfterTenacity(target, ticks, now)
}

func (w *World) releaseWarriorR(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(warriorHeroID).ReleaseR; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func warriorRDamage(target *Entity, skill config.SkillConfig, level int) float64 {
	if h := heroHooksFor(warriorHeroID).WarriorRDamage; h != nil {
		return h(target, skill, level)
	}
	return 0
}

func (w *World) WarriorTrueDamageAfterReduction(target *Entity, rawDamage float64, tick uint64) int {
	return trueDamageAfterReduction(target, rawDamage, tick)
}

func (w *World) stopWarriorE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(warriorHeroID).StopE; h != nil {
		h(w, entity, state, skill, tick, tickRate)
	}
}

func (w *World) tickWarriorJudgment(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(warriorHeroID).TickE; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) WarriorAttackCrits(attacker *Entity, target *Entity, tick uint64) bool {
	return w.attackCrits(attacker, target, tick)
}

func (w *World) WarriorPhysicalDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	return physicalDamageAfterResistance(attacker, target, rawDamage, tick)
}

func ApplyWarriorPhysicalDefenseShred(w *World, target *Entity, percent float64, untilTick uint64) {
	applyPhysicalDefenseShred(w, target, percent, untilTick)
}

func applyPhysicalDefenseShred(w *World, target *Entity, percent float64, untilTick uint64) {
	if target == nil || percent <= 0 {
		return
	}
	if target.Combat.PhysicalDefenseShredAmount > 0 {
		target.Stats.PhysicalDefense += target.Combat.PhysicalDefenseShredAmount
	}
	shred := target.Stats.PhysicalDefense * clamp(percent, 0, 1)
	target.Stats.PhysicalDefense -= shred
	if target.Stats.PhysicalDefense < 0 {
		target.Stats.PhysicalDefense = 0
	}
	target.Combat.PhysicalDefenseShredAmount = shred
	target.Combat.PhysicalDefenseShredUntil = untilTick
}
