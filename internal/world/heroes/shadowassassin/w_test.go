package shadowassassin

import (
	"l-battle/internal/protocol"
	"l-battle/internal/world"
	"testing"
)

func TestCutthroatCastSpendsManaStartsCooldownAndCreatesFan(t *testing.T) {
	w, source, _ := qTestWorld(t)
	state := source.Skills[wID]
	state.Level = 2
	source.Skills[wID] = state
	source.Stats.MP = 100

	CastW(w, source, protocol.CastInput{SkillID: wID, TargetX: source.Position.X + 850, TargetY: source.Position.Y}, state, w.SkillConfig(wID), 10, 20)

	if source.Stats.MP != 35 {
		t.Fatalf("mana = %v, want 35", source.Stats.MP)
	}
	if got := source.Skills[wID].CooldownUntilTick; got != 210 {
		t.Fatalf("cooldown = %d, want 210", got)
	}
	projectiles := w.SkillEffects()
	count := 0
	for _, effect := range projectiles {
		if effect.Kind == "shadow_assassin_w" {
			count++
		}
	}
	if count != 7 {
		t.Fatalf("blade count = %d, want 7", count)
	}
}

func TestCutthroatOutboundAndReturnEachHitOnce(t *testing.T) {
	w, source, target := qTestWorld(t)
	source.Stats.BonusAttack = 100
	target.Stats.PhysicalDefense = 0
	target.Stats.HP = 1000
	target.Stats.MaxHP = 1000
	hits := make(map[string]bool)
	projectile := &world.Projectile{
		SkillID:  wID,
		Damage:   1,
		Radius:   55,
		HitIDs:   hits,
		Position: world.Vector2{X: target.Position.X + 100, Y: target.Position.Y},
	}
	previous := world.Vector2{X: target.Position.X - 100, Y: target.Position.Y}

	ResolveWProjectile(w, source, projectile, previous, 10, 20)
	ResolveWProjectile(w, source, projectile, previous, 10, 20)
	if got := 1000 - target.Stats.HP; got != 90 {
		t.Fatalf("outbound damage = %v, want 90", got)
	}
	projectile.Returning = true
	projectile.EffectTicks = 10
	ResolveWProjectile(w, source, projectile, previous, 11, 20)
	ResolveWProjectile(w, source, projectile, previous, 11, 20)
	if got := 1000 - target.Stats.HP; got != 189 {
		t.Fatalf("two-pass damage = %v, want 189", got)
	}
	if target.Control.MoveSpeedSlow != 0.2 || target.Control.MoveSpeedSlowUntil != 51 {
		t.Fatalf("slow = %v/%d, want 0.2/51", target.Control.MoveSpeedSlow, target.Control.MoveSpeedSlowUntil)
	}
}

func TestCutthroatProjectileReturnsWithoutEndpointDelay(t *testing.T) {
	w, source, target := qTestWorld(t)
	state := source.Skills[wID]
	state.Level = 1
	source.Skills[wID] = state
	target.Position = world.Vector2{X: source.Position.X + 400, Y: source.Position.Y}
	target.Stats.PhysicalDefense = 0
	target.Stats.HP = 1000

	CastW(w, source, protocol.CastInput{SkillID: wID, TargetX: source.Position.X + 850, TargetY: source.Position.Y}, state, w.SkillConfig(wID), 10, 20)
	for tick := uint64(11); tick <= 35; tick++ {
		w.Tick(tick, 20)
	}
	if got := 1000 - target.Stats.HP; got != 63 {
		t.Fatalf("flight damage = %v, want 63", got)
	}
}
