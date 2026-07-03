package world

import (
	"l-battle/internal/config"
	"math"
	"testing"
)

func TestSwordConfiguredStatsAtLevel18(t *testing.T) {
	heroes, err := config.LoadHeroes("../../configs/heroes.json")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get(swordHeroID)
	if !ok {
		t.Fatal("sword hero not found")
	}
	stats := heroStatsAtLevel(hero, MaxHeroLevel)

	if hero.Resource != "sword_intent" {
		t.Fatalf("sword resource = %s, want sword_intent", hero.Resource)
	}
	if stats.MaxHP != 1969 {
		t.Fatalf("sword level 18 hp = %d, want 1969", stats.MaxHP)
	}
	if stats.Attack != 111 {
		t.Fatalf("sword level 18 attack = %f, want 111", stats.Attack)
	}
	if math.Abs(stats.AttackSpeed-1.111715) > 0.000001 {
		t.Fatalf("sword level 18 attack speed = %f, want 1.111715", stats.AttackSpeed)
	}
	if stats.PhysicalDefense != 87.8 {
		t.Fatalf("sword level 18 armor = %f, want 87.8", stats.PhysicalDefense)
	}
	if stats.MagicDefense != 53.25 {
		t.Fatalf("sword level 18 magic resist = %f, want 53.25", stats.MagicDefense)
	}
}

func TestInfinityEdgeKeepsBaseCritMultiplierForSword(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Base.CritChance = 0
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 3500

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("infinity_edge"), 1, nil, 20)

	if got := w.critDamageMultiplier(player); got != 2.5 {
		t.Fatalf("crit damage multiplier = %f, want 2.5", got)
	}
}

func TestSwordCritOverflowConvertsToBonusAttack(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get(swordHeroID)
	if !ok {
		t.Fatal("missing sword hero")
	}
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 5600
	baseAttack := player.Stats.Attack
	baseBonusAttack := player.Stats.BonusAttack

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("phantom_dancer"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("phantom_dancer"), 2, nil, 20)

	if player.Stats.CritChance != 0.6 {
		t.Fatalf("raw crit chance = %f, want 0.6", player.Stats.CritChance)
	}
	if got := w.DisplayCritChance(player); got != 1 {
		t.Fatalf("display crit chance = %f, want 1", got)
	}
	if math.Abs(player.Stats.Attack-(baseAttack+10)) > 0.000001 {
		t.Fatalf("attack = %f, want %f", player.Stats.Attack, baseAttack+10)
	}
	if math.Abs(player.Stats.BonusAttack-(baseBonusAttack+10)) > 0.000001 {
		t.Fatalf("bonus attack = %f, want %f", player.Stats.BonusAttack, baseBonusAttack+10)
	}
}

func TestSwordPassiveChargesWhileMoving(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	w.ApplyInput("p1", protocolPlayerInputMove(player.Position.X+100, player.Position.Y), 1, nil, 20)
	w.Tick(2, 20)

	if player.Passive.SwordIntent <= 0 {
		t.Fatalf("sword intent did not charge while moving")
	}
	if player.Passive.MaxSwordIntent != 100 {
		t.Fatalf("max sword intent = %f, want 100", player.Passive.MaxSwordIntent)
	}
}

func TestSwordPassiveIntentChargeByLevelConfig(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	player.Level = 1
	w.chargeSwordIntent(player, 59)
	if math.Abs(player.Passive.SwordIntent-1) > 0.001 {
		t.Fatalf("level 1 sword intent = %f, want 1", player.Passive.SwordIntent)
	}

	player.Level = 7
	player.Passive.SwordIntent = 0
	w.chargeSwordIntent(player, 52.5)
	if math.Abs(player.Passive.SwordIntent-1) > 0.001 {
		t.Fatalf("level 7 sword intent = %f, want 1", player.Passive.SwordIntent)
	}

	player.Level = 13
	player.Passive.SwordIntent = 0
	w.chargeSwordIntent(player, 46)
	if math.Abs(player.Passive.SwordIntent-1) > 0.001 {
		t.Fatalf("level 13 sword intent = %f, want 1", player.Passive.SwordIntent)
	}
}

