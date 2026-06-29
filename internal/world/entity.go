package world

type Vector2 struct {
	X float64
	Y float64
}

type Entity struct {
	ID           string
	Kind         EntityKind
	Team         Team
	PlayerID     string
	HeroID       string
	Level        int
	Exp          float64
	TotalExp     float64
	NextLevelExp float64
	Position     Vector2
	Stats        Stats
	Radius       float64
	Skills       map[string]SkillState
	Combat       CombatState
	Passive      PassiveState
	Intent       IntentState
}

type EntityKind string

const (
	EntityKindPlayer       EntityKind = "player"
	EntityKindEnemyHero    EntityKind = "enemy_hero"
	EntityKindSiegeMinion  EntityKind = "siege_minion"
	EntityKindMeleeMinion  EntityKind = "melee_minion"
	EntityKindRangedMinion EntityKind = "ranged_minion"
	EntityKindTower        EntityKind = "tower"
	EntityKindCrystal      EntityKind = "crystal"
	EntityKindBarracks     EntityKind = "barracks"
	EntityKindDummy        EntityKind = "dummy"
)

type Team string

const (
	TeamBlue    Team = "blue"
	TeamRed     Team = "red"
	TeamNeutral Team = "neutral"
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
	CritChance      float64
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

type PassiveState struct {
	SwordIntent    float64
	MaxSwordIntent float64
	Shield         int
	MaxShield      int
}

type IntentState struct {
	MoveTarget       *Vector2
	AttackTargetID   string
	AttackPausedTill uint64
}
