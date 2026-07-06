package world

import (
	"l-battle/internal/protocol"
	"math"
	"testing"
)

func TestMageConfiguredStatsAtLevel18(t *testing.T) {
	heroes := testWorld(t).heroes
	hero, ok := heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	if hero.Resource != "mp" {
		t.Fatalf("mage resource = %s, want mp", hero.Resource)
	}
	stats := heroStatsAtLevel(hero, MaxHeroLevel)
	if stats.MaxHP != 2090 {
		t.Fatalf("mage level 18 hp = %d, want 2090", stats.MaxHP)
	}
	if math.Abs(stats.MaxMP-879.5) > 0.000001 {
		t.Fatalf("mage level 18 mp = %f, want 879.5", stats.MaxMP)
	}
	if math.Abs(stats.Attack-106.2) > 0.000001 {
		t.Fatalf("mage level 18 attack = %f, want 106.2", stats.Attack)
	}
	if math.Abs(stats.AttackSpeed-0.907833) > 0.000001 {
		t.Fatalf("mage level 18 attack speed = %f, want 0.907833", stats.AttackSpeed)
	}
	if stats.AttackWindupSeconds != 0.234 {
		t.Fatalf("mage attack windup = %f, want 0.234", stats.AttackWindupSeconds)
	}
	if math.Abs(stats.PhysicalDefense-98.9) > 0.000001 {
		t.Fatalf("mage level 18 armor = %f, want 98.9", stats.PhysicalDefense)
	}
	if math.Abs(stats.MagicDefense-52.1) > 0.000001 {
		t.Fatalf("mage level 18 magic resist = %f, want 52.1", stats.MagicDefense)
	}
	if stats.MoveSpeed != 330 {
		t.Fatalf("mage move speed = %f, want 330", stats.MoveSpeed)
	}
	if stats.AttackRange != 550 {
		t.Fatalf("mage attack range = %f, want 550", stats.AttackRange)
	}
	if math.Abs(stats.HPRegen5-14.85) > 0.000001 {
		t.Fatalf("mage level 18 hp regen = %f, want 14.85", stats.HPRegen5)
	}
	if math.Abs(stats.MPRegen5-21.6) > 0.000001 {
		t.Fatalf("mage level 18 mp regen = %f, want 21.6", stats.MPRegen5)
	}
}

func TestMagePassiveSkillHitAppliesAndRefreshesIllumination(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	target := w.entities["enemy:hero-1"]

	w.onHeroSkillHit(player, target, 10, 20)
	if target.Control.MageIlluminationBy != player.ID {
		t.Fatalf("illumination source = %q, want %q", target.Control.MageIlluminationBy, player.ID)
	}
	if target.Control.MageIlluminationUntil != 130 {
		t.Fatalf("illumination until = %d, want 130", target.Control.MageIlluminationUntil)
	}

	w.onHeroSkillHit(player, target, 40, 20)
	if target.Control.MageIlluminationUntil != 160 {
		t.Fatalf("refreshed illumination until = %d, want 160", target.Control.MageIlluminationUntil)
	}
}

func TestMagePassiveBasicAttackDetonatesIllumination(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	player.Stats.AbilityPower = 100
	target := &Entity{
		ID:     "target",
		Kind:   EntityKindEnemyHero,
		Team:   TeamRed,
		Stats:  Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
		Radius: 18,
	}

	w.onHeroSkillHit(player, target, 10, 20)
	w.onHeroBasicHit(player, target, 20, 20)

	if target.Stats.HP != 950 {
		t.Fatalf("target hp = %d, want 950", target.Stats.HP)
	}
	if target.Combat.LastDamage != 50 || target.Combat.LastDamageType != "magic" {
		t.Fatalf("last damage = %d/%s, want 50/magic", target.Combat.LastDamage, target.Combat.LastDamageType)
	}
	if target.Control.MageIlluminationUntil != 0 || target.Control.MageIlluminationBy != "" {
		t.Fatalf("illumination = %q until %d, want cleared", target.Control.MageIlluminationBy, target.Control.MageIlluminationUntil)
	}
}

