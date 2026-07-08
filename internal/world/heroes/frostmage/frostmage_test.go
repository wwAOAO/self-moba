package frostmage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestIceShardDamagesSlowsAndShattersBehindTarget(t *testing.T) {
	w, source := testWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 500
	source.Stats.AbilityPower = 100
	learnQ(source, 1)
	first := spawnEnemyHero(t, w, 1400, 1000)
	second := spawnEnemyHero(t, w, 1600, 1000)
	first.Stats.MagicDefense = 0
	second.Stats.MagicDefense = 0

	CastQ(w, source, protocol.CastInput{SkillID: qID, TargetX: 1825, TargetY: 1000}, source.Skills[qID], w.SkillConfig(qID), 10, 20)

	if got := source.Stats.MP; got != 440 {
		t.Fatalf("mp after q = %v, want 440", got)
	}
	if got := source.Skills[qID].CooldownUntilTick; got != 170 {
		t.Fatalf("q cooldown = %d, want 170", got)
	}
	w.Tick(14, 20)
	if first.Combat.LastDamage != 0 {
		t.Fatal("q hit before windup finished")
	}

	for tick := uint64(15); tick <= 25; tick++ {
		w.Tick(tick, 20)
	}

	if got := first.Combat.LastDamage; got != 150 {
		t.Fatalf("first damage = %d, want 150", got)
	}
	if got := first.Control.MoveSpeedSlow; got != 0.16 {
		t.Fatalf("first slow = %f, want 0.16", got)
	}
	if got := second.Combat.LastDamage; got != 150 {
		t.Fatalf("second damage = %d, want 150", got)
	}
	if second.Control.MoveSpeedSlow != 0 {
		t.Fatalf("second slow = %f, want 0", second.Control.MoveSpeedSlow)
	}
}

func TestRingOfFrostDamagesAndRootsNearbyEnemies(t *testing.T) {
	w, source := testWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.AbilityPower = 100
	learnW(source, 2)
	near := spawnEnemyHero(t, w, 1300, 1000)
	far := spawnEnemyHero(t, w, 1600, 1000)
	near.Stats.MagicDefense = 0
	far.Stats.MagicDefense = 0

	CastW(w, source, protocol.CastInput{SkillID: wID}, source.Skills[wID], w.SkillConfig(wID), 10, 20)

	if got := source.Skills[wID].CooldownUntilTick; got != 270 {
		t.Fatalf("w cooldown = %d, want 270", got)
	}
	if got := near.Combat.LastDamage; got != 175 {
		t.Fatalf("near damage = %d, want 175", got)
	}
	if got := near.Control.RootedUntilTick; got != 34 {
		t.Fatalf("near root until = %d, want 34", got)
	}
	if far.Combat.LastDamage != 0 {
		t.Fatalf("far damage = %d, want 0", far.Combat.LastDamage)
	}
}

func TestGlacialPathDamagesAlongPathAndRecastsToClaw(t *testing.T) {
	w, source := testWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 500
	source.Stats.AbilityPower = 100
	learnE(source, 1)
	target := spawnEnemyHero(t, w, 1200, 1000)
	target.Stats.MagicDefense = 0

	CastE(w, source, protocol.CastInput{SkillID: eID, TargetX: 2050, TargetY: 1000}, source.Skills[eID], w.SkillConfig(eID), 10, 20)

	if got := source.Stats.MP; got != 460 {
		t.Fatalf("mp after e = %v, want 460", got)
	}
	if got := source.Skills[eID].CooldownUntilTick; got != 490 {
		t.Fatalf("e cooldown = %d, want 490", got)
	}
	w.Tick(11, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatal("e hit before windup finished")
	}
	for tick := uint64(12); tick <= 14; tick++ {
		w.Tick(tick, 20)
	}
	if got := target.Combat.LastDamage; got != 130 {
		t.Fatalf("e damage = %d, want 130", got)
	}
	before := source.Position
	for tick := uint64(15); tick <= 22; tick++ {
		w.Tick(tick, 20)
	}
	if !RecastE(w, source, protocol.CastInput{SkillID: eID}, source.Skills[eID], w.SkillConfig(eID), 22, 20) {
		t.Fatal("e recast was not handled")
	}
	if source.Position.X <= before.X {
		t.Fatalf("source x after recast = %f, want > %f", source.Position.X, before.X)
	}
	if source.Passive.FrostEProjectileID != "" {
		t.Fatalf("e projectile state = %q, want empty", source.Passive.FrostEProjectileID)
	}
}

