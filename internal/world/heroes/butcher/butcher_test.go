package butcher

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"math"
	"strings"
	"testing"
)

func TestMeatHookDamagesAndPullsEnemy(t *testing.T) {
	w, butcher := testWorld(t)
	prepareQ(butcher, 1)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1500, 1000)
	target := w.EntityByID(targetID)
	startHP := target.Stats.HP

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 2300, TargetY: 1000}}, 10, nil, 20)
	if butcher.Stats.MP != 75 {
		t.Fatalf("mana = %v, want 75", butcher.Stats.MP)
	}
	if got := butcher.Skills[qID].CooldownUntilTick; got != 430 {
		t.Fatalf("cooldown = %d, want 430", got)
	}

	sawPull := false
	for tick := uint64(11); tick <= 40; tick++ {
		w.Tick(tick, 20)
		if target.Control.DashUntilTick > tick {
			sawPull = true
			if target.Control.AirborneUntilTick < target.Control.DashUntilTick {
				t.Fatalf("airborne until %d, pull until %d", target.Control.AirborneUntilTick, target.Control.DashUntilTick)
			}
		}
	}
	if !sawPull {
		t.Fatal("hook did not start pulling the target")
	}
	if got := startHP - target.Stats.HP; got != 257 {
		t.Fatalf("damage = %v, want 257", got)
	}
	if target.Position.X > 1042 || target.Position.X < 1040 {
		t.Fatalf("pulled x = %v, want about 1041", target.Position.X)
	}
}

func TestMeatHookPullsAllyWithoutDamageAndHalvesCooldown(t *testing.T) {
	w, butcher := testWorld(t)
	prepareQ(butcher, 1)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	allyID, _ := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamBlue, 1500, 1000)
	ally := w.EntityByID(allyID)
	startHP := ally.Stats.HP

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 2300, TargetY: 1000}}, 10, nil, 20)
	for tick := uint64(11); tick <= 40; tick++ {
		w.Tick(tick, 20)
	}
	if ally.Stats.HP != startHP {
		t.Fatalf("ally hp = %v, want %v", ally.Stats.HP, startHP)
	}
	if got := butcher.Skills[qID].CooldownUntilTick; got != 220 {
		t.Fatalf("ally hook cooldown = %d, want 220", got)
	}
	if ally.Position.X > 1044 {
		t.Fatalf("ally was not pulled to butcher: x=%v", ally.Position.X)
	}
}

func TestMeatHookKillsFirstEnemyMinion(t *testing.T) {
	w, butcher := testWorld(t)
	prepareQ(butcher, 1)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	minionID, _ := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamRed, 1300, 1000)
	heroID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1600, 1000)
	hero := w.EntityByID(heroID)
	startHP := hero.Stats.HP

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: qID, TargetX: 2300, TargetY: 1000}}, 10, nil, 20)
	for tick := uint64(11); tick <= 30; tick++ {
		w.Tick(tick, 20)
	}
	if w.EntityByID(minionID) != nil {
		t.Fatal("first enemy minion should be killed and removed")
	}
	if hero.Stats.HP != startHP {
		t.Fatalf("hero behind minion took damage: %v -> %v", startHP, hero.Stats.HP)
	}
}