func TestMagePassiveDisplayDamageIncludesBasicAttackAndIllumination(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	player.Stats.AbilityPower = 100
	target := &Entity{
		ID:     "target",
		Kind:   EntityKindEnemyHero,
		Team:   TeamRed,
		Stats:  Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
		Radius: 18,
	}

	target.Combat.LastHitTick = 10
	w.applyDamage(player, target, 37, 20)
	w.onHeroSkillHit(player, target, 10, 20)
	w.onHeroBasicHit(player, target, 10, 20)

	if len(target.Combat.DamageEvents) != 2 {
		t.Fatalf("damage events = %#v, want 2 events", target.Combat.DamageEvents)
	}
	if target.Combat.DamageEvents[0].Damage != 37 || target.Combat.DamageEvents[0].DamageType != "physical" {
		t.Fatalf("first damage event = %#v, want 37 physical", target.Combat.DamageEvents[0])
	}
	if target.Combat.DamageEvents[1].Damage != 50 || target.Combat.DamageEvents[1].DamageType != "magic" {
		t.Fatalf("second damage event = %#v, want 50 magic", target.Combat.DamageEvents[1])
	}
}

func TestMagePassiveUltimateHitDetonatesAndReappliesIllumination(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	player.Level = 18
	target := &Entity{
		ID:     "target",
		Kind:   EntityKindEnemyHero,
		Team:   TeamRed,
		Stats:  Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
		Radius: 18,
	}

	w.onHeroSkillHit(player, target, 10, 20)
	w.ApplyMageIlluminationOnUltimateHit(player, target, 20, 20)

	if target.Stats.HP != 800 {
		t.Fatalf("target hp = %d, want 800", target.Stats.HP)
	}
	if target.Control.MageIlluminationBy != player.ID {
		t.Fatalf("illumination source after ult = %q, want %q", target.Control.MageIlluminationBy, player.ID)
	}
	if target.Control.MageIlluminationUntil != 140 {
		t.Fatalf("illumination until after ult = %d, want 140", target.Control.MageIlluminationUntil)
	}
}

func TestMageQFiresAfterWindupConsumesManaAndStartsCooldownOnRelease(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageQSkillID, 1)
	startMP := player.Stats.MP

	w.ApplyInput("mage", protocolPlayerInputCast(mageQSkillID, player.Position.X+600, player.Position.Y), 10, w.skills, 20)
	if player.Stats.MP != startMP-50 {
		t.Fatalf("mp after q start = %f, want %f", player.Stats.MP, startMP-50)
	}
	if !player.Mage.LightBindingPending || player.Mage.LightBindingReleaseTick != 15 {
		t.Fatalf("pending q = %v release %d, want true/15", player.Mage.LightBindingPending, player.Mage.LightBindingReleaseTick)
	}
	if player.Skills[mageQSkillID].CooldownUntilTick != 0 {
		t.Fatalf("cooldown before release = %d, want 0", player.Skills[mageQSkillID].CooldownUntilTick)
	}

	w.Tick(14, 20)
	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles before windup release = %d, want 0", len(w.projectiles))
	}
	w.Tick(15, 20)
	if len(w.projectiles) != 1 {
		t.Fatalf("projectiles after windup release = %d, want 1", len(w.projectiles))
	}
	if player.Skills[mageQSkillID].CooldownUntilTick != 315 {
		t.Fatalf("cooldown until = %d, want 315", player.Skills[mageQSkillID].CooldownUntilTick)
	}
}

