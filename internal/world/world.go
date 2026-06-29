package world

import (
	"math"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

const (
	DefaultMapWidth  = 2000.0
	DefaultMapHeight = 2000.0
)

type World struct {
	width    float64
	height   float64
	entities map[string]*Entity
}

func NewWorld() *World {
	w := &World{
		width:    DefaultMapWidth,
		height:   DefaultMapHeight,
		entities: make(map[string]*Entity),
	}
	w.SpawnTrainingDummy()
	return w
}

func (w *World) SpawnHero(playerID string, hero config.HeroConfig) {
	if _, ok := w.entities[playerID]; ok {
		return
	}
	skillIDs := hero.Skills.SkillIDs()
	skills := make(map[string]SkillState, len(skillIDs))
	for _, skillID := range skillIDs {
		skills[skillID] = SkillState{SkillID: skillID}
	}
	w.entities[playerID] = &Entity{
		ID:       "player:" + playerID,
		Kind:     EntityKindPlayer,
		PlayerID: playerID,
		HeroID:   hero.HeroID,
		Stats: Stats{
			HP:              hero.Base.HP,
			MaxHP:           hero.Base.HP,
			MP:              hero.Base.MP,
			MaxMP:           hero.Base.MP,
			Attack:          hero.Base.Attack,
			PhysicalDefense: hero.Base.PhysicalDefense,
			MagicDefense:    hero.Base.MagicDefense,
			MoveSpeed:       hero.Base.MoveSpeed,
			AttackRange:     hero.Base.AttackRange,
			AttackSpeed:     hero.Base.AttackSpeed,
		},
		Radius: hero.Radius,
		Skills: skills,
		Position: Vector2{
			X: w.width / 2,
			Y: w.height / 2,
		},
	}
}

func (w *World) SpawnTrainingDummy() {
	w.spawnDummy("dummy:training-1", w.width/2+180, w.height/2)
	w.spawnDummy("dummy:training-2", w.width/2+180, w.height/2+200)
}

func (w *World) spawnDummy(id string, x float64, y float64) {
	if _, ok := w.entities[id]; ok {
		return
	}
	w.entities[id] = &Entity{
		ID:   id,
		Kind: EntityKindDummy,
		Stats: Stats{
			HP:              3000,
			MaxHP:           3000,
			PhysicalDefense: 10,
			MagicDefense:    10,
		},
		Radius: 28,
		Position: Vector2{
			X: x,
			Y: y,
		},
	}
}

func (w *World) RemovePlayer(playerID string) {
	delete(w.entities, playerID)
}

func (w *World) ApplyInput(playerID string, input protocol.PlayerInput, tick uint64, skills *config.SkillStore, tickRate int) {
	entity := w.entities[playerID]
	if entity == nil {
		return
	}
	dx, dy := normalize(input.MoveX, input.MoveY)
	entity.Position.X += dx * entity.Stats.MoveSpeed
	entity.Position.Y += dy * entity.Stats.MoveSpeed
	entity.Position.X = clamp(entity.Position.X, 0, w.width)
	entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
	if input.Cast != nil {
		w.applyCast(entity, *input.Cast, tick, skills, tickRate)
	}
	if input.Attack != nil {
		w.applyAttack(entity, *input.Attack, tick, tickRate)
	}
}

func (w *World) Entities() []Entity {
	entities := make([]Entity, 0, len(w.entities))
	for _, entity := range w.entities {
		entities = append(entities, *entity)
	}
	return entities
}

func (w *World) Players() []Entity {
	return w.entitiesByKind(EntityKindPlayer)
}

func (w *World) Dummies() []Entity {
	return w.entitiesByKind(EntityKindDummy)
}

func (w *World) Size() (float64, float64) {
	return w.width, w.height
}

func (w *World) entitiesByKind(kind EntityKind) []Entity {
	entities := make([]Entity, 0, len(w.entities))
	for _, entity := range w.entities {
		if entity.Kind == kind {
			entities = append(entities, *entity)
		}
	}
	return entities
}

func normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (w *World) applyCast(entity *Entity, cast protocol.CastInput, tick uint64, skills *config.SkillStore, tickRate int) {
	state, ok := entity.Skills[cast.SkillID]
	if !ok {
		return
	}
	if tick < state.CooldownUntilTick {
		return
	}
	skill, ok := skills.Get(cast.SkillID)
	if !ok {
		return
	}
	state.CooldownUntilTick = tick + cooldownTicks(skill.CooldownMS, tickRate)
	entity.Skills[cast.SkillID] = state
}

func cooldownTicks(cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 {
		return 0
	}
	ticks := math.Ceil(float64(cooldownMS) / 1000 * float64(tickRate))
	return uint64(ticks)
}

func (w *World) applyAttack(attacker *Entity, attack protocol.AttackInput, tick uint64, tickRate int) {
	if attacker.Kind != EntityKindPlayer || tick < attacker.Combat.NextAttackTick {
		return
	}
	target := w.entities[attack.TargetID]
	if target == nil || target.Stats.HP <= 0 {
		return
	}
	if distance(attacker.Position, target.Position) > attacker.Stats.AttackRange+attacker.Radius+target.Radius {
		return
	}

	damage := attacker.Stats.Attack - target.Stats.PhysicalDefense
	if damage < 1 {
		damage = 1
	}
	target.Combat.LastHitTick = tick
	target.Combat.LastDamage = damage
	if target.Kind != EntityKindDummy {
		target.Stats.HP -= damage
		if target.Stats.HP < 0 {
			target.Stats.HP = 0
		}
	}
	attacker.Combat.NextAttackTick = tick + attackCooldownTicks(attacker.Stats.AttackSpeed, tickRate)
}

func attackCooldownTicks(attackSpeed float64, tickRate int) uint64 {
	if attackSpeed <= 0 {
		return uint64(tickRate)
	}
	ticks := math.Ceil(float64(tickRate) / attackSpeed)
	if ticks < 1 {
		return 1
	}
	return uint64(ticks)
}

func distance(a Vector2, b Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}
