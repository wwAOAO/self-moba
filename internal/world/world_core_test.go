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

func TestMovingEntitiesDoNotOverlapCollisionRadius(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.MoveSpeed = 345
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	minion := &Entity{
		ID:       "spawn:test-minion-collision",
		Kind:     EntityKindMeleeMinion,
		Team:     TeamRed,
		Position: Vector2{X: player.Position.X + 30, Y: player.Position.Y},
		Radius:   14,
		Stats:    Stats{HP: 420, MaxHP: 420},
	}
	w.entities[minion.ID] = minion

	w.ApplyInput("p1", protocolPlayerInputMove(player.Position.X+100, player.Position.Y), 1, nil, 20)
	w.Tick(2, 20)

	if got, want := distance(player.Position, minion.Position), player.Radius+minion.Radius; got < want-0.001 {
		t.Fatalf("entity distance = %f, want at least %f", got, want)
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
	w.Tick(7, 20)

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

	w.Tick(target.Death.RespawnTick, 20)

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
	placeEntity(player, 3000, 3000)
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

func TestFountainsSpawnAtTeamSpawnPositions(t *testing.T) {
	w := testWorld(t)
	blue := w.entities["spawn:fountain:blue"]
	red := w.entities["spawn:fountain:red"]
	if blue == nil || red == nil {
		t.Fatal("missing fountains")
	}
	if blue.Position != w.spawnPosition(TeamBlue) || red.Position != w.spawnPosition(TeamRed) {
		t.Fatalf("fountain positions = %+v/%+v, want %+v/%+v", blue.Position, red.Position, w.spawnPosition(TeamBlue), w.spawnPosition(TeamRed))
	}
	if canAttackTarget(&Entity{ID: "attacker", Team: TeamRed, Stats: Stats{HP: 100}}, blue) {
		t.Fatal("fountain should not be attackable")
	}
}

func TestFountainRegeneratesFriendlyHero(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.HP = player.Stats.MaxHP - 100
	player.Stats.MP = player.Stats.MaxMP - 100

	w.Tick(1, 20)

	if player.Stats.HP != player.Stats.MaxHP-80 {
		t.Fatalf("hp = %d, want %d", player.Stats.HP, player.Stats.MaxHP-80)
	}
	if player.Stats.MP != player.Stats.MaxMP-98 {
		t.Fatalf("mp = %f, want %f", player.Stats.MP, player.Stats.MaxMP-98)
	}
}

func TestEnemyInFountainRangeGetsShot(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("red", hero, TeamRed)
	target := w.entities[playerEntityID("red")]
	target.Position = w.spawnPosition(TeamBlue)
	target.Stats.PhysicalDefense = 0
	target.Stats.MagicDefense = 0
	startHP := target.Stats.HP

	w.Tick(1, 20)
	if len(w.projectiles) == 0 {
		t.Fatal("fountain should fire projectile")
	}
	for tick := uint64(2); tick <= 20; tick++ {
		w.Tick(tick, 20)
		if target.Stats.HP < startHP {
			break
		}
	}

	if target.Stats.HP != startHP-250 {
		t.Fatalf("target hp = %d, want %d", target.Stats.HP, startHP-250)
	}
	if len(target.Combat.DamageEvents) != 3 {
		t.Fatalf("damage events = %+v, want 3", target.Combat.DamageEvents)
	}
}

func TestEnemyMinionInFountainRangeGetsShot(t *testing.T) {
	w := testWorld(t)
	minion := &Entity{
		ID:       "spawn:red-minion-in-fountain",
		Kind:     EntityKindMeleeMinion,
		Team:     TeamRed,
		Position: w.spawnPosition(TeamBlue),
		Radius:   14,
		Stats:    Stats{HP: 420, MaxHP: 420},
	}
	w.entities[minion.ID] = minion

	w.Tick(1, 20)

	if len(w.projectiles) == 0 {
		t.Fatal("fountain should fire projectile at enemy minion")
	}
}

func TestMinionWavesSpawnEvery30Seconds(t *testing.T) {
	w := testWorld(t)

	w.Tick(599, 20)
	if got := countLaneMinions(w); got != 0 {
		t.Fatalf("lane minions before 30s = %d, want 0", got)
	}
	w.Tick(600, 20)
	if got := countLaneMinions(w); got != 2 {
		t.Fatalf("lane minions immediately after first wave starts = %d, want 2", got)
	}
	for tick := uint64(601); tick <= 700; tick++ {
		w.Tick(tick, 20)
	}
	if got := countLaneMinions(w); got != 14 {
		t.Fatalf("lane minions after first wave finishes = %d, want 14", got)
	}
	w.Tick(1200, 20)
	for tick := uint64(1201); tick <= 1300; tick++ {
		w.Tick(tick, 20)
	}
	if got := countLaneMinions(w); got != 28 {
		t.Fatalf("lane minions after second wave = %d, want 28", got)
	}
}

func TestMinionWaveComposition(t *testing.T) {
	w := testWorld(t)
	w.spawnMinionWave(TeamBlue, 1)

	counts := map[EntityKind]int{}
	for _, entity := range w.entities {
		if entity.Lane.Active && entity.Team == TeamBlue {
			counts[entity.Kind]++
		}
	}
	if counts[EntityKindMeleeMinion] != 3 || counts[EntityKindRangedMinion] != 3 || counts[EntityKindSiegeMinion] != 1 {
		t.Fatalf("wave counts = %+v, want 3 melee, 3 ranged, 1 siege", counts)
	}
}

func TestLaneMinionMovesTowardEnemyFountain(t *testing.T) {
	w := testWorld(t)
	w.spawnMinionWave(TeamBlue, 1)
	minion := firstLaneMinion(w, TeamBlue)
	if minion == nil {
		t.Fatal("missing blue lane minion")
	}
	startDistance := distance(minion.Position, w.spawnPosition(TeamRed))

	for tick := uint64(2); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}

	if got := distance(minion.Position, w.spawnPosition(TeamRed)); got >= startDistance {
		t.Fatalf("minion distance to enemy fountain = %f, want less than %f", got, startDistance)
	}
}

func TestLaneMinionAttacksEnemyOnRoute(t *testing.T) {
	w := testWorld(t)
	w.spawnMinionWave(TeamBlue, 1)
	blue := firstLaneMinion(w, TeamBlue)
	if blue == nil {
		t.Fatal("missing blue lane minion")
	}
	red := &Entity{
		ID:       "spawn:test-red-minion",
		Kind:     EntityKindMeleeMinion,
		Team:     TeamRed,
		Position: Vector2{X: blue.Position.X + 50, Y: blue.Position.Y},
		Radius:   14,
		Stats:    Stats{HP: 420, MaxHP: 420, Attack: 32, MoveSpeed: 260, AttackRange: 70, AttackSpeed: 0.8},
		Lane:     LaneState{Active: true, RouteTarget: w.spawnPosition(TeamBlue), LastOnLaneTick: 1},
	}
	w.entities[red.ID] = red

	for tick := uint64(2); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}

	if red.Stats.HP >= red.Stats.MaxHP {
		t.Fatalf("red minion hp = %d, want damaged", red.Stats.HP)
	}
}

