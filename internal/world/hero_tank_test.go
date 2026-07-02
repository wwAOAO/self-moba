package world

import (
	"l-battle/internal/config"
	"math"
	"testing"
)

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
	if player.Skills[tankQSkillID].CooldownUntilTick != 0 {
		t.Fatalf("tank q cooldown during windup = %d, want 0", player.Skills[tankQSkillID].CooldownUntilTick)
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
	if player.Skills[tankQSkillID].CooldownUntilTick != 175 {
		t.Fatalf("tank q cooldown = %d, want 175", player.Skills[tankQSkillID].CooldownUntilTick)
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

func TestTankQAlwaysHitsLockedTargetAtEnd(t *testing.T) {
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
	target.Stats.MagicDefense = 0

	w.ApplyInput("tank", protocolPlayerInputCastTarget(tankQSkillID, target.ID, target.Position.X, target.Position.Y), 10, w.skills, 20)
	w.Tick(15, 20)
	target.Position = Vector2{X: player.Position.X + 2000, Y: player.Position.Y + 2000}
	for tick := uint64(16); tick <= 30; tick++ {
		w.Tick(tick, 20)
		if len(w.projectiles) == 0 {
			break
		}
	}

	if target.Combat.LastHitTick == 0 {
		t.Fatal("tank q should hit locked target at end")
	}
	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles = %d, want 0 after forced hit", len(w.projectiles))
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
	if player.Skills[tankESkillID].CooldownUntilTick != 0 {
		t.Fatalf("tank e cooldown during windup = %d, want 0", player.Skills[tankESkillID].CooldownUntilTick)
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
	if player.Skills[tankESkillID].CooldownUntilTick != 155 {
		t.Fatalf("tank e cooldown = %d, want 155", player.Skills[tankESkillID].CooldownUntilTick)
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