func TestRotToggleDamageDebuffsAndLinger(t *testing.T) {
	w, butcher := testWorld(t)
	prepareW(butcher, 1)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1100, 1000)
	target := w.EntityByID(targetID)
	target.Stats.MagicDefense = 0
	startHP := target.Stats.HP

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID}}, 10, nil, 20)
	if !butcher.Passive.ButcherWActive || butcher.Skills[wID].CooldownUntilTick != 20 {
		t.Fatalf("rot activation = active:%v cooldown:%d", butcher.Passive.ButcherWActive, butcher.Skills[wID].CooldownUntilTick)
	}
	if buffs := w.ActiveBuffs(butcher, 10); len(buffs) != 1 || buffs[0].ID != wID || buffs[0].ExpiresAtTick != 0 {
		t.Fatalf("rot buff = %+v", buffs)
	}
	for tick := uint64(11); tick <= 20; tick++ {
		w.Tick(tick, 20)
	}
	if got := startHP - target.Stats.HP; got != 15 {
		t.Fatalf("rot half-second damage = %v, want 15", got)
	}
	if butcher.Stats.MP < 196 || butcher.Stats.MP >= 196.1 {
		t.Fatalf("mana = %v, want 4 spent plus normal regen", butcher.Stats.MP)
	}
	if math.Abs(target.Control.MoveSpeedSlow-0.14) > 0.000001 || target.Control.MoveSpeedSlowUntil != 30 {
		t.Fatalf("slow = %v until %d, want 0.14 until 30", target.Control.MoveSpeedSlow, target.Control.MoveSpeedSlowUntil)
	}
	if target.Control.GrievousWounds != 0.4 || target.Control.GrievousWoundsUntil != 30 {
		t.Fatalf("grievous wounds = %v until %d, want 0.4 until 30", target.Control.GrievousWounds, target.Control.GrievousWoundsUntil)
	}

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID}}, 20, nil, 20)
	if butcher.Passive.ButcherWActive || butcher.Skills[wID].CooldownUntilTick != 30 {
		t.Fatalf("rot deactivation = active:%v cooldown:%d", butcher.Passive.ButcherWActive, butcher.Skills[wID].CooldownUntilTick)
	}
	if got := world.EffectiveMoveSpeedAtTick(target, 29); got != target.Stats.MoveSpeed*0.86 {
		t.Fatalf("lingering move speed = %v", got)
	}
	w.Tick(30, 20)
	if got := world.EffectiveMoveSpeedAtTick(target, 30); got != target.Stats.MoveSpeed || target.Stats.GrievousWounds != 0 {
		t.Fatalf("expired aura move speed/grievous = %v/%v", got, target.Stats.GrievousWounds)
	}
}

func TestRotSelfDamageIsNonlethal(t *testing.T) {
	w, butcher := testWorld(t)
	prepareW(butcher, 1)
	butcher.Stats.HP = 5
	butcher.Stats.MagicDefense = 0
	applyWPeriodicDamage(w, butcher, w.SkillConfig(wID), 1, 20, 20, 0.5)
	if butcher.Stats.HP != 1 {
		t.Fatalf("self hp = %v, want 1", butcher.Stats.HP)
	}
}

func TestRotTurnsOffWhenManaIsInsufficient(t *testing.T) {
	w, butcher := testWorld(t)
	prepareW(butcher, 1)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	butcher.Stats.MP = 3
	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: wID}}, 10, nil, 20)
	for tick := uint64(11); tick <= 20; tick++ {
		w.Tick(tick, 20)
	}
	if butcher.Passive.ButcherWActive {
		t.Fatal("rot should turn off when it cannot pay the next mana tick")
	}
	if butcher.Skills[wID].CooldownUntilTick != 30 {
		t.Fatalf("automatic shutdown cooldown = %d, want 30", butcher.Skills[wID].CooldownUntilTick)
	}
}

func TestMeatShieldBlocksDamageAndExpires(t *testing.T) {
	w, butcher := testWorld(t)
	prepareE(butcher, 1)
	butcher.Stats.PhysicalDefense = 12
	butcher.Stats.MagicDefense = 30

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID}}, 10, nil, 20)
	if butcher.Stats.MP != 165 {
		t.Fatalf("mana = %v, want 165", butcher.Stats.MP)
	}
	if got := butcher.Skills[eID].CooldownUntilTick; got != 410 {
		t.Fatalf("cooldown = %d, want 410", got)
	}
	if got := butcher.Passive.ButcherEUntil; got != 90 {
		t.Fatalf("duration = %d, want expiry at 90", got)
	}
	wantBlock := 8 + 35 + 12*0.05 + 30*0.05
	if got := DamageBlock(w, butcher); math.Abs(got-wantBlock) > 0.0001 {
		t.Fatalf("damage block = %v, want %v", got, wantBlock)
	}
	buffs := w.ActiveBuffs(butcher, 10)
	if len(buffs) != 1 || buffs[0].ID != eID || buffs[0].ExpiresAtTick != 90 {
		t.Fatalf("meat shield buff = %+v", buffs)
	}

	w.Tick(90, 20)
	if butcher.Passive.ButcherEUntil != 0 || butcher.Passive.ButcherEEffectID != "" {
		t.Fatalf("expired meat shield state = %+v", butcher.Passive)
	}
	if got := DamageBlock(w, butcher); math.Abs(got-8) > 0.0001 {
		t.Fatalf("expired damage block = %v, want passive 8", got)
	}
}

