package ninja

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"testing"
)

func TestPassiveDamageHeroCooldownAndScaling(t *testing.T) {
	skill := passiveSkill()
	attacker := &world.Entity{ID: "ninja", HeroID: heroID, Level: 1, Team: world.TeamBlue}
	target := &world.Entity{ID: "enemy", Kind: world.EntityKindEnemyHero, Team: world.TeamRed, Stats: world.Stats{HP: 499, MaxHP: 1000}}

	if got := passiveRawDamage(attacker, target, skill, 10, 20); got != 50 {
		t.Fatalf("level 1 damage = %f, want 50", got)
	}
	if got := passiveRawDamage(attacker, target, skill, 20, 20); got != 0 {
		t.Fatalf("cooldown damage = %f, want 0", got)
	}
	attacker.Level = 18
	if got := passiveRawDamage(attacker, target, skill, 210, 20); got != 100 {
		t.Fatalf("level 18 damage after cooldown = %f, want 100", got)
	}
}

func TestPassiveDamageMonsterRules(t *testing.T) {
	skill := passiveSkill()
	attacker := &world.Entity{ID: "ninja", HeroID: heroID, Level: 18, Team: world.TeamBlue}
	monster := &world.Entity{ID: "red", Kind: world.EntityKindRedBuff, Team: world.TeamNeutral, Stats: world.Stats{HP: 4999, MaxHP: 10000}}
	epic := &world.Entity{ID: "baron", Kind: world.EntityKindBaronNashor, Team: world.TeamNeutral, Stats: world.Stats{HP: 4999, MaxHP: 10000}}

	if got := passiveRawDamage(attacker, monster, skill, 1, 20); got != 750 {
		t.Fatalf("monster damage = %f, want 750", got)
	}
	if got := passiveRawDamage(attacker, epic, skill, 1, 20); got != 175 {
		t.Fatalf("epic monster damage = %f, want 175", got)
	}
}

func TestQDamageUsesFirstAndLaterHitScaling(t *testing.T) {
	skill := qSkill()
	attacker := &world.Entity{Stats: world.Stats{BonusAttack: 11}}

	if got := qRawDamage(attacker, skill, 2, 1); got != 131 {
		t.Fatalf("first hit damage = %f, want 131", got)
	}
	if got := qRawDamage(attacker, skill, 2, 2); math.Abs(got-78.6) > 0.000001 {
		t.Fatalf("later hit damage = %f, want 78.6", got)
	}
}

func TestQReleasesProjectileAfterWindup(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	source := &world.Entity{
		ID:       "ninja",
		HeroID:   heroID,
		Team:     world.TeamBlue,
		Position: world.Vector2{X: 100, Y: 100},
		Stats:    world.Stats{MP: 200, MaxMP: 200, AttackSpeed: 1},
		Skills:   map[string]world.SkillState{qID: {SkillID: qID, Level: 1}},
		Ninja: world.NinjaState{
			ShadowPosition:  world.Vector2{X: 100, Y: 200},
			ShadowExpiresAt: 100,
		},
	}

	CastQ(w, source, protocol.CastInput{SkillID: qID, TargetX: 1000, TargetY: 100}, source.Skills[qID], qSkill(), 10, 20)

	if source.Ninja.QReleaseTick != 15 || source.Control.ActionLockedUntilTick != 15 {
		t.Fatalf("q release=%d locked=%d, want 15/15", source.Ninja.QReleaseTick, source.Control.ActionLockedUntilTick)
	}
	if countNinjaShurikens(w) != 0 {
		t.Fatal("q should not fire before windup")
	}

	Tick(w, source, 15, 20)
	if countNinjaShurikens(w) != 2 {
		t.Fatalf("q shuriken count = %d, want 2", countNinjaShurikens(w))
	}
	shadowDirChecked := false
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "ninja_shuriken" && effect.Radius != 87.5 {
			t.Fatalf("q missile radius = %f, want 87.5", effect.Radius)
		}
		if effect.Kind == "ninja_shuriken" && effect.Start.Y == 200 {
			shadowDirChecked = true
			want := -100 / math.Hypot(900, -100)
			if math.Abs(effect.Dir.Y-want) > 0.000001 {
				t.Fatalf("shadow q dirY = %f, want %f", effect.Dir.Y, want)
			}
		}
	}
	if !shadowDirChecked {
		t.Fatal("missing shadow q projectile")
	}
	if source.Ninja.QPending {
		t.Fatal("q pending should be cleared after release")
	}
	if got, want := source.Skills[qID].CooldownUntilTick, uint64(135); got != want {
		t.Fatalf("q cooldown = %d, want %d", got, want)
	}
}

