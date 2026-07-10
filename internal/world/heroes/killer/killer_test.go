package killer

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"testing"
)

func TestBouncingBladeHitsThreeEnemiesAndDropsDaggerBehindFinalTarget(t *testing.T) {
	w, killer, _, primary := testWorld(t)
	second := spawnTestHero(t, w, "q-second", world.TeamRed)
	third := spawnTestHero(t, w, "q-third", world.TeamRed)
	fourth := spawnTestHero(t, w, "q-fourth", world.TeamRed)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(primary, 1400, 1000)
	prepareQTarget(second, 1500, 1000)
	prepareQTarget(third, 1600, 1000)
	prepareQTarget(fourth, 1900, 1000)
	killer.Stats.AbilityPower = 100
	state := killer.Skills[qID]
	state.Level = 1
	killer.Skills[qID] = state

	CastQ(w, killer, protocol.CastInput{SkillID: qID, TargetID: fourth.ID, TargetX: primary.Position.X, TargetY: primary.Position.Y}, state, w.SkillConfig(qID), 10, 20)
	if got := killer.Passive.KillerQReleaseTick; got != 15 {
		t.Fatalf("q release tick = %d, want 15", got)
	}
	if got := killer.Skills[qID].CooldownUntilTick; got != 230 {
		t.Fatalf("q cooldown tick = %d, want 230", got)
	}
	if effectByKind(w.SkillEffects(), "killer_q_cast_range") == nil {
		t.Fatal("successful q cast did not create its range effect")
	}
	w.Tick(14, 20)
	if primary.Combat.LastDamage != 0 {
		t.Fatal("q dealt damage before cast windup finished")
	}
	for tick := uint64(15); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}

	for _, target := range []*world.Entity{primary, second, third} {
		if got := target.Combat.LastDamage; got != 105 {
			t.Fatalf("target %s damage = %d, want 105", target.ID, got)
		}
	}
	if fourth.Combat.LastDamage != 0 {
		t.Fatalf("fourth target damage = %d, want 0", fourth.Combat.LastDamage)
	}
	dagger := effectByKind(w.SkillEffects(), "killer_q_dagger")
	if dagger == nil {
		t.Fatal("q did not create a dagger")
	}
	if math.Abs(dagger.Start.X-1950) > 0.001 || math.Abs(dagger.Start.Y-1000) > 0.001 {
		t.Fatalf("dagger position = (%v, %v), want behind final target at (1950, 1000)", dagger.Start.X, dagger.Start.Y)
	}
	if got := dagger.ExpiresAt - dagger.CreatedAt; got != 80 {
		t.Fatalf("dagger duration = %d ticks, want 80", got)
	}
	if len(killer.Passive.KillerDaggers) != 1 {
		t.Fatalf("dagger state count = %d, want 1", len(killer.Passive.KillerDaggers))
	}
	expiresAt := killer.Passive.KillerDaggers[0].ExpiresAt
	w.Tick(expiresAt, 20)
	if len(killer.Passive.KillerDaggers) != 0 || effectByKind(w.SkillEffects(), "killer_q_dagger") != nil {
		t.Fatal("expired q dagger was not removed")
	}
}

func TestBouncingBladeRangeOnlyAppearsAfterSuccessfulCast(t *testing.T) {
	w, killer, _, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	state := killer.Skills[qID]
	state.Level = 1
	killer.Skills[qID] = state
	w.ForEachEntity(func(entity *world.Entity) {
		if entity != nil && entity.ID != killer.ID {
			entity.Team = killer.Team
		}
	})

	CastQ(w, killer, protocol.CastInput{SkillID: qID, TargetX: 1300, TargetY: 1000}, state, w.SkillConfig(qID), 10, 20)
	if killer.Passive.KillerQPending || effectByKind(w.SkillEffects(), "killer_q_cast_range") != nil {
		t.Fatal("failed q cast displayed its range")
	}

	target.Team = world.TeamRed
	prepareQTarget(target, 1300, 1000)
	CastQ(w, killer, protocol.CastInput{SkillID: qID, TargetX: 1300, TargetY: 1000}, state, w.SkillConfig(qID), 11, 20)
	if !killer.Passive.KillerQPending || effectByKind(w.SkillEffects(), "killer_q_cast_range") == nil {
		t.Fatal("successful q cast did not display its range")
	}
}

