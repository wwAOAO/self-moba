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
