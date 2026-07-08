package world

import (
	"l-battle/internal/config"
	"math"
	"testing"
)

func TestAttackTargetAutoAttacksOnServerTick(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)

	target := w.entities["dummy:training-1"]
	target.Position = Vector2{X: w.entities[playerEntityID("p1")].Position.X + 100, Y: w.entities[playerEntityID("p1")].Position.Y}
	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 1, nil, 20)
	w.Tick(2, 20)
	tickAttackRelease(t, w, w.entities[playerEntityID("p1")], 20)

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
	tickAttackRelease(t, w, attacker, 20)

	if target.Stats.HP >= startHP {
		t.Fatalf("target hp = %v, start hp = %v; opposing player was not damaged", target.Stats.HP, startHP)
	}
}

func TestGlacialBucklerAppliesManaArmorAndAbilityHaste(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 1250
	baseArmor := player.Stats.PhysicalDefense
	baseMP := player.Stats.MaxMP
	baseHaste := player.Stats.AbilityHaste

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("glacial_buckler"), 1, nil, 20)

	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "glacial_buckler" {
		t.Fatalf("equipment = %+v, want glacial_buckler", player.Equipment)
	}
	if player.Stats.PhysicalDefense != baseArmor+45 {
		t.Fatalf("physical defense = %f, want %f", player.Stats.PhysicalDefense, baseArmor+45)
	}
	if player.Stats.MaxMP != baseMP+300 {
		t.Fatalf("max mp = %f, want %f", player.Stats.MaxMP, baseMP+300)
	}
	if player.Stats.AbilityHaste != baseHaste+15 {
		t.Fatalf("ability haste = %f, want %f", player.Stats.AbilityHaste, baseHaste+15)
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
	tickAttackRelease(t, w, player, 20)

	if target.Stats.HP >= startHP {
		t.Fatalf("enemy hero hp = %v, want below %v after basic attack", target.Stats.HP, startHP)
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
			t.Fatalf("damage after %f resistance = %v, want %v", tt.resistance, got, tt.want)
		}
	}
}

func TestPenetrationAppliesPercentBeforeFlatAndCannotCreateNegativeResistance(t *testing.T) {
	effective := effectiveResistance(100, 0.3, 15)
	if effective != 55 {
		t.Fatalf("effective resistance = %f, want 55", effective)
	}
	if got := damageAfterResistance(100, effective, 0); got != 65 {
		t.Fatalf("damage after penetration = %v, want 65", got)
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
		t.Fatalf("damage after resistance and reduction = %v, want 20", got)
	}
}

func TestDamageReductionStacksMultiplicatively(t *testing.T) {
	reduction := stackDamageReduction(0.2, 0.5)
	if math.Abs(reduction-0.6) > 0.0001 {
		t.Fatalf("stacked reduction = %f, want 0.6", reduction)
	}
	if got := damageAfterResistance(100, 0, reduction); got != 40 {
		t.Fatalf("damage after stacked reduction = %v, want 40", got)
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
		t.Fatalf("hp after true damage = %v, want 925", target.Stats.HP)
	}
}

func TestRangedMinionBasicAttackIsPhysicalWithMagicBonusAgainstNonHero(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{ID: "minion:ranged", Kind: EntityKindRangedMinion, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1, Attack: 100}}
	target := &Entity{
		ID:   "structure:target",
		Kind: EntityKindCrystal,
		Team: TeamRed,
		Stats: Stats{
			HP:              1000,
			MaxHP:           1000,
			PhysicalDefense: 0,
			MagicDefense:    0,
		},
	}
	target.Combat.LastHitTick = 10

	w.applyMinionBasicAttackDamage(attacker, target, 10, 20)

	if got := 1000 - target.Stats.HP; got != 120 {
		t.Fatalf("ranged minion damage = %v, want 120", got)
	}
	wantTypes := []string{"physical", "magic"}
	wantDamage := []int{100, 20}
	if len(target.Combat.DamageEvents) != len(wantTypes) {
		t.Fatalf("damage events = %+v, want 2 events", target.Combat.DamageEvents)
	}
	for i := range wantTypes {
		event := target.Combat.DamageEvents[i]
		if event.DamageType != wantTypes[i] || event.Damage != wantDamage[i] || !event.BasicAttack {
			t.Fatalf("event %d = %+v, want %s/%d basic", i, event, wantTypes[i], wantDamage[i])
		}
	}
}

func TestRangedMinionBasicAttackIsOnlyPhysicalAgainstHero(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{ID: "minion:ranged", Kind: EntityKindRangedMinion, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1, Attack: 100}}
	target := &Entity{ID: "hero:target", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}
	target.Combat.LastHitTick = 10

	w.applyMinionBasicAttackDamage(attacker, target, 10, 20)

	if got := 1000 - target.Stats.HP; got != 60 {
		t.Fatalf("ranged minion hero damage = %v, want 60", got)
	}
	if len(target.Combat.DamageEvents) != 1 || target.Combat.DamageEvents[0].DamageType != "physical" {
		t.Fatalf("damage events = %+v, want one physical event", target.Combat.DamageEvents)
	}
}

