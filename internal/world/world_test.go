package world

import (
	"math"
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

func TestOpposingTeamPlayersCanAttackEachOther(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.AttackRange = 1000
	w.SpawnHero("blue", hero, TeamBlue)
	w.SpawnHero("red", hero, TeamRed)
	attacker := w.entities[playerEntityID("blue")]
	target := w.entities[playerEntityID("red")]
	target.Position = attacker.Position
	startHP := target.Stats.HP

	w.ApplyInput("blue", protocolPlayerInputAttack(target.ID), 1, nil, 20)
	w.Tick(2, 20)

	if target.Stats.HP >= startHP {
		t.Fatalf("target hp = %d, start hp = %d; opposing player was not damaged", target.Stats.HP, startHP)
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
	stepValue := float64(steps)

	if stats.MaxHP != hero.Base.HP+hero.Growth.HP*steps {
		t.Fatalf("max hp = %d", stats.MaxHP)
	}
	if stats.Attack != hero.Base.Attack+hero.Growth.Attack*stepValue {
		t.Fatalf("attack = %f", stats.Attack)
	}
	if stats.PhysicalDefense != hero.Base.PhysicalDefense+hero.Growth.PhysicalDefense*stepValue {
		t.Fatalf("physical defense = %f", stats.PhysicalDefense)
	}
	if stats.MagicDefense != hero.Base.MagicDefense+hero.Growth.MagicDefense*stepValue {
		t.Fatalf("magic defense = %f", stats.MagicDefense)
	}
	if stats.MoveSpeed != hero.Base.MoveSpeed+hero.Growth.MoveSpeed*float64(steps) {
		t.Fatalf("move speed = %f", stats.MoveSpeed)
	}
	if stats.AttackRange != hero.Base.AttackRange+hero.Growth.AttackRange*float64(steps) {
		t.Fatalf("attack range = %f", stats.AttackRange)
	}
	if stats.BaseAttackSpeed != hero.Base.AttackSpeed*(1+hero.Growth.AttackSpeed*float64(steps)) {
		t.Fatalf("base attack speed = %f", stats.BaseAttackSpeed)
	}
	if stats.AttackSpeed != stats.BaseAttackSpeed {
		t.Fatalf("attack speed = %f", stats.AttackSpeed)
	}
	if stats.CritChance != hero.Base.CritChance+hero.Growth.CritChance*float64(steps) {
		t.Fatalf("crit chance = %f", stats.CritChance)
	}
}

func TestSwordConfiguredStatsAtLevel18(t *testing.T) {
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(swordHeroID)
	if !ok {
		t.Fatal("sword hero not found")
	}
	stats := heroStatsAtLevel(hero, MaxHeroLevel)

	if hero.Resource != "sword_intent" {
		t.Fatalf("sword resource = %s, want sword_intent", hero.Resource)
	}
	if stats.MaxHP != 1969 {
		t.Fatalf("sword level 18 hp = %d, want 1969", stats.MaxHP)
	}
	if stats.Attack != 111 {
		t.Fatalf("sword level 18 attack = %f, want 111", stats.Attack)
	}
	if math.Abs(stats.AttackSpeed-1.111715) > 0.000001 {
		t.Fatalf("sword level 18 attack speed = %f, want 1.111715", stats.AttackSpeed)
	}
	if stats.PhysicalDefense != 87.8 {
		t.Fatalf("sword level 18 armor = %f, want 87.8", stats.PhysicalDefense)
	}
	if stats.MagicDefense != 53.25 {
		t.Fatalf("sword level 18 magic resist = %f, want 53.25", stats.MagicDefense)
	}
}

func TestTankConfiguredStatsAtLevel18(t *testing.T) {
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("tank")
	if !ok {
		t.Fatal("tank hero not found")
	}
	stats := heroStatsAtLevel(hero, MaxHeroLevel)

	if stats.MaxHP != 2569 {
		t.Fatalf("tank level 18 hp = %d, want 2569", stats.MaxHP)
	}
	if math.Abs(stats.MaxMP-1302.2) > 0.000001 {
		t.Fatalf("tank level 18 mp = %f, want 1302.2", stats.MaxMP)
	}
	if math.Abs(stats.Attack-129.97) > 0.000001 {
		t.Fatalf("tank level 18 attack = %f, want 129.97", stats.Attack)
	}
	if math.Abs(stats.PhysicalDefense-103.75) > 0.000001 {
		t.Fatalf("tank level 18 armor = %f, want 103.75", stats.PhysicalDefense)
	}
	if math.Abs(stats.MagicDefense-53.35) > 0.000001 {
		t.Fatalf("tank level 18 magic resist = %f, want 53.35", stats.MagicDefense)
	}
	if math.Abs(stats.AttackSpeed-1.006764) > 0.000001 {
		t.Fatalf("tank level 18 attack speed = %f, want 1.006764", stats.AttackSpeed)
	}
	if stats.MoveSpeed != 335 {
		t.Fatalf("tank move speed = %f, want 335", stats.MoveSpeed)
	}
	if stats.AttackRange != 125 {
		t.Fatalf("tank attack range = %f, want 125", stats.AttackRange)
	}
	if math.Abs(stats.HPRegen5-16.35) > 0.000001 {
		t.Fatalf("tank level 18 hp regen = %f, want 16.35", stats.HPRegen5)
	}
	if math.Abs(stats.MPRegen5-16.67) > 0.000001 {
		t.Fatalf("tank level 18 mp regen = %f, want 16.67", stats.MPRegen5)
	}
}

func TestArcherConfiguredStatsAtLevel18(t *testing.T) {
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("archer")
	if !ok {
		t.Fatal("archer hero not found")
	}
	stats := heroStatsAtLevel(hero, MaxHeroLevel)

	if hero.Resource != "mp" {
		t.Fatalf("archer resource = %s, want mp", hero.Resource)
	}
	if stats.MaxHP != 2357 {
		t.Fatalf("archer level 18 hp = %d, want 2357", stats.MaxHP)
	}
	if math.Abs(stats.MaxMP-824) > 0.000001 {
		t.Fatalf("archer level 18 mp = %f, want 824", stats.MaxMP)
	}
	if math.Abs(stats.Attack-109.32) > 0.000001 {
		t.Fatalf("archer level 18 attack = %f, want 109.32", stats.Attack)
	}
	if math.Abs(stats.AttackSpeed-1.030494) > 0.000001 {
		t.Fatalf("archer level 18 attack speed = %f, want 1.030494", stats.AttackSpeed)
	}
	if math.Abs(stats.PhysicalDefense-104.2) > 0.000001 {
		t.Fatalf("archer level 18 armor = %f, want 104.2", stats.PhysicalDefense)
	}
	if math.Abs(stats.MagicDefense-52.1) > 0.000001 {
		t.Fatalf("archer level 18 magic resist = %f, want 52.1", stats.MagicDefense)
	}
	if stats.MoveSpeed != 325 {
		t.Fatalf("archer move speed = %f, want 325", stats.MoveSpeed)
	}
	if stats.AttackRange != 600 {
		t.Fatalf("archer attack range = %f, want 600", stats.AttackRange)
	}
	if math.Abs(stats.HPRegen5-12.85) > 0.000001 {
		t.Fatalf("archer level 18 hp regen = %f, want 12.85", stats.HPRegen5)
	}
	if math.Abs(stats.MPRegen5-13.77) > 0.000001 {
		t.Fatalf("archer level 18 mp regen = %f, want 13.77", stats.MPRegen5)
	}
}

func TestArcherBasicAttackFiresArrowProjectile(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	hero.Base.Attack = 100
	hero.Base.CritChance = 0
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	id, ok := w.SpawnObject(EntityKindEnemyHero, TeamRed, player.Position.X+300, player.Position.Y)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}
	target := w.entities[id]
	target.Stats.PhysicalDefense = 0
	startHP := target.Stats.HP

	w.ApplyInput("archer", protocolPlayerInputAttack(id), 1, nil, 20)
	w.Tick(2, 20)

	if target.Stats.HP != startHP {
		t.Fatalf("archer arrow should not damage instantly: hp = %d, want %d", target.Stats.HP, startHP)
	}
	assertSkillEffect(t, w.SkillEffects(), "basic_arrow")

	for tick := uint64(3); tick < 20; tick++ {
		w.Tick(tick, 20)
		if target.Stats.HP < startHP {
			return
		}
	}
	t.Fatalf("archer arrow did not damage target after travel; hp = %d want below %d", target.Stats.HP, startHP)
}

func TestArcherQStacksOnBasicAttackHit(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerQSkillID, 1)
	target := &Entity{ID: "target", Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.addArcherFocusStack(player, 10, 20)
	w.addArcherFocusStack(player, 20, 20)

	if player.Archer.FocusStacks != 2 {
		t.Fatalf("focus stacks = %d, want 2", player.Archer.FocusStacks)
	}
	if player.Archer.FocusExpireTick != 100 {
		t.Fatalf("focus expire = %d, want 100", player.Archer.FocusExpireTick)
	}
	w.applyArcherFocusOnBasicHit(player, target, 30, 20)
	if player.Archer.FocusStacks != 3 {
		t.Fatalf("focus stacks after hit = %d, want 3", player.Archer.FocusStacks)
	}
}

func TestArcherQConsumesStacksAndGrantsAttackSpeed(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerQSkillID, 1)
	player.Archer.FocusStacks = 4
	player.Archer.FocusExpireTick = 100
	startMP := player.Stats.MP

	w.ApplyInput("archer", protocolPlayerInputCast(archerQSkillID, player.Position.X, player.Position.Y), 10, nil, 20)

	if player.Archer.FocusStacks != 0 {
		t.Fatalf("focus stacks = %d, want 0 after active", player.Archer.FocusStacks)
	}
	if player.Archer.FocusActiveUntil != 110 {
		t.Fatalf("focus active until = %d, want 110", player.Archer.FocusActiveUntil)
	}
	if math.Abs(player.Stats.MP-(startMP-50)) > 0.000001 {
		t.Fatalf("mp = %f, want %f", player.Stats.MP, startMP-50)
	}
	wantAttackSpeed := player.Stats.AttackSpeed * 1.2
	if math.Abs(EffectiveAttackSpeedAtTick(player, 11)-wantAttackSpeed) > 0.000001 {
		t.Fatalf("attack speed = %f, want %f", EffectiveAttackSpeedAtTick(player, 11), wantAttackSpeed)
	}
	w.addArcherFocusStack(player, 20, 20)
	if player.Archer.FocusStacks != 0 {
		t.Fatalf("focus stacks during active = %d, want 0", player.Archer.FocusStacks)
	}
}

func TestArcherQRequiresFullStacks(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerQSkillID, 1)
	player.Archer.FocusStacks = 3
	player.Archer.FocusExpireTick = 100
	startMP := player.Stats.MP

	w.ApplyInput("archer", protocolPlayerInputCast(archerQSkillID, player.Position.X, player.Position.Y), 10, nil, 20)

	if player.Archer.FocusActiveUntil != 0 {
		t.Fatalf("focus active until = %d, want 0 when not full stacks", player.Archer.FocusActiveUntil)
	}
	if player.Archer.FocusStacks != 3 {
		t.Fatalf("focus stacks = %d, want unchanged 3", player.Archer.FocusStacks)
	}
	if player.Stats.MP != startMP {
		t.Fatalf("mp = %f, want unchanged %f", player.Stats.MP, startMP)
	}
}