func TestWCreatesShadowAndRecastSwaps(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	source := &world.Entity{
		ID:       "ninja",
		HeroID:   heroID,
		Team:     world.TeamBlue,
		Position: world.Vector2{X: 100, Y: 100},
		Radius:   16,
		Stats:    world.Stats{MP: 200, MaxMP: 200, AttackSpeed: 1},
		Skills:   map[string]world.SkillState{wID: {SkillID: wID, Level: 1}},
	}

	CastW(w, source, protocolCast(750, 100), source.Skills[wID], wSkill(), 10, 20)

	if source.Stats.MP != 180 {
		t.Fatalf("mp after w = %f, want 180", source.Stats.MP)
	}
	if source.Ninja.ShadowPosition.X != 750 || source.Ninja.ShadowPosition.Y != 100 {
		t.Fatalf("shadow position = %+v, want 750,100", source.Ninja.ShadowPosition)
	}
	if source.Ninja.ShadowExpiresAt != 110 {
		t.Fatalf("shadow expires = %d, want 110", source.Ninja.ShadowExpiresAt)
	}
	effect := onlyNinjaShadowEffect(t, w)
	if effect.Start.X != 100 || effect.End.X != 750 || effect.Speed != 130 {
		t.Fatalf("shadow effect start=%+v end=%+v speed=%f, want x 100->750 speed 130", effect.Start, effect.End, effect.Speed)
	}

	if SpecialRecast(w, source, protocolCast(0, 0), source.Skills[wID], wSkill(), 12, 20) {
		t.Fatal("w recast should wait until shadow arrives")
	}
	if source.Position.X != 100 || source.Ninja.ShadowPosition.X != 750 {
		t.Fatalf("early recast moved source=%+v shadow=%+v", source.Position, source.Ninja.ShadowPosition)
	}
	if !SpecialRecast(w, source, protocolCast(0, 0), source.Skills[wID], wSkill(), 20, 20) {
		t.Fatal("w recast should swap")
	}
	if source.Position.X != 750 || source.Ninja.ShadowPosition.X != 100 {
		t.Fatalf("swap positions source=%+v shadow=%+v", source.Position, source.Ninja.ShadowPosition)
	}
	if SpecialRecast(w, source, protocolCast(0, 0), source.Skills[wID], wSkill(), 21, 20) {
		t.Fatal("w recast should only swap once")
	}

	Tick(w, source, 110, 20)
	if source.Ninja.ShadowExpiresAt != 0 {
		t.Fatalf("shadow did not expire: %+v", source.Ninja)
	}
}