func TestSiegeMinionBasicAttackIsPhysicalOnly(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{ID: "minion:siege", Kind: EntityKindSiegeMinion, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1, Attack: 100}}
	target := &Entity{ID: "structure:target", Kind: EntityKindCrystal, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}
	target.Combat.LastHitTick = 10

	w.applyMinionBasicAttackDamage(attacker, target, 10, 20)

	if got := 1000 - target.Stats.HP; got != 100 {
		t.Fatalf("siege minion damage = %v, want 100", got)
	}
	if len(target.Combat.DamageEvents) != 1 || target.Combat.DamageEvents[0].DamageType != "physical" || !target.Combat.DamageEvents[0].BasicAttack {
		t.Fatalf("damage events = %+v, want one physical basic attack", target.Combat.DamageEvents)
	}
}

func TestSiegeMinionBasicAttackSplashBurnsEverySecondForFiveSeconds(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{ID: "minion:siege", Kind: EntityKindSiegeMinion, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1, Attack: 100}}
	target := &Entity{ID: "structure:target", Kind: EntityKindCrystal, Team: TeamRed, Position: Vector2{X: 1000, Y: 1000}, Stats: Stats{HP: 1000, MaxHP: 1000}}
	near := &Entity{ID: "structure:near", Kind: EntityKindCrystal, Team: TeamRed, Position: Vector2{X: 1200, Y: 1000}, Stats: Stats{HP: 1000, MaxHP: 1000}}
	far := &Entity{ID: "structure:far", Kind: EntityKindCrystal, Team: TeamRed, Position: Vector2{X: 1400, Y: 1000}, Stats: Stats{HP: 1000, MaxHP: 1000}}
	w.entities[attacker.ID] = attacker
	w.entities[target.ID] = target
	w.entities[near.ID] = near
	w.entities[far.ID] = far
	target.Combat.LastHitTick = 10

	w.applyMinionBasicAttackDamage(attacker, target, 10, 20)

	if got := 1000 - target.Stats.HP; got != 100 {
		t.Fatalf("primary damage = %v, want 100", got)
	}
	if got := 1000 - near.Stats.HP; got != 0 {
		t.Fatalf("near immediate splash damage = %v, want 0", got)
	}
	if got := 1000 - far.Stats.HP; got != 0 {
		t.Fatalf("far splash damage = %v, want 0", got)
	}

	for tick := uint64(11); tick <= 10+secondsToTicks(5, 20); tick++ {
		w.tickSiegeMinionSplashBurns(tick, 20)
	}
	if got := 1000 - near.Stats.HP; got != 550 {
		t.Fatalf("near splash burn damage = %v, want 550", got)
	}
	if got := near.Combat.LastDamageType; got != "magic" {
		t.Fatalf("near splash damage type = %q, want magic", got)
	}
}

func TestSiegeMinionSplashAttackFiresCannonballProjectile(t *testing.T) {
	w := testWorld(t)
	attacker := &Entity{ID: "minion:siege", Kind: EntityKindSiegeMinion, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1, Attack: 100}}
	target := &Entity{ID: "structure:target", Kind: EntityKindCrystal, Team: TeamRed, Position: Vector2{X: 1000, Y: 1000}, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.fireBasicAttackProjectile(attacker, target, 10, 20)
	if got := countProjectilesByKind(w, siegeCannonballProjectileKind); got != 1 {
		t.Fatalf("ready siege projectile count = %v, want 1 cannonball", got)
	}

	w.projectiles = map[string]*Projectile{}
	attacker.Combat.NextSiegeSplashTick = 20
	w.fireBasicAttackProjectile(attacker, target, 11, 20)
	if got := countProjectilesByKind(w, basicArrowProjectileKind); got != 1 {
		t.Fatalf("cooling siege projectile count = %v, want 1 basic projectile", got)
	}
}

func TestMeleeMinionReducesDamageFromNonHeroUnits(t *testing.T) {
	w := testWorld(t)
	target := &Entity{ID: "minion:melee", Kind: EntityKindMeleeMinion, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}
	minion := &Entity{ID: "minion:ranged", Kind: EntityKindRangedMinion, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1}}
	hero := &Entity{ID: "hero:blue", Kind: EntityKindEnemyHero, Team: TeamBlue, Stats: Stats{HP: 1, MaxHP: 1}}

	target.Combat.LastHitTick = 10
	w.applyResolvedDamage(minion, target, 100, "physical", sustainSingleTargetSkill, 20)
	if got := 1000 - target.Stats.HP; got != 85 {
		t.Fatalf("non-hero damage = %v, want 85", got)
	}

	target.Stats.HP = 1000
	target.Combat.LastHitTick = 20
	w.applyResolvedDamage(hero, target, 100, "physical", sustainSingleTargetSkill, 20)
	if got := 1000 - target.Stats.HP; got != 100 {
		t.Fatalf("hero damage = %v, want 100", got)
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
	heroes, err := config.LoadHeroes("../../configs/heroes")
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
		t.Fatalf("warrior q cooldown with 100 haste = %v, want 90", got)
	}
}