func TestArcherQActiveAddsBonusPhysicalDamage(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{
		ID:     "player:archer",
		HeroID: archerHeroID,
		Team:   TeamBlue,
		Kind:   EntityKindPlayer,
		Stats:  Stats{HP: 1000, MaxHP: 1000, Attack: 100},
		Archer: ArcherState{
			FocusActiveUntil:  100,
			FocusActiveLevel:  1,
			FocusBonusADRatio: 1.05,
		},
	}
	target := &Entity{
		ID:    "target",
		Team:  TeamRed,
		Stats: Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
	}

	damage := w.archerFocusBonusDamage(attacker, target, 20)

	if damage != 105 {
		t.Fatalf("focus bonus damage = %d, want 105", damage)
	}
}

func TestArcherWFiresVolleyArrowsConsumesManaAndCooldown(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerWSkillID, 1)
	startMP := player.Stats.MP

	w.ApplyInput("archer", protocolPlayerInputCast(archerWSkillID, player.Position.X+1000, player.Position.Y), 10, nil, 20)

	if len(w.projectiles) != 7 {
		t.Fatalf("volley projectiles = %d, want 7", len(w.projectiles))
	}
	if math.Abs(player.Stats.MP-(startMP-70)) > 0.000001 {
		t.Fatalf("mp = %f, want %f", player.Stats.MP, startMP-70)
	}
	if player.Skills[archerWSkillID].CooldownUntilTick != 290 {
		t.Fatalf("cooldown until = %d, want 290", player.Skills[archerWSkillID].CooldownUntilTick)
	}
}

func TestArcherWTargetTakesDamageOnceAndIsSlowed(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	hero.Base.Attack = 100
	hero.Base.CritChance = 0
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerWSkillID, 1)
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: player.Position.X + 300, Y: player.Position.Y},
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
		Radius:   80,
	}
	w.entities[target.ID] = target

	w.ApplyInput("archer", protocolPlayerInputCast(archerWSkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	for tick := uint64(11); tick < 30; tick++ {
		w.Tick(tick, 20)
	}

	if target.Stats.HP != 880 {
		t.Fatalf("target hp = %d, want 880 after one volley hit", target.Stats.HP)
	}
	if math.Abs(target.Control.MoveSpeedSlow-0.2) > 0.000001 {
		t.Fatalf("slow = %f, want 0.2", target.Control.MoveSpeedSlow)
	}
}

func TestArcherEStartsWithTwoChargesAndLaunchesHawk(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerESkillID, 1)
	w.refreshArcherSkillOnUpgrade(player, archerESkillID)

	if player.Skills[archerESkillID].Stacks != 2 {
		t.Fatalf("hawk charges = %d, want 2", player.Skills[archerESkillID].Stacks)
	}

	w.ApplyInput("archer", protocolPlayerInputCast(archerESkillID, player.Position.X+600, player.Position.Y), 10, nil, 20)

	state := player.Skills[archerESkillID]
	if state.Stacks != 1 {
		t.Fatalf("hawk charges after cast = %d, want 1", state.Stacks)
	}
	if state.CooldownUntilTick != 110 {
		t.Fatalf("hawk cooldown until = %d, want 110", state.CooldownUntilTick)
	}
	if state.StacksExpireTick != 1810 {
		t.Fatalf("hawk recharge tick = %d, want 1810", state.StacksExpireTick)
	}
	assertSkillEffect(t, w.SkillEffects(), "archer_hawk")
}

func TestArcherERechargeRestoresCharge(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerESkillID, 1)
	w.refreshArcherSkillOnUpgrade(player, archerESkillID)
	state := player.Skills[archerESkillID]
	state.Stacks = 1
	state.StacksExpireTick = 100
	player.Skills[archerESkillID] = state

	w.tickArcherHawkCharges(player, 100, 20)

	state = player.Skills[archerESkillID]
	if state.Stacks != 2 {
		t.Fatalf("hawk charges = %d, want 2", state.Stacks)
	}
	if state.StacksExpireTick != 0 {
		t.Fatalf("hawk recharge tick = %d, want 0 at full charges", state.StacksExpireTick)
	}
}

func TestArcherRFiresCrystalArrow(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerRSkillID, 1)
	startMP := player.Stats.MP

	w.ApplyInput("archer", protocolPlayerInputCast(archerRSkillID, player.Position.X+3000, player.Position.Y), 10, nil, 20)

	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles before cast delay = %d, want 0", len(w.projectiles))
	}
	if math.Abs(player.Stats.MP-(startMP-100)) > 0.000001 {
		t.Fatalf("mp = %f, want %f", player.Stats.MP, startMP-100)
	}
	if player.Skills[archerRSkillID].CooldownUntilTick != 2010 {
		t.Fatalf("r cooldown until = %d, want 2010", player.Skills[archerRSkillID].CooldownUntilTick)
	}
	if player.Control.ActionLockedUntilTick != 15 {
		t.Fatalf("cast delay lock until = %d, want 15", player.Control.ActionLockedUntilTick)
	}
	w.Tick(15, 20)
	if len(w.projectiles) != 1 {
		t.Fatalf("projectiles after cast delay = %d, want 1", len(w.projectiles))
	}
	assertSkillEffect(t, w.SkillEffects(), "archer_crystal_arrow")
}

func TestArcherRHitsFirstHeroStunsByTravelDistanceAndSplashes(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerRSkillID, 1)
	primary := &Entity{
		ID:       "primary",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: player.Position.X + 1500, Y: player.Position.Y},
		Stats:    Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
		Radius:   18,
	}
	splash := &Entity{
		ID:       "splash",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: primary.Position.X, Y: primary.Position.Y + 250},
		Stats:    Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
		Radius:   18,
	}
	w.entities[primary.ID] = primary
	w.entities[splash.ID] = splash

	w.ApplyInput("archer", protocolPlayerInputCast(archerRSkillID, primary.Position.X+500, primary.Position.Y), 10, nil, 20)
	for tick := uint64(11); tick < 50; tick++ {
		w.Tick(tick, 20)
		if primary.Combat.LastDamage > 0 {
			break
		}
	}

	if primary.Stats.HP != 800 {
		t.Fatalf("primary hp = %d, want 800", primary.Stats.HP)
	}
	if splash.Stats.HP != 900 {
		t.Fatalf("splash hp = %d, want 900", splash.Stats.HP)
	}
	if primary.Control.StunnedUntilTick-primary.Combat.LastHitTick != 70 {
		t.Fatalf("stun ticks = %d, want 70", primary.Control.StunnedUntilTick-primary.Combat.LastHitTick)
	}
	if math.Abs(splash.Control.MoveSpeedSlow-0.2) > 0.000001 {
		t.Fatalf("splash slow = %f, want 0.2", splash.Control.MoveSpeedSlow)
	}
}

func TestArcherRStunScalesBetweenZeroAnd1400Distance(t *testing.T) {
	skill := config.SkillConfig{
		Meta: map[string]float64{
			"minStunSeconds":  1,
			"maxStunSeconds":  3.5,
			"maxStunDistance": 1400,
		},
	}
	if got := archerRStunTicks(&Projectile{Traveled: 0}, skill, 20); got != 20 {
		t.Fatalf("stun at 0 = %d, want 20", got)
	}
	if got := archerRStunTicks(&Projectile{Traveled: 700}, skill, 20); got != 45 {
		t.Fatalf("stun at 700 = %d, want 45", got)
	}
	if got := archerRStunTicks(&Projectile{Traveled: 1400}, skill, 20); got != 70 {
		t.Fatalf("stun at 1400 = %d, want 70", got)
	}
	if got := archerRStunTicks(&Projectile{Traveled: 2000}, skill, 20); got != 70 {
		t.Fatalf("stun past 1400 = %d, want 70", got)
	}
}

func TestArcherRProjectileSpeedScalesWithTravelDistance(t *testing.T) {
	projectile := &Projectile{
		Range:    6000,
		SpeedMin: 1500,
		SpeedMax: 2100,
	}

	updateProjectileSpeed(projectile, 20)
	if projectile.SpeedPerTick != 75 {
		t.Fatalf("speed at start = %f, want 75", projectile.SpeedPerTick)
	}
	projectile.Traveled = 3000
	updateProjectileSpeed(projectile, 20)
	if projectile.SpeedPerTick != 90 {
		t.Fatalf("speed halfway = %f, want 90", projectile.SpeedPerTick)
	}
	projectile.Traveled = 6000
	updateProjectileSpeed(projectile, 20)
	if projectile.SpeedPerTick != 105 {
		t.Fatalf("speed at end = %f, want 105", projectile.SpeedPerTick)
	}
}

func TestTankGraniteShieldStartsWithMaxHealthShield(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}

	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]

	if player.Passive.MaxShield != 67 {
		t.Fatalf("tank max shield = %d, want 67", player.Passive.MaxShield)
	}
	if player.Passive.Shield != 67 {
		t.Fatalf("tank shield = %d, want 67", player.Passive.Shield)
	}
}

func TestTankGraniteShieldResetsAfterTenSecondsWithoutDamage(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	source := &Entity{Kind: EntityKindEnemyHero, Team: TeamRed}
	player.Combat.LastHitTick = 5

	w.applyDamage(source, player, 40, 20)

	if player.Stats.HP != player.Stats.MaxHP {
		t.Fatalf("tank hp = %d, want full hp %d while shield absorbs", player.Stats.HP, player.Stats.MaxHP)
	}
	if player.Passive.Shield != 27 {
		t.Fatalf("tank shield after damage = %d, want 27", player.Passive.Shield)
	}
	if player.Passive.LastRegenBreakTick != 5 {
		t.Fatalf("tank shield break tick = %d, want 5", player.Passive.LastRegenBreakTick)
	}

	w.Tick(204, 20)
	if player.Passive.Shield != 27 {
		t.Fatalf("tank shield before reset = %d, want 27", player.Passive.Shield)
	}
	w.Tick(205, 20)
	if player.Passive.Shield != player.Passive.MaxShield {
		t.Fatalf("tank shield after reset = %d, want %d", player.Passive.Shield, player.Passive.MaxShield)
	}
}

func TestTankQFiresShardDealsMagicDamageAndStealsMoveSpeed(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	hero.Base.AbilityPower = 100
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	nearby := w.entities["enemy:blue-hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	nearby.Team = TeamRed
	nearby.Position = target.Position
	target.Stats.MagicDefense = 0
	nearby.Stats.MagicDefense = 0
	startHP := target.Stats.HP
	startNearbyHP := nearby.Stats.HP
	startMP := player.Stats.MP

	w.ApplyInput("tank", protocolPlayerInputCastTarget(tankQSkillID, target.ID, target.Position.X, target.Position.Y), 10, w.skills, 20)

	if math.Abs(player.Stats.MP-(startMP-70)) > 0.000001 {
		t.Fatalf("tank mp = %f, want %f", player.Stats.MP, startMP-70)
	}
	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles during windup = %d, want 0", len(w.projectiles))
	}
	if !player.Tank.SeismicShardPending {
		t.Fatal("tank q should be pending during windup")
	}
	if player.Tank.SeismicShardReleaseTick != 15 {
		t.Fatalf("tank q release tick = %d, want 15", player.Tank.SeismicShardReleaseTick)
	}
	if player.Control.ActionLockedUntilTick != 15 {
		t.Fatalf("tank q action lock until = %d, want 15", player.Control.ActionLockedUntilTick)
	}
	w.Tick(15, 20)
	if len(w.projectiles) != 1 {
		t.Fatalf("projectiles = %d, want 1", len(w.projectiles))
	}
	for tick := uint64(16); tick <= 25; tick++ {
		w.Tick(tick, 20)
		if target.Combat.LastDamage > 0 {
			break
		}
	}
	if got := startHP - target.Stats.HP; got != 130 {
		t.Fatalf("tank q damage = %d, want 130", got)
	}
	if player.Control.MoveSpeedBonusUntil != target.Control.MoveSpeedSlowUntil {
		t.Fatalf("move speed effect until mismatch: source %d target %d", player.Control.MoveSpeedBonusUntil, target.Control.MoveSpeedSlowUntil)
	}
	if math.Abs(player.Control.MoveSpeedBonusFlat-0.84) > 0.000001 || target.Control.MoveSpeedSlow != 0.2 {
		t.Fatalf("move speed steal = %f/%f, want 0.84/0.2", player.Control.MoveSpeedBonusFlat, target.Control.MoveSpeedSlow)
	}
	hitTick := player.Control.MoveSpeedBonusUntil - 60
	if math.Abs(EffectiveMoveSpeedAtTick(player, hitTick)-335.84) > 0.000001 {
		t.Fatalf("tank stolen move speed = %f, want 335.84", EffectiveMoveSpeedAtTick(player, hitTick))
	}
	if math.Abs(EffectiveMoveSpeedAtTick(target, hitTick)-3.36) > 0.000001 {
		t.Fatalf("target slowed move speed = %f, want 3.36", EffectiveMoveSpeedAtTick(target, hitTick))
	}
	if nearby.Stats.HP != startNearbyHP {
		t.Fatalf("nearby hp = %d, want unchanged %d", nearby.Stats.HP, startNearbyHP)
	}
	if EffectiveMoveSpeedAtTick(player, player.Control.MoveSpeedBonusUntil) != player.Stats.MoveSpeed {
		t.Fatalf("tank move speed should expire to base")
	}
}