func TestSwordPassiveHeroDamageTriggersShield(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	source := w.entities["enemy:hero-1"]

	w.applyDamage(source, player, 150, 20)

	if player.Passive.SwordIntent != 0 {
		t.Fatalf("sword intent = %f, want 0 after shield trigger", player.Passive.SwordIntent)
	}
	if player.Passive.MaxShield != 125 {
		t.Fatalf("max shield = %d, want 125", player.Passive.MaxShield)
	}
	if player.Passive.Shield != 0 {
		t.Fatalf("shield = %d, want 0 after absorbing 150 damage", player.Passive.Shield)
	}
	if player.Stats.HP != player.Stats.MaxHP-25 {
		t.Fatalf("hp = %d, want %d", player.Stats.HP, player.Stats.MaxHP-25)
	}
}

func TestSwordPassiveShieldValueByLevelConfig(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	tests := []struct {
		level int
		want  int
	}{
		{level: 1, want: 125},
		{level: 6, want: 160},
		{level: 12, want: 275},
		{level: 18, want: 600},
	}
	for _, tt := range tests {
		player.Level = tt.level
		if got := w.swordShieldValue(player); got != tt.want {
			t.Fatalf("level %d shield = %d, want %d", tt.level, got, tt.want)
		}
	}
}

func TestSwordPassiveShieldExpiresAfterConfiguredDuration(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	source := w.entities["enemy:hero-1"]
	player.Combat.LastHitTick = 10

	w.applyDamage(source, player, 25, 20)
	if player.Passive.Shield != 100 {
		t.Fatalf("shield after absorb = %d, want 100", player.Passive.Shield)
	}
	if player.Passive.ShieldExpireTick != 30 {
		t.Fatalf("shield expire tick = %d, want 30", player.Passive.ShieldExpireTick)
	}

	w.Tick(29, 20)
	if player.Passive.Shield != 100 {
		t.Fatalf("shield before expire = %d, want 100", player.Passive.Shield)
	}
	w.Tick(30, 20)
	if player.Passive.Shield != 0 || player.Passive.MaxShield != 0 {
		t.Fatalf("shield after expire = %d/%d, want 0/0", player.Passive.Shield, player.Passive.MaxShield)
	}
}

func TestSwordPassiveMinionDamageDoesNotTriggerShield(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	source := w.entities["minion:red-melee-1"]

	w.applyDamage(source, player, 50, 20)

	if player.Passive.SwordIntent != player.Passive.MaxSwordIntent {
		t.Fatalf("sword intent = %f, want unchanged full intent", player.Passive.SwordIntent)
	}
	if player.Passive.Shield != 0 {
		t.Fatalf("shield = %d, want 0", player.Passive.Shield)
	}
	if player.Stats.HP != player.Stats.MaxHP-50 {
		t.Fatalf("hp = %d, want %d", player.Stats.HP, player.Stats.MaxHP-50)
	}
}

func TestSwordPassiveDoublesCritChance(t *testing.T) {
	attacker := &Entity{
		HeroID: swordHeroID,
		Stats:  Stats{CritChance: 0.4},
	}

	w := testWorld(t)
	if got := w.critChance(attacker); got != 0.8 {
		t.Fatalf("crit chance = %f, want 0.8", got)
	}
}

func TestSwordPassiveZeroCritChanceStaysZero(t *testing.T) {
	attacker := &Entity{
		HeroID: swordHeroID,
		Stats:  Stats{CritChance: 0},
	}

	w := testWorld(t)
	if got := w.critChance(attacker); got != 0 {
		t.Fatalf("crit chance = %f, want 0", got)
	}
}

func TestSwordPassiveCritFinalDamageIsReducedBy10Percent(t *testing.T) {
	attacker := &Entity{
		ID:     "player:p1",
		HeroID: swordHeroID,
		Stats: Stats{
			Attack:     100,
			CritChance: 1,
		},
	}
	target := &Entity{
		ID:    "dummy:target",
		Stats: Stats{PhysicalDefense: 10},
	}

	w := testWorld(t)
	damage := w.attackDamage(attacker, target, 1)

	if damage != 164 {
		t.Fatalf("damage = %d, want 164", damage)
	}
}

