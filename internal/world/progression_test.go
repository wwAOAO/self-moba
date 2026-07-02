package world

import "testing"

func TestHeroStatsAtMaxLevelUseGrowth(t *testing.T) {
	hero := testHeroConfig()
	hero.Growth.HP = 10
	hero.Growth.MP = 2
	hero.Growth.Attack = 3
	hero.Growth.PhysicalDefense = 4
	hero.Growth.MagicDefense = 5
	hero.Growth.MoveSpeed = 0.1
	hero.Growth.AttackRange = 1
	hero.Growth.AttackSpeed = 0.01

	stats := heroStatsAtLevel(hero, MaxHeroLevel)
	steps := MaxHeroLevel - MinHeroLevel
	stepValue := float64(steps)

	if stats.MaxHP != hero.Base.HP+hero.Growth.HP*steps {
		t.Fatalf("max hp = %d", stats.MaxHP)
	}
	if stats.Attack != hero.Base.Attack+hero.Growth.Attack*stepValue {
		t.Fatalf("attack = %f", stats.Attack)
	}
	if stats.PhysicalDefense != hero.Base.PhysicalDefense+hero.Growth.PhysicalDefense*stepValue {
		t.Fatalf("physical defense = %f", stats.PhysicalDefense)
	}
	if stats.MagicDefense != hero.Base.MagicDefense+hero.Growth.MagicDefense*stepValue {
		t.Fatalf("magic defense = %f", stats.MagicDefense)
	}
	if stats.MoveSpeed != hero.Base.MoveSpeed+hero.Growth.MoveSpeed*float64(steps) {
		t.Fatalf("move speed = %f", stats.MoveSpeed)
	}
	if stats.AttackRange != hero.Base.AttackRange+hero.Growth.AttackRange*float64(steps) {
		t.Fatalf("attack range = %f", stats.AttackRange)
	}
	if stats.BaseAttackSpeed != hero.Base.AttackSpeed*(1+hero.Growth.AttackSpeed*float64(steps)) {
		t.Fatalf("base attack speed = %f", stats.BaseAttackSpeed)
	}
	if stats.AttackSpeed != stats.BaseAttackSpeed {
		t.Fatalf("attack speed = %f", stats.AttackSpeed)
	}
	if stats.CritChance != hero.Base.CritChance+hero.Growth.CritChance*float64(steps) {
		t.Fatalf("crit chance = %f", stats.CritChance)
	}
}

func TestHeroStatsLevelIsClamped(t *testing.T) {
	hero := testHeroConfig()
	hero.Growth.HP = 10

	low := heroStatsAtLevel(hero, -1)
	high := heroStatsAtLevel(hero, 99)

	if low.MaxHP != hero.Base.HP {
		t.Fatalf("low level max hp = %d", low.MaxHP)
	}
	if high.MaxHP != hero.Base.HP+hero.Growth.HP*(MaxHeroLevel-MinHeroLevel) {
		t.Fatalf("high level max hp = %d", high.MaxHP)
	}
}

func TestMinionKillGrantsExperience(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.Attack = 1000
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := w.entities["minion:red-melee-1"]

	w.ApplyInput("p1", protocolPlayerInputAttack(target.ID), 1, nil, 20)
	w.Tick(2, 20)

	if player.TotalExp != 62 {
		t.Fatalf("total exp = %f, want 62", player.TotalExp)
	}
	if player.Gold != 20 {
		t.Fatalf("gold = %f, want 20", player.Gold)
	}
}

func TestJungleKillGrantsGold(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := &Entity{
		ID:     "monster:red-buff",
		Kind:   EntityKindRedBuff,
		Team:   TeamNeutral,
		Stats:  Stats{HP: 1, MaxHP: 1},
		Radius: 20,
	}
	w.entities[target.ID] = target

	w.applyKillReward(player, target)

	if player.TotalExp != 100 {
		t.Fatalf("total exp = %f, want 100", player.TotalExp)
	}
	if player.Gold != 160 {
		t.Fatalf("gold = %f, want 160", player.Gold)
	}
}