func TestTankQPicksNearestTargetToCursorAndTracksIt(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	other := w.entities["enemy:blue-hero-1"]
	other.Team = TeamRed
	other.Position = Vector2{X: player.Position.X + 260, Y: player.Position.Y}

	w.ApplyInput("tank", protocolPlayerInputCast(tankQSkillID, target.Position.X, target.Position.Y), 10, w.skills, 20)
	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles during windup = %d, want 0", len(w.projectiles))
	}
	w.Tick(15, 20)
	if len(w.projectiles) != 1 {
		t.Fatalf("projectiles = %d, want 1", len(w.projectiles))
	}
	var projectile *Projectile
	for _, value := range w.projectiles {
		projectile = value
	}
	if projectile.TargetID != target.ID {
		t.Fatalf("projectile target id = %q, want cursor-nearest %q", projectile.TargetID, target.ID)
	}
	target.Position.Y += 200
	w.Tick(16, 20)
	if projectile.Dir.Y <= 0 {
		t.Fatalf("tracking projectile dir y = %f, want positive toward moved target", projectile.Dir.Y)
	}
}

func TestTankWPassiveArmorScalesWithShield(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankWSkillID, 1)

	w.refreshTankWPassive(player)
	if math.Abs(player.Stats.PhysicalDefense-52) > 0.000001 {
		t.Fatalf("shielded tank armor = %f, want 52", player.Stats.PhysicalDefense)
	}
	player.Passive.Shield = 0
	w.refreshTankWPassive(player)
	if math.Abs(player.Stats.PhysicalDefense-44) > 0.000001 {
		t.Fatalf("unshielded tank armor = %f, want 44", player.Stats.PhysicalDefense)
	}
}

func TestTankWEmpowersAttackAndAftershocksCone(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	hero.Base.AttackRange = 300
	hero.Base.AbilityPower = 100
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankWSkillID, 1)
	player.Passive.Shield = 0
	w.refreshTankWPassive(player)
	target := w.entities["enemy:hero-1"]
	side := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	side.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	side.Position = Vector2{X: player.Position.X + 180, Y: player.Position.Y + 30}
	target.Stats.PhysicalDefense = 0
	side.Stats.PhysicalDefense = 0
	startTargetHP := target.Stats.HP
	startSideHP := side.Stats.HP
	startMP := player.Stats.MP

	w.ApplyInput("tank", protocolPlayerInputCast(tankWSkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)

	if math.Abs(player.Stats.MP-(startMP-30)) > 0.000001 {
		t.Fatalf("tank mp = %f, want %f", player.Stats.MP, startMP-30)
	}
	if player.Tank.ThunderclapAftershockUntil != 110 {
		t.Fatalf("aftershock until = %d, want 110", player.Tank.ThunderclapAftershockUntil)
	}

	w.ApplyInput("tank", protocolPlayerInputAttack(target.ID), 11, nil, 20)
	w.Tick(12, 20)

	if got := startTargetHP - target.Stats.HP; got != 171 {
		t.Fatalf("primary damage = %d, want 171", got)
	}
	if got := startSideHP - side.Stats.HP; got != 52 {
		t.Fatalf("aftershock damage = %d, want 52", got)
	}
	if target.Combat.LastDamage != 171 {
		t.Fatalf("primary last damage = %d, want combined 171", target.Combat.LastDamage)
	}
	if player.Tank.ThunderclapEmpowerUntil != 0 {
		t.Fatalf("empower until = %d, want consumed", player.Tank.ThunderclapEmpowerUntil)
	}
}

func TestTankEDealsMagicDamageAndSlowsAttackSpeed(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	hero.Base.AbilityPower = 100
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankESkillID, 1)
	target := w.entities["enemy:hero-1"]
	outside := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	outside.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	outside.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	target.Stats.MagicDefense = 0
	outside.Stats.MagicDefense = 0
	startHP := target.Stats.HP
	startOutsideHP := outside.Stats.HP
	startMP := player.Stats.MP

	w.ApplyInput("tank", protocolPlayerInputCast(tankESkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)

	if math.Abs(player.Stats.MP-(startMP-50)) > 0.000001 {
		t.Fatalf("tank mp = %f, want %f", player.Stats.MP, startMP-50)
	}
	if player.Skills[tankESkillID].CooldownUntilTick != 150 {
		t.Fatalf("tank e cooldown = %d, want 150", player.Skills[tankESkillID].CooldownUntilTick)
	}
	if got := startHP - target.Stats.HP; got != 0 {
		t.Fatalf("tank e damage before windup release = %d, want 0", got)
	}
	if !player.Tank.GroundSlamPending {
		t.Fatal("tank e should be pending during windup")
	}
	if player.Tank.GroundSlamReleaseTick != 15 {
		t.Fatalf("tank e release tick = %d, want 15", player.Tank.GroundSlamReleaseTick)
	}
	if player.Control.ActionLockedUntilTick != 15 {
		t.Fatalf("tank e action lock until = %d, want 15", player.Control.ActionLockedUntilTick)
	}
	w.Tick(15, 20)
	if got := startHP - target.Stats.HP; got != 136 {
		t.Fatalf("tank e damage = %d, want 136", got)
	}
	if outside.Stats.HP != startOutsideHP {
		t.Fatalf("outside hp = %d, want unchanged %d", outside.Stats.HP, startOutsideHP)
	}
	if target.Control.AttackSpeedSlow != 0.3 || target.Control.AttackSpeedSlowUntil != 75 {
		t.Fatalf("attack speed slow = %f until %d, want 0.3 until 75", target.Control.AttackSpeedSlow, target.Control.AttackSpeedSlowUntil)
	}
	if math.Abs(EffectiveAttackSpeedAtTick(target, 16)-0.7) > 0.000001 {
		t.Fatalf("slowed attack speed = %f, want 0.7", EffectiveAttackSpeedAtTick(target, 16))
	}
	if EffectiveAttackSpeedAtTick(target, 75) != target.Stats.AttackSpeed {
		t.Fatalf("attack speed should recover at expire tick")
	}
}

func TestTankRChargesTargetAreaDamagesLandingCircleOnly(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	hero.Base.AbilityPower = 100
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankRSkillID, 1)
	pathTarget := w.entities["enemy:hero-1"]
	landingTarget := w.entities["enemy:blue-hero-1"]
	outside := w.entities["dummy:training-1"]
	pathTarget.Team = TeamRed
	landingTarget.Team = TeamRed
	outside.Team = TeamRed
	pathTarget.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y + 40}
	landingPoint := Vector2{X: player.Position.X + 900, Y: player.Position.Y}
	landingTarget.Position = Vector2{X: landingPoint.X + 120, Y: landingPoint.Y}
	outside.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y + 250}
	pathTarget.Stats.MagicDefense = 0
	landingTarget.Stats.MagicDefense = 0
	outside.Stats.MagicDefense = 0
	startPathHP := pathTarget.Stats.HP
	startLandingHP := landingTarget.Stats.HP
	startOutsideHP := outside.Stats.HP
	startMP := player.Stats.MP

	w.ApplyInput("tank", protocolPlayerInputCast(tankRSkillID, landingPoint.X, landingPoint.Y), 10, w.skills, 20)

	if math.Abs(player.Stats.MP-(startMP-100)) > 0.000001 {
		t.Fatalf("tank mp = %f, want %f", player.Stats.MP, startMP-100)
	}
	if distance(player.Position, landingPoint) <= 0.000001 {
		t.Fatalf("tank should not arrive before dash ticks")
	}
	landingTick := player.Control.DashUntilTick
	if landingTick <= 10 {
		t.Fatalf("dash until = %d, want after cast tick", landingTick)
	}
	if player.Skills[tankRSkillID].CooldownUntilTick != 2610 {
		t.Fatalf("tank r cooldown = %d, want 2610", player.Skills[tankRSkillID].CooldownUntilTick)
	}
	if pathTarget.Stats.HP != startPathHP || landingTarget.Stats.HP != startLandingHP {
		t.Fatalf("tank r should not damage before arrival")
	}
	w.Tick(landingTick-1, 20)
	if landingTarget.Stats.HP != startLandingHP {
		t.Fatalf("landing target hp before arrival = %d, want %d", landingTarget.Stats.HP, startLandingHP)
	}
	w.Tick(landingTick, 20)
	if distance(player.Position, landingPoint) > 0.000001 {
		t.Fatalf("tank position = %+v, want landing %+v", player.Position, landingPoint)
	}
	assertSkillEffect(t, w.SkillEffects(), "tank_r_impact")
	if pathTarget.Stats.HP != startPathHP {
		t.Fatalf("path target hp = %d, want unchanged %d", pathTarget.Stats.HP, startPathHP)
	}
	if got := startLandingHP - landingTarget.Stats.HP; got != 280 {
		t.Fatalf("landing target damage = %d, want 280", got)
	}
	if outside.Stats.HP != startOutsideHP {
		t.Fatalf("outside hp = %d, want unchanged %d", outside.Stats.HP, startOutsideHP)
	}
	wantAirborneUntil := landingTick + secondsToTicks(1.5, 20)
	if pathTarget.Control.AirborneUntilTick != 0 || landingTarget.Control.AirborneUntilTick != wantAirborneUntil {
		t.Fatalf("airborne until = %d/%d, want 0/%d", pathTarget.Control.AirborneUntilTick, landingTarget.Control.AirborneUntilTick, wantAirborneUntil)
	}
}

func TestTankROutOfRangeMovesIntoCastRangeThenCasts(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankRSkillID, 1)
	start := player.Position
	targetPoint := Vector2{X: start.X + 1400, Y: start.Y}
	startMP := player.Stats.MP

	w.ApplyInput("tank", protocolPlayerInputCast(tankRSkillID, targetPoint.X, targetPoint.Y), 10, w.skills, 20)

	if !player.Tank.UnstoppableCastPending {
		t.Fatal("tank r cast should be pending out of range")
	}
	if math.Abs(player.Stats.MP-startMP) > 0.000001 {
		t.Fatalf("mp changed before cast = %f, want %f", player.Stats.MP, startMP)
	}
	if player.Skills[tankRSkillID].CooldownUntilTick != 0 {
		t.Fatalf("cooldown before real cast = %d, want 0", player.Skills[tankRSkillID].CooldownUntilTick)
	}

	for tick := uint64(11); tick <= 35; tick++ {
		w.Tick(tick, 20)
	}

	if player.Tank.UnstoppableCastPending {
		t.Fatal("tank r cast should release after moving into range")
	}
	if player.Control.DashUntilTick == 0 {
		t.Fatal("tank r should start dash after reaching cast range")
	}
	if math.Abs(player.Stats.MP-(startMP-100)) > 0.000001 {
		t.Fatalf("mp after real cast = %f, want %f", player.Stats.MP, startMP-100)
	}
	if player.Skills[tankRSkillID].CooldownUntilTick == 0 {
		t.Fatal("tank r cooldown should start after real cast")
	}
}

