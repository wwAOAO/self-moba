package world_test

import (
	"math"
	"testing"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	_ "l-battle/internal/world/heroes/doctor"
)

func TestDoctorPassiveResistsRootAndCanisterPickup(t *testing.T) {
	w := newDoctorPassiveWorld(t)
	w.SpawnHero("doc", doctorHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Stats.HP = 500
	doctor.Position = world.Vector2{X: 1000, Y: 1000}

	if doctor.Stats.HPRegen5 != 9.25 {
		t.Fatalf("hp regen/5 = %v, want 9.25", doctor.Stats.HPRegen5)
	}
	if applied := w.ApplyRoot(doctor, 80, 5, 10); applied {
		t.Fatal("root applied, want passive to resist it")
	}
	if doctor.Control.RootedUntilTick != 0 {
		t.Fatalf("root until = %v, want 0", doctor.Control.RootedUntilTick)
	}
	if got, want := doctor.Stats.HP, 465.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after passive = %v, want %v", got, want)
	}
	if doctor.Skills["doctor_passive"].CooldownUntilTick != 455 {
		t.Fatalf("passive cooldown = %v, want 455", doctor.Skills["doctor_passive"].CooldownUntilTick)
	}
	assertEffectKind(t, w.SkillEffects(), "doctor_canister")

	doctor.Stats.HP = 600
	w.Tick(6, 10)
	if got, want := doctor.Stats.HP, 640.185; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after pickup = %v, want %v", got, want)
	}
	if doctor.Skills["doctor_passive"].CooldownUntilTick != 305 {
		t.Fatalf("passive cooldown after pickup = %v, want 305", doctor.Skills["doctor_passive"].CooldownUntilTick)
	}
	assertNoEffectKind(t, w.SkillEffects(), "doctor_canister")
}

func TestDoctorCanisterCanBeCrushedByEnemyHero(t *testing.T) {
	w := newDoctorPassiveWorld(t)
	w.SpawnHero("doc", doctorHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Stats.HP = 500
	doctor.Position = world.Vector2{X: 1000, Y: 1000}

	w.ApplyStun(doctor, 80, 5, 10)
	doctor.Position = world.Vector2{X: 900, Y: 900}
	enemyID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1000, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	enemy := w.EntityByID(enemyID)
	enemy.Stats.HP = 1000
	w.Tick(6, 10)

	assertNoEffectKind(t, w.SkillEffects(), "doctor_canister")
	if got, want := doctor.Stats.HP, 465.185; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after enemy crush = %v, want %v", got, want)
	}
}

func TestDoctorQRefundsHalfCostOnHit(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 500
	doctor.Stats.HPRegen5 = 0
	state := doctor.Skills["doctor_q"]
	state.Level = 1
	doctor.Skills["doctor_q"] = state

	enemyID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	enemy := w.EntityByID(enemyID)
	enemy.Stats.HP = 1000
	enemy.Stats.MaxHP = 1000
	enemy.Stats.MagicDefense = 0

	w.ApplyInput("doc", doctorQCast(enemy), 10, nil, 20)
	if got, want := doctor.Stats.HP, 450.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after q cost = %v, want %v", got, want)
	}
	if got, want := doctor.Skills["doctor_q"].CooldownUntilTick, uint64(90); got != want {
		t.Fatalf("q cooldown = %v, want %v", got, want)
	}

	tickDoctorQHit(t, w, enemy, 11, 25, 20)
	if got, want := enemy.Combat.LastDamage, 200; got != want {
		t.Fatalf("q damage = %v, want %v", got, want)
	}
	if got, want := doctor.Stats.HP, 475.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after q hit refund = %v, want %v", got, want)
	}
	if enemy.Control.MoveSpeedSlow <= 0 {
		t.Fatal("q did not apply slow")
	}
}

func TestDoctorQRefundsFullCostOnKill(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 500
	doctor.Stats.HPRegen5 = 0
	state := doctor.Skills["doctor_q"]
	state.Level = 1
	doctor.Skills["doctor_q"] = state

	enemyID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	enemy := w.EntityByID(enemyID)
	enemy.Stats.HP = 70
	enemy.Stats.MaxHP = 1000
	enemy.Stats.MagicDefense = 0

	w.ApplyInput("doc", doctorQCast(enemy), 10, nil, 20)
	tickDoctorQHit(t, w, enemy, 11, 25, 20)
	if got, want := doctor.Stats.HP, 500.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after q kill refund = %v, want %v", got, want)
	}
	if enemy.Stats.HP != 0 {
		t.Fatalf("enemy hp = %v, want 0", enemy.Stats.HP)
	}
}

