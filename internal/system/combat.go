package system

import "l-battle/internal/world"

func ApplyDamage(target *world.Entity, amount int) {
	if target == nil || amount <= 0 {
		return
	}
}
