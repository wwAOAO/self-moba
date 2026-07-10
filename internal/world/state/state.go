package state

import (
	"l-battle/internal/world/geom"
	"l-battle/internal/world/model"
)

type BuffState struct {
	ID              string
	Name            string
	Stacks          int
	Tooltip         string
	AbilityHaste    float64
	ExpiresAtTick   uint64
	ExplosionAtTick uint64
	Negative        bool
}

type RegenState struct {
	HPRemainder float64
	MPRemainder float64
}

type EquipmentSlot struct {
	EquipmentID               string
	Name                      string
	Stacks                    float64
	LowHealthShieldUsed       bool
	LowHealthDamageReduce     float64
	LowHealthShieldThreshold  float64
	LowHealthShieldMin        int
	LowHealthShieldMax        int
	OutOfCombatMoveSpeed      float64
	OutOfCombatRequiredTicks  uint64
	ManaShieldCooldownUntil   uint64
	RagebladeHitCount         int
	StackExpireTick           uint64
	SunfireActiveUntil        uint64
	SunfireNextTick           uint64
	SunfireStackExpireTick    uint64
	WarmogNextTick            uint64
	WarmogRequiredTicks       uint64
	EndlessDespairActiveUntil uint64
	EndlessDespairNextTick    uint64
	PhysicalShieldMaxAmount   int
	PhysicalShieldAmount      int
	PhysicalShieldStartTick   uint64
	PhysicalShieldExpireTick  uint64
	LifeStealOverhealShield   int
	StoneplateShieldRatio     float64
	StoneplateResistPercent   float64
	StoneplateCooldownTicks   uint64
	StoneplateCooldownUntil   uint64
	StoneplateBreakTick       uint64
	StoneplateShieldActive    bool
	StoneplateShieldAmount    int
	HeroDamageBonusUntil      uint64
	EchoCharge                float64
	EchoCooldownUntil         uint64
}

type Stats struct {
	HP                   float64
	MaxHP                float64
	BonusHP              float64
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
	Tenacity             float64
	SlowResist           float64
	BasicAttackBlock     float64
	CritDamageReduce     float64
	MoveSpeed            float64
	AttackRange          float64
	AttackSpeed          float64
	AttackWindupSeconds  float64
	BaseAttackSpeed      float64
	AttackSpeedBonus     float64
	AttackSpeedRatio     float64
	AttackSpeedSlow      float64
	CritChance           float64
	Omnivamp             float64
	LifeSteal            float64
	HealingPower         float64
	GrievousWounds       float64
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
	DamageEvents               []DamageEvent
	DamageEventsTick           uint64
	NextSiegeSplashTick        uint64
	PhysicalDefenseShredUntil  uint64
	PhysicalDefenseShredAmount float64
	BlackCleaverStacks         int
	BlackCleaverUntil          uint64
}

type DamageEvent struct {
	Damage      int
	DamageType  string
	BasicAttack bool
	SourceID    string
}

type ControlState struct {
	AirborneUntilTick       uint64
	UntargetableUntilTick   uint64
	DashUntilTick           uint64
	DashStartTick           uint64
	DashStart               geom.Vector2
	DashEnd                 geom.Vector2
	ActionLockedUntilTick   uint64
	StunnedUntilTick        uint64
	TauntedUntilTick        uint64
	SuppressedUntilTick     uint64
	SilencedUntilTick       uint64
	TenacityUntilTick       uint64
	MoveSpeedBonusFlat      float64
	MoveSpeedBonusUntil     uint64
	MoveSpeedSlow           float64
	MoveSpeedSlowUntil      uint64
	AttackDamageReduction   float64
	AttackDamageReduceUntil uint64
	UndyingRageUntil        uint64
	UndyingRageMinHP        float64
	RootedUntilTick         uint64
	AttackSpeedSlow         float64
	AttackSpeedSlowUntil    uint64
	GrievousWounds          float64
	GrievousWoundsUntil     uint64
	MageIlluminationUntil   uint64
	MageIlluminationBy      string
	MageFinalSparkBy        string
	MageFinalSparkUntil     uint64
	MageFinalSparkRefund    float64
}

