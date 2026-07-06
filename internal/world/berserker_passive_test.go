package world

import "testing"

func TestBerserkerQOuterRingBleedsAndHeals(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	outerHero := w.entities["enemy:hero-1"]
	innerMinion := w.entities["minion:red-melee-1"]
	bandHero := &Entity{
		ID:     "enemy:q-band",
		Kind:   EntityKindEnemyHero,
		Team:   TeamRed,
		Radius: 18,
		Stats:  Stats{HP: 1000, MaxHP: 1000},
	}
	outsideHero := &Entity{
		ID:     "enemy:q-outside",
		Kind:   EntityKindEnemyHero,
		Team:   TeamRed,
		Radius: 18,
		Stats:  Stats{HP: 1000, MaxHP: 1000},
	}
	outerMonster := &Entity{
		ID:       "monster:blue",
		Kind:     EntityKindBlueBuff,
		Team:     TeamNeutral,
		Position: Vector2{X: 1380, Y: 1000},
		Radius:   20,
		Stats:    Stats{HP: 1000, MaxHP: 1000},
	}
	w.entities[bandHero.ID] = bandHero
	w.entities[outsideHero.ID] = outsideHero
	w.entities[outerMonster.ID] = outerMonster
	placeEntity(source, 1000, 1000)
	placeEntity(bandHero, 1310, 1000)
	placeEntity(outerHero, 1360, 1000)
	placeEntity(outsideHero, 1430, 1000)
	placeEntity(innerMinion, 1100, 1000)
	source.Stats.HP = source.Stats.MaxHP - 500
	source.Stats.MP = 100
	outerHero.Stats.HP = 1000
	outerHero.Stats.MaxHP = 1000
	outerHero.Stats.PhysicalDefense = 0
	bandHero.Stats.PhysicalDefense = 0
	outsideHero.Stats.PhysicalDefense = 0
	innerMinion.Stats.HP = 1000
	innerMinion.Stats.MaxHP = 1000
	innerMinion.Stats.PhysicalDefense = 0
	learnSkill(source, berserkerQSkillID, 1)

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerQSkillID, source.Position.X, source.Position.Y), 10, nil, 20)

	if got := outerHero.Stats.HP; got != 1000 {
		t.Fatalf("outer hero hp during windup = %d, want 1000", got)
	}
	if got := source.Stats.MP; got != 70 {
		t.Fatalf("mp after q cast = %f, want 70", got)
	}
	effect := onlyBerserkerQEffect(t, w)
	if got, want := effect.ExpiresAt, uint64(25); got != want {
		t.Fatalf("q range expires = %d, want %d", got, want)
	}
	if got, want := effect.Range, 425.0; got != want {
		t.Fatalf("q outer range = %f, want %f", got, want)
	}
	if got, want := effect.Radius, 300.0; got != want {
		t.Fatalf("q inner radius = %f, want %f", got, want)
	}

	w.Tick(25, 20)

	if effect := findBerserkerQEffect(w); effect != nil {
		t.Fatalf("q range after release = %+v, want nil", *effect)
	}
	if got := outerHero.Stats.HP; got != 886 {
		t.Fatalf("outer hero hp = %d, want 886", got)
	}
	if got := bandHero.Stats.HP; got != 886 {
		t.Fatalf("band hero hp = %d, want 886", got)
	}
	if got := outsideHero.Stats.HP; got != 1000 {
		t.Fatalf("outside hero hp = %d, want 1000", got)
	}
	if got := innerMinion.Stats.HP; got != 960 {
		t.Fatalf("inner minion hp = %d, want 960", got)
	}
	if got := outerHero.Passive.Bleeds[source.ID].Stacks; got != 1 {
		t.Fatalf("outer bleed stacks = %d, want 1", got)
	}
	if got := bandHero.Passive.Bleeds[source.ID].Stacks; got != 1 {
		t.Fatalf("band bleed stacks = %d, want 1", got)
	}
	if got := innerMinion.Passive.Bleeds[source.ID].Stacks; got != 0 {
		t.Fatalf("inner bleed stacks = %d, want 0", got)
	}
	if got, want := source.Stats.HP, source.Stats.MaxHP-245; got != want {
		t.Fatalf("hp after q heal = %d, want %d", got, want)
	}
	if got, want := source.Skills[berserkerQSkillID].CooldownUntilTick, uint64(205); got != want {
		t.Fatalf("q cooldown = %d, want %d", got, want)
	}
}

