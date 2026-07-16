package monk

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"testing"
)

func TestFlurryEmpowersTwoAttacksAndRefundsCooldowns(t *testing.T) {
	w, player := testWorld(t)
	learn(player, qID, 1)
	learn(player, wID, 1)
	learn(player, eID, 1)

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID}}, 10, nil, 20)
	if player.Passive.MonkFlurryAttacks != 2 || player.Passive.MonkFlurryUntil != 70 {
		t.Fatalf("flurry state = attacks %d until %d, want 2/70", player.Passive.MonkFlurryAttacks, player.Passive.MonkFlurryUntil)
	}
	if got, want := world.EffectiveAttackSpeedAtTick(player, 10), player.Stats.AttackSpeed*1.4; math.Abs(got-want) > 0.000001 {
		t.Fatalf("flurry attack speed = %f, want %f", got, want)
	}

	player.Stats.MP = 100
	player.Stats.MPRegen5 = 0
	player.Position = world.Vector2{X: 1000, Y: 1000}
	state := player.Skills[wID]
	state.CooldownUntilTick = 100
	player.Skills[wID] = state

	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, player.Position.X+100, player.Position.Y)
	if !ok {
		t.Fatal("spawn target failed")
	}
	attackTick := player.Combat.NextAttackTick
	w.ApplyInput("monk", protocol.PlayerInput{Attack: &protocol.AttackInput{TargetID: targetID}}, attackTick, nil, 20)
	w.Tick(attackTick, 20)
	w.Tick(player.Combat.AttackReleaseTick, 20)
	if got := player.Stats.MP; got != 120 {
		t.Fatalf("energy after first flurry attack = %f, want 120", got)
	}
	if got := player.Skills[wID].CooldownUntilTick; got != 90 {
		t.Fatalf("w cooldown after first flurry attack = %d, want 90", got)
	}
	if player.Passive.MonkFlurryAttacks != 1 {
		t.Fatalf("flurry attacks after first release = %d, want 1", player.Passive.MonkFlurryAttacks)
	}

	attackTick = player.Combat.NextAttackTick
	w.ApplyInput("monk", protocol.PlayerInput{Attack: &protocol.AttackInput{TargetID: targetID}}, attackTick, nil, 20)
	w.Tick(attackTick, 20)
	w.Tick(player.Combat.AttackReleaseTick, 20)
	if got := player.Stats.MP; got != 130 {
		t.Fatalf("energy after second flurry attack = %f, want 130", got)
	}
	if player.Passive.MonkFlurryAttacks != 0 || player.Passive.MonkFlurryUntil != 0 {
		t.Fatalf("flurry should be consumed: %+v", player.Passive)
	}
}

func TestFlurryEnergyRefundTiers(t *testing.T) {
	tests := []struct {
		level  int
		first  float64
		second float64
	}{
		{level: 1, first: 20, second: 10},
		{level: 7, first: 30, second: 15},
		{level: 13, first: 40, second: 20},
	}
	for _, tt := range tests {
		if got := energyRefund(tt.level, 1); got != tt.first {
			t.Fatalf("level %d first refund = %f, want %f", tt.level, got, tt.first)
		}
		if got := energyRefund(tt.level, 2); got != tt.second {
			t.Fatalf("level %d second refund = %f, want %f", tt.level, got, tt.second)
		}
	}
}

