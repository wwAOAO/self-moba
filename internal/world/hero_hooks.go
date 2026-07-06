package world

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"math"
)

type HeroCastHandler func(w *World, entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int)
type HeroTickHandler func(w *World, entity *Entity, tick uint64, tickRate int)
type HeroHitHandler func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)
type HeroKillHandler func(w *World, killer *Entity, target *Entity)
type HeroSkillUpgradeHandler func(w *World, entity *Entity, skillID string)

type HeroHooks struct {
	Cast           map[string]HeroCastHandler
	Tick           HeroTickHandler
	OnBasicHit     HeroHitHandler
	OnSkillHit     HeroHitHandler
	OnKill         HeroKillHandler
	OnSkillUpgrade HeroSkillUpgradeHandler

	FocusBonusDamage func(w *World, attacker *Entity, target *Entity, tick uint64) int
	ApplyFrostShot   func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)
	WDamage          func(w *World, entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64) int
	ArcherRDamage    func(w *World, entity *Entity, target *Entity, skill config.SkillConfig, level int, tick uint64, multiplier float64) int
	RStunTicks       func(projectile *Projectile, skill config.SkillConfig, tickRate int) uint64
	ApplyRSplash     func(w *World, source *Entity, primary *Entity, projectile *Projectile, skill config.SkillConfig, tick uint64, tickRate int)

	RageCritChance func(entity *Entity, skill config.SkillConfig) float64

	ActivateEZone      func(w *World, entity *Entity, center Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int)
	DetonateE          func(w *World, entity *Entity, skill config.SkillConfig, tick uint64, tickRate int)
	MageQDamage        func(w *World, attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, multiplier float64, tick uint64) int
	ApplyUltimateIllum func(w *World, source *Entity, target *Entity, tick uint64, tickRate int)

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
	RefreshGranite    func(w *World, entity *Entity)
	RefreshGraniteMax func(w *World, entity *Entity)
	WBonusDamage      func(w *World, attacker *Entity, tick uint64) float64
	TankQDamage       func(w *World, attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int

	StopE          func(w *World, entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int)
	QBonusDamage   func(w *World, attacker *Entity, tick uint64) float64
	WarriorRDamage func(target *Entity, skill config.SkillConfig, level int) float64
}

var heroHooks = map[string]HeroHooks{}

func RegisterHeroHooks(heroID string, hooks HeroHooks) {
	heroHooks[heroID] = hooks
}

func heroHooksFor(heroID string) HeroHooks {
	return heroHooks[heroID]
}

func heroHooksForEntity(entity *Entity) HeroHooks {
	if entity == nil {
		return HeroHooks{}
	}
	return heroHooksFor(entity.HeroID)
}

func heroCastHandlerFor(heroID string, skillID string) HeroCastHandler {
	if handlers := heroHooksFor(heroID).Cast; handlers != nil {
		return handlers[skillID]
	}
	return nil
}

func (w *World) tickHero(entity *Entity, tick uint64, tickRate int) {
	if h := heroHooksForEntity(entity).Tick; h != nil {
		h(w, entity, tick, tickRate)
	}
}

func (w *World) onHeroBasicHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksForEntity(source).OnBasicHit; h != nil {
		h(w, source, target, tick, tickRate)
	}
}

func (w *World) onHeroSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksForEntity(source).OnSkillHit; h != nil {
		h(w, source, target, tick, tickRate)
	}
}

func (w *World) onHeroKill(killer *Entity, target *Entity) {
	for _, hooks := range heroHooks {
		if hooks.OnKill != nil {
			hooks.OnKill(w, killer, target)
		}
	}
}

func (w *World) onHeroSkillUpgrade(entity *Entity, skillID string) {
	if h := heroHooksForEntity(entity).OnSkillUpgrade; h != nil {
		h(w, entity, skillID)
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

func bladeRageCritChance(entity *Entity, skill config.SkillConfig) float64 {
	if h := heroHooksFor(bladeHeroID).RageCritChance; h != nil {
		return h(entity, skill)
	}
	return 0
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

func (w *World) activateMageEZone(entity *Entity, center Vector2, level int, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ActivateEZone; h != nil {
		h(w, entity, center, level, skill, tick, tickRate)
	}
}

func (w *World) detonateMageE(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).DetonateE; h != nil {
		h(w, entity, skill, tick, tickRate)
	}
}

func (w *World) MageTargetsInRadius(entity *Entity, center Vector2, radius float64) []*Entity {
	return w.targetsInRadius(entity, center, radius)
}

func (w *World) ApplyMageMoveSpeedSlow(target *Entity, slow float64, until uint64) {
	applyMoveSpeedSlow(target, slow, until)
}

func (w *World) ApplyMageIlluminationOnSkillHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	w.onHeroSkillHit(source, target, tick, tickRate)
}

func (w *World) ApplyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	w.applyMageIlluminationOnUltimateHit(source, target, tick, tickRate)
}

func (w *World) applyMageIlluminationOnUltimateHit(source *Entity, target *Entity, tick uint64, tickRate int) {
	if h := heroHooksFor(mageHeroID).ApplyUltimateIllum; h != nil {
		h(w, source, target, tick, tickRate)
	}
}

func (w *World) MagePassiveSkill(entity *Entity) config.SkillConfig {
	return w.heroPassiveSkill(entity)
}

func MageSkillMetaCurveByLevel(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	return skillMetaCurveByLevel(skill, valueKey, levelKey, level, fallback)
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

func (w *World) refreshTankWPassive(entity *Entity) {
	if h := heroHooksFor(tankHeroID).RefreshWPassive; h != nil {
		h(w, entity)
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
