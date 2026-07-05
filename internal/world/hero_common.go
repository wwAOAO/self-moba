package world

func (w *World) lockAttackAfterCast(entity *Entity, tick uint64, tickRate int) {
	nextAttackTick := tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(entity, tick), tickRate)
	if entity.Combat.NextAttackTick < nextAttackTick {
		entity.Combat.NextAttackTick = nextAttackTick
	}
}