func TestBouncingBladeChoosesEnemyNearestCursorIgnoringSelectedTarget(t *testing.T) {
	w, killer, _, first := testWorld(t)
	second := spawnTestHero(t, w, "q-cursor-second", world.TeamRed)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(first, 1300, 1000)
	prepareQTarget(second, 1550, 1000)
	state := killer.Skills[qID]
	state.Level = 1
	killer.Skills[qID] = state

	CastQ(w, killer, protocol.CastInput{
		SkillID:  qID,
		TargetID: first.ID,
		TargetX:  1540,
		TargetY:  1000,
	}, state, w.SkillConfig(qID), 10, 20)

	if killer.Passive.KillerQTargetID != second.ID {
		t.Fatalf("q cursor target = %s, want %s", killer.Passive.KillerQTargetID, second.ID)
	}
}

func TestBouncingBladeDaggerTravelsBeforeBecomingPickable(t *testing.T) {
	w, killer, _, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(target, 1400, 1000)
	projectile := &world.Projectile{
		Start: world.Vector2{X: 1000, Y: 1000},
		Dir:   world.Vector2{X: 1, Y: 0},
	}

	spawnQDagger(w, killer, target, projectile, 20, 20)

	if len(killer.Passive.KillerAirborneDaggers) != 1 || len(killer.Passive.KillerDaggers) != 0 {
		t.Fatalf("q airborne/ground dagger count = %d/%d, want 1/0", len(killer.Passive.KillerAirborneDaggers), len(killer.Passive.KillerDaggers))
	}
	airborne := effectByKind(w.SkillEffects(), "killer_q_dagger_airborne")
	if airborne == nil {
		t.Fatal("q dagger travel effect is missing")
	}
	if airborne.Start != target.Position || airborne.End != (world.Vector2{X: 1750, Y: 1000}) {
		t.Fatalf("q dagger trajectory = %+v -> %+v, want target -> (1750, 1000)", airborne.Start, airborne.End)
	}
	if got := airborne.ExpiresAt - airborne.CreatedAt; got != 7 {
		t.Fatalf("q dagger travel ticks = %d, want 7", got)
	}

	w.Tick(26, 20)
	if len(killer.Passive.KillerDaggers) != 0 {
		t.Fatal("q dagger became pickable before reaching its landing point")
	}
	w.Tick(27, 20)
	if len(killer.Passive.KillerAirborneDaggers) != 0 || len(killer.Passive.KillerDaggers) != 1 {
		t.Fatalf("landed q dagger airborne/ground count = %d/%d, want 0/1", len(killer.Passive.KillerAirborneDaggers), len(killer.Passive.KillerDaggers))
	}
	if effectByKind(w.SkillEffects(), "killer_q_dagger_airborne") != nil || effectByKind(w.SkillEffects(), "killer_q_dagger") == nil {
		t.Fatal("q dagger travel effect did not transition to ground effect")
	}
}