func TestTankROutOfRangeCastCanBeCanceledByMove(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(tankHeroID)
	if !ok {
		t.Fatal("tank hero not found")
	}
	w.SpawnHero("tank", hero, TeamBlue)
	player := w.entities[playerEntityID("tank")]
	learnSkill(player, tankRSkillID, 1)
	start := player.Position
	startMP := player.Stats.MP

	w.ApplyInput("tank", protocolPlayerInputCast(tankRSkillID, start.X+1400, start.Y), 10, w.skills, 20)
	w.ApplyInput("tank", protocolPlayerInputMove(start.X, start.Y+200), 11, w.skills, 20)

	if player.Tank.UnstoppableCastPending {
		t.Fatal("tank r pending cast should be canceled by move")
	}
	for tick := uint64(12); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}
	if player.Control.DashUntilTick != 0 {
		t.Fatalf("dash until = %d, want no r dash", player.Control.DashUntilTick)
	}
	if math.Abs(player.Stats.MP-startMP) > 0.000001 {
		t.Fatalf("mp after canceled cast = %f, want %f", player.Stats.MP, startMP)
	}
	if player.Skills[tankRSkillID].CooldownUntilTick != 0 {
		t.Fatalf("cooldown after canceled cast = %d, want 0", player.Skills[tankRSkillID].CooldownUntilTick)
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
	if player.SkillPoints != 2 {
		t.Fatalf("skill points = %d, want 2", player.SkillPoints)
	}
}

func TestUpgradeSkillUsesSkillPointAndCapsBySlot(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.SkillPoints = 10

	for i := 0; i < 6; i++ {
		w.ApplyInput("p1", protocolPlayerInputUpgrade("q"), 1, nil, 20)
	}
	for i := 0; i < 4; i++ {
		w.ApplyInput("p1", protocolPlayerInputUpgrade("r"), 1, nil, 20)
	}

	if player.Skills[swordQSkillID].Level != MaxBasicSkillLevel {
		t.Fatalf("q level = %d, want %d", player.Skills[swordQSkillID].Level, MaxBasicSkillLevel)
	}
	if player.Skills[swordRSkillID].Level != MaxUltSkillLevel {
		t.Fatalf("r level = %d, want %d", player.Skills[swordRSkillID].Level, MaxUltSkillLevel)
	}
	if player.SkillPoints != 2 {
		t.Fatalf("skill points = %d, want 2", player.SkillPoints)
	}
}

func TestUpgradeSkillWorksWhileActionLocked(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Control.ActionLockedUntilTick = 100

	w.ApplyInput("p1", protocolPlayerInputUpgrade("q"), 1, nil, 20)

	if player.Skills[swordQSkillID].Level != 1 {
		t.Fatalf("q level = %d, want 1", player.Skills[swordQSkillID].Level)
	}
	if player.SkillPoints != 0 {
		t.Fatalf("skill points = %d, want 0", player.SkillPoints)
	}
}

func TestDebugLevelUpRaisesHeroAndGrantsSkillPoint(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	w.ApplyInput("p1", protocolPlayerInputDebugLevelUp(), 1, nil, 20)

	if player.Level != 2 {
		t.Fatalf("level = %d, want 2", player.Level)
	}
	if player.SkillPoints != 2 {
		t.Fatalf("skill points = %d, want 2", player.SkillPoints)
	}
	player.Level = MaxHeroLevel
	player.SkillPoints = 0
	w.ApplyInput("p1", protocolPlayerInputDebugLevelUp(), 2, nil, 20)
	if player.Level != MaxHeroLevel {
		t.Fatalf("level = %d, want max %d", player.Level, MaxHeroLevel)
	}
	if player.SkillPoints != 0 {
		t.Fatalf("skill points = %d, want unchanged 0", player.SkillPoints)
	}
}

func TestUnlearnedSkillCannotCast(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := w.entities["dummy:training-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 1, nil, 20)

	if target.Combat.LastDamage != 0 {
		t.Fatal("unlearned q should not damage target")
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

func TestSpawnObjectEnemyHeroHasLevelRewardData(t *testing.T) {
	w := testWorld(t)
	id, ok := w.SpawnObject(EntityKindEnemyHero, TeamRed, 500, 600)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}

	entity := w.entities[id]
	if entity == nil {
		t.Fatalf("spawned enemy hero %s not found", id)
	}
	if entity.Level != MinHeroLevel {
		t.Fatalf("enemy hero level = %d, want %d", entity.Level, MinHeroLevel)
	}
	if entity.NextLevelExp != 280 {
		t.Fatalf("enemy hero next level exp = %f, want 280", entity.NextLevelExp)
	}
}

func TestEnemyHeroCanBeBasicAttacked(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.Attack = 100
	hero.Base.AttackRange = 300
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	id, ok := w.SpawnObject(EntityKindEnemyHero, TeamRed, player.Position.X+100, player.Position.Y)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}
	target := w.entities[id]
	target.Stats.PhysicalDefense = 0
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputAttack(id), 1, nil, 20)
	w.Tick(2, 20)

	if target.Stats.HP >= startHP {
		t.Fatalf("enemy hero hp = %d, want below %d after basic attack", target.Stats.HP, startHP)
	}
}

func TestEnemyHeroKillGrantsExperienceAndRemovesTarget(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.Attack = 2000
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	id, ok := w.SpawnObject(EntityKindEnemyHero, TeamRed, player.Position.X+100, player.Position.Y)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}

	w.ApplyInput("p1", protocolPlayerInputAttack(id), 1, nil, 20)
	w.Tick(2, 20)

	if player.TotalExp != 210 {
		t.Fatalf("total exp = %f, want 210", player.TotalExp)
	}
	if _, ok := w.entities[id]; ok {
		t.Fatalf("dead enemy hero %s should be removed", id)
	}
	if player.Intent.AttackTargetID != "" {
		t.Fatalf("attack target id = %q, want empty", player.Intent.AttackTargetID)
	}
}

func TestSwordPassiveChargesWhileMoving(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	w.ApplyInput("p1", protocolPlayerInputMove(player.Position.X+100, player.Position.Y), 1, nil, 20)
	w.Tick(2, 20)

	if player.Passive.SwordIntent <= 0 {
		t.Fatalf("sword intent did not charge while moving")
	}
	if player.Passive.MaxSwordIntent != 100 {
		t.Fatalf("max sword intent = %f, want 100", player.Passive.MaxSwordIntent)
	}
}

func TestSwordPassiveIntentChargeByLevelConfig(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	player.Level = 1
	w.chargeSwordIntent(player, 59)
	if math.Abs(player.Passive.SwordIntent-1) > 0.001 {
		t.Fatalf("level 1 sword intent = %f, want 1", player.Passive.SwordIntent)
	}

	player.Level = 7
	player.Passive.SwordIntent = 0
	w.chargeSwordIntent(player, 52.5)
	if math.Abs(player.Passive.SwordIntent-1) > 0.001 {
		t.Fatalf("level 7 sword intent = %f, want 1", player.Passive.SwordIntent)
	}

	player.Level = 13
	player.Passive.SwordIntent = 0
	w.chargeSwordIntent(player, 46)
	if math.Abs(player.Passive.SwordIntent-1) > 0.001 {
		t.Fatalf("level 13 sword intent = %f, want 1", player.Passive.SwordIntent)
	}
}

func TestSwordPassiveHeroDamageTriggersShield(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	source := w.entities["enemy:hero-1"]

	w.applyDamage(source, player, 150, 20)

	if player.Passive.SwordIntent != 0 {
		t.Fatalf("sword intent = %f, want 0 after shield trigger", player.Passive.SwordIntent)
	}
	if player.Passive.MaxShield != 125 {
		t.Fatalf("max shield = %d, want 125", player.Passive.MaxShield)
	}
	if player.Passive.Shield != 0 {
		t.Fatalf("shield = %d, want 0 after absorbing 150 damage", player.Passive.Shield)
	}
	if player.Stats.HP != player.Stats.MaxHP-25 {
		t.Fatalf("hp = %d, want %d", player.Stats.HP, player.Stats.MaxHP-25)
	}
}

func TestSwordPassiveShieldValueByLevelConfig(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	tests := []struct {
		level int
		want  int
	}{
		{level: 1, want: 125},
		{level: 6, want: 160},
		{level: 12, want: 275},
		{level: 18, want: 600},
	}
	for _, tt := range tests {
		player.Level = tt.level
		if got := w.swordShieldValue(player); got != tt.want {
			t.Fatalf("level %d shield = %d, want %d", tt.level, got, tt.want)
		}
	}
}

func TestSwordPassiveShieldExpiresAfterConfiguredDuration(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	source := w.entities["enemy:hero-1"]
	player.Combat.LastHitTick = 10

	w.applyDamage(source, player, 25, 20)
	if player.Passive.Shield != 100 {
		t.Fatalf("shield after absorb = %d, want 100", player.Passive.Shield)
	}
	if player.Passive.ShieldExpireTick != 30 {
		t.Fatalf("shield expire tick = %d, want 30", player.Passive.ShieldExpireTick)
	}

	w.Tick(29, 20)
	if player.Passive.Shield != 100 {
		t.Fatalf("shield before expire = %d, want 100", player.Passive.Shield)
	}
	w.Tick(30, 20)
	if player.Passive.Shield != 0 || player.Passive.MaxShield != 0 {
		t.Fatalf("shield after expire = %d/%d, want 0/0", player.Passive.Shield, player.Passive.MaxShield)
	}
}

func TestSwordPassiveMinionDamageDoesNotTriggerShield(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	source := w.entities["minion:red-melee-1"]

	w.applyDamage(source, player, 50, 20)

	if player.Passive.SwordIntent != player.Passive.MaxSwordIntent {
		t.Fatalf("sword intent = %f, want unchanged full intent", player.Passive.SwordIntent)
	}
	if player.Passive.Shield != 0 {
		t.Fatalf("shield = %d, want 0", player.Passive.Shield)
	}
	if player.Stats.HP != player.Stats.MaxHP-50 {
		t.Fatalf("hp = %d, want %d", player.Stats.HP, player.Stats.MaxHP-50)
	}
}

func TestSwordPassiveDoublesCritChance(t *testing.T) {
	attacker := &Entity{
		HeroID: swordHeroID,
		Stats:  Stats{CritChance: 0.4},
	}

	w := testWorld(t)
	if got := w.critChance(attacker); got != 0.8 {
		t.Fatalf("crit chance = %f, want 0.8", got)
	}
}

func TestSwordPassiveZeroCritChanceStaysZero(t *testing.T) {
	attacker := &Entity{
		HeroID: swordHeroID,
		Stats:  Stats{CritChance: 0},
	}

	w := testWorld(t)
	if got := w.critChance(attacker); got != 0 {
		t.Fatalf("crit chance = %f, want 0", got)
	}
}

func TestSwordPassiveCritDamageIsReducedTo190Percent(t *testing.T) {
	attacker := &Entity{
		ID:     "player:p1",
		HeroID: swordHeroID,
		Stats: Stats{
			Attack:     100,
			CritChance: 1,
		},
	}
	target := &Entity{
		ID:    "dummy:target",
		Stats: Stats{PhysicalDefense: 10},
	}

	w := testWorld(t)
	damage := w.attackDamage(attacker, target, 1)

	if damage != 173 {
		t.Fatalf("damage = %d, want 173", damage)
	}
}