func TestFrozenTombEnemyDamagesAndStunsAfterWindup(t *testing.T) {
	w, source := testWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 500
	source.Stats.AbilityPower = 100
	learnR(source, 2)
	target := spawnEnemyHero(t, w, 1400, 1000)
	target.Stats.MagicDefense = 0

	CastR(w, source, protocol.CastInput{SkillID: rID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, source.Skills[rID], w.SkillConfig(rID), 10, 20)

	if got := source.Stats.MP; got != 400 {
		t.Fatalf("mp after r = %v, want 400", got)
	}
	if got := source.Skills[rID].CooldownUntilTick; got != 2010 {
		t.Fatalf("r cooldown = %d, want 2010", got)
	}
	w.Tick(17, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatal("r hit before windup finished")
	}
	w.Tick(18, 20)
	if got := target.Combat.LastDamage; got != 325 {
		t.Fatalf("r damage = %d, want 325", got)
	}
	if got := target.Control.StunnedUntilTick; got != 48 {
		t.Fatalf("r stun until = %d, want 48", got)
	}
}

func TestFrozenTombSelfHealsBecomesUntargetableAndExplodes(t *testing.T) {
	w, source := testWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 500
	source.Stats.AbilityPower = 100
	source.Stats.HP = 250
	learnR(source, 1)
	near := spawnEnemyHero(t, w, 1300, 1000)
	far := spawnEnemyHero(t, w, 1700, 1000)
	near.Stats.MagicDefense = 0
	far.Stats.MagicDefense = 0

	CastR(w, source, protocol.CastInput{SkillID: rID, TargetID: source.ID, TargetX: source.Position.X, TargetY: source.Position.Y}, source.Skills[rID], w.SkillConfig(rID), 10, 20)

	if source.Control.UntargetableUntilTick != 60 {
		t.Fatalf("self r untargetable until = %d, want 60", source.Control.UntargetableUntilTick)
	}
	if source.Stats.DamageReduce != 1 {
		t.Fatalf("self r damage reduce = %f, want 1", source.Stats.DamageReduce)
	}
	damage := w.MagicDamageAfterResistance(near, source, 200, 11)
	source.Combat.LastHitTick = 11
	w.ApplyMagicDamage(near, source, damage, 20)
	if source.Stats.HP < 250 {
		t.Fatalf("self r hp after damage = %f, want >= 250", source.Stats.HP)
	}
	for tick := uint64(11); tick <= 60; tick++ {
		w.Tick(tick, 20)
	}
	if source.Stats.HP <= 250 {
		t.Fatalf("self r hp after heal = %f, want > 250", source.Stats.HP)
	}
	if source.Stats.DamageReduce != 0 {
		t.Fatalf("self r damage reduce after end = %f, want 0", source.Stats.DamageReduce)
	}
	if got := near.Combat.LastDamage; got != 225 {
		t.Fatalf("self r explosion damage = %d, want 225", got)
	}
	if got := near.Control.MoveSpeedSlow; got != 0.3 {
		t.Fatalf("self r slow = %f, want 0.3", got)
	}
	if far.Combat.LastDamage != 0 {
		t.Fatalf("far damage = %d, want 0", far.Combat.LastDamage)
	}
}

func TestIcebornSubjugationServantSlowsHeroAndExplodes(t *testing.T) {
	w, source := testWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.AbilityPower = 100

	dead := spawnEnemyHero(t, w, 1100, 1000)
	hero := spawnEnemyHero(t, w, 1200, 1000)
	minionID, ok := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamRed, 1050, 1000)
	if !ok {
		t.Fatal("spawn minion failed")
	}
	minion := w.EntityByID(minionID)
	hero.Stats.MagicDefense = 0
	dead.Stats.HP = 0
	dead.Combat.LastHitTick = 10

	w.ApplyKillReward(source, dead)
	w.KillPlayer(dead, 10, 20)

	if got := len(source.Passive.FrostServants); got != 1 {
		t.Fatalf("servants = %d, want 1", got)
	}

	w.Tick(11, 20)

	if hero.Control.MoveSpeedSlow != 0.3 {
		t.Fatalf("hero slow = %f, want 0.3", hero.Control.MoveSpeedSlow)
	}
	if minion.Control.MoveSpeedSlow != 0 {
		t.Fatalf("minion slow = %f, want 0", minion.Control.MoveSpeedSlow)
	}

	w.Tick(90, 20)

	if got := hero.Combat.LastDamage; got != 50 {
		t.Fatalf("explosion damage = %d, want 50", got)
	}
	if got := len(source.Passive.FrostServants); got != 0 {
		t.Fatalf("servants after explosion = %d, want 0", got)
	}
}

func spawnEnemyHero(t *testing.T, w *world.World, x float64, y float64) *world.Entity {
	t.Helper()
	id, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, x, y)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}
	target := w.EntityByID(id)
	if target == nil {
		t.Fatal("enemy hero not found")
	}
	return target
}

