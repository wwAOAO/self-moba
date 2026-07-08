package firemage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestSearReleasesProjectileAfterWindup(t *testing.T) {
	w, source := testWorld(t)
	learnQ(source, 1)
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.MP = 100

	CastQ(w, source, protocol.CastInput{SkillID: qID, TargetX: 1600, TargetY: 1000}, source.Skills[qID], w.SkillConfig(qID), 10, 20)
	if got, want := source.Stats.MP, 50.0; got != want {
		t.Fatalf("mp after sear = %v, want %v", got, want)
	}
	if got, want := source.Skills[qID].CooldownUntilTick, uint64(170); got != want {
		t.Fatalf("cooldown tick = %d, want %d", got, want)
	}

	w.Tick(14, 20)

	if hasSearProjectile(w) {
		t.Fatal("sear projectile released before windup finished")
	}

	w.Tick(15, 20)

	if !hasSearProjectile(w) {
		t.Fatal("sear projectile was not released after windup")
	}
}

func TestSearHitDamagesAndAddsBurn(t *testing.T) {
	w, source := testWorld(t)
	target := searTarget(t, w, source)
	learnQ(source, 1)

	castSear(w, source, 10)
	w.Tick(15, 20)
	w.Tick(16, 20)
	w.Tick(17, 20)

	if got, want := target.Combat.LastDamage, 80; got != want {
		t.Fatalf("sear damage = %d, want %d", got, want)
	}
	if got := target.Passive.FireBurns[source.ID].Stacks; got != 1 {
		t.Fatalf("burn stacks = %d, want 1", got)
	}
}

func TestSearStunsAlreadyBurningTarget(t *testing.T) {
	w, source := testWorld(t)
	target := searTarget(t, w, source)
	learnQ(source, 1)
	applyBurn(w, source, target, 10, 20)

	castSear(w, source, 20)
	w.Tick(25, 20)
	w.Tick(26, 20)
	w.Tick(27, 20)

	if got, want := target.Control.StunnedUntilTick, uint64(67); got != want {
		t.Fatalf("stunned until = %d, want %d", got, want)
	}
	if got := target.Passive.FireBurns[source.ID].Stacks; got != 2 {
		t.Fatalf("burn stacks = %d, want 2", got)
	}
}

func TestPillarOfFlameTriggersAfterDelay(t *testing.T) {
	w, source := testWorld(t)
	target := pillarTarget(t, w, source)
	learnW(source, 1)
	source.Stats.MP = 200

	CastW(w, source, protocol.CastInput{SkillID: wID, TargetX: 1200, TargetY: 1000}, source.Skills[wID], w.SkillConfig(wID), 10, 20)
	if got, want := source.Stats.MP, 130.0; got != want {
		t.Fatalf("mp after pillar = %v, want %v", got, want)
	}
	if got, want := source.Skills[wID].CooldownUntilTick, uint64(210); got != want {
		t.Fatalf("cooldown tick = %d, want %d", got, want)
	}

	w.Tick(24, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatalf("pillar damaged before delay: %d", target.Combat.LastDamage)
	}
	w.Tick(25, 20)

	if got, want := target.Combat.LastDamage, 75; got != want {
		t.Fatalf("pillar damage = %d, want %d", got, want)
	}
	if got := target.Passive.FireBurns[source.ID].Stacks; got != 1 {
		t.Fatalf("burn stacks = %d, want 1", got)
	}
}

func TestPillarOfFlameDealsBonusDamageToBurningTargets(t *testing.T) {
	w, source := testWorld(t)
	target := pillarTarget(t, w, source)
	learnW(source, 1)
	applyBurn(w, source, target, 10, 20)

	castPillar(w, source, 11)
	w.Tick(26, 20)

	if got, want := target.Combat.LastDamage, 94; got != want {
		t.Fatalf("pillar burning damage = %d, want %d", got, want)
	}
	if got := target.Passive.FireBurns[source.ID].Stacks; got != 2 {
		t.Fatalf("burn stacks = %d, want 2", got)
	}
}

