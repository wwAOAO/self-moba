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
	MoveX             float64             `json:"moveX,omitempty"`
	MoveY             float64             `json:"moveY,omitempty"`
	Move              *MoveInput          `json:"move,omitempty"`
	Attack            *AttackInput        `json:"attack,omitempty"`
	Cast              *CastInput          `json:"cast,omitempty"`
	UpgradeSkill      *UpgradeSkillInput  `json:"upgradeSkill,omitempty"`
	BuyEquipment      *BuyEquipmentInput  `json:"buyEquipment,omitempty"`
	SellEquipment     *SellEquipmentInput `json:"sellEquipment,omitempty"`
	DebugLevelUp      bool                `json:"debugLevelUp,omitempty"`
	DebugAbilityHaste *float64            `json:"debugAbilityHaste,omitempty"`
	DebugGold         float64             `json:"debugGold,omitempty"`
	ClientSeq         uint64              `json:"clientSeq"`
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

type BuyEquipmentInput struct {
	EquipmentID string `json:"equipmentId"`
}

type SellEquipmentInput struct {
	Slot int `json:"slot"`
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

// PlayerSnapshot 描述客户端渲染和交互所需的玩家英雄权威状态。
type PlayerSnapshot struct {
	// PlayerID 是玩家在房间内的唯一标识。
	PlayerID string `json:"playerId"`
	// HeroID 是玩家当前使用的英雄配置标识。
	HeroID string `json:"heroId"`
	// Team 是玩家所属阵营。
	Team string `json:"team"`
	// Level 是英雄当前等级。
	Level int `json:"level"`
	// MaxLevel 是英雄允许达到的最高等级。
	MaxLevel int `json:"maxLevel"`
	// SkillPoints 是当前未分配的技能点数。
	SkillPoints int `json:"skillPoints"`
	// Gold 是玩家当前持有的金币数。
	Gold float64 `json:"gold"`
	// Equipment 是固定槽位顺序的装备快照。
	Equipment []EquipmentSlot `json:"equipment"`
	// Exp 是当前等级内的经验值。
	Exp float64 `json:"exp"`
	// TotalExp 是玩家累计获得的经验值。
	TotalExp float64 `json:"totalExp"`
	// NextLevelExp 是升到下一级所需的累计经验值。
	NextLevelExp float64 `json:"nextLevelExp"`
	// Message 是本次状态变更产生的可选提示消息。
	Message string `json:"message,omitempty"`
	// MessageTick 是提示消息产生的服务端 tick。
	MessageTick uint64 `json:"messageTick,omitempty"`
	// X 是英雄中心在地图中的横坐标，单位为世界单位。
	X float64 `json:"x"`
	// Y 是英雄中心在地图中的纵坐标，单位为世界单位。
	Y float64 `json:"y"`
	// Radius 是英雄的权威碰撞半径，单位为世界单位。
	Radius float64 `json:"radius"`
	// Stats 是英雄当前生效的战斗属性。
	Stats StatsSnapshot `json:"stats"`
	// Skills 是英雄各技能的等级和冷却状态。
	Skills []SkillSnapshot `json:"skills"`
	// Buffs 是英雄当前生效的增益和减益状态。
	Buffs []BuffSnapshot `json:"buffs,omitempty"`
	// Passive 是跨英雄通用的被动状态快照。
	Passive PassiveSnapshot `json:"passive"`
	// LastHitTick 是英雄最近一次受到伤害的服务端 tick。
	LastHitTick uint64 `json:"lastHitTick"`
	// LastDamage 是英雄最近一次受到的伤害值。
	LastDamage int `json:"lastDamage"`
	// LastDamageType 是英雄最近一次受到的伤害类型。
	LastDamageType string `json:"lastDamageType"`
	// DamageEvents 是尚需客户端展示的伤害事件。
	DamageEvents []DamageEventSnapshot `json:"damageEvents,omitempty"`
	// Dead 表示英雄当前是否处于死亡状态。
	Dead bool `json:"dead"`
	// RespawnTick 是英雄计划复活的服务端 tick。
	RespawnTick uint64 `json:"respawnTick"`
	// RespawnIn 是距离复活的剩余秒数。
	RespawnIn float64 `json:"respawnIn"`
	// Control 是英雄当前的控制状态。
	Control ControlSnapshot `json:"control"`
	// Sword 是剑客英雄的专属状态。
	Sword SwordSnapshot `json:"sword"`
	// Warrior 是圣骑士英雄的专属状态。
	Warrior WarriorSnapshot `json:"warrior"`
	// Tank 是石头人英雄的专属状态。
	Tank TankSnapshot `json:"tank"`
	// Archer 是弓箭手英雄的专属状态。
	Archer ArcherSnapshot `json:"archer"`
	// Ninja 是忍者英雄的专属状态。
	Ninja NinjaSnapshot `json:"ninja"`
}

type BuffSnapshot struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Stacks          int     `json:"stacks,omitempty"`
	Tooltip         string  `json:"tooltip,omitempty"`
	ExpiresAtTick   uint64  `json:"expiresAtTick,omitempty"`
	ExplosionAtTick uint64  `json:"explosionAtTick,omitempty"`
	Negative        bool    `json:"negative,omitempty"`
	AbilityHaste    float64 `json:"abilityHaste,omitempty"`
}