func TestBerserkerQWindupAllowsMovement(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	placeEntity(source, 1000, 1000)
	source.Stats.MP = 100
	source.Stats.MoveSpeed = 100
	learnSkill(source, berserkerQSkillID, 1)

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerQSkillID, source.Position.X, source.Position.Y), 10, nil, 20)
	w.ApplyInput("berserker", protocolPlayerInputMove(1200, 1000), 11, nil, 20)
	w.Tick(12, 20)

	if source.Position.X <= 1000 {
		t.Fatalf("berserker did not move during q windup: %+v", source.Position)
	}
	foundEffect := false
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "berserker_q" && effect.Start != source.Position {
			t.Fatalf("q range center = %+v, want %+v", effect.Start, source.Position)
		}
		if effect.Kind == "berserker_q" {
			foundEffect = true
		}
	}
	if !foundEffect {
		t.Fatal("missing berserker q range effect")
	}
}

func onlyBerserkerQEffect(t *testing.T, w *World) SkillEffect {
	t.Helper()
	effect := findBerserkerQEffect(w)
	if effect == nil {
		t.Fatal("missing berserker q range effect")
	}
	return *effect
}

func findBerserkerQEffect(w *World) *SkillEffect {
	for _, effect := range w.skillEffects {
		if effect.Kind == "berserker_q" {
			return &effect
		}
	}
	return nil
}

func findBerserkerREffect(w *World) *SkillEffect {
	for _, effect := range w.skillEffects {
		if effect.Kind == "berserker_r" {
			return &effect
		}
	}
	return nil
}

func assertBuff(t *testing.T, buffs []BuffState, id string) {
	t.Helper()
	for _, buff := range buffs {
		if buff.ID == id {
			return
		}
	}
	t.Fatalf("missing buff %s", id)
}

func TestBerserkerWEmpowersNextAttackSlowsAndBleedReducesCooldown(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1100, 1000)
	source.Stats.MP = 100
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.PhysicalDefense = 0
	learnSkill(source, berserkerWSkillID, 3)
	target.Passive.Bleeds = map[string]BleedState{
		source.ID: {Stacks: 2, ExpiresAtTick: 200, NextTick: 200},
	}

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerWSkillID, source.Position.X, source.Position.Y), 10, nil, 20)

	if got := source.Stats.MP; got != 70 {
		t.Fatalf("mp after w cast = %f, want 70", got)
	}
	if got := source.Skills[berserkerWSkillID].Stacks; got != 1 {
		t.Fatalf("w stacks after cast = %d, want 1", got)
	}

	w.ApplyInput("berserker", protocolPlayerInputAttack(target.ID), 11, nil, 20)
	w.Tick(12, 20)
	w.Tick(source.Combat.AttackReleaseTick, 20)

	if got := target.Stats.HP; got != 898 {
		t.Fatalf("target hp = %d, want 898", got)
	}
	if got := target.Passive.Bleeds[source.ID].Stacks; got != 3 {
		t.Fatalf("bleed stacks = %d, want 3", got)
	}
	if got, want := source.Skills[berserkerWSkillID].CooldownUntilTick, uint64(97); got != want {
		t.Fatalf("w cooldown = %d, want %d", got, want)
	}
	if got := target.Control.MoveSpeedSlow; got != 0.3 {
		t.Fatalf("move slow = %f, want 0.3", got)
	}
	if got := target.Control.AttackSpeedSlow; got != 0.3 {
		t.Fatalf("attack speed slow = %f, want 0.3", got)
	}
}

func TestBerserkerWResetsBasicAttack(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1100, 1000)
	source.Stats.MP = 100
	source.Stats.AttackSpeed = 1
	target.Stats.PhysicalDefense = 0
	learnSkill(source, berserkerWSkillID, 1)

	w.ApplyInput("berserker", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, source, 20)
	if got, want := source.Combat.NextAttackTick, uint64(30); got != want {
		t.Fatalf("next attack after first hit = %d, want %d", got, want)
	}

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerWSkillID, target.Position.X, target.Position.Y), 16, nil, 20)
	w.Tick(16, 20)

	if got := source.Combat.PendingAttackTargetID; got != target.ID {
		t.Fatalf("pending attack after w reset = %q, want %q", got, target.ID)
	}
	if got, want := source.Combat.AttackReleaseTick, uint64(21); got != want {
		t.Fatalf("second attack release tick = %d, want %d", got, want)
	}
}