func TestPillarOfFlameOutOfRangeWalksBeforeCastingAtTargetPoint(t *testing.T) {
	w, source := testWorld(t)
	learnW(source, 1)
	source.Stats.MP = 200
	source.Position = world.Vector2{X: 1000, Y: 1000}
	target := spawnEnemyHero(t, w, 2200, 1000)
	target.Stats.MagicDefense = 0

	CastW(w, source, protocol.CastInput{SkillID: wID, TargetX: 2200, TargetY: 1000}, source.Skills[wID], w.SkillConfig(wID), 10, 20)

	if got := source.Stats.MP; got != 200 {
		t.Fatalf("mp before walking into range = %v, want 200", got)
	}
	if !source.Passive.FireWCastPending {
		t.Fatal("out-of-range pillar should wait until caster walks into range")
	}
	if source.Intent.MoveTarget == nil || source.Intent.MoveTarget.X != 1300 || source.Intent.MoveTarget.Y != 1000 {
		t.Fatalf("move target = %+v, want 1300,1000", source.Intent.MoveTarget)
	}

	source.Position = *source.Intent.MoveTarget
	ReleasePreparedW(w, source, 11, 20)
	w.Tick(26, 20)

	if got, want := target.Combat.LastDamage, 75; got != want {
		t.Fatalf("pillar damage at original target = %d, want %d", got, want)
	}
}

