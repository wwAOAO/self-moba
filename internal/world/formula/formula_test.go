package formula

import (
	"math"
	"testing"
)

func TestDamageAfterResistance(t *testing.T) {
	tests := []struct {
		resistance float64
		want       int
	}{
		{resistance: 0, want: 100},
		{resistance: 50, want: 67},
		{resistance: 100, want: 50},
		{resistance: -50, want: 200},
	}
	for _, tt := range tests {
		if got := DamageAfterResistance(100, tt.resistance, 0); got != tt.want {
			t.Fatalf("damage after %f resistance = %d, want %d", tt.resistance, got, tt.want)
		}
	}
}

func TestEffectiveResistance(t *testing.T) {
	if got := EffectiveResistance(100, 0.3, 15); got != 55 {
		t.Fatalf("effective resistance = %f, want 55", got)
	}
	if got := EffectiveResistance(20, 0.5, 40); got != 0 {
		t.Fatalf("penetration-created resistance = %f, want 0", got)
	}
	if got := EffectiveResistance(-50, 0.3, 15); got != -50 {
		t.Fatalf("forced negative resistance = %f, want -50", got)
	}
}

func TestFinalAttackSpeed(t *testing.T) {
	if got := FinalAttackSpeed(0.65, 0.5, 0.7, 0); math.Abs(got-0.8775) > 0.000001 {
		t.Fatalf("attack speed = %f, want 0.8775", got)
	}
	if got := FinalAttackSpeed(2, 1, 1, 0); got != 2.5 {
		t.Fatalf("attack speed cap = %f, want 2.5", got)
	}
	if got := FinalAttackSpeed(1, 1, 1, 0.3); got != 1.4 {
		t.Fatalf("slowed attack speed = %f, want 1.4", got)
	}
}

func TestStackingFormulas(t *testing.T) {
	if got := StackDamageReduction(0.2, 0.5); math.Abs(got-0.6) > 0.0001 {
		t.Fatalf("stacked reduction = %f, want 0.6", got)
	}
	if got := StackTenacity(0.3, 0.6); math.Abs(got-0.72) > 0.0001 {
		t.Fatalf("stacked tenacity = %f, want 0.72", got)
	}
}

func TestDeterministicCritRoll(t *testing.T) {
	first := DeterministicCritRoll("a", "b", 10)
	second := DeterministicCritRoll("a", "b", 10)
	if first != second {
		t.Fatalf("crit roll should be deterministic: %f != %f", first, second)
	}
	if first < 0 || first >= 1 {
		t.Fatalf("crit roll = %f, want in [0, 1)", first)
	}
}
