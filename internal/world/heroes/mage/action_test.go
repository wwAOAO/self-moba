package mage

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestMageCastsExposeAnimationWindows(t *testing.T) {
	tests := []struct {
		name       string
		skillID    string
		windup     float64
		wantEndsAt uint64
		apply      func(*world.World, *world.Entity, protocol.CastInput, world.SkillState, config.SkillConfig, uint64, int)
	}{
		{name: "q", skillID: qID, windup: 0.25, wantEndsAt: 18, apply: ApplyQ},
		{name: "w", skillID: wID, windup: 0.2, wantEndsAt: 18, apply: ApplyW},
		{name: "e", skillID: eID, windup: 0.25, wantEndsAt: 18, apply: ApplyE},
		{name: "r", skillID: rID, windup: 0.5, wantEndsAt: 20, apply: ApplyR},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := world.NewWorld(nil, nil, nil, nil, nil)
			entity := &world.Entity{
				ID:       "mage",
				HeroID:   heroID,
				Position: world.Vector2{X: 1000, Y: 1000},
				Stats:    world.Stats{MP: 1000},
				Skills:   make(map[string]world.SkillState),
			}
			test.apply(
				w,
				entity,
				protocol.CastInput{TargetX: 1200, TargetY: 1000},
				world.SkillState{SkillID: test.skillID, Level: 1},
				config.SkillConfig{Meta: map[string]float64{"castWindupSeconds": test.windup}},
				10,
				20,
			)

			if entity.Action.Name != test.name || entity.Action.SkillID != test.skillID || entity.Action.StartedAtTick != 10 || entity.Action.EndsAtTick != test.wantEndsAt {
				t.Fatalf("mage %s action = %+v", test.name, entity.Action)
			}
			if entity.Facing.X <= 0 || entity.Facing.Y != 0 {
				t.Fatalf("mage %s facing = %+v, want right", test.name, entity.Facing)
			}
		})
	}
}
