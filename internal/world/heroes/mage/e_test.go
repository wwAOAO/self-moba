package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/world"
	"testing"
)

func TestAddEBurstEffectUsesDetonationPositionAndRadius(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	entity := &world.Entity{ID: "mage", HeroID: heroID, Team: world.TeamBlue}
	center := world.Vector2{X: 1200, Y: 900}

	addEBurstEffect(w, entity, center, config.SkillConfig{Meta: map[string]float64{"radius": 310}}, 20, 20)

	var found *world.SkillEffect
	for _, effect := range w.SkillEffects() {
		if effect.Kind == "mage_lucent_singularity_burst" {
			copy := effect
			found = &copy
			break
		}
	}
	if found == nil {
		t.Fatal("missing lucent singularity burst effect")
	}
	if found.Start != center {
		t.Fatalf("burst center = %+v, want %+v", found.Start, center)
	}
	if found.Radius != 310 {
		t.Fatalf("burst radius = %v, want 310", found.Radius)
	}
	if found.CreatedAt != 20 || found.ExpiresAt != 36 {
		t.Fatalf("burst ticks = %d-%d, want 20-36", found.CreatedAt, found.ExpiresAt)
	}
}