func TestSonicWaveMarksAndEchoStrikeExecutes(t *testing.T) {
	w, player := testWorld(t)
	learn(player, qID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	player.Stats.MPRegen5 = 0
	player.Stats.BonusAttack = 100

	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("spawn target failed")
	}
	target := w.EntityByID(targetID)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 1300, TargetY: 1000}}, 10, nil, 20)
	w.Tick(15, 20)
	hitTick := tickUntilMarked(t, w, player, target, 16, 30, 20)
	assertSkillEffect(t, w, "monk_q_mark")
	if got := target.Combat.LastDamage; got != 160 {
		t.Fatalf("sonic wave damage = %d, want 160", got)
	}
	if player.Passive.MonkQMarkTargetID != target.ID || player.Passive.MonkQMarkUntil != hitTick+60 {
		t.Fatalf("q mark target/until = %q/%d, want %q/%d", player.Passive.MonkQMarkTargetID, player.Passive.MonkQMarkUntil, target.ID, hitTick+60)
	}
	if got := player.Stats.MP; got != 150 {
		t.Fatalf("energy after sonic wave = %f, want 150", got)
	}

	target.Stats.HP = 400
	beforeMP := player.Stats.MP
	castTick := hitTick + 1
	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID}}, castTick, nil, 20)
	assertSkillEffect(t, w, "monk_q_echo")
	if got := player.Stats.MP; got != beforeMP {
		t.Fatalf("echo strike energy = %f, want %f", got, beforeMP)
	}
	if got := target.Stats.HP; got != 400 {
		t.Fatalf("echo strike dealt damage before arrival: hp = %f, want 400", got)
	}
	if player.Passive.MonkQMarkTargetID != "" {
		t.Fatalf("q mark should be consumed: %+v", player.Passive)
	}
	if got, want := player.Control.DashUntilTick, castTick+4; got != want {
		t.Fatalf("echo strike dash until = %d, want %d at 1400 units/s", got, want)
	}
	w.Tick(castTick+1, 20)
	interruptedAt := player.Position
	w.ApplyAirborne(player, castTick+10, castTick+1, 20)
	if player.Control.DashUntilTick != 0 {
		t.Fatalf("airborne should interrupt echo strike, dash until = %d", player.Control.DashUntilTick)
	}
	if hasSkillEffect(w, "monk_q_echo") {
		t.Fatal("airborne should remove echo strike effect")
	}
	w.Tick(castTick+4, 20)
	if got := target.Stats.HP; got != 400 {
		t.Fatalf("interrupted echo strike damage: hp = %f, want 400", got)
	}
	if player.Position != interruptedAt {
		t.Fatalf("interrupted echo strike moved from %+v to %+v", interruptedAt, player.Position)
	}
}

func TestEchoStrikeDealsDamageOnlyAtDestination(t *testing.T) {
	w, player := testWorld(t)
	learn(player, qID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	player.Stats.BonusAttack = 100
	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("spawn target failed")
	}
	target := w.EntityByID(targetID)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 1300, TargetY: 1000}}, 10, nil, 20)
	w.Tick(15, 20)
	hitTick := tickUntilMarked(t, w, player, target, 16, 30, 20)
	target.Stats.HP = 400

	castTick := hitTick + 1
	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID}}, castTick, nil, 20)
	arrivalTick := player.Control.DashUntilTick
	for tick := castTick + 1; tick < arrivalTick; tick++ {
		w.Tick(tick, 20)
		if target.Stats.HP != 400 {
			t.Fatalf("echo strike dealt damage at tick %d before arrival %d: hp = %f", tick, arrivalTick, target.Stats.HP)
		}
	}
	w.Tick(arrivalTick, 20)
	if got := target.Combat.LastDamage; got != 208 {
		t.Fatalf("echo strike damage at destination = %d, want 208", got)
	}
}

func TestSonicWaveBlockedByWindWallDoesNotMark(t *testing.T) {
	w, player := testWorld(t)
	learn(player, qID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("spawn target failed")
	}
	target := w.EntityByID(targetID)
	w.PutWindWall(world.WindWall{
		ID:        "windwall:test",
		Team:      world.TeamRed,
		Center:    world.Vector2{X: 1150, Y: 1000},
		Dir:       world.Vector2{X: 0, Y: 1},
		Width:     500,
		ExpiresAt: 100,
	})

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 1300, TargetY: 1000}}, 10, nil, 20)
	for tick := uint64(15); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}
	if target.Combat.LastDamage != 0 {
		t.Fatalf("blocked sonic wave damage = %d, want 0", target.Combat.LastDamage)
	}
	if player.Passive.MonkQMarkTargetID != "" {
		t.Fatalf("blocked sonic wave should not mark: %+v", player.Passive)
	}
}

func TestSonicWaveUsesCursorDirectionNotTargetID(t *testing.T) {
	w, player := testWorld(t)
	learn(player, qID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200

	lockedID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1000, 1300)
	if !ok {
		t.Fatal("spawn locked target failed")
	}
	cursorID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("spawn cursor target failed")
	}
	locked := w.EntityByID(lockedID)
	cursor := w.EntityByID(cursorID)
	locked.Stats.PhysicalDefense = 0
	cursor.Stats.PhysicalDefense = 0

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{
		SkillID:  qID,
		TargetID: locked.ID,
		TargetX:  1300,
		TargetY:  1000,
	}}, 10, nil, 20)
	w.Tick(15, 20)
	tickUntilMarked(t, w, player, cursor, 16, 30, 20)

	if locked.Combat.LastDamage != 0 {
		t.Fatalf("sonic wave should ignore target id, locked target damage = %d", locked.Combat.LastDamage)
	}
	if cursor.Combat.LastDamage == 0 {
		t.Fatal("sonic wave did not hit cursor-direction target")
	}
}

