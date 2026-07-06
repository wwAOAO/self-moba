package archer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
)

func ApplyW(w *world.World, entity *world.Entity, cast protocol.CastInput, state world.SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != heroID {
		return
	}
	manaCost := skillList(skill, "manaCost", state.Level, []float64{70, 65, 60, 55, 50})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Skills[wID] = world.SkillState{
		SkillID:           state.SkillID,
		Level:             state.Level,
		CooldownUntilTick: tick + cooldownTicksFor(entity, int(skillList(skill, "cooldownMs", state.Level, []float64{14000, 11500, 9000, 6500, 4000})), tickRate),
		Stacks:            state.Stacks,
		StacksExpireTick:  state.StacksExpireTick,
	}
	w.LockAttackAfterCast(entity, tick, tickRate)

	target := world.Vector2{X: cast.TargetX, Y: cast.TargetY}
	dx, dy := normalize(target.X-entity.Position.X, target.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	centerAngle := math.Atan2(dy, dx)
	arrowCount := int(math.Round(skillList(skill, "arrowCount", state.Level, []float64{7, 8, 9, 10, 11})))
	if arrowCount < 1 {
		arrowCount = 1
	}
	coneAngle := skillMeta(skill, "coneAngleDegrees", 48) * math.Pi / 180
	startAngle := centerAngle - coneAngle/2
	step := 0.0
	if arrowCount > 1 {
		step = coneAngle / float64(arrowCount-1)
	}
	speedPerTick := skillMeta(skill, "projectileSpeed", 1500)
	if tickRate > 0 {
		speedPerTick /= float64(tickRate)
	}
	rangeValue := skillRange(skill, 1200)
	radius := skillMeta(skill, "projectileRadius", 16)
	groupID := w.NextProjectileID("projectile:archer_w_group:")
	for i := 0; i < arrowCount; i++ {
		angle := startAngle + step*float64(i)
		w.PutProjectile(&world.Projectile{
			ID:           w.NextProjectileID("projectile:archer_w:"),
			Kind:         "archer_volley_arrow",
			Team:         entity.Team,
			SourceID:     entity.ID,
			SkillID:      wID,
			GroupID:      groupID,
			Position:     entity.Position,
			Start:        entity.Position,
			Dir:          world.Vector2{X: math.Cos(angle), Y: math.Sin(angle)},
			SpeedPerTick: speedPerTick,
			Range:        rangeValue,
			Radius:       radius,
			Damage:       state.Level,
			CreatedAt:    tick,
			ExpiresAt:    tick + secondsToTicks(rangeValue/skillMeta(skill, "projectileSpeed", 1500)+0.2, tickRate),
			HitIDs:       make(map[string]bool),
		})
	}
}

func WDamage(w *world.World, entity *world.Entity, target *world.Entity, skill config.SkillConfig, level int, tick uint64) int {
	baseDamage := skillList(skill, "baseDamage", level, []float64{20, 35, 50, 65, 80})
	rawDamage := baseDamage + entity.Stats.Attack*skillMeta(skill, "adRatio", 1)
	return w.ArcherPhysicalDamageAfterResistance(entity, target, rawDamage, tick)
}