func TestBouncingBladeIsBlockedByWindWall(t *testing.T) {
	w, killer, _, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(target, 1400, 1000)
	state := killer.Skills[qID]
	state.Level = 1
	killer.Skills[qID] = state
	w.PutWindWall(world.WindWall{
		ID:        "windwall:killer-q",
		Team:      world.TeamRed,
		Center:    world.Vector2{X: 1200, Y: 1000},
		Dir:       world.Vector2{X: 0, Y: 1},
		Width:     500,
		ExpiresAt: 100,
	})

	CastQ(w, killer, protocol.CastInput{SkillID: qID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, state, w.SkillConfig(qID), 10, 20)
	for tick := uint64(15); tick <= 30; tick++ {
		w.Tick(tick, 20)
	}

	if target.Combat.LastDamage != 0 {
		t.Fatalf("blocked q damage = %d, want 0", target.Combat.LastDamage)
	}
	if effectByKind(w.SkillEffects(), "killer_q_dagger") != nil || len(killer.Passive.KillerDaggers) != 0 {
		t.Fatal("blocked q created a dagger")
	}
}

func TestDaggerPickupSlashesEnemiesAroundKiller(t *testing.T) {
	w, killer, _, first := testWorld(t)
	second := spawnTestHero(t, w, "dagger-second", world.TeamRed)
	outside := spawnTestHero(t, w, "dagger-outside", world.TeamRed)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(first, 1200, 1000)
	prepareQTarget(second, 1000, 1300)
	prepareQTarget(outside, 1400, 1000)
	killer.Level = 1
	killer.Stats.BonusAttack = 100
	killer.Stats.AbilityPower = 100
	addTestDagger(w, killer, world.Vector2{X: 1040, Y: 1000}, 10, 90)

	w.Tick(11, 20)

	for _, target := range []*world.Entity{first, second} {
		if got := target.Combat.LastDamage; got != 203 {
			t.Fatalf("target %s dagger damage = %d, want 203", target.ID, got)
		}
	}
	if outside.Combat.LastDamage != 0 {
		t.Fatalf("outside target dagger damage = %d, want 0", outside.Combat.LastDamage)
	}
	if len(killer.Passive.KillerDaggers) != 0 || effectByKind(w.SkillEffects(), "killer_q_dagger") != nil {
		t.Fatal("picked up dagger was not removed")
	}
	if effectByKind(w.SkillEffects(), "killer_dagger_slash") == nil {
		t.Fatal("dagger pickup did not create slash effect")
	}
}

func TestDaggerSlashDamageScalingAndAPBreakpoints(t *testing.T) {
	w, killer, _, _ := testWorld(t)
	skill := w.SkillConfig(passiveID)
	killer.Stats.BonusAttack = 100
	killer.Stats.AbilityPower = 100

	killer.Level = 1
	if got := daggerSlashRawDamage(killer, skill); math.Abs(got-203) > 0.001 {
		t.Fatalf("level 1 raw damage = %v, want 203", got)
	}
	killer.Level = 18
	if got := daggerSlashRawDamage(killer, skill); math.Abs(got-405) > 0.001 {
		t.Fatalf("level 18 raw damage = %v, want 405", got)
	}

	for _, test := range []struct {
		level int
		want  float64
	}{
		{level: 1, want: 0.7},
		{level: 5, want: 0.7},
		{level: 6, want: 0.8},
		{level: 10, want: 0.8},
		{level: 11, want: 0.9},
		{level: 15, want: 0.9},
		{level: 16, want: 1.0},
		{level: 18, want: 1.0},
	} {
		if got := steppedValue(skill, "daggerAPRatio", "daggerAPRatioLevels", test.level, 0); got != test.want {
			t.Fatalf("level %d AP ratio = %v, want %v", test.level, got, test.want)
		}
	}
}

func TestPreparationIsInstantDecaysMoveSpeedAndLandsDaggerAfterOneSecond(t *testing.T) {
	w, killer, _, _ := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	state := killer.Skills[wID]
	state.Level = 1
	killer.Skills[wID] = state

	CastW(w, killer, protocol.CastInput{SkillID: wID}, state, w.SkillConfig(wID), 10, 20)

	if got := killer.Skills[wID].CooldownUntilTick; got != 310 {
		t.Fatalf("w cooldown tick = %d, want 310", got)
	}
	if killer.Control.ActionLockedUntilTick != 0 {
		t.Fatalf("instant w action lock = %d, want 0", killer.Control.ActionLockedUntilTick)
	}
	if len(killer.Passive.KillerAirborneDaggers) != 1 || len(killer.Passive.KillerDaggers) != 0 {
		t.Fatalf("dagger states airborne/ground = %d/%d, want 1/0", len(killer.Passive.KillerAirborneDaggers), len(killer.Passive.KillerDaggers))
	}
	if effectByKind(w.SkillEffects(), "killer_w_dagger_airborne") == nil {
		t.Fatal("w did not create airborne dagger effect")
	}
	if got := world.EffectiveMoveSpeedAtTick(killer, 10); math.Abs(got-517.5) > 0.001 {
		t.Fatalf("w initial move speed = %v, want 517.5", got)
	}
	if got := world.EffectiveMoveSpeedAtTick(killer, 22); math.Abs(got-434.7) > 0.001 {
		t.Fatalf("w decayed move speed = %v, want 434.7", got)
	}
	if got := world.EffectiveMoveSpeedAtTick(killer, 35); math.Abs(got-345) > 0.001 {
		t.Fatalf("w expired move speed = %v, want 345", got)
	}

	w.Tick(29, 20)
	if len(killer.Passive.KillerDaggers) != 0 || effectByKind(w.SkillEffects(), "killer_dagger_slash") != nil {
		t.Fatal("airborne w dagger became pickable before one second")
	}
	killer.Position = world.Vector2{X: 1200, Y: 1000}
	w.Tick(30, 20)
	if len(killer.Passive.KillerAirborneDaggers) != 0 || len(killer.Passive.KillerDaggers) != 1 {
		t.Fatalf("landed dagger states airborne/ground = %d/%d, want 0/1", len(killer.Passive.KillerAirborneDaggers), len(killer.Passive.KillerDaggers))
	}
	if got := killer.Passive.KillerDaggers[0].Position; got != (world.Vector2{X: 1000, Y: 1000}) {
		t.Fatalf("w dagger position = %+v, want cast position", got)
	}
	if got := killer.Passive.KillerDaggers[0].ExpiresAt; got != 110 {
		t.Fatalf("w ground dagger expiry = %d, want 110", got)
	}
	if effectByKind(w.SkillEffects(), "killer_w_dagger") == nil {
		t.Fatal("landed w dagger effect is missing")
	}

	killer.Position = world.Vector2{X: 1000, Y: 1000}
	w.Tick(31, 20)
	if len(killer.Passive.KillerDaggers) != 0 || effectByKind(w.SkillEffects(), "killer_dagger_slash") == nil {
		t.Fatal("landed w dagger was not pickable")
	}
}

func TestPreparationRankValues(t *testing.T) {
	w, _, _, _ := testWorld(t)
	skill := w.SkillConfig(wID)
	for index, want := range []float64{15000, 14000, 13000, 12000, 11000} {
		if got := skillList(skill, "cooldownMs", index+1, nil); got != want {
			t.Fatalf("rank %d cooldown = %v, want %v", index+1, got, want)
		}
	}
	for index, want := range []float64{0.5, 0.6, 0.7, 0.8, 0.9} {
		if got := skillList(skill, "moveSpeedBonus", index+1, nil); got != want {
			t.Fatalf("rank %d move speed bonus = %v, want %v", index+1, got, want)
		}
	}
}

func TestShunpoBlinksToEnemyDealsDamageAndGrantsReduction(t *testing.T) {
	w, killer, _, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(target, 1600, 1000)
	killer.Stats.AbilityPower = 100
	killer.Stats.BonusAttack = 100
	state := killer.Skills[eID]
	state.Level = 1
	killer.Skills[eID] = state

	CastE(w, killer, protocol.CastInput{SkillID: eID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, state, w.SkillConfig(eID), 10, 20)

	if killer.Position != target.Position {
		t.Fatalf("e position = %+v, want %+v", killer.Position, target.Position)
	}
	if got := target.Combat.LastDamage; got != 135 {
		t.Fatalf("e damage = %d, want 135", got)
	}
	if got := killer.Skills[eID].CooldownUntilTick; got != 290 {
		t.Fatalf("e cooldown tick = %d, want 290", got)
	}
	if killer.Control.ActionLockedUntilTick != 0 {
		t.Fatalf("instant e action lock = %d, want 0", killer.Control.ActionLockedUntilTick)
	}
	if effectByKind(w.SkillEffects(), "killer_e") == nil {
		t.Fatal("e blink effect is missing")
	}
	killer.Stats.PhysicalDefense = 0
	killer.Stats.MagicDefense = 0
	if got := w.MagicDamageAfterResistance(target, killer, 100, 11); got != 80 {
		t.Fatalf("damage during e reduction = %d, want 80", got)
	}
	if got := w.MagicDamageAfterResistance(target, killer, 100, 70); got != 100 {
		t.Fatalf("damage after e reduction = %d, want 100", got)
	}
}

func TestShunpoToDaggerPicksItUpAndUsesReducedCooldown(t *testing.T) {
	w, killer, _, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(target, 1650, 1000)
	addTestDagger(w, killer, world.Vector2{X: 1600, Y: 1000}, 1, 100)
	state := killer.Skills[eID]
	state.Level = 5
	killer.Skills[eID] = state

	CastE(w, killer, protocol.CastInput{SkillID: eID, TargetID: target.ID, TargetX: 1600, TargetY: 1000}, state, w.SkillConfig(eID), 10, 20)

	if killer.Position != (world.Vector2{X: 1600, Y: 1000}) {
		t.Fatalf("dagger e position = %+v, want (1600, 1000)", killer.Position)
	}
	if got := killer.Skills[eID].CooldownUntilTick; got != 46 {
		t.Fatalf("dagger e cooldown tick = %d, want 46", got)
	}
	if len(killer.Passive.KillerDaggers) != 0 || effectByKind(w.SkillEffects(), "killer_q_dagger") != nil {
		t.Fatal("e did not consume targeted dagger")
	}
	if effectByKind(w.SkillEffects(), "killer_dagger_slash") == nil {
		t.Fatal("e dagger pickup did not trigger passive slash")
	}
	if got := target.Combat.LastDamage; got != 68 {
		t.Fatalf("e dagger passive damage = %d, want 68", got)
	}
}

func TestShunpoCanBlinkToAllyWithoutDamagingIt(t *testing.T) {
	w, killer, ally, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(ally, 1500, 1000)
	prepareQTarget(target, 1650, 1000)
	state := killer.Skills[eID]
	state.Level = 1
	killer.Skills[eID] = state

	CastE(w, killer, protocol.CastInput{SkillID: eID, TargetID: target.ID, TargetX: ally.Position.X, TargetY: ally.Position.Y}, state, w.SkillConfig(eID), 10, 20)

	if killer.Position != ally.Position {
		t.Fatalf("ally e position = %+v, want %+v", killer.Position, ally.Position)
	}
	if ally.Combat.LastDamage != 0 {
		t.Fatalf("ally e damage = %d, want 0", ally.Combat.LastDamage)
	}
}

func TestShunpoTargetRulesAndRange(t *testing.T) {
	source := &world.Entity{ID: "source", Kind: world.EntityKindPlayer, Team: world.TeamBlue, Position: world.Vector2{X: 1000, Y: 1000}, Stats: world.Stats{HP: 1000}}
	tests := []struct {
		name string
		unit *world.Entity
		want bool
	}{
		{name: "ally hero", unit: eTestTarget("ally", world.EntityKindPlayer, world.TeamBlue, 1500), want: true},
		{name: "enemy minion", unit: eTestTarget("minion", world.EntityKindMeleeMinion, world.TeamRed, 1500), want: true},
		{name: "neutral monster", unit: eTestTarget("monster", world.EntityKindBlueBuff, world.TeamNeutral, 1500), want: true},
		{name: "fruit", unit: eTestTarget("fruit", world.EntityKindFruit, world.TeamNeutral, 1500), want: true},
		{name: "tower", unit: eTestTarget("tower", world.EntityKindTower, world.TeamRed, 1500), want: false},
		{name: "ward", unit: eTestTarget("ward", world.EntityKindWard, world.TeamRed, 1500), want: false},
		{name: "barracks", unit: eTestTarget("barracks", world.EntityKindBarracks, world.TeamRed, 1500), want: false},
		{name: "out of range", unit: eTestTarget("far", world.EntityKindEnemyHero, world.TeamRed, 1800), want: false},
		{name: "self", unit: source, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validEEntityTarget(source, test.unit, 700, 10); got != test.want {
				t.Fatalf("valid target = %v, want %v", got, test.want)
			}
		})
	}
}

func TestShunpoAllowsSpawnedFruitAndRejectsWardAsCandidate(t *testing.T) {
	w, killer, _, _ := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	state := killer.Skills[eID]
	state.Level = 1
	killer.Skills[eID] = state
	fruitID, ok := w.SpawnObject(world.EntityKindFruit, world.TeamNeutral, 1500, 1000)
	if !ok {
		t.Fatal("failed to spawn fruit")
	}
	fruit := w.EntityByID(fruitID)

	CastE(w, killer, protocol.CastInput{SkillID: eID, TargetX: fruit.Position.X, TargetY: fruit.Position.Y}, state, w.SkillConfig(eID), 10, 20)
	if killer.Position != fruit.Position || fruit.Stats.HP != 1 {
		t.Fatalf("fruit e position/hp = %+v/%v, want %+v/1", killer.Position, fruit.Stats.HP, fruit.Position)
	}

	wardID, ok := w.SpawnObject(world.EntityKindWard, world.TeamRed, 1500, 1000)
	if !ok {
		t.Fatal("failed to spawn ward")
	}
	ward := w.EntityByID(wardID)
	if validEEntityTarget(killer, ward, 700, 20) {
		t.Fatal("ward became a valid e candidate")
	}
}

func TestShunpoRankValues(t *testing.T) {
	w, _, _, _ := testWorld(t)
	skill := w.SkillConfig(eID)
	for index, want := range []float64{14000, 12500, 11000, 9500, 8000} {
		if got := skillList(skill, "cooldownMs", index+1, nil); got != want {
			t.Fatalf("rank %d cooldown = %v, want %v", index+1, got, want)
		}
	}
	for index, want := range []float64{30, 45, 60, 75, 90} {
		if got := skillList(skill, "baseDamage", index+1, nil); got != want {
			t.Fatalf("rank %d damage = %v, want %v", index+1, got, want)
		}
	}
}

func TestDeathLotusFiresTenSegmentsAndAppliesGrievousWounds(t *testing.T) {
	w, killer, _, target := testWorld(t)
	prepareQTarget(killer, 3000, 3000)
	prepareQTarget(target, 3200, 3000)
	killer.Stats.AbilityPower = 100
	killer.Stats.BonusAttack = 100
	state := killer.Skills[rID]
	state.Level = 1
	killer.Skills[rID] = state

	CastR(w, killer, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)
	if got := killer.Skills[rID].CooldownUntilTick; got != 1810 {
		t.Fatalf("r cooldown tick = %d, want 1810", got)
	}
	if got := killer.Passive.KillerRExpireTick; got != 60 {
		t.Fatalf("r channel expiry = %d, want 60", got)
	}
	if got := countEffectsByKind(w.SkillEffects(), "killer_r"); got != 1 {
		t.Fatalf("initial r projectiles = %d, want 1", got)
	}
	hits := 0
	for tick := uint64(11); tick <= 61; tick++ {
		beforeHit := target.Combat.LastHitTick
		w.Tick(tick, 20)
		if target.Combat.LastHitTick != beforeHit {
			hits++
		}
	}
	if hits != 10 {
		t.Fatalf("r hit segments = %d, want 10", hits)
	}
	if target.Control.GrievousWounds != 0.4 {
		t.Fatalf("r grievous wounds = %v, want 0.4", target.Control.GrievousWounds)
	}
	if got := target.Control.GrievousWoundsUntil; got != target.Combat.LastHitTick+60 {
		t.Fatalf("r grievous expiry = %d, want last hit + 60 (%d)", got, target.Combat.LastHitTick+60)
	}
	if killer.Passive.KillerRExpireTick != 0 || effectByKind(w.SkillEffects(), "killer_r_channel") != nil {
		t.Fatal("r channel state or effect remained after 2.5 seconds")
	}
}

func TestDeathLotusTargetsNearestThreeEnemyHeroesOnly(t *testing.T) {
	w, killer, _, first := testWorld(t)
	second := spawnTestHero(t, w, "r-second", world.TeamRed)
	third := spawnTestHero(t, w, "r-third", world.TeamRed)
	fourth := spawnTestHero(t, w, "r-fourth", world.TeamRed)
	minionID, ok := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamRed, 1100, 1100)
	if !ok {
		t.Fatal("failed to spawn r minion")
	}
	minion := w.EntityByID(minionID)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(first, 1100, 1000)
	prepareQTarget(second, 1200, 1000)
	prepareQTarget(third, 1300, 1000)
	prepareQTarget(fourth, 1400, 1000)
	state := killer.Skills[rID]
	state.Level = 1
	killer.Skills[rID] = state

	CastR(w, killer, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)
	if got := countEffectsByKind(w.SkillEffects(), "killer_r"); got != 3 {
		t.Fatalf("r first segment projectile count = %d, want 3", got)
	}
	for tick := uint64(11); tick <= 14; tick++ {
		w.Tick(tick, 20)
	}
	for _, target := range []*world.Entity{first, second, third} {
		if target.Combat.LastDamage != 40 {
			t.Fatalf("nearest target %s damage = %d, want 40", target.ID, target.Combat.LastDamage)
		}
	}
	if fourth.Combat.LastDamage != 0 || minion.Combat.LastDamage != 0 {
		t.Fatalf("excluded fourth/minion damage = %d/%d, want 0/0", fourth.Combat.LastDamage, minion.Combat.LastDamage)
	}
}