func TestMageQHitsUpToTwoTargetsRootsAndAppliesIllumination(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageQSkillID, 1)
	player.Stats.AbilityPower = 100
	player.Position = Vector2{X: 1000, Y: 1000}
	first := &Entity{ID: "first", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1100, Y: 1000}, Radius: 18, Stats: Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0}}
	second := &Entity{ID: "second", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1200, Y: 1000}, Radius: 18, Stats: Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0}}
	third := &Entity{ID: "third", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1300, Y: 1000}, Radius: 18, Stats: Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0}}
	w.entities[first.ID] = first
	w.entities[second.ID] = second
	w.entities[third.ID] = third

	w.ApplyInput("mage", protocolPlayerInputCast(mageQSkillID, 1600, 1000), 10, w.skills, 20)
	for tick := uint64(15); tick < 25; tick++ {
		w.Tick(tick, 20)
	}

	if first.Stats.HP != 880 {
		t.Fatalf("first hp = %d, want 880", first.Stats.HP)
	}
	if second.Stats.HP != 940 {
		t.Fatalf("second hp = %d, want 940", second.Stats.HP)
	}
	if third.Stats.HP != 1000 {
		t.Fatalf("third hp = %d, want 1000", third.Stats.HP)
	}
	if first.Control.RootedUntilTick <= first.Combat.LastHitTick || second.Control.RootedUntilTick <= second.Combat.LastHitTick {
		t.Fatalf("root missing first=%d/%d second=%d/%d", first.Control.RootedUntilTick, first.Combat.LastHitTick, second.Control.RootedUntilTick, second.Combat.LastHitTick)
	}
	if first.Control.MageIlluminationBy != player.ID || second.Control.MageIlluminationBy != player.ID {
		t.Fatalf("illumination source first=%q second=%q want %q", first.Control.MageIlluminationBy, second.Control.MageIlluminationBy, player.ID)
	}
	if len(w.projectiles) != 0 {
		t.Fatalf("projectiles after two hits = %d, want 0", len(w.projectiles))
	}
}

func TestRootPreventsMovementButAllowsAttackAndCast(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageQSkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}
	player.Control.RootedUntilTick = 50
	target := &Entity{ID: "target", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1200, Y: 1000}, Radius: 18, Stats: Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0}}
	w.entities[target.ID] = target

	w.ApplyInput("mage", protocol.PlayerInput{Move: &protocol.MoveInput{TargetX: 1400, TargetY: 1000}}, 10, w.skills, 20)
	if player.Intent.MoveTarget != nil {
		t.Fatal("rooted move should not set move target")
	}
	before := player.Position
	w.tickPlayer(player, 11, 20)
	if player.Position != before {
		t.Fatalf("rooted player moved from %+v to %+v", before, player.Position)
	}

	w.ApplyInput("mage", protocolPlayerInputAttack(target.ID), 12, w.skills, 20)
	if player.Intent.AttackTargetID != target.ID {
		t.Fatalf("attack target = %q, want %q", player.Intent.AttackTargetID, target.ID)
	}
	w.ApplyInput("mage", protocolPlayerInputCast(mageQSkillID, target.Position.X, target.Position.Y), 13, w.skills, 20)
	if !player.Mage.LightBindingPending {
		t.Fatal("root should still allow casting")
	}
}

func TestMageBasicAttackFiresProjectile(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	player.Position = Vector2{X: 1000, Y: 1000}
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: 1200, Y: 1000},
		Radius:   18,
		Stats:    Stats{HP: 1000, MaxHP: 1000},
	}
	w.entities[target.ID] = target

	w.resolveBasicAttack(player, target, 10, 20)

	if len(w.projectiles) != 1 {
		t.Fatalf("projectiles = %d, want 1", len(w.projectiles))
	}
	if target.Combat.LastHitTick != 0 {
		t.Fatalf("mage ranged attack should not hit instantly, last hit tick = %d", target.Combat.LastHitTick)
	}
}

func TestMageWShieldsSelfOnCastAndAgainOnReturn(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageWSkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}

	w.ApplyInput("mage", protocolPlayerInputCast(mageWSkillID, 1300, 1000), 10, w.skills, 20)
	w.Tick(14, 20)
	if player.Passive.Shield != 50 {
		t.Fatalf("shield on release = %d, want 50", player.Passive.Shield)
	}
	for tick := uint64(15); tick < 35; tick++ {
		w.Tick(tick, 20)
	}
	if player.Passive.Shield != 100 {
		t.Fatalf("shield after return = %d, want 100", player.Passive.Shield)
	}
	if len(player.Passive.ShieldLayers) != 2 {
		t.Fatalf("shield layers = %d, want 2", len(player.Passive.ShieldLayers))
	}
}