func TestEchoStrikeTracksMovingMarkedTarget(t *testing.T) {
	w, player := testWorld(t)
	learn(player, qID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200

	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1300, 1000)
	if !ok {
		t.Fatal("spawn target failed")
	}
	target := w.EntityByID(targetID)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 1300, TargetY: 1000}}, 10, nil, 20)
	w.Tick(15, 20)
	hitTick := tickUntilMarked(t, w, player, target, 16, 30, 20)

	castTick := hitTick + 1
	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID}}, castTick, nil, 20)
	originalEnd := player.Control.DashEnd
	target.Position = world.Vector2{X: 1900, Y: 1000}

	for tick := castTick + 1; tick <= castTick+30; tick++ {
		w.Tick(tick, 20)
	}
	if player.Position.X <= originalEnd.X+100 {
		t.Fatalf("echo strike stopped near old target position: got %+v old end %+v", player.Position, originalEnd)
	}
	if distance(player.Position, target.Position) > player.Radius+target.Radius+1 {
		t.Fatalf("echo strike did not follow moving target: player %+v target %+v", player.Position, target.Position)
	}
}

func TestSafeguardShieldsAllyHeroAndHalvesCooldown(t *testing.T) {
	w, player := testWorld(t)
	learn(player, wID, 2)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	player.Stats.MPRegen5 = 0
	player.Stats.AbilityPower = 50
	allyHero, ok := testHeroConfig(t)
	if !ok {
		t.Fatal("monk hero missing")
	}
	w.SpawnHero("ally", allyHero, world.TeamBlue)
	ally := w.EntityByID("player:ally")
	ally.Position = world.Vector2{X: 1300, Y: 1000}

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID, TargetID: ally.ID}}, 10, nil, 20)
	assertSkillEffect(t, w, "monk_w_safeguard")

	if got := player.Stats.MP; got != 150 {
		t.Fatalf("energy after safeguard = %f, want 150", got)
	}
	if got := player.Skills[wID].CooldownUntilTick; got != 90 {
		t.Fatalf("w cooldown = %d, want 90", got)
	}
	if player.Passive.Shield != 140 || ally.Passive.Shield != 140 {
		t.Fatalf("shields self/ally = %d/%d, want 140/140", player.Passive.Shield, ally.Passive.Shield)
	}
	if player.Passive.MonkWRecastUntil != 70 {
		t.Fatalf("w recast until = %d, want 70", player.Passive.MonkWRecastUntil)
	}
	if got, want := player.Control.DashUntilTick, uint64(14); got != want {
		t.Fatalf("safeguard dash until = %d, want %d at 1400 units/s", got, want)
	}
	if player.Position != (world.Vector2{X: 1000, Y: 1000}) {
		t.Fatalf("safeguard should not teleport, position = %+v", player.Position)
	}
	w.Tick(14, 20)
	if distance(player.Position, ally.Position) > player.Radius+ally.Radius+0.000001 {
		t.Fatalf("safeguard should move next to ally, positions self=%+v ally=%+v", player.Position, ally.Position)
	}
}

func TestSafeguardCanTargetSelf(t *testing.T) {
	w, player := testWorld(t)
	learn(player, wID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID, TargetX: player.Position.X, TargetY: player.Position.Y}}, 10, nil, 20)

	if player.Position != (world.Vector2{X: 1000, Y: 1000}) {
		t.Fatalf("self safeguard moved player to %+v", player.Position)
	}
	if player.Passive.Shield != 60 {
		t.Fatalf("self safeguard shield = %d, want 60", player.Passive.Shield)
	}
	if got := player.Skills[wID].CooldownUntilTick; got != 90 {
		t.Fatalf("self safeguard cooldown = %d, want 90", got)
	}
}

func TestSafeguardToMinionDoesNotHalveCooldown(t *testing.T) {
	w, player := testWorld(t)
	learn(player, wID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	minionID, ok := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamBlue, 1300, 1000)
	if !ok {
		t.Fatal("spawn minion failed")
	}

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID, TargetID: minionID}}, 10, nil, 20)

	if got := player.Skills[wID].CooldownUntilTick; got != 170 {
		t.Fatalf("w cooldown to minion = %d, want 170", got)
	}
	if player.Passive.Shield != 60 {
		t.Fatalf("self shield = %d, want 60", player.Passive.Shield)
	}
}

