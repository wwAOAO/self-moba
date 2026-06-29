package system

import "l-battle/internal/world"

type Buff struct {
	ID       string
	Owner    *world.Entity
	Duration int
}

func TickBuff(_ Buff) {
}