func TestMinionKillRewardGrantsGold(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := &Entity{
		ID:     "minion:test-melee",
		Kind:   EntityKindMeleeMinion,
		Team:   TeamRed,
		Stats:  Stats{HP: 1, MaxHP: 1},
		Radius: 14,
	}

	w.applyKillReward(player, target)

	if player.TotalExp != 62 {
		t.Fatalf("total exp = %f, want 62", player.TotalExp)
	}
	if player.Gold != 20 {
		t.Fatalf("gold = %f, want 20", player.Gold)
	}
}

func TestBaronKillGrantsTeamGold(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	w.SpawnHero("p2", hero, TeamBlue)
	killer := w.entities[playerEntityID("p1")]
	ally := w.entities[playerEntityID("p2")]
	target := &Entity{
		ID:     "monster:baron",
		Kind:   EntityKindBaronNashor,
		Team:   TeamNeutral,
		Stats:  Stats{HP: 1, MaxHP: 1},
		Radius: 80,
	}
	w.entities[target.ID] = target

	w.applyKillReward(killer, target)

	if killer.Gold != 300 {
		t.Fatalf("killer gold = %f, want 300", killer.Gold)
	}
	if ally.Gold != 300 {
		t.Fatalf("ally gold = %f, want 300", ally.Gold)
	}
}

func TestHeroKillRewardGrantsGold(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := &Entity{
		ID:           "enemy:test-hero",
		Kind:         EntityKindEnemyHero,
		Team:         TeamRed,
		Level:        1,
		NextLevelExp: 280,
		Stats:        Stats{HP: 1, MaxHP: 1},
		Radius:       18,
	}

	w.applyKillReward(player, target)

	if player.TotalExp != 210 {
		t.Fatalf("total exp = %f, want 210", player.TotalExp)
	}
	if player.Gold != 300 {
		t.Fatalf("gold = %f, want 300", player.Gold)
	}
}

func TestExperienceLevelsUpAndRecalculatesStats(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	startMaxHP := player.Stats.MaxHP

	w.addExperience(player, 280)

	if player.Level != 2 {
		t.Fatalf("level = %d, want 2", player.Level)
	}
	if player.Exp != 0 {
		t.Fatalf("exp = %f, want 0", player.Exp)
	}
	if player.NextLevelExp != 340 {
		t.Fatalf("next level exp = %f, want 340", player.NextLevelExp)
	}
	if player.Stats.MaxHP <= startMaxHP {
		t.Fatalf("max hp did not grow: got %d start %d", player.Stats.MaxHP, startMaxHP)
	}
	if player.SkillPoints != 2 {
		t.Fatalf("skill points = %d, want 2", player.SkillPoints)
	}
}

func TestUpgradeSkillUsesSkillPointAndCapsBySlot(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	hero.Skills.R = swordRSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.SkillPoints = 10

	for i := 0; i < 6; i++ {
		w.ApplyInput("p1", protocolPlayerInputUpgrade("q"), 1, nil, 20)
	}
	for i := 0; i < 4; i++ {
		w.ApplyInput("p1", protocolPlayerInputUpgrade("r"), 1, nil, 20)
	}

	if player.Skills[swordQSkillID].Level != MaxBasicSkillLevel {
		t.Fatalf("q level = %d, want %d", player.Skills[swordQSkillID].Level, MaxBasicSkillLevel)
	}
	if player.Skills[swordRSkillID].Level != MaxUltSkillLevel {
		t.Fatalf("r level = %d, want %d", player.Skills[swordRSkillID].Level, MaxUltSkillLevel)
	}
	if player.SkillPoints != 2 {
		t.Fatalf("skill points = %d, want 2", player.SkillPoints)
	}
}

func TestUpgradeSkillWorksWhileActionLocked(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Control.ActionLockedUntilTick = 100

	w.ApplyInput("p1", protocolPlayerInputUpgrade("q"), 1, nil, 20)

	if player.Skills[swordQSkillID].Level != 1 {
		t.Fatalf("q level = %d, want 1", player.Skills[swordQSkillID].Level)
	}
	if player.SkillPoints != 0 {
		t.Fatalf("skill points = %d, want 0", player.SkillPoints)
	}
}

