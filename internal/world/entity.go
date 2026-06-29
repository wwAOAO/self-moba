package world

type Vector2 struct {
	X float64
	Y float64
}

type Entity struct {
	ID       string
	Kind     EntityKind
	PlayerID string
	HeroID   string
	Position Vector2
	Stats    Stats
	Radius   float64
	Skills   map[string]SkillState
	Combat   CombatState
}

type EntityKind string

const (
	EntityKindPlayer EntityKind = "player"
	EntityKindDummy  EntityKind = "dummy"
)

type Stats struct {
	HP              int
	MaxHP           int
	MP              int
	MaxMP           int
	Attack          int
	PhysicalDefense int
	MagicDefense    int
	MoveSpeed       float64
	AttackRange     float64
	AttackSpeed     float64
}

type SkillState struct {
	SkillID           string
	CooldownUntilTick uint64
}

type CombatState struct {
	NextAttackTick uint64
	LastHitTick    uint64
	LastDamage     int
}
