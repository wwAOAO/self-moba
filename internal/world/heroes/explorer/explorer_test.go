package explorer

import (
	"l-battle/internal/world"
	"math"
	"testing"
)

func TestRisingSpellForceRefreshesSharedExpiryAndClearsAllStacks(t *testing.T) {
	source := &world.Entity{ID: "explorer", HeroID: heroID, Team: world.TeamBlue}
	target := &world.Entity{ID: "target", Team: world.TeamRed, Stats: world.Stats{HP: 100}}

	for tick := uint64(0); tick < 6; tick++ {
		OnSkillHit(nil, source, target, tick, 20)
	}
	if got := source.Passive.ExplorerSpellForceStacks; got != 5 {
		t.Fatalf("stacks = %d, want 5", got)
	}
	if got := source.Passive.ExplorerSpellForceExpiresAt; got != 125 {
		t.Fatalf("shared expiry = %d, want 125", got)
	}

	stats := world.Stats{BaseAttackSpeed: 0.625, AttackSpeed: 0.625, AttackSpeedRatio: 0.638}
	ApplyStats(nil, source, &stats)
	if math.Abs(stats.AttackSpeedBonus-0.5) > 0.000001 {
		t.Fatalf("attack speed bonus = %f, want 0.5", stats.AttackSpeedBonus)
	}

	Tick(nil, source, 121, 20)
	if got := source.Passive.ExplorerSpellForceStacks; got != 5 {
		t.Fatalf("stacks before shared expiry = %d, want 5", got)
	}

	OnSkillHit(nil, source, target, 121, 20)
	if got := source.Passive.ExplorerSpellForceExpiresAt; got != 241 {
		t.Fatalf("refreshed expiry = %d, want 241", got)
	}
	Tick(nil, source, 240, 20)
	if source.Passive.ExplorerSpellForceStacks != 5 {
		t.Fatalf("stacks before refreshed expiry = %d, want 5", source.Passive.ExplorerSpellForceStacks)
	}
	Tick(nil, source, 241, 20)
	if source.Passive.ExplorerSpellForceStacks != 0 || source.Passive.ExplorerSpellForceExpiresAt != 0 {
		t.Fatalf("expired stacks/expiry = %d/%d, want 0/0", source.Passive.ExplorerSpellForceStacks, source.Passive.ExplorerSpellForceExpiresAt)
	}
}
