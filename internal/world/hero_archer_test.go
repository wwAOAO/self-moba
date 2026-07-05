package world

import (
	"l-battle/internal/config"
	"math"
	"testing"
)

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
	placeEntity(player, 3000, 3000)
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
	tickAttackRelease(t, w, player, 20)
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
	placeEntity(player, 3000, 3000)
	learnSkill(player, archerWSkillID, 1)
	startMP := player.Stats.MP

	w.ApplyInput("archer", protocolPlayerInputCast(archerWSkillID, player.Position.X+1000, player.Position.Y), 10, nil, 20)

	if got := countProjectilesByKind(w, "archer_volley_arrow"); got != 7 {
		t.Fatalf("volley projectiles = %d, want 7", got)
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
	placeEntity(player, 3000, 3000)
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

func TestArcherWProjectileDisappearsAfterHittingEnemy(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	hero.Base.Attack = 100
	hero.Base.CritChance = 0
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	placeEntity(player, 3000, 3000)
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
	if got := countProjectilesByKind(w, "archer_volley_arrow"); got != 7 {
		t.Fatalf("volley projectiles before hit = %d, want 7", got)
	}
	for tick := uint64(11); tick < 30; tick++ {
		w.Tick(tick, 20)
	}
	if got := countProjectilesByKind(w, "archer_volley_arrow"); got >= 7 {
		t.Fatalf("volley projectile count after hit = %d, want less than 7", got)
	}
}

func TestArcherWRepeatedGroupHitProjectileStillDisappears(t *testing.T) {
	w := testWorld(t)
	source := &Entity{
		ID:       "player:archer",
		Kind:     EntityKindPlayer,
		HeroID:   archerHeroID,
		Team:     TeamBlue,
		Position: Vector2{X: 100, Y: 100},
		Stats:    Stats{HP: 1000, MaxHP: 1000, Attack: 100},
	}
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: 130, Y: 100},
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
		Radius:   20,
	}
	w.entities[source.ID] = source
	w.entities[target.ID] = target
	w.projectileHits["volley"] = map[string]bool{target.ID: true}
	w.projectiles["arrow"] = &Projectile{
		ID:           "arrow",
		Kind:         "archer_volley_arrow",
		SkillID:      archerWSkillID,
		GroupID:      "volley",
		SourceID:     source.ID,
		Team:         TeamBlue,
		Position:     Vector2{X: 120, Y: 100},
		Dir:          Vector2{X: 1, Y: 0},
		Radius:       20,
		SpeedPerTick: 1,
		Range:        200,
		ExpiresAt:    100,
		HitIDs:       make(map[string]bool),
	}

	w.Tick(10, 20)

	if _, ok := w.projectiles["arrow"]; ok {
		t.Fatal("repeated archer w group-hit projectile should disappear")
	}
	if target.Stats.HP != 1000 {
		t.Fatalf("target hp = %d, want unchanged for repeated group hit", target.Stats.HP)
	}
}

func TestProjectileSweptCollisionRemovesArcherWBeforeItVisuallyPassesTarget(t *testing.T) {
	w := testWorld(t)
	source := &Entity{
		ID:       "player:archer",
		Kind:     EntityKindPlayer,
		HeroID:   archerHeroID,
		Team:     TeamBlue,
		Position: Vector2{X: 100, Y: 100},
		Stats:    Stats{HP: 1000, MaxHP: 1000, Attack: 100},
	}
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: 160, Y: 100},
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
		Radius:   20,
	}
	w.entities[source.ID] = source
	w.entities[target.ID] = target
	w.projectiles["arrow"] = &Projectile{
		ID:           "arrow",
		Kind:         "archer_volley_arrow",
		SkillID:      archerWSkillID,
		GroupID:      "volley",
		SourceID:     source.ID,
		Team:         TeamBlue,
		Position:     Vector2{X: 100, Y: 100},
		Dir:          Vector2{X: 1, Y: 0},
		Radius:       5,
		SpeedPerTick: 120,
		Range:        300,
		ExpiresAt:    100,
		HitIDs:       make(map[string]bool),
	}

	w.Tick(10, 20)

	if _, ok := w.projectiles["arrow"]; ok {
		t.Fatal("fast archer w projectile should be removed when its path crosses the target")
	}
	if target.Stats.HP >= 1000 {
		t.Fatalf("target hp = %d, want damaged by swept collision", target.Stats.HP)
	}
}

