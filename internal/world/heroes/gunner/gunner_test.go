package gunner

import (
	"testing"

	"l-battle/internal/config"
	"l-battle/internal/world"
)

func TestGunnerCastEffectsCarryTheirVisualTiming(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	entity := &world.Entity{ID: "gunner", HeroID: heroID, Team: world.TeamBlue}
	dir := world.Vector2{X: 1}

	addQMuzzleEffect(w, entity, dir, 2, config.SkillConfig{}, 20, 20)
	addWEffect(w, entity, 3, config.SkillConfig{}, 30, 110)

	var muzzle, active *world.SkillEffect
	for _, effect := range w.SkillEffects() {
		effect := effect
		switch effect.Kind {
		case "gunner_q_muzzle":
			muzzle = &effect
		case "gunner_w":
			active = &effect
		}
	}
	if muzzle == nil || muzzle.Dir != dir || muzzle.Range != 150 || muzzle.Width != 44 {
		t.Fatalf("q muzzle effect = %+v", muzzle)
	}
	if muzzle.CreatedAt != 20 || muzzle.ExpiresAt != 25 || muzzle.Count != 2 {
		t.Fatalf("q muzzle timing = %+v", muzzle)
	}
	if active == nil || active.SourceID != entity.ID || active.Radius != 105 {
		t.Fatalf("w active effect = %+v", active)
	}
	if active.CreatedAt != 30 || active.ExpiresAt != 110 || active.Count != 3 {
		t.Fatalf("w active timing = %+v", active)
	}
}