func TestSwordPassiveFinalDamageReductionAppliesAfterInfinityEdge(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.CritChance = 1
	player.Gold = 3500
	target := &Entity{
		ID:    "dummy:target",
		Stats: Stats{PhysicalDefense: 0},
	}

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("infinity_edge"), 1, nil, 20)
	player.Stats.Attack = 100
	damage := w.attackDamage(player, target, 1)

	if damage != 225 {
		t.Fatalf("damage = %d, want 225", damage)
	}
}

func TestTrueDamageIsAbsorbedBySwordShield(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Passive = "sword_edge"
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Passive.SwordIntent = player.Passive.MaxSwordIntent
	player.Combat.LastHitTick = 10
	source := w.entities["enemy:hero-1"]

	w.applyTrueDamage(source, player, 100, 20)

	if player.Stats.HP != player.Stats.MaxHP {
		t.Fatalf("hp after shielded true damage = %d, want %d", player.Stats.HP, player.Stats.MaxHP)
	}
	if player.Passive.Shield != 25 {
		t.Fatalf("shield after true damage = %d, want 25", player.Passive.Shield)
	}
}

func TestSwordQDamagesTargetAndAddsStack(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	target := w.entities["dummy:training-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 1, nil, 20)

	if target.Combat.LastDamage != 0 {
		t.Fatalf("sword q damage before windup release = %d, want 0", target.Combat.LastDamage)
	}
	tickSwordQRelease(t, w, player, 20)

	state := player.Skills[swordQSkillID]
	if target.Combat.LastDamage <= 0 {
		t.Fatal("sword q should damage target")
	}
	if state.Stacks != 1 {
		t.Fatalf("q stacks = %d, want 1", state.Stacks)
	}
	if state.CooldownUntilTick <= 1 {
		t.Fatalf("q cooldown was not set")
	}
}

func TestSwordQCooldownUsesSkillLevelAndAttackSpeedPercent(t *testing.T) {
	w := testWorld(t)
	skill := w.skillConfig(swordQSkillID)

	if got := swordQCooldownTicksByBonus(0, skill, 1, 20); got != 120 {
		t.Fatalf("level 1 q cooldown ticks = %d, want 120", got)
	}
	if got := swordQCooldownTicksByBonus(1, skill, MaxBasicSkillLevel, 20); got != 32 {
		t.Fatalf("level 5 q cooldown ticks = %d, want 32", got)
	}
	if got := swordQCooldownTicksByBonus(10, skill, MaxBasicSkillLevel, 20); got != 27 {
		t.Fatalf("level 5 q min cooldown ticks = %d, want 27", got)
	}
}

func TestSwordQCooldownUsesAttackSpeedBonusNotPanelAttackSpeed(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Base.AttackSpeed = 0.697
	hero.Growth.AttackSpeed = 0
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Stats.AttackSpeed = 0.697
	player.Stats.AttackSpeedBonus = 0

	if got := w.swordQCooldownTicks(player, w.skillConfig(swordQSkillID), 1, 20); got != 120 {
		t.Fatalf("base attack speed q cooldown ticks = %d, want 120", got)
	}
	player.Stats.AttackSpeed = 0.697 * 2
	player.Stats.AttackSpeedBonus = 1
	if got := w.swordQCooldownTicks(player, w.skillConfig(swordQSkillID), 1, 20); got != 48 {
		t.Fatalf("100%% bonus attack speed q cooldown ticks = %d, want 48", got)
	}
}

func TestSwordQWindupUsesAttackSpeedBonus(t *testing.T) {
	w := testWorld(t)
	skill := w.skillConfig(swordQSkillID)
	entity := &Entity{Stats: Stats{AttackSpeedBonus: 0}}

	if got := swordQWindupSeconds(entity, skill); math.Abs(got-0.328) > 0.000001 {
		t.Fatalf("base windup = %f, want 0.328", got)
	}
	entity.Stats.AttackSpeedBonus = 0.5
	want := 0.328 / 1.5
	if got := swordQWindupSeconds(entity, skill); math.Abs(got-want) > 0.000001 {
		t.Fatalf("50%% bonus attack speed windup = %f, want %f", got, want)
	}
	entity.Stats.AttackSpeedBonus = 10
	if got := swordQWindupSeconds(entity, skill); math.Abs(got-0.09) > 0.000001 {
		t.Fatalf("capped windup = %f, want 0.09", got)
	}
}

func TestSwordQIgnoresAbilityHaste(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	player.Stats.AbilityHaste = 100

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 10, w.skills, 20)

	if got := player.Skills[swordQSkillID].CooldownUntilTick; got != 0 {
		t.Fatalf("sword q cooldown during windup = %d, want 0", got)
	}
	releaseTick := tickSwordQRelease(t, w, player, 20)
	want := releaseTick + w.swordQCooldownTicks(player, w.skillConfig(swordQSkillID), 1, 20)
	if got := player.Skills[swordQSkillID].CooldownUntilTick; got != want {
		t.Fatalf("sword q cooldown with ability haste = %d, want %d", got, want)
	}
}

