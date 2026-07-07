package world

import "testing"

func TestRobotRPassiveMarksBasicAttacksAndTriggersArc(t *testing.T) {
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
	robot.Stats.AbilityPower = 100
	target.Stats.MagicDefense = 0
	target.Stats.HP = 2000
	target.Stats.MaxHP = 2000
	learnSkill(robot, "robot_r", 1)

	for tick := uint64(10); tick <= 12; tick++ {
		w.onHeroBasicHit(robot, target, tick, 20)
	}
	if got := target.Passive.RobotArcMarks[robot.ID].Stacks; got != 3 {
		t.Fatalf("arc stacks = %d, want 3", got)
	}

	w.Tick(29, 20)
	if target.Combat.LastDamage != 0 {
		t.Fatalf("arc damaged early = %d", target.Combat.LastDamage)
	}
	w.Tick(30, 20)

	if got := target.Combat.LastDamage; got != 240 {
		t.Fatalf("arc damage = %d, want 240", got)
	}
	if _, ok := target.Passive.RobotArcMarks[robot.ID]; ok {
		t.Fatal("arc mark should be consumed")
	}
}

func TestRobotRActiveDamagesSilencesRemovesShieldsAndDisablesPassive(t *testing.T) {
	w := testWorld(t)
	hero, ok := w.heroes.Get("robot")
	if !ok {
		t.Fatal("robot hero not found")
	}
	w.SpawnHero("robot", hero, TeamBlue)
	robot := w.entities[playerEntityID("robot")]
	target := w.entities["enemy:hero-1"]
	outsideID, _ := w.SpawnObject(EntityKindEnemyHero, TeamRed, 3000, 3000)
	outside := w.entities[outsideID]
	placeEntity(robot, 1000, 1000)
	placeEntity(target, 1300, 1000)
	robot.Stats.MP = 200
	robot.Stats.MPRegen5 = 0
	robot.Stats.AbilityPower = 50
	target.Stats.HP = 2000
	target.Stats.MaxHP = 2000
	target.Stats.MagicDefense = 0
	target.Passive.Shield = 150
	target.Passive.MaxShield = 150
	outside.Passive.Shield = 100
	outside.Passive.MaxShield = 100
	learnSkill(robot, "robot_r", 1)

	w.ApplyInput("robot", protocolPlayerInputCast("robot_r", robot.Position.X, robot.Position.Y), 10, nil, 20)
	if got := robot.Stats.MP; got != 100 {
		t.Fatalf("mp after r cast = %f, want 100", got)
	}
	if got := target.Combat.LastDamage; got != 0 {
		t.Fatalf("r damaged before windup = %d, want 0", got)
	}
	w.Tick(14, 20)

	if got := target.Combat.LastDamage; got != 300 {
		t.Fatalf("r active damage = %d, want 300", got)
	}
	assertSkillEffect(t, w.SkillEffects(), "robot_r")
	if target.Passive.Shield != 0 || target.Passive.MaxShield != 0 {
		t.Fatalf("target shield = %d/%d, want 0/0", target.Passive.Shield, target.Passive.MaxShield)
	}
	if got, want := target.Control.SilencedUntilTick, uint64(24); got != want {
		t.Fatalf("silence until = %d, want %d", got, want)
	}
	if outside.Passive.Shield != 100 || outside.Combat.LastDamage != 0 {
		t.Fatal("outside target should not be affected")
	}
	if got, want := robot.Skills["robot_r"].CooldownUntilTick, uint64(1214); got != want {
		t.Fatalf("r cooldown = %d, want %d", got, want)
	}

	w.onHeroBasicHit(robot, target, 15, 20)
	if len(target.Passive.RobotArcMarks) != 0 {
		t.Fatal("r passive should be disabled during cooldown")
	}
	robot.Skills["robot_r"] = SkillState{SkillID: "robot_r", Level: 1, CooldownUntilTick: 15}
	w.onHeroBasicHit(robot, target, 16, 20)
	if len(target.Passive.RobotArcMarks) == 0 {
		t.Fatal("r passive should resume after cooldown")
	}
}
