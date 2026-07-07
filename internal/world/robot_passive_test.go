package world

import "testing"

func TestRobotManaBarrierConsumesManaAndRefundsRemainingShield(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	placeEntity(robot, 3000, 3000)
	robot.Stats.HP = 130
	robot.Stats.MP = robot.Stats.MaxMP
	robot.Stats.MPRegen5 = 0
	robot.Combat.LastHitTick = 10

	w.applyDamage(nil, robot, 20, 20)

	if got, want := robot.Passive.Shield, 80; got != want {
		t.Fatalf("shield = %v, want %v", got, want)
	}
	if got, want := robot.Stats.MP, robot.Stats.MaxMP-80; got != want {
		t.Fatalf("mp after trigger = %f, want %f", got, want)
	}
	assertBuff(t, w.ActiveBuffs(robot, 10), "robot_mana_barrier")

	robot.Combat.LastHitTick = 20
	w.applyDamage(nil, robot, 30, 20)
	if got, want := robot.Passive.Shield, 50; got != want {
		t.Fatalf("shield after damage = %v, want %v", got, want)
	}
	robot.Stats.MPRegen5 = 0

	w.Tick(210, 20)
	if got := robot.Passive.Shield; got != 0 {
		t.Fatalf("shield after expire = %v, want 0", got)
	}
	if got, want := robot.Stats.MP, robot.Stats.MaxMP-30; got != want {
		t.Fatalf("mp after refund = %f, want %f", got, want)
	}

	robot.Combat.LastHitTick = 211
	w.applyDamage(nil, robot, 1, 20)
	if got := robot.Passive.Shield; got != 0 {
		t.Fatalf("shield during cooldown = %v, want 0", got)
	}
}