func TestArcherPassiveAppliesFrostSlow(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{
		ID:     "player:archer",
		HeroID: archerHeroID,
		Level:  1,
		Skills: map[string]SkillState{
			"archer_focus": {SkillID: "archer_focus", Level: 1},
		},
		Stats: Stats{CritChance: 0},
	}
	target := &Entity{ID: "target", Stats: Stats{HP: 1000, MaxHP: 1000}}
	target.Combat.LastHitTick = 10

	w.applyDamage(attacker, target, 10, 20)

	if math.Abs(target.Control.MoveSpeedSlow-0.2) > 0.000001 {
		t.Fatalf("slow = %f, want 0.2", target.Control.MoveSpeedSlow)
	}
	if target.Control.MoveSpeedSlowUntil != 50 {
		t.Fatalf("slow until = %d, want 50", target.Control.MoveSpeedSlowUntil)
	}
}

func TestArcherPassiveCritDoublesSlowWithoutCritDamage(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{
		ID:     "player:archer",
		HeroID: archerHeroID,
		Level:  18,
		Skills: map[string]SkillState{
			"archer_focus": {SkillID: "archer_focus", Level: 1},
		},
		Stats: Stats{
			Attack:     100,
			CritChance: 1,
		},
	}
	target := &Entity{
		ID:    "target",
		Stats: Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
	}
	target.Combat.LastHitTick = 10

	damage := w.attackDamage(attacker, target, 10)
	w.applyDamage(attacker, target, damage, 20)

	if damage != 100 {
		t.Fatalf("archer crit basic damage = %d, want 100 without crit bonus", damage)
	}
	if math.Abs(target.Control.MoveSpeedSlow-0.6) > 0.000001 {
		t.Fatalf("crit slow = %f, want 0.6", target.Control.MoveSpeedSlow)
	}
}

func TestArcherPassiveDealsBonusDamageToSlowedTargets(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{
		ID:     "player:archer",
		HeroID: archerHeroID,
		Level:  1,
		Skills: map[string]SkillState{
			"archer_focus": {SkillID: "archer_focus", Level: 1},
		},
		Stats: Stats{
			Attack:     100,
			CritChance: 0,
		},
	}
	target := &Entity{
		ID:      "target",
		Stats:   Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
		Control: ControlState{MoveSpeedSlow: 0.2, MoveSpeedSlowUntil: 20},
	}

	damage := w.attackDamage(attacker, target, 10)

	if damage != 110 {
		t.Fatalf("damage against slowed target = %d, want 110", damage)
	}
}

func TestDamageAfterResistanceUsesPercentageReduction(t *testing.T) {
	tests := []struct {
		resistance float64
		want       int
	}{
		{resistance: 0, want: 100},
		{resistance: 50, want: 67},
		{resistance: 100, want: 50},
		{resistance: 200, want: 33},
		{resistance: 300, want: 25},
		{resistance: 400, want: 20},
		{resistance: -25, want: 133},
		{resistance: -50, want: 200},
	}
	for _, tt := range tests {
		if got := damageAfterResistance(100, tt.resistance, 0); got != tt.want {
			t.Fatalf("damage after %f resistance = %d, want %d", tt.resistance, got, tt.want)
		}
	}
}

func TestPenetrationAppliesPercentBeforeFlatAndCannotCreateNegativeResistance(t *testing.T) {
	effective := effectiveResistance(100, 0.3, 15)
	if effective != 55 {
		t.Fatalf("effective resistance = %f, want 55", effective)
	}
	if got := damageAfterResistance(100, effective, 0); got != 65 {
		t.Fatalf("damage after penetration = %d, want 65", got)
	}
	if effective := effectiveResistance(20, 0.5, 40); effective != 0 {
		t.Fatalf("penetration-created resistance = %f, want 0", effective)
	}
	if effective := effectiveResistance(-50, 0.3, 15); effective != -50 {
		t.Fatalf("forced negative resistance = %f, want -50", effective)
	}
}

func TestDamageReductionAppliesAfterResistance(t *testing.T) {
	if got := damageAfterResistance(100, 100, 0.6); got != 20 {
		t.Fatalf("damage after resistance and reduction = %d, want 20", got)
	}
}

func TestDamageReductionStacksMultiplicatively(t *testing.T) {
	reduction := stackDamageReduction(0.2, 0.5)
	if math.Abs(reduction-0.6) > 0.0001 {
		t.Fatalf("stacked reduction = %f, want 0.6", reduction)
	}
	if got := damageAfterResistance(100, 0, reduction); got != 40 {
		t.Fatalf("damage after stacked reduction = %d, want 40", got)
	}
}

func TestTrueDamageIgnoresResistanceAndUsesDamageReduction(t *testing.T) {
	target := &Entity{
		Stats: Stats{
			HP:              1000,
			MaxHP:           1000,
			PhysicalDefense: 400,
			MagicDefense:    400,
			DamageReduce:    0.25,
		},
	}
	w := testWorld(t)

	w.applyTrueDamage(nil, target, 100, 20)

	if target.Stats.HP != 925 {
		t.Fatalf("hp after true damage = %d, want 925", target.Stats.HP)
	}
}

func TestTrueDamageIsAbsorbedBySwordShield(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	player.Combat.LastHitTick = 10
	source := w.entities["enemy:hero-1"]

	w.applyTrueDamage(source, player, 100, 20)

	if player.Stats.HP != player.Stats.MaxHP {
		t.Fatalf("hp after shielded true damage = %d, want %d", player.Stats.HP, player.Stats.MaxHP)
	}
	if player.Passive.Shield != 25 {
		t.Fatalf("shield after true damage = %d, want 25", player.Passive.Shield)
	}
}

func TestSwordQDamagesTargetAndAddsStack(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	target := w.entities["dummy:training-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 1, nil, 20)

	if target.Combat.LastDamage != 0 {
		t.Fatalf("sword q damage before windup release = %d, want 0", target.Combat.LastDamage)
	}
	tickSwordQRelease(t, w, player, 20)

	state := player.Skills[swordQSkillID]
	if target.Combat.LastDamage <= 0 {
		t.Fatal("sword q should damage target")
	}
	if state.Stacks != 1 {
		t.Fatalf("q stacks = %d, want 1", state.Stacks)
	}
	if state.CooldownUntilTick <= 1 {
		t.Fatalf("q cooldown was not set")
	}
}

func TestSwordQCooldownUsesSkillLevelAndAttackSpeedPercent(t *testing.T) {
	w := testWorld(t)
	skill := w.skillConfig(swordQSkillID)

	if got := swordQCooldownTicksByBonus(0, skill, 1, 20); got != 120 {
		t.Fatalf("level 1 q cooldown ticks = %d, want 120", got)
	}
	if got := swordQCooldownTicksByBonus(1, skill, MaxBasicSkillLevel, 20); got != 32 {
		t.Fatalf("level 5 q cooldown ticks = %d, want 32", got)
	}
	if got := swordQCooldownTicksByBonus(10, skill, MaxBasicSkillLevel, 20); got != 27 {
		t.Fatalf("level 5 q min cooldown ticks = %d, want 27", got)
	}
}

func TestSwordQCooldownUsesAttackSpeedBonusNotPanelAttackSpeed(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Base.AttackSpeed = 0.697
	hero.Growth.AttackSpeed = 0
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.AttackSpeed = 0.697
	player.Stats.AttackSpeedBonus = 0

	if got := w.swordQCooldownTicks(player, w.skillConfig(swordQSkillID), 1, 20); got != 120 {
		t.Fatalf("base attack speed q cooldown ticks = %d, want 120", got)
	}
	player.Stats.AttackSpeed = 0.697 * 2
	player.Stats.AttackSpeedBonus = 1
	if got := w.swordQCooldownTicks(player, w.skillConfig(swordQSkillID), 1, 20); got != 48 {
		t.Fatalf("100%% bonus attack speed q cooldown ticks = %d, want 48", got)
	}
}

func TestSwordQWindupUsesAttackSpeedBonus(t *testing.T) {
	w := testWorld(t)
	skill := w.skillConfig(swordQSkillID)
	entity := &Entity{Stats: Stats{AttackSpeedBonus: 0}}

	if got := swordQWindupSeconds(entity, skill); math.Abs(got-0.328) > 0.000001 {
		t.Fatalf("base windup = %f, want 0.328", got)
	}
	entity.Stats.AttackSpeedBonus = 0.5
	want := 0.328 / 1.5
	if got := swordQWindupSeconds(entity, skill); math.Abs(got-want) > 0.000001 {
		t.Fatalf("50%% bonus attack speed windup = %f, want %f", got, want)
	}
	entity.Stats.AttackSpeedBonus = 10
	if got := swordQWindupSeconds(entity, skill); math.Abs(got-0.09) > 0.000001 {
		t.Fatalf("capped windup = %f, want 0.09", got)
	}
}

func TestFinalAttackSpeedUsesBonusRatioSlowAndCap(t *testing.T) {
	if got := finalAttackSpeed(0.65, 0.5, 0.7, 0); math.Abs(got-0.8775) > 0.000001 {
		t.Fatalf("attack speed = %f, want 0.8775", got)
	}
	if got := finalAttackSpeed(2, 1, 1, 0); got != 2.5 {
		t.Fatalf("attack speed cap = %f, want 2.5", got)
	}
	if got := finalAttackSpeed(1, 1, 1, 0.3); got != 1.4 {
		t.Fatalf("slowed attack speed = %f, want 1.4", got)
	}
	if got := finalAttackSpeed(1, 1, 0, 0); got != 1 {
		t.Fatalf("zero ratio attack speed = %f, want 1", got)
	}
}

func TestHeroConfigLoadsAttackWindups(t *testing.T) {
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]float64{
		swordHeroID:   0.328,
		tankHeroID:    0.25,
		warriorHeroID: 0.273,
		archerHeroID:  0.274,
	}
	for heroID, want := range cases {
		hero, ok := heroes.Get(heroID)
		if !ok {
			t.Fatalf("missing hero %s", heroID)
		}
		if math.Abs(hero.Base.AttackWindupSeconds-want) > 0.000001 {
			t.Fatalf("%s attack windup = %f, want %f", heroID, hero.Base.AttackWindupSeconds, want)
		}
	}
}

func TestAbilityHasteReducesSkillCooldowns(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = warriorHeroID
	hero.Skills.Q = warriorQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorQSkillID, 1)
	player.Stats.AbilityHaste = 100

	w.ApplyInput("p1", protocolPlayerInputCast(warriorQSkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)

	if got := player.Skills[warriorQSkillID].CooldownUntilTick; got != 90 {
		t.Fatalf("warrior q cooldown with 100 haste = %d, want 90", got)
	}
}

func TestSwordQIgnoresAbilityHaste(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	player.Stats.AbilityHaste = 100

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 10, w.skills, 20)

	want := uint64(10) + w.swordQCooldownTicks(player, w.skillConfig(swordQSkillID), 1, 20)
	if got := player.Skills[swordQSkillID].CooldownUntilTick; got != want {
		t.Fatalf("sword q cooldown with ability haste = %d, want %d", got, want)
	}
}

func TestBasicAttackUsesWindupBeforeDamage(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.AttackRange = 1000
	hero.Base.AttackSpeed = 1
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 100, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatalf("damage during attack windup = %d, want 0", target.Combat.LastDamage)
	}
	w.Tick(15, 20)
	if target.Combat.LastDamage <= 0 {
		t.Fatal("basic attack should damage after windup")
	}
	if player.Combat.NextAttackTick != 30 {
		t.Fatalf("next attack tick = %d, want 30", player.Combat.NextAttackTick)
	}
}