type SwordState struct {
	SweepingBladeStacks      int
	SweepingBladeTargetUntil map[string]uint64
	LastBreathUntilTick      uint64
	QPending                 bool
	QReleaseTick             uint64
	QTarget                  geom.Vector2
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
	CrystalArrowTarget      geom.Vector2
	CrystalArrowLevel       int
}

type MageState struct {
	LightBindingPending           bool
	LightBindingReleaseTick       uint64
	LightBindingTarget            geom.Vector2
	LightBindingLevel             int
	PrismaticBarrierPending       bool
	PrismaticBarrierReleaseTick   uint64
	PrismaticBarrierTarget        geom.Vector2
	PrismaticBarrierLevel         int
	LucentSingularityPending      bool
	LucentSingularityReleaseTick  uint64
	LucentSingularityTarget       geom.Vector2
	LucentSingularityActive       bool
	LucentSingularityCenter       geom.Vector2
	LucentSingularityExpireTick   uint64
	LucentSingularityLevel        int
	LucentSingularityEffectID     string
	LucentSingularityProjectileID string
	FinalSparkPending             bool
	FinalSparkReleaseTick         uint64
	FinalSparkTarget              geom.Vector2
	FinalSparkLevel               int
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
	UnstoppableCastTarget      geom.Vector2
	UnstoppableCastLevel       int
	UnstoppableImpactPending   bool
	UnstoppableImpactTick      uint64
	UnstoppableImpactLevel     int
	UnstoppableImpactRadius    float64
	UnstoppableKnockupTicks    uint64
}

type BerserkerState struct {
	BloodRageUntil              uint64
	ApprehendDir                geom.Vector2
	NoxianGuillotineTarget      string
	NoxianGuillotineLevel       int
	NoxianGuillotineCastPending bool
	NoxianGuillotineCastTarget  string
	NoxianGuillotineRecast      uint64
	NoxianGuillotineRestore     uint64
}

type NinjaState struct {
	ShadowPosition       geom.Vector2
	ShadowExpiresAt      uint64
	ShadowEffectID       string
	ShadowReadyTick      uint64
	ShadowRecastSkillID  string
	ShadowRecastUntil    uint64
	RShadowPosition      geom.Vector2
	RShadowExpiresAt     uint64
	RShadowEffectID      string
	RShadowRecastUntil   uint64
	QPending             bool
	QReleaseTick         uint64
	QTarget              geom.Vector2
	QLevel               int
	PendingShadowQTarget geom.Vector2
	PendingShadowQLevel  int
	PendingShadowQGroup  string
	PendingShadowELevel  int
	PendingShadowEGroup  string
	PendingShadowEHitIDs map[string]bool
	SkillHitMarks        map[string]uint8
	SkillEnergyRefunded  map[string]bool
	RPending             bool
	RCastPending         bool
	RCastTargetID        string
	RCastLevel           int
	RCastPoint           geom.Vector2
	RCastPointSet        bool
	RReleaseTick         uint64
	RDashEndTick         uint64
	RTargetID            string
	RLevel               int
	RMarkTargetID        string
	RMarkTriggerTick     uint64
	RMarkDamage          float64
	RMarkLevel           int
}