func TestSwordQDamagesAllEnemiesInMouseDirection(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	first := w.entities["dummy:training-1"]
	second := w.entities["dummy:training-2"]
	outside := w.entities["enemy:hero-1"]
	first.Position = Vector2{X: player.Position.X + 180, Y: player.Position.Y}
	second.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y + 20}
	outside.Position = Vector2{X: player.Position.X, Y: player.Position.Y + 300}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, first.Position.X, first.Position.Y), 1, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if first.Combat.LastDamage <= 0 {
		t.Fatal("first target should take sword q damage")
	}
	if second.Combat.LastDamage <= 0 {
		t.Fatal("second target should take sword q damage")
	}
	if outside.Combat.LastDamage != 0 {
		t.Fatalf("outside target damage = %d, want 0", outside.Combat.LastDamage)
	}
}

func TestSwordQThirdHitBecomesWhirlwindAndKnocksUp(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	for id, entity := range w.entities {
		if id != player.ID && id != target.ID && entity.Team != player.Team {
			delete(w.entities, id)
		}
	}
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 100, nil, 20)

	if target.Combat.LastDamage != 0 {
		t.Fatal("whirlwind q should not damage target before projectile reaches it")
	}
	hitTick := uint64(0)
	for tick := uint64(101); tick <= 140; tick++ {
		w.Tick(tick, 20)
		if target.Combat.LastDamage > 0 {
			hitTick = tick
			break
		}
	}
	if target.Combat.LastDamage <= 0 {
		t.Fatal("whirlwind q should damage target after projectile reaches it")
	}
	if target.Control.AirborneUntilTick != hitTick+20 {
		t.Fatalf("airborne until = %d, want %d", target.Control.AirborneUntilTick, hitTick+20)
	}
	state := player.Skills[swordQSkillID]
	if state.Stacks != 0 {
		t.Fatalf("q stacks = %d, want reset to 0", state.Stacks)
	}
}

func TestSwordWhirlwindQProjectileHitsOnlyAfterCollision(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	nearPath := w.entities["dummy:training-1"]
	nearWhirlwindEdge := w.entities["dummy:training-2"]
	outsideWhirlwind := w.entities["enemy:hero-1"]
	nearPath.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y + 20}
	nearWhirlwindEdge.Position = Vector2{X: player.Position.X + 600, Y: player.Position.Y + 20}
	outsideWhirlwind.Position = Vector2{X: player.Position.X + 600, Y: player.Position.Y + 180}
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+900, player.Position.Y), 100, nil, 20)

	if nearPath.Combat.LastDamage != 0 || nearWhirlwindEdge.Combat.LastDamage != 0 {
		t.Fatal("whirlwind q should not damage targets on cast")
	}
	for tick := uint64(101); tick <= 117; tick++ {
		w.Tick(tick, 20)
		if nearPath.Combat.LastDamage > 0 {
			break
		}
	}
	if nearPath.Combat.LastDamage <= 0 {
		t.Fatal("target near whirlwind path should take damage")
	}
	if nearWhirlwindEdge.Combat.LastDamage != 0 {
		t.Fatal("farther target should not be hit before projectile reaches it")
	}
	for tick := uint64(118); tick <= 126; tick++ {
		w.Tick(tick, 20)
		if nearWhirlwindEdge.Combat.LastDamage > 0 {
			break
		}
	}
	if nearWhirlwindEdge.Combat.LastDamage <= 0 {
		t.Fatal("target inside whirlwind radius should take damage")
	}
	if outsideWhirlwind.Combat.LastDamage != 0 {
		t.Fatalf("outside target damage = %d, want 0", outsideWhirlwind.Combat.LastDamage)
	}
}