func TestDeathLotusAllowsMovementAndShunpoButBlocksOtherActions(t *testing.T) {
	w, killer, ally, target := testWorld(t)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(ally, 1400, 1000)
	prepareQTarget(target, 2000, 1000)
	state := killer.Skills[rID]
	state.Level = 1
	killer.Skills[rID] = state
	CastR(w, killer, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)

	if got := world.EffectiveMoveSpeedAtTick(killer, 11); math.Abs(got-69) > 0.001 {
		t.Fatalf("r move speed = %v, want 69", got)
	}
	qState := killer.Skills[qID]
	qState.Level = 1
	CastQ(w, killer, protocol.CastInput{SkillID: qID, TargetID: target.ID}, qState, w.SkillConfig(qID), 11, 20)
	wState := killer.Skills[wID]
	wState.Level = 1
	CastW(w, killer, protocol.CastInput{SkillID: wID}, wState, w.SkillConfig(wID), 11, 20)
	if killer.Passive.KillerQPending || killer.Skills[qID].CooldownUntilTick != 0 || killer.Skills[wID].CooldownUntilTick != 0 {
		t.Fatal("r channel allowed q or w")
	}

	target.Position = world.Vector2{X: 1100, Y: 1000}
	killer.Intent.AttackTargetID = target.ID
	w.Tick(11, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatalf("basic attack during r dealt %d damage", target.Combat.LastDamage)
	}
	eState := killer.Skills[eID]
	eState.Level = 1
	CastE(w, killer, protocol.CastInput{SkillID: eID, TargetID: target.ID, TargetX: ally.Position.X, TargetY: ally.Position.Y}, eState, w.SkillConfig(eID), 12, 20)
	if killer.Position != ally.Position || !channelingR(killer, 12) {
		t.Fatalf("r e combo position/channel = %+v/%v", killer.Position, channelingR(killer, 12))
	}
}

