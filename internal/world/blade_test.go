package world

import (
	"math"
	"testing"
)

func TestBladeQBloodlustAttackBonusScalesWithMissingHP(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", hero, TeamBlue)
	player := w.entities[playerEntityID("blade")]
	baseAttack := player.Stats.Attack
	learnSkill(player, "blade_q", 3)

	w.recalculatePlayerStats(player)
	if got, want := player.Stats.Attack, baseAttack+15; math.Abs(got-want) > 0.000001 {
		t.Fatalf("full hp attack = %f, want %f", got, want)
	}

	player.Stats.HP = player.Stats.MaxHP / 2
	w.recalculatePlayerStats(player)
	want := baseAttack + 15 + 50*0.25
	if got := player.Stats.Attack; math.Abs(got-want) > 0.000001 {
		t.Fatalf("half hp attack = %f, want %f", got, want)
	}
}

func TestBladeQConsumesRageAndHeals(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", hero, TeamBlue)
	player := w.entities[playerEntityID("blade")]
	learnSkill(player, "blade_q", 2)
	player.Stats.HP = player.Stats.MaxHP - 200
	player.Stats.MP = 40
	player.Stats.AbilityPower = 10

	w.ApplyInput("blade", protocolPlayerInputCast("blade_q", player.Position.X, player.Position.Y), 10, nil, 20)

	if player.Stats.MP != 0 {
		t.Fatalf("rage after q = %f, want 0", player.Stats.MP)
	}
	if got, want := player.Stats.HP, player.Stats.MaxHP-200+93; got != want {
		t.Fatalf("hp after q = %f, want %f", got, want)
	}
	if got, want := player.Skills["blade_q"].CooldownUntilTick, uint64(250); got != want {
		t.Fatalf("cooldown tick = %v, want %v", got, want)
	}
	assertSkillEffect(t, w.SkillEffects(), "blade_q_heal")
}

func TestBladeWReducesEnemyHeroAttackAndSlows(t *testing.T) {
	w := testWorld(t)
	bladeHero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", bladeHero, TeamBlue)
	w.SpawnHero("red", testHeroConfig(), TeamRed)
	blade := w.entities[playerEntityID("blade")]
	target := w.entities[playerEntityID("red")]
	placeEntity(blade, 3000, 3000)
	placeEntity(target, 3200, 3000)
	learnSkill(blade, "blade_w", 3)
	baseAttack := target.Stats.Attack

	w.ApplyInput("blade", protocolPlayerInputCast("blade_w", blade.Position.X, blade.Position.Y), 10, nil, 20)

	if got := target.Stats.Attack; got != baseAttack {
		t.Fatalf("target attack during windup = %f, want %f", got, baseAttack)
	}
	if got, want := blade.Skills["blade_w"].StacksExpireTick, uint64(16); got != want {
		t.Fatalf("w release tick = %v, want %v", got, want)
	}

	w.Tick(16, 20)

	if got, want := target.Stats.Attack, baseAttack-50; got != want {
		t.Fatalf("target attack = %f, want %f", got, want)
	}
	if got, want := target.Control.MoveSpeedSlow, 0.45; math.Abs(got-want) > 0.000001 {
		t.Fatalf("target slow = %f, want %f", got, want)
	}
	if got, want := blade.Skills["blade_w"].CooldownUntilTick, uint64(296); got != want {
		t.Fatalf("cooldown tick = %v, want %v", got, want)
	}

	w.Tick(96, 20)
	if got := target.Stats.Attack; got != baseAttack {
		t.Fatalf("target attack after expire = %f, want %f", got, baseAttack)
	}
}

func TestBladeWRequiresNearbyEnemyHero(t *testing.T) {
	w := testWorld(t)
	bladeHero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", bladeHero, TeamBlue)
	blade := w.entities[playerEntityID("blade")]
	minion := w.entities["minion:red-melee-1"]
	placeEntity(blade, 100, 100)
	placeEntity(minion, 120, 100)
	learnSkill(blade, "blade_w", 1)

	w.ApplyInput("blade", protocolPlayerInputCast("blade_w", blade.Position.X, blade.Position.Y), 10, nil, 20)

	if got := blade.Skills["blade_w"].CooldownUntilTick; got != 0 {
		t.Fatalf("cooldown tick = %v, want 0", got)
	}
	if got := minion.Control.MoveSpeedSlow; got != 0 {
		t.Fatalf("minion slow = %f, want 0", got)
	}
}

func TestBladeWWindupIgnoresAttackSpeedAndAbilityHaste(t *testing.T) {
	w := testWorld(t)
	bladeHero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", bladeHero, TeamBlue)
	w.SpawnHero("red", testHeroConfig(), TeamRed)
	blade := w.entities[playerEntityID("blade")]
	target := w.entities[playerEntityID("red")]
	placeEntity(blade, 3000, 3000)
	placeEntity(target, 3200, 3000)
	learnSkill(blade, "blade_w", 1)
	blade.Stats.AttackSpeedBonus = 10
	blade.Stats.AbilityHaste = 500

	w.ApplyInput("blade", protocolPlayerInputCast("blade_w", blade.Position.X, blade.Position.Y), 10, nil, 20)

	if got, want := blade.Skills["blade_w"].StacksExpireTick, uint64(16); got != want {
		t.Fatalf("w release tick = %v, want fixed %v", got, want)
	}
}

