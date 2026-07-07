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
		t.Fatalf("target hp = %d, start hp = %d; opposing player was not damaged", target.Stats.HP, startHP)
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
		t.Fatalf("enemy hero hp = %d, want below %d after basic attack", target.Stats.HP, startHP)
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
		t.Fatalf("warrior q cooldown with 100 haste = %d, want 90", got)
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
		t.Fatalf("next attack after first hit = %d, want %d", got, want)
	}

	w.ApplyInput("warrior", protocolPlayerInputCast(warriorQSkillID, target.Position.X, target.Position.Y), 16, nil, 20)
	w.Tick(16, 20)

	if got := player.Combat.PendingAttackTargetID; got != target.ID {
		t.Fatalf("pending attack after q reset = %q, want %q", got, target.ID)
	}
	if got, want := player.Combat.AttackReleaseTick, uint64(22); got != want {
		t.Fatalf("second attack release tick = %d, want %d", got, want)
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
	if len(target.Combat.DamageEvents) != 1 || !target.Combat.DamageEvents[0].BasicAttack {
		t.Fatalf("basic attack damage events = %+v", target.Combat.DamageEvents)
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
		t.Fatalf("basic arrows during attack windup = %d, want 0", got)
	}
	tickAttackRelease(t, w, player, 20)
	if got := countProjectilesByKind(w, "basic_arrow"); got == 0 {
		t.Fatal("ranged basic attack should fire projectile after windup")
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
		t.Fatalf("basic arrows after release = %d, want 1", got)
	}

	w.killPlayer(target, 17, 20)

	if got := countProjectilesByKind(w, "basic_arrow"); got != 0 {
		t.Fatalf("basic arrows after target death = %d, want 0", got)
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
		t.Fatalf("first target damage = %d, want 150", got)
	}
	gunner.Skills[gunnerWSkillID] = SkillState{SkillID: gunnerWSkillID, CooldownUntilTick: 100}
	w.onHeroBasicHit(gunner, target, 10, 20)
	if got := gunner.Skills[gunnerWSkillID].CooldownUntilTick; got != 60 {
		t.Fatalf("w cooldown after passive = %d, want 60", got)
	}
	if got := w.attackDamage(gunner, target, 20, 20); got != 100 {
		t.Fatalf("same target damage = %d, want 100", got)
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
		t.Fatalf("max-level minion damage = %d, want 150", got)
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
		t.Fatalf("first q damage before projectile hit = %d, want 0", got)
	}
	tickUntilDamage(t, w, first, 10, 40, 20)
	tickUntilDamage(t, w, second, 10, 40, 20)

	if got := first.Combat.LastDamage; got != 120 {
		t.Fatalf("first q damage = %d, want 120", got)
	}
	if got := second.Combat.LastDamage; got != 120 {
		t.Fatalf("second q damage = %d, want 120", got)
	}
	if got := gunner.Skills["gunner_q"].CooldownUntilTick; got != 150 {
		t.Fatalf("q cooldown = %d, want 150", got)
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
		t.Fatalf("first q damage = %d, want 120", got)
	}
	if got := second.Combat.LastDamage; got != 0 {
		t.Fatalf("second q damage = %d, want 0", got)
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
		t.Fatalf("second q crit damage = %d, want 240", got)
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
		t.Fatalf("w cooldown = %d, want 250", got)
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
		t.Fatalf("first e tick damage = %d, want 24", got)
	}
	if math.Abs(target.Control.MoveSpeedSlow-0.46) > 0.000001 {
		t.Fatalf("e slow = %f, want 0.46", target.Control.MoveSpeedSlow)
	}
	if got := gunner.Skills["gunner_e"].CooldownUntilTick; got != 370 {
		t.Fatalf("e cooldown = %d, want 370", got)
	}
	if len(w.SkillEffects()) == 0 {
		t.Fatal("e should create a visible area effect")
	}

	for tick := uint64(11); tick <= 45; tick++ {
		w.Tick(tick, 20)
	}
	if got := startHP - target.Stats.HP; got != 192 {
		t.Fatalf("total e damage after 8 ticks = %d, want 192", got)
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
		t.Fatalf("r should not damage before projectile flight, got %d", target.Combat.LastDamage)
	}
	for tick := uint64(11); tick <= 90; tick++ {
		w.Tick(tick, 20)
	}

	if got := 3000 - target.Stats.HP; got != 1120 {
		t.Fatalf("r total damage = %d, want 1120", got)
	}
	if got := 3000 - inside.Stats.HP; got != 1120 {
		t.Fatalf("r inside cone damage = %d, want 1120", got)
	}
	if got := side.Combat.LastDamage; got != 0 {
		t.Fatalf("side target damage = %d, want 0", got)
	}
	if got := gunner.Skills["gunner_r"].CooldownUntilTick; got != 2410 {
		t.Fatalf("r cooldown = %d, want 2410", got)
	}
	if gunner.Passive.GunnerRExpireTick != 0 {
		t.Fatalf("r state should be cleared: expire=%d", gunner.Passive.GunnerRExpireTick)
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
		t.Fatalf("r crit wave damage = %d, want 160", got)
	}
}
