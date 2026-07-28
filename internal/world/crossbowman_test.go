package world

import "testing"

func TestCrossbowmanUsesRangedBasicAttacks(t *testing.T) {
	if !isRangedBasicAttacker(&Entity{HeroID: crossbowmanHeroID}) {
		t.Fatal("crossbowman should use ranged basic attacks")
	}
}