func TestBladeEDashesDamagesPathAndGainsRage(t *testing.T) {
	w := testWorld(t)
	bladeHero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", bladeHero, TeamBlue)
	blade := w.entities[playerEntityID("blade")]
	placeEntity(blade, 1000, 1000)
	target := w.entities["minion:red-melee-1"]
	offPath := w.entities["minion:red-ranged-1"]
	placeEntity(target, 1300, 1000)
	placeEntity(offPath, 1300, 1100)
	learnSkill(blade, "blade_e", 2)
	targetHP := target.Stats.HP
	offPathHP := offPath.Stats.HP

	w.ApplyInput("blade", protocolPlayerInputCast("blade_e", 1650, 1000), 10, nil, 20)

	if target.Stats.HP >= targetHP {
		t.Fatalf("path target hp = %v, want below %v", target.Stats.HP, targetHP)
	}
	if offPath.Stats.HP != offPathHP {
		t.Fatalf("off path hp = %v, want %v", offPath.Stats.HP, offPathHP)
	}
	if blade.Stats.MP != 2 {
		t.Fatalf("rage after e = %f, want 2", blade.Stats.MP)
	}
	if got, want := blade.Control.DashEnd, (Vector2{X: 1650, Y: 1000}); got != want {
		t.Fatalf("dash end = %+v, want %+v", got, want)
	}
	if got, want := blade.Skills["blade_e"].CooldownUntilTick, uint64(250); got != want {
		t.Fatalf("cooldown tick = %v, want %v", got, want)
	}
	assertSkillEffect(t, w.SkillEffects(), "blade_e_whirlwind")
}

func TestBladeBasicAttackCritRefundsECooldown(t *testing.T) {
	w := testWorld(t)
	bladeHero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", bladeHero, TeamBlue)
	blade := w.entities[playerEntityID("blade")]
	target := w.entities["minion:red-melee-1"]
	learnSkill(blade, "blade_e", 1)
	blade.Stats.CritChance = 1
	state := blade.Skills["blade_e"]
	state.CooldownUntilTick = 200
	blade.Skills["blade_e"] = state

	w.onHeroBasicHit(blade, target, 20, 20)

	if got, want := blade.Skills["blade_e"].CooldownUntilTick, uint64(185); got != want {
		t.Fatalf("minion crit refunded cooldown to %v, want %v", got, want)
	}

	heroTarget := w.entities["enemy:hero-1"]
	heroTarget.Team = TeamRed
	w.onHeroBasicHit(blade, heroTarget, 40, 20)
	if got, want := blade.Skills["blade_e"].CooldownUntilTick, uint64(155); got != want {
		t.Fatalf("hero crit refunded cooldown to %v, want %v", got, want)
	}
}

func TestBladeRGrantsRageAndPreventsDeath(t *testing.T) {
	w := testWorld(t)
	bladeHero, ok := w.heroes.Get(bladeHeroID)
	if !ok {
		t.Fatal("blade hero not found")
	}
	w.SpawnHero("blade", bladeHero, TeamBlue)
	blade := w.entities[playerEntityID("blade")]
	learnSkill(blade, "blade_r", 2)
	blade.Stats.HP = 100

	w.ApplyInput("blade", protocolPlayerInputCast("blade_r", blade.Position.X, blade.Position.Y), 10, nil, 20)

	if got, want := blade.Stats.MP, float64(75); got != want {
		t.Fatalf("rage after r = %f, want %f", got, want)
	}
	if got, want := blade.Control.UndyingRageUntil, uint64(110); got != want {
		t.Fatalf("undying until = %v, want %v", got, want)
	}
	if got, want := blade.Control.UndyingRageMinHP, 50.0; got != want {
		t.Fatalf("min hp = %f, want %f", got, want)
	}
	if got, want := blade.Skills["blade_r"].CooldownUntilTick, uint64(2010); got != want {
		t.Fatalf("cooldown tick = %v, want %v", got, want)
	}
	assertSkillEffect(t, w.SkillEffects(), "blade_r_rage")

	blade.Combat.LastHitTick = 20
	w.applyDamage(nil, blade, 1000, 20)
	if blade.Stats.HP != 50 {
		t.Fatalf("hp during r = %v, want 50", blade.Stats.HP)
	}
	if blade.Death.Dead {
		t.Fatal("blade should not die during r")
	}

	blade.Combat.LastHitTick = 111
	w.applyDamage(nil, blade, 1000, 20)
	if blade.Stats.HP != 0 {
		t.Fatalf("hp after r expired = %v, want 0", blade.Stats.HP)
	}
}