func TestArcherWArrowBodyCollisionRemovesProjectile(t *testing.T) {
	w := testWorld(t)
	source := &Entity{
		ID:       "player:archer",
		Kind:     EntityKindPlayer,
		HeroID:   archerHeroID,
		Team:     TeamBlue,
		Position: Vector2{X: 100, Y: 100},
		Stats:    Stats{HP: 1000, MaxHP: 1000, Attack: 100},
	}
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: 113, Y: 119},
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
		Radius:   8,
	}
	w.entities[source.ID] = source
	w.entities[target.ID] = target
	w.projectiles["arrow"] = &Projectile{
		ID:           "arrow",
		Kind:         "archer_volley_arrow",
		SkillID:      archerWSkillID,
		GroupID:      "volley",
		SourceID:     source.ID,
		Team:         TeamBlue,
		Position:     Vector2{X: 100, Y: 100},
		Dir:          Vector2{X: 1, Y: 0},
		Radius:       16,
		SpeedPerTick: 1,
		Range:        300,
		ExpiresAt:    100,
		HitIDs:       make(map[string]bool),
	}

	w.Tick(10, 20)

	if _, ok := w.projectiles["arrow"]; ok {
		t.Fatal("archer w projectile should be removed when its arrow body collides with target")
	}
	if target.Stats.HP >= 1000 {
		t.Fatalf("target hp = %d, want damaged by arrow body collision", target.Stats.HP)
	}
}