func TestSwordWhirlwindQCanComboIntoLastBreathAfterHit(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	for id, entity := range w.entities {
		if id != player.ID && id != target.ID && entity.Team != player.Team {
			delete(w.entities, id)
		}
	}
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 100, nil, 20)
	hitTick := uint64(0)
	for tick := uint64(101); tick <= 130; tick++ {
		w.Tick(tick, 20)
		if target.Control.AirborneUntilTick > tick {
			hitTick = tick
			break
		}
	}
	if hitTick == 0 {
		t.Fatalf("target airborne until = %d, want airborne after whirlwind q hit", target.Control.AirborneUntilTick)
	}
	rTick := hitTick + 1
	if target.Control.AirborneUntilTick <= rTick {
		t.Fatalf("target airborne until = %d, want after r cast tick", target.Control.AirborneUntilTick)
	}

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), rTick, nil, 20)

	if player.Skills[swordRSkillID].CooldownUntilTick == 0 {
		t.Fatal("r should cast on target knocked up by whirlwind q")
	}
	if target.Combat.LastDamage <= 0 {
		t.Fatal("r should damage airborne target")
	}
}

func TestSwordWCreatesWindWallAndExpires(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.W = swordWSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordWSkillID, 1)

	w.ApplyInput("p1", protocolPlayerInputCast(swordWSkillID, player.Position.X+100, player.Position.Y), 10, nil, 20)

	walls := w.WindWalls()
	if len(walls) != 1 {
		t.Fatalf("wind wall count = %d, want 1", len(walls))
	}
	if walls[0].Width != 300 {
		t.Fatalf("wind wall width = %f, want 300", walls[0].Width)
	}
	if walls[0].ExpiresAt != 90 {
		t.Fatalf("wind wall expires at = %d, want 90", walls[0].ExpiresAt)
	}
	state := player.Skills[swordWSkillID]
	if state.CooldownUntilTick != 530 {
		t.Fatalf("wind wall cooldown until = %d, want 530", state.CooldownUntilTick)
	}

	w.Tick(90, 20)

	if got := len(w.WindWalls()); got != 0 {
		t.Fatalf("wind wall count after expiry = %d, want 0", got)
	}
}

func TestWindWallBlocksEnemyProjectileOnly(t *testing.T) {
	w := testWorld(t)
	w.windWalls["wall"] = WindWall{
		ID:        "wall",
		Team:      TeamBlue,
		Center:    Vector2{X: 100, Y: 100},
		Dir:       Vector2{X: 0, Y: 1},
		Width:     300,
		ExpiresAt: 100,
	}

	if !w.BlocksProjectile(TeamRed, Vector2{X: 0, Y: 100}, Vector2{X: 200, Y: 100}) {
		t.Fatal("enemy projectile should be blocked")
	}
	if w.BlocksProjectile(TeamBlue, Vector2{X: 0, Y: 100}, Vector2{X: 200, Y: 100}) {
		t.Fatal("same-team projectile should not be blocked")
	}
}

