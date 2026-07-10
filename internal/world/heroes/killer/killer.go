package killer

import "l-battle/internal/world"

func init() {
	world.RegisterHeroHooks("killer", world.HeroHooks{})
}