func TestDeathLotusIsInterruptedByHardControl(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*world.Entity)
	}{
		{name: "stun", apply: func(entity *world.Entity) { entity.Control.StunnedUntilTick = 20 }},
		{name: "airborne", apply: func(entity *world.Entity) { entity.Control.AirborneUntilTick = 20 }},
		{name: "root", apply: func(entity *world.Entity) { entity.Control.RootedUntilTick = 20 }},
		{name: "taunt", apply: func(entity *world.Entity) { entity.Control.TauntedUntilTick = 20 }},
		{name: "suppression", apply: func(entity *world.Entity) { entity.Control.SuppressedUntilTick = 20 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w, killer, _, target := testWorld(t)
			prepareQTarget(killer, 1000, 1000)
			prepareQTarget(target, 2000, 1000)
			state := killer.Skills[rID]
			state.Level = 1
			killer.Skills[rID] = state
			CastR(w, killer, protocol.CastInput{SkillID: rID}, state, w.SkillConfig(rID), 10, 20)
			test.apply(killer)
			w.Tick(11, 20)
			if killer.Passive.KillerRExpireTick != 0 || effectByKind(w.SkillEffects(), "killer_r_channel") != nil {
				t.Fatal("hard control did not interrupt r")
			}
		})
	}
}

func TestDeathLotusRankValues(t *testing.T) {
	w, killer, _, target := testWorld(t)
	skill := w.SkillConfig(rID)
	prepareQTarget(killer, 1000, 1000)
	prepareQTarget(target, 1200, 1000)
	killer.Stats.AbilityPower = 100
	killer.Stats.BonusAttack = 100
	for index, want := range []float64{90000, 60000, 45000} {
		if got := skillList(skill, "cooldownMs", index+1, nil); got != want {
			t.Fatalf("rank %d cooldown = %v, want %v", index+1, got, want)
		}
	}
	for index, want := range []int{79, 127, 175} {
		if got := RDamage(w, killer, target, skill, index+1, 10); got != want {
			t.Fatalf("rank %d segment damage = %d, want %d", index+1, got, want)
		}
	}
}