func TestWarriorQResetsBasicAttack(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(warriorHeroID)
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero("warrior", hero, TeamBlue)
	player := w.entities[playerEntityID("warrior")]
	target := w.entities["enemy:hero-1"]
	placeEntity(player, 1000, 1000)
	placeEntity(target, 1100, 1000)
	player.Stats.AttackSpeed = 1
	target.Stats.PhysicalDefense = 0
	learnSkill(player, warriorQSkillID, 1)

	w.ApplyInput("warrior", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, player, 20)
	if got, want := player.Combat.NextAttackTick, uint64(30); got != want {
		t.Fatalf("next attack after first hit = %v, want %v", got, want)
	}

	w.ApplyInput("warrior", protocolPlayerInputCast(warriorQSkillID, target.Position.X, target.Position.Y), 16, nil, 20)
	w.Tick(16, 20)

	if got := player.Combat.PendingAttackTargetID; got != target.ID {
		t.Fatalf("pending attack after q reset = %q, want %q", got, target.ID)
	}
	if got, want := player.Combat.AttackReleaseTick, uint64(22); got != want {
		t.Fatalf("second attack release tick = %v, want %v", got, want)
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
		t.Fatalf("damage during attack windup = %v, want 0", target.Combat.LastDamage)
	}
	w.Tick(15, 20)
	if target.Combat.LastDamage <= 0 {
		t.Fatal("basic attack should damage after windup")
	}
	if len(target.Combat.DamageEvents) != 1 || !target.Combat.DamageEvents[0].BasicAttack {
		t.Fatalf("basic attack damage events = %+v", target.Combat.DamageEvents)
	}
	if player.Combat.NextAttackTick != 30 {
		t.Fatalf("next attack tick = %v, want 30", player.Combat.NextAttackTick)
	}
}

func TestBasicAttackWindupCompletesWhenTargetMovesOutOfRange(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.AttackRange = 120
	hero.Base.AttackSpeed = 1
	hero.Base.MoveSpeed = 400
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Stats.PhysicalDefense = 0
	placeEntity(player, 1000, 1000)
	placeEntity(target, 1140, 1000)

	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	placeEntity(target, 1800, 1000)
	before := player.Position
	w.Tick(11, 20)
	if player.Position != before {
		t.Fatalf("player moved during attack windup: got %+v want %+v", player.Position, before)
	}
	tickAttackRelease(t, w, player, 20)

	if target.Combat.LastDamage <= 0 {
		t.Fatal("basic attack should complete after windup even if target moved out of range")
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
	tickAttackRelease(t, w, player, 20)
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
	placeEntity(player, 3000, 3000)
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}

	w.ApplyInput("archer", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	if got := countProjectilesByKind(w, "basic_arrow"); got != 0 {
		t.Fatalf("basic arrows during attack windup = %v, want 0", got)
	}
	tickAttackRelease(t, w, player, 20)
	if got := countProjectilesByKind(w, "basic_arrow"); got == 0 {
		t.Fatal("ranged basic attack should fire projectile after windup")
	}
}

func TestExplorerBasicAttackFiresProjectile(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	player := w.entities[playerEntityID("explorer")]
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}

	w.ApplyInput("explorer", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, player, 20)

	if got := countProjectilesByKind(w, "basic_arrow"); got == 0 {
		t.Fatal("explorer basic attack should fire projectile")
	}
}

