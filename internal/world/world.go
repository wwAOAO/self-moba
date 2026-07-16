package world

import (
	"l-battle/internal/config"
)

const (
	DefaultMapWidth    = 8000.0
	DefaultMapHeight   = 8000.0
	MinHeroLevel       = 1
	MaxHeroLevel       = 18
	MaxBasicSkillLevel = 5
	MaxUltSkillLevel   = 3
	swordHeroID        = "sword"
	warriorHeroID      = "warrior"
	archerHeroID       = "archer"
	tankHeroID         = "tank"
	mageHeroID         = "mage"
	gunnerHeroID       = "gunner"
	explorerHeroID     = "explorer"
	frostmageHeroID    = "frostmage"
	fireMageHeroID     = "fire_mage"
	bladeHeroID        = "blade"
	berserkerHeroID    = "berserker"
	ninjaHeroID        = "ninja"
	respawnSeconds     = 20
	doctorQSkillID     = "doctor_q"
	doctorWSkillID     = "doctor_w"
	berserkerQSkillID  = "berserker_q"
	berserkerWSkillID  = "berserker_w"
	berserkerESkillID  = "berserker_e"
	berserkerRSkillID  = "berserker_r"
	swordQSkillID      = "sword_cut"
	swordQStackTicks   = 6
	swordWSkillID      = "sword_wind_wall"
	swordESkillID      = "sword_sweeping_blade"
	swordRSkillID      = "sword_storm"
	warriorQSkillID    = "slash"
	warriorWSkillID    = "dash"
	warriorESkillID    = "judgment"
	warriorRSkillID    = "justice"
	tankQSkillID       = "slam"
	tankWSkillID       = "guard"
	tankESkillID       = "taunt"
	tankRSkillID       = "earthquake"
	archerQSkillID     = "shot"
	archerWSkillID     = "roll"
	archerESkillID     = "trap"
	archerRSkillID     = "arrow_rain"
	mageQSkillID       = "mage_q"
	mageWSkillID       = "mage_w"
	mageESkillID       = "mage_e"
	mageRSkillID       = "mage_r"
	fireMageQSkillID   = "fire_mage_q"
	fireMageRSkillID   = "fire_mage_r"
	frostmageQSkillID  = "frostmage_q"
	frostmageESkillID  = "frostmage_e"
	gunnerQSkillID     = "gunner_q"
	gunnerWSkillID     = "gunner_w"
	gunnerRSkillID     = "gunner_r"
	robotQSkillID      = "robot_q"
	explorerQSkillID   = "explorer_q"
	explorerWSkillID   = "explorer_w"
	explorerESkillID   = "explorer_e"
	explorerRSkillID   = "explorer_r"
	ninjaQSkillID      = "ninja_q"
	ninjaRSkillID      = "ninja_r"
	killerQSkillID     = "killer_q"
	killerRSkillID     = "killer_r"
	monkQSkillID       = "monk_q"
	windWallDuration   = 4
)

type World struct {
	width               float64
	height              float64
	entities            map[string]*Entity
	heroes              *config.HeroStore
	skills              *config.SkillStore
	levels              *config.LevelConfig
	rewards             *config.RewardConfig
	equipment           *config.EquipmentStore
	nextObjectID        int
	nextWallID          int
	nextProjectileID    int
	nextEffectID        int
	nextPassiveGoldTick uint64
	windWalls           map[string]WindWall
	projectiles         map[string]*Projectile
	projectileHits      map[string]map[string]bool
	skillEffects        map[string]SkillEffect
	equipmentBurns      map[string]EquipmentBurn
	siegeSplashBurns    map[string]EquipmentBurn
	nextMinionWaveTick  uint64
	minionWaveNumber    int
	pendingMinionSpawns []PendingMinionSpawn
}

