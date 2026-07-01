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

	w.applyMageIlluminationOnSkillHit(player, target, 10, 20)
	if target.Control.MageIlluminationBy != player.ID {
		t.Fatalf("illumination source = %q, want %q", target.Control.MageIlluminationBy, player.ID)
	}
	if target.Control.MageIlluminationUntil != 130 {
		t.Fatalf("illumination until = %d, want 130", target.Control.MageIlluminationUntil)
	}

	w.applyMageIlluminationOnSkillHit(player, target, 40, 20)
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

	w.applyMageIlluminationOnSkillHit(player, target, 10, 20)
	w.triggerMageIlluminationOnBasicAttack(player, target, 20, 20)

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

	w.applyMageIlluminationOnSkillHit(player, target, 10, 20)
	w.applyMageIlluminationOnUltimateHit(player, target, 20, 20)

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