type PassiveState struct {
	SwordIntent                 float64
	MaxSwordIntent              float64
	ButcherFlesh                int
	ButcherQPending             bool
	ButcherQRelease             uint64
	ButcherQTarget              geom.Vector2
	ButcherQLevel               int
	ButcherWActive              bool
	ButcherWNextTick            uint64
	ButcherWLevel               int
	ButcherWEffectID            string
	ButcherEUntil               uint64
	ButcherELevel               int
	ButcherEEffectID            string
	ButcherRTargetID            string
	ButcherRStartPosition       geom.Vector2
	ButcherRUntil               uint64
	ButcherRNextTick            uint64
	ButcherRLevel               int
	ButcherREffectID            string
	ButcherRPreviousStunUntil   uint64
	ButcherRAppliedStunUntil    uint64
	KillerVoracityMarks         map[string]KillerVoracityMark
	KillerQPending              bool
	KillerQReleaseTick          uint64
	KillerQTargetID             string
	KillerQLevel                int
	KillerDaggers               []KillerDaggerState
	KillerAirborneDaggers       []KillerAirborneDaggerState
	KillerWMoveSpeedStartTick   uint64
	KillerWMoveSpeedUntilTick   uint64
	KillerWMoveSpeedBonus       float64
	KillerEDamageReduceUntil    uint64
	KillerEDamageReduction      float64
	KillerRStartTick            uint64
	KillerRExpireTick           uint64
	KillerRNextTick             uint64
	KillerRLevel                int
	KillerRSegmentsFired        int
	KillerREffectID             string
	KillerRMoveSpeedMultiplier  float64
	GunnerTargetID              string
	GunnerWActiveUntil          uint64
	GunnerWAttackSpeed          float64
	GunnerWMoveSpeed            float64
	GunnerECenter               geom.Vector2
	GunnerEExpireTick           uint64
	GunnerENextTick             uint64
	GunnerELevel                int
	GunnerEEffectID             string
	GunnerRDir                  geom.Vector2
	GunnerRStartTick            uint64
	GunnerRExpireTick           uint64
	GunnerRNextTick             uint64
	GunnerRLevel                int
	GunnerRWaves                int
	GunnerRWaveCount            int
	GunnerREffectID             string
	RobotShieldUntil            uint64
	RobotShieldCDUntil          uint64
	RobotShieldMana             int
	RobotQPending               bool
	RobotQReleaseTick           uint64
	RobotQTarget                geom.Vector2
	RobotQLevel                 int
	RobotWStartTick             uint64
	RobotWUntil                 uint64
	RobotWLevel                 int
	RobotWMoveSpeed             float64
	RobotArcMarks               map[string]RobotArcState
	ExplorerSpellForceStacks    int
	ExplorerSpellForceExpiresAt uint64
	ExplorerFluxMarks           map[string]ExplorerFluxState
	ExplorerQPending            bool
	ExplorerQRelease            uint64
	ExplorerQTarget             geom.Vector2
	ExplorerQLevel              int
	ExplorerWTarget             geom.Vector2
	ExplorerWLevel              int
	ExplorerEPending            bool
	ExplorerERelease            uint64
	ExplorerETarget             geom.Vector2
	ExplorerELevel              int
	ExplorerRPending            bool
	ExplorerRRelease            uint64
	ExplorerRTarget             geom.Vector2
	ExplorerRLevel              int
	FrostServants               []FrostServantState
	FrostQPending               bool
	FrostQRelease               uint64
	FrostQTarget                geom.Vector2
	FrostQLevel                 int
	FrostEPending               bool
	FrostERelease               uint64
	FrostETarget                geom.Vector2
	FrostELevel                 int
	FrostEProjectileID          string
	FrostERecastTick            uint64
	FrostRPending               bool
	FrostRRelease               uint64
	FrostRTargetID              string
	FrostRLevel                 int
	FrostRSelfUntil             uint64
	FrostRSelfLevel             int
	FrostRSelfEffectID          string
	FrostRSelfHealLeft          float64
	FrostRSelfHealTicks         uint64
	FrostROldDamageReduce       float64
	FireBurns                   map[string]FireBurnState
	FireManaUntil               uint64
	FireManaNextTick            uint64
	FireQPending                bool
	FireQReleaseTick            uint64
	FireQTarget                 geom.Vector2
	FireQLevel                  int
	FireWPending                bool
	FireWTriggerTick            uint64
	FireWCenter                 geom.Vector2
	FireWLevel                  int
	FireWCastPending            bool
	FireWCastTarget             geom.Vector2
	FireWCastLevel              int
	FireRPending                bool
	FireRReleaseTick            uint64
	FireRTargetID               string
	FireRLevel                  int
	DoctorPassiveCooldownUntil  uint64
	DoctorCanisterEffectID      string
	DoctorCanisterPosition      geom.Vector2
	DoctorCanisterRadius        float64
	DoctorCanisterExpiresAt     uint64
	DoctorQPending              bool
	DoctorQRelease              uint64
	DoctorQTarget               geom.Vector2
	DoctorQLevel                int
	DoctorQHealthCost           float64
	DoctorWActiveUntil          uint64
	DoctorWNextDamageTick       uint64
	DoctorWLevel                int
	DoctorWGrayHealth           float64
	DoctorWEffectID             string
	DoctorRUntil                uint64
	DoctorRNextHealTick         uint64
	DoctorRLevel                int
	DoctorREffectID             string
	MonkFlurryUntil             uint64
	MonkFlurryAttacks           int
	MonkFlurryHitIndex          int
	MonkQPending                bool
	MonkQRelease                uint64
	MonkQTarget                 geom.Vector2
	MonkQLevel                  int
	MonkQMarkTargetID           string
	MonkQMarkUntil              uint64
	MonkQMarkLevel              int
	MonkQMarkEffectID           string
	MonkWRecastUntil            uint64
	MonkWIronWillUntil          uint64
	MonkWIronWillLevel          int
	MonkEPending                bool
	MonkERelease                uint64
	MonkELevel                  int
	MonkERecastUntil            uint64
	MonkEHitIDs                 map[string]bool
	MonkESlows                  map[string]MonkESlowState
	MonkRPending                bool
	MonkRRelease                uint64
	MonkRTargetID               string
	MonkRLevel                  int
	NinjaSoulCooldowns          map[string]uint64
	Shield                      int
	MaxShield                   int
	ShieldExpireTick            uint64
	ShieldLayers                []ShieldLayer
	LastRegenBreakTick          uint64
	NextRegenTick               uint64
	NextFountainTick            uint64
	Bleeds                      map[string]BleedState
}