func TestConflagrationInstantlyDamagesAndBurnsTarget(t *testing.T) {
	w, source := testWorld(t)
	target := singleTarget(t, w, source)
	learnE(source, 1)
	source.Stats.MP = 200

	CastE(w, source, protocol.CastInput{SkillID: eID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, source.Skills[eID], w.SkillConfig(eID), 10, 20)

	if got, want := source.Stats.MP, 130.0; got != want {
		t.Fatalf("mp after conflagration = %v, want %v", got, want)
	}
	if got, want := source.Skills[eID].CooldownUntilTick, uint64(250); got != want {
		t.Fatalf("cooldown tick = %d, want %d", got, want)
	}
	if got, want := target.Combat.LastDamage, 70; got != want {
		t.Fatalf("conflagration damage = %d, want %d", got, want)
	}
	if got := target.Passive.FireBurns[source.ID].Stacks; got != 1 {
		t.Fatalf("burn stacks = %d, want 1", got)
	}
}

func TestConflagrationSpreadsBurnFromBurningTarget(t *testing.T) {
	w, source := testWorld(t)
	target := singleTarget(t, w, source)
	near := spawnEnemyHero(t, w, 1600, 1000)
	far := spawnEnemyHero(t, w, 2300, 1000)
	learnE(source, 1)
	applyBurn(w, source, target, 10, 20)
	source.Stats.MP = 200

	CastE(w, source, protocol.CastInput{SkillID: eID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, source.Skills[eID], w.SkillConfig(eID), 11, 20)

	if got := target.Passive.FireBurns[source.ID].Stacks; got != 2 {
		t.Fatalf("primary burn stacks = %d, want 2", got)
	}
	if got := near.Passive.FireBurns[source.ID].Stacks; got != 1 {
		t.Fatalf("near burn stacks = %d, want 1", got)
	}
	if got := far.Passive.FireBurns[source.ID].Stacks; got != 0 {
		t.Fatalf("far burn stacks = %d, want 0", got)
	}
}

func TestPyroclasmReleasesAfterWindupAndHits(t *testing.T) {
	w, source := testWorld(t)
	target := singleTarget(t, w, source)
	learnR(source, 1)
	source.Stats.MP = 200

	CastR(w, source, protocol.CastInput{SkillID: rID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, source.Skills[rID], w.SkillConfig(rID), 10, 20)
	if got, want := source.Stats.MP, 100.0; got != want {
		t.Fatalf("mp after pyroclasm = %v, want %v", got, want)
	}
	if got, want := source.Skills[rID].CooldownUntilTick, uint64(2110); got != want {
		t.Fatalf("cooldown tick = %d, want %d", got, want)
	}

	w.Tick(13, 20)
	if hasPyroclasmProjectile(w) {
		t.Fatal("pyroclasm projectile released before windup finished")
	}
	w.Tick(14, 20)
	if !hasPyroclasmProjectile(w) {
		t.Fatal("pyroclasm projectile was not released after windup")
	}
	tickUntilDamage(t, w, target, 15, 25)

	if got, want := target.Combat.LastDamage, 150; got != want {
		t.Fatalf("pyroclasm damage = %d, want %d", got, want)
	}
	if got := target.Passive.FireBurns[source.ID].Stacks; got != 1 {
		t.Fatalf("burn stacks = %d, want 1", got)
	}
}

func TestPyroclasmBouncesBetweenEnemies(t *testing.T) {
	w, source := testWorld(t)
	first := singleTarget(t, w, source)
	second := spawnEnemyHero(t, w, 1600, 1000)
	second.Stats.MagicDefense = 0
	learnR(source, 1)
	castPyroclasm(w, source, first, 10)

	hitTick := tickUntilDamage(t, w, first, 14, 25)
	tickUntilDamage(t, w, second, hitTick+1, 45)

	if got := first.Passive.FireBurns[source.ID].Stacks; got < 1 {
		t.Fatalf("first burn stacks = %d, want at least 1", got)
	}
	if got := second.Passive.FireBurns[source.ID].Stacks; got < 1 {
		t.Fatalf("second burn stacks = %d, want at least 1", got)
	}
}

func TestPyroclasmPrefersHeroesWhenCurrentTargetBurns(t *testing.T) {
	w, source := testWorld(t)
	first := singleTarget(t, w, source)
	minionID, ok := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamRed, 1400, 1000)
	if !ok {
		t.Fatal("spawn minion failed")
	}
	minion := w.EntityByID(minionID)
	minion.Stats.MagicDefense = 0
	hero := spawnEnemyHero(t, w, 1800, 1000)
	hero.Stats.MagicDefense = 0
	learnR(source, 1)
	applyBurn(w, source, first, 10, 20)
	castPyroclasm(w, source, first, 11)

	for tick := uint64(15); tick <= 22; tick++ {
		w.Tick(tick, 20)
	}
	if minion.Combat.LastDamage != 0 {
		t.Fatalf("pyroclasm hit minion before hero: %d", minion.Combat.LastDamage)
	}
	tickUntilDamage(t, w, hero, 23, 50)
}

func TestPyroclasmCanBounceThroughCasterWithoutDamagingThem(t *testing.T) {
	w, source := testWorld(t)
	first := singleTarget(t, w, source)
	second := spawnEnemyHero(t, w, 700, 1000)
	second.Stats.MagicDefense = 0
	learnR(source, 1)
	source.Stats.HP = source.Stats.MaxHP
	beforeHP := source.Stats.HP
	castPyroclasm(w, source, first, 10)

	hitTick := tickUntilDamage(t, w, first, 14, 25)
	tickUntilDamage(t, w, second, hitTick+1, 50)

	if source.Stats.HP != beforeHP {
		t.Fatalf("caster hp = %v, want %v", source.Stats.HP, beforeHP)
	}
}

func TestPyroclasmCanBounceToCasterWithoutAnotherEnemy(t *testing.T) {
	w, source := testWorld(t)
	first := singleTarget(t, w, source)
	learnR(source, 1)
	beforeHP := source.Stats.HP
	castPyroclasm(w, source, first, 10)

	tickUntilDamage(t, w, first, 14, 25)

	if source.Stats.HP != beforeHP {
		t.Fatalf("caster hp = %v, want %v", source.Stats.HP, beforeHP)
	}
	if !pyroclasmMovingToward(w, source.Position) {
		t.Fatal("pyroclasm should bounce to caster when no other enemy is in range")
	}
}

func TestBlazeStacksAndTicks(t *testing.T) {
	w, source := testWorld(t)
	target := spawnEnemyHero(t, w, 1000, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.MagicDefense = 0

	for i := 0; i < 3; i++ {
		applySkillHit(w, source, target, 10)
	}

	burn := target.Passive.FireBurns[source.ID]
	if burn.Stacks != 3 {
		t.Fatalf("burn stacks = %d, want 3", burn.Stacks)
	}
	if burn.ExplosionAtTick != 50 {
		t.Fatalf("explosion tick = %d, want 50", burn.ExplosionAtTick)
	}

	w.Tick(30, 20)

	if target.Stats.HP != 937 {
		t.Fatalf("target hp = %v, want 937", target.Stats.HP)
	}
}

func TestBlazeDamageTriggersLiandrys(t *testing.T) {
	w, source := testWorld(t)
	source.Gold = 3100
	w.ApplyInput("fire", protocol.PlayerInput{BuyEquipment: &protocol.BuyEquipmentInput{EquipmentID: "liandrys_anguish"}}, 1, nil, 20)
	target := spawnEnemyHero(t, w, 1000, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.MagicDefense = 0

	applyBurn(w, source, target, 10, 20)
	w.Tick(30, 20)
	w.Tick(50, 20)

	if target.Stats.HP != 920 {
		t.Fatalf("target hp = %v, want 920", target.Stats.HP)
	}
}

func TestBlazeExplodesAtThreeStacks(t *testing.T) {
	w, source := testWorld(t)
	target := spawnEnemyHero(t, w, 1000, 1000)
	target.Stats.MagicDefense = 0
	nearID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1200, 1000)
	if !ok {
		t.Fatal("spawn nearby enemy failed")
	}
	near := w.EntityByID(nearID)
	near.Stats.MagicDefense = 0

	for i := 0; i < 3; i++ {
		applySkillHit(w, source, target, 10)
	}
	w.Tick(50, 20)

	if near.Combat.LastDamage <= 0 {
		t.Fatal("nearby enemy should take blaze explosion damage")
	}
	if _, ok := target.Passive.FireBurns[source.ID]; ok {
		t.Fatal("burn should be consumed by explosion")
	}
}

func TestBlazeKillStartsManaRestore(t *testing.T) {
	w, source := testWorld(t)
	source.Stats.MP = 0
	source.Stats.MaxMP = 1000
	source.Stats.MPRegen5 = 0
	source.Position = world.Vector2{X: 3000, Y: 3000}
	target := spawnEnemyHero(t, w, 1000, 1000)
	target.Stats.HP = 5
	target.Stats.MaxHP = 100
	target.Stats.MagicDefense = 0

	for i := 0; i < 3; i++ {
		applySkillHit(w, source, target, 10)
	}
	w.Tick(30, 20)
	w.Tick(130, 20)

	if source.Stats.MP != 30 {
		t.Fatalf("mp after first restore tick = %v, want 30", source.Stats.MP)
	}
}

func applySkillHit(w *world.World, source *world.Entity, target *world.Entity, tick uint64) {
	target.Combat.LastHitTick = tick
	w.ApplyMagicDamage(source, target, 1, 20)
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

func castSear(w *world.World, source *world.Entity, tick uint64) {
	source.Stats.MP = 100
	CastQ(w, source, protocol.CastInput{SkillID: qID, TargetX: source.Position.X + 600, TargetY: source.Position.Y}, source.Skills[qID], w.SkillConfig(qID), tick, 20)
}

func castPillar(w *world.World, source *world.Entity, tick uint64) {
	source.Stats.MP = 200
	CastW(w, source, protocol.CastInput{SkillID: wID, TargetX: source.Position.X + 200, TargetY: source.Position.Y}, source.Skills[wID], w.SkillConfig(wID), tick, 20)
}

func castPyroclasm(w *world.World, source *world.Entity, target *world.Entity, tick uint64) {
	source.Stats.MP = 200
	CastR(w, source, protocol.CastInput{SkillID: rID, TargetID: target.ID, TargetX: target.Position.X, TargetY: target.Position.Y}, source.Skills[rID], w.SkillConfig(rID), tick, 20)
}

func hasSearProjectile(w *world.World) bool {
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "fire_mage_q" {
			return true
		}
	}
	return false
}

func hasPyroclasmProjectile(w *world.World) bool {
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "fire_mage_r" {
			return true
		}
	}
	return false
}

func pyroclasmMovingToward(w *world.World, target world.Vector2) bool {
	for _, effect := range w.SkillEffects() {
		if effect.Kind != "fire_mage_r" {
			continue
		}
		dx, dy := normalize(target.X-effect.Start.X, target.Y-effect.Start.Y)
		if dx*effect.Dir.X+dy*effect.Dir.Y > 0.99 {
			return true
		}
	}
	return false
}

func tickUntilDamage(t *testing.T, w *world.World, target *world.Entity, from uint64, to uint64) uint64 {
	t.Helper()
	for tick := from; tick <= to; tick++ {
		w.Tick(tick, 20)
		if target.Combat.LastDamage > 0 {
			return tick
		}
	}
	t.Fatalf("%s did not take damage by tick %d", target.ID, to)
	return to
}

func searTarget(t *testing.T, w *world.World, source *world.Entity) *world.Entity {
	t.Helper()
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.AbilityPower = 0
	target := spawnEnemyHero(t, w, 1300, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.MagicDefense = 0
	return target
}

func pillarTarget(t *testing.T, w *world.World, source *world.Entity) *world.Entity {
	t.Helper()
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.AbilityPower = 0
	target := spawnEnemyHero(t, w, 1200, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.MagicDefense = 0
	return target
}

func singleTarget(t *testing.T, w *world.World, source *world.Entity) *world.Entity {
	t.Helper()
	source.Position = world.Vector2{X: 1000, Y: 1000}
	source.Stats.AbilityPower = 0
	target := spawnEnemyHero(t, w, 1200, 1000)
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	target.Stats.MagicDefense = 0
	return target
}

func spawnEnemyHero(t *testing.T, w *world.World, x float64, y float64) *world.Entity {
	t.Helper()
	id, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, x, y)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}
	target := w.EntityByID(id)
	if target == nil {
		t.Fatal("spawned enemy hero not found")
	}
	return target
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
	equipment, err := config.LoadEquipment("../../../../configs/equipment")
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(heroes, skills, nil, nil, equipment)
	hero, ok := heroes.Get(heroID)
	if !ok {
		t.Fatal("fire mage hero not found")
	}
	w.SpawnHero("fire", hero, world.TeamBlue)
	source := w.EntityByID("player:fire")
	if source == nil {
		t.Fatal("fire mage entity not found")
	}
	return w, source
}