func TestArcherWPointBlankVolleyProjectilesAllDisappearOnSameTarget(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	hero.Base.Attack = 100
	hero.Base.CritChance = 0
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	placeEntity(player, 3000, 3000)
	learnSkill(player, archerWSkillID, 1)
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: player.Position.X + 45, Y: player.Position.Y},
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0},
		Radius:   18,
	}
	w.entities[target.ID] = target

	w.ApplyInput("archer", protocolPlayerInputCast(archerWSkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.Tick(11, 20)

	if got := countProjectilesByKind(w, "archer_volley_arrow"); got != 0 {
		t.Fatalf("point blank volley projectiles after hit = %d, want 0", got)
	}
	if target.Stats.HP != 880 {
		t.Fatalf("target hp = %d, want one volley hit only", target.Stats.HP)
	}
}

func TestArcherWEffectUsesCurrentPositionForFrontendSnapshotSmoothing(t *testing.T) {
	w := testWorld(t)
	projectile := &Projectile{
		ID:           "arrow",
		Kind:         "archer_volley_arrow",
		SkillID:      archerWSkillID,
		Team:         TeamBlue,
		Start:        Vector2{X: 100, Y: 100},
		Position:     Vector2{X: 180, Y: 100},
		Dir:          Vector2{X: 1, Y: 0},
		SpeedPerTick: 75,
		Range:        1200,
		Radius:       16,
		CreatedAt:    10,
		ExpiresAt:    100,
		HitIDs:       make(map[string]bool),
	}
	w.projectiles[projectile.ID] = projectile

	effects := w.SkillEffects()

	if len(effects) != 1 {
		t.Fatalf("effect count = %d, want 1", len(effects))
	}
	if effects[0].Start != projectile.Position {
		t.Fatalf("effect start = %+v, want current position %+v", effects[0].Start, projectile.Position)
	}
	if effects[0].CreatedAt != projectile.CreatedAt {
		t.Fatalf("effect created at = %d, want %d", effects[0].CreatedAt, projectile.CreatedAt)
	}
}

func TestArcherREffectUsesCurrentPositionForFrontendSnapshotSmoothing(t *testing.T) {
	w := testWorld(t)
	projectile := &Projectile{
		ID:           "arrow",
		Kind:         "archer_crystal_arrow",
		SkillID:      archerRSkillID,
		Team:         TeamBlue,
		Start:        Vector2{X: 100, Y: 100},
		Position:     Vector2{X: 260, Y: 100},
		Dir:          Vector2{X: 1, Y: 0},
		SpeedPerTick: 75,
		Range:        6000,
		Radius:       130,
		CreatedAt:    10,
		ExpiresAt:    100,
		HitIDs:       make(map[string]bool),
	}
	w.projectiles[projectile.ID] = projectile

	effects := w.SkillEffects()

	if len(effects) != 1 {
		t.Fatalf("effect count = %d, want 1", len(effects))
	}
	if effects[0].Start != projectile.Position {
		t.Fatalf("effect start = %+v, want current position %+v", effects[0].Start, projectile.Position)
	}
	if effects[0].CreatedAt != projectile.CreatedAt {
		t.Fatalf("effect created at = %d, want %d", effects[0].CreatedAt, projectile.CreatedAt)
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

func TestArcherEUsesMousePointAsLandingLocation(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	learnSkill(player, archerESkillID, 1)
	w.refreshArcherSkillOnUpgrade(player, archerESkillID)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 100, Y: player.Position.Y}
	landing := Vector2{X: player.Position.X + 700, Y: player.Position.Y + 250}

	w.ApplyInput("archer", protocolPlayerInputCastTarget(archerESkillID, target.ID, landing.X, landing.Y), 10, nil, 20)

	effects := w.SkillEffects()
	if len(effects) != 1 {
		t.Fatalf("effect count = %d, want 1", len(effects))
	}
	if effects[0].Kind != "archer_hawk" {
		t.Fatalf("effect kind = %s, want archer_hawk", effects[0].Kind)
	}
	if effects[0].End != landing {
		t.Fatalf("hawk landing = %+v, want mouse point %+v", effects[0].End, landing)
	}
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
	if player.Skills[archerRSkillID].CooldownUntilTick != 0 {
		t.Fatalf("r cooldown during cast delay = %d, want 0", player.Skills[archerRSkillID].CooldownUntilTick)
	}
	if player.Control.ActionLockedUntilTick != 15 {
		t.Fatalf("cast delay lock until = %d, want 15", player.Control.ActionLockedUntilTick)
	}
	w.Tick(15, 20)
	if len(w.projectiles) != 1 {
		t.Fatalf("projectiles after cast delay = %d, want 1", len(w.projectiles))
	}
	if player.Skills[archerRSkillID].CooldownUntilTick != 2015 {
		t.Fatalf("r cooldown until = %d, want 2015", player.Skills[archerRSkillID].CooldownUntilTick)
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

func TestArcherFocusBasicArrowDisplaysThreeArrowsWithoutExtraProjectiles(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	player := w.entities[playerEntityID("archer")]
	placeEntity(player, 3000, 3000)
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	player.Archer.FocusActiveUntil = 100

	w.ApplyInput("archer", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, player, 20)

	if got := countProjectilesByKind(w, "basic_arrow"); got != 1 {
		t.Fatalf("basic arrow projectile count = %d, want 1", got)
	}
	effects := w.SkillEffects()
	if len(effects) != 1 {
		t.Fatalf("effect count = %d, want 1", len(effects))
	}
	if effects[0].Kind != "basic_arrow" {
		t.Fatalf("effect kind = %s, want basic_arrow", effects[0].Kind)
	}
	if effects[0].Count != 3 {
		t.Fatalf("display arrow count = %d, want 3", effects[0].Count)
	}
}
