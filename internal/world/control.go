package world

import "math"

const (
	doctorHeroID                 = "doctor"
	doctorPassiveSkillID         = "doctor_passive"
	doctorCanisterEffectKind     = "doctor_canister"
	doctorPassiveCooldownSeconds = 45
	doctorPassiveCostRatio       = 0.07
	doctorCanisterRadius         = 55
)

func (w *World) ApplyStun(target *Entity, until uint64, tick uint64, tickRate int) bool {
	if target == nil {
		return false
	}
	if w.consumeDoctorPassive(target, tick, tickRate) {
		return false
	}
	target.Control.StunnedUntilTick = until
	return true
}

func (w *World) ApplyRoot(target *Entity, until uint64, tick uint64, tickRate int) bool {
	if target == nil {
		return false
	}
	if w.consumeDoctorPassive(target, tick, tickRate) {
		return false
	}
	target.Control.RootedUntilTick = until
	return true
}

func (w *World) ApplyAirborne(target *Entity, until uint64, tick uint64, tickRate int) bool {
	if target == nil {
		return false
	}
	if w.consumeDoctorPassive(target, tick, tickRate) {
		return false
	}
	target.Control.AirborneUntilTick = until
	return true
}

func (w *World) ExtendAirborne(target *Entity, ticks uint64, tick uint64, tickRate int) bool {
	if target == nil {
		return false
	}
	if w.consumeDoctorPassive(target, tick, tickRate) {
		return false
	}
	target.Control.AirborneUntilTick += ticks
	return true
}

func (w *World) consumeDoctorPassive(target *Entity, tick uint64, tickRate int) bool {
	if w == nil || target == nil || target.HeroID != doctorHeroID || target.Stats.HP <= 0 || tickRate <= 0 {
		return false
	}
	if tick < target.Passive.DoctorPassiveCooldownUntil {
		return false
	}
	skill := w.skillConfig(doctorPassiveSkillID)
	cooldownSeconds := skillMetaRange(skill, "cooldownSeconds", doctorPassiveCooldownSeconds)
	costRatio := skillMetaRange(skill, "healthCostCurrentRatio", doctorPassiveCostRatio)
	radius := skillMetaRange(skill, "canisterRadius", doctorCanisterRadius)
	cooldownUntil := tick + secondsToTicks(cooldownSeconds, tickRate)

	beforeHP := target.Stats.HP
	target.Stats.HP = math.Max(1, target.Stats.HP-target.Stats.HP*costRatio)
	w.refreshPlayerStatsAfterHPChange(target, beforeHP)

	target.Passive.DoctorPassiveCooldownUntil = cooldownUntil
	if state, ok := target.Skills[doctorPassiveSkillID]; ok {
		state.CooldownUntilTick = cooldownUntil
		target.Skills[doctorPassiveSkillID] = state
	}
	w.removeDoctorCanister(target)
	id := w.NextEffectID("effect:doctor_canister:")
	target.Passive.DoctorCanisterEffectID = id
	target.Passive.DoctorCanisterPosition = target.Position
	target.Passive.DoctorCanisterRadius = radius
	target.Passive.DoctorCanisterExpiresAt = cooldownUntil
	w.PutSkillEffect(SkillEffect{
		ID:           id,
		Kind:         doctorCanisterEffectKind,
		Team:         target.Team,
		SourceID:     target.ID,
		SourceHeroID: target.HeroID,
		Start:        target.Position,
		Radius:       radius,
		CreatedAt:    tick,
		ExpiresAt:    cooldownUntil,
	})
	return true
}

func (w *World) removeDoctorCanister(entity *Entity) {
	if w == nil || entity == nil {
		return
	}
	w.RemoveSkillEffect(entity.Passive.DoctorCanisterEffectID)
	entity.Passive.DoctorCanisterEffectID = ""
	entity.Passive.DoctorCanisterPosition = Vector2{}
	entity.Passive.DoctorCanisterRadius = 0
	entity.Passive.DoctorCanisterExpiresAt = 0
}

func (w *World) RemoveDoctorCanister(entity *Entity) {
	w.removeDoctorCanister(entity)
}