type DummySnapshot struct {
	ID             string                `json:"id"`
	X              float64               `json:"x"`
	Y              float64               `json:"y"`
	Radius         float64               `json:"radius"`
	Stats          StatsSnapshot         `json:"stats"`
	Buffs          []BuffSnapshot        `json:"buffs,omitempty"`
	LastHitTick    uint64                `json:"lastHitTick"`
	LastDamage     int                   `json:"lastDamage"`
	LastDamageType string                `json:"lastDamageType"`
	DamageEvents   []DamageEventSnapshot `json:"damageEvents,omitempty"`
}

type UnitSnapshot struct {
	ID             string                `json:"id"`
	Kind           string                `json:"kind"`
	Team           string                `json:"team"`
	X              float64               `json:"x"`
	Y              float64               `json:"y"`
	Radius         float64               `json:"radius"`
	Stats          StatsSnapshot         `json:"stats"`
	Buffs          []BuffSnapshot        `json:"buffs,omitempty"`
	LastHitTick    uint64                `json:"lastHitTick"`
	LastDamage     int                   `json:"lastDamage"`
	LastDamageType string                `json:"lastDamageType"`
	DamageEvents   []DamageEventSnapshot `json:"damageEvents,omitempty"`
	Control        ControlSnapshot       `json:"control"`
}

type DamageEventSnapshot struct {
	Damage      int    `json:"damage"`
	DamageType  string `json:"damageType"`
	BasicAttack bool   `json:"basicAttack,omitempty"`
	SourceID    string `json:"sourceId,omitempty"`
}

type EquipmentSlot struct {
	EquipmentID string `json:"equipmentId"`
	Name        string `json:"name"`
	Stacks      int    `json:"stacks,omitempty"`
}

type EffectSnapshot struct {
	ID           string  `json:"id"`
	Kind         string  `json:"kind"`
	Team         string  `json:"team"`
	SourceID     string  `json:"sourceId,omitempty"`
	SourceHeroID string  `json:"sourceHeroId,omitempty"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	EndX         float64 `json:"endX"`
	EndY         float64 `json:"endY"`
	DirX         float64 `json:"dirX"`
	DirY         float64 `json:"dirY"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Radius       float64 `json:"radius"`
	Range        float64 `json:"range"`
	Speed        float64 `json:"speed"`
	CreatedAt    uint64  `json:"createdAt"`
	ExpiresAt    uint64  `json:"expiresAt"`
	Count        int     `json:"count"`
}

type StatsSnapshot struct {
	HP                   float64 `json:"hp"`
	MaxHP                float64 `json:"maxHp"`
	BonusHP              float64 `json:"bonusHp"`
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
	Omnivamp             float64 `json:"omnivamp"`
	LifeSteal            float64 `json:"lifeSteal"`
	HealingPower         float64 `json:"healingPower"`
	GrievousWounds       float64 `json:"grievousWounds"`
}

type SkillSnapshot struct {
	SkillID           string `json:"skillId"`
	Level             int    `json:"level"`
	CooldownUntilTick uint64 `json:"cooldownUntilTick"`
	Stacks            int    `json:"stacks"`
	StacksExpireTick  uint64 `json:"stacksExpireTick"`
}

type PassiveSnapshot struct {
	SwordIntent        float64           `json:"swordIntent"`
	MaxSwordIntent     float64           `json:"maxSwordIntent"`
	NinjaSoulCooldowns map[string]uint64 `json:"ninjaSoulCooldowns,omitempty"`
	Shield             int               `json:"shield"`
	MaxShield          int               `json:"maxShield"`
	MonkQMarkUntil     uint64            `json:"monkQMarkUntil,omitempty"`
	MonkWRecastUntil   uint64            `json:"monkWRecastUntil,omitempty"`
	MonkERecastUntil   uint64            `json:"monkERecastUntil,omitempty"`
}

type ControlSnapshot struct {
	AirborneUntilTick     uint64  `json:"airborneUntilTick"`
	DashUntilTick         uint64  `json:"dashUntilTick"`
	InvisibleUntilTick    uint64  `json:"invisibleUntilTick,omitempty"`
	ActionLockedUntilTick uint64  `json:"actionLockedUntilTick"`
	StunnedUntilTick      uint64  `json:"stunnedUntilTick"`
	SilencedUntilTick     uint64  `json:"silencedUntilTick"`
	TenacityUntilTick     uint64  `json:"tenacityUntilTick"`
	MoveSpeedSlow         float64 `json:"moveSpeedSlow"`
	MoveSpeedSlowUntil    uint64  `json:"moveSpeedSlowUntil"`
	RootedUntilTick       uint64  `json:"rootedUntilTick,omitempty"`
	MageIlluminationUntil uint64  `json:"mageIlluminationUntil,omitempty"`
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

type NinjaSnapshot struct {
	ShadowX             float64 `json:"shadowX"`
	ShadowY             float64 `json:"shadowY"`
	ShadowExpiresAt     uint64  `json:"shadowExpiresAt"`
	ShadowReadyTick     uint64  `json:"shadowReadyTick,omitempty"`
	RShadowX            float64 `json:"rShadowX"`
	RShadowY            float64 `json:"rShadowY"`
	RShadowExpiresAt    uint64  `json:"rShadowExpiresAt"`
	ShadowRecastSkillID string  `json:"shadowRecastSkillId,omitempty"`
	RShadowRecastUntil  uint64  `json:"rShadowRecastUntil,omitempty"`
}

type Error struct {
	Message string `json:"message"`
}