func TestVoracityRefreshesBasicSkillsAndRefundsUltimateOnAssist(t *testing.T) {
	w, killer, ally, target := testWorld(t)
	setCooldowns(killer, 500)

	target.Combat.LastHitTick = 100
	w.ApplyDamage(killer, target, 1, 20)
	target.Combat.LastHitTick = 160
	w.ApplyKillReward(ally, target)

	for _, skillID := range []string{qID, wID, eID} {
		if got := killer.Skills[skillID].CooldownUntilTick; got != 160 {
			t.Fatalf("%s cooldown = %d, want 160", skillID, got)
		}
	}
	if got := killer.Skills[rID].CooldownUntilTick; got != 200 {
		t.Fatalf("r cooldown = %d, want 200", got)
	}
}

func TestVoracityDoesNotTriggerAfterDamageWindow(t *testing.T) {
	w, killer, ally, target := testWorld(t)
	setCooldowns(killer, 500)

	target.Combat.LastHitTick = 100
	w.ApplyDamage(killer, target, 1, 20)
	target.Combat.LastHitTick = 161
	w.ApplyKillReward(ally, target)

	for _, skillID := range []string{qID, wID, eID, rID} {
		if got := killer.Skills[skillID].CooldownUntilTick; got != 500 {
			t.Fatalf("%s cooldown = %d, want 500", skillID, got)
		}
	}
}