func TestWindWallRemovesEnemyProjectile(t *testing.T) {
	w := testWorld(t)
	w.windWalls["wall"] = WindWall{
		ID:        "wall",
		Team:      TeamBlue,
		Center:    Vector2{X: 100, Y: 100},
		Dir:       Vector2{X: 0, Y: 1},
		Width:     300,
		ExpiresAt: 100,
	}
	w.projectiles["p"] = &Projectile{
		ID:           "p",
		Kind:         "mage_q",
		Team:         TeamRed,
		SkillID:      mageQSkillID,
		Position:     Vector2{X: 50, Y: 100},
		Dir:          Vector2{X: 1, Y: 0},
		SpeedPerTick: 100,
		Range:        500,
		ExpiresAt:    100,
		HitIDs:       map[string]bool{},
	}

	w.tickProjectiles(1, 20)

	if _, ok := w.projectiles["p"]; ok {
		t.Fatal("enemy projectile crossing wind wall should be removed")
	}
}

func TestSwordEDashesThroughTargetAndAppliesPerTargetCooldown(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	hero.Base.BonusAttack = 40
	hero.Base.AbilityPower = 50
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}
	startPosition := player.Position
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)

	if target.Stats.HP >= startHP {
		t.Fatal("sword e should damage target")
	}
	if player.Position != startPosition {
		t.Fatalf("player position = %+v, want unchanged at cast tick %+v", player.Position, startPosition)
	}
	w.Tick(13, 20)
	if player.Position.X <= startPosition.X || player.Position.X >= target.Position.X {
		t.Fatalf("player x = %f, want moving toward target from %f", player.Position.X, startPosition.X)
	}
	w.Tick(17, 20)
	if player.Position.X <= target.Position.X {
		t.Fatalf("player x = %f, should finish dash through target x=%f", player.Position.X, target.Position.X)
	}
	if player.Passive.SwordIntent <= 0 {
		t.Fatal("sword e movement should charge sword intent")
	}
	if player.Sword.SweepingBladeStacks != 1 {
		t.Fatalf("e stacks = %d, want 1", player.Sword.SweepingBladeStacks)
	}
	afterFirstHP := target.Stats.HP
	player.Skills[swordESkillID] = SkillState{SkillID: swordESkillID, Level: 1}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 20, nil, 20)

	if target.Stats.HP != afterFirstHP {
		t.Fatalf("target hp = %d, want unchanged %d while per-target cooldown active", target.Stats.HP, afterFirstHP)
	}
}

func TestSwordEPicksUnitNearestToCursorPoint(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	nearPlayer := w.entities["enemy:hero-1"]
	nearCursor := w.entities["enemy:blue-hero-1"]
	nearPlayer.Team = TeamRed
	nearCursor.Team = TeamRed
	nearPlayer.Position = Vector2{X: player.Position.X + 160, Y: player.Position.Y}
	nearCursor.Position = Vector2{X: player.Position.X + 260, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, nearCursor.Position.X, nearCursor.Position.Y), 10, nil, 20)

	if nearCursor.Combat.LastDamage <= 0 {
		t.Fatal("sword e should hit unit nearest to cursor point")
	}
	if nearPlayer.Combat.LastDamage != 0 {
		t.Fatalf("nearer-to-player unit damage = %d, want 0", nearPlayer.Combat.LastDamage)
	}
}

func TestSwordEStoresPerTargetCooldown(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)

	if player.Sword.SweepingBladeTargetUntil[target.ID] != 210 {
		t.Fatalf("target cooldown until = %d, want 210", player.Sword.SweepingBladeTargetUntil[target.ID])
	}
}

func TestSwordEIgnoresAttackInputDuringDash(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.E = swordESkillID
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 11, nil, 20)

	if player.Intent.AttackTargetID != "" {
		t.Fatalf("attack target during e dash = %q, want empty", player.Intent.AttackTargetID)
	}
}