func TestDoctorWDealsPeriodicDamage(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 500
	state := doctor.Skills["doctor_w"]
	state.Level = 1
	doctor.Skills["doctor_w"] = state

	enemyID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1100, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	enemy := w.EntityByID(enemyID)
	enemy.Stats.HP = 1000
	enemy.Stats.MaxHP = 1000
	enemy.Stats.MagicDefense = 0

	w.ApplyInput("doc", doctorWCast(), 10, nil, 20)
	if got, want := doctor.Stats.HP, 475.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after w cost = %v, want %v", got, want)
	}
	if got, want := doctor.Skills["doctor_w"].CooldownUntilTick, uint64(350); got != want {
		t.Fatalf("w cooldown = %v, want %v", got, want)
	}
	assertEffectKind(t, w.SkillEffects(), "doctor_w")

	w.Tick(30, 20)
	if got, want := enemy.Combat.LastDamage, 20; got != want {
		t.Fatalf("w periodic damage = %v, want %v", got, want)
	}
}

func TestDoctorWRecastHealsAllGrayHealthOnHeroHit(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 500
	state := doctor.Skills["doctor_w"]
	state.Level = 1
	doctor.Skills["doctor_w"] = state

	enemyID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1100, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	enemy := w.EntityByID(enemyID)
	enemy.Stats.MagicDefense = 0

	w.ApplyInput("doc", doctorWCast(), 10, nil, 20)
	doctor.Combat.LastHitTick = 11
	w.ApplyTrueDamage(enemy, doctor, 100, 20)
	w.ApplyInput("doc", doctorWCast(), 12, nil, 20)

	if got, want := doctor.Stats.HP, 400.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after w hero heal = %v, want %v", got, want)
	}
	if got, want := enemy.Combat.LastDamage, 20; got != want {
		t.Fatalf("w burst damage = %v, want %v", got, want)
	}
	assertNoEffectKind(t, w.SkillEffects(), "doctor_w")
}

func TestDoctorWRecastHealsHalfGrayHealthOnOnlyNonHeroHit(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 500
	state := doctor.Skills["doctor_w"]
	state.Level = 1
	doctor.Skills["doctor_w"] = state

	minionID, ok := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamRed, 1100, 1000)
	if !ok {
		t.Fatal("failed to spawn minion")
	}
	minion := w.EntityByID(minionID)
	minion.Stats.MagicDefense = 0

	w.ApplyInput("doc", doctorWCast(), 10, nil, 20)
	doctor.Combat.LastHitTick = 11
	w.ApplyTrueDamage(minion, doctor, 100, 20)
	w.ApplyInput("doc", doctorWCast(), 12, nil, 20)

	if got, want := doctor.Stats.HP, 387.5; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after w non-hero heal = %v, want %v", got, want)
	}
	if got, want := minion.Combat.LastDamage, 20; got != want {
		t.Fatalf("w burst damage = %v, want %v", got, want)
	}
}

func TestDoctorEPassiveAddsAttackFromMaxHP(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	state := doctor.Skills["doctor_e"]
	state.Level = 1
	doctor.Skills["doctor_e"] = state

	w.RefreshPlayerStats(doctor)
	if got, want := doctor.Stats.Attack, 79.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("attack with e passive = %v, want %v", got, want)
	}
	if got, want := doctor.Stats.BonusAttack, 20.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("bonus attack with e passive = %v, want %v", got, want)
	}
}

func TestDoctorEEmpoweredAttackScalesWithMissingHPAndExpiresOnHit(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 510
	state := doctor.Skills["doctor_e"]
	state.Level = 1
	doctor.Skills["doctor_e"] = state
	w.RefreshPlayerStats(doctor)

	enemyID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1190, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	enemy := w.EntityByID(enemyID)
	enemy.Stats.HP = 1000
	enemy.Stats.MaxHP = 1000
	enemy.Stats.PhysicalDefense = 0

	w.ApplyInput("doc", doctorECast(), 10, nil, 20)
	if got, want := doctor.Stats.HP, 500.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after e cost = %v, want %v", got, want)
	}
	if got, want := doctor.Skills["doctor_e"].CooldownUntilTick, uint64(190); got != want {
		t.Fatalf("e cooldown = %v, want %v", got, want)
	}
	if got, want := doctor.Stats.AttackRange, 175.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("attack range during e = %v, want %v", got, want)
	}

	w.ApplyInput("doc", protocol.PlayerInput{Attack: &protocol.AttackInput{TargetID: enemy.ID}}, 11, nil, 20)
	w.Tick(11, 20)
	w.Tick(16, 20)
	if got, want := enemy.Combat.LastDamage, 181; got != want {
		t.Fatalf("empowered e damage = %v, want %v", got, want)
	}
	if got := doctor.Skills["doctor_e"].Stacks; got != 0 {
		t.Fatalf("e stacks after hit = %v, want 0", got)
	}
	if got, want := doctor.Stats.AttackRange, 125.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("attack range after e hit = %v, want %v", got, want)
	}
}