func NewWorld(heroes *config.HeroStore, skills *config.SkillStore, levels *config.LevelConfig, rewards *config.RewardConfig, equipment *config.EquipmentStore) *World {
	w := &World{
		width:            DefaultMapWidth,
		height:           DefaultMapHeight,
		entities:         make(map[string]*Entity),
		heroes:           heroes,
		skills:           skills,
		levels:           levels,
		rewards:          rewards,
		equipment:        equipment,
		windWalls:        make(map[string]WindWall),
		projectiles:      make(map[string]*Projectile),
		projectileHits:   make(map[string]map[string]bool),
		skillEffects:     make(map[string]SkillEffect),
		equipmentBurns:   make(map[string]EquipmentBurn),
		siegeSplashBurns: make(map[string]EquipmentBurn),
	}
	w.SpawnBattleUnits()
	return w
}

func (w *World) Tick(tick uint64, tickRate int) {
	w.expireWindWalls(tick)
	w.expireSkillEffects(tick)
	w.tickMinionWaves(tick, tickRate)
	w.tickProjectiles(tick, tickRate)
	w.tickEquipmentBurns(tick, tickRate)
	w.tickSiegeMinionSplashBurns(tick, tickRate)
	w.tickFountains(tick, tickRate)
	w.tickPassiveGold(tick, tickRate)
	for _, entity := range w.entities {
		w.tickHeroEntity(entity, tick, tickRate)
		w.tickUntargetable(entity, tick)
		w.tickPhysicalDefenseShred(entity, tick)
		w.tickAttackDamageReduction(entity, tick)
		w.tickGrievousWounds(entity, tick)
		w.tickDashMovement(entity, tick, tickRate)
		w.tickSwordShield(entity, tick)
		tickEquipmentPhysicalDamageShield(entity, tick)
		w.tickStoneplateShield(entity, tick)
		w.tickSunfire(entity, tick, tickRate)
		w.tickEndlessDespair(entity, tick, tickRate)
		w.tickFountainForTarget(entity, tick, tickRate)
		if entity.Lane.Active {
			w.releasePendingAttack(entity, tick, tickRate)
			w.tickLaneMinion(entity, tick, tickRate)
			continue
		}
		if entity.Kind == EntityKindTower {
			w.releasePendingAttack(entity, tick, tickRate)
			w.tickTower(entity, tick, tickRate)
			continue
		}
		if entity.Kind != EntityKindPlayer {
			continue
		}
		w.tickEquipmentStacks(entity, tick)
		w.tickHero(entity, tick, tickRate)
		w.releasePendingAttack(entity, tick, tickRate)
		w.tickRespawn(entity, tick)
		if entity.Death.Dead || entity.Stats.HP <= 0 {
			continue
		}
		w.tickBaseRegen(entity, tickRate)
		w.tickEquipmentPercentRegen(entity, tick, tickRate)
		w.tickWarmog(entity, tick, tickRate)
		w.tickPlayer(entity, tick, tickRate)
	}
}

func (w *World) tickUntargetable(entity *Entity, tick uint64) {
	if entity != nil && entity.Control.UntargetableUntilTick > 0 && tick >= entity.Control.UntargetableUntilTick {
		entity.Control.UntargetableUntilTick = 0
	}
}

func (w *World) tickPhysicalDefenseShred(entity *Entity, tick uint64) {
	if entity == nil || entity.Combat.PhysicalDefenseShredAmount <= 0 || tick < entity.Combat.PhysicalDefenseShredUntil {
		return
	}
	entity.Stats.PhysicalDefense += entity.Combat.PhysicalDefenseShredAmount
	entity.Combat.PhysicalDefenseShredAmount = 0
	entity.Combat.PhysicalDefenseShredUntil = 0
}