func TestBasicAttackWindupUsesAttackSpeedBonus(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.AttackRange = 1000
	hero.Base.AttackSpeed = 1
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.AttackSpeedBonus = 1
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 100, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	w.Tick(12, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatalf("damage before shortened windup = %d, want 0", target.Combat.LastDamage)
	}
	w.Tick(13, 20)
	if target.Combat.LastDamage <= 0 {
		t.Fatal("basic attack should damage after shortened windup")
	}
}

func TestRangedBasicAttackFiresProjectileAfterWindup(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}

	w.ApplyInput("archer", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles during attack windup = %d, want 0", len(w.projectiles))
	}
	w.Tick(16, 20)
	if len(w.projectiles) == 0 {
		t.Fatal("ranged basic attack should fire projectile after windup")
	}
}

func TestCastingSkillLocksAutoAttackForAttackInterval(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Base.AttackRange = 1000
	hero.Base.AttackSpeed = 1
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}
	player.Intent.AttackTargetID = target.ID

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 1, nil, 20)
	qDamageTick := target.Combat.LastHitTick
	w.Tick(1, 20)

	if target.Combat.LastHitTick != qDamageTick {
		t.Fatalf("auto attack should not fire on same tick as skill cast: got hit tick %d want %d", target.Combat.LastHitTick, qDamageTick)
	}
	tickSwordQRelease(t, w, player, 20)
	if player.Combat.NextAttackTick != 28 {
		t.Fatalf("next attack tick = %d, want 28", player.Combat.NextAttackTick)
	}
}

func TestSwordQDamagesAllEnemiesInMouseDirection(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	first := w.entities["dummy:training-1"]
	second := w.entities["dummy:training-2"]
	outside := w.entities["enemy:hero-1"]
	first.Position = Vector2{X: player.Position.X + 180, Y: player.Position.Y}
	second.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y + 20}
	outside.Position = Vector2{X: player.Position.X, Y: player.Position.Y + 300}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, first.Position.X, first.Position.Y), 1, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if first.Combat.LastDamage <= 0 {
		t.Fatal("first target should take sword q damage")
	}
	if second.Combat.LastDamage <= 0 {
		t.Fatal("second target should take sword q damage")
	}
	if outside.Combat.LastDamage != 0 {
		t.Fatalf("outside target damage = %d, want 0", outside.Combat.LastDamage)
	}
}

func TestSwordQThirdHitBecomesWhirlwindAndKnocksUp(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	for id, entity := range w.entities {
		if id != player.ID && id != target.ID && entity.Team != player.Team {
			delete(w.entities, id)
		}
	}
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 100, nil, 20)

	if target.Combat.LastDamage != 0 {
		t.Fatal("whirlwind q should not damage target before projectile reaches it")
	}
	hitTick := uint64(0)
	for tick := uint64(101); tick <= 140; tick++ {
		w.Tick(tick, 20)
		if target.Combat.LastDamage > 0 {
			hitTick = tick
			break
		}
	}
	if target.Combat.LastDamage <= 0 {
		t.Fatal("whirlwind q should damage target after projectile reaches it")
	}
	if target.Control.AirborneUntilTick != hitTick+20 {
		t.Fatalf("airborne until = %d, want %d", target.Control.AirborneUntilTick, hitTick+20)
	}
	state := player.Skills[swordQSkillID]
	if state.Stacks != 0 {
		t.Fatalf("q stacks = %d, want reset to 0", state.Stacks)
	}
}

func TestSwordWhirlwindQProjectileHitsOnlyAfterCollision(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	nearPath := w.entities["dummy:training-1"]
	nearWhirlwindEdge := w.entities["dummy:training-2"]
	outsideWhirlwind := w.entities["enemy:hero-1"]
	nearPath.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y + 20}
	nearWhirlwindEdge.Position = Vector2{X: player.Position.X + 600, Y: player.Position.Y + 20}
	outsideWhirlwind.Position = Vector2{X: player.Position.X + 600, Y: player.Position.Y + 180}
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+900, player.Position.Y), 100, nil, 20)

	if nearPath.Combat.LastDamage != 0 || nearWhirlwindEdge.Combat.LastDamage != 0 {
		t.Fatal("whirlwind q should not damage targets on cast")
	}
	for tick := uint64(101); tick <= 117; tick++ {
		w.Tick(tick, 20)
		if nearPath.Combat.LastDamage > 0 {
			break
		}
	}
	if nearPath.Combat.LastDamage <= 0 {
		t.Fatal("target near whirlwind path should take damage")
	}
	if nearWhirlwindEdge.Combat.LastDamage != 0 {
		t.Fatal("farther target should not be hit before projectile reaches it")
	}
	for tick := uint64(118); tick <= 126; tick++ {
		w.Tick(tick, 20)
		if nearWhirlwindEdge.Combat.LastDamage > 0 {
			break
		}
	}
	if nearWhirlwindEdge.Combat.LastDamage <= 0 {
		t.Fatal("target inside whirlwind radius should take damage")
	}
	if outsideWhirlwind.Combat.LastDamage != 0 {
		t.Fatalf("outside target damage = %d, want 0", outsideWhirlwind.Combat.LastDamage)
	}
}

func TestSwordWhirlwindQCanComboIntoLastBreathAfterHit(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	for id, entity := range w.entities {
		if id != player.ID && id != target.ID && entity.Team != player.Team {
			delete(w.entities, id)
		}
	}
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 100, nil, 20)
	hitTick := uint64(0)
	for tick := uint64(101); tick <= 130; tick++ {
		w.Tick(tick, 20)
		if target.Control.AirborneUntilTick > tick {
			hitTick = tick
			break
		}
	}
	if hitTick == 0 {
		t.Fatalf("target airborne until = %d, want airborne after whirlwind q hit", target.Control.AirborneUntilTick)
	}
	rTick := hitTick + 1
	if target.Control.AirborneUntilTick <= rTick {
		t.Fatalf("target airborne until = %d, want after r cast tick", target.Control.AirborneUntilTick)
	}

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), rTick, nil, 20)

	if player.Skills[swordRSkillID].CooldownUntilTick == 0 {
		t.Fatal("r should cast on target knocked up by whirlwind q")
	}
	if target.Combat.LastDamage <= 0 {
		t.Fatal("r should damage airborne target")
	}
}

func TestSwordWCreatesWindWallAndExpires(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.W = swordWSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordWSkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(swordWSkillID, player.Position.X+100, player.Position.Y), 10, nil, 20)

	walls := w.WindWalls()
	if len(walls) != 1 {
		t.Fatalf("wind wall count = %d, want 1", len(walls))
	}
	if walls[0].Width != 300 {
		t.Fatalf("wind wall width = %f, want 300", walls[0].Width)
	}
	if walls[0].ExpiresAt != 90 {
		t.Fatalf("wind wall expires at = %d, want 90", walls[0].ExpiresAt)
	}
	state := player.Skills[swordWSkillID]
	if state.CooldownUntilTick != 530 {
		t.Fatalf("wind wall cooldown until = %d, want 530", state.CooldownUntilTick)
	}

	w.Tick(90, 20)

	if got := len(w.WindWalls()); got != 0 {
		t.Fatalf("wind wall count after expiry = %d, want 0", got)
	}
}

func TestWarriorToughnessRegeneratesAfterOutOfCombat(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.HP = 500
	player.Passive.LastRegenBreakTick = 0

	w.Tick(160, 20)

	if player.Stats.HP != 509 {
		t.Fatalf("hp after toughness regen = %d, want 509", player.Stats.HP)
	}
}

func TestWarriorToughnessRegenRatioUsesLevelTable(t *testing.T) {
	w := testWorld(t)
	skill := w.skillConfig("warrior_toughness")

	if got := warriorToughnessRegenRatio(1, skill); got != 0.015 {
		t.Fatalf("level 1 toughness ratio = %f, want 0.015", got)
	}
	if got := warriorToughnessRegenRatio(10, skill); got != 0.0582 {
		t.Fatalf("level 10 toughness ratio = %f, want 0.0582", got)
	}
	if got := warriorToughnessRegenRatio(18, skill); got != 0.101 {
		t.Fatalf("level 18 toughness ratio = %f, want 0.101", got)
	}
}

func TestWarriorToughnessHeroDamageBreaksRegen(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.HP = 500
	source := w.entities["enemy:hero-1"]
	player.Combat.LastHitTick = 100

	w.applyDamage(source, player, 10, 20)
	w.Tick(259, 20)

	if player.Stats.HP != 490 {
		t.Fatalf("hp before out-of-combat = %d, want 490", player.Stats.HP)
	}
	w.Tick(260, 20)
	if player.Stats.HP != 499 {
		t.Fatalf("hp after out-of-combat = %d, want 499", player.Stats.HP)
	}
}

func TestWarriorToughnessMinionDamageDoesNotBreakRegen(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.HP = 500
	source := w.entities["minion:red-melee-1"]
	player.Combat.LastHitTick = 100

	w.applyDamage(source, player, 10, 20)
	w.Tick(160, 20)

	if player.Stats.HP != 499 {
		t.Fatalf("hp after minion damage toughness regen = %d, want 499", player.Stats.HP)
	}
}

func TestWarriorQEmpowersNextAttack(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300 + player.Radius + target.Radius, Y: player.Position.Y}
	startHP := target.Stats.HP
	player.Combat.NextAttackTick = 999

	w.ApplyInput("p1", protocolPlayerInputCast(warriorQSkillID, target.Position.X, target.Position.Y), 10, w.skills, 20)

	if player.Skills[warriorQSkillID].CooldownUntilTick != 170 {
		t.Fatalf("warrior q cooldown = %d, want 170", player.Skills[warriorQSkillID].CooldownUntilTick)
	}
	if player.Warrior.DecisiveStrikeUntilTick != 100 {
		t.Fatalf("warrior q empower until = %d, want 100", player.Warrior.DecisiveStrikeUntilTick)
	}
	if player.Warrior.DecisiveStrikeSpeedUntilTick != 40 {
		t.Fatalf("warrior q speed until = %d, want 40", player.Warrior.DecisiveStrikeSpeedUntilTick)
	}
	if got := EffectiveMoveSpeedAtTick(player, 11); math.Abs(got-442) > 0.000001 {
		t.Fatalf("warrior q effective move speed = %f, want 442", got)
	}
	if got := EffectiveMoveSpeedAtTick(player, 40); got != 340 {
		t.Fatalf("expired warrior q move speed = %f, want 340", got)
	}
	if player.Combat.NextAttackTick != 10 {
		t.Fatalf("next attack tick = %d, want 10 reset", player.Combat.NextAttackTick)
	}

	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 11, nil, 20)
	w.Tick(12, 20)

	if got := startHP - target.Stats.HP; got != 146 {
		t.Fatalf("warrior q damage = %d, want 146", got)
	}
	if target.Control.SilencedUntilTick != 42 {
		t.Fatalf("silenced until = %d, want 42", target.Control.SilencedUntilTick)
	}
	if player.Warrior.DecisiveStrikeUntilTick != 0 {
		t.Fatalf("warrior q empower should be consumed, got %d", player.Warrior.DecisiveStrikeUntilTick)
	}
}

func TestWarriorQExpiresWithoutEmpoweringAttack(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300 + player.Radius + target.Radius, Y: player.Position.Y}
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(warriorQSkillID, target.Position.X, target.Position.Y), 10, w.skills, 20)
	player.Combat.NextAttackTick = 101
	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 101, nil, 20)
	w.Tick(101, 20)

	if got := startHP - target.Stats.HP; got != 0 {
		t.Fatalf("expired warrior q should not reach or damage target, got %d", got)
	}
	if target.Control.SilencedUntilTick != 0 {
		t.Fatalf("expired warrior q silence = %d, want 0", target.Control.SilencedUntilTick)
	}
}