func TestSwordEQMakesCircularQ(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	sideTarget := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	sideTarget.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	sideTarget.Position = Vector2{X: target.Position.X, Y: target.Position.Y + 120}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.Tick(14, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 14, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if sideTarget.Combat.LastDamage <= 0 {
		t.Fatal("q during e dash should become circular aoe and hit nearby target")
	}
}

func TestSwordEQDamageUsesReleasePositionAfterDashMovement(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	dashTarget := w.entities["enemy:hero-1"]
	oldPositionTarget := w.entities["enemy:blue-hero-1"]
	newPositionTarget := &Entity{
		ID:       "enemy:eq-release-position",
		Kind:     EntityKindEnemyHero,
		Team:     TeamRed,
		HeroID:   warriorHeroID,
		Position: Vector2{X: player.Position.X + 680, Y: player.Position.Y},
		Radius:   30,
		Stats:    Stats{HP: 1000, MaxHP: 1000, PhysicalDefense: 0, MagicDefense: 0},
		Skills:   make(map[string]SkillState),
	}
	w.entities[newPositionTarget.ID] = newPositionTarget
	dashTarget.Team = TeamRed
	oldPositionTarget.Team = TeamRed
	dashTarget.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	oldPositionTarget.Position = Vector2{X: player.Position.X, Y: player.Position.Y + 450}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, dashTarget.Position.X, dashTarget.Position.Y), 10, nil, 20)
	w.Tick(14, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 14, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if oldPositionTarget.Combat.LastDamage != 0 {
		t.Fatalf("old position target damage = %d, want 0", oldPositionTarget.Combat.LastDamage)
	}
	if newPositionTarget.Combat.LastDamage <= 0 {
		t.Fatal("new position target should be hit by EQ at release position")
	}
}

func TestSwordQBeforeEQWindowDoesNotBecomeCircular(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	sideTarget := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	sideTarget.Team = TeamRed
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	sideTarget.Position = Vector2{X: target.Position.X, Y: target.Position.Y + 120}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 11, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if sideTarget.Combat.LastDamage != 0 {
		t.Fatalf("side target damage = %d, want 0 before eq window", sideTarget.Combat.LastDamage)
	}
}

func TestSwordEQWithWhirlwindStacksKnocksUpAndClearsStacks(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.E = swordESkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordESkillID, 1)
	target := w.entities["enemy:hero-1"]
	sideTarget := w.entities["enemy:blue-hero-1"]
	target.Team = TeamRed
	sideTarget.Team = TeamRed
	sideTarget.HeroID = warriorHeroID
	sideTarget.Warrior.CourageFrontUntilTick = 30
	sideTarget.Warrior.CourageFrontTenacity = 0.6
	sideTarget.Control.TenacityUntilTick = 30
	target.Position = Vector2{X: player.Position.X + 150, Y: player.Position.Y}
	sideTarget.Position = Vector2{X: target.Position.X, Y: target.Position.Y + 120}

	w.ApplyInput("p1", protocolPlayerInputCast(swordESkillID, target.Position.X, target.Position.Y), 10, nil, 20)
	w.Tick(14, 20)
	player.Skills[swordQSkillID] = SkillState{SkillID: swordQSkillID, Level: 1, Stacks: 2, StacksExpireTick: 200}
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, player.Position.X+1, player.Position.Y), 14, nil, 20)
	tickSwordQRelease(t, w, player, 20)

	if sideTarget.Control.AirborneUntilTick != 41 {
		t.Fatalf("side target airborne until = %d, want 41", sideTarget.Control.AirborneUntilTick)
	}
	qState := player.Skills[swordQSkillID]
	if qState.Stacks != 0 {
		t.Fatalf("q stacks = %d, want 0", qState.Stacks)
	}
	if qState.StacksExpireTick != 0 {
		t.Fatalf("q stacks expire tick = %d, want 0", qState.StacksExpireTick)
	}
}

