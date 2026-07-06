package world

import (
	"l-battle/internal/config"
	"math"
	"testing"
)

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
	placeEntity(player, 3000, 3000)
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
	placeEntity(player, 3000, 3000)
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
	placeEntity(player, 3000, 3000)
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
	placeEntity(player, 3000, 3000)
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
	releaseTick := tickAttackRelease(t, w, player, 20)

	if got := startHP - target.Stats.HP; got != 146 {
		t.Fatalf("warrior q damage = %d, want 146", got)
	}
	wantSilencedUntil := releaseTick + secondsToTicks(1.5, 20)
	if target.Control.SilencedUntilTick != wantSilencedUntil {
		t.Fatalf("silenced until = %d, want %d", target.Control.SilencedUntilTick, wantSilencedUntil)
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
	placeEntity(player, 3000, 3000)
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
	placeEntity(player, 3000, 3000)
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

	if player.Combat.PendingAttackTargetID != "" {
		t.Fatalf("pending attack target = %q, want empty while spinning", player.Combat.PendingAttackTargetID)
	}
	if target.Stats.HP != hpAfterFirstSpin {
		t.Fatalf("target hp = %d, want unchanged %d immediately after blocked auto input", target.Stats.HP, hpAfterFirstSpin)
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
	if player.Skills[warriorRSkillID].CooldownUntilTick != 0 {
		t.Fatalf("r cooldown during windup = %d, want 0", player.Skills[warriorRSkillID].CooldownUntilTick)
	}
	w.Tick(19, 20)

	if target.Stats.HP != 350 {
		t.Fatalf("target hp = %d, want 350", target.Stats.HP)
	}
	if target.Combat.LastDamage != 250 {
		t.Fatalf("last damage = %d, want 250", target.Combat.LastDamage)
	}
	if player.Skills[warriorRSkillID].CooldownUntilTick != 2419 {
		t.Fatalf("r cooldown until = %d, want 2419", player.Skills[warriorRSkillID].CooldownUntilTick)
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

	w.onHeroKill(player, &Entity{Kind: EntityKindMeleeMinion})
	if math.Abs(player.Stats.PhysicalDefense-(baseArmor+0.2)) > 0.0001 {
		t.Fatalf("armor after minion kill = %f, want %f", player.Stats.PhysicalDefense, baseArmor+0.2)
	}
	if math.Abs(player.Stats.MagicDefense-(baseMagic+0.2)) > 0.0001 {
		t.Fatalf("magic resist after minion kill = %f, want %f", player.Stats.MagicDefense, baseMagic+0.2)
	}

	w.onHeroKill(player, &Entity{Kind: EntityKindEnemyHero})
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

	w.onHeroKill(player, &Entity{Kind: EntityKindMeleeMinion})

	if player.Stats.PhysicalDefense != baseArmor {
		t.Fatalf("armor = %f, want unchanged %f", player.Stats.PhysicalDefense, baseArmor)
	}
	if player.Stats.MagicDefense != baseMagic {
		t.Fatalf("magic resist = %f, want unchanged %f", player.Stats.MagicDefense, baseMagic)
	}
}
