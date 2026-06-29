package config

import "testing"

func TestLevelConfigTotalExp(t *testing.T) {
	levels, err := LoadLevels("../../configs/levels.json")
	if err != nil {
		t.Fatal(err)
	}
	if levels.MaxLevel != 18 {
		t.Fatalf("max level = %d, want 18", levels.MaxLevel)
	}
	if levels.TotalExp != 20985 {
		t.Fatalf("total exp = %d, want 20985", levels.TotalExp)
	}
	if nextExp, ok := levels.NextExp(17); !ok || nextExp != 1915 {
		t.Fatalf("level 17 next exp = %d, ok=%v", nextExp, ok)
	}
	if nextExp, ok := levels.NextExp(18); !ok || nextExp != 0 {
		t.Fatalf("level 18 next exp = %d, ok=%v", nextExp, ok)
	}
}
