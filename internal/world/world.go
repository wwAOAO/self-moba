package world

import (
	"l-battle/internal/config"
)

const (
	DefaultMapWidth    = 6000.0
	DefaultMapHeight   = 6000.0
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
	bladeHeroID        = "blade"
	berserkerHeroID    = "berserker"
	ninjaHeroID        = "ninja"
	respawnSeconds     = 20
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
	gunnerQSkillID     = "gunner_q"
	gunnerWSkillID     = "gunner_w"
	gunnerRSkillID     = "gunner_r"
	ninjaQSkillID      = "ninja_q"
	ninjaRSkillID      = "ninja_r"
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
	windWalls           map[string]WindWall
	projectiles         map[string]*Projectile
	projectileHits      map[string]map[string]bool
	skillEffects        map[string]SkillEffect
	equipmentBurns      map[string]EquipmentBurn
	nextMinionWaveTick  uint64
	minionWaveNumber    int
	pendingMinionSpawns []PendingMinionSpawn
}

func NewWorld(heroes *config.HeroStore, skills *config.SkillStore, levels *config.LevelConfig, rewards *config.RewardConfig, equipment *config.EquipmentStore) *World {
	w := &World{
		width:          DefaultMapWidth,
		height:         DefaultMapHeight,
		entities:       make(map[string]*Entity),
		heroes:         heroes,
		skills:         skills,
		levels:         levels,
		rewards:        rewards,
		equipment:      equipment,
		windWalls:      make(map[string]WindWall),
		projectiles:    make(map[string]*Projectile),
		projectileHits: make(map[string]map[string]bool),
		skillEffects:   make(map[string]SkillEffect),
		equipmentBurns: make(map[string]EquipmentBurn),
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
	w.tickFountains(tick, tickRate)
	for _, entity := range w.entities {
		w.tickHeroEntity(entity, tick, tickRate)
		w.tickUntargetable(entity, tick)
		w.tickPhysicalDefenseShred(entity, tick)
		w.tickAttackDamageReduction(entity, tick)
		w.tickSwordShield(entity, tick)
		tickEquipmentPhysicalDamageShield(entity, tick)
		w.tickStoneplateShield(entity, tick)
		w.tickSunfire(entity, tick, tickRate)
		w.tickFountainForTarget(entity, tick, tickRate)
		if entity.Lane.Active {
			w.releasePendingAttack(entity, tick, tickRate)
			w.tickLaneMinion(entity, tick, tickRate)
			continue
		}
		if entity.Kind != EntityKindPlayer {
			continue
		}
		w.tickEquipmentStacks(entity, tick)
		w.tickHero(entity, tick, tickRate)
		w.tickDashMovement(entity, tick, tickRate)
		w.releasePendingAttack(entity, tick, tickRate)
		w.tickRespawn(entity, tick)
		if entity.Death.Dead || entity.Stats.HP <= 0 {
			continue
		}
		w.tickBaseRegen(entity, tickRate)
		w.tickEquipmentPercentRegen(entity, tick, tickRate)
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
	w.chargeSwordIntent(entity, distance(before, entity.Position))
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