func TestMeatShieldLevelFiveValues(t *testing.T) {
	w, butcher := testWorld(t)
	prepareE(butcher, 5)
	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: eID}}, 10, nil, 20)
	if got := butcher.Skills[eID].CooldownUntilTick; got != 330 {
		t.Fatalf("cooldown = %d, want 330", got)
	}
	if got := butcher.Passive.ButcherEUntil; got != 170 {
		t.Fatalf("duration = %d, want expiry at 170", got)
	}
	wantBlock := 8 + 95 + butcher.Stats.PhysicalDefense*0.05 + butcher.Stats.MagicDefense*0.05
	if got := DamageBlock(w, butcher); math.Abs(got-wantBlock) > 0.0001 {
		t.Fatalf("damage block = %v, want %v", got, wantBlock)
	}
}

func TestDismemberHeroDamageStunAndDuration(t *testing.T) {
	w, butcher := testWorld(t)
	prepareR(butcher, 1)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	butcher.Stats.BonusAttack = 20
	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1120, 1000)
	target := w.EntityByID(targetID)
	target.Stats.MagicDefense = 0
	startHP := target.Stats.HP

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID, TargetID: target.ID}}, 10, nil, 20)
	if butcher.Stats.MP != 100 || butcher.Skills[rID].CooldownUntilTick != 610 {
		t.Fatalf("r cost/cooldown = %v/%d, want 100/610", butcher.Stats.MP, butcher.Skills[rID].CooldownUntilTick)
	}
	if butcher.Passive.ButcherRUntil != 69 || target.Control.StunnedUntilTick != 69 {
		t.Fatalf("r hero duration/stun = %d/%d, want 69/69", butcher.Passive.ButcherRUntil, target.Control.StunnedUntilTick)
	}
	if got := startHP - target.Stats.HP; got != 90 {
		t.Fatalf("initial r damage = %v, want 90", got)
	}
	for tick := uint64(11); tick <= 69; tick++ {
		w.Tick(tick, 20)
	}
	if got := startHP - target.Stats.HP; got != 270 {
		t.Fatalf("total hero r damage = %v, want 270", got)
	}
	if butcher.Passive.ButcherRUntil != 0 || target.Control.StunnedUntilTick != 69 {
		t.Fatalf("r did not finish cleanly: until=%d stun=%d", butcher.Passive.ButcherRUntil, target.Control.StunnedUntilTick)
	}
}

func TestDismemberMinionLastsSixDamageTicks(t *testing.T) {
	w, butcher := testWorld(t)
	prepareR(butcher, 3)
	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	targetID, _ := w.SpawnObject(world.EntityKindMeleeMinion, world.TeamRed, 1120, 1000)
	target := w.EntityByID(targetID)
	target.Stats.HP = 5000
	target.Stats.MaxHP = 5000
	target.Stats.MagicDefense = 0

	w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID, TargetID: target.ID}}, 10, nil, 20)
	if butcher.Passive.ButcherRUntil != 128 {
		t.Fatalf("r minion expiry = %d, want 128", butcher.Passive.ButcherRUntil)
	}
	for tick := uint64(11); tick <= 128; tick++ {
		w.Tick(tick, 20)
	}
	if got := 5000 - target.Stats.HP; got != 1050 {
		t.Fatalf("total minion r damage = %v, want 1050", got)
	}
}

func TestDismemberIsInterruptedByMovementAndControl(t *testing.T) {
	for _, test := range []struct {
		name      string
		interrupt func(*world.World, *world.Entity)
	}{
		{name: "movement", interrupt: func(w *world.World, butcher *world.Entity) {
			w.ApplyInput("butcher", protocol.PlayerInput{Move: &protocol.MoveInput{TargetX: 1400, TargetY: 1000}}, 11, nil, 20)
		}},
		{name: "stun", interrupt: func(_ *world.World, butcher *world.Entity) {
			butcher.Control.StunnedUntilTick = 30
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			w, butcher := testWorld(t)
			prepareR(butcher, 1)
			butcher.Position = world.Vector2{X: 1000, Y: 1000}
			targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1120, 1000)
			target := w.EntityByID(targetID)
			w.ApplyInput("butcher", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID, TargetID: target.ID}}, 10, nil, 20)
			test.interrupt(w, butcher)
			w.Tick(12, 20)
			if butcher.Passive.ButcherRUntil != 0 || target.Control.StunnedUntilTick > 12 {
				t.Fatalf("r remained after %s: until=%d stun=%d", test.name, butcher.Passive.ButcherRUntil, target.Control.StunnedUntilTick)
			}
		})
	}
}

