package world

import (
	"testing"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
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
	w.SpawnHero("p1", hero, TeamBlue)

	player := w.entities[playerEntityID("p1")]
	startX := player.Position.X
	w.ApplyInput("p1", protocolPlayerInputMove(startX+100, player.Position.Y), 1, nil, 20)
	w.Tick(2, 20)

	if player.Position.X <= startX {
		t.Fatalf("player did not move toward server move target: got x=%f start=%f", player.Position.X, startX)
	}
}

func TestAttackTargetAutoAttacksOnServerTick(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)

	target := w.entities["dummy:training-1"]
	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 1, nil, 20)
	w.Tick(2, 20)

	if target.Combat.LastDamage <= 0 {
		t.Fatalf("target was not attacked by server tick")
	}
}

func TestHeroStatsAtMaxLevelUseGrowth(t *testing.T) {
	hero := testHeroConfig()
	hero.Growth.HP = 10
	hero.Growth.MP = 2
	hero.Growth.Attack = 3
	hero.Growth.PhysicalDefense = 4
	hero.Growth.MagicDefense = 5
	hero.Growth.MoveSpeed = 0.1
	hero.Growth.AttackRange = 1
	hero.Growth.AttackSpeed = 0.01

	stats := heroStatsAtLevel(hero, MaxHeroLevel)
	steps := MaxHeroLevel - MinHeroLevel

	if stats.MaxHP != hero.Base.HP+hero.Growth.HP*steps {
		t.Fatalf("max hp = %d", stats.MaxHP)
	}
	if stats.Attack != hero.Base.Attack+hero.Growth.Attack*steps {
		t.Fatalf("attack = %d", stats.Attack)
	}
	if stats.PhysicalDefense != hero.Base.PhysicalDefense+hero.Growth.PhysicalDefense*steps {
		t.Fatalf("physical defense = %d", stats.PhysicalDefense)
	}
	if stats.MagicDefense != hero.Base.MagicDefense+hero.Growth.MagicDefense*steps {
		t.Fatalf("magic defense = %d", stats.MagicDefense)
	}
	if stats.MoveSpeed != hero.Base.MoveSpeed+hero.Growth.MoveSpeed*float64(steps) {
		t.Fatalf("move speed = %f", stats.MoveSpeed)
	}
	if stats.AttackRange != hero.Base.AttackRange+hero.Growth.AttackRange*float64(steps) {
		t.Fatalf("attack range = %f", stats.AttackRange)
	}
	if stats.AttackSpeed != hero.Base.AttackSpeed+hero.Growth.AttackSpeed*float64(steps) {
		t.Fatalf("attack speed = %f", stats.AttackSpeed)
	}
}

func TestHeroStatsLevelIsClamped(t *testing.T) {
	hero := testHeroConfig()
	hero.Growth.HP = 10

	low := heroStatsAtLevel(hero, -1)
	high := heroStatsAtLevel(hero, 99)

	if low.MaxHP != hero.Base.HP {
		t.Fatalf("low level max hp = %d", low.MaxHP)
	}
	if high.MaxHP != hero.Base.HP+hero.Growth.HP*(MaxHeroLevel-MinHeroLevel) {
		t.Fatalf("high level max hp = %d", high.MaxHP)
	}
}

func TestMinionKillGrantsExperience(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.Attack = 1000
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := w.entities["minion:red-melee-1"]

	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 1, nil, 20)
	w.Tick(2, 20)

	if player.TotalExp != 62 {
		t.Fatalf("total exp = %f, want 62", player.TotalExp)
	}
}

func TestExperienceLevelsUpAndRecalculatesStats(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	startMaxHP := player.Stats.MaxHP

	w.addExperience(player, 280)

	if player.Level != 2 {
		t.Fatalf("level = %d, want 2", player.Level)
	}
	if player.Exp != 0 {
		t.Fatalf("exp = %f, want 0", player.Exp)
	}
	if player.NextLevelExp != 340 {
		t.Fatalf("next level exp = %f, want 340", player.NextLevelExp)
	}
	if player.Stats.MaxHP <= startMaxHP {
		t.Fatalf("max hp did not grow: got %d start %d", player.Stats.MaxHP, startMaxHP)
	}
}

func assertPlayerTeam(t *testing.T, w *World, playerID string, want Team) {
	t.Helper()
	entity := w.entities[playerEntityID(playerID)]
	if entity == nil {
		t.Fatalf("player %s not found", playerID)
	}
	if entity.Team != want {
		t.Fatalf("player %s team = %s, want %s", playerID, entity.Team, want)
	}
}

func testWorld(t *testing.T) *World {
	t.Helper()
	hero := testHeroConfig()
	heroes, err := config.NewHeroStore([]config.HeroConfig{hero})
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
	return NewWorld(heroes, levels, rewards)
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

func protocolPlayerInputMove(x float64, y float64) protocol.PlayerInput {
	return protocol.PlayerInput{
		Move: &protocol.MoveInput{
			TargetX: x,
			TargetY: y,
		},
	}
}

func protocolPlayerInputAttack(targetID string) protocol.PlayerInput {
	return protocol.PlayerInput{
		Attack: &protocol.AttackInput{
			TargetID: targetID,
		},
	}
}