func TestExplorerQHitsAppliesOnHitAndReducesCooldowns(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	explorer.Stats.Attack = 100
	explorer.Stats.AbilityPower = 50
	learnSkill(explorer, explorerQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	placeEntity(explorer, 1000, 1000)
	placeEntity(target, 1300, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerQSkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	if got := explorer.Stats.MP; got != 347 {
		t.Fatalf("mp after q cast = %f, want 347", got)
	}
	w.Tick(14, 20)
	if got := countProjectilesByKind(w, "explorer_q"); got != 0 {
		t.Fatalf("q projectile during windup = %d, want 0", got)
	}
	w.Tick(15, 20)
	if got := countProjectilesByKind(w, "explorer_q"); got != 1 {
		t.Fatalf("q projectile after windup = %d, want 1", got)
	}
	tickUntilDamage(t, w, target, 15, 30, 20)

	if got := target.Combat.LastDamage; got != 170 {
		t.Fatalf("q damage = %d, want 170", got)
	}
	if len(target.Combat.DamageEvents) != 1 || !target.Combat.DamageEvents[0].BasicAttack {
		t.Fatalf("q should apply on-hit basic attack context: %+v", target.Combat.DamageEvents)
	}
	if got := explorer.Skills[explorerQSkillID].CooldownUntilTick; got != 90 {
		t.Fatalf("q cooldown after refund = %d, want 90", got)
	}
	if got := explorer.Stats.AttackSpeedBonus; math.Abs(got-0.1) > 0.000001 {
		t.Fatalf("passive attack speed bonus = %f, want 0.1", got)
	}
}

func TestExplorerQResetsBasicAttack(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	learnSkill(explorer, explorerQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	placeEntity(explorer, 1000, 1000)
	placeEntity(target, 1300, 1000)
	explorer.Intent.AttackTargetID = target.ID
	explorer.Combat.NextAttackTick = 80

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerQSkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.Tick(15, 20)

	if got, want := explorer.Combat.NextAttackTick, uint64(47); got != want {
		t.Fatalf("next attack tick after q reset = %d, want %d", got, want)
	}
	if got := explorer.Combat.PendingAttackTargetID; got != target.ID {
		t.Fatalf("pending attack after q reset = %q, want %q", got, target.ID)
	}
	if got, want := explorer.Combat.AttackReleaseTick, uint64(20); got != want {
		t.Fatalf("attack release tick after q reset = %d, want %d", got, want)
	}
}

func TestExplorerWAttachesAndSkillDetonatesWithManaRefund(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	explorer.Stats.MP = 200
	explorer.Stats.MaxMP = 375
	explorer.Stats.Attack = 100
	explorer.Stats.BonusAttack = 20
	explorer.Stats.AbilityPower = 100
	learnSkill(explorer, explorerQSkillID, 1)
	learnSkill(explorer, explorerWSkillID, 1)
	target := w.entities["enemy:hero-1"]
	placeEntity(explorer, 1000, 1000)
	placeEntity(target, 1300, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0
	target.Stats.MagicDefense = 0

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerWSkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	if got := countProjectilesByKind(w, "explorer_w"); got != 0 {
		t.Fatalf("w projectile during windup = %d, want 0", got)
	}
	for tick := uint64(11); tick <= 30 && len(target.Passive.ExplorerFluxMarks) == 0; tick++ {
		w.Tick(tick, 20)
	}
	if len(target.Passive.ExplorerFluxMarks) != 1 {
		t.Fatalf("w mark missing: %+v", target.Passive.ExplorerFluxMarks)
	}
	if target.Combat.LastDamage != 0 {
		t.Fatalf("w attach damage = %d, want 0", target.Combat.LastDamage)
	}
	explorer.Stats.Attack = 100
	explorer.Stats.BonusAttack = 20
	explorer.Stats.AbilityPower = 100

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerQSkillID, target.Position.X, target.Position.Y), 31, nil, 20)
	tickUntilDamage(t, w, target, 31, 50, 20)

	if _, ok := target.Passive.ExplorerFluxMarks[explorer.ID]; ok {
		t.Fatal("w mark should be consumed")
	}
	if got := target.Combat.LastDamage; got != 190 {
		t.Fatalf("w detonation damage = %d, want 190", got)
	}
	if got := explorer.Stats.MP; got < 211 || got > 212 {
		t.Fatalf("mp after skill detonation refund = %f, want about 211", got)
	}
	if len(target.Combat.DamageEvents) != 2 || !target.Combat.DamageEvents[0].BasicAttack || target.Combat.DamageEvents[1].DamageType != "magic" {
		t.Fatalf("damage events = %+v, want q on-hit then w magic", target.Combat.DamageEvents)
	}
}

func TestExplorerEReleasesAfterWindupAndPrioritizesWTarget(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	explorer.Stats.MP = 200
	explorer.Stats.MaxMP = 375
	explorer.Stats.BonusAttack = 20
	explorer.Stats.AbilityPower = 100
	learnSkill(explorer, explorerWSkillID, 1)
	learnSkill(explorer, explorerESkillID, 1)
	near := w.entities["enemy:hero-1"]
	markedID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 1800, 1000)
	marked := w.entities[markedID]
	placeEntity(explorer, 1000, 1000)
	placeEntity(near, 1575, 1000)
	placeEntity(marked, 1800, 1000)
	near.Stats.HP = 1000
	near.Stats.MaxHP = 1000
	near.Stats.MagicDefense = 0
	marked.Stats.HP = 1000
	marked.Stats.MaxHP = 1000
	marked.Stats.MagicDefense = 0
	marked.Passive.ExplorerFluxMarks = map[string]ExplorerFluxState{
		explorer.ID: {Level: 1, ExpiresAt: 100},
	}

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerESkillID, 1475, 1000), 10, nil, 20)
	if got := explorer.Position.X; got != 1000 {
		t.Fatalf("e should not blink during windup, x=%f", got)
	}
	w.Tick(14, 20)
	if got := countProjectilesByKind(w, "explorer_e"); got != 0 {
		t.Fatalf("e projectile during windup = %d, want 0", got)
	}
	w.Tick(15, 20)
	if got := explorer.Position.X; got != 1475 {
		t.Fatalf("e blink x = %f, want 1475", got)
	}
	if got := countProjectilesByKind(w, "explorer_e"); got != 1 {
		t.Fatalf("e projectile after windup = %d, want 1", got)
	}
	tickUntilDamage(t, w, marked, 15, 30, 20)

	if near.Combat.LastDamage != 0 {
		t.Fatalf("near unmarked target damage = %d, want 0", near.Combat.LastDamage)
	}
	if _, ok := marked.Passive.ExplorerFluxMarks[explorer.ID]; ok {
		t.Fatal("w mark should be consumed by e")
	}
	if len(marked.Combat.DamageEvents) != 2 || marked.Combat.DamageEvents[0].Damage != 167 || marked.Combat.DamageEvents[1].Damage != 190 {
		t.Fatalf("marked damage events = %+v, want e magic then w detonation", marked.Combat.DamageEvents)
	}
	if got := explorer.Stats.AttackSpeedBonus; math.Abs(got-0.1) > 0.000001 {
		t.Fatalf("passive attack speed bonus = %f, want 0.1", got)
	}
}

