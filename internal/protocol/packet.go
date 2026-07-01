package protocol

import "encoding/json"

type PacketType string

const (
	PacketJoinRoom    PacketType = "join_room"
	PacketLeave       PacketType = "leave"
	PacketInput       PacketType = "input"
	PacketSpawnObject PacketType = "spawn_object"
	PacketSnapshot    PacketType = "snapshot"
	PacketError       PacketType = "error"
)

type Packet struct {
	Type     PacketType       `json:"type"`
	RoomID   string           `json:"roomId,omitempty"`
	PlayerID string           `json:"playerId,omitempty"`
	Seq      uint64           `json:"seq,omitempty"`
	Payload  *json.RawMessage `json:"payload,omitempty"`
}

type JoinRoom struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
	HeroID   string `json:"heroId"`
	Team     string `json:"team"`
}

type PlayerInput struct {
	MoveX        float64            `json:"moveX,omitempty"`
	MoveY        float64            `json:"moveY,omitempty"`
	Move         *MoveInput         `json:"move,omitempty"`
	Attack       *AttackInput       `json:"attack,omitempty"`
	Cast         *CastInput         `json:"cast,omitempty"`
	UpgradeSkill *UpgradeSkillInput `json:"upgradeSkill,omitempty"`
	DebugLevelUp bool               `json:"debugLevelUp,omitempty"`
	ClientSeq    uint64             `json:"clientSeq"`
}

type MoveInput struct {
	TargetX float64 `json:"targetX"`
	TargetY float64 `json:"targetY"`
}

type AttackInput struct {
	TargetID string `json:"targetId"`
	Clear    bool   `json:"clear,omitempty"`
}

type CastInput struct {
	SkillID  string  `json:"skillId"`
	TargetID string  `json:"targetId,omitempty"`
	TargetX  float64 `json:"targetX"`
	TargetY  float64 `json:"targetY"`
}

type UpgradeSkillInput struct {
	Slot string `json:"slot"`
}

