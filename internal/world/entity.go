package world

type Vector2 struct {
	X float64
	Y float64
}

type WindWall struct {
	ID        string
	Team      Team
	Center    Vector2
	Dir       Vector2
	Width     float64
	ExpiresAt uint64
}

type SkillEffect struct {
	ID        string
	Kind      string
	Team      Team
	Start     Vector2
	End       Vector2
	Dir       Vector2
	Range     float64
	Radius    float64
	Width     float64
	Height    float64
	Count     int
	Speed     float64
	CreatedAt uint64
	ExpiresAt uint64
}

type Projectile struct {
	ID           string
	Kind         string
	Team         Team
	SourceID     string
	TargetID     string
	SkillID      string
	GroupID      string
	Position     Vector2
	Start        Vector2
	Dir          Vector2
	SpeedPerTick float64
	SpeedMin     float64
	SpeedMax     float64
	Range        float64
	Radius       float64
	DisplayRange float64
	DisplayCount int
	Traveled     float64
	Damage       int
	KnockupTicks uint64
	EffectRatio  float64
	EffectTicks  uint64
	CreatedAt    uint64
	ExpiresAt    uint64
	HitIDs       map[string]bool
}

type Entity struct {
	ID           string
	Kind         EntityKind
	Team         Team
	PlayerID     string
	HeroID       string
	Level        int
	SkillPoints  int
	Exp          float64
	TotalExp     float64
	NextLevelExp float64
	Position     Vector2
	Stats        Stats
	Radius       float64
	Skills       map[string]SkillState
	Combat       CombatState
	Control      ControlState
	Passive      PassiveState
	Sword        SwordState
	Warrior      WarriorState
	Archer       ArcherState
	Tank         TankState
	Death        DeathState
	Intent       IntentState
	Regen        RegenState
}

type RegenState struct {
	HPRemainder float64
	MPRemainder float64
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
	HP                   int
	MaxHP                int
	BonusHP              int
	MP                   float64
	MaxMP                float64
	HPRegen5             float64
	MPRegen5             float64
	Attack               float64
	BonusAttack          float64
	AbilityPower         int
	AbilityHaste         float64
	DamageReduce         float64
	PhysicalDefense      float64
	MagicDefense         float64
	BonusPhysicalDefense float64
	BonusMagicDefense    float64
	PhysicalPenPercent   float64
	PhysicalPenFlat      float64
	PhysicalDamageReduce float64
	MagicPenPercent      float64
	MagicPenFlat         float64
	MagicDamageReduce    float64
	MoveSpeed            float64
	AttackRange          float64
	AttackSpeed          float64
	AttackWindupSeconds  float64
	BaseAttackSpeed      float64
	AttackSpeedBonus     float64
	AttackSpeedRatio     float64
	AttackSpeedSlow      float64
	CritChance           float64
}

type SkillState struct {
	SkillID           string
	Level             int
	CooldownUntilTick uint64
	Stacks            int
	StacksExpireTick  uint64
}

type CombatState struct {
	NextAttackTick             uint64
	PendingAttackTargetID      string
	AttackReleaseTick          uint64
	LastHitTick                uint64
	LastDamage                 int
	LastDamageType             string
	PhysicalDefenseShredUntil  uint64
	PhysicalDefenseShredAmount float64
}

type ControlState struct {
	AirborneUntilTick     uint64
	DashUntilTick         uint64
	DashStartTick         uint64
	DashStart             Vector2
	DashEnd               Vector2
	ActionLockedUntilTick uint64
	StunnedUntilTick      uint64
	SilencedUntilTick     uint64
	TenacityUntilTick     uint64
	MoveSpeedBonusFlat    float64
	MoveSpeedBonusUntil   uint64
	MoveSpeedSlow         float64
	MoveSpeedSlowUntil    uint64
	AttackSpeedSlow       float64
	AttackSpeedSlowUntil  uint64
}

type SwordState struct {
	SweepingBladeStacks      int
	SweepingBladeTargetUntil map[string]uint64
	LastBreathUntilTick      uint64
	QPending                 bool
	QReleaseTick             uint64
	QTarget                  Vector2
	QForm                    string
	QRange                   float64
}

type WarriorState struct {
	DecisiveStrikeUntilTick      uint64
	DecisiveStrikeSpeedUntilTick uint64
	DecisiveStrikeLevel          int
	DecisiveStrikeMoveSpeedBonus float64
	CourageUntilTick             uint64
	CourageFrontUntilTick        uint64
	CourageFrontDamageReduce     float64
	CourageFrontTenacity         float64
	CourageBackDamageReduce      float64
	CouragePassiveResistGain     float64
	JudgmentUntilTick            uint64
	JudgmentNextSpinTick         uint64
	JudgmentSpinIntervalTicks    uint64
	JudgmentSpinsRemaining       int
	JudgmentLevel                int
	JudgmentHits                 map[string]int
	JusticePending               bool
	JusticeReleaseTick           uint64
	JusticeTargetID              string
	JusticeLevel                 int
}

type ArcherState struct {
	FocusStacks             int
	FocusExpireTick         uint64
	FocusActiveUntil        uint64
	FocusActiveLevel        int
	FocusAttackSpeed        float64
	FocusBonusADRatio       float64
	CrystalArrowPending     bool
	CrystalArrowReleaseTick uint64
	CrystalArrowTarget      Vector2
	CrystalArrowLevel       int
}

type TankState struct {
	ThunderclapArmorBonus      float64
	ThunderclapEmpowerUntil    uint64
	ThunderclapAftershockUntil uint64
	ThunderclapLevel           int
	SeismicShardPending        bool
	SeismicShardReleaseTick    uint64
	SeismicShardTargetID       string
	SeismicShardLevel          int
	GroundSlamPending          bool
	GroundSlamReleaseTick      uint64
	GroundSlamLevel            int
	UnstoppableCastPending     bool
	UnstoppableCastTarget      Vector2
	UnstoppableCastLevel       int
	UnstoppableImpactPending   bool
	UnstoppableImpactTick      uint64
	UnstoppableImpactLevel     int
	UnstoppableImpactRadius    float64
	UnstoppableKnockupTicks    uint64
}

type PassiveState struct {
	SwordIntent        float64
	MaxSwordIntent     float64
	Shield             int
	MaxShield          int
	ShieldExpireTick   uint64
	LastRegenBreakTick uint64
	NextRegenTick      uint64
}

type DeathState struct {
	Dead              bool
	RespawnTick       uint64
	RespawnTickRate   int
	RespawnSeconds    int
	LastDeathPosition Vector2
}

type IntentState struct {
	MoveTarget       *Vector2
	AttackTargetID   string
	AttackPausedTill uint64
}