func TestExplorerEInterruptedDuringWindupDoesNotBlinkOrFire(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	learnSkill(explorer, explorerESkillID, 1)
	placeEntity(explorer, 1000, 1000)

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerESkillID, 1475, 1000), 10, nil, 20)
	explorer.Control.StunnedUntilTick = 20
	w.Tick(11, 20)
	w.Tick(15, 20)

	if explorer.Passive.ExplorerEPending {
		t.Fatal("e pending should be cleared after interrupt")
	}
	if got := explorer.Position.X; got != 1000 {
		t.Fatalf("interrupted e blink x = %f, want 1000", got)
	}
	if got := countProjectilesByKind(w, "explorer_e"); got != 0 {
		t.Fatalf("interrupted e projectiles = %d, want 0", got)
	}
}

func TestExplorerEUsesRawMouseDirection(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	learnSkill(explorer, explorerESkillID, 1)
	placeEntity(explorer, 1000, 1000)

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerESkillID, 100000, 1000), 10, nil, 20)
	w.Tick(15, 20)

	if got := explorer.Position.X; got != 1475 {
		t.Fatalf("e blink x = %f, want 1475", got)
	}
	if got := explorer.Position.Y; got != 1000 {
		t.Fatalf("e blink y = %f, want 1000", got)
	}
}

func TestExplorerRDealsReducedDamageToMinionsAndNonEpicMonsters(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(explorerHeroID)
	if !ok {
		t.Fatal("explorer hero not found")
	}
	w.SpawnHero("explorer", hero, TeamBlue)
	explorer := w.entities[playerEntityID("explorer")]
	explorer.Stats.MP = 300
	learnSkill(explorer, explorerRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	minion := w.entities["minion:red-melee-1"]
	monster := &Entity{
		ID:       "spawn:test-gromp",
		Kind:     EntityKindGromp,
		Team:     TeamNeutral,
		Position: Vector2{X: 1900, Y: 1000},
		Radius:   40,
		Stats:    Stats{HP: 1000, MaxHP: 1000},
	}
	w.entities[monster.ID] = monster
	placeEntity(explorer, 1000, 1000)
	placeEntity(target, 1500, 1000)
	placeEntity(minion, 1700, 1000)
	target.Stats.HP = 2000
	target.Stats.MaxHP = 2000
	target.Stats.MagicDefense = 0
	minion.Stats.HP = 1000
	minion.Stats.MaxHP = 1000
	minion.Stats.MagicDefense = 0

	w.ApplyInput("explorer", protocolPlayerInputCast(explorerRSkillID, 3000, 1000), 10, nil, 20)
	if got := explorer.Stats.MP; got != 200 {
		t.Fatalf("mp after r cast = %f, want 200", got)
	}
	w.Tick(29, 20)
	if got := countProjectilesByKind(w, "explorer_r"); got != 0 {
		t.Fatalf("r projectile before channel completes = %d, want 0", got)
	}
	w.Tick(30, 20)
	if got := countProjectilesByKind(w, "explorer_r"); got != 1 {
		t.Fatalf("r projectile after channel = %d, want 1", got)
	}
	if projectile := projectileByKind(w, "explorer_r"); projectile == nil || projectile.Range != 5000 {
		t.Fatalf("r projectile range = %+v, want map edge range 5000", projectile)
	}
	for tick := uint64(31); tick <= 50 && (target.Combat.LastDamage == 0 || minion.Combat.LastDamage == 0 || monster.Combat.LastDamage == 0); tick++ {
		w.Tick(tick, 20)
	}
	if target.Combat.LastDamage == 0 || minion.Combat.LastDamage == 0 || monster.Combat.LastDamage == 0 {
		t.Fatalf("r did not hit all targets: hero=%d minion=%d monster=%d", target.Combat.LastDamage, minion.Combat.LastDamage, monster.Combat.LastDamage)
	}

	if got := target.Combat.LastDamage; got != 350 {
		t.Fatalf("r hero damage = %d, want 350", got)
	}
	if got := minion.Combat.LastDamage; got != 150 {
		t.Fatalf("r minion damage = %d, want 150", got)
	}
	if got := monster.Combat.LastDamage; got != 150 {
		t.Fatalf("r monster damage = %d, want 150", got)
	}
	if got := explorer.Skills[explorerRSkillID].CooldownUntilTick; got != 2410 {
		t.Fatalf("r cooldown = %d, want 2410", got)
	}
	if got := explorer.Stats.AttackSpeedBonus; math.Abs(got-0.3) > 0.000001 {
		t.Fatalf("passive attack speed bonus = %f, want 0.3", got)
	}
}

func TestArcherRUsesMapEdgeRange(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	archer := w.entities[playerEntityID("archer")]
	archer.Stats.MP = 200
	learnSkill(archer, archerRSkillID, 1)
	placeEntity(archer, 1000, 1000)

	w.ApplyInput("archer", protocolPlayerInputCast(archerRSkillID, 100000, 1000), 10, nil, 20)
	w.Tick(14, 20)
	if got := countProjectilesByKind(w, "archer_crystal_arrow"); got != 0 {
		t.Fatalf("archer r projectile before release = %d, want 0", got)
	}
	w.Tick(15, 20)
	projectile := projectileByKind(w, "archer_crystal_arrow")
	if projectile == nil {
		t.Fatal("archer r projectile missing after release")
	}
	if got := projectile.Range; got != 5000 {
		t.Fatalf("archer r range = %f, want map edge range 5000", got)
	}
}

func TestFrostMageBasicAttackFiresProjectile(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(frostmageHeroID)
	if !ok {
		t.Fatal("frost mage hero not found")
	}
	w.SpawnHero("frostmage", hero, TeamBlue)
	player := w.entities[playerEntityID("frostmage")]
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}

	w.ApplyInput("frostmage", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, player, 20)

	if got := countProjectilesByKind(w, "basic_arrow"); got == 0 {
		t.Fatal("frost mage basic attack should fire projectile")
	}
}

