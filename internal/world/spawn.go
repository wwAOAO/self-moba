package world

import (
	"l-battle/internal/config"
	"strconv"
)

func (w *World) SpawnHero(playerID string, hero config.HeroConfig, team Team) {
	if team != TeamRed {
		team = TeamBlue
	}
	entityID := playerEntityID(playerID)
	skillIDs := hero.Skills.SkillIDs()
	skills := make(map[string]SkillState, len(skillIDs))
	for _, skillID := range skillIDs {
		skills[skillID] = SkillState{SkillID: skillID}
	}
	position := w.spawnPosition(team)
	level := MinHeroLevel
	stats := heroStatsAtLevel(hero, level)
	w.applySwordCritOverflowStats(&Entity{HeroID: hero.HeroID}, &stats)
	nextLevelExp := w.nextLevelExp(level)
	startingSkillPoints := 1
	if entity := w.entities[entityID]; entity != nil {
		entity.Team = team
		entity.HeroID = hero.HeroID
		entity.Level = level
		entity.SkillPoints = startingSkillPoints
		entity.Gold = 0
		entity.Equipment = nil
		entity.Exp = 0
		entity.TotalExp = 0
		entity.NextLevelExp = nextLevelExp
		entity.Stats = stats
		entity.Radius = hero.Radius
		entity.Skills = skills
		entity.Position = position
		entity.Combat = CombatState{}
		entity.Passive = w.passiveStateForHero(hero)
		entity.Sword = swordStateForHero(hero.HeroID)
		entity.Warrior = WarriorState{}
		entity.Archer = ArcherState{}
		entity.Tank = TankState{}
		entity.Death = DeathState{}
		return
	}
	w.entities[entityID] = &Entity{
		ID:           entityID,
		Kind:         EntityKindPlayer,
		Team:         team,
		PlayerID:     playerID,
		HeroID:       hero.HeroID,
		Level:        level,
		SkillPoints:  startingSkillPoints,
		Gold:         0,
		Equipment:    nil,
		Exp:          0,
		TotalExp:     0,
		NextLevelExp: nextLevelExp,
		Stats:        stats,
		Radius:       hero.Radius,
		Skills:       skills,
		Position:     position,
		Passive:      w.passiveStateForHero(hero),
		Sword:        swordStateForHero(hero.HeroID),
	}
}

