package shadowassassin

import (
	"l-battle/internal/config"
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestAssassinsPathCastSpendsManaStartsCooldownAndResetsAttack(t *testing.T) {
	w, source, _ := qTestWorld(t)
	state := source.Skills[qID]
	state.Level = 3
	source.Skills[qID] = state
	source.Stats.MP = 100
	source.Combat.NextAttackTick = 99
	source.Combat.PendingAttackTargetID = "player:target"
	source.Combat.AttackReleaseTick = 15

	CastQ(w, source, protocol.CastInput{SkillID: qID}, state, w.SkillConfig(qID), 10, 20)

	if source.Stats.MP != 50 {
		t.Fatalf("mana = %v, want 50", source.Stats.MP)
	}
	if got := source.Skills[qID].CooldownUntilTick; got != 130 {
		t.Fatalf("cooldown = %d, want 130", got)
	}
	if !source.ShadowAssassin.QEmpowered || source.ShadowAssassin.QLevel != 3 {
		t.Fatalf("q empower = %+v", source.ShadowAssassin)
	}
	if source.Combat.NextAttackTick != 10 || source.Combat.PendingAttackTargetID != "" || source.Combat.AttackReleaseTick != 0 {
		t.Fatalf("attack was not reset: %+v", source.Combat)
	}
	if !hasEffect(w.SkillEffects(), "shadow_assassin_q_cast") {
		t.Fatal("q cast effect was not created")
	}
	buffs := ActiveBuffs(w, source, 10)
	if len(buffs) != 1 || buffs[0].ID != "shadow_assassin_q_ready" {
		t.Fatalf("q ready buff = %+v", buffs)
	}
}

func TestAssassinsPathBonusDamageUsesBonusAD(t *testing.T) {
	w, source, _ := qTestWorld(t)
	source.ShadowAssassin.QEmpowered = true
	source.ShadowAssassin.QLevel = 2
	source.Stats.BonusAttack = 100

	if got := QBonusPhysicalDamage(w, source, nil, 10, 20); got != 90 {
		t.Fatalf("bonus damage = %d, want 90", got)
	}
}

func TestAssassinsPathHeroHitConsumesEmpowerAndAppliesBleed(t *testing.T) {
	w, source, target := qTestWorld(t)
	source.ShadowAssassin.QEmpowered = true
	source.ShadowAssassin.QLevel = 1
	target.Stats.PhysicalDefense = 0
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000

	OnBasicHit(w, source, target, 10, 20)

	if source.ShadowAssassin.QEmpowered || source.ShadowAssassin.QLevel != 0 {
		t.Fatal("q empower was not consumed")
	}
	if !hasEffect(w.SkillEffects(), "shadow_assassin_q_hit") {
		t.Fatal("q hit effect was not created")
	}
	if buffs := ActiveBuffs(w, source, 10); len(buffs) != 0 {
		t.Fatalf("q ready buff remained after hit: %+v", buffs)
	}
	bleed, ok := target.ShadowAssassin.Bleeds[source.ID]
	if !ok || bleed.ExpiresAtTick != 130 || bleed.NextTick != 30 {
		t.Fatalf("bleed = %+v", bleed)
	}
	if target.Control.MoveSpeedSlow != 0.1 || target.Control.MoveSpeedSlowUntil != 130 {
		t.Fatalf("slow = %v/%d, want 0.1/130", target.Control.MoveSpeedSlow, target.Control.MoveSpeedSlowUntil)
	}
	if !hasEffect(w.SkillEffects(), "shadow_assassin_q_reveal") {
		t.Fatal("bleed did not reveal the target")
	}

	for tick := uint64(30); tick <= 130; tick += 20 {
		w.Tick(tick, 20)
	}
	if got := 1000 - target.Stats.HP; got != 10 {
		t.Fatalf("bleed damage = %v, want 10", got)
	}
	if len(target.ShadowAssassin.Bleeds) != 0 {
		t.Fatal("bleed remained after its final tick")
	}
}

func TestAssassinsPathDoesNotBleedNonHero(t *testing.T) {
	w, source, _ := qTestWorld(t)
	target := &world.Entity{ID: "minion", Kind: world.EntityKindMeleeMinion, Team: world.TeamRed, Stats: world.Stats{HP: 100}}
	source.ShadowAssassin.QEmpowered = true
	source.ShadowAssassin.QLevel = 1

	OnBasicHit(w, source, target, 10, 20)

	if source.ShadowAssassin.QEmpowered || len(target.ShadowAssassin.Bleeds) != 0 {
		t.Fatalf("non-hero hit state source=%+v target=%+v", source.ShadowAssassin, target.ShadowAssassin)
	}
}

func qTestWorld(t *testing.T) (*world.World, *world.Entity, *world.Entity) {
	t.Helper()
	skills, err := config.NewSkillStore([]config.SkillConfig{
		{SkillID: passiveID, Type: "passive", Meta: map[string]float64{"physicalDamageBonus": 0.1}},
		{
			SkillID:    qID,
			CooldownMS: 8000,
			Type:       "self_buff",
			Meta: map[string]float64{
				"bonusAdRatio":         0.3,
				"bleedBonusAdRatio":    1.2,
				"bleedDurationSeconds": 6,
				"bleedTickSeconds":     1,
				"bleedSlow":            0.1,
			},
			MetaLists: map[string][]float64{
				"manaCost":    {40, 45, 50, 55, 60},
				"cooldownMs":  {8000, 7000, 6000, 5000, 4000},
				"bonusDamage": {30, 60, 90, 120, 150},
				"bleedDamage": {10, 20, 30, 40, 50},
			},
		},
		{
			SkillID:    wID,
			CooldownMS: 10000,
			Range:      850,
			Meta: map[string]float64{
				"bonusAdRatio":              0.6,
				"bladeCount":                7,
				"coneAngleDegrees":          50,
				"projectileSpeed":           1700,
				"projectileRadius":          55,
				"projectileLifetimeSeconds": 2.5,
				"slowSeconds":               2,
			},
			MetaLists: map[string][]float64{
				"manaCost":   {60, 65, 70, 75, 80},
				"baseDamage": {30, 55, 80, 105, 130},
				"slow":       {0.2, 0.25, 0.3, 0.35, 0.4},
			},
		},
		{
			SkillID:    eID,
			CooldownMS: 18000,
			Range:      700,
			Meta: map[string]float64{
				"damageAmpSeconds": 3,
				"landingPadding":   8,
				"rootSeconds":      1,
				"effectSeconds":    0.3,
			},
			MetaLists: map[string][]float64{
				"manaCost":   {35, 40, 45, 50, 55},
				"cooldownMs": {18000, 16000, 14000, 12000, 10000},
				"damageAmp":  {0.03, 0.06, 0.09, 0.12, 0.15},
			},
		},
		{
			SkillID:    rID,
			CooldownMS: 75000,
			Meta: map[string]float64{
				"bonusAdRatio":          0.9,
				"bladeCount":            24,
				"invisibilitySeconds":   2.5,
				"moveSpeedBonus":        0.4,
				"projectileRadius":      55,
				"projectileRange":       550,
				"projectileSpeed":       1200,
				"returnLifetimeSeconds": 5,
			},
			MetaLists: map[string][]float64{
				"manaCost":   {80, 90, 100},
				"cooldownMs": {75000, 65000, 55000},
				"damage":     {120, 190, 260},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	hero := config.HeroConfig{
		HeroID: heroID,
		Base: config.BaseStats{
			HP:          1000,
			MP:          500,
			Attack:      100,
			MoveSpeed:   335,
			AttackRange: 125,
			AttackSpeed: 1,
		},
		Radius: 16,
		Skills: config.HeroSkills{
			Passive: passiveID,
			Q:       qID,
			W:       wID,
			E:       eID,
			R:       rID,
		},
	}
	w := world.NewWorld(nil, skills, nil, nil, nil)
	w.SpawnHero("source", hero, world.TeamBlue)
	w.SpawnHero("target", hero, world.TeamRed)
	source := w.EntityByID("player:source")
	target := w.EntityByID("player:target")
	source.Position = world.Vector2{X: 5000, Y: 5000}
	target.Position = world.Vector2{X: 5200, Y: 5000}
	return w, source, target
}

func hasEffect(effects []world.SkillEffect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}