func TestDoctorRBuffsAndHealsOverTime(t *testing.T) {
	w := newDoctorQWorld(t)
	w.SpawnHero("doc", doctorQHeroConfig(), world.TeamBlue)
	doctor := w.EntityByID("player:doc")
	doctor.Position = world.Vector2{X: 1000, Y: 1000}
	doctor.Stats.HP = 500
	state := doctor.Skills["doctor_r"]
	state.Level = 1
	doctor.Skills["doctor_r"] = state
	w.RefreshPlayerStats(doctor)

	w.ApplyInput("doc", doctorRCast(), 10, nil, 20)
	if got, want := doctor.Stats.HP, 400.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after r cost = %v, want %v", got, want)
	}
	if got, want := doctor.Skills["doctor_r"].CooldownUntilTick, uint64(2210); got != want {
		t.Fatalf("r cooldown = %v, want %v", got, want)
	}
	if got, want := doctor.Skills["doctor_r"].StacksExpireTick, uint64(250); got != want {
		t.Fatalf("r expire tick = %v, want %v", got, want)
	}
	if got, want := doctor.Stats.MoveSpeed, 448.5; math.Abs(got-want) > 0.001 {
		t.Fatalf("move speed during r = %v, want %v", got, want)
	}
	if got, want := doctor.Stats.Attack, 67.85; math.Abs(got-want) > 0.001 {
		t.Fatalf("attack during r = %v, want %v", got, want)
	}

	w.Tick(30, 20)
	if got, want := doctor.Stats.HP, 480.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after first r heal = %v, want %v", got, want)
	}
	w.Tick(50, 20)
	if got, want := doctor.Stats.HP, 560.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("hp after second r heal = %v, want %v", got, want)
	}
	w.Tick(250, 20)
	if got := doctor.Skills["doctor_r"].Stacks; got != 0 {
		t.Fatalf("r stacks after expiry = %v, want 0", got)
	}
	if got, want := doctor.Stats.MoveSpeed, 345.0; math.Abs(got-want) > 0.001 {
		t.Fatalf("move speed after r = %v, want %v", got, want)
	}
}

