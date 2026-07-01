package world

import "testing"

func TestBuyEquipmentSpendsGoldAndAppliesStats(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 500
	baseAttack := player.Stats.Attack
	baseMaxHP := player.Stats.MaxHP

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("small_blade"), 1, nil, 20)

	if player.Gold != 25 {
		t.Fatalf("gold = %f, want 25", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "small_blade" {
		t.Fatalf("equipment = %+v, want small_blade", player.Equipment)
	}
	if player.Stats.Attack != baseAttack+8 {
		t.Fatalf("attack = %f, want %f", player.Stats.Attack, baseAttack+8)
	}
	if player.Stats.MaxHP != baseMaxHP+80 {
		t.Fatalf("max hp = %d, want %d", player.Stats.MaxHP, baseMaxHP+80)
	}
	if player.Stats.Omnivamp != 0.025 {
		t.Fatalf("omnivamp = %f, want 0.025", player.Stats.Omnivamp)
	}
}

func TestSellEquipmentRefundsHalfPriceAndRemovesStats(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 500
	baseAttack := player.Stats.Attack

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("small_blade"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputSellEquipment(1), 2, nil, 20)

	if player.Gold != 262 {
		t.Fatalf("gold = %f, want 262", player.Gold)
	}
	if len(player.Equipment) != 0 {
		t.Fatalf("equipment = %+v, want empty", player.Equipment)
	}
	if player.Stats.Attack != baseAttack {
		t.Fatalf("attack = %f, want %f", player.Stats.Attack, baseAttack)
	}
}

func TestBuyEquipmentStopsAtSixSlots(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 10000

	for i := 0; i < 7; i++ {
		w.ApplyInput("p1", protocolPlayerInputBuyEquipment("iron_bow"), uint64(i+1), nil, 20)
	}

	if len(player.Equipment) != 6 {
		t.Fatalf("equipment slots = %d, want 6", len(player.Equipment))
	}
	if player.Gold != 7600 {
		t.Fatalf("gold = %f, want 7600", player.Gold)
	}
}

func TestBuyEquipmentAllowsOnlyOneBootsGroup(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 1000
	baseMoveSpeed := player.Stats.MoveSpeed

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("boots"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("boots"), 2, nil, 20)

	if len(player.Equipment) != 1 {
		t.Fatalf("equipment slots = %d, want 1", len(player.Equipment))
	}
	if player.Gold != 650 {
		t.Fatalf("gold = %f, want 650", player.Gold)
	}
	if player.Stats.MoveSpeed != baseMoveSpeed+25 {
		t.Fatalf("move speed = %f, want %f", player.Stats.MoveSpeed, baseMoveSpeed+25)
	}
}

func TestBuyCompositeEquipmentCanBuyFullPriceWithoutComponents(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 740
	baseMagicDefense := player.Stats.MagicDefense

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("negatron_cloak"), 1, nil, 20)

	if player.Gold != 0 {
		t.Fatalf("gold = %f, want 0", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "negatron_cloak" {
		t.Fatalf("equipment = %+v, want negatron_cloak", player.Equipment)
	}
	if player.Stats.MagicDefense != baseMagicDefense+48 {
		t.Fatalf("magic defense = %f, want %f", player.Stats.MagicDefense, baseMagicDefense+48)
	}
}

func TestBuyCompositeEquipmentConsumesComponentAndCombineCost(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 740

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("null_magic_mantle"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("negatron_cloak"), 2, nil, 20)

	if player.Gold != 0 {
		t.Fatalf("gold = %f, want 0", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "negatron_cloak" {
		t.Fatalf("equipment = %+v, want negatron_cloak", player.Equipment)
	}
}

func TestBuyCompositeEquipmentBuysAffordableMissingComponent(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 500

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("negatron_cloak"), 1, nil, 20)

	if player.Gold != 50 {
		t.Fatalf("gold = %f, want 50", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "null_magic_mantle" {
		t.Fatalf("equipment = %+v, want null_magic_mantle", player.Equipment)
	}
}

func TestBuyCompositeEquipmentShowsNotEnoughGold(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 200

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("negatron_cloak"), 7, nil, 20)

	if player.Gold != 200 {
		t.Fatalf("gold = %f, want 200", player.Gold)
	}
	if len(player.Equipment) != 0 {
		t.Fatalf("equipment = %+v, want empty", player.Equipment)
	}
	if player.Message != "金币不足" || player.MessageTick != 7 {
		t.Fatalf("message = %q tick=%d, want 金币不足 tick=7", player.Message, player.MessageTick)
	}
}

func TestBuyChainVestConsumesClothArmorAndCombineCost(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 800
	baseArmor := player.Stats.PhysicalDefense

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("cloth_armor"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("chain_vest"), 2, nil, 20)

	if player.Gold != 0 {
		t.Fatalf("gold = %f, want 0", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "chain_vest" {
		t.Fatalf("equipment = %+v, want chain_vest", player.Equipment)
	}
	if player.Stats.PhysicalDefense != baseArmor+30 {
		t.Fatalf("physical defense = %f, want %f", player.Stats.PhysicalDefense, baseArmor+30)
	}
}

func TestInfinityEdgeConsumesComponentsAndAppliesCritDamage(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 3500
	baseAttack := player.Stats.Attack
	baseCrit := player.Stats.CritChance

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("bf_sword"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("pickaxe"), 2, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("cloak_of_agility"), 3, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("infinity_edge"), 4, nil, 20)

	if player.Gold != 0 {
		t.Fatalf("gold = %f, want 0", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "infinity_edge" {
		t.Fatalf("equipment = %+v, want infinity_edge", player.Equipment)
	}
	if player.Stats.Attack != baseAttack+70 {
		t.Fatalf("attack = %f, want %f", player.Stats.Attack, baseAttack+70)
	}
	if player.Stats.CritChance != baseCrit+0.2 {
		t.Fatalf("crit chance = %f, want %f", player.Stats.CritChance, baseCrit+0.2)
	}
	if got := w.critDamageMultiplier(player); got != 2.5 {
		t.Fatalf("crit damage multiplier = %f, want 2.5", got)
	}
}

func TestDuplicateInfinityEdgeCritDamageOnlyAppliesOnce(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 7000

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("infinity_edge"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("infinity_edge"), 2, nil, 20)

	if len(player.Equipment) != 2 {
		t.Fatalf("equipment slots = %d, want 2", len(player.Equipment))
	}
	if got := w.critDamageMultiplier(player); got != 2.5 {
		t.Fatalf("crit damage multiplier = %f, want 2.5", got)
	}
}

func TestPhantomDancerConsumesTwoRecurveBowsAndCloak(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 2800
	baseAttackSpeedBonus := player.Stats.AttackSpeedBonus
	baseCrit := player.Stats.CritChance
	baseMoveSpeed := player.Stats.MoveSpeed

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("recurve_bow"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("cloak_of_agility"), 2, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("recurve_bow"), 3, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("phantom_dancer"), 4, nil, 20)

	if player.Gold != 0 {
		t.Fatalf("gold = %f, want 0", player.Gold)
	}
	if len(player.Equipment) != 1 || player.Equipment[0].EquipmentID != "phantom_dancer" {
		t.Fatalf("equipment = %+v, want phantom_dancer", player.Equipment)
	}
	if player.Stats.AttackSpeedBonus != baseAttackSpeedBonus+0.3 {
		t.Fatalf("attack speed bonus = %f, want %f", player.Stats.AttackSpeedBonus, baseAttackSpeedBonus+0.3)
	}
	if player.Stats.CritChance != baseCrit+0.3 {
		t.Fatalf("crit chance = %f, want %f", player.Stats.CritChance, baseCrit+0.3)
	}
	if player.Stats.MoveSpeed != baseMoveSpeed*1.07 {
		t.Fatalf("move speed = %f, want %f", player.Stats.MoveSpeed, baseMoveSpeed*1.07)
	}
}

func TestPhantomDancerLowHealthShieldAndDamageReduction(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 2800
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("phantom_dancer"), 1, nil, 20)
	player.Stats.HP = int(float64(player.Stats.MaxHP) * 0.31)
	attacker := &Entity{ID: "enemy", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.applyDamage(attacker, player, 20, 20)

	if player.Passive.Shield != 200 {
		t.Fatalf("shield = %d, want 200", player.Passive.Shield)
	}
	if got := equipmentLowHealthDamageReduce(player); got != 0.1 {
		t.Fatalf("damage reduce = %f, want 0.1", got)
	}
	beforeShield := player.Passive.Shield
	player.Stats.PhysicalDefense = 0
	damage := physicalDamageAfterResistance(attacker, player, 100, player.Combat.LastHitTick)
	w.applyResolvedDamage(attacker, player, damage, "physical", sustainSingleTargetSkill, 20)
	if player.Passive.Shield != beforeShield-90 {
		t.Fatalf("shield after reduced damage = %d, want %d", player.Passive.Shield, beforeShield-90)
	}
}

func TestCatalystRestoresHpAndMpOnLevelUp(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 900
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("catalyst_of_aeons"), 1, nil, 20)
	player.Stats.HP = 100
	player.Stats.MP = 50

	w.debugLevelUp(player)

	if player.Stats.HP <= 100 {
		t.Fatalf("hp = %d, want restored above 100", player.Stats.HP)
	}
	if player.Stats.MP <= 50 {
		t.Fatalf("mp = %f, want restored above 50", player.Stats.MP)
	}
}

func TestRaptorCloakAddsOutOfCombatMoveSpeed(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 850
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("raptor_cloak"), 1, nil, 20)
	baseMoveSpeed := player.Stats.MoveSpeed
	player.Combat.LastHitTick = 10

	if got := EffectiveMoveSpeedAtTick(player, 50); got != baseMoveSpeed {
		t.Fatalf("in combat move speed = %f, want %f", got, baseMoveSpeed)
	}
	if got := EffectiveMoveSpeedAtTick(player, 111); got != baseMoveSpeed+20 {
		t.Fatalf("out of combat move speed = %f, want %f", got, baseMoveSpeed+20)
	}
}

func TestDuplicateEquipmentEffectsOnlyApplyOnce(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 3000
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("raptor_cloak"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("raptor_cloak"), 2, nil, 20)
	baseMoveSpeed := player.Stats.MoveSpeed

	if len(player.Equipment) != 2 {
		t.Fatalf("equipment slots = %d, want 2", len(player.Equipment))
	}
	if got := EffectiveMoveSpeedAtTick(player, 200); got != baseMoveSpeed+20 {
		t.Fatalf("out of combat move speed = %f, want %f", got, baseMoveSpeed+20)
	}
}

func TestDuplicateEquipmentStatsStillStack(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 1000
	baseAttack := player.Stats.Attack

	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("short_sword"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("short_sword"), 2, nil, 20)

	if player.Stats.Attack != baseAttack+20 {
		t.Fatalf("attack = %f, want %f", player.Stats.Attack, baseAttack+20)
	}
}

func TestSeekersArmguardGainsStatsOnUnitKill(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 950
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("seekers_armguard"), 1, nil, 20)
	baseArmor := player.Stats.PhysicalDefense
	baseAP := player.Stats.AbilityPower
	target := &Entity{ID: "unit", Kind: EntityKindDummy, Team: TeamRed, Stats: Stats{HP: 0, MaxHP: 100}}

	for i := 0; i < 5; i++ {
		w.applyKillReward(player, target)
	}

	if player.Equipment[0].Stacks != 5 {
		t.Fatalf("stacks = %f, want 5", player.Equipment[0].Stacks)
	}
	if player.Stats.PhysicalDefense != baseArmor+1 {
		t.Fatalf("physical defense = %f, want %f", player.Stats.PhysicalDefense, baseArmor+1)
	}
	if player.Stats.AbilityPower != baseAP+1 {
		t.Fatalf("ability power = %d, want %d", player.Stats.AbilityPower, baseAP+1)
	}
}

func TestDuplicateSeekersArmguardGrowthOnlyAppliesOnce(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 2000
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("seekers_armguard"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("seekers_armguard"), 2, nil, 20)
	target := &Entity{ID: "unit", Kind: EntityKindDummy, Team: TeamRed, Stats: Stats{HP: 0, MaxHP: 100}}

	w.applyKillReward(player, target)

	if player.Equipment[0].Stacks != 1 {
		t.Fatalf("first stacks = %f, want 1", player.Equipment[0].Stacks)
	}
	if player.Equipment[1].Stacks != 0 {
		t.Fatalf("second stacks = %f, want 0", player.Equipment[1].Stacks)
	}
}

func TestWoodenShieldHealsAfterHeroHit(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 1000
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("wooden_shield"), 1, nil, 20)
	player.Stats.HP = 500
	attacker := &Entity{ID: "enemy", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.applyDamage(attacker, player, 100, 20)

	if player.Stats.HP != 405 {
		t.Fatalf("hp after hero hit heal = %d, want 405", player.Stats.HP)
	}
}

func TestDuplicateWoodenShieldHeroHitHealOnlyAppliesOnce(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 1000
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("wooden_shield"), 1, nil, 20)
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("wooden_shield"), 2, nil, 20)
	player.Stats.HP = 500
	attacker := &Entity{ID: "enemy", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.applyDamage(attacker, player, 100, 20)

	if player.Stats.HP != 405 {
		t.Fatalf("hp after duplicate hero hit heal = %d, want 405", player.Stats.HP)
	}
}

func TestEquipmentOmnivampHealsFromActualDamage(t *testing.T) {
	w := testWorld(t)
	hero := testHeroConfig()
	w.SpawnHero("p1", hero, TeamBlue)
	player := w.entities[playerEntityID("p1")]
	player.Gold = 500
	w.ApplyInput("p1", protocolPlayerInputBuyEquipment("small_blade"), 1, nil, 20)
	player.Stats.HP = 500
	target := &Entity{ID: "target", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.applyDamage(player, target, 200, 20)

	if player.Stats.HP != 505 {
		t.Fatalf("hp after omnivamp = %d, want 505", player.Stats.HP)
	}
}

func TestAOEOmnivampUsesDecayHealingPowerAndGrievousWounds(t *testing.T) {
	w := testWorld(t)
	source := &Entity{
		ID:   "source",
		Kind: EntityKindPlayer,
		Team: TeamBlue,
		Stats: Stats{
			HP:             500,
			MaxHP:          1000,
			Omnivamp:       0.2,
			HealingPower:   0.3,
			GrievousWounds: 0.4,
		},
	}
	target := &Entity{ID: "target", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.applyAOEDamage(source, target, 200, "magic", 20)

	if source.Stats.HP != 510 {
		t.Fatalf("hp after aoe omnivamp = %d, want 510", source.Stats.HP)
	}
}

func TestLifeStealOnlyAppliesToBasicAttacks(t *testing.T) {
	w := testWorld(t)
	source := &Entity{
		ID:   "source",
		Kind: EntityKindPlayer,
		Team: TeamBlue,
		Stats: Stats{
			HP:        500,
			MaxHP:     1000,
			LifeSteal: 0.2,
		},
	}
	target := &Entity{ID: "target", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}}

	w.applyDamage(source, target, 100, 20)
	if source.Stats.HP != 500 {
		t.Fatalf("hp after skill with life steal = %d, want 500", source.Stats.HP)
	}
	w.applyBasicAttackDamage(source, target, 100, 20)
	if source.Stats.HP != 520 {
		t.Fatalf("hp after basic attack life steal = %d, want 520", source.Stats.HP)
	}
}
