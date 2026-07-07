package explorer

import (
	"l-battle/internal/world"
	"math"
	"testing"
)

func TestRisingSpellForceStacksAndExpiresIndependently(t *testing.T) {
	source := &world.Entity{ID: "explorer", HeroID: heroID, Team: world.TeamBlue}
	target := &world.Entity{ID: "target", Team: world.TeamRed, Stats: world.Stats{HP: 100}}

	for tick := uint64(0); tick < 6; tick++ {
		OnSkillHit(nil, source, target, tick, 20)
	}
	if got := len(source.Passive.ExplorerSpellForce); got != 5 {
		t.Fatalf("stacks = %d, want 5", got)
	}

	stats := world.Stats{BaseAttackSpeed: 0.625, AttackSpeed: 0.625, AttackSpeedRatio: 0.638}
	ApplyStats(nil, source, &stats)
	if math.Abs(stats.AttackSpeedBonus-0.5) > 0.000001 {
		t.Fatalf("attack speed bonus = %f, want 0.5", stats.AttackSpeedBonus)
	}

	Tick(nil, source, 121, 20)
	if got := len(source.Passive.ExplorerSpellForce); got != 4 {
		t.Fatalf("stacks after first independent expiry = %d, want 4", got)
	}
}