func TestVoracityUltimateRefundStopsAtReady(t *testing.T) {
	w, killer, ally, target := testWorld(t)
	setCooldowns(killer, 200)

	target.Combat.LastHitTick = 100
	w.ApplyDamage(killer, target, 1, 20)
	target.Combat.LastHitTick = 150
	w.ApplyKillReward(ally, target)

	if got := killer.Skills[rID].CooldownUntilTick; got != 150 {
		t.Fatalf("r cooldown = %d, want 150", got)
	}
}

func setCooldowns(entity *world.Entity, until uint64) {
	for _, skillID := range []string{qID, wID, eID, rID} {
		state := entity.Skills[skillID]
		state.CooldownUntilTick = until
		entity.Skills[skillID] = state
	}
}

func prepareQTarget(entity *world.Entity, x float64, y float64) {
	entity.Position = world.Vector2{X: x, Y: y}
	entity.Stats.HP = 1000
	entity.Stats.MaxHP = 1000
	entity.Stats.PhysicalDefense = 0
	entity.Stats.MagicDefense = 0
}

func addTestDagger(w *world.World, killer *world.Entity, position world.Vector2, createdAt uint64, expiresAt uint64) {
	effectID := w.NextEffectID("effect:test_killer_dagger:")
	killer.Passive.KillerDaggers = append(killer.Passive.KillerDaggers, world.KillerDaggerState{
		EffectID:  effectID,
		Position:  position,
		ExpiresAt: expiresAt,
	})
	w.PutSkillEffect(world.SkillEffect{
		ID:           effectID,
		Kind:         "killer_q_dagger",
		Team:         killer.Team,
		SourceID:     killer.ID,
		SourceHeroID: killer.HeroID,
		Start:        position,
		Radius:       32,
		CreatedAt:    createdAt,
		ExpiresAt:    expiresAt,
	})
}

