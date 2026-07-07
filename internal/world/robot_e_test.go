package world

import "testing"

func TestRobotEEmpowersNextAttackResetsAndKnocksUp(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	target := w.entities["enemy:hero-1"]
	placeEntity(robot, 1000, 1000)
	placeEntity(target, 1100, 1000)
	robot.Stats.Attack = 100
	robot.Stats.AttackSpeed = 1
	robot.Stats.AbilityPower = 80
	robot.Stats.MP = 100
	robot.Stats.MPRegen5 = 0
	target.Stats.HP = 2000
	target.Stats.MaxHP = 2000
	target.Stats.PhysicalDefense = 0
	learnSkill(robot, "robot_e", 1)

	w.ApplyInput("robot", protocolPlayerInputAttack(target.ID), 10, nil, 20)
	w.Tick(10, 20)
	tickAttackRelease(t, w, robot, 20)
	if got, want := robot.Combat.NextAttackTick, uint64(30); got != want {
		t.Fatalf("next attack after first hit = %v, want %v", got, want)
	}

	placeEntity(target, 1280, 1000)
	target.Control.DashUntilTick = 100
	target.Control.DashStartTick = 10
	target.Control.DashStart = target.Position
	target.Control.DashEnd = target.Position
	target.Combat.PendingAttackTargetID = robot.ID
	target.Combat.AttackReleaseTick = 80
	w.ApplyInput("robot", protocolPlayerInputCast("robot_e", target.Position.X, target.Position.Y), 16, nil, 20)
	if got := robot.Stats.MP; got != 75 {
		t.Fatalf("mp after e = %f, want 75", got)
	}
	if got, want := robot.Skills["robot_e"].CooldownUntilTick, uint64(196); got != want {
		t.Fatalf("e cooldown = %v, want %v", got, want)
	}
	if got, want := robot.Stats.AttackRange, 300.0; got != want {
		t.Fatalf("e attack range = %f, want %f", got, want)
	}
	robot.Stats.AbilityPower = 80
	wantDamage := w.PhysicalDamageAfterResistance(robot, target, robot.Stats.Attack*2+float64(robot.Stats.AbilityPower)*0.25, 21)
	w.Tick(16, 20)
	if got := robot.Combat.PendingAttackTargetID; got != target.ID {
		t.Fatalf("pending attack after e reset = %q, want %q", got, target.ID)
	}
	tickAttackRelease(t, w, robot, 20)

	if got := target.Combat.LastDamage; got != wantDamage {
		t.Fatalf("e damage = %d, want %d", got, wantDamage)
	}
	if got, want := target.Control.AirborneUntilTick, uint64(41); got != want {
		t.Fatalf("e knockup until = %v, want %v", got, want)
	}
	if target.Control.DashUntilTick != 0 || target.Combat.PendingAttackTargetID != "" || target.Combat.AttackReleaseTick != 0 {
		t.Fatalf("target was not interrupted: control=%+v combat=%+v", target.Control, target.Combat)
	}
	if robot.Skills["robot_e"].Stacks != 0 {
		t.Fatal("e should be consumed by next basic attack")
	}
	if got, want := robot.Stats.AttackRange, 125.0; got != want {
		t.Fatalf("attack range after e consume = %f, want %f", got, want)
	}
}

func TestRobotEExpiresUnused(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	robot.Stats.MP = 100
	robot.Stats.MPRegen5 = 0
	learnSkill(robot, "robot_e", 1)

	w.ApplyInput("robot", protocolPlayerInputCast("robot_e", robot.Position.X, robot.Position.Y), 10, nil, 20)
	assertBuff(t, w.ActiveBuffs(robot, 10), "robot_power_fist")
	w.Tick(110, 20)

	if got := robot.Skills["robot_e"].Stacks; got != 0 {
		t.Fatalf("e stacks after expiry = %d, want 0", got)
	}
	if got, want := robot.Stats.AttackRange, 125.0; got != want {
		t.Fatalf("attack range after e expiry = %f, want %f", got, want)
	}
}