func TestFireMageBasicAttackFiresProjectile(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(fireMageHeroID)
	if !ok {
		t.Fatal("fire mage hero not found")
	}
	w.SpawnHero("fire_mage", hero, TeamBlue)
	player := w.entities[playerEntityID("fire_mage")]
	target := w.entities["enemy:hero-1"]
	target.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}

	w.ApplyInput("fire_mage", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, player, 20)

	if got := countProjectilesByKind(w, "basic_arrow"); got == 0 {
		t.Fatal("fire mage basic attack should fire projectile")
	}
}

func TestTargetedProjectileDisappearsWhenTargetDies(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(archerHeroID)
	if !ok {
		t.Fatal("archer hero not found")
	}
	w.SpawnHero("archer", hero, TeamBlue)
	w.SpawnHero("red", testHeroConfig(), TeamRed)
	archer := w.entities[playerEntityID("archer")]
	target := w.entities[playerEntityID("red")]
	placeEntity(archer, 1000, 1000)
	placeEntity(target, 1500, 1000)

	w.ApplyInput("archer", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, archer, 20)
	if got := countProjectilesByKind(w, "basic_arrow"); got != 1 {
		t.Fatalf("basic arrows after release = %v, want 1", got)
	}

	w.killPlayer(target, 17, 20)

	if got := countProjectilesByKind(w, "basic_arrow"); got != 0 {
		t.Fatalf("basic arrows after target death = %v, want 0", got)
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
		t.Fatalf("auto attack should not fire on same tick as skill cast: got hit tick %v want %v", target.Combat.LastHitTick, qDamageTick)
	}
	tickSwordQRelease(t, w, player, 20)
	if player.Combat.NextAttackTick != 28 {
		t.Fatalf("next attack tick = %v, want 28", player.Combat.NextAttackTick)
	}
}

func TestSwordEAllowsSkillCastsOnlyNearDashEnd(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(swordHeroID)
	if !ok {
		t.Fatal("sword hero not found")
	}
	w.SpawnHero("sword", hero, TeamBlue)
	player := w.entities[playerEntityID("sword")]
	target := w.entities["enemy:hero-1"]
	placeEntity(player, 1000, 1000)
	placeEntity(target, 1200, 1000)
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)

	w.ApplyInput("sword", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.ApplyInput("sword", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 11, nil, 20)
	if player.Sword.QPending {
		t.Fatal("sword q should not cast before final 0.2s of e dash")
	}

	w.ApplyInput("sword", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 13, nil, 20)

	if !player.Sword.QPending {
		t.Fatal("sword q should be pending during eq combo")
	}
	if got := player.Sword.QForm; got != "circle" {
		t.Fatalf("sword q form = %q, want circle", got)
	}
}

func TestGunnerPassiveDamagesOnlyNewTargetsAndReducesWCooldown(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	gunner.Stats.Attack = 100
	target := w.entities["enemy:hero-1"]
	target.Stats.PhysicalDefense = 0

	if got := w.attackDamage(gunner, target, 10, 20); got != 150 {
		t.Fatalf("first target damage = %v, want 150", got)
	}
	gunner.Skills[gunnerWSkillID] = SkillState{SkillID: gunnerWSkillID, CooldownUntilTick: 100}
	w.onHeroBasicHit(gunner, target, 10, 20)
	if got := gunner.Skills[gunnerWSkillID].CooldownUntilTick; got != 60 {
		t.Fatalf("w cooldown after passive = %v, want 60", got)
	}
	if got := w.attackDamage(gunner, target, 20, 20); got != 100 {
		t.Fatalf("same target damage = %v, want 100", got)
	}
}

func TestGunnerPassiveUsesLowerMinionDamage(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	gunner.Level = MaxHeroLevel
	gunner.Stats.Attack = 100
	target := w.entities["minion:red-melee-1"]
	target.Stats.PhysicalDefense = 0

	if got := w.attackDamage(gunner, target, 10, 20); got != 150 {
		t.Fatalf("max-level minion damage = %v, want 150", got)
	}
}

func TestGunnerQBouncesToEnemyBehindTarget(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	gunner.Stats.Attack = 100
	first := w.entities["enemy:hero-1"]
	secondID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 1400, 1000)
	second := w.entities[secondID]
	placeEntity(gunner, 1000, 1000)
	placeEntity(first, 1200, 1000)
	first.Stats.PhysicalDefense = 0
	second.Stats.PhysicalDefense = 0
	learnSkill(gunner, "gunner_q", 1)

	w.ApplyInput("gunner", protocolPlayerInputCastTarget("gunner_q", first.ID, first.Position.X, first.Position.Y), 10, nil, 20)
	if got := first.Combat.LastDamage; got != 0 {
		t.Fatalf("first q damage before projectile hit = %v, want 0", got)
	}
	tickUntilDamage(t, w, first, 10, 40, 20)
	tickUntilDamage(t, w, second, 10, 40, 20)

	if got := first.Combat.LastDamage; got != 120 {
		t.Fatalf("first q damage = %v, want 120", got)
	}
	if got := second.Combat.LastDamage; got != 120 {
		t.Fatalf("second q damage = %v, want 120", got)
	}
	if got := gunner.Skills["gunner_q"].CooldownUntilTick; got != 150 {
		t.Fatalf("q cooldown = %v, want 150", got)
	}
}

