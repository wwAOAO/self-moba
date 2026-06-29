package battle

import (
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func BuildSnapshot(roomID string, tick uint64, w *world.World) protocol.Snapshot {
	players := w.Players()
	dummies := w.Dummies()
	width, height := w.Size()
	snapshot := protocol.Snapshot{
		RoomID: roomID,
		Tick:   tick,
		Map: protocol.MapSnapshot{
			Width:  width,
			Height: height,
		},
		Players: make([]protocol.PlayerSnapshot, 0, len(players)),
		Dummies: make([]protocol.DummySnapshot, 0, len(dummies)),
	}
	for _, entity := range players {
		snapshot.Players = append(snapshot.Players, protocol.PlayerSnapshot{
			PlayerID: entity.PlayerID,
			HeroID:   entity.HeroID,
			X:        entity.Position.X,
			Y:        entity.Position.Y,
			Stats:    buildStatsSnapshot(entity.Stats),
			Skills:   buildSkillSnapshots(entity.Skills),
		})
	}
	for _, entity := range dummies {
		snapshot.Dummies = append(snapshot.Dummies, protocol.DummySnapshot{
			ID:          entity.ID,
			X:           entity.Position.X,
			Y:           entity.Position.Y,
			Radius:      entity.Radius,
			Stats:       buildStatsSnapshot(entity.Stats),
			LastHitTick: entity.Combat.LastHitTick,
			LastDamage:  entity.Combat.LastDamage,
		})
	}
	return snapshot
}

func buildStatsSnapshot(stats world.Stats) protocol.StatsSnapshot {
	return protocol.StatsSnapshot{
		HP:              stats.HP,
		MaxHP:           stats.MaxHP,
		MP:              stats.MP,
		MaxMP:           stats.MaxMP,
		Attack:          stats.Attack,
		PhysicalDefense: stats.PhysicalDefense,
		MagicDefense:    stats.MagicDefense,
		MoveSpeed:       stats.MoveSpeed,
		AttackRange:     stats.AttackRange,
		AttackSpeed:     stats.AttackSpeed,
	}
}

func buildSkillSnapshots(states map[string]world.SkillState) []protocol.SkillSnapshot {
	skills := make([]protocol.SkillSnapshot, 0, len(states))
	for _, state := range states {
		skills = append(skills, protocol.SkillSnapshot{
			SkillID:           state.SkillID,
			CooldownUntilTick: state.CooldownUntilTick,
		})
	}
	return skills
}
