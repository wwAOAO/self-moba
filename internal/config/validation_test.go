package config

import "testing"

// TestHeroRadiiMatchWorldScale 验证所有英雄采用与地图比例匹配的职业体型半径。
func TestHeroRadiiMatchWorldScale(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{
		"archer": 28, "crossbowman": 28, "gunner": 28, "explorer": 28,
		"mage": 28, "fire_mage": 28, "frostmage": 28,
		"killer": 30, "ninja": 30, "shadow_assassin": 30,
		"blade": 32, "sword": 32, "monk": 32,
		"berserker": 34, "warrior": 34, "doctor": 34,
		"robot": 36, "butcher": 38, "tank": 38,
	}
	for heroID, radius := range want {
		hero, ok := heroes.Get(heroID)
		if !ok {
			t.Fatalf("hero %s is missing", heroID)
		}
		if hero.Radius != radius {
			t.Fatalf("hero %s radius = %v, want %v", heroID, hero.Radius, radius)
		}
	}
}

func TestValidateGameConfigLoadsCurrentTables(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	levels, err := LoadLevels("../../configs/levels.json")
	if err != nil {
		t.Fatal(err)
	}
	rewards, err := LoadRewards("../../configs/rewards.json")
	if err != nil {
		t.Fatal(err)
	}
	equipment, err := LoadEquipment("../../configs/equipment")
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateGameConfig(GameConfig{
		Heroes:    heroes,
		Skills:    skills,
		Levels:    levels,
		Rewards:   rewards,
		Equipment: equipment,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestKillerStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("killer")
	if !ok {
		t.Fatal("killer hero is missing")
	}

	if hero.Resource != "none" {
		t.Fatalf("resource = %q, want none", hero.Resource)
	}
	if hero.Base.HP != 602 || hero.Growth.HP != 94 ||
		hero.Base.HPRegen5 != 7.5 || hero.Growth.HPRegen5 != 0.7 ||
		hero.Base.Attack != 53 || hero.Growth.Attack != 3.2 ||
		hero.Base.AttackSpeed != 0.656 || hero.Growth.AttackSpeed != 0.0244 ||
		hero.Base.PhysicalDefense != 28 || hero.Growth.PhysicalDefense != 4 ||
		hero.Base.MagicDefense != 34 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.AttackRange != 125 {
		t.Fatalf("killer stats do not match the configured base attributes: %+v", hero)
	}
}

func TestShadowAssassinStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("shadow_assassin")
	if !ok {
		t.Fatal("shadow assassin hero is missing")
	}

	if hero.Resource != "mp" {
		t.Fatalf("resource = %q, want mp", hero.Resource)
	}
	if hero.Name != "影刃" {
		t.Fatalf("name = %q, want 影刃", hero.Name)
	}
	if hero.Base.HP != 588 || hero.Growth.HP != 95 ||
		hero.Base.HPRegen5 != 8.5 || hero.Growth.HPRegen5 != 0.75 ||
		hero.Base.MP != 377.2 || hero.Growth.MP != 37 ||
		hero.Base.MPRegen5 != 7.6 || hero.Growth.MPRegen5 != 0.8 ||
		hero.Base.Attack != 68 || hero.Growth.Attack != 3.1 ||
		hero.Base.AttackSpeed != 0.625 || hero.Growth.AttackSpeed != 0.029 ||
		hero.Base.PhysicalDefense != 30 || hero.Growth.PhysicalDefense != 3.5 ||
		hero.Base.MagicDefense != 39 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.MoveSpeed != 335 || hero.Base.AttackRange != 125 {
		t.Fatalf("shadow assassin stats do not match the configured base attributes: %+v", hero)
	}
}

func TestShadowAssassinQConfig(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	skill, ok := skills.Get("shadow_assassin_q")
	if !ok {
		t.Fatal("shadow assassin q is missing")
	}
	if skill.Name != "刺客诡道" || skill.Type != "self_buff" || skill.CooldownMS != 8000 ||
		skill.Meta["bonusAdRatio"] != 0.3 || skill.Meta["bleedBonusAdRatio"] != 1.2 ||
		skill.Meta["bleedDurationSeconds"] != 6 || skill.Meta["bleedSlow"] != 0.1 {
		t.Fatalf("shadow assassin q metadata does not match: %+v", skill)
	}
	for key, want := range map[string][]float64{
		"manaCost":    {40, 45, 50, 55, 60},
		"cooldownMs":  {8000, 7000, 6000, 5000, 4000},
		"bonusDamage": {30, 60, 90, 120, 150},
		"bleedDamage": {10, 20, 30, 40, 50},
	} {
		got := skill.MetaLists[key]
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s = %v, want %v", key, got, want)
			}
		}
	}
}

