package battle

import (
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func BuildSnapshot(roomID string, tick uint64, w *world.World) protocol.Snapshot {
	players := w.Players()
	dummies := w.Dummies()
	units := w.Units()
	walls := w.WindWalls()
	effects := w.SkillEffects()
	width, height := w.Size()
	snapshot := protocol.Snapshot{
		RoomID: roomID,
		Tick:   tick,
		Map: protocol.MapSnapshot{
			Width:  width,
			Height: height,
		},
		Players: make([]protocol.PlayerSnapshot, 0, len(players)),
		Units:   make([]protocol.UnitSnapshot, 0, len(units)),
		Dummies: make([]protocol.DummySnapshot, 0, len(dummies)),
		Effects: make([]protocol.EffectSnapshot, 0, len(walls)+len(effects)),
	}
	for _, entity := range players {
		stats := buildStatsSnapshot(entity.Stats)
		stats.MoveSpeed = world.EffectiveMoveSpeedAtTick(&entity, tick)
		stats.AttackSpeed = world.EffectiveAttackSpeedAtTick(&entity, tick)
		snapshot.Players = append(snapshot.Players, protocol.PlayerSnapshot{
			PlayerID:       entity.PlayerID,
			HeroID:         entity.HeroID,
			Team:           string(entity.Team),
			Level:          entity.Level,
			MaxLevel:       world.MaxHeroLevel,
			SkillPoints:    entity.SkillPoints,
			Exp:            entity.Exp,
			TotalExp:       entity.TotalExp,
			NextLevelExp:   entity.NextLevelExp,
			X:              entity.Position.X,
			Y:              entity.Position.Y,
			Stats:          stats,
			Skills:         buildSkillSnapshots(entity.Skills),
			Passive:        buildPassiveSnapshot(entity.Passive),
			LastHitTick:    entity.Combat.LastHitTick,
			LastDamage:     entity.Combat.LastDamage,
			LastDamageType: entity.Combat.LastDamageType,
			Dead:           entity.Death.Dead,
			RespawnTick:    entity.Death.RespawnTick,
			RespawnIn:      respawnInSeconds(tick, entity.Death),
			Control:        buildControlSnapshot(entity.Control),
			Sword:          buildSwordSnapshot(entity.Sword),
			Warrior:        buildWarriorSnapshot(entity.Warrior),
			Tank:           buildTankSnapshot(entity.Tank),
			Archer:         buildArcherSnapshot(entity.Archer),
		})
	}
	for _, entity := range dummies {
		snapshot.Dummies = append(snapshot.Dummies, protocol.DummySnapshot{
			ID:             entity.ID,
			X:              entity.Position.X,
			Y:              entity.Position.Y,
			Radius:         entity.Radius,
			Stats:          buildStatsSnapshot(entity.Stats),
			LastHitTick:    entity.Combat.LastHitTick,
			LastDamage:     entity.Combat.LastDamage,
			LastDamageType: entity.Combat.LastDamageType,
		})
	}
	for _, entity := range units {
		stats := buildStatsSnapshot(entity.Stats)
		stats.MoveSpeed = world.EffectiveMoveSpeedAtTick(&entity, tick)
		stats.AttackSpeed = world.EffectiveAttackSpeedAtTick(&entity, tick)
		snapshot.Units = append(snapshot.Units, protocol.UnitSnapshot{
			ID:             entity.ID,
			Kind:           string(entity.Kind),
			Team:           string(entity.Team),
			X:              entity.Position.X,
			Y:              entity.Position.Y,
			Radius:         entity.Radius,
			Stats:          stats,
			LastHitTick:    entity.Combat.LastHitTick,
			LastDamage:     entity.Combat.LastDamage,
			LastDamageType: entity.Combat.LastDamageType,
			Control:        buildControlSnapshot(entity.Control),
		})
	}
	for _, effect := range walls {
		snapshot.Effects = append(snapshot.Effects, protocol.EffectSnapshot{
			ID:        effect.ID,
			Kind:      "wind_wall",
			Team:      string(effect.Team),
			X:         effect.Center.X,
			Y:         effect.Center.Y,
			DirX:      effect.Dir.X,
			DirY:      effect.Dir.Y,
			Width:     effect.Width,
			ExpiresAt: effect.ExpiresAt,
		})
	}
	for _, effect := range effects {
		snapshot.Effects = append(snapshot.Effects, protocol.EffectSnapshot{
			ID:        effect.ID,
			Kind:      effect.Kind,
			Team:      string(effect.Team),
			X:         effect.Start.X,
			Y:         effect.Start.Y,
			EndX:      effect.End.X,
			EndY:      effect.End.Y,
			DirX:      effect.Dir.X,
			DirY:      effect.Dir.Y,
			Width:     effect.Width,
			Height:    effect.Height,
			Radius:    effect.Radius,
			Range:     effect.Range,
			Speed:     effect.Speed,
			CreatedAt: effect.CreatedAt,
			ExpiresAt: effect.ExpiresAt,
			Count:     effect.Count,
		})
	}
	return snapshot
}

func buildStatsSnapshot(stats world.Stats) protocol.StatsSnapshot {
	return protocol.StatsSnapshot{
		HP:                   stats.HP,
		MaxHP:                stats.MaxHP,
		BonusHP:              stats.BonusHP,
		MP:                   stats.MP,
		MaxMP:                stats.MaxMP,
		HPRegen5:             stats.HPRegen5,
		MPRegen5:             stats.MPRegen5,
		Attack:               stats.Attack,
		BonusAttack:          stats.BonusAttack,
		AbilityPower:         stats.AbilityPower,
		AbilityHaste:         stats.AbilityHaste,
		DamageReduce:         stats.DamageReduce,
		PhysicalDefense:      stats.PhysicalDefense,
		BonusPhysicalDefense: stats.BonusPhysicalDefense,
		PhysicalPenPercent:   stats.PhysicalPenPercent,
		PhysicalPenFlat:      stats.PhysicalPenFlat,
		PhysicalDamageReduce: stats.PhysicalDamageReduce,
		MagicDefense:         stats.MagicDefense,
		BonusMagicDefense:    stats.BonusMagicDefense,
		MagicPenPercent:      stats.MagicPenPercent,
		MagicPenFlat:         stats.MagicPenFlat,
		MagicDamageReduce:    stats.MagicDamageReduce,
		MoveSpeed:            stats.MoveSpeed,
		AttackRange:          stats.AttackRange,
		AttackSpeed:          stats.AttackSpeed,
		BaseAttackSpeed:      stats.BaseAttackSpeed,
		AttackSpeedBonus:     stats.AttackSpeedBonus,
		AttackSpeedRatio:     stats.AttackSpeedRatio,
		AttackSpeedSlow:      stats.AttackSpeedSlow,
		CritChance:           stats.CritChance,
	}
}

func buildSkillSnapshots(states map[string]world.SkillState) []protocol.SkillSnapshot {
	skills := make([]protocol.SkillSnapshot, 0, len(states))
	for _, state := range states {
		skills = append(skills, protocol.SkillSnapshot{
			SkillID:           state.SkillID,
			Level:             state.Level,
			CooldownUntilTick: state.CooldownUntilTick,
			Stacks:            state.Stacks,
			StacksExpireTick:  state.StacksExpireTick,
		})
	}
	return skills
}

func buildPassiveSnapshot(state world.PassiveState) protocol.PassiveSnapshot {
	return protocol.PassiveSnapshot{
		SwordIntent:    state.SwordIntent,
		MaxSwordIntent: state.MaxSwordIntent,
		Shield:         state.Shield,
		MaxShield:      state.MaxShield,
	}
}

func buildControlSnapshot(state world.ControlState) protocol.ControlSnapshot {
	return protocol.ControlSnapshot{
		AirborneUntilTick:     state.AirborneUntilTick,
		DashUntilTick:         state.DashUntilTick,
		ActionLockedUntilTick: state.ActionLockedUntilTick,
		StunnedUntilTick:      state.StunnedUntilTick,
		SilencedUntilTick:     state.SilencedUntilTick,
		TenacityUntilTick:     state.TenacityUntilTick,
		MoveSpeedSlow:         state.MoveSpeedSlow,
		MoveSpeedSlowUntil:    state.MoveSpeedSlowUntil,
	}
}

func buildWarriorSnapshot(state world.WarriorState) protocol.WarriorSnapshot {
	return protocol.WarriorSnapshot{
		JudgmentUntilTick: state.JudgmentUntilTick,
	}
}

func buildTankSnapshot(state world.TankState) protocol.TankSnapshot {
	return protocol.TankSnapshot{
		ThunderclapAftershockUntil: state.ThunderclapAftershockUntil,
	}
}

func buildArcherSnapshot(state world.ArcherState) protocol.ArcherSnapshot {
	return protocol.ArcherSnapshot{
		FocusStacks:      state.FocusStacks,
		FocusExpireTick:  state.FocusExpireTick,
		FocusActiveUntil: state.FocusActiveUntil,
		FocusAttackSpeed: state.FocusAttackSpeed,
	}
}

func buildSwordSnapshot(state world.SwordState) protocol.SwordSnapshot {
	targetUntil := make(map[string]uint64, len(state.SweepingBladeTargetUntil))
	for id, until := range state.SweepingBladeTargetUntil {
		targetUntil[id] = until
	}
	return protocol.SwordSnapshot{
		SweepingBladeTargetUntil: targetUntil,
	}
}

func respawnInSeconds(tick uint64, death world.DeathState) float64 {
	if !death.Dead || death.RespawnTick <= tick || death.RespawnTickRate <= 0 {
		return 0
	}
	return float64(death.RespawnTick-tick) / float64(death.RespawnTickRate)
}
