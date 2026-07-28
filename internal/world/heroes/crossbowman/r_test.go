package crossbowman

import (
	"testing"

	"l-battle/internal/protocol"
	"l-battle/internal/world"
)

func TestFinalHourManaCooldownDurationAndBonusAttackByRank(t *testing.T) {
	cooldownSeconds := []uint64{100, 85, 70}
	durationSeconds := []uint64{8, 10, 12}
	bonusAttack := []float64{35, 50, 65}
	for index := range cooldownSeconds {
		w, source := tumbleTestWorld(t)
		source.Stats.MP = 200
		learnFinalHour(source, index+1)

		w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID}}, 10, nil, 20)

		if source.Stats.MP != 120 {
			t.Fatalf("rank %d mana after R = %v, want 120", index+1, source.Stats.MP)
		}
		if got, want := source.Skills[rID].CooldownUntilTick, uint64(10)+cooldownSeconds[index]*20; got != want {
			t.Fatalf("rank %d cooldown = %d, want %d", index+1, got, want)
		}
		if got, want := source.Crossbowman.UltimateUntilTick, uint64(10)+durationSeconds[index]*20; got != want {
			t.Fatalf("rank %d duration end = %d, want %d", index+1, got, want)
		}
		if source.Crossbowman.UltimateLevel != index+1 {
			t.Fatalf("ultimate level = %d, want %d", source.Crossbowman.UltimateLevel, index+1)
		}
		if got, want := source.Stats.Attack, 60+bonusAttack[index]; got != want {
			t.Fatalf("rank %d attack = %v, want %v", index+1, got, want)
		}
		if got, want := source.Stats.BonusAttack, bonusAttack[index]; got != want {
			t.Fatalf("rank %d bonus attack = %v, want %v", index+1, got, want)
		}
		if !hasEffect(w.SkillEffects(), "crossbowman_final_hour") {
			t.Fatalf("rank %d final hour effect missing", index+1)
		}
		buffs := ActiveBuffs(w, source, 10)
		if len(buffs) == 0 || buffs[0].ID != rID || buffs[0].ExpiresAtTick != source.Crossbowman.UltimateUntilTick {
			t.Fatalf("rank %d final hour buff = %+v", index+1, buffs)
		}
	}
}

func TestFinalHourRejectsInsufficientManaAndExpiresCleanly(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Stats.MP = 79
	learnFinalHour(source, 1)

	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID}}, 10, nil, 20)
	if source.Crossbowman.UltimateUntilTick != 0 || source.Skills[rID].CooldownUntilTick != 0 || source.Stats.Attack != 60 {
		t.Fatalf("low-mana R changed state: stats=%+v skill=%+v state=%+v", source.Stats, source.Skills[rID], source.Crossbowman)
	}

	source.Stats.MP = 80
	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID}}, 11, nil, 20)
	expiry := source.Crossbowman.UltimateUntilTick
	w.Tick(expiry, 20)
	if source.Crossbowman.UltimateUntilTick != 0 || source.Crossbowman.UltimateLevel != 0 {
		t.Fatalf("expired R state = %+v, want cleared", source.Crossbowman)
	}
	if source.Stats.Attack != 60 || source.Stats.BonusAttack != 0 {
		t.Fatalf("stats after R expiry = attack %v bonus %v, want 60/0", source.Stats.Attack, source.Stats.BonusAttack)
	}
}

func TestFinalHourExtendsWhenMarkedHeroDiesToAlly(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Stats.MP = 100
	learnFinalHour(source, 1)
	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID}}, 10, nil, 20)
	originalUntil := source.Crossbowman.UltimateUntilTick
	target := &world.Entity{
		ID:     "enemy-hero",
		Kind:   world.EntityKindEnemyHero,
		Team:   world.TeamRed,
		Stats:  world.Stats{HP: 1000, MaxHP: 1000},
		Radius: 16,
	}
	target.Combat.LastHitTick = 20
	w.ApplyDamage(source, target, 10, 20)
	if got, want := source.Crossbowman.UltimateDamagedHeroes[target.ID], uint64(80); got != want {
		t.Fatalf("damage mark expiry = %d, want %d", got, want)
	}

	target.Combat.LastHitTick = 79
	ally := &world.Entity{ID: "ally", Kind: world.EntityKindPlayer, Team: world.TeamBlue}
	w.ApplyKillReward(ally, target)
	if got, want := source.Crossbowman.UltimateUntilTick, originalUntil+80; got != want {
		t.Fatalf("extended R expiry = %d, want %d", got, want)
	}
	if _, exists := source.Crossbowman.UltimateDamagedHeroes[target.ID]; exists {
		t.Fatal("consumed R damage mark was not cleared")
	}
}

func TestFinalHourLethalDamageMarksBeforeOwnKillExtension(t *testing.T) {
	w, source := tumbleTestWorld(t)
	source.Stats.MP = 100
	learnFinalHour(source, 1)
	w.ApplyInput("vayne", protocol.PlayerInput{Cast: &protocol.CastInput{SkillID: rID}}, 10, nil, 20)
	originalUntil := source.Crossbowman.UltimateUntilTick
	target := &world.Entity{
		ID:    "enemy-hero",
		Kind:  world.EntityKindEnemyHero,
		Team:  world.TeamRed,
		Stats: world.Stats{HP: 5, MaxHP: 1000},
	}
	target.Combat.LastHitTick = 20
	w.ApplyDamage(source, target, 10, 20)
	if target.Stats.HP != 0 {
		t.Fatalf("lethal damage left target at %v HP", target.Stats.HP)
	}
	w.ApplyKillReward(source, target)
	if got, want := source.Crossbowman.UltimateUntilTick, originalUntil+80; got != want {
		t.Fatalf("R expiry after own lethal hit = %d, want %d", got, want)
	}
}

func TestFinalHourDoesNotExtendWithoutValidActiveMark(t *testing.T) {
	tests := []struct {
		name        string
		markUntil   uint64
		deathTick   uint64
		activeUntil uint64
	}{
		{name: "not marked", deathTick: 30, activeUntil: 100},
		{name: "mark expired", markUntil: 29, deathTick: 30, activeUntil: 100},
		{name: "ultimate expired", markUntil: 50, deathTick: 30, activeUntil: 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, source := tumbleTestWorld(t)
			source.Crossbowman.UltimateUntilTick = tt.activeUntil
			source.Crossbowman.UltimateLevel = 1
			source.Crossbowman.UltimateTickRate = 20
			source.Crossbowman.UltimateDamagedHeroes = map[string]uint64{}
			if tt.markUntil > 0 {
				source.Crossbowman.UltimateDamagedHeroes["enemy-hero"] = tt.markUntil
			}
			target := &world.Entity{ID: "enemy-hero", Kind: world.EntityKindEnemyHero, Team: world.TeamRed}
			target.Combat.LastHitTick = tt.deathTick

			OnKill(w, nil, target)
			if source.Crossbowman.UltimateUntilTick != tt.activeUntil {
				t.Fatalf("R expiry = %d, want unchanged %d", source.Crossbowman.UltimateUntilTick, tt.activeUntil)
			}
		})
	}
}

func learnFinalHour(source *world.Entity, level int) {
	state := source.Skills[rID]
	state.Level = level
	source.Skills[rID] = state
}
