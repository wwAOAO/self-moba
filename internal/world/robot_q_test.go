package world

import "testing"

func TestRobotQPullsFirstValidTarget(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	placeEntity(robot, 1000, 3000)
	robot.Stats.MP = 200
	robot.Stats.AbilityPower = 20
	learnSkill(robot, robotQSkillID, 1)

	tower := &Entity{ID: "structure:q-tower", Kind: EntityKindTower, Team: TeamRed, Position: Vector2{X: 1200, Y: 3000}, Radius: 80, Stats: Stats{HP: 1000, MaxHP: 1000}}
	near := &Entity{ID: "minion:q-near", Kind: EntityKindMeleeMinion, Team: TeamRed, Position: Vector2{X: 1360, Y: 3000}, Radius: 20, Stats: Stats{HP: 1000, MaxHP: 1000}}
	far := &Entity{ID: "minion:q-far", Kind: EntityKindMeleeMinion, Team: TeamRed, Position: Vector2{X: 1370, Y: 3000}, Radius: 20, Stats: Stats{HP: 1000, MaxHP: 1000}}
	w.entities[tower.ID] = tower
	w.entities[near.ID] = near
	w.entities[far.ID] = far

	w.ApplyInput("robot", protocolPlayerInputCast(robotQSkillID, 2000, 3000), 10, nil, 20)
	if got := robot.Stats.MP; got != 100 {
		t.Fatalf("mp after q cast = %f, want 100", got)
	}
	w.Tick(15, 20)
	if got, want := robot.Skills[robotQSkillID].CooldownUntilTick, uint64(415); got != want {
		t.Fatalf("q cooldown = %v, want %v", got, want)
	}

	for tick := uint64(16); tick <= 22; tick++ {
		w.Tick(tick, 20)
	}

	if got := near.Stats.HP; got != 900 {
		t.Fatalf("near hp = %v, want 900", got)
	}
	if got := far.Stats.HP; got != 1000 {
		t.Fatalf("far hp = %v, want 1000", got)
	}
	if got, want := near.Position.X, 1039.0; got != want {
		t.Fatalf("near x after pull = %f, want %f", got, want)
	}
	if near.Control.StunnedUntilTick != 22 {
		t.Fatalf("stun until = %v, want 22", near.Control.StunnedUntilTick)
	}
}

func TestRobotQBlockedByWindWall(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	placeEntity(robot, 1000, 3000)
	robot.Stats.MP = 200
	learnSkill(robot, robotQSkillID, 1)
	target := &Entity{ID: "minion:q-blocked", Kind: EntityKindMeleeMinion, Team: TeamRed, Position: Vector2{X: 1360, Y: 3000}, Radius: 20, Stats: Stats{HP: 1000, MaxHP: 1000}}
	w.entities[target.ID] = target
	w.PutWindWall(WindWall{ID: "windwall:test", Team: TeamRed, Center: Vector2{X: 1180, Y: 3000}, Dir: Vector2{X: 0, Y: 1}, Width: 500, ExpiresAt: 100})

	w.ApplyInput("robot", protocolPlayerInputCast(robotQSkillID, 2000, 3000), 10, nil, 20)
	w.Tick(15, 20)
	w.Tick(16, 20)
	w.Tick(17, 20)

	if got := countProjectilesByKind(w, "robot_q"); got != 0 {
		t.Fatalf("robot q projectiles = %d, want 0", got)
	}
	if got := target.Stats.HP; got != 1000 {
		t.Fatalf("target hp = %v, want 1000", got)
	}
}