func TestWMovingShadowDelaysCopiedE(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	source := &world.Entity{
		ID:       "ninja",
		HeroID:   heroID,
		Team:     world.TeamBlue,
		Position: world.Vector2{X: 1000, Y: 1000},
		Radius:   16,
		Stats:    world.Stats{MP: 200, MaxMP: 200, BonusAttack: 10, AttackSpeed: 1},
		Skills: map[string]world.SkillState{
			wID: {SkillID: wID, Level: 1},
			eID: {SkillID: eID, Level: 1},
		},
	}
	bodyID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1100, 1000)
	shadowID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1550, 1000)
	bodyTarget := w.EntityByID(bodyID)
	shadowTarget := w.EntityByID(shadowID)
	bodyTarget.Stats.HP, bodyTarget.Stats.MaxHP, bodyTarget.Stats.PhysicalDefense = 1000, 1000, 0
	shadowTarget.Stats.HP, shadowTarget.Stats.MaxHP, shadowTarget.Stats.PhysicalDefense = 1000, 1000, 0

	CastW(w, source, protocolCast(1650, 1000), source.Skills[wID], wSkill(), 10, 20)
	CastE(w, source, protocol.CastInput{}, source.Skills[eID], eSkill(), 12, 20)

	if bodyTarget.Stats.HP != 923 {
		t.Fatalf("body e hp = %d, want 923", bodyTarget.Stats.HP)
	}
	if shadowTarget.Stats.HP != 1000 {
		t.Fatalf("moving shadow should not e yet, hp = %d", shadowTarget.Stats.HP)
	}
	Tick(w, source, 15, 20)
	if shadowTarget.Stats.HP != 923 {
		t.Fatalf("delayed shadow e hp = %d, want 923", shadowTarget.Stats.HP)
	}
}

func TestWPassiveRefundsOncePerSkillCast(t *testing.T) {
	source := &world.Entity{
		ID:     "ninja",
		HeroID: heroID,
		Team:   world.TeamBlue,
		Stats:  world.Stats{MP: 100, MaxMP: 200},
		Skills: map[string]world.SkillState{wID: {SkillID: wID, Level: 3}},
	}
	target := &world.Entity{ID: "target", Kind: world.EntityKindEnemyHero, Team: world.TeamRed, Stats: world.Stats{HP: 1000, MaxHP: 1000}}

	SkillHit(nil, source, target, qID, "cast:1", false, 10, 20)
	if source.Stats.MP != 100 {
		t.Fatalf("mp after first side hit = %f, want 100", source.Stats.MP)
	}
	SkillHit(nil, source, target, qID, "cast:1", true, 11, 20)
	if source.Stats.MP != 140 {
		t.Fatalf("mp after both sides hit = %f, want 140", source.Stats.MP)
	}
	SkillHit(nil, source, &world.Entity{ID: "target2", Kind: world.EntityKindEnemyHero, Team: world.TeamRed}, qID, "cast:1", false, 12, 20)
	SkillHit(nil, source, &world.Entity{ID: "target2", Kind: world.EntityKindEnemyHero, Team: world.TeamRed}, qID, "cast:1", true, 13, 20)
	if source.Stats.MP != 140 {
		t.Fatalf("mp after same cast second refund = %f, want 140", source.Stats.MP)
	}
}

func TestEDamagesOnceSlowsRefundsAndReducesWCooldown(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	source := &world.Entity{
		ID:       "ninja",
		HeroID:   heroID,
		Team:     world.TeamBlue,
		Position: world.Vector2{X: 1000, Y: 1000},
		Stats:    world.Stats{MP: 100, MaxMP: 200, BonusAttack: 10, AttackSpeed: 1},
		Skills: map[string]world.SkillState{
			eID: {SkillID: eID, Level: 2},
			wID: {SkillID: wID, Level: 2, CooldownUntilTick: 500},
		},
		Ninja: world.NinjaState{
			ShadowPosition:  world.Vector2{X: 1300, Y: 1000},
			ShadowExpiresAt: 100,
		},
	}
	bothID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1150, 1000)
	shadowID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1350, 1000)
	both := w.EntityByID(bothID)
	shadowOnly := w.EntityByID(shadowID)
	both.Stats.HP, both.Stats.MaxHP, both.Stats.PhysicalDefense = 1000, 1000, 0
	shadowOnly.Stats.HP, shadowOnly.Stats.MaxHP, shadowOnly.Stats.PhysicalDefense = 1000, 1000, 0

	CastE(w, source, protocol.CastInput{}, source.Skills[eID], eSkill(), 10, 20)

	if source.Stats.MP != 95 {
		t.Fatalf("mp after e cost and refund = %f, want 95", source.Stats.MP)
	}
	if both.Stats.HP != 900 || shadowOnly.Stats.HP != 900 {
		t.Fatalf("hp after e both=%d shadow=%d, want 900/900", both.Stats.HP, shadowOnly.Stats.HP)
	}
	if both.Control.MoveSpeedSlow != 0.375 || shadowOnly.Control.MoveSpeedSlow != 0.25 {
		t.Fatalf("slows both=%f shadow=%f", both.Control.MoveSpeedSlow, shadowOnly.Control.MoveSpeedSlow)
	}
	if both.Control.MoveSpeedSlowUntil != 40 || shadowOnly.Control.MoveSpeedSlowUntil != 40 {
		t.Fatalf("slow until both=%d shadow=%d, want 40", both.Control.MoveSpeedSlowUntil, shadowOnly.Control.MoveSpeedSlowUntil)
	}
	if got, want := source.Skills[wID].CooldownUntilTick, uint64(380); got != want {
		t.Fatalf("w cooldown after hero hits = %d, want %d", got, want)
	}
	if got, want := source.Skills[eID].CooldownUntilTick, uint64(100); got != want {
		t.Fatalf("e cooldown = %d, want %d", got, want)
	}
	effects := ninjaEEffects(w)
	if len(effects) != 2 {
		t.Fatalf("ninja e range effects = %d, want 2", len(effects))
	}
	if effects[0].Radius != 290 || effects[0].ExpiresAt != 17 {
		t.Fatalf("ninja e effect radius/expires = %f/%d, want 290/17", effects[0].Radius, effects[0].ExpiresAt)
	}
}