func TestDebugLevelUpRaisesHeroAndGrantsSkillPoint(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	w.ApplyInput("p1", protocolPlayerInputDebugLevelUp(), 1, nil, 20)

	if player.Level != 2 {
		t.Fatalf("level = %d, want 2", player.Level)
	}
	if player.SkillPoints != 2 {
		t.Fatalf("skill points = %d, want 2", player.SkillPoints)
	}
	player.Level = MaxHeroLevel
	player.SkillPoints = 0
	w.ApplyInput("p1", protocolPlayerInputDebugLevelUp(), 2, nil, 20)
	if player.Level != MaxHeroLevel {
		t.Fatalf("level = %d, want max %d", player.Level, MaxHeroLevel)
	}
	if player.SkillPoints != 0 {
		t.Fatalf("skill points = %d, want unchanged 0", player.SkillPoints)
	}
}

func TestUnlearnedSkillCannotCast(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.HeroID = swordHeroID
	hero.Skills.Q = swordQSkillID
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	target := w.entities["dummy:training-1"]
	target.Position = Vector2{X: player.Position.X + 200, Y: player.Position.Y}

	w.ApplyInput("p1", protocolPlayerInputCast(swordQSkillID, target.Position.X, target.Position.Y), 1, nil, 20)

	if target.Combat.LastDamage != 0 {
		t.Fatal("unlearned q should not damage target")
	}
}

func TestSpawnObjectEnemyHeroHasLevelRewardData(t *testing.T) {
	w := testWorld(t)
	id, ok := w.SpawnObject(EntityKindEnemyHero, TeamRed, 500, 600)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}

	entity := w.entities[id]
	if entity == nil {
		t.Fatalf("spawned enemy hero %s not found", id)
	}
	if entity.Level != MinHeroLevel {
		t.Fatalf("enemy hero level = %d, want %d", entity.Level, MinHeroLevel)
	}
	if entity.NextLevelExp != 280 {
		t.Fatalf("enemy hero next level exp = %f, want 280", entity.NextLevelExp)
	}
}

func TestEnemyHeroKillGrantsExperienceAndRemovesTarget(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	hero.Base.Attack = 2000
	hero.Base.AttackRange = 1000
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	id, ok := w.SpawnObject(EntityKindEnemyHero, TeamRed, player.Position.X+100, player.Position.Y)
	if !ok {
		t.Fatal("spawn enemy hero failed")
	}

	w.ApplyInput("p1", protocolPlayerInputAttack(id), 1, nil, 20)
	w.Tick(2, 20)

	if player.TotalExp != 210 {
		t.Fatalf("total exp = %f, want 210", player.TotalExp)
	}
	if player.Gold != 300 {
		t.Fatalf("gold = %f, want 300", player.Gold)
	}
	if _, ok := w.entities[id]; ok {
		t.Fatalf("dead enemy hero %s should be removed", id)
	}
	if player.Intent.AttackTargetID != "" {
		t.Fatalf("attack target id = %q, want empty", player.Intent.AttackTargetID)
	}
}

func TestDebugAbilityHasteToggleInput(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	w.ApplyInput("p1", protocolPlayerInputDebugAbilityHaste(200), 1, nil, 20)
	if player.Stats.AbilityHaste != 200 {
		t.Fatalf("ability haste = %f, want 200", player.Stats.AbilityHaste)
	}
	w.ApplyInput("p1", protocolPlayerInputDebugAbilityHaste(0), 2, nil, 20)
	if player.Stats.AbilityHaste != 0 {
		t.Fatalf("ability haste = %f, want 0", player.Stats.AbilityHaste)
	}
}

func TestDebugAbilityHasteBuffSurvivesLevelUp(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]

	w.ApplyInput("p1", protocolPlayerInputDebugAbilityHaste(200), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputDebugLevelUp(), 2, nil, 20)

	if player.Stats.AbilityHaste != 200 {
		t.Fatalf("ability haste after level up = %f, want 200", player.Stats.AbilityHaste)
	}
	if len(player.Buffs) != 1 || player.Buffs[0].ID != debugAbilityHasteBuffID {
		t.Fatalf("debug haste buff missing after level up: %+v", player.Buffs)
	}
}
