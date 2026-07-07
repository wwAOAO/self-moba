package world

import (
	"math"
	"testing"
)

func TestRobotWBuffDecaysAndEndsWithSlow(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	placeEntity(robot, 3000, 3000)
	robot.Stats.MP = 200
	robot.Stats.MPRegen5 = 0
	learnSkill(robot, "robot_w", 1)
	baseMoveSpeed := robot.Stats.MoveSpeed
	baseAttackSpeed := robot.Stats.AttackSpeed

	w.ApplyInput("robot", protocolPlayerInputCast("robot_w", robot.Position.X, robot.Position.Y), 10, nil, 20)

	if got := robot.Stats.MP; got != 125 {
		t.Fatalf("mp after w = %f, want 125", got)
	}
	if got, want := robot.Skills["robot_w"].CooldownUntilTick, uint64(310); got != want {
		t.Fatalf("w cooldown = %v, want %v", got, want)
	}
	if got, want := robot.Stats.MoveSpeed, baseMoveSpeed*1.7; math.Abs(got-want) > 0.000001 {
		t.Fatalf("move speed after w = %f, want %f", got, want)
	}
	if robot.Stats.AttackSpeed <= baseAttackSpeed {
		t.Fatalf("attack speed after w = %f, want above %f", robot.Stats.AttackSpeed, baseAttackSpeed)
	}
	assertBuff(t, w.ActiveBuffs(robot, 10), "robot_overdrive")

	w.Tick(60, 20)
	if got, want := robot.Stats.MoveSpeed, baseMoveSpeed*1.1; math.Abs(got-want) > 0.000001 {
		t.Fatalf("move speed after decay = %f, want %f", got, want)
	}

	w.Tick(110, 20)
	if got := robot.Passive.RobotWUntil; got != 0 {
		t.Fatalf("w active after expire = %v, want 0", got)
	}
	if got, want := EffectiveMoveSpeedAtTick(robot, 111), baseMoveSpeed*0.7; math.Abs(got-want) > 0.000001 {
		t.Fatalf("move speed during penalty = %f, want %f", got, want)
	}
}

func TestRobotWBasicAttackBonusMagicDamage(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	robot.Passive.RobotWUntil = 100
	robot.Passive.RobotWLevel = 1
	heroTarget := &Entity{ID: "enemy:robot-w-hero", Kind: EntityKindEnemyHero, Team: TeamRed, Stats: Stats{HP: 1000, MaxHP: 1000}, Radius: 18}
	minionTarget := &Entity{ID: "minion:robot-w-cap", Kind: EntityKindMeleeMinion, Team: TeamRed, Stats: Stats{HP: 100000, MaxHP: 100000}, Radius: 20}

	if got := w.heroBasicAttackBonusMagicDamage(robot, heroTarget, 10, 20); got != 10 {
		t.Fatalf("hero bonus damage = %d, want 10", got)
	}
	if got := w.heroBasicAttackBonusMagicDamage(robot, minionTarget, 10, 20); got != 60 {
		t.Fatalf("minion capped bonus damage = %d, want 60", got)
	}
	if got := w.heroBasicAttackBonusMagicDamage(robot, heroTarget, 100, 20); got != 0 {
		t.Fatalf("expired bonus damage = %d, want 0", got)
	}
}