func (w *World) SpawnBattleUnits() {
	w.spawnUnit("enemy:blue-hero-1", EntityKindEnemyHero, TeamBlue, w.width/2-420, w.height/2+220, 18, Stats{
		HP:              1200,
		MaxHP:           1200,
		MP:              500,
		MaxMP:           500,
		Attack:          82,
		PhysicalDefense: 26,
		MagicDefense:    18,
		MoveSpeed:       4.2,
		AttackRange:     150,
		AttackSpeed:     1,
	})
	w.spawnUnit("minion:blue-melee-1", EntityKindMeleeMinion, TeamBlue, w.width/2-360, w.height/2+70, 14, Stats{
		HP:              420,
		MaxHP:           420,
		Attack:          32,
		PhysicalDefense: 8,
		MagicDefense:    4,
		MoveSpeed:       3,
		AttackRange:     70,
		AttackSpeed:     0.8,
	})
	w.spawnUnit("minion:blue-ranged-1", EntityKindRangedMinion, TeamBlue, w.width/2-430, w.height/2, 13, Stats{
		HP:              300,
		MaxHP:           300,
		Attack:          38,
		PhysicalDefense: 5,
		MagicDefense:    5,
		MoveSpeed:       3,
		AttackRange:     360,
		AttackSpeed:     0.7,
	})
	w.spawnUnit("minion:blue-siege-1", EntityKindSiegeMinion, TeamBlue, w.width/2-500, w.height/2-80, 18, Stats{
		HP:              680,
		MaxHP:           680,
		Attack:          62,
		PhysicalDefense: 14,
		MagicDefense:    8,
		MoveSpeed:       2.4,
		AttackRange:     430,
		AttackSpeed:     0.55,
	})
	w.spawnUnit("structure:blue-tower-1", EntityKindTower, TeamBlue, w.width/2-700, w.height/2+240, 34, Stats{
		HP:              2600,
		MaxHP:           2600,
		Attack:          180,
		PhysicalDefense: 80,
		MagicDefense:    60,
		AttackRange:     620,
		AttackSpeed:     0.75,
	})
	w.spawnUnit("structure:blue-barracks-1", EntityKindBarracks, TeamBlue, w.width/2-760, w.height/2-80, 40, Stats{
		HP:              3200,
		MaxHP:           3200,
		PhysicalDefense: 55,
		MagicDefense:    45,
	})
	w.spawnUnit("structure:blue-crystal", EntityKindCrystal, TeamBlue, w.width/2-900, w.height/2-260, 48, Stats{
		HP:              4500,
		MaxHP:           4500,
		PhysicalDefense: 70,
		MagicDefense:    70,
	})
	w.spawnUnit("enemy:hero-1", EntityKindEnemyHero, TeamRed, w.width/2+420, w.height/2-220, 18, Stats{
		HP:              1200,
		MaxHP:           1200,
		MP:              500,
		MaxMP:           500,
		Attack:          82,
		PhysicalDefense: 26,
		MagicDefense:    18,
		MoveSpeed:       4.2,
		AttackRange:     150,
		AttackSpeed:     1,
	})
	w.spawnUnit("minion:red-melee-1", EntityKindMeleeMinion, TeamRed, w.width/2+360, w.height/2-70, 14, Stats{
		HP:              420,
		MaxHP:           420,
		Attack:          32,
		PhysicalDefense: 8,
		MagicDefense:    4,
		MoveSpeed:       3,
		AttackRange:     70,
		AttackSpeed:     0.8,
	})
	w.spawnUnit("minion:red-ranged-1", EntityKindRangedMinion, TeamRed, w.width/2+430, w.height/2, 13, Stats{
		HP:              300,
		MaxHP:           300,
		Attack:          38,
		PhysicalDefense: 5,
		MagicDefense:    5,
		MoveSpeed:       3,
		AttackRange:     360,
		AttackSpeed:     0.7,
	})
	w.spawnUnit("minion:red-siege-1", EntityKindSiegeMinion, TeamRed, w.width/2+500, w.height/2+80, 18, Stats{
		HP:              680,
		MaxHP:           680,
		Attack:          62,
		PhysicalDefense: 14,
		MagicDefense:    8,
		MoveSpeed:       2.4,
		AttackRange:     430,
		AttackSpeed:     0.55,
	})
	w.spawnUnit("structure:red-tower-1", EntityKindTower, TeamRed, w.width/2+700, w.height/2-240, 34, Stats{
		HP:              2600,
		MaxHP:           2600,
		Attack:          180,
		PhysicalDefense: 80,
		MagicDefense:    60,
		AttackRange:     620,
		AttackSpeed:     0.75,
	})
	w.spawnUnit("structure:red-barracks-1", EntityKindBarracks, TeamRed, w.width/2+760, w.height/2+80, 40, Stats{
		HP:              3200,
		MaxHP:           3200,
		PhysicalDefense: 55,
		MagicDefense:    45,
	})
	w.spawnUnit("structure:red-crystal", EntityKindCrystal, TeamRed, w.width/2+900, w.height/2+260, 48, Stats{
		HP:              4500,
		MaxHP:           4500,
		PhysicalDefense: 70,
		MagicDefense:    70,
	})
}

func (w *World) SpawnTrainingDummy() {
	w.spawnDummy("dummy:training-1", w.width/2+180, w.height/2)
	w.spawnDummy("dummy:training-2", w.width/2+180, w.height/2+200)
}