func TestWarriorWActiveGrantsShieldDamageReductionAndTenacity(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorWSkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(warriorWSkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)

	if player.Passive.Shield != 70 {
		t.Fatalf("warrior w shield = %d, want 70", player.Passive.Shield)
	}
	if player.Warrior.CourageUntilTick != 90 {
		t.Fatalf("courage until = %d, want 90", player.Warrior.CourageUntilTick)
	}
	if player.Warrior.CourageFrontUntilTick != 25 {
		t.Fatalf("courage front until = %d, want 25", player.Warrior.CourageFrontUntilTick)
	}
	if player.Control.TenacityUntilTick != 25 {
		t.Fatalf("tenacity until = %d, want 25", player.Control.TenacityUntilTick)
	}
	if player.Skills[warriorWSkillID].CooldownUntilTick != 490 {
		t.Fatalf("w cooldown = %d, want 490", player.Skills[warriorWSkillID].CooldownUntilTick)
	}

	frontDamage := damageAfterResistance(100, 0, player.damageReductionForType("physical", 12))
	if frontDamage != 40 {
		t.Fatalf("front damage = %d, want 40", frontDamage)
	}
	backDamage := damageAfterResistance(100, 0, player.damageReductionForType("physical", 30))
	if backDamage != 70 {
		t.Fatalf("back damage = %d, want 70", backDamage)
	}
	expiredDamage := damageAfterResistance(100, 0, player.damageReductionForType("physical", 90))
	if expiredDamage != 100 {
		t.Fatalf("expired damage = %d, want 100", expiredDamage)
	}
}

func TestWarriorWFrontTenacityReducesControlDuration(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorWSkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(warriorWSkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)

	controlTicks := secondsToTicks(1.5, 20)
	if got := controlTicksAfterTenacity(player, controlTicks, 12); got != 12 {
		t.Fatalf("front tenacity control ticks = %d, want 12", got)
	}
	if got := controlTicksAfterTenacity(player, controlTicks, 25); got != 30 {
		t.Fatalf("expired front tenacity control ticks = %d, want 30", got)
	}
}

func TestTenacityStacksMultiplicatively(t *testing.T) {
	if got := stackTenacity(0.3, 0.6); math.Abs(got-0.72) > 0.0001 {
		t.Fatalf("stacked tenacity = %f, want 0.72", got)
	}
}

func TestWarriorESpinsDamageNearestAndShredsArmor(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorESkillID, 1)
	near := w.entities["enemy:hero-1"]
	far := w.entities["enemy:blue-hero-1"]
	near.Team = TeamRed
	far.Team = TeamRed
	near.Position = Vector2{X: player.Position.X + 80, Y: player.Position.Y}
	far.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	near.Stats.PhysicalDefense = 0
	far.Stats.PhysicalDefense = 0
	nearStartHP := near.Stats.HP
	farStartHP := far.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(warriorESkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)
	for tick := uint64(10); tick <= 70; tick++ {
		w.Tick(tick, 20)
	}

	if got := nearStartHP - near.Stats.HP; got != 259 {
		t.Fatalf("near target damage = %d, want 259", got)
	}
	if got := farStartHP - far.Stats.HP; got != 210 {
		t.Fatalf("far target damage = %d, want 210", got)
	}
	if near.Combat.PhysicalDefenseShredUntil != 170 {
		t.Fatalf("near shred until = %d, want 170", near.Combat.PhysicalDefenseShredUntil)
	}
}

func TestWarriorESecondCastEndsEarlyAndRefundsRemainingDuration(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorESkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(warriorESkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)
	if player.Skills[warriorESkillID].CooldownUntilTick != 0 {
		t.Fatalf("e cooldown until while active = %d, want 0", player.Skills[warriorESkillID].CooldownUntilTick)
	}

	w.ApplyInput("p1", protocolPlayerInputCast(warriorESkillID, player.Position.X, player.Position.Y), 30, w.skills, 20)

	if player.Warrior.JudgmentUntilTick != 0 {
		t.Fatalf("judgment until = %d, want stopped", player.Warrior.JudgmentUntilTick)
	}
	if player.Skills[warriorESkillID].CooldownUntilTick != 170 {
		t.Fatalf("refunded cooldown until = %d, want 170", player.Skills[warriorESkillID].CooldownUntilTick)
	}
}

func TestWarriorECooldownStartsAfterSpinEnds(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorESkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(warriorESkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)
	for tick := uint64(10); tick <= 70; tick++ {
		w.Tick(tick, 20)
		if tick < 70 && player.Skills[warriorESkillID].CooldownUntilTick != 0 {
			t.Fatalf("e cooldown during spin at tick %d = %d, want 0", tick, player.Skills[warriorESkillID].CooldownUntilTick)
		}
	}

	if player.Skills[warriorESkillID].CooldownUntilTick != 250 {
		t.Fatalf("e cooldown after spin = %d, want 250", player.Skills[warriorESkillID].CooldownUntilTick)
	}
}

func TestWarriorEAttackSpeedAddsSpins(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.AttackSpeedBonus = 0.5
	learnSkill(player, warriorESkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(warriorESkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)

	if player.Warrior.JudgmentSpinsRemaining != 9 {
		t.Fatalf("judgment spins = %d, want 9", player.Warrior.JudgmentSpinsRemaining)
	}
}

func TestWarriorECannotAutoAttackWhileSpinning(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 100, Y: player.Position.Y}
	target.Stats.PhysicalDefense = 0
	player.Combat.NextAttackTick = 10
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(warriorESkillID, player.Position.X, player.Position.Y), 10, w.skills, 20)
	w.Tick(10, 20)
	hpAfterFirstSpin := target.Stats.HP
	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 11, nil, 20)
	w.Tick(11, 20)

	if target.Stats.HP != hpAfterFirstSpin {
		t.Fatalf("target hp = %d, want unchanged %d from blocked auto attack", target.Stats.HP, hpAfterFirstSpin)
	}
	if startHP-hpAfterFirstSpin <= 0 {
		t.Fatal("judgment spin should still damage target")
	}
}

func TestWarriorRDealsMissingHealthTrueDamage(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 250, Y: player.Position.Y}
	target.Stats.MaxHP = 1000
	target.Stats.HP = 600
	target.Stats.DamageReduce = 0
	target.Stats.PhysicalDefense = 999

	w.ApplyInput("p1", protocolPlayerInputCast(warriorRSkillID, target.Position.X, target.Position.Y), 10, w.skills, 20)

	if target.Stats.HP != 600 {
		t.Fatalf("target hp before windup release = %d, want 600", target.Stats.HP)
	}
	if !player.Warrior.JusticePending {
		t.Fatal("warrior r should be pending during windup")
	}
	if player.Warrior.JusticeReleaseTick != 19 {
		t.Fatalf("warrior r release tick = %d, want 19", player.Warrior.JusticeReleaseTick)
	}
	if player.Control.ActionLockedUntilTick != 19 {
		t.Fatalf("warrior r action lock until = %d, want 19", player.Control.ActionLockedUntilTick)
	}
	w.Tick(19, 20)

	if target.Stats.HP != 350 {
		t.Fatalf("target hp = %d, want 350", target.Stats.HP)
	}
	if target.Combat.LastDamage != 250 {
		t.Fatalf("last damage = %d, want 250", target.Combat.LastDamage)
	}
	if player.Skills[warriorRSkillID].CooldownUntilTick != 2410 {
		t.Fatalf("r cooldown until = %d, want 2410", player.Skills[warriorRSkillID].CooldownUntilTick)
	}
}

func TestWarriorRWindupIsFixedAndIgnoresAttackSpeed(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorRSkillID, 1)
	player.Stats.AttackSpeedBonus = 10
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 250, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(warriorRSkillID, target.Position.X, target.Position.Y), 10, w.skills, 20)

	if player.Warrior.JusticeReleaseTick != 19 {
		t.Fatalf("warrior r release tick = %d, want fixed 19", player.Warrior.JusticeReleaseTick)
	}
}

func TestWarriorRDamageScalesByRank(t *testing.T) {
	skill := config.SkillConfig{
		MetaLists: map[string][]float64{
			"baseDamage":     {150, 250, 350},
			"missingHPRatio": {0.25, 0.3, 0.35},
		},
	}
	target := &Entity{Stats: Stats{MaxHP: 2000, HP: 1000}}

	if got := warriorRDamage(target, skill, 3); got != 700 {
		t.Fatalf("rank 3 damage = %f, want 700", got)
	}
}

func TestWarriorRRequiresTargetInRange(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 900, Y: player.Position.Y}
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(warriorRSkillID, target.Position.X, target.Position.Y), 10, w.skills, 20)

	if target.Stats.HP != startHP {
		t.Fatalf("target hp = %d, want unchanged %d", target.Stats.HP, startHP)
	}
	if player.Skills[warriorRSkillID].CooldownUntilTick != 0 {
		t.Fatalf("r cooldown until = %d, want 0 for invalid target", player.Skills[warriorRSkillID].CooldownUntilTick)
	}
}

func TestWarriorWPassiveKillGrantsPermanentResists(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, warriorWSkillID, 1)
	baseArmor := player.Stats.PhysicalDefense
	baseMagic := player.Stats.MagicDefense

	w.applyWarriorWPassiveKill(player, &Entity{Kind: EntityKindMeleeMinion})
	if math.Abs(player.Stats.PhysicalDefense-(baseArmor+0.2)) > 0.0001 {
		t.Fatalf("armor after minion kill = %f, want %f", player.Stats.PhysicalDefense, baseArmor+0.2)
	}
	if math.Abs(player.Stats.MagicDefense-(baseMagic+0.2)) > 0.0001 {
		t.Fatalf("magic resist after minion kill = %f, want %f", player.Stats.MagicDefense, baseMagic+0.2)
	}

	w.applyWarriorWPassiveKill(player, &Entity{Kind: EntityKindEnemyHero})
	if math.Abs(player.Warrior.CouragePassiveResistGain-1.2) > 0.0001 {
		t.Fatalf("passive resist gain = %f, want 1.2", player.Warrior.CouragePassiveResistGain)
	}
}

func TestWarriorWPassiveRequiresLearnedSkill(t *testing.T) {
	w := testWorld(t)
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	baseArmor := player.Stats.PhysicalDefense
	baseMagic := player.Stats.MagicDefense

	w.applyWarriorWPassiveKill(player, &Entity{Kind: EntityKindMeleeMinion})

	if player.Stats.PhysicalDefense != baseArmor {
		t.Fatalf("armor = %f, want unchanged %f", player.Stats.PhysicalDefense, baseArmor)
	}
	if player.Stats.MagicDefense != baseMagic {
		t.Fatalf("magic resist = %f, want unchanged %f", player.Stats.MagicDefense, baseMagic)
	}
}

func TestWindWallBlocksEnemyProjectileOnly(t *testing.T) {
	w := testWorld(t)
	w.windWalls["wall"] = WindWall{
		ID:        "wall",
		Team:      TeamBlue,
		Center:    Vector2{X: 100, Y: 100},
		Dir:       Vector2{X: 0, Y: 1},
		Width:     300,
		ExpiresAt: 100,
	}

	if !w.BlocksProjectile(TeamRed, Vector2{X: 0, Y: 100}, Vector2{X: 200, Y: 100}) {
		t.Fatal("enemy projectile should be blocked")
	}
	if w.BlocksProjectile(TeamBlue, Vector2{X: 0, Y: 100}, Vector2{X: 200, Y: 100}) {
		t.Fatal("same-team projectile should not be blocked")
	}
}