type SpawnObject struct {
	Kind string  `json:"kind"`
	Team string  `json:"team"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type Snapshot struct {
	RoomID  string           `json:"roomId"`
	Tick    uint64           `json:"tick"`
	Map     MapSnapshot      `json:"map"`
	Players []PlayerSnapshot `json:"players"`
	Units   []UnitSnapshot   `json:"units"`
	Dummies []DummySnapshot  `json:"dummies"`
	Effects []EffectSnapshot `json:"effects"`
}

type MapSnapshot struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type PlayerSnapshot struct {
	PlayerID       string          `json:"playerId"`
	HeroID         string          `json:"heroId"`
	Team           string          `json:"team"`
	Level          int             `json:"level"`
	MaxLevel       int             `json:"maxLevel"`
	SkillPoints    int             `json:"skillPoints"`
	Exp            float64         `json:"exp"`
	TotalExp       float64         `json:"totalExp"`
	NextLevelExp   float64         `json:"nextLevelExp"`
	X              float64         `json:"x"`
	Y              float64         `json:"y"`
	Stats          StatsSnapshot   `json:"stats"`
	Skills         []SkillSnapshot `json:"skills"`
	Passive        PassiveSnapshot `json:"passive"`
	LastHitTick    uint64          `json:"lastHitTick"`
	LastDamage     int             `json:"lastDamage"`
	LastDamageType string          `json:"lastDamageType"`
	Dead           bool            `json:"dead"`
	RespawnTick    uint64          `json:"respawnTick"`
	RespawnIn      float64         `json:"respawnIn"`
	Control        ControlSnapshot `json:"control"`
	Sword          SwordSnapshot   `json:"sword"`
	Warrior        WarriorSnapshot `json:"warrior"`
	Tank           TankSnapshot    `json:"tank"`
	Archer         ArcherSnapshot  `json:"archer"`
}

type DummySnapshot struct {
	ID             string        `json:"id"`
	X              float64       `json:"x"`
	Y              float64       `json:"y"`
	Radius         float64       `json:"radius"`
	Stats          StatsSnapshot `json:"stats"`
	LastHitTick    uint64        `json:"lastHitTick"`
	LastDamage     int           `json:"lastDamage"`
	LastDamageType string        `json:"lastDamageType"`
}

type UnitSnapshot struct {
	ID             string          `json:"id"`
	Kind           string          `json:"kind"`
	Team           string          `json:"team"`
	X              float64         `json:"x"`
	Y              float64         `json:"y"`
	Radius         float64         `json:"radius"`
	Stats          StatsSnapshot   `json:"stats"`
	LastHitTick    uint64          `json:"lastHitTick"`
	LastDamage     int             `json:"lastDamage"`
	LastDamageType string          `json:"lastDamageType"`
	Control        ControlSnapshot `json:"control"`
}

type EffectSnapshot struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Team      string  `json:"team"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	EndX      float64 `json:"endX"`
	EndY      float64 `json:"endY"`
	DirX      float64 `json:"dirX"`
	DirY      float64 `json:"dirY"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Radius    float64 `json:"radius"`
	Range     float64 `json:"range"`
	Speed     float64 `json:"speed"`
	CreatedAt uint64  `json:"createdAt"`
	ExpiresAt uint64  `json:"expiresAt"`
}

type StatsSnapshot struct {
	HP                   int     `json:"hp"`
	MaxHP                int     `json:"maxHp"`
	BonusHP              int     `json:"bonusHp"`
	MP                   float64 `json:"mp"`
	MaxMP                float64 `json:"maxMp"`
	HPRegen5             float64 `json:"hpRegen5"`
	MPRegen5             float64 `json:"mpRegen5"`
	Attack               float64 `json:"attack"`
	BonusAttack          float64 `json:"bonusAttack"`
	AbilityPower         int     `json:"abilityPower"`
	AbilityHaste         float64 `json:"abilityHaste"`
	DamageReduce         float64 `json:"damageReduce"`
	PhysicalDefense      float64 `json:"physicalDefense"`
	BonusPhysicalDefense float64 `json:"bonusPhysicalDefense"`
	PhysicalPenPercent   float64 `json:"physicalPenPercent"`
	PhysicalPenFlat      float64 `json:"physicalPenFlat"`
	PhysicalDamageReduce float64 `json:"physicalDamageReduce"`
	MagicDefense         float64 `json:"magicDefense"`
	BonusMagicDefense    float64 `json:"bonusMagicDefense"`
	MagicPenPercent      float64 `json:"magicPenPercent"`
	MagicPenFlat         float64 `json:"magicPenFlat"`
	MagicDamageReduce    float64 `json:"magicDamageReduce"`
	MoveSpeed            float64 `json:"moveSpeed"`
	AttackRange          float64 `json:"attackRange"`
	AttackSpeed          float64 `json:"attackSpeed"`
	BaseAttackSpeed      float64 `json:"baseAttackSpeed"`
	AttackSpeedBonus     float64 `json:"attackSpeedBonus"`
	AttackSpeedRatio     float64 `json:"attackSpeedRatio"`
	AttackSpeedSlow      float64 `json:"attackSpeedSlow"`
	CritChance           float64 `json:"critChance"`
}

type SkillSnapshot struct {
	SkillID           string `json:"skillId"`
	Level             int    `json:"level"`
	CooldownUntilTick uint64 `json:"cooldownUntilTick"`
	Stacks            int    `json:"stacks"`
	StacksExpireTick  uint64 `json:"stacksExpireTick"`
}

type PassiveSnapshot struct {
	SwordIntent    float64 `json:"swordIntent"`
	MaxSwordIntent float64 `json:"maxSwordIntent"`
	Shield         int     `json:"shield"`
	MaxShield      int     `json:"maxShield"`
}

type ControlSnapshot struct {
	AirborneUntilTick     uint64  `json:"airborneUntilTick"`
	DashUntilTick         uint64  `json:"dashUntilTick"`
	ActionLockedUntilTick uint64  `json:"actionLockedUntilTick"`
	StunnedUntilTick      uint64  `json:"stunnedUntilTick"`
	SilencedUntilTick     uint64  `json:"silencedUntilTick"`
	TenacityUntilTick     uint64  `json:"tenacityUntilTick"`
	MoveSpeedSlow         float64 `json:"moveSpeedSlow"`
	MoveSpeedSlowUntil    uint64  `json:"moveSpeedSlowUntil"`
}

type WarriorSnapshot struct {
	JudgmentUntilTick uint64 `json:"judgmentUntilTick"`
}

type TankSnapshot struct {
	ThunderclapAftershockUntil uint64 `json:"thunderclapAftershockUntil"`
}

type ArcherSnapshot struct {
	FocusStacks      int     `json:"focusStacks"`
	FocusExpireTick  uint64  `json:"focusExpireTick"`
	FocusActiveUntil uint64  `json:"focusActiveUntil"`
	FocusAttackSpeed float64 `json:"focusAttackSpeed"`
}

type SwordSnapshot struct {
	SweepingBladeTargetUntil map[string]uint64 `json:"sweepingBladeTargetUntil"`
}

type Error struct {
	Message string `json:"message"`
}