func TestRWindupDashMarkDamageAndRecast(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	sourceID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamBlue, 100, 100)
	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 500, 100)
	source := w.EntityByID(sourceID)
	target := w.EntityByID(targetID)
	source.HeroID = heroID
	source.Radius = 16
	source.Stats = world.Stats{HP: 1000, MaxHP: 1000, Attack: 100, AttackSpeed: 1}
	source.Skills = map[string]world.SkillState{rID: {SkillID: rID, Level: 2}}
	source.Intent.MoveTarget = &world.Vector2{X: 900, Y: 900}
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0

	CastR(w, source, protocol.CastInput{SkillID: rID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, source.Skills[rID], rSkill(), 10, 20)

	if source.Ninja.RReleaseTick != 22 || source.Ninja.RDashEndTick != 29 {
		t.Fatalf("r timing release=%d dashEnd=%d, want 22/29", source.Ninja.RReleaseTick, source.Ninja.RDashEndTick)
	}
	if source.Intent.MoveTarget != nil {
		t.Fatalf("r should clear move target, got %+v", source.Intent.MoveTarget)
	}
	if source.Control.UntargetableUntilTick != 29 || source.Control.ActionLockedUntilTick != 29 {
		t.Fatalf("r control untargetable=%d locked=%d, want 29/29", source.Control.UntargetableUntilTick, source.Control.ActionLockedUntilTick)
	}
	if got, want := source.Skills[rID].CooldownUntilTick, uint64(2210); got != want {
		t.Fatalf("r cooldown = %d, want %d", got, want)
	}
	if hits := w.TargetsInRadius(target, source.Position, 100); len(hits) != 0 {
		t.Fatalf("untargetable source should not be targetable, got %d hits", len(hits))
	}

	source.Ninja.ShadowPosition = world.Vector2{X: 25, Y: 25}
	source.Ninja.ShadowExpiresAt = 200
	source.Ninja.ShadowRecastSkillID = wID
	source.Ninja.ShadowRecastUntil = 200
	Tick(w, source, 22, 20)
	if source.Control.DashUntilTick != 29 {
		t.Fatalf("dash until = %d, want 29", source.Control.DashUntilTick)
	}
	if source.Ninja.ShadowPosition.X != 25 || source.Ninja.ShadowRecastSkillID != wID {
		t.Fatalf("r should not overwrite w shadow: %+v", source.Ninja)
	}
	if source.Ninja.RShadowPosition.X != 100 || source.Ninja.RShadowExpiresAt != 172 || source.Ninja.RShadowRecastUntil != 142 {
		t.Fatalf("r shadow = %+v expires %d recastUntil %d, want x=100 expires=172 recastUntil=142", source.Ninja.RShadowPosition, source.Ninja.RShadowExpiresAt, source.Ninja.RShadowRecastUntil)
	}

	Tick(w, source, 29, 20)
	if source.Position.X != 500 || source.Control.UntargetableUntilTick != 0 {
		t.Fatalf("after dash position=%+v untargetable=%d, want x=500 untargetable=0", source.Position, source.Control.UntargetableUntilTick)
	}
	if source.Ninja.RMarkTargetID != target.ID || source.Ninja.RMarkTriggerTick != 89 {
		t.Fatalf("mark target=%q trigger=%d, want %q/89", source.Ninja.RMarkTargetID, source.Ninja.RMarkTriggerTick, target.ID)
	}

	w.ApplyDamage(source, target, 100, 20)
	target.Control.UntargetableUntilTick = 200
	Tick(w, source, 89, 20)
	if target.Stats.HP != 760 {
		t.Fatalf("target hp after mark = %d, want 760", target.Stats.HP)
	}
	if source.Ninja.RMarkTargetID != "" || source.Ninja.RMarkDamage != 0 {
		t.Fatalf("mark state not cleared: %+v", source.Ninja)
	}

	if !SpecialRecast(w, source, protocol.CastInput{SkillID: rID}, source.Skills[rID], rSkill(), 90, 20) {
		t.Fatal("r recast should swap with shadow")
	}
	if source.Position.X != 100 || source.Ninja.RShadowPosition.X != 500 {
		t.Fatalf("r recast positions source=%+v shadow=%+v, want source x=100 shadow x=500", source.Position, source.Ninja.RShadowPosition)
	}
	if SpecialRecast(w, source, protocol.CastInput{SkillID: rID}, source.Skills[rID], rSkill(), 142, 20) {
		t.Fatal("r recast should be closed during final shadow linger")
	}
}