func TestMageWShieldsAlliedPlayerItTouches(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	w.SpawnHero("ally", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	ally := w.entities[playerEntityID("ally")]
	learnSkill(player, mageWSkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}
	ally.Position = Vector2{X: 1150, Y: 1000}

	w.ApplyInput("mage", protocolPlayerInputCast(mageWSkillID, 1400, 1000), 10, w.skills, 20)
	for tick := uint64(14); tick < 18; tick++ {
		w.Tick(tick, 20)
	}
	if ally.Passive.Shield != 50 {
		t.Fatalf("ally shield = %d, want 50", ally.Passive.Shield)
	}
}

func TestMageWReturnsOnOriginalCastLine(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageWSkillID, 1)
	projectile := &Projectile{
		SkillID:      mageWSkillID,
		Returning:    true,
		Position:     Vector2{X: 1300, Y: 1000},
		Dir:          Vector2{X: -1, Y: 0},
		Start:        Vector2{X: 1000, Y: 1000},
		Range:        2300,
		Traveled:     1200,
		SpeedPerTick: 10,
		SpeedMin:     10,
	}
	player.Position = Vector2{X: 1000, Y: 1300}

	if projectile.Dir.X != -1 || projectile.Dir.Y != 0 {
		t.Fatalf("return dir = %+v, want original cast line", projectile.Dir)
	}
}

func TestMageWReturnsToCastOriginEvenIfCasterMoves(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageWSkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}

	w.ApplyInput("mage", protocolPlayerInputCast(mageWSkillID, 2150, 1000), 10, w.skills, 20)
	player.Position = Vector2{X: 1000, Y: 1450}
	returnTick := uint64(0)
	seenProjectile := false
	for tick := uint64(14); tick < 140; tick++ {
		w.Tick(tick, 20)
		if len(w.projectiles) > 0 {
			seenProjectile = true
		}
		if seenProjectile && len(w.projectiles) == 0 {
			returnTick = tick
			break
		}
	}
	if len(w.projectiles) != 0 {
		t.Fatal("mage w should disappear after returning to cast origin")
	}
	if returnTick == 0 {
		t.Fatal("mage w return tick not recorded")
	}
	if player.Passive.Shield < 100 {
		t.Fatalf("self shield at return tick %d = %d, want two layers", returnTick, player.Passive.Shield)
	}
}

func TestMageWSpeedSlowsOutAndAcceleratesBack(t *testing.T) {
	projectile := &Projectile{
		SkillID:  mageWSkillID,
		Range:    2000,
		SpeedMin: 120,
	}
	updateMageWProjectileSpeed(projectile)
	outStart := projectile.SpeedPerTick
	projectile.Traveled = 900
	updateMageWProjectileSpeed(projectile)
	outEnd := projectile.SpeedPerTick
	if outEnd >= outStart {
		t.Fatalf("out speed = %f then %f, want slowing", outStart, outEnd)
	}

	projectile.Returning = true
	projectile.Traveled = 1100
	updateMageWProjectileSpeed(projectile)
	backStart := projectile.SpeedPerTick
	projectile.Traveled = 1900
	updateMageWProjectileSpeed(projectile)
	backEnd := projectile.SpeedPerTick
	if backEnd <= backStart {
		t.Fatalf("return speed = %f then %f, want accelerating", backStart, backEnd)
	}
}

func TestMageEPlacesZoneSlowsAndManualDetonates(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageESkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: 1150, Y: 1000},
		Radius:   18,
		Stats:    Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
	}
	w.entities[target.ID] = target

	w.ApplyInput("mage", protocolPlayerInputCast(mageESkillID, 1150, 1000), 10, w.skills, 20)
	w.Tick(15, 20)
	if player.Skills[mageESkillID].CooldownUntilTick != 0 {
		t.Fatalf("cooldown during e projectile = %d, want 0", player.Skills[mageESkillID].CooldownUntilTick)
	}
	if player.Mage.LucentSingularityActive {
		t.Fatal("mage e should fly before placing zone")
	}
	if len(w.projectiles) != 1 {
		t.Fatalf("mage e projectile count = %d, want 1", len(w.projectiles))
	}
	for tick := uint64(16); tick <= 18; tick++ {
		w.Tick(tick, 20)
	}
	if !player.Mage.LucentSingularityActive {
		t.Fatal("mage e should be active after projectile arrives")
	}
	if player.Skills[mageESkillID].CooldownUntilTick != 0 {
		t.Fatalf("cooldown before e detonate = %d, want 0", player.Skills[mageESkillID].CooldownUntilTick)
	}
	if target.Control.MoveSpeedSlow != 0.3 {
		t.Fatalf("zone slow = %f, want 0.3", target.Control.MoveSpeedSlow)
	}

	w.ApplyInput("mage", protocolPlayerInputCast(mageESkillID, 1150, 1000), 19, w.skills, 20)
	if target.Stats.HP != 920 {
		t.Fatalf("target hp = %d, want 920", target.Stats.HP)
	}
	if target.Control.MoveSpeedSlowUntil != 39 {
		t.Fatalf("detonate slow until = %d, want 39", target.Control.MoveSpeedSlowUntil)
	}
	if player.Mage.LucentSingularityActive {
		t.Fatal("mage e should be inactive after detonation")
	}
	if player.Skills[mageESkillID].CooldownUntilTick != 219 {
		t.Fatalf("cooldown after e detonate = %d, want 219", player.Skills[mageESkillID].CooldownUntilTick)
	}
}