func TestIronWillGrantsBaseArmorAndOmnivamp(t *testing.T) {
	w, player := testWorld(t)
	learn(player, wID, 3)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	minionID, ok := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamBlue, 1300, 1000)
	if !ok {
		t.Fatal("spawn minion failed")
	}

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID, TargetID: minionID}}, 10, nil, 20)
	beforeArmor := player.Stats.PhysicalDefense
	beforeOmnivamp := player.Stats.Omnivamp
	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID}}, 11, nil, 20)
	assertSkillEffect(t, w, "monk_w_iron_will")

	if got := player.Stats.MP; got != 120 {
		t.Fatalf("energy after iron will = %f, want 120", got)
	}
	wantArmor := beforeArmor + (beforeArmor-player.Stats.BonusPhysicalDefense)*0.15
	if math.Abs(player.Stats.PhysicalDefense-wantArmor) > 0.000001 {
		t.Fatalf("iron will armor = %f, want %f", player.Stats.PhysicalDefense, wantArmor)
	}
	if math.Abs(player.Stats.Omnivamp-(beforeOmnivamp+0.15)) > 0.000001 {
		t.Fatalf("iron will omnivamp = %f, want %f", player.Stats.Omnivamp, beforeOmnivamp+0.15)
	}
	if player.Passive.MonkWRecastUntil != 0 || player.Passive.MonkWIronWillUntil != 91 {
		t.Fatalf("w recast/iron will state = %d/%d, want 0/91", player.Passive.MonkWRecastUntil, player.Passive.MonkWIronWillUntil)
	}
}

func TestTempestDamagesRevealsAndEnablesCripple(t *testing.T) {
	w, player := testWorld(t)
	learn(player, eID, 1)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	player.Stats.MPRegen5 = 0
	player.Stats.BonusAttack = 40

	insideID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1200, 1000)
	if !ok {
		t.Fatal("spawn inside target failed")
	}
	outsideID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1500, 1000)
	if !ok {
		t.Fatal("spawn outside target failed")
	}
	inside := w.EntityByID(insideID)
	outside := w.EntityByID(outsideID)
	inside.Stats.MagicDefense = 0
	outside.Stats.MagicDefense = 0

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID}}, 10, nil, 20)
	if got := player.Stats.MP; got != 150 {
		t.Fatalf("energy after tempest cast = %f, want 150", got)
	}
	if got := player.Skills[eID].CooldownUntilTick; got != 210 {
		t.Fatalf("e cooldown = %d, want 210", got)
	}
	w.Tick(15, 20)
	assertSkillEffect(t, w, "monk_e_tempest")

	if got := inside.Combat.LastDamage; got != 100 {
		t.Fatalf("tempest damage = %d, want 100", got)
	}
	if outside.Combat.LastDamage != 0 {
		t.Fatalf("outside target damage = %d, want 0", outside.Combat.LastDamage)
	}
	if player.Passive.MonkERecastUntil != 75 || !player.Passive.MonkEHitIDs[inside.ID] {
		t.Fatalf("e recast state = until %d hits %+v, want inside until 75", player.Passive.MonkERecastUntil, player.Passive.MonkEHitIDs)
	}
}

func TestCrippleSlowsOnlyTempestHitsAndDecays(t *testing.T) {
	w, player := testWorld(t)
	learn(player, eID, 3)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.MP = 200
	player.Stats.MPRegen5 = 0

	insideID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1200, 1000)
	if !ok {
		t.Fatal("spawn inside target failed")
	}
	outsideID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1500, 1000)
	if !ok {
		t.Fatal("spawn outside target failed")
	}
	inside := w.EntityByID(insideID)
	outside := w.EntityByID(outsideID)

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID}}, 10, nil, 20)
	w.Tick(15, 20)
	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID}}, 16, nil, 20)
	assertSkillEffect(t, w, "monk_e_cripple")

	if got := player.Stats.MP; got != 120 {
		t.Fatalf("energy after cripple = %f, want 120", got)
	}
	if got := inside.Control.MoveSpeedSlow; math.Abs(got-0.4) > 0.000001 {
		t.Fatalf("initial slow = %f, want 0.4", got)
	}
	if outside.Control.MoveSpeedSlow != 0 {
		t.Fatalf("outside slow = %f, want 0", outside.Control.MoveSpeedSlow)
	}
	w.Tick(56, 20)
	if got := inside.Control.MoveSpeedSlow; math.Abs(got-0.2) > 0.000001 {
		t.Fatalf("decayed slow = %f, want 0.2", got)
	}
	w.Tick(96, 20)
	if inside.Control.MoveSpeedSlow != 0 || inside.Control.MoveSpeedSlowUntil != 0 {
		t.Fatalf("expired slow = %f/%d, want 0/0", inside.Control.MoveSpeedSlow, inside.Control.MoveSpeedSlowUntil)
	}
}