func (w *World) spawnDummy(id string, x float64, y float64) {
	w.spawnUnit(id, EntityKindDummy, TeamNeutral, x, y, 28, Stats{
		HP:              3000,
		MaxHP:           3000,
		PhysicalDefense: 10,
		MagicDefense:    10,
	})
}

func (w *World) SpawnObject(kind EntityKind, team Team, x float64, y float64) (string, bool) {
	if team != TeamBlue && team != TeamRed && team != TeamNeutral {
		team = TeamNeutral
	}
	stats, radius, ok := unitTemplate(kind)
	if !ok {
		return "", false
	}
	if kind == EntityKindDummy {
		team = TeamNeutral
	}
	w.nextObjectID++
	id := "spawn:" + string(kind) + ":" + strconv.Itoa(w.nextObjectID)
	w.spawnUnit(id, kind, team, clamp(x, 0, w.width), clamp(y, 0, w.height), radius, stats)
	return id, true
}

func (w *World) spawnUnit(id string, kind EntityKind, team Team, x float64, y float64, radius float64, stats Stats) {
	if _, ok := w.entities[id]; ok {
		return
	}
	entity := &Entity{
		ID:     id,
		Kind:   kind,
		Team:   team,
		Stats:  stats,
		Radius: radius,
		Position: Vector2{
			X: x,
			Y: y,
		},
	}
	if kind == EntityKindEnemyHero {
		entity.Level = MinHeroLevel
		entity.NextLevelExp = w.nextLevelExp(entity.Level)
	}
	w.entities[id] = entity
}

func unitTemplate(kind EntityKind) (Stats, float64, bool) {
	switch kind {
	case EntityKindDummy:
		return Stats{HP: 3000, MaxHP: 3000, PhysicalDefense: 10, MagicDefense: 10}, 28, true
	case EntityKindEnemyHero:
		return Stats{HP: 1200, MaxHP: 1200, MP: 500, MaxMP: 500, Attack: 82, PhysicalDefense: 26, MagicDefense: 18, MoveSpeed: 4.2, AttackRange: 150, AttackSpeed: 1}, 18, true
	case EntityKindMeleeMinion:
		return Stats{HP: 420, MaxHP: 420, Attack: 32, PhysicalDefense: 8, MagicDefense: 4, MoveSpeed: 3, AttackRange: 70, AttackSpeed: 0.8}, 14, true
	case EntityKindRangedMinion:
		return Stats{HP: 300, MaxHP: 300, Attack: 38, PhysicalDefense: 5, MagicDefense: 5, MoveSpeed: 3, AttackRange: 360, AttackSpeed: 0.7}, 13, true
	case EntityKindSiegeMinion:
		return Stats{HP: 680, MaxHP: 680, Attack: 62, PhysicalDefense: 14, MagicDefense: 8, MoveSpeed: 2.4, AttackRange: 430, AttackSpeed: 0.55}, 18, true
	case EntityKindTower:
		return Stats{HP: 2600, MaxHP: 2600, Attack: 180, PhysicalDefense: 80, MagicDefense: 60, AttackRange: 620, AttackSpeed: 0.75}, 34, true
	case EntityKindBarracks:
		return Stats{HP: 3200, MaxHP: 3200, PhysicalDefense: 55, MagicDefense: 45}, 40, true
	case EntityKindCrystal:
		return Stats{HP: 4500, MaxHP: 4500, PhysicalDefense: 70, MagicDefense: 70}, 48, true
	default:
		return Stats{}, 0, false
	}
}

func (w *World) RemovePlayer(playerID string) {
	delete(w.entities, playerID)
	delete(w.entities, playerEntityID(playerID))
}

func (w *World) spawnPosition(team Team) Vector2 {
	if team == TeamRed {
		return Vector2{
			X: w.width/2 + 160,
			Y: w.height/2 - 160,
		}
	}
	return Vector2{
		X: w.width/2 - 160,
		Y: w.height/2 + 160,
	}
}

func playerEntityID(playerID string) string {
	return "player:" + playerID
}