func learnQ(source *world.Entity, level int) {
	state := source.Skills[qID]
	state.Level = level
	source.Skills[qID] = state
}

func learnW(source *world.Entity, level int) {
	state := source.Skills[wID]
	state.Level = level
	source.Skills[wID] = state
}

func learnE(source *world.Entity, level int) {
	state := source.Skills[eID]
	state.Level = level
	source.Skills[eID] = state
}

func learnR(source *world.Entity, level int) {
	state := source.Skills[rID]
	state.Level = level
	source.Skills[rID] = state
}

func testWorld(t *testing.T) (*world.World, *world.Entity) {
	t.Helper()
	heroes, err := config.NewHeroStore([]config.HeroConfig{{
		HeroID: heroID,
		Base: config.BaseStats{
			HP:          506,
			MP:          304,
			Attack:      50,
			MoveSpeed:   325,
			AttackRange: 550,
			AttackSpeed: 0.625,
		},
		Radius: 16,
		Skills: config.HeroSkills{
			Passive: passiveID,
			Q:       qID,
			W:       wID,
			E:       eID,
			R:       rID,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.NewSkillStore([]config.SkillConfig{
		{
			SkillID: passiveID,
			Type:    "passive",
			Meta: map[string]float64{
				"spawnRadius":     1350,
				"seekRadius":      700,
				"slowRadius":      450,
				"slow":            0.3,
				"slowSeconds":     0.25,
				"durationSeconds": 4,
				"moveSpeed":       325,
				"explosionRadius": 450,
				"apRatio":         0.5,
			},
		},
		{
			SkillID:    qID,
			Range:      825,
			CooldownMS: 8000,
			Type:       "projectile_line",
			Meta: map[string]float64{
				"apRatio":               0.8,
				"castWindupSeconds":     0.25,
				"projectileSpeed":       2200,
				"projectileRadius":      75,
				"shardProjectileRadius": 90,
				"shatterDistance":       700,
				"shatterSearchRadius":   100,
				"effectRange":           950,
				"slowSeconds":           1.5,
			},
			MetaLists: map[string][]float64{
				"manaCost":   {60, 63, 66, 69, 72},
				"cooldownMs": {8000, 7000, 6000, 5000, 4000},
				"baseDamage": {70, 100, 130, 160, 190},
				"slow":       {0.16, 0.19, 0.22, 0.25, 0.28},
			},
		},
		{
			SkillID:    wID,
			Range:      450,
			CooldownMS: 14000,
			Type:       "aoe_root",
			Meta: map[string]float64{
				"apRatio": 0.7,
				"radius":  450,
			},
			MetaLists: map[string][]float64{
				"cooldownMs":  {14000, 13000, 12000, 11000, 10000},
				"baseDamage":  {70, 105, 140, 175, 210},
				"rootSeconds": {1.1, 1.2, 1.3, 1.4, 1.5},
			},
		},
		{
			SkillID:    eID,
			Range:      1050,
			CooldownMS: 24000,
			Type:       "projectile_line",
			Meta: map[string]float64{
				"manaCost":           40,
				"apRatio":            0.6,
				"castWindupSeconds":  0.1,
				"recastDelaySeconds": 0.5,
				"durationSeconds":    1.25,
				"projectileRadius":   90,
				"projectileMinSpeed": 400,
				"projectileMaxSpeed": 1600,
			},
			MetaLists: map[string][]float64{
				"cooldownMs": {24000, 21000, 18000, 15000, 12000},
				"baseDamage": {70, 105, 140, 175, 210},
			},
		},
		{
			SkillID:    rID,
			Range:      550,
			CooldownMS: 120000,
			Type:       "target_or_self",
			Meta: map[string]float64{
				"manaCost":               100,
				"apRatio":                0.75,
				"selfHealAPRatio":        0.25,
				"enemyCastWindupSeconds": 0.375,
				"selfDurationSeconds":    2.5,
				"stunSeconds":            1.5,
				"selfRadius":             550,
				"slowSeconds":            1.5,
				"targetPickPadding":      80,
			},
			MetaLists: map[string][]float64{
				"cooldownMs": {120000, 100000, 80000},
				"baseDamage": {150, 250, 350},
				"selfHeal":   {90, 140, 190},
				"slow":       {0.3, 0.45, 0.75},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(heroes, skills, nil, nil, nil)
	hero, ok := heroes.Get(heroID)
	if !ok {
		t.Fatal("frost mage hero not found")
	}
	w.SpawnHero("frost", hero, world.TeamBlue)
	source := w.EntityByID("player:frost")
	if source == nil {
		t.Fatal("frost mage entity not found")
	}
	return w, source
}