func TestSwordEDashesThroughTargetAndAppliesPerTargetCooldown(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	hero.Base.BonusAttack = 40
	hero.Base.AbilityPower = 50
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}
	startPosition := player.Position
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)

	if target.Stats.HP >= startHP {
		t.Fatal("sword e should damage target")
	}
	if player.Position != startPosition {
		t.Fatalf("player position = %+v, want unchanged at cast tick %+v", player.Position, startPosition)
	}
	w.Tick(13, 20)
	if player.Position.X <= startPosition.X || player.Position.X >= target.Position.X {
		t.Fatalf("player x = %f, want moving toward target from %f", player.Position.X, startPosition.X)
	}
	w.Tick(17, 20)
	if player.Position.X <= target.Position.X {
		t.Fatalf("player x = %f, should finish dash through target x=%f", player.Position.X, target.Position.X)
	}
	if player.Passive.SwordIntent <= 0 {
		t.Fatal("sword e movement should charge sword intent")
	}
	if player.Sword.SweepingBladeStacks != 1 {
		t.Fatalf("e stacks = %d, want 1", player.Sword.SweepingBladeStacks)
	}
	afterFirstHP := target.Stats.HP
	player.Skills[swordESkillID] = SkillState{SkillID: swordESkillID, Level: 1}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 20, nil, 20)

	if target.Stats.HP != afterFirstHP {
		t.Fatalf("target hp = %d, want unchanged %d while per-target cooldown active", target.Stats.HP, afterFirstHP)
	}
}

func TestSwordEPicksUnitNearestToCursorPoint(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	nearPlayer := w.entities["enemy:hero-1"]
	nearCursor := w.entities["enemy:blue-hero-1"]
	nearPlayer.Team = TeamRed
	nearCursor.Team = TeamRed
	nearPlayer.Position = Vector2{X: player.Position.X + 160, Y: player.Position.Y}
	nearCursor.Position = Vector2{X: player.Position.X + 260, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, nearCursor.Position.X, nearCursor.Position.Y), 10, nil, 20)

	if nearCursor.Combat.LastDamage <= 0 {
		t.Fatal("sword e should hit unit nearest to cursor point")
	}
	if nearPlayer.Combat.LastDamage != 0 {
		t.Fatalf("nearer-to-player unit damage = %d, want 0", nearPlayer.Combat.LastDamage)
	}
}

func TestSwordEStoresPerTargetCooldown(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)

	if player.Sword.SweepingBladeTargetUntil[target.ID] != 210 {
		t.Fatalf("target cooldown until = %d, want 210", player.Sword.SweepingBladeTargetUntil[target.ID])
	}
}

func TestSwordEIgnoresAttackInputDuringDash(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 11, nil, 20)

	if player.Intent.AttackTargetID != "" {
		t.Fatalf("attack target during e dash = %q, want empty", player.Intent.AttackTargetID)
	}
}

func TestSwordEQMakesCircularQ(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	sideTarget := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	sideTarget.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	sideTarget.Position = Vector2{X: target.Position.X, Y: target.Position.Y + 120}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.Tick(14, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 14, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if sideTarget.Combat.LastDamage <= 0 {
		t.Fatal("q during e dash should become circular aoe and hit nearby target")
	}
}

func TestSwordEQDamageUsesReleasePositionAfterDashMovement(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	dashTarget := w.entities["enemy:hero-1"]
	oldPositionTarget := w.entities["enemy:blue-hero-1"]
	newPositionTarget := &Entity{
		ID:       "enemy:eq-release-position",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		HeroID:   warriorHeroID,
		Position: Vector2{X: player.Position.X + 680, Y: player.Position.Y},
		Radius:   30,
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0, MagicDefense: 0},
		Skills:   make(map[string]SkillState),
	}
	w.entities[newPositionTarget.ID] = newPositionTarget
	dashTarget.Team = TeamRed
	oldPositionTarget.Team = TeamRed
	dashTarget.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	oldPositionTarget.Position = Vector2{X: player.Position.X, Y: player.Position.Y + 450}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, dashTarget.Position.X, dashTarget.Position.Y), 10, nil, 20)
	w.Tick(14, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 14, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if oldPositionTarget.Combat.LastDamage != 0 {
		t.Fatalf("old position target damage = %d, want 0", oldPositionTarget.Combat.LastDamage)
	}
	if newPositionTarget.Combat.LastDamage <= 0 {
		t.Fatal("new position target should be hit by EQ at release position")
	}
}

func TestSwordQBeforeEQWindowDoesNotBecomeCircular(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	sideTarget := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	sideTarget.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	sideTarget.Position = Vector2{X: target.Position.X, Y: target.Position.Y + 120}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 11, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if sideTarget.Combat.LastDamage != 0 {
		t.Fatalf("side target damage = %d, want 0 before eq window", sideTarget.Combat.LastDamage)
	}
}

func TestSwordEQWithWhirlwindStacksKnocksUpAndClearsStacks(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	sideTarget := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	sideTarget.Team = TeamRed
	sideTarget.HeroID = warriorHeroID
	sideTarget.Warrior.CourageFrontUntilTick = 30
	sideTarget.Warrior.CourageFrontTenacity = 0.6
	sideTarget.Control.TenacityUntilTick = 30
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	sideTarget.Position = Vector2{X: target.Position.X, Y: target.Position.Y + 120}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.Tick(14, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 14, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if sideTarget.Control.AirborneUntilTick != 41 {
		t.Fatalf("side target airborne until = %d, want 41", sideTarget.Control.AirborneUntilTick)
	}
	qState := player.Skills[swordQSkillID]
	if qState.Stacks != 0 {
		t.Fatalf("q stacks = %d, want 0", qState.Stacks)
	}
	if qState.StacksExpireTick != 0 {
		t.Fatalf("q stacks expire tick = %d, want 0", qState.StacksExpireTick)
	}
}

func TestSwordRRequiresAirborneHeroAndAppliesLastBreathState(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	hero.Base.BonusAttack = 60
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	target.Control.AirborneUntilTick = 120
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), 100, nil, 20)

	if target.Stats.HP >= startHP {
		t.Fatal("sword r should damage airborne enemy hero")
	}
	if target.Control.AirborneUntilTick != 140 {
		t.Fatalf("airborne until = %d, want extended to 140", target.Control.AirborneUntilTick)
	}
	if player.Passive.Shield != w.swordShieldValue(player) {
		t.Fatalf("shield = %d, want %d", player.Passive.Shield, w.swordShieldValue(player))
	}
	qState := player.Skills[swordQSkillID]
	if qState.Stacks != 0 {
		t.Fatalf("q stacks = %d, want 0", qState.Stacks)
	}
	if qState.StacksExpireTick != 0 {
		t.Fatalf("q stacks expire tick = %d, want 0", qState.StacksExpireTick)
	}
	if player.Sword.LastBreathUntilTick != 400 {
		t.Fatalf("last breath until = %d, want 400", player.Sword.LastBreathUntilTick)
	}
	if player.Skills[swordRSkillID].CooldownUntilTick != 1700 {
		t.Fatalf("r cooldown until = %d, want 1700", player.Skills[swordRSkillID].CooldownUntilTick)
	}
}

func TestSwordRAutoTargetsAirborneHeroInRange(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	target.Control.AirborneUntilTick = 130

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, player.Position.X+1200, player.Position.Y+900), 110, nil, 20)

	if player.Skills[swordRSkillID].CooldownUntilTick == 0 {
		t.Fatal("r should auto target airborne enemy hero in range")
	}
}

func TestSwordRLocksSelfActionsForOneSecond(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	target.Control.AirborneUntilTick = 130

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), 100, nil, 20)
	if player.Control.ActionLockedUntilTick != 120 {
		t.Fatalf("action locked until = %d, want 120", player.Control.ActionLockedUntilTick)
	}
	start := player.Position
	player.Intent.AttackTargetID = target.ID
	nextAttack := player.Combat.NextAttackTick

	w.ApplyInput("p1", protocolPlayerInputMove(start.X+500, start.Y), 101, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 101, nil, 20)
	w.Tick(101, 20)

	if player.Position != start {
		t.Fatalf("player moved while action locked: got %+v want %+v", player.Position, start)
	}
	if player.Combat.NextAttackTick != nextAttack {
		t.Fatalf("next attack tick changed while action locked: got %d want %d", player.Combat.NextAttackTick, nextAttack)
	}
	if player.Skills[swordQSkillID].CooldownUntilTick != 0 {
		t.Fatalf("q cooldown = %d, want 0 while action locked", player.Skills[swordQSkillID].CooldownUntilTick)
	}
}

func TestSwordRDoesNotCastOnNonAirborneHero(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), 100, nil, 20)

	if target.Stats.HP != startHP {
		t.Fatalf("target hp = %d, want unchanged %d", target.Stats.HP, startHP)
	}
	if player.Skills[swordRSkillID].CooldownUntilTick != 0 {
		t.Fatal("r should not enter cooldown without valid airborne target")
	}
}

func TestLastBreathDoesNotPenetrateArmorYet(t *testing.T) {
	attacker := &Entity{
		ID:     "player:p1",
		HeroID: swordHeroID,
		Stats:  Stats{Attack: 200},
		Sword:  SwordState{LastBreathUntilTick: 200},
	}
	target := &Entity{
		ID: "target",
		Stats: Stats{
			PhysicalDefense:      80,
			BonusPhysicalDefense: 40,
		},
	}

	w := testWorld(t)
	damage := w.attackDamage(attacker, target, 100)

	if damage != 111 {
		t.Fatalf("damage = %d, want 111", damage)
	}
}

func TestSpawnObjectRejectsUnsupportedKind(t *testing.T) {
	w := testWorld(t)
	if _, ok := w.SpawnObject(EntityKind("bad_kind"), TeamRed, 500, 600); ok {
		t.Fatal("unsupported kind should be rejected")
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

func learnSkill(entity *Entity, skillID string, level int) {
	state := entity.Skills[skillID]
	state.SkillID = skillID
	state.Level = level
	entity.Skills[skillID] = state
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
	return NewWorld(heroes, skills, levels, rewards)
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

func protocolPlayerInputCast(skillID string, targetX float64, targetY float64) protocol.PlayerInput {
	return protocolPlayerInputCastTarget(skillID, "", targetX, targetY)
}

func protocolPlayerInputCastTarget(skillID string, targetID string, targetX float64, targetY float64) protocol.PlayerInput {
	return protocol.PlayerInput{
		Cast: &protocol.CastInput{
			SkillID:  skillID,
			TargetID: targetID,
			TargetX:  targetX,
			TargetY:  targetY,
		},
	}
}

func protocolPlayerInputUpgrade(slot string) protocol.PlayerInput {
	return protocol.PlayerInput{
		UpgradeSkill: &protocol.UpgradeSkillInput{
			Slot: slot,
		},
	}
}

func protocolPlayerInputDebugLevelUp() protocol.PlayerInput {
	return protocol.PlayerInput{
		DebugLevelUp: true,
	}
}

func assertSkillEffect(t *testing.T, effects []SkillEffect, kind string) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			return
		}
	}
	t.Fatalf("missing skill effect %s", kind)
}

func tickSwordQRelease(t *testing.T, w *World, entity *Entity, tickRate int) uint64 {
	t.Helper()
	if entity == nil || !entity.Sword.QPending {
		t.Fatal("sword q is not pending")
	}
	releaseTick := entity.Sword.QReleaseTick
	w.Tick(releaseTick, tickRate)
	if entity.Sword.QPending {
		t.Fatalf("sword q still pending after release tick %d", releaseTick)
	}
	return releaseTick
}