func TestDragonRageKnocksBackAndDamagesCollisionTargets(t *testing.T) {
	w, player := testWorld(t)
	learn(player, rID, 2)
	player.Position = world.Vector2{X: 1000, Y: 1000}
	player.Stats.BonusAttack = 100

	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1200, 1000)
	if !ok {
		t.Fatal("spawn kick target failed")
	}
	collideID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1700, 1000)
	if !ok {
		t.Fatal("spawn collision target failed")
	}
	missID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1700, 1200)
	if !ok {
		t.Fatal("spawn miss target failed")
	}
	target := w.EntityByID(targetID)
	collide := w.EntityByID(collideID)
	miss := w.EntityByID(missID)
	target.Stats.PhysicalDefense = 0
	target.Stats.BonusHP = 500
	collide.Stats.PhysicalDefense = 0
	miss.Stats.PhysicalDefense = 0

	w.ApplyInput("monk", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID, TargetID: target.ID}}, 10, nil, 20)
	if got := player.Skills[rID].CooldownUntilTick; got != 1710 {
		t.Fatalf("r cooldown = %d, want 1710", got)
	}
	if player.Passive.MonkFlurryAttacks != 2 {
		t.Fatalf("r should trigger flurry, got %d attacks", player.Passive.MonkFlurryAttacks)
	}
	w.Tick(15, 20)

	if got := target.Combat.LastDamage; got != 600 {
		t.Fatalf("kick target damage = %d, want 600", got)
	}
	if got := collide.Combat.LastDamage; got != 660 {
		t.Fatalf("collision damage = %d, want 660", got)
	}
	if miss.Combat.LastDamage != 0 {
		t.Fatalf("miss target damage = %d, want 0", miss.Combat.LastDamage)
	}
	if target.Control.DashUntilTick != 35 || target.Control.AirborneUntilTick != 35 {
		t.Fatalf("kick target control = dash %d airborne %d, want 35/35", target.Control.DashUntilTick, target.Control.AirborneUntilTick)
	}
	if collide.Control.AirborneUntilTick != 35 {
		t.Fatalf("collision airborne = %d, want 35", collide.Control.AirborneUntilTick)
	}
	w.Tick(35, 20)
	if math.Abs(target.Position.X-2400) > 0.000001 || math.Abs(target.Position.Y-1000) > 0.000001 {
		t.Fatalf("kick target end position = %+v, want 2400/1000", target.Position)
	}
}

func testWorld(t *testing.T) (*world.World, *world.Entity) {
	t.Helper()
	heroes, err := config.LoadHeroes("../../../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.LoadSkills("../../../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(heroes, skills, nil, nil, nil)
	hero, ok := heroes.Get(heroID)
	if !ok {
		t.Fatal("monk hero missing")
	}
	w.SpawnHero("monk", hero, world.TeamBlue)
	player := w.EntityByID("player:monk")
	if player == nil {
		t.Fatal("monk player missing")
	}
	return w, player
}

func testHeroConfig(t *testing.T) (config.HeroConfig, bool) {
	t.Helper()
	heroes, err := config.LoadHeroes("../../../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	return heroes.Get(heroID)
}

func tickUntilMarked(t *testing.T, w *world.World, player *world.Entity, target *world.Entity, from uint64, to uint64, tickRate int) uint64 {
	t.Helper()
	for tick := from; tick <= to; tick++ {
		w.Tick(tick, tickRate)
		if player.Passive.MonkQMarkTargetID == target.ID {
			return tick
		}
	}
	t.Fatalf("target %s was not marked by tick %d", target.ID, to)
	return 0
}

func learn(entity *world.Entity, skillID string, level int) {
	state := entity.Skills[skillID]
	state.SkillID = skillID
	state.Level = level
	entity.Skills[skillID] = state
}

func assertSkillEffect(t *testing.T, w *world.World, kind string) {
	t.Helper()
	if hasSkillEffect(w, kind) {
		return
	}
	t.Fatalf("missing skill effect %q", kind)
}

func hasSkillEffect(w *world.World, kind string) bool {
	for _, effect := range w.SkillEffects() {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}