func TestGunnerQDoesNotBounceWithoutEnemyBehindTarget(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	gunner.Stats.Attack = 100
	first := w.entities["enemy:hero-1"]
	secondID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 1000, 1400)
	second := w.entities[secondID]
	placeEntity(gunner, 1000, 1000)
	placeEntity(first, 1200, 1000)
	first.Stats.PhysicalDefense = 0
	second.Stats.PhysicalDefense = 0
	learnSkill(gunner, "gunner_q", 1)

	w.ApplyInput("gunner", protocolPlayerInputCastTarget("gunner_q", first.ID, first.Position.X, first.Position.Y), 10, nil, 20)
	tickUntilDamage(t, w, first, 10, 40, 20)

	if got := first.Combat.LastDamage; got != 120 {
		t.Fatalf("first q damage = %v, want 120", got)
	}
	if got := second.Combat.LastDamage; got != 0 {
		t.Fatalf("second q damage = %v, want 0", got)
	}
}

func TestGunnerQSecondHitCritsWhenFirstHitKills(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	gunner.Stats.Attack = 100
	first := w.entities["enemy:hero-1"]
	secondID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 1400, 1000)
	second := w.entities[secondID]
	placeEntity(gunner, 1000, 1000)
	placeEntity(first, 1200, 1000)
	first.Stats.HP = 100
	first.Stats.PhysicalDefense = 0
	second.Stats.PhysicalDefense = 0
	learnSkill(gunner, "gunner_q", 1)

	w.ApplyInput("gunner", protocolPlayerInputCastTarget("gunner_q", first.ID, first.Position.X, first.Position.Y), 10, nil, 20)
	tickUntilDamage(t, w, second, 10, 40, 20)

	if got := second.Combat.LastDamage; got != 240 {
		t.Fatalf("second q crit damage = %v, want 240", got)
	}
}

func TestGunnerWPassiveMoveSpeedScalesOutOfCombat(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	learnSkill(gunner, gunnerWSkillID, 1)
	baseMoveSpeed := gunner.Stats.MoveSpeed

	w.Tick(80, 20)
	if got := gunner.Stats.MoveSpeed; got != baseMoveSpeed+30 {
		t.Fatalf("w passive move speed after 4s = %f, want %f", got, baseMoveSpeed+30)
	}
	w.Tick(140, 20)
	if got := gunner.Stats.MoveSpeed; got != baseMoveSpeed+60 {
		t.Fatalf("w passive move speed after 7s = %f, want %f", got, baseMoveSpeed+60)
	}
	gunner.Combat.LastHitTick = 150
	w.Tick(151, 20)
	if got := gunner.Stats.MoveSpeed; got != baseMoveSpeed {
		t.Fatalf("w passive move speed after damage = %f, want %f", got, baseMoveSpeed)
	}
}

func TestGunnerWActiveGrantsFullMoveSpeedAndAttackSpeed(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	learnSkill(gunner, gunnerWSkillID, 1)
	baseMoveSpeed := gunner.Stats.MoveSpeed
	baseAttackSpeed := EffectiveAttackSpeedAtTick(gunner, 10)

	w.ApplyInput("gunner", protocolPlayerInputCast(gunnerWSkillID, gunner.Position.X, gunner.Position.Y), 10, nil, 20)

	if got := gunner.Stats.MoveSpeed; got != baseMoveSpeed+60 {
		t.Fatalf("w active move speed = %f, want %f", got, baseMoveSpeed+60)
	}
	if got := EffectiveAttackSpeedAtTick(gunner, 10); math.Abs(got-baseAttackSpeed*1.4) > 0.000001 {
		t.Fatalf("w active attack speed = %f, want %f", got, baseAttackSpeed*1.4)
	}
	if got := gunner.Skills[gunnerWSkillID].CooldownUntilTick; got != 250 {
		t.Fatalf("w cooldown = %v, want 250", got)
	}
	assertBuff(t, w.ActiveBuffs(gunner, 10), "gunner_w_active")
}

func TestGunnerEDealsAreaDamageSlowAndExpires(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	target := w.entities["enemy:hero-1"]
	placeEntity(gunner, 1000, 1000)
	placeEntity(target, 1200, 1000)
	gunner.Stats.AbilityPower = 100
	target.Stats.MagicDefense = 0
	startHP := target.Stats.HP
	learnSkill(gunner, "gunner_e", 1)

	w.ApplyInput("gunner", protocolPlayerInputCast("gunner_e", target.Position.X, target.Position.Y), 10, nil, 20)

	if got := target.Combat.LastDamage; got != 24 {
		t.Fatalf("first e tick damage = %v, want 24", got)
	}
	if math.Abs(target.Control.MoveSpeedSlow-0.46) > 0.000001 {
		t.Fatalf("e slow = %f, want 0.46", target.Control.MoveSpeedSlow)
	}
	if got := gunner.Skills["gunner_e"].CooldownUntilTick; got != 370 {
		t.Fatalf("e cooldown = %v, want 370", got)
	}
	if len(w.SkillEffects()) == 0 {
		t.Fatal("e should create a visible area effect")
	}

	for tick := uint64(11); tick <= 45; tick++ {
		w.Tick(tick, 20)
	}
	if got := startHP - target.Stats.HP; got != 192 {
		t.Fatalf("total e damage after 8 ticks = %v, want 192", got)
	}
	w.Tick(50, 20)
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "gunner_e" {
			t.Fatal("e effect should expire after 2s")
		}
	}
}

