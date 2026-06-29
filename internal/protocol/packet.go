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
	MoveX     float64      `json:"moveX,omitempty"`
	MoveY     float64      `json:"moveY,omitempty"`
	Move      *MoveInput   `json:"move,omitempty"`
	Attack    *AttackInput `json:"attack,omitempty"`
	Cast      *CastInput   `json:"cast,omitempty"`
	ClientSeq uint64       `json:"clientSeq"`
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
	SkillID string  `json:"skillId"`
	TargetX float64 `json:"targetX"`
	TargetY float64 `json:"targetY"`
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
}

type MapSnapshot struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type PlayerSnapshot struct {
	PlayerID     string          `json:"playerId"`
	HeroID       string          `json:"heroId"`
	Team         string          `json:"team"`
	Level        int             `json:"level"`
	MaxLevel     int             `json:"maxLevel"`
	Exp          float64         `json:"exp"`
	TotalExp     float64         `json:"totalExp"`
	NextLevelExp float64         `json:"nextLevelExp"`
	X            float64         `json:"x"`
	Y            float64         `json:"y"`
	Stats        StatsSnapshot   `json:"stats"`
	Skills       []SkillSnapshot `json:"skills"`
	Passive      PassiveSnapshot `json:"passive"`
	LastHitTick  uint64          `json:"lastHitTick"`
	LastDamage   int             `json:"lastDamage"`
}

type DummySnapshot struct {
	ID          string        `json:"id"`
	X           float64       `json:"x"`
	Y           float64       `json:"y"`
	Radius      float64       `json:"radius"`
	Stats       StatsSnapshot `json:"stats"`
	LastHitTick uint64        `json:"lastHitTick"`
	LastDamage  int           `json:"lastDamage"`
}

type UnitSnapshot struct {
	ID          string        `json:"id"`
	Kind        string        `json:"kind"`
	Team        string        `json:"team"`
	X           float64       `json:"x"`
	Y           float64       `json:"y"`
	Radius      float64       `json:"radius"`
	Stats       StatsSnapshot `json:"stats"`
	LastHitTick uint64        `json:"lastHitTick"`
	LastDamage  int           `json:"lastDamage"`
}

type StatsSnapshot struct {
	HP              int     `json:"hp"`
	MaxHP           int     `json:"maxHp"`
	MP              int     `json:"mp"`
	MaxMP           int     `json:"maxMp"`
	Attack          int     `json:"attack"`
	PhysicalDefense int     `json:"physicalDefense"`
	MagicDefense    int     `json:"magicDefense"`
	MoveSpeed       float64 `json:"moveSpeed"`
	AttackRange     float64 `json:"attackRange"`
	AttackSpeed     float64 `json:"attackSpeed"`
	CritChance      float64 `json:"critChance"`
}

type SkillSnapshot struct {
	SkillID           string `json:"skillId"`
	CooldownUntilTick uint64 `json:"cooldownUntilTick"`
}

type PassiveSnapshot struct {
	SwordIntent    float64 `json:"swordIntent"`
	MaxSwordIntent float64 `json:"maxSwordIntent"`
	Shield         int     `json:"shield"`
	MaxShield      int     `json:"maxShield"`
}

type Error struct {
	Message string `json:"message"`
}
