package crossbowman

import (
	"math"
	"testing"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func TestTumbleMovesSpendsManaResetsAttackAndDefersCooldown(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 100
	source.Combat.NextAttackTick = 99
	source.Combat.PendingAttackTargetID = "target"
	source.Combat.AttackReleaseTick = 80
	source.Intent.MoveTarget = &world.Vector2{X: 5000, Y: 5000}
	state := source.Skills[qID]
	state.Level = 3
	source.Skills[qID] = state

	CastQ(w, source, protocol.CastInput{TargetX: 2000, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)

	if got, want := source.Position, (world.Vector2{X: 1300, Y: 1000}); got != want {
		t.Fatalf("tumble position = %+v, want %+v", got, want)
	}
	if source.Stats.MP != 70 {
		t.Fatalf("mana after tumble = %v, want 70", source.Stats.MP)
	}
	if source.Combat.NextAttackTick != 10 || source.Combat.PendingAttackTargetID != "" || source.Combat.AttackReleaseTick != 0 {
		t.Fatalf("attack reset = %+v, want next attack at 10 with no pending windup", source.Combat)
	}
	if source.Intent.MoveTarget != nil {
		t.Fatalf("successful tumble kept old move target %+v", *source.Intent.MoveTarget)
	}
	if got := source.Skills[qID].CooldownUntilTick; got != 0 {
		t.Fatalf("cooldown during empower = %d, want 0", got)
	}
	if got, want := source.Crossbowman.TumblePendingCooldownTicks, uint64(80); got != want {
		t.Fatalf("pending rank 3 cooldown = %d, want %d", got, want)
	}
	if got, want := source.Crossbowman.TumbleEmpowerUntilTick, uint64(130); got != want {
		t.Fatalf("empower expiry = %d, want %d", got, want)
	}
	if source.Control.InvisibleUntilTick != 0 {
		t.Fatalf("normal tumble invisibility = %d, want 0", source.Control.InvisibleUntilTick)
	}
	if !hasEffect(w.SkillEffects(), "crossbowman_tumble") {
		t.Fatal("tumble effect was not created")
	}
}

func TestTumbleClampsAtMapBoundaryAndRejectsInsufficientMana(t *testing.T) {
	w, source := tumbleTestWorld(t)
	state := source.Skills[qID]
	state.Level = 1
	source.Skills[qID] = state
	source.Position = world.Vector2{X: 7900, Y: 1000}
	source.Stats.MP = 30

	CastQ(w, source, protocol.CastInput{TargetX: 9000, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)
	if got, want := source.Position, (world.Vector2{X: 8000, Y: 1000}); got != want {
		t.Fatalf("boundary tumble position = %+v, want %+v", got, want)
	}

	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 29
	source.Crossbowman.TumbleEmpowerUntilTick = 0
	source.Crossbowman.TumbleLevel = 0
	source.Crossbowman.TumblePendingCooldownTicks = 0
	oldMoveTarget := &world.Vector2{X: 4000, Y: 1000}
	source.Intent.MoveTarget = oldMoveTarget
	state.CooldownUntilTick = 0
	CastQ(w, source, protocol.CastInput{TargetX: 2000, TargetY: 1000}, state, w.SkillConfig(qID), 20, 20)
	if got, want := source.Position, (world.Vector2{X: 1000, Y: 1000}); got != want {
		t.Fatalf("low mana tumble moved to %+v, want %+v", got, want)
	}
	if source.Intent.MoveTarget != oldMoveTarget {
		t.Fatalf("failed tumble changed move target to %+v, want original target", source.Intent.MoveTarget)
	}
}

func TestTumbleUsesCursorDirectionInsteadOfSelectedEnemy(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 100
	state := source.Skills[qID]
	state.Level = 1
	source.Skills[qID] = state

	CastQ(w, source, protocol.CastInput{
		TargetID: "enemy-to-the-east",
		TargetX:  1000,
		TargetY:  2000,
	}, state, w.SkillConfig(qID), 10, 20)

	if got, want := source.Position, (world.Vector2{X: 1000, Y: 1300}); got != want {
		t.Fatalf("tumble with selected enemy moved to %+v, want cursor-directed %+v", got, want)
	}
}

func TestTumbleDamageRatiosExpiryAndConsumption(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Stats.Attack = 100
	wants := []int{75, 85, 95, 105, 115}
	for level, want := range wants {
		source.Crossbowman.TumbleLevel = level + 1
		source.Crossbowman.TumbleEmpowerUntilTick = 100
		if got := QBonusPhysicalDamage(w, source, nil, 50, 20); got != want {
			t.Fatalf("rank %d bonus damage = %d, want %d", level+1, got, want)
		}
	}

	source.Crossbowman.TumbleLevel = 1
	source.Crossbowman.TumbleEmpowerUntilTick = 100
	ConsumeQ(w, source, nil, 99, 20)
	if source.Crossbowman.TumbleEmpowerUntilTick != 0 || source.Crossbowman.TumbleLevel != 0 {
		t.Fatalf("consumed tumble state = %+v, want cleared", source.Crossbowman)
	}

	source.Crossbowman.TumbleLevel = 1
	source.Crossbowman.TumbleEmpowerUntilTick = 100
	if got := QBonusPhysicalDamage(w, source, nil, 100, 20); got != 0 {
		t.Fatalf("expired tumble bonus = %d, want 0", got)
	}
}

func TestTumbleBonusDamageDoesNotReceiveCritMultiplier(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.Attack = 100
	source.Stats.CritChance = 1
	source.Stats.AttackSpeed = 1
	source.Stats.AttackWindupSeconds = 0.1
	source.Stats.MP = 100
	state := source.Skills[qID]
	state.Level = 1
	source.Skills[qID] = state

	CastQ(w, source, protocol.CastInput{TargetX: 2000, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)
	targetID, ok := w.SpawnObject(world.EntityKindDummy, world.TeamNeutral, 1450, 1000)
	if !ok {
		t.Fatal("failed to spawn target dummy")
	}
	target := w.EntityByID(targetID)
	target.Stats.PhysicalDefense = 0
	w.ApplyInput("vayne", protocol.PlayerInput{Attack: &protocol.AttackInput{TargetID: targetID}}, 11, nil, 20)

	for tick := uint64(11); tick <= 20 && target.Combat.LastDamage == 0; tick++ {
		w.Tick(tick, 20)
	}
	if got, want := target.Combat.LastDamage, 275; got != want {
		t.Fatalf("critical empowered attack damage = %d, want %d (200 crit + 75 Q)", got, want)
	}
	if source.Crossbowman.TumbleEmpowerUntilTick != 0 {
		t.Fatal("empowerment was not consumed by the basic attack")
	}
}

func TestTumbleUltimateCooldownReductionAndInvisibility(t *testing.T) {
	for ultimateLevel, wantCooldownTicks := range []uint64{56, 48, 40} {
		w, source := tumbleTestWorld(t)
		source.Position = world.Vector2{X: 1000, Y: 1000}
		source.Stats.MP = 100
		state := source.Skills[qID]
		state.Level = 3
		source.Skills[qID] = state
		source.Crossbowman.UltimateUntilTick = 100
		source.Crossbowman.UltimateLevel = ultimateLevel + 1

		CastQ(w, source, protocol.CastInput{TargetX: 2000, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)

		if got := source.Skills[qID].CooldownUntilTick; got != 0 {
			t.Fatalf("R rank %d cooldown during empower = %d, want 0", ultimateLevel+1, got)
		}
		if got := source.Crossbowman.TumblePendingCooldownTicks; got != wantCooldownTicks {
			t.Fatalf("R rank %d pending tumble cooldown = %d, want %d", ultimateLevel+1, got, wantCooldownTicks)
		}
		if got, want := source.Control.InvisibleUntilTick, uint64(30); got != want {
			t.Fatalf("R rank %d invisibility expiry = %d, want %d", ultimateLevel+1, got, want)
		}
		enemy := &world.Entity{ID: "enemy", Team: world.TeamRed, Stats: world.Stats{HP: 100}}
		if world.CanAttackTarget(enemy, source) {
			t.Fatalf("R rank %d invisible crossbowman remained targetable", ultimateLevel+1)
		}
		w.Tick(30, 20)
		if source.Control.InvisibleUntilTick != 0 {
			t.Fatalf("R rank %d invisibility did not expire", ultimateLevel+1)
		}
		ConsumeQ(w, source, nil, 31, 20)
		if got, want := source.Skills[qID].CooldownUntilTick, uint64(31)+wantCooldownTicks; got != want {
			t.Fatalf("R rank %d cooldown after empower = %d, want %d", ultimateLevel+1, got, want)
		}
	}
}

func TestTumbleCooldownByRank(t *testing.T) {
	for level, seconds := range []float64{6, 5, 4, 3, 2} {
		w, source := tumbleTestWorld(t)
		source.Position = world.Vector2{X: 1000, Y: 1000}
		source.Stats.MP = 100
		state := source.Skills[qID]
		state.Level = level + 1
		source.Skills[qID] = state
		CastQ(w, source, protocol.CastInput{TargetX: 2000, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)
		cooldownTicks := uint64(math.Ceil(seconds * 20))
		if got := source.Skills[qID].CooldownUntilTick; got != 0 {
			t.Fatalf("rank %d cooldown during empower = %d, want 0", level+1, got)
		}
		if got := source.Crossbowman.TumblePendingCooldownTicks; got != cooldownTicks {
			t.Fatalf("rank %d pending cooldown = %d, want %d", level+1, got, cooldownTicks)
		}
		position := source.Position
		mana := source.Stats.MP
		CastQ(w, source, protocol.CastInput{TargetX: 1000, TargetY: 2000}, source.Skills[qID], w.SkillConfig(qID), 11, 20)
		if source.Position != position || source.Stats.MP != mana {
			t.Fatalf("rank %d recast during empower changed position/mana", level+1)
		}
		ConsumeQ(w, source, nil, 12, 20)
		want := uint64(12) + cooldownTicks
		if got := source.Skills[qID].CooldownUntilTick; got != want {
			t.Fatalf("rank %d cooldown after empower = %d, want %d", level+1, got, want)
		}
	}
}

func TestTumbleCooldownStartsWhenEmpowerExpires(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 100
	state := source.Skills[qID]
	state.Level = 1
	source.Skills[qID] = state
	CastQ(w, source, protocol.CastInput{TargetX: 2000, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)

	w.Tick(129, 20)
	if got := source.Skills[qID].CooldownUntilTick; got != 0 {
		t.Fatalf("cooldown before empower expiry = %d, want 0", got)
	}
	w.Tick(130, 20)
	if got, want := source.Skills[qID].CooldownUntilTick, uint64(250); got != want {
		t.Fatalf("cooldown after natural empower expiry = %d, want %d", got, want)
	}
	if source.Crossbowman.TumblePendingCooldownTicks != 0 {
		t.Fatalf("pending cooldown after expiry = %d, want 0", source.Crossbowman.TumblePendingCooldownTicks)
	}
}

func tumbleTestWorld(t *testing.T) (*world.World, *world.Entity) {
	t.Helper()
	skills, err := config.NewSkillStore([]config.SkillConfig{
		{
			SkillID: passiveID,
			Range:   2000,
			Meta: map[string]float64{
				"moveSpeedBonus":         30,
				"ultimateMoveSpeedBonus": 90,
				"lingerSeconds":          2,
			},
		},
		{
			SkillID:    qID,
			CooldownMS: 6000,
			Range:      300,
			Meta: map[string]float64{
				"manaCost":               30,
				"empowerDurationSeconds": 6,
				"invisibilitySeconds":    1,
				"effectSeconds":          0.35,
			},
			MetaLists: map[string][]float64{
				"cooldownMs":                {6000, 5000, 4000, 3000, 2000},
				"totalAdRatio":              {0.75, 0.85, 0.95, 1.05, 1.15},
				"ultimateCooldownReduction": {0.3, 0.4, 0.5},
			},
		},
		{
			SkillID: wID,
			Type:    "passive",
			Meta: map[string]float64{
				"stackDurationSeconds": 5,
				"procEffectSeconds":    0.45,
			},
			MetaLists: map[string][]float64{
				"minimumDamage":    {50, 65, 80, 95, 110},
				"maxHealthRatio":   {0.06, 0.07, 0.08, 0.09, 0.1},
				"monsterDamageCap": {140, 155, 170, 185, 200},
			},
		},
		{
			SkillID:    eID,
			CooldownMS: 20000,
			Range:      550,
			Type:       "targeted_projectile",
			Meta: map[string]float64{
				"manaCost":                  90,
				"castWindupSeconds":         0.2,
				"bonusAdRatio":              0.5,
				"projectileSpeed":           2200,
				"projectileRadius":          18,
				"projectileLifetimeSeconds": 2,
				"knockbackDistance":         470,
				"knockbackSeconds":          0.25,
				"stunSeconds":               1.5,
				"impactEffectSeconds":       0.5,
			},
			MetaLists: map[string][]float64{
				"cooldownMs": {20000, 18000, 16000, 14000, 12000},
				"baseDamage": {50, 85, 120, 155, 190},
			},
		},
		{
			SkillID:    rID,
			CooldownMS: 100000,
			Type:       "self_buff",
			Meta: map[string]float64{
				"manaCost":          80,
				"damageMarkSeconds": 3,
				"extensionSeconds":  4,
			},
			MetaLists: map[string][]float64{
				"cooldownMs":      {100000, 85000, 70000},
				"durationSeconds": {8, 10, 12},
				"bonusAttack":     {35, 50, 65},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	hero := config.HeroConfig{
		HeroID:   heroID,
		Resource: "mp",
		Radius:   16,
		Base: config.BaseStats{
			HP:                  515,
			MP:                  231.8,
			Attack:              60,
			MoveSpeed:           330,
			AttackRange:         550,
			AttackSpeed:         0.658,
			AttackWindupSeconds: 0.267,
			AttackSpeedRatio:    0.658,
		},
		Skills: config.HeroSkills{Passive: passiveID, Q: qID, W: wID, E: eID, R: rID},
	}
	heroes, err := config.NewHeroStore([]config.HeroConfig{hero})
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(heroes, skills, nil, nil, nil)
	w.SpawnHero("vayne", hero, world.TeamBlue)
	source := w.EntityByID("player:vayne")
	if source == nil {
		t.Fatal("crossbowman was not spawned")
	}
	return w, source
}

func hasEffect(effects []world.SkillEffect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}