func eTestTarget(id string, kind world.EntityKind, team world.Team, x float64) *world.Entity {
	return &world.Entity{
		ID:       id,
		Kind:     kind,
		Team:     team,
		Position: world.Vector2{X: x, Y: 1000},
		Radius:   18,
		Stats:    world.Stats{HP: 1000, MaxHP: 1000},
	}
}

func spawnTestHero(t *testing.T, w *world.World, playerID string, team world.Team) *world.Entity {
	t.Helper()
	heroes, err := config.LoadHeroes("../../../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("warrior")
	if !ok {
		t.Fatal("warrior hero not found")
	}
	w.SpawnHero(playerID, hero, team)
	entity := w.EntityByID("player:" + playerID)
	if entity == nil {
		t.Fatalf("spawned hero %s not found", playerID)
	}
	return entity
}

func effectByKind(effects []world.SkillEffect, kind string) *world.SkillEffect {
	for index := range effects {
		if effects[index].Kind == kind {
			return &effects[index]
		}
	}
	return nil
}

func countEffectsByKind(effects []world.SkillEffect, kind string) int {
	count := 0
	for _, effect := range effects {
		if effect.Kind == kind {
			count++
		}
	}
	return count
}

func testWorld(t *testing.T) (*world.World, *world.Entity, *world.Entity, *world.Entity) {
	t.Helper()
	heroes, err := config.LoadHeroes("../../../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.LoadSkills("../../../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	equipment, err := config.LoadEquipment("../../../../configs/equipment")
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(heroes, skills, nil, nil, equipment)
	killerConfig, ok := heroes.Get(heroID)
	if !ok {
		t.Fatal("killer hero not found")
	}
	allyConfig, ok := heroes.Get("warrior")
	if !ok {
		t.Fatal("ally hero not found")
	}
	w.SpawnHero("killer", killerConfig, world.TeamBlue)
	w.SpawnHero("ally", allyConfig, world.TeamBlue)
	w.SpawnHero("target", allyConfig, world.TeamRed)
	return w, w.EntityByID("player:killer"), w.EntityByID("player:ally"), w.EntityByID("player:target")
}