func TestRCanSelectTargetByClickPointOutOfRange(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	sourceID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamBlue, 100, 100)
	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1200, 100)
	source := w.EntityByID(sourceID)
	target := w.EntityByID(targetID)
	source.HeroID = heroID
	target.Stats.HP = 1000

	got := rTarget(w, source, protocol.CastInput{SkillID: rID, TargetID: target.ID, TargetX: 1200, TargetY: 100}, rSkill())

	if got != target {
		t.Fatalf("r target = %v, want %v", got, target)
	}
}

func TestROutOfRangePreparesMoveAndRelease(t *testing.T) {
	skills, err := config.NewSkillStore([]config.SkillConfig{rSkill()})
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(nil, skills, nil, nil, nil)
	sourceID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamBlue, 100, 100)
	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1200, 100)
	source := w.EntityByID(sourceID)
	target := w.EntityByID(targetID)
	source.HeroID = heroID
	source.Stats = world.Stats{HP: 1000, MaxHP: 1000, Attack: 100, AttackSpeed: 1}
	source.Skills = map[string]world.SkillState{rID: {SkillID: rID, Level: 2}}
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000

	CastR(w, source, protocol.CastInput{SkillID: rID, TargetID: target.ID, TargetX: 1200, TargetY: 100}, source.Skills[rID], rSkill(), 10, 20)

	if !source.Ninja.RCastPending || source.Ninja.RCastTargetID != target.ID || source.Ninja.RCastLevel != 2 {
		t.Fatalf("prepared r state = %+v, want pending target level 2", source.Ninja)
	}
	if source.Ninja.RPending || source.Skills[rID].CooldownUntilTick != 0 {
		t.Fatalf("out of range r should not start/cooldown: pending=%v cooldown=%d", source.Ninja.RPending, source.Skills[rID].CooldownUntilTick)
	}
	if source.Intent.MoveTarget == nil || source.Intent.MoveTarget.X != 575 || source.Intent.MoveTarget.Y != 100 {
		t.Fatalf("move target = %+v, want 575,100", source.Intent.MoveTarget)
	}

	source.Position = *source.Intent.MoveTarget
	ReleasePreparedR(w, source, 20, 20)

	if source.Ninja.RCastPending || !source.Ninja.RPending {
		t.Fatalf("release should clear prepared and start r: %+v", source.Ninja)
	}
	if source.Ninja.RReleaseTick != 32 || source.Ninja.RDashEndTick != 39 {
		t.Fatalf("r timing release=%d dashEnd=%d, want 32/39", source.Ninja.RReleaseTick, source.Ninja.RDashEndTick)
	}
	if got, want := source.Skills[rID].CooldownUntilTick, uint64(2220); got != want {
		t.Fatalf("r cooldown = %d, want %d", got, want)
	}
}