func TestPassiveDamageReductionAndFlesh(t *testing.T) {
	w, butcher := testWorld(t)
	sourceID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 100, 100)
	if !ok {
		t.Fatal("spawn enemy hero")
	}
	source := w.EntityByID(sourceID)
	butcher.Position = world.Vector2{X: 100, Y: 100}

	startHP := butcher.Stats.HP
	w.ApplyTrueDamage(source, butcher, 100, 20)
	if got := startHP - butcher.Stats.HP; got != 92 {
		t.Fatalf("level 1 true damage taken = %v, want 92", got)
	}

	butcher.Stats.HP = butcher.Stats.MaxHP
	startHP = butcher.Stats.HP
	damage := w.MagicDamageAfterResistance(source, butcher, 130, 0)
	w.ApplyMagicDamage(source, butcher, damage, 20)
	if got := startHP - butcher.Stats.HP; got != 88 {
		t.Fatalf("level 1 magic damage taken = %v, want 88", got)
	}

	baseMaxHP := butcher.Stats.MaxHP
	baseRegen := butcher.Stats.HPRegen5
	baseAttack := butcher.Stats.Attack
	w.ApplyKillReward(nil, source)
	if butcher.Passive.ButcherFlesh != 3 {
		t.Fatalf("nearby flesh = %d, want 3", butcher.Passive.ButcherFlesh)
	}
	if math.Abs(butcher.Stats.MaxHP-(baseMaxHP+0.66)) > 0.0001 ||
		math.Abs(butcher.Stats.HPRegen5-(baseRegen+0.015)) > 0.0001 ||
		math.Abs(butcher.Stats.Attack-(baseAttack+0.03)) > 0.0001 {
		t.Fatalf("flesh stats = %+v", butcher.Stats)
	}

	butcher.Position = world.Vector2{X: 1000, Y: 1000}
	w.ApplyKillReward(nil, source)
	if butcher.Passive.ButcherFlesh != 3 {
		t.Fatalf("flesh outside range = %d, want 3", butcher.Passive.ButcherFlesh)
	}
	w.ApplyKillReward(butcher, source)
	if butcher.Passive.ButcherFlesh != 6 {
		t.Fatalf("killer flesh outside range = %d, want 6", butcher.Passive.ButcherFlesh)
	}
	buffs := w.ActiveBuffs(butcher, 0)
	if len(buffs) != 1 || buffs[0].Stacks != 6 || buffs[0].ExpiresAtTick != 0 || !strings.Contains(buffs[0].Tooltip, "最大生命 +1.32") {
		t.Fatalf("flesh buff = %+v", buffs)
	}
}

func TestPassiveScalesAtLevel18(t *testing.T) {
	w, butcher := testWorld(t)
	butcher.Level = 18
	w.RefreshPlayerStats(butcher)
	if got := DamageBlock(w, butcher); math.Abs(got-26) > 0.0001 {
		t.Fatalf("damage block = %v, want 26", got)
	}
	if got := MagicDamageReduction(w, butcher, 0); math.Abs(got-0.16) > 0.0001 {
		t.Fatalf("magic reduction = %v, want 0.16", got)
	}

	targetID, _ := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, butcher.Position.X, butcher.Position.Y)
	w.ApplyKillReward(nil, w.EntityByID(targetID))
	if butcher.Passive.ButcherFlesh != 12 {
		t.Fatalf("level 18 flesh = %d, want 12", butcher.Passive.ButcherFlesh)
	}
}

func testWorld(t *testing.T) (*world.World, *world.Entity) {
	t.Helper()
	heroes, err := config.LoadHeroes("../../../../configs/heroes")
	if err != nil {
		t.Fatal(err)
	}
	skills, err := config.LoadSkills("../../../../configs/skills")
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(heroes, skills, nil, nil, nil)
	hero, ok := heroes.Get(heroID)
	if !ok {
		t.Fatal("butcher config missing")
	}
	w.SpawnHero("butcher", hero, world.TeamBlue)
	return w, w.EntityByID("player:butcher")
}

func prepareQ(entity *world.Entity, level int) {
	state := entity.Skills[qID]
	state.Level = level
	entity.Skills[qID] = state
}

func prepareW(entity *world.Entity, level int) {
	state := entity.Skills[wID]
	state.Level = level
	entity.Skills[wID] = state
}

func prepareE(entity *world.Entity, level int) {
	state := entity.Skills[eID]
	state.Level = level
	entity.Skills[eID] = state
}

func prepareR(entity *world.Entity, level int) {
	state := entity.Skills[rID]
	state.Level = level
	entity.Skills[rID] = state
}