type KillerVoracityMark struct {
	DamagedAt uint64
	ExpiresAt uint64
	TickRate  int
}

type KillerDaggerState struct {
	EffectID  string
	Position  geom.Vector2
	ExpiresAt uint64
}

type KillerAirborneDaggerState struct {
	EffectID           string
	Position           geom.Vector2
	Direction          geom.Vector2
	LandsAt            uint64
	GroundEffectKind   string
	GroundEffectPrefix string
}

type BleedState struct {
	Stacks        int
	ExpiresAtTick uint64
	NextTick      uint64
	Remainder     float64
}

type RobotArcState struct {
	Stacks      int
	TriggerTick uint64
}

type ExplorerFluxState struct {
	Level     int
	ExpiresAt uint64
}

type FireBurnState struct {
	Stacks          int
	ExpiresAtTick   uint64
	NextTick        uint64
	ExplosionAtTick uint64
}

type FrostServantState struct {
	ID        string
	Position  geom.Vector2
	ExpiresAt uint64
	EffectID  string
}

type MonkESlowState struct {
	StartTick uint64
	Until     uint64
	Slow      float64
}

type ShieldLayer struct {
	Amount    int
	ExpiresAt uint64
}

type DeathState struct {
	Dead              bool
	RespawnTick       uint64
	RespawnTickRate   int
	RespawnSeconds    int
	LastDeathPosition geom.Vector2
}

type IntentState struct {
	MoveTarget       *geom.Vector2
	AttackTargetID   string
	AttackPausedTill uint64
}

type LaneState struct {
	Active         bool
	RouteTarget    geom.Vector2
	LastOnLaneTick uint64
}

type PendingMinionSpawn struct {
	Team       model.Team
	Kind       model.EntityKind
	Index      int
	WaveNumber int
	SpawnTick  uint64
}