func TestGunnerRChannelsConeWaves(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	target := w.entities["enemy:hero-1"]
	insideID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 1800, 1100)
	inside := w.entities[insideID]
	sideID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 1000, 1800)
	side := w.entities[sideID]
	placeEntity(gunner, 1000, 1000)
	placeEntity(target, 1800, 1000)
	gunner.Stats.Attack = 100
	target.Stats.HP = 3000
	target.Stats.MaxHP = 3000
	target.Stats.PhysicalDefense = 0
	inside.Stats.HP = 3000
	inside.Stats.MaxHP = 3000
	inside.Stats.PhysicalDefense = 0
	side.Stats.PhysicalDefense = 0
	learnSkill(gunner, "gunner_r", 1)

	w.ApplyInput("gunner", protocolPlayerInputCast("gunner_r", 2000, 1000), 10, nil, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatalf("r should not damage before projectile flight, got %v", target.Combat.LastDamage)
	}
	for tick := uint64(11); tick <= 90; tick++ {
		w.Tick(tick, 20)
	}

	if got := 3000 - target.Stats.HP; got != 1120 {
		t.Fatalf("r total damage = %v, want 1120", got)
	}
	if got := 3000 - inside.Stats.HP; got != 1120 {
		t.Fatalf("r inside cone damage = %v, want 1120", got)
	}
	if got := side.Combat.LastDamage; got != 0 {
		t.Fatalf("side target damage = %v, want 0", got)
	}
	if got := gunner.Skills["gunner_r"].CooldownUntilTick; got != 2410 {
		t.Fatalf("r cooldown = %v, want 2410", got)
	}
	if gunner.Passive.GunnerRExpireTick != 0 {
		t.Fatalf("r state should be cleared: expire=%v", gunner.Passive.GunnerRExpireTick)
	}
}

func TestGunnerRChannelCancelsOnMove(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	placeEntity(gunner, 1000, 1000)
	learnSkill(gunner, "gunner_r", 1)

	w.ApplyInput("gunner", protocolPlayerInputCast("gunner_r", 2000, 1000), 10, nil, 20)
	firstWaveCount := countProjectilesByKind(w, "gunner_r")
	w.ApplyInput("gunner", protocolPlayerInputMove(1200, 1000), 11, nil, 20)

	if gunner.Passive.GunnerRExpireTick != 0 || gunner.Control.ActionLockedUntilTick != 0 {
		t.Fatalf("r channel state expire=%d lock=%d, want canceled", gunner.Passive.GunnerRExpireTick, gunner.Control.ActionLockedUntilTick)
	}
	if gunner.Intent.MoveTarget == nil {
		t.Fatal("move input should continue after canceling r")
	}
	for tick := uint64(12); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}
	if got := countProjectilesByKind(w, "gunner_r"); got != firstWaveCount {
		t.Fatalf("gunner r waves after cancel = %d, want %d", got, firstWaveCount)
	}
}

func TestGunnerRWaveCanCrit(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	target := w.entities["enemy:hero-1"]
	placeEntity(gunner, 1000, 1000)
	placeEntity(target, 1800, 1000)
	gunner.Stats.Attack = 100
	gunner.Stats.CritChance = 1
	target.Stats.PhysicalDefense = 0
	learnSkill(gunner, "gunner_r", 1)

	w.ApplyInput("gunner", protocolPlayerInputCast("gunner_r", 2000, 1000), 10, nil, 20)
	tickUntilDamage(t, w, target, 10, 30, 20)

	if got := target.Combat.LastDamage; got != 160 {
		t.Fatalf("r crit wave damage = %v, want 160", got)
	}
}

func TestGunnerRProjectileEffectCarriesConeSpread(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(gunnerHeroID)
	if !ok {
		t.Fatal("gunner hero not found")
	}
	w.SpawnHero("gunner", hero, TeamBlue)
	gunner := w.entities[playerEntityID("gunner")]
	placeEntity(gunner, 1000, 1000)
	learnSkill(gunner, "gunner_r", 1)

	w.ApplyInput("gunner", protocolPlayerInputCast("gunner_r", 2000, 1000), 10, nil, 20)
	w.Tick(12, 20)

	for _, effect := range w.SkillEffects() {
		if effect.Kind != "gunner_r" || effect.Speed == 0 {
			continue
		}
		if effect.Width != 33.75 || effect.Count != 7 || effect.End != gunner.Position {
			t.Fatalf("r projectile effect = %+v, want width 33.75 count 7 end %+v", effect, gunner.Position)
		}
		return
	}
	t.Fatal("gunner r projectile effect missing")
}