func TestShadowAssassinWConfig(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	skill, ok := skills.Get("shadow_assassin_w")
	if !ok {
		t.Fatal("shadow assassin w is missing")
	}
	if skill.Name != "斩草除根" || skill.Type != "directional_projectile" || skill.CooldownMS != 10000 || skill.Range != 850 ||
		skill.Meta["bonusAdRatio"] != 0.6 || skill.Meta["slowSeconds"] != 2 {
		t.Fatalf("shadow assassin w metadata does not match: %+v", skill)
	}
	for key, want := range map[string][]float64{
		"manaCost":   {60, 65, 70, 75, 80},
		"baseDamage": {30, 55, 80, 105, 130},
		"slow":       {0.2, 0.25, 0.3, 0.35, 0.4},
	} {
		got := skill.MetaLists[key]
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s = %v, want %v", key, got, want)
			}
		}
	}
}

func TestShadowAssassinEConfig(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	skill, ok := skills.Get("shadow_assassin_e")
	if !ok {
		t.Fatal("shadow assassin e is missing")
	}
	if skill.Name != "割喉之战" || skill.Type != "targeted_blink" || skill.CooldownMS != 18000 || skill.Range != 700 ||
		skill.Meta["rootSeconds"] != 1 || skill.Meta["damageAmpSeconds"] != 3 {
		t.Fatalf("shadow assassin e metadata does not match: %+v", skill)
	}
	for key, want := range map[string][]float64{
		"manaCost":   {35, 40, 45, 50, 55},
		"cooldownMs": {18000, 16000, 14000, 12000, 10000},
		"damageAmp":  {0.03, 0.06, 0.09, 0.12, 0.15},
	} {
		got := skill.MetaLists[key]
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s = %v, want %v", key, got, want)
			}
		}
	}
}

func TestShadowAssassinRConfig(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	skill, ok := skills.Get("shadow_assassin_r")
	if !ok {
		t.Fatal("shadow assassin r is missing")
	}
	if skill.Name != "暗影突袭" || skill.Type != "self_buff" || skill.CooldownMS != 75000 ||
		skill.Meta["invisibilitySeconds"] != 2.5 || skill.Meta["moveSpeedBonus"] != 0.4 || skill.Meta["bonusAdRatio"] != 0.9 {
		t.Fatalf("shadow assassin r metadata does not match: %+v", skill)
	}
	for key, want := range map[string][]float64{
		"manaCost":   {80, 90, 100},
		"cooldownMs": {75000, 65000, 55000},
		"damage":     {120, 190, 260},
	} {
		got := skill.MetaLists[key]
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s = %v, want %v", key, got, want)
			}
		}
	}
}

func TestCrossbowmanStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("crossbowman")
	if !ok {
		t.Fatal("crossbowman hero is missing")
	}

	if hero.Resource != "mp" {
		t.Fatalf("resource = %q, want mp", hero.Resource)
	}
	if hero.Base.HP != 515 || hero.Growth.HP != 89 ||
		hero.Base.HPRegen5 != 3.5 || hero.Growth.HPRegen5 != 0.55 ||
		hero.Base.MP != 231.8 || hero.Growth.MP != 35 ||
		hero.Base.MPRegen5 != 6.972 || hero.Growth.MPRegen5 != 0.4 ||
		hero.Base.Attack != 60 || hero.Growth.Attack != 2.36 ||
		hero.Base.AttackSpeed != 0.658 || hero.Growth.AttackSpeed != 0.033 ||
		hero.Base.PhysicalDefense != 23 || hero.Growth.PhysicalDefense != 3.4 ||
		hero.Base.MagicDefense != 30 || hero.Growth.MagicDefense != 0.5 ||
		hero.Base.MoveSpeed != 330 || hero.Base.AttackRange != 550 {
		t.Fatalf("crossbowman stats do not match the configured base attributes: %+v", hero)
	}
}

func TestMonkStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("monk")
	if !ok {
		t.Fatal("monk hero is missing")
	}

	if hero.Resource != "energy" {
		t.Fatalf("resource = %q, want energy", hero.Resource)
	}
	if hero.Base.HP != 428 || hero.Growth.HP != 85 ||
		hero.Base.HPRegen5 != 6.25 || hero.Growth.HPRegen5 != 0.7 ||
		hero.Base.MP != 200 || hero.Base.MPRegen5 != 50 ||
		hero.Base.Attack != 55.8 || hero.Growth.Attack != 3.2 ||
		hero.Base.AttackSpeed != 0.651 || hero.Growth.AttackSpeed != 0.03 ||
		hero.Base.PhysicalDefense != 16 || hero.Growth.PhysicalDefense != 3.7 ||
		hero.Base.MagicDefense != 30 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.MoveSpeed != 350 || hero.Base.AttackRange != 125 {
		t.Fatalf("monk stats do not match the configured base attributes: %+v", hero)
	}
}

func TestButcherStats(t *testing.T) {
	heroes, err := LoadHeroes("../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	hero, ok := heroes.Get("butcher")
	if !ok {
		t.Fatal("butcher hero is missing")
	}

	if hero.Resource != "mp" {
		t.Fatalf("resource = %q, want mp", hero.Resource)
	}
	if hero.Base.HP != 432 || hero.Growth.HP != 86 ||
		hero.Base.HPRegen5 != 7.45 || hero.Growth.HPRegen5 != 0.55 ||
		hero.Base.MP != 200 || hero.Growth.MP != 50 ||
		hero.Base.MPRegen5 != 7.45 || hero.Growth.MPRegen5 != 0.7 ||
		hero.Base.Attack != 52 || hero.Growth.Attack != 3.3 ||
		hero.Base.AttackSpeed != 0.613 || hero.Growth.AttackSpeed != 0.0098 ||
		hero.Base.PhysicalDefense != 12 || hero.Growth.PhysicalDefense != 1.25 ||
		hero.Base.MagicDefense != 30 || hero.Growth.MagicDefense != 1.25 ||
		hero.Base.MoveSpeed != 325 || hero.Base.AttackRange != 175 {
		t.Fatalf("butcher stats do not match the configured base attributes: %+v", hero)
	}
}

func TestValidateEquipmentComponentsRejectsCycles(t *testing.T) {
	equipment, err := NewEquipmentStore([]EquipmentConfig{
		{EquipmentID: "a", Name: "A", Price: 100, Tier: 2, Components: []string{"b"}},
		{EquipmentID: "b", Name: "B", Price: 100, Tier: 2, Components: []string{"a"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateEquipmentComponents(equipment); err == nil {
		t.Fatal("cycle should be rejected")
	}
}

func TestValidateEquipmentComponentsRejectsHigherTierComponents(t *testing.T) {
	equipment, err := NewEquipmentStore([]EquipmentConfig{
		{EquipmentID: "tier2", Name: "Tier 2", Price: 100, Tier: 2, Components: []string{"tier3"}},
		{EquipmentID: "tier3", Name: "Tier 3", Price: 100, Tier: 3},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateEquipmentComponents(equipment); err == nil {
		t.Fatal("higher-tier component should be rejected")
	}
}