func TestSwordRRequiresAirborneHeroAndAppliesLastBreathState(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	hero.Base.BonusAttack = 60
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	target.Control.AirborneUntilTick = 120
	startHP := target.Stats.HP
	player.Passive.SwordIntent = 12

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), 100, nil, 20)

	if target.Stats.HP >= startHP {
		t.Fatal("sword r should damage airborne enemy hero")
	}
	if target.Control.AirborneUntilTick != 140 {
		t.Fatalf("airborne until = %d, want extended to 140", target.Control.AirborneUntilTick)
	}
	if player.Passive.Shield != w.swordShieldValue(player) {
		t.Fatalf("shield = %d, want %d", player.Passive.Shield, w.swordShieldValue(player))
	}
	if player.Passive.SwordIntent != player.Passive.MaxSwordIntent {
		t.Fatalf("sword intent = %f, want %f", player.Passive.SwordIntent, player.Passive.MaxSwordIntent)
	}
	qState := player.Skills[swordQSkillID]
	if qState.Stacks != 0 {
		t.Fatalf("q stacks = %d, want 0", qState.Stacks)
	}
	if qState.StacksExpireTick != 0 {
		t.Fatalf("q stacks expire tick = %d, want 0", qState.StacksExpireTick)
	}
	if player.Sword.LastBreathUntilTick != 400 {
		t.Fatalf("last breath until = %d, want 400", player.Sword.LastBreathUntilTick)
	}
	if player.Skills[swordRSkillID].CooldownUntilTick != 1700 {
		t.Fatalf("r cooldown until = %d, want 1700", player.Skills[swordRSkillID].CooldownUntilTick)
	}
}

func TestSwordRAutoTargetsAirborneHeroInRange(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 500, Y: player.Position.Y}
	target.Control.AirborneUntilTick = 130

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, player.Position.X+1200, player.Position.Y+900), 110, nil, 20)

	if player.Skills[swordRSkillID].CooldownUntilTick == 0 {
		t.Fatal("r should auto target airborne enemy hero in range")
	}
}

func TestSwordRLocksSelfActionsForOneSecond(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordQSkillID, 1)
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	target.Control.AirborneUntilTick = 130

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), 100, nil, 20)
	if player.Control.ActionLockedUntilTick != 120 {
		t.Fatalf("action locked until = %d, want 120", player.Control.ActionLockedUntilTick)
	}
	start := player.Position
	player.Intent.AttackTargetID = target.ID
	nextAttack := player.Combat.NextAttackTick

	w.ApplyInput("p1", protocolPlayerInputMove(start.X+500, start.Y), 101, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 101, nil, 20)
	w.Tick(101, 20)

	if player.Position != start {
		t.Fatalf("player moved while action locked: got %+v want %+v", player.Position, start)
	}
	if player.Combat.NextAttackTick != nextAttack {
		t.Fatalf("next attack tick changed while action locked: got %d want %d", player.Combat.NextAttackTick, nextAttack)
	}
	if player.Skills[swordQSkillID].CooldownUntilTick != 0 {
		t.Fatalf("q cooldown = %d, want 0 while action locked", player.Skills[swordQSkillID].CooldownUntilTick)
	}
}

func TestSwordRDoesNotCastOnNonAirborneHero(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	learnSkill(player, swordRSkillID, 1)
	target := w.entities["enemy:hero-1"]
	target.Position = Vector2{X: player.Position.X + 300, Y: player.Position.Y}
	startHP := target.Stats.HP

	w.ApplyInput("p1", protocolPlayerInputCast(swordRSkillID, target.Position.X, target.Position.Y), 100, nil, 20)

	if target.Stats.HP != startHP {
		t.Fatalf("target hp = %d, want unchanged %d", target.Stats.HP, startHP)
	}
	if player.Skills[swordRSkillID].CooldownUntilTick != 0 {
		t.Fatal("r should not enter cooldown without valid airborne target")
	}
}

func TestLastBreathDoesNotPenetrateArmorYet(t *testing.T) {
	attacker := &Entity{
		ID:     "player:p1",
		HeroID: swordHeroID,
		Stats:  Stats{Attack: 200},
		Sword:  SwordState{LastBreathUntilTick: 200},
	}
	target := &Entity{
		ID: "target",
		Stats: Stats{
			PhysicalDefense:      80,
			BonusPhysicalDefense: 40,
		},
	}

	w := testWorld(t)
	damage := w.attackDamage(attacker, target, 100)

	if damage != 111 {
		t.Fatalf("damage = %d, want 111", damage)
	}
}