func TestBerserkerEPassiveGrantsArmorPenOnUpgrade(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]

	w.ApplyInput("berserker", protocolPlayerInputUpgrade("e"), 1, nil, 20)

	if got := source.Skills[berserkerESkillID].Level; got != 1 {
		t.Fatalf("e level = %d, want 1", got)
	}
	if got := source.Stats.PhysicalPenPercent; got != 0.2 {
		t.Fatalf("armor pen = %f, want 0.2", got)
	}
}

func TestBerserkerEPullsConeTargetsAndSlows(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	offCone := w.entities["minion:red-melee-1"]
	tower := w.entities["structure:red-tower-1"]
	baron := &Entity{
		ID:       "monster:baron",
		Kind:     EntityKindBaronNashor,
		Team:     TeamNeutral,
		Position: Vector2{X: 1340, Y: 1000},
		Radius:   40,
		Stats:    Stats{HP: 8000, MaxHP: 8000},
	}
	w.entities[baron.ID] = baron
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1300, 1000)
	placeEntity(offCone, 1000, 1300)
	placeEntity(tower, 1320, 1000)
	source.Stats.MP = 100
	learnSkill(source, berserkerESkillID, 2)

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerESkillID, 1500, 1000), 10, nil, 20)

	if got := source.Stats.MP; got != 55 {
		t.Fatalf("mp after e = %f, want 55", got)
	}
	if got := target.Position; got != (Vector2{X: 1300, Y: 1000}) {
		t.Fatalf("target moved during windup to %+v", got)
	}

	w.Tick(15, 20)

	if got, want := distance(source.Position, target.Position), source.Radius+target.Radius+1; got != want {
		t.Fatalf("pulled target distance = %f, want %f", got, want)
	}
	if got := target.Control.MoveSpeedSlow; got != 0.4 {
		t.Fatalf("target slow = %f, want 0.4", got)
	}
	if got, want := target.Control.MoveSpeedSlowUntil, uint64(35); got != want {
		t.Fatalf("slow until = %d, want %d", got, want)
	}
	if got, want := target.Control.AirborneUntilTick, uint64(20); got != want {
		t.Fatalf("airborne until = %d, want %d", got, want)
	}
	if got := target.Passive.Bleeds[source.ID].Stacks; got != 0 {
		t.Fatalf("bleed stacks = %d, want 0", got)
	}
	if got := offCone.Position; got != (Vector2{X: 1000, Y: 1300}) {
		t.Fatalf("off cone moved to %+v", got)
	}
	if got := tower.Position; got != (Vector2{X: 1320, Y: 1000}) {
		t.Fatalf("tower moved to %+v", got)
	}
	if got := baron.Position; got != (Vector2{X: 1340, Y: 1000}) {
		t.Fatalf("baron moved to %+v", got)
	}
	if got, want := source.Skills[berserkerESkillID].CooldownUntilTick, uint64(435); got != want {
		t.Fatalf("e cooldown = %d, want %d", got, want)
	}
}