func TestMageEAutoDetonatesAfterDuration(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageESkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}
	target := &Entity{
		ID:       "target",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		Position: Vector2{X: 1150, Y: 1000},
		Radius:   18,
		Stats:    Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0},
	}
	w.entities[target.ID] = target

	w.ApplyInput("mage", protocolPlayerInputCast(mageESkillID, 1150, 1000), 10, w.skills, 20)
	for tick := uint64(15); tick <= 118; tick++ {
		w.Tick(tick, 20)
	}
	if target.Stats.HP != 920 {
		t.Fatalf("target hp after auto detonate = %d, want 920", target.Stats.HP)
	}
}

func TestMageRFiresBeamAfterWindupHitsLineAndAppliesIllumination(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageRSkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}
	target := &Entity{ID: "target", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1500, Y: 1040}, Radius: 18, Stats: Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0}}
	miss := &Entity{ID: "miss", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1500, Y: 1150}, Radius: 18, Stats: Stats{HP: 1000, MaxHP: 1000, MagicDefense: 0}}
	w.entities[target.ID] = target
	w.entities[miss.ID] = miss

	w.ApplyInput("mage", protocolPlayerInputCast(mageRSkillID, 2000, 1000), 10, w.skills, 20)
	if !player.Mage.FinalSparkPending || player.Mage.FinalSparkReleaseTick != 20 {
		t.Fatalf("pending r = %v release %d, want true/20", player.Mage.FinalSparkPending, player.Mage.FinalSparkReleaseTick)
	}
	if player.Skills[mageRSkillID].CooldownUntilTick != 0 {
		t.Fatalf("cooldown before release = %d, want 0", player.Skills[mageRSkillID].CooldownUntilTick)
	}
	w.Tick(20, 20)

	if target.Stats.HP != 700 {
		t.Fatalf("target hp = %d, want 700", target.Stats.HP)
	}
	if miss.Stats.HP != 1000 {
		t.Fatalf("miss hp = %d, want 1000", miss.Stats.HP)
	}
	if target.Control.MageIlluminationBy != player.ID {
		t.Fatalf("illumination source = %q, want %q", target.Control.MageIlluminationBy, player.ID)
	}
	if player.Skills[mageRSkillID].CooldownUntilTick != 1620 {
		t.Fatalf("r cooldown = %d, want 1620", player.Skills[mageRSkillID].CooldownUntilTick)
	}
}

func TestMageRRefundsCooldownWhenMarkedHeroDies(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(mageHeroID)
	if !ok {
		t.Fatal("mage hero not found")
	}
	w.SpawnHero("mage", hero, TeamBlue)
	player := w.entities[playerEntityID("mage")]
	learnSkill(player, mageRSkillID, 1)
	player.Position = Vector2{X: 1000, Y: 1000}
	target := &Entity{ID: "target", Kind: EntityKindEnemyHero, Team: TeamRed, Position: Vector2{X: 1500, Y: 1000}, Radius: 18, Stats: Stats{HP: 300, MaxHP: 300, MagicDefense: 0}}
	w.entities[target.ID] = target

	w.ApplyInput("mage", protocolPlayerInputCast(mageRSkillID, 2000, 1000), 10, w.skills, 20)
	w.Tick(20, 20)

	if target.Stats.HP != 0 {
		t.Fatalf("target hp = %d, want 0", target.Stats.HP)
	}
	if player.Skills[mageRSkillID].CooldownUntilTick != 1140 {
		t.Fatalf("r cooldown after refund = %d, want 1140", player.Skills[mageRSkillID].CooldownUntilTick)
	}
}
