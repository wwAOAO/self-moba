package crossbowman

import (
	"testing"

	"l-battle/internal/config"
	"l-battle/internal/world"
)

func TestNightHunterActivatesTowardEnemyHeroAndLingers(t *testing.T) {
	w, source, target := passiveTestWorld(t)
	source.Intent.MoveTarget = &world.Vector2{X: 2000, Y: 1000}

	Tick(w, source, 10, 20)

	if got, want := source.Crossbowman.NightHunterUntil, uint64(50); got != want {
		t.Fatalf("night hunter expiry = %d, want %d", got, want)
	}
	if got, want := world.EffectiveMoveSpeedAtTick(source, 10), 360.0; got != want {
		t.Fatalf("move speed while pursuing = %v, want %v", got, want)
	}

	source.Intent.MoveTarget = &world.Vector2{X: 500, Y: 1000}
	Tick(w, source, 11, 20)
	if got := world.EffectiveMoveSpeedAtTick(source, 49); got != 360 {
		t.Fatalf("move speed during linger = %v, want 360", got)
	}
	if got := world.EffectiveMoveSpeedAtTick(source, 50); got != 330 {
		t.Fatalf("move speed after linger = %v, want 330", got)
	}

	target.Position = world.Vector2{X: 3001, Y: 1000}
	source.Crossbowman = world.CrossbowmanState{}
	source.Intent.MoveTarget = &world.Vector2{X: 4000, Y: 1000}
	Tick(w, source, 60, 20)
	if source.Crossbowman.NightHunterUntil != 0 {
		t.Fatal("night hunter activated outside its 2000 range")
	}
}

func TestNightHunterUsesUltimateMoveSpeedBonus(t *testing.T) {
	w, source, _ := passiveTestWorld(t)
	source.Intent.MoveTarget = &world.Vector2{X: 2000, Y: 1000}
	source.Crossbowman.UltimateUntilTick = 100

	Tick(w, source, 10, 20)

	if got, want := world.EffectiveMoveSpeedAtTick(source, 10), 420.0; got != want {
		t.Fatalf("move speed during ultimate = %v, want %v", got, want)
	}
	buffs := ActiveBuffs(w, source, 10)
	if len(buffs) != 1 || buffs[0].ID != passiveID || buffs[0].Tooltip != "+90 移动速度" {
		t.Fatalf("night hunter buff = %+v", buffs)
	}
}

func TestNightHunterRequiresMovementTowardEnemy(t *testing.T) {
	w, source, _ := passiveTestWorld(t)
	source.Intent.MoveTarget = &world.Vector2{X: 500, Y: 1000}

	Tick(w, source, 10, 20)

	if source.Crossbowman.NightHunterUntil != 0 {
		t.Fatal("night hunter activated while moving away from the enemy")
	}
}

func passiveTestWorld(t *testing.T) (*world.World, *world.Entity, *world.Entity) {
	t.Helper()
	skills, err := config.NewSkillStore([]config.SkillConfig{{
		SkillID: passiveID,
		Range:   2000,
		Meta: map[string]float64{
			"moveSpeedBonus":         30,
			"ultimateMoveSpeedBonus": 90,
			"lingerSeconds":          2,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	w := world.NewWorld(nil, skills, nil, nil, nil)
	source := &world.Entity{
		ID:       "crossbowman",
		HeroID:   heroID,
		Kind:     world.EntityKindPlayer,
		Team:     world.TeamBlue,
		Position: world.Vector2{X: 1000, Y: 1000},
		Stats:    world.Stats{HP: 515, MaxHP: 515, MoveSpeed: 330, AttackRange: 550},
	}
	targetID, ok := w.SpawnObject(world.EntityKindEnemyHero, world.TeamRed, 1500, 1000)
	if !ok {
		t.Fatal("failed to spawn enemy hero")
	}
	target := w.EntityByID(targetID)
	return w, source, target
}
