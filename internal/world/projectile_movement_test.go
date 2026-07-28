package world

import "testing"

func TestShadowAssassinBladesReturnImmediatelyAtWorldBoundary(t *testing.T) {
	for _, skillID := range []string{shadowAssassinWSkillID, shadowAssassinRSkillID} {
		t.Run(skillID, func(t *testing.T) {
			source := &Entity{ID: "source", Radius: 18, Position: Vector2{X: 100, Y: 100}}
			projectile := &Projectile{
				ID:           "blade",
				SourceID:     source.ID,
				SkillID:      skillID,
				Position:     Vector2{X: 5, Y: 50},
				Dir:          Vector2{X: -1},
				SpeedPerTick: 10,
				Range:        1000,
			}
			w := projectileTestWorld(source, projectile)

			previous := w.moveProjectile(projectile, 20)
			w.handleProjectileAfterMove(projectile.ID, source, projectile, previous, 12, 20)

			if !projectile.Returning || projectile.EffectTicks != 12 {
				t.Fatalf("boundary return state = returning:%v tick:%d", projectile.Returning, projectile.EffectTicks)
			}
			if projectile.Dir.X <= 0 {
				t.Fatalf("boundary return direction = %+v, want direction toward source", projectile.Dir)
			}
		})
	}
}

func TestShadowAssassinWReturnTracksCurrentSourcePosition(t *testing.T) {
	source := &Entity{ID: "source", Radius: 18, Position: Vector2{X: 100, Y: 200}}
	projectile := &Projectile{
		ID:           "blade",
		SourceID:     source.ID,
		SkillID:      shadowAssassinWSkillID,
		Position:     Vector2{X: 100, Y: 100},
		Start:        Vector2{X: 0, Y: 100},
		Dir:          Vector2{X: -1},
		SpeedPerTick: 20,
		Range:        1000,
		Returning:    true,
	}
	w := projectileTestWorld(source, projectile)

	w.moveProjectile(projectile, 20)

	if projectile.Dir.X != 0 || projectile.Dir.Y != 1 {
		t.Fatalf("return direction = %+v, want direction toward current source", projectile.Dir)
	}
	if projectile.Position != (Vector2{X: 100, Y: 120}) {
		t.Fatalf("return position = %+v, want {100 120}", projectile.Position)
	}
}

func TestShadowAssassinBladesAreCollectedAtSource(t *testing.T) {
	for _, skillID := range []string{shadowAssassinWSkillID, shadowAssassinRSkillID} {
		t.Run(skillID, func(t *testing.T) {
			source := &Entity{ID: "source", Radius: 18, Position: Vector2{X: 100, Y: 100}}
			projectile := &Projectile{
				ID:        "blade",
				SourceID:  source.ID,
				SkillID:   skillID,
				Position:  Vector2{X: 117, Y: 100},
				Returning: true,
			}
			w := projectileTestWorld(source, projectile)

			w.finishProjectileIfNeeded(projectile.ID, source, projectile, 12, 20)

			if w.ProjectileByID(projectile.ID) != nil {
				t.Fatal("returning blade was not collected inside source radius")
			}
		})
	}
}

func TestShadowAssassinRStopsClientExtrapolationAtMaxRange(t *testing.T) {
	source := &Entity{ID: "source", Radius: 18, Position: Vector2{X: 100, Y: 100}}
	projectile := &Projectile{
		ID:           "blade",
		Kind:         "shadow_assassin_r",
		SourceID:     source.ID,
		SkillID:      shadowAssassinRSkillID,
		Position:     Vector2{X: 650, Y: 100},
		SpeedPerTick: 60,
		Range:        550,
		Traveled:     550,
	}
	w := projectileTestWorld(source, projectile)

	effects := w.SkillEffects()
	if len(effects) != 1 || effects[0].Speed != 0 {
		t.Fatalf("outbound endpoint snapshot speed = %+v, want 0", effects)
	}
	if projectile.SpeedPerTick != 60 {
		t.Fatalf("runtime speed = %v, want 60", projectile.SpeedPerTick)
	}

	projectile.Returning = true
	effects = w.SkillEffects()
	if len(effects) != 1 || effects[0].Speed != 60 {
		t.Fatalf("return snapshot speed = %+v, want 60", effects)
	}
}

func projectileTestWorld(source *Entity, projectile *Projectile) *World {
	return &World{
		width:        1000,
		height:       1000,
		entities:     map[string]*Entity{source.ID: source},
		projectiles:  map[string]*Projectile{projectile.ID: projectile},
		skillEffects: make(map[string]SkillEffect),
	}
}
