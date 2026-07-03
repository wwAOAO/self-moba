package config

import "testing"

func TestSkillConfigLoadsSwordCutRanges(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills.json")
	if err != nil {
		t.Fatal(err)
	}

	skill, ok := skills.Get("sword_cut")
	if !ok {
		t.Fatal("sword_cut skill not found")
	}
	if skill.Range != 475 {
		t.Fatalf("sword_cut range = %f, want 475", skill.Range)
	}
	if skill.Meta["whirlwindRange"] != 900 {
		t.Fatalf("sword_cut whirlwind range = %f, want 900", skill.Meta["whirlwindRange"])
	}
	if skill.Meta["eqRadius"] != 375 {
		t.Fatalf("sword_cut eq radius = %f, want 375", skill.Meta["eqRadius"])
	}
	if len(skill.MetaLists["baseDamage"]) != 5 || skill.MetaLists["baseDamage"][4] != 120 {
		t.Fatalf("sword_cut base damage = %#v", skill.MetaLists["baseDamage"])
	}
}

func TestSkillConfigLoadsSwordSkillTables(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills.json")
	if err != nil {
		t.Fatal(err)
	}

	windWall, ok := skills.Get("sword_wind_wall")
	if !ok {
		t.Fatal("sword_wind_wall skill not found")
	}
	if windWall.MetaLists["width"][0] != 300 || windWall.MetaLists["cooldownMs"][4] != 18000 {
		t.Fatalf("wind wall config = meta:%#v lists:%#v", windWall.Meta, windWall.MetaLists)
	}

	sweepingBlade, ok := skills.Get("sword_sweeping_blade")
	if !ok {
		t.Fatal("sword_sweeping_blade skill not found")
	}
	if sweepingBlade.Meta["apRatio"] != 0.6 || sweepingBlade.MetaLists["targetCooldownMs"][0] != 10000 {
		t.Fatalf("sweeping blade config = meta:%#v lists:%#v", sweepingBlade.Meta, sweepingBlade.MetaLists)
	}

	lastBreath, ok := skills.Get("sword_storm")
	if !ok {
		t.Fatal("sword_storm skill not found")
	}
	if lastBreath.Meta["lastBreathDurationSeconds"] != 15 || lastBreath.MetaLists["cooldownMs"][2] != 30000 {
		t.Fatalf("last breath config = meta:%#v lists:%#v", lastBreath.Meta, lastBreath.MetaLists)
	}
}

func TestSkillConfigLoadsWarriorJudgment(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills.json")
	if err != nil {
		t.Fatal(err)
	}

	judgment, ok := skills.Get("judgment")
	if !ok {
		t.Fatal("judgment skill not found")
	}
	if judgment.Range != 220 || judgment.Meta["baseSpins"] != 7 {
		t.Fatalf("judgment config = meta:%#v lists:%#v", judgment.Meta, judgment.MetaLists)
	}
	if judgment.MetaLists["cooldownMs"][0] != 9000 || judgment.MetaLists["cooldownMs"][4] != 6000 {
		t.Fatalf("judgment cooldowns = %#v", judgment.MetaLists["cooldownMs"])
	}
	if judgment.MetaLists["critAdRatio"][4] != 0.676 {
		t.Fatalf("judgment crit ad ratio = %#v", judgment.MetaLists["critAdRatio"])
	}

	justice, ok := skills.Get("justice")
	if !ok {
		t.Fatal("justice skill not found")
	}
	if justice.MetaLists["baseDamage"][2] != 350 || justice.MetaLists["missingHPRatio"][2] != 0.35 {
		t.Fatalf("justice config = meta:%#v lists:%#v", justice.Meta, justice.MetaLists)
	}
	if justice.MetaLists["cooldownMs"][0] != 120000 || justice.MetaLists["cooldownMs"][2] != 80000 {
		t.Fatalf("justice cooldowns = %#v", justice.MetaLists["cooldownMs"])
	}
}

func TestSkillConfigLoadsFromDirectory(t *testing.T) {
	skills, err := LoadSkills("../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := skills.Get("sword_cut"); !ok {
		t.Fatal("sword_cut skill not found from directory")
	}
	if _, ok := skills.Get("mage_r"); !ok {
		t.Fatal("mage_r skill not found from directory")
	}
}