func TestBerserkerRExecutesRefreshesAndGrantsVision(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	w.SpawnHero("red", testHeroConfig(), TeamRed)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	nextTarget := w.entities[playerEntityID("red")]
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1200, 1000)
	placeEntity(nextTarget, 1220, 1000)
	source.Stats.MP = 100
	target.Stats.HP = 200
	target.Stats.MaxHP = 1200
	nextTarget.Stats.HP = 100
	nextTarget.Stats.MaxHP = 1000
	learnSkill(source, berserkerRSkillID, 1)
	target.Passive.Bleeds = map[string]BleedState{
		source.ID: {Stacks: 5, ExpiresAtTick: 200, NextTick: 200},
	}

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerRSkillID, target.Position.X, target.Position.Y), 10, nil, 20)

	if got := target.Stats.HP; got != 200 {
		t.Fatalf("target hp during r windup = %d, want 200", got)
	}
	if got := source.Stats.MP; got != 0 {
		t.Fatalf("mp after r cast = %f, want 0", got)
	}

	w.Tick(20, 20)

	if target.Stats.HP != 0 {
		t.Fatalf("target hp after r = %d, want 0", target.Stats.HP)
	}
	if got, want := source.Berserker.BloodRageUntil, uint64(120); got != want {
		t.Fatalf("blood rage until = %d, want %d", got, want)
	}
	if got, want := source.Berserker.NoxianGuillotineRecast, uint64(420); got != want {
		t.Fatalf("r recast until = %d, want %d", got, want)
	}
	if got, want := distance(source.Position, target.Position), 175.0; got != want {
		t.Fatalf("r jump distance = %f, want %f", got, want)
	}
	if got, want := source.Skills[berserkerRSkillID].CooldownUntilTick, uint64(2420); got != want {
		t.Fatalf("r cooldown after execute = %d, want %d", got, want)
	}
	assertBuff(t, w.ActiveBuffs(source, 20), "berserker_noxian_guillotine_recast")
	if len(w.skillEffects) == 0 {
		t.Fatal("r should create vision effect")
	}

	mpBeforeRecast := source.Stats.MP
	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerRSkillID, nextTarget.Position.X, nextTarget.Position.Y), 30, nil, 20)
	if got := source.Stats.MP; got != mpBeforeRecast {
		t.Fatalf("mp after free recast = %f, want %f", got, mpBeforeRecast)
	}
	w.Tick(40, 20)
	if nextTarget.Stats.HP != 0 {
		t.Fatalf("next target hp after recast = %d, want 0", nextTarget.Stats.HP)
	}
	if got, want := source.Skills[berserkerRSkillID].CooldownUntilTick, uint64(2440); got != want {
		t.Fatalf("r cooldown after recast execute = %d, want %d", got, want)
	}
}

func TestBerserkerRWalksIntoRangeBeforeWindup(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1700, 1000)
	source.Stats.MP = 100
	source.Stats.MPRegen5 = 0
	source.Stats.MoveSpeed = 4800
	target.Stats.HP = 1000
	learnSkill(source, berserkerRSkillID, 1)

	w.ApplyInput("berserker", protocolPlayerInputCastTarget(berserkerRSkillID, target.ID, target.Position.X, target.Position.Y), 10, nil, 20)

	if got := source.Stats.MP; got != 100 {
		t.Fatalf("mp before r range = %f, want 100", got)
	}
	if !source.Berserker.NoxianGuillotineCastPending {
		t.Fatal("r should wait for range")
	}
	if effect := findBerserkerREffect(w); effect != nil {
		t.Fatalf("r range effect before range = %+v, want nil", *effect)
	}

	w.Tick(11, 20)

	if source.Berserker.NoxianGuillotineCastPending {
		t.Fatal("r should start windup after walking into range")
	}
	if got := source.Stats.MP; got != 0 {
		t.Fatalf("mp after r windup starts = %f, want 0", got)
	}
	if got := source.Skills[berserkerRSkillID].Stacks; got != 1 {
		t.Fatalf("r stacks after range = %d, want 1", got)
	}
	effect := findBerserkerREffect(w)
	if effect == nil {
		t.Fatal("missing r range effect after range")
	}
	if got, want := effect.ExpiresAt, uint64(21); got != want {
		t.Fatalf("r range expires = %d, want %d", got, want)
	}
}

