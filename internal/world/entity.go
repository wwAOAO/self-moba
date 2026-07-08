package world

import (
	"l-battle/internal/world/effect"
	"l-battle/internal/world/geom"
	"l-battle/internal/world/model"
	"l-battle/internal/world/projectile"
	"l-battle/internal/world/state"
)

type Vector2 = geom.Vector2
type EntityKind = model.EntityKind
type Team = model.Team
type Projectile = projectile.Projectile
type WindWall = effect.WindWall
type EquipmentBurn = effect.EquipmentBurn
type SkillEffect = effect.SkillEffect
type BuffState = state.BuffState
type RegenState = state.RegenState
type EquipmentSlot = state.EquipmentSlot
type Stats = state.Stats
type SkillState = state.SkillState
type CombatState = state.CombatState
type DamageEvent = state.DamageEvent
type ControlState = state.ControlState
type SwordState = state.SwordState
type WarriorState = state.WarriorState
type ArcherState = state.ArcherState
type MageState = state.MageState
type TankState = state.TankState
type BerserkerState = state.BerserkerState
type NinjaState = state.NinjaState
type PassiveState = state.PassiveState
type BleedState = state.BleedState
type RobotArcState = state.RobotArcState
type ExplorerFluxState = state.ExplorerFluxState
type FireBurnState = state.FireBurnState
type ShieldLayer = state.ShieldLayer
type DeathState = state.DeathState
type IntentState = state.IntentState
type LaneState = state.LaneState
type PendingMinionSpawn = state.PendingMinionSpawn

type Entity struct {
	ID           string
	Kind         EntityKind
	Team         Team
	PlayerID     string
	HeroID       string
	Level        int
	SkillPoints  int
	Gold         float64
	Equipment    []EquipmentSlot
	Buffs        []BuffState
	Exp          float64
	TotalExp     float64
	NextLevelExp float64
	Position     Vector2
	Stats        Stats
	Message      string
	MessageTick  uint64
	Radius       float64
	Skills       map[string]SkillState
	Combat       CombatState
	Control      ControlState
	Passive      PassiveState
	Sword        SwordState
	Warrior      WarriorState
	Archer       ArcherState
	Mage         MageState
	Tank         TankState
	Berserker    BerserkerState
	Ninja        NinjaState
	Death        DeathState
	Intent       IntentState
	Lane         LaneState
	Regen        RegenState
}

const (
	EntityKindPlayer       = model.EntityKindPlayer
	EntityKindEnemyHero    = model.EntityKindEnemyHero
	EntityKindSiegeMinion  = model.EntityKindSiegeMinion
	EntityKindSuperMinion  = model.EntityKindSuperMinion
	EntityKindMeleeMinion  = model.EntityKindMeleeMinion
	EntityKindRangedMinion = model.EntityKindRangedMinion
	EntityKindBlueBuff     = model.EntityKindBlueBuff
	EntityKindRedBuff      = model.EntityKindRedBuff
	EntityKindGromp        = model.EntityKindGromp
	EntityKindRaptor       = model.EntityKindRaptor
	EntityKindMurkWolf     = model.EntityKindMurkWolf
	EntityKindKrugCamp     = model.EntityKindKrugCamp
	EntityKindBaronNashor  = model.EntityKindBaronNashor
	EntityKindTower        = model.EntityKindTower
	EntityKindCrystal      = model.EntityKindCrystal
	EntityKindBarracks     = model.EntityKindBarracks
	EntityKindFountain     = model.EntityKindFountain
	EntityKindDummy        = model.EntityKindDummy
	TeamBlue               = model.TeamBlue
	TeamRed                = model.TeamRed
	TeamNeutral            = model.TeamNeutral
)
