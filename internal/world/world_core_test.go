package world

import (
	"l-battle/internal/config"
	"math"
	"testing"
)

func TestSpawnHeroRefreshesTeamOnRejoin(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()

	w.SpawnHero("p1", hero, TeamRed)
	assertPlayerTeam(t, w, "p1", TeamRed)

	w.RemovePlayer("p1")
	w.SpawnHero("p1", hero, TeamBlue)
	assertPlayerTeam(t, w, "p1", TeamBlue)

	w.RemovePlayer("p1")
	w.SpawnHero("p1", hero, TeamRed)
	assertPlayerTeam(t, w, "p1", TeamRed)
}

func TestSpawnHeroOverwritesExistingPlayerTeam(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()

	w.SpawnHero("p1", hero, TeamRed)
	w.SpawnHero("p1", hero, TeamBlue)
	w.SpawnHero("p1", hero, TeamRed)

	assertPlayerTeam(t, w, "p1", TeamRed)
}

func TestMoveTargetAdvancesOnServerTick(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.MoveSpeed = 345
	w.SpawnHero("p1", hero, TeamBlue)

	player := w.entities[playerEntityID("p1")]
	startX := player.Position.X
	w.ApplyInput("p1", protocolPlayerInputMove(startX+100, player.Position.Y), 1, nil, 20)
	w.Tick(2, 20)

	if player.Position.X <= startX {
		t.Fatalf("player did not move toward server move target: got x=%f start=%f", player.Position.X, startX)
	}
	if got := player.Position.X - startX; math.Abs(got-17.25) > 0.001 {
		t.Fatalf("move distance = %f, want 17.25", got)
	}
}

func TestDeadPlayerStaysInWorldAndRespawnsAfter20Seconds(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.Attack = 2000
	hero.Base.AttackRange = DefaultMapWidth
	w.SpawnHero("blue", hero, TeamBlue)
	w.SpawnHero("red", hero, TeamRed)
	attacker := w.entities[playerEntityID("blue")]
	target := w.entities[playerEntityID("red")]
	target.Position = Vector2{X: 1200, Y: 900}

	w.ApplyInput("blue", protocolPlayerInputAttack(target.ID), 1, nil, 20)
	w.Tick(2, 20)

	if w.entities[target.ID] == nil {
		t.Fatal("dead player should stay in world")
	}
	if !target.Death.Dead {
		t.Fatal("target should be marked dead")
	}
	if canAttackTarget(attacker, target) {
		t.Fatal("dead player should not be attackable")
	}
	if target.Position.X != 1200 || target.Position.Y != 900 {
		t.Fatalf("dead player position = %+v, want death position", target.Position)
	}

	w.Tick(2+uint64(respawnSeconds*20), 20)

	if target.Death.Dead {
		t.Fatal("target should respawn after 20 seconds")
	}
	if target.Stats.HP != target.Stats.MaxHP {
		t.Fatalf("respawn hp = %d, want %d", target.Stats.HP, target.Stats.MaxHP)
	}
	spawn := w.spawnPosition(TeamRed)
	if target.Position != spawn {
		t.Fatalf("respawn position = %+v, want %+v", target.Position, spawn)
	}
}

func TestBaseRegenRestoresHPAndMPOverTime(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.HPRegen5 = 10
	hero.Base.MPRegen5 = 5
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.HP = player.Stats.MaxHP - 20
	player.Stats.MP = player.Stats.MaxMP - 20

	for tick := uint64(1); tick <= 100; tick++ {
		w.Tick(tick, 20)
	}

	if player.Stats.HP != player.Stats.MaxHP-10 {
		t.Fatalf("hp after 5s regen = %d, want %d", player.Stats.HP, player.Stats.MaxHP-10)
	}
	if math.Abs(player.Stats.MP-(player.Stats.MaxMP-15)) > 0.000001 {
		t.Fatalf("mp after 5s regen = %f, want %f", player.Stats.MP, player.Stats.MaxMP-15)
	}
}

func TestBaseRegenDoesNotExceedMaxHPOrMP(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.HPRegen5 = 10
	hero.Base.MPRegen5 = 10
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.HP = player.Stats.MaxHP - 1
	player.Stats.MP = player.Stats.MaxMP - 1

	for tick := uint64(1); tick <= 100; tick++ {
		w.Tick(tick, 20)
	}

	if player.Stats.HP != player.Stats.MaxHP {
		t.Fatalf("hp after regen = %d, want max %d", player.Stats.HP, player.Stats.MaxHP)
	}
	if player.Stats.MP != player.Stats.MaxMP {
		t.Fatalf("mp after regen = %f, want max %f", player.Stats.MP, player.Stats.MaxMP)
	}
}

func TestSpawnObjectCreatesUnit(t *testing.T) {
	w := testWorld(t)
	id, ok := w.SpawnObject(EntityKindMeleeMinion, TeamRed, 500, 600)
	if !ok {
		t.Fatal("spawn object failed")
	}
	entity := w.entities[id]
	if entity == nil {
		t.Fatalf("spawned entity %s not found", id)
	}
	if entity.Kind != EntityKindMeleeMinion || entity.Team != TeamRed {
		t.Fatalf("spawned entity kind/team = %s/%s", entity.Kind, entity.Team)
	}
}

func TestSpawnObjectRejectsUnsupportedKind(t *testing.T) {
	w := testWorld(t)
	if _, ok := w.SpawnObject(EntityKind("bad_kind"), TeamRed, 500, 600); ok {
		t.Fatal("unsupported kind should be rejected")
	}
}

func testWorld(t *testing.T) *World {
	t.Helper()
	loadedHeroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	heroConfigs := loadedHeroes.All()
	heroConfigs = append(heroConfigs, testHeroConfig())
	heroes, err := config.NewHeroStore(heroConfigs)
	if err != nil {
		t.Fatal(err)
	}
	levels, err := config.LoadLevels("../../configs/levels.json")
	if err != nil {
		t.Fatal(err)
	}
	rewards, err := config.LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.LoadSkills("../../configs/skills.json")
	if err != nil {
		t.Fatal(err)
	}
	equipment, err := config.LoadEquipment("../../configs/equipment.json")
	if err != nil {
		t.Fatal(err)
	}
	w := NewWorld(heroes, skills, levels, rewards, equipment)
	w.SpawnTrainingDummy()
	return w
}

func testHeroConfig() config.HeroConfig {
	return config.HeroConfig{
		HeroID: "test_hero",
		Base: config.BaseStats{
			HP:              1000,
			MP:              100,
			Attack:          50,
			PhysicalDefense: 10,
			MagicDefense:    10,
			MoveSpeed:       5,
			AttackRange:     120,
			AttackSpeed:     1,
			CritChance:      0.1,
		},
		Growth: config.BaseStats{
			HP:              10,
			MP:              2,
			Attack:          1,
			PhysicalDefense: 1,
			MagicDefense:    1,
			MoveSpeed:       0,
			AttackRange:     0,
			AttackSpeed:     0.01,
			CritChance:      0.001,
		},
		Radius: 12,
		Skills: config.HeroSkills{
			Passive: "passive",
			Q:       "q",
			W:       "w",
			E:       "e",
			R:       "r",
		},
	}
}