func TestBerserkerRPreparedCastCancelsOnMove(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1700, 1000)
	source.Stats.MP = 100
	learnSkill(source, berserkerRSkillID, 1)

	w.ApplyInput("berserker", protocolPlayerInputCastTarget(berserkerRSkillID, target.ID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.ApplyInput("berserker", protocolPlayerInputMove(900, 1000), 11, nil, 20)

	if source.Berserker.NoxianGuillotineCastPending {
		t.Fatal("r prepared cast should cancel on move")
	}
	if got := source.Stats.MP; got != 100 {
		t.Fatalf("mp after cancel = %f, want 100", got)
	}
	if got := source.Skills[berserkerRSkillID].Stacks; got != 0 {
		t.Fatalf("r stacks after cancel = %d, want 0", got)
	}
}

func TestBerserkerRLevelThreeKillPermanentlyRefreshesCooldown(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	placeEntity(source, 1000, 1000)
	placeEntity(target, 1200, 1000)
	source.Stats.MP = 100
	target.Stats.HP = 100
	learnSkill(source, berserkerRSkillID, 3)

	w.ApplyInput("berserker", protocolPlayerInputCast(berserkerRSkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	if got := source.Stats.MP; got != 100 {
		t.Fatalf("mp after rank 3 r cast = %f, want 100", got)
	}
	w.Tick(20, 20)

	if got, want := source.Skills[berserkerRSkillID].CooldownUntilTick, uint64(20); got != want {
		t.Fatalf("r cooldown after rank 3 execute = %d, want %d", got, want)
	}
	if got := source.Berserker.NoxianGuillotineRecast; got != 0 {
		t.Fatalf("rank 3 recast window = %d, want 0", got)
	}
}

func TestBerserkerBleedStacksBloodRageAndTicks(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	baseAttack := source.Stats.Attack

	for tick := uint64(1); tick <= 5; tick++ {
		target.Combat.LastHitTick = tick
		w.applyBasicAttackDamage(source, target, 1, 20)
	}

	bleed := target.Passive.Bleeds[source.ID]
	if bleed.Stacks != 5 {
		t.Fatalf("bleed stacks = %d, want 5", bleed.Stacks)
	}
	if got, want := source.Berserker.BloodRageUntil, uint64(105); got != want {
		t.Fatalf("blood rage until = %d, want %d", got, want)
	}
	if got, want := source.Stats.Attack, baseAttack+30; got != want {
		t.Fatalf("blood rage attack = %f, want %f", got, want)
	}

	beforeHP := target.Stats.HP
	w.Tick(21, 20)
	if target.Stats.HP >= beforeHP {
		t.Fatalf("bleed tick hp = %d, want below %d", target.Stats.HP, beforeHP)
	}
	if len(target.Combat.DamageEvents) != 1 || target.Combat.DamageEvents[0].BasicAttack {
		t.Fatalf("bleed damage events = %+v", target.Combat.DamageEvents)
	}
}

func TestBerserkerBleedDealsBonusDamageToMonsters(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	monster := &Entity{
		ID:     "monster:blue",
		Kind:   EntityKindBlueBuff,
		Team:   TeamNeutral,
		Radius: 20,
		Stats:  Stats{HP: 1000, MaxHP: 1000},
	}
	w.entities[monster.ID] = monster
	monster.Passive.Bleeds = map[string]BleedState{
		source.ID: {Stacks: 1, ExpiresAtTick: 200, NextTick: 21},
	}

	w.Tick(21, 20)

	if got, want := monster.Stats.HP, 994; got != want {
		t.Fatalf("monster hp after bleed tick = %d, want %d", got, want)
	}
}

func TestBerserkerBloodRageDamageAppliesFullBleed(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["enemy:hero-1"]
	source.Berserker.BloodRageUntil = 100
	w.recalculatePlayerStats(source)

	target.Combat.LastHitTick = 1
	w.applyDamage(source, target, 1, 20)

	if got := target.Passive.Bleeds[source.ID].Stacks; got != 5 {
		t.Fatalf("bleed stacks during blood rage = %d, want 5", got)
	}
}

func TestBerserkerMinionBleedDoesNotTriggerBloodRage(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(berserkerHeroID)
	if !ok {
		t.Fatal("berserker hero not found")
	}
	w.SpawnHero("berserker", hero, TeamBlue)
	source := w.entities[playerEntityID("berserker")]
	target := w.entities["minion:red-melee-1"]

	for tick := uint64(1); tick <= 5; tick++ {
		target.Combat.LastHitTick = tick
		w.applyBasicAttackDamage(source, target, 1, 20)
	}

	if got := target.Passive.Bleeds[source.ID].Stacks; got != 5 {
		t.Fatalf("minion bleed stacks = %d, want 5", got)
	}
	if got := source.Berserker.BloodRageUntil; got != 0 {
		t.Fatalf("blood rage from minion = %d, want 0", got)
	}
}
