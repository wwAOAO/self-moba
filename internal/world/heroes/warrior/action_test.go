package warrior

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestWarriorCastsExposeAnimationWindows(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	entity := &world.Entity{
		ID:     "warrior",
		HeroID: heroID,
		Team:   world.TeamBlue,
		Stats:  world.Stats{HP: 1000, MaxHP: 1000},
		Skills: make(map[string]world.SkillState),
	}
	tests := []struct {
		name       string
		skillID    string
		wantEndsAt uint64
		apply      func()
	}{
		{name: "q", skillID: qID, wantEndsAt: 18, apply: func() {
			ApplyQ(w, entity, world.SkillState{SkillID: qID, Level: 1}, config.SkillConfig{}, 10, 20)
		}},
		{name: "w", skillID: wID, wantEndsAt: 19, apply: func() {
			ApplyW(w, entity, world.SkillState{SkillID: wID, Level: 1}, config.SkillConfig{}, 10, 20)
		}},
		{name: "e", skillID: eID, wantEndsAt: 70, apply: func() {
			ApplyE(w, entity, world.SkillState{SkillID: eID, Level: 1}, config.SkillConfig{}, 10, 20)
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.apply()
			if entity.Action.Name != test.name || entity.Action.SkillID != test.skillID || entity.Action.StartedAtTick != 10 || entity.Action.EndsAtTick != test.wantEndsAt {
				t.Fatalf("warrior %s action = %+v", test.name, entity.Action)
			}
		})
	}
}

func TestWarriorRActionFacesTargetAndEStopClearsAction(t *testing.T) {
	w := world.NewWorld(nil, nil, nil, nil, nil)
	var target *world.Entity
	w.ForEachEntity(func(candidate *world.Entity) {
		if target == nil && candidate.Team == world.TeamRed && candidate.Stats.HP > 0 {
			target = candidate
		}
	})
	if target == nil {
		t.Fatal("missing red target")
	}
	entity := &world.Entity{
		ID:       "warrior",
		HeroID:   heroID,
		Team:     world.TeamBlue,
		Position: world.Vector2{X: target.Position.X - 100, Y: target.Position.Y},
		Stats:    world.Stats{HP: 1000, MaxHP: 1000},
		Skills:   make(map[string]world.SkillState),
	}
	ApplyR(
		w,
		entity,
		protocol.CastInput{TargetX: target.Position.X, TargetY: target.Position.Y},
		world.SkillState{SkillID: rID, Level: 1},
		config.SkillConfig{Range: 400},
		10,
		20,
	)
	if entity.Action.Name != "r" || entity.Action.EndsAtTick != 19 || entity.Facing.X <= 0 || entity.Facing.Y != 0 {
		t.Fatalf("warrior r action = %+v, facing = %+v", entity.Action, entity.Facing)
	}

	ApplyE(w, entity, world.SkillState{SkillID: eID, Level: 1}, config.SkillConfig{}, 20, 20)
	StopE(w, entity, world.SkillState{SkillID: eID, Level: 1}, config.SkillConfig{}, 21, 20)
	if entity.Action != (world.ActionState{}) {
		t.Fatalf("warrior e action after stop = %+v", entity.Action)
	}
}