func newDoctorPassiveWorld(t *testing.T) *world.World {
	t.Helper()
	heroes, err := config.NewHeroStore([]config.HeroConfig{doctorHeroConfig()})
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.NewSkillStore([]config.SkillConfig{{
		SkillID:    "doctor_passive",
		Name:       "自由之足",
		CooldownMS: 45000,
		Type:       "passive",
		Meta: map[string]float64{
			"cooldownSeconds":               45,
			"healthCostCurrentRatio":        0.07,
			"canisterHealMaxHPRatio":        0.04,
			"canisterCooldownRefundSeconds": 15,
			"canisterRadius":                55,
			"extraHPRegen5MaxHPRatio":       0.002,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return world.NewWorld(heroes, skills, nil, nil, nil)
}

func newDoctorQWorld(t *testing.T) *world.World {
	t.Helper()
	heroes, err := config.NewHeroStore([]config.HeroConfig{doctorQHeroConfig()})
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.NewSkillStore([]config.SkillConfig{
		{
			SkillID: "doctor_passive",
			Name:    "自由之足",
			Type:    "passive",
			Meta: map[string]float64{
				"extraHPRegen5MaxHPRatio": 0,
			},
		},
		{
			SkillID:    "doctor_q",
			Name:       "病毒屠刀",
			CooldownMS: 4000,
			Range:      975,
			Type:       "active",
			Meta: map[string]float64{
				"castWindupSeconds": 0.25,
				"projectileWidth":   120,
				"projectileSpeed":   2000,
				"slow":              0.4,
				"slowSeconds":       2,
			},
			MetaLists: map[string][]float64{
				"healthCost":       {50, 60, 70, 80, 90},
				"cooldownMs":       {4000, 4000, 4000, 4000, 4000},
				"currentHPRatio":   {0.2, 0.225, 0.25, 0.275, 0.3},
				"minimumDamage":    {80, 130, 180, 230, 280},
				"monsterMaxDamage": {300, 375, 450, 525, 600},
			},
		},
		{
			SkillID:    "doctor_w",
			Name:       "电击疗法",
			CooldownMS: 17000,
			Range:      325,
			Type:       "active",
			Meta: map[string]float64{
				"healthCostCurrentRatio": 0.05,
				"durationSeconds":        4,
				"damageIntervalSeconds":  1,
				"radius":                 325,
				"bonusHPRatio":           0.07,
				"heroGrayHealRatio":      1,
				"nonHeroGrayHealRatio":   0.5,
			},
			MetaLists: map[string][]float64{
				"cooldownMs":      {17000, 16500, 16000, 15500, 15000},
				"damagePerSecond": {20, 35, 50, 65, 80},
				"grayHealthRatio": {0.25, 0.3, 0.35, 0.4, 0.45},
				"burstDamage":     {20, 35, 50, 65, 80},
			},
		},
		{
			SkillID:    "doctor_e",
			Name:       "大力行医",
			CooldownMS: 9000,
			Type:       "active",
			Meta: map[string]float64{
				"durationSeconds":         4,
				"attackRangeBonus":        50,
				"totalADRatio":            1,
				"missingHPCapRatio":       0.7,
				"missingHPDamageBonusCap": 0.4,
			},
			MetaLists: map[string][]float64{
				"healthCost":       {10, 25, 40, 55, 70},
				"cooldownMs":       {9000, 8250, 7500, 6750, 6000},
				"maxHPAttackRatio": {0.02, 0.0225, 0.025, 0.0275, 0.03},
			},
		},
		{
			SkillID:    "doctor_r",
			Name:       "极限剂量",
			CooldownMS: 110000,
			Type:       "active",
			Meta: map[string]float64{
				"healthCostCurrentRatio": 0.2,
				"durationSeconds":        12,
				"healIntervalSeconds":    1,
			},
			MetaLists: map[string][]float64{
				"cooldownMs":         {110000, 100000, 90000},
				"moveSpeedBonus":     {0.3, 0.5, 0.7},
				"maxHPHealPerSecond": {0.08, 0.11, 0.14},
				"attackBonusRatio":   {0.15, 0.25, 0.35},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return world.NewWorld(heroes, skills, nil, nil, nil)
}

func doctorQCast(target *world.Entity) protocol.PlayerInput {
	return protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: "doctor_q", TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}}
}

func doctorWCast() protocol.PlayerInput {
	return protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: "doctor_w"}}
}

func doctorECast() protocol.PlayerInput {
	return protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: "doctor_e"}}
}

func doctorRCast() protocol.PlayerInput {
	return protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: "doctor_r"}}
}

func tickDoctorQHit(t *testing.T, w *world.World, target *world.Entity, from uint64, to uint64, tickRate int) {
	t.Helper()
	for tick := from; tick <= to; tick++ {
		w.Tick(tick, tickRate)
		if target.Combat.LastDamage > 0 || target.Stats.HP == 0 {
			return
		}
	}
	t.Fatalf("doctor q did not hit by tick %v", to)
}

func doctorHeroConfig() config.HeroConfig {
	return config.HeroConfig{
		HeroID: "doctor",
		Name:   "医生",
		Base: config.BaseStats{
			HP:              1000,
			HPRegen5:        7.25,
			Attack:          59,
			PhysicalDefense: 20,
			MagicDefense:    31,
			MoveSpeed:       345,
			AttackRange:     125,
			AttackSpeed:     0.625,
		},
		Radius: 18,
		Skills: config.HeroSkills{
			Passive: "doctor_passive",
			Q:       "doctor_q",
			W:       "doctor_w",
			E:       "doctor_e",
			R:       "doctor_r",
		},
	}
}

func doctorQHeroConfig() config.HeroConfig {
	hero := doctorHeroConfig()
	hero.Base.HPRegen5 = 0
	return hero
}

func assertEffectKind(t *testing.T, effects []world.SkillEffect, kind string) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			return
		}
	}
	t.Fatalf("missing effect %s", kind)
}

func assertNoEffectKind(t *testing.T, effects []world.SkillEffect, kind string) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			t.Fatalf("unexpected effect %s", kind)
		}
	}
}