func passiveSkill() config.SkillConfig {
	return config.SkillConfig{
		Meta: map[string]float64{
			"heroCooldownSeconds":     10,
			"monsterDamageMultiplier": 0.75,
			"epicMonsterDamageCap":    175,
		},
		MetaLists: map[string][]float64{
			"maxHPRatio":       {0.05, 0.1},
			"maxHPRatioLevels": {1, 18},
		},
	}
}

func eSkill() config.SkillConfig {
	return config.SkillConfig{
		Range: 290,
		Meta: map[string]float64{
			"manaCost":               40,
			"bonusAdRatio":           0.7,
			"slowSeconds":            1.5,
			"wCooldownRefundSeconds": 3,
		},
		MetaLists: map[string][]float64{
			"baseDamage": {70, 92.5, 115, 137.5, 160},
			"shadowSlow": {0.2, 0.25, 0.3, 0.35, 0.4},
			"doubleSlow": {0.3, 0.375, 0.45, 0.45, 0.6},
			"cooldownMs": {5000, 4500, 4000, 3500, 3000},
		},
	}
}

func wSkill() config.SkillConfig {
	return config.SkillConfig{
		Range: 650,
		Meta: map[string]float64{
			"shadowDurationSeconds": 5,
			"shadowDashSeconds":     0.25,
		},
		MetaLists: map[string][]float64{
			"manaCost":     {20, 25, 30, 35, 40},
			"cooldownMs":   {20000, 19000, 18000, 17000, 16000},
			"energyRefund": {30, 35, 40, 45, 50},
		},
	}
}

func onlyNinjaShadowEffect(t *testing.T, w *world.World) world.SkillEffect {
	t.Helper()
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "ninja_shadow" {
			return effect
		}
	}
	t.Fatal("missing ninja shadow effect")
	return world.SkillEffect{}
}

func protocolCast(x float64, y float64) protocol.CastInput {
	return protocol.CastInput{SkillID: wID, TargetX: x, TargetY: y}
}

func qSkill() config.SkillConfig {
	return config.SkillConfig{
		Range: 900,
		Meta: map[string]float64{
			"projectileRadius":  87.5,
			"castWindupSeconds": 0.25,
			"firstBonusAdRatio": 1,
			"laterBonusAdRatio": 0.6,
		},
		MetaLists: map[string][]float64{
			"firstBaseDamage": {80, 120, 160, 200, 240},
			"laterBaseDamage": {48, 72, 96, 120, 144},
		},
	}
}

func countNinjaShurikens(w *world.World) int {
	count := 0
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "ninja_shuriken" {
			count++
		}
	}
	return count
}

func ninjaEEffects(w *world.World) []world.SkillEffect {
	effects := []world.SkillEffect{}
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "ninja_e" {
			effects = append(effects, effect)
		}
	}
	return effects
}

func rSkill() config.SkillConfig {
	return config.SkillConfig{
		SkillID: rID,
		Range:   625,
		Meta: map[string]float64{
			"castWindupSeconds":     0.6,
			"dashSeconds":           0.35,
			"markDelaySeconds":      3,
			"shadowDurationSeconds": 7.5,
			"shadowLingerSeconds":   1.5,
			"targetPickPadding":     80,
			"attackRatio":           1,
		},
		MetaLists: map[string][]float64{
			"storedDamageRatio": {0.25, 0.4, 0.55},
			"cooldownMs":        {120000, 110000, 100000},
		},
	}
}