func TestLaneMinionReturnsAfterLeavingRouteTooLong(t *testing.T) {
	w := testWorld(t)
	w.spawnMinionWave(TeamBlue, 1)
	minion := firstLaneMinion(w, TeamBlue)
	if minion == nil {
		t.Fatal("missing blue lane minion")
	}
	routeStart := w.spawnPosition(TeamBlue)
	routeEnd := w.spawnPosition(TeamRed)
	dx, dy := normalize(routeEnd.X-routeStart.X, routeEnd.Y-routeStart.Y)
	offRoute := Vector2{X: minion.Position.X - dy*900, Y: minion.Position.Y + dx*900}
	target := &Entity{
		ID:       "spawn:far-red-hero",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: offRoute,
		Radius:   18,
		Stats:    Stats{HP: 1200, MaxHP: 1200},
	}
	w.entities[target.ID] = target
	minion.Position = offRoute
	minion.Intent.AttackTargetID = target.ID
	minion.Lane.LastOnLaneTick = 1

	w.Tick(102, 20)

	if minion.Intent.AttackTargetID != "" {
		t.Fatalf("attack target = %q, want cleared", minion.Intent.AttackTargetID)
	}
}

func TestBaseRegenDoesNotExceedMaxHPOrMP(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.HPRegen5 = 10
	hero.Base.MPRegen5 = 10
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	placeEntity(player, 3000, 3000)
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

func countLaneMinions(w *World) int {
	total := 0
	for _, entity := range w.entities {
		if entity.Lane.Active {
			total++
		}
	}
	return total
}

func firstLaneMinion(w *World, team Team) *Entity {
	for _, entity := range w.entities {
		if entity.Lane.Active && entity.Team == team {
			return entity
		}
	}
	return nil
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
	skills, err := config.LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	equipment, err := config.LoadEquipment("../../configs/equipment")
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