func (w *World) tickDashMovement(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.Control.DashUntilTick == 0 || tick < entity.Control.DashStartTick {
		return
	}
	if entity.Control.DashUntilTick <= entity.Control.DashStartTick {
		entity.Control.DashUntilTick = 0
		return
	}
	before := entity.Position
	if tick >= entity.Control.DashUntilTick {
		entity.Position = entity.Control.DashEnd
		entity.Control.DashUntilTick = 0
		entity.Control.DashStartTick = 0
		w.resolveTankRImpact(entity, tick, tickRate)
	} else {
		progress := float64(tick-entity.Control.DashStartTick) / float64(entity.Control.DashUntilTick-entity.Control.DashStartTick)
		entity.Position = Vector2{
			X: entity.Control.DashStart.X + (entity.Control.DashEnd.X-entity.Control.DashStart.X)*progress,
			Y: entity.Control.DashStart.Y + (entity.Control.DashEnd.Y-entity.Control.DashStart.Y)*progress,
		}
	}
	moved := distance(before, entity.Position)
	w.chargeSwordIntent(entity, moved)
	w.chargeEquipmentOnMove(entity, moved)
}

func (w *World) tickRespawn(entity *Entity, tick uint64) {
	if !entity.Death.Dead || tick < entity.Death.RespawnTick {
		return
	}
	entity.Death = DeathState{}
	entity.Position = w.spawnPosition(entity.Team)
	entity.Stats.HP = entity.Stats.MaxHP
	if entity.HeroID == bladeHeroID {
		entity.Stats.MP = 0
	} else {
		entity.Stats.MP = entity.Stats.MaxMP
	}
	entity.Intent = IntentState{}
	entity.Control = ControlState{}
	w.removeDoctorCanister(entity)
	entity.Sword = swordStateForHero(entity.HeroID)
	entity.Warrior = WarriorState{}
	entity.Archer = ArcherState{}
	entity.Mage = MageState{}
	entity.Tank = TankState{}
	entity.Berserker = BerserkerState{}
	entity.Ninja = NinjaState{}
	entity.Passive.GunnerTargetID = ""
	entity.Passive.GunnerWActiveUntil = 0
	entity.Passive.GunnerWAttackSpeed = 0
	entity.Passive.GunnerWMoveSpeed = 0
	entity.Passive.GunnerECenter = Vector2{}
	entity.Passive.GunnerEExpireTick = 0
	entity.Passive.GunnerENextTick = 0
	entity.Passive.GunnerELevel = 0
	entity.Passive.GunnerEEffectID = ""
	entity.Passive.GunnerRDir = Vector2{}
	entity.Passive.GunnerRStartTick = 0
	entity.Passive.GunnerRExpireTick = 0
	entity.Passive.GunnerRNextTick = 0
	entity.Passive.GunnerRLevel = 0
	entity.Passive.GunnerRWaves = 0
	entity.Passive.GunnerRWaveCount = 0
	entity.Passive.GunnerREffectID = ""
	entity.Passive.KillerRStartTick = 0
	entity.Passive.KillerRExpireTick = 0
	entity.Passive.KillerRNextTick = 0
	entity.Passive.KillerRLevel = 0
	entity.Passive.KillerRSegmentsFired = 0
	entity.Passive.KillerREffectID = ""
	entity.Passive.KillerRMoveSpeedMultiplier = 0
	entity.Passive.ButcherRTargetID = ""
	entity.Passive.ButcherRStartPosition = Vector2{}
	entity.Passive.ButcherRUntil = 0
	entity.Passive.ButcherRNextTick = 0
	entity.Passive.ButcherRLevel = 0
	entity.Passive.ButcherREffectID = ""
	entity.Passive.ButcherRPreviousStunUntil = 0
	entity.Passive.ButcherRAppliedStunUntil = 0
	entity.Passive.RobotShieldUntil = 0
	entity.Passive.RobotShieldMana = 0
	entity.Passive.RobotQPending = false
	entity.Passive.RobotQReleaseTick = 0
	entity.Passive.RobotQTarget = Vector2{}
	entity.Passive.RobotQLevel = 0
	entity.Passive.RobotWStartTick = 0
	entity.Passive.RobotWUntil = 0
	entity.Passive.RobotWLevel = 0
	entity.Passive.RobotWMoveSpeed = 0
	entity.Passive.RobotArcMarks = nil
	entity.Passive.ExplorerSpellForceStacks = 0
	entity.Passive.ExplorerSpellForceExpiresAt = 0
	entity.Passive.ExplorerFluxMarks = nil
	entity.Passive.ExplorerQPending = false
	entity.Passive.ExplorerQRelease = 0
	entity.Passive.ExplorerQTarget = Vector2{}
	entity.Passive.ExplorerQLevel = 0
	entity.Passive.ExplorerWTarget = Vector2{}
	entity.Passive.ExplorerWLevel = 0
	entity.Passive.ExplorerEPending = false
	entity.Passive.ExplorerERelease = 0
	entity.Passive.ExplorerETarget = Vector2{}
	entity.Passive.ExplorerELevel = 0
	entity.Passive.ExplorerRPending = false
	entity.Passive.ExplorerRRelease = 0
	entity.Passive.ExplorerRTarget = Vector2{}
	entity.Passive.ExplorerRLevel = 0
	entity.Passive.FrostServants = nil
	entity.Passive.FrostQPending = false
	entity.Passive.FrostQRelease = 0
	entity.Passive.FrostQTarget = Vector2{}
	entity.Passive.FrostQLevel = 0
	entity.Passive.FrostEPending = false
	entity.Passive.FrostERelease = 0
	entity.Passive.FrostETarget = Vector2{}
	entity.Passive.FrostELevel = 0
	entity.Passive.FrostEProjectileID = ""
	entity.Passive.FrostERecastTick = 0
	entity.Passive.FrostRPending = false
	entity.Passive.FrostRRelease = 0
	entity.Passive.FrostRTargetID = ""
	entity.Passive.FrostRLevel = 0
	entity.Passive.FrostRSelfUntil = 0
	entity.Passive.FrostRSelfLevel = 0
	entity.Passive.FrostRSelfEffectID = ""
	entity.Passive.FrostRSelfHealLeft = 0
	entity.Passive.FrostRSelfHealTicks = 0
	entity.Passive.FrostROldDamageReduce = 0
	entity.Passive.MonkFlurryUntil = 0
	entity.Passive.MonkFlurryAttacks = 0
	entity.Passive.MonkFlurryHitIndex = 0
	entity.Passive.MonkQPending = false
	entity.Passive.MonkQRelease = 0
	entity.Passive.MonkQTarget = Vector2{}
	entity.Passive.MonkQLevel = 0
	entity.Passive.MonkQMarkTargetID = ""
	entity.Passive.MonkQMarkUntil = 0
	entity.Passive.MonkQMarkLevel = 0
	entity.Passive.MonkQMarkEffectID = ""
	entity.Passive.MonkWRecastUntil = 0
	entity.Passive.MonkWIronWillUntil = 0
	entity.Passive.MonkWIronWillLevel = 0
	entity.Passive.MonkEPending = false
	entity.Passive.MonkERelease = 0
	entity.Passive.MonkELevel = 0
	entity.Passive.MonkERecastUntil = 0
	entity.Passive.MonkEHitIDs = nil
	entity.Passive.MonkESlows = nil
	entity.Passive.MonkRPending = false
	entity.Passive.MonkRRelease = 0
	entity.Passive.MonkRTargetID = ""
	entity.Passive.MonkRLevel = 0
	entity.Passive.Shield = 0
	entity.Passive.MaxShield = 0
	entity.Passive.ShieldExpireTick = 0
	entity.Passive.Bleeds = nil
	entity.Passive.LastRegenBreakTick = tick
	entity.Passive.NextRegenTick = 0
	w.refreshTankGraniteShield(entity)
	w.refreshTankWPassive(entity)
	entity.Combat.NextAttackTick = tick
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

func (w *World) Units() []Entity {
	units := make([]Entity, 0, len(w.entities))
	for _, entity := range w.entities {
		if entity.Kind != EntityKindPlayer {
			units = append(units, *entity)
		}
	}
	return units
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
