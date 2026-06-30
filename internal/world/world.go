package world

import (
	"math"
	"strconv"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
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
	tankHeroID         = "tank"
	critRollModulo     = 10000
	respawnSeconds     = 20
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
	windWallDuration   = 4
)

type World struct {
	width            float64
	height           float64
	entities         map[string]*Entity
	heroes           *config.HeroStore
	skills           *config.SkillStore
	levels           *config.LevelConfig
	rewards          *config.RewardConfig
	nextObjectID     int
	nextWallID       int
	nextProjectileID int
	nextEffectID     int
	windWalls        map[string]WindWall
	projectiles      map[string]*Projectile
	skillEffects     map[string]SkillEffect
}

func NewWorld(heroes *config.HeroStore, skills *config.SkillStore, levels *config.LevelConfig, rewards *config.RewardConfig) *World {
	w := &World{
		width:        DefaultMapWidth,
		height:       DefaultMapHeight,
		entities:     make(map[string]*Entity),
		heroes:       heroes,
		skills:       skills,
		levels:       levels,
		rewards:      rewards,
		windWalls:    make(map[string]WindWall),
		projectiles:  make(map[string]*Projectile),
		skillEffects: make(map[string]SkillEffect),
	}
	w.SpawnBattleUnits()
	w.SpawnTrainingDummy()
	return w
}

func (w *World) SpawnHero(playerID string, hero config.HeroConfig, team Team) {
	if team != TeamRed {
		team = TeamBlue
	}
	entityID := playerEntityID(playerID)
	skillIDs := hero.Skills.SkillIDs()
	skills := make(map[string]SkillState, len(skillIDs))
	for _, skillID := range skillIDs {
		skills[skillID] = SkillState{SkillID: skillID}
	}
	position := w.spawnPosition(team)
	level := MinHeroLevel
	stats := heroStatsAtLevel(hero, level)
	nextLevelExp := w.nextLevelExp(level)
	startingSkillPoints := 1
	if entity := w.entities[entityID]; entity != nil {
		entity.Team = team
		entity.HeroID = hero.HeroID
		entity.Level = level
		entity.SkillPoints = startingSkillPoints
		entity.Exp = 0
		entity.TotalExp = 0
		entity.NextLevelExp = nextLevelExp
		entity.Stats = stats
		entity.Radius = hero.Radius
		entity.Skills = skills
		entity.Position = position
		entity.Combat = CombatState{}
		entity.Passive = w.passiveStateForHero(hero)
		entity.Sword = swordStateForHero(hero.HeroID)
		entity.Warrior = WarriorState{}
		entity.Tank = TankState{}
		entity.Death = DeathState{}
		return
	}
	w.entities[entityID] = &Entity{
		ID:           entityID,
		Kind:         EntityKindPlayer,
		Team:         team,
		PlayerID:     playerID,
		HeroID:       hero.HeroID,
		Level:        level,
		SkillPoints:  startingSkillPoints,
		Exp:          0,
		TotalExp:     0,
		NextLevelExp: nextLevelExp,
		Stats:        stats,
		Radius:       hero.Radius,
		Skills:       skills,
		Position:     position,
		Passive:      w.passiveStateForHero(hero),
		Sword:        swordStateForHero(hero.HeroID),
	}
}

func (w *World) SpawnBattleUnits() {
	w.spawnUnit("enemy:blue-hero-1", EntityKindEnemyHero, TeamBlue, w.width/2-420, w.height/2+220, 18, Stats{
		HP:              1200,
		MaxHP:           1200,
		MP:              500,
		MaxMP:           500,
		Attack:          82,
		PhysicalDefense: 26,
		MagicDefense:    18,
		MoveSpeed:       4.2,
		AttackRange:     150,
		AttackSpeed:     1,
	})
	w.spawnUnit("minion:blue-melee-1", EntityKindMeleeMinion, TeamBlue, w.width/2-360, w.height/2+70, 14, Stats{
		HP:              420,
		MaxHP:           420,
		Attack:          32,
		PhysicalDefense: 8,
		MagicDefense:    4,
		MoveSpeed:       3,
		AttackRange:     70,
		AttackSpeed:     0.8,
	})
	w.spawnUnit("minion:blue-ranged-1", EntityKindRangedMinion, TeamBlue, w.width/2-430, w.height/2, 13, Stats{
		HP:              300,
		MaxHP:           300,
		Attack:          38,
		PhysicalDefense: 5,
		MagicDefense:    5,
		MoveSpeed:       3,
		AttackRange:     360,
		AttackSpeed:     0.7,
	})
	w.spawnUnit("minion:blue-siege-1", EntityKindSiegeMinion, TeamBlue, w.width/2-500, w.height/2-80, 18, Stats{
		HP:              680,
		MaxHP:           680,
		Attack:          62,
		PhysicalDefense: 14,
		MagicDefense:    8,
		MoveSpeed:       2.4,
		AttackRange:     430,
		AttackSpeed:     0.55,
	})
	w.spawnUnit("structure:blue-tower-1", EntityKindTower, TeamBlue, w.width/2-700, w.height/2+240, 34, Stats{
		HP:              2600,
		MaxHP:           2600,
		Attack:          180,
		PhysicalDefense: 80,
		MagicDefense:    60,
		AttackRange:     620,
		AttackSpeed:     0.75,
	})
	w.spawnUnit("structure:blue-barracks-1", EntityKindBarracks, TeamBlue, w.width/2-760, w.height/2-80, 40, Stats{
		HP:              3200,
		MaxHP:           3200,
		PhysicalDefense: 55,
		MagicDefense:    45,
	})
	w.spawnUnit("structure:blue-crystal", EntityKindCrystal, TeamBlue, w.width/2-900, w.height/2-260, 48, Stats{
		HP:              4500,
		MaxHP:           4500,
		PhysicalDefense: 70,
		MagicDefense:    70,
	})
	w.spawnUnit("enemy:hero-1", EntityKindEnemyHero, TeamRed, w.width/2+420, w.height/2-220, 18, Stats{
		HP:              1200,
		MaxHP:           1200,
		MP:              500,
		MaxMP:           500,
		Attack:          82,
		PhysicalDefense: 26,
		MagicDefense:    18,
		MoveSpeed:       4.2,
		AttackRange:     150,
		AttackSpeed:     1,
	})
	w.spawnUnit("minion:red-melee-1", EntityKindMeleeMinion, TeamRed, w.width/2+360, w.height/2-70, 14, Stats{
		HP:              420,
		MaxHP:           420,
		Attack:          32,
		PhysicalDefense: 8,
		MagicDefense:    4,
		MoveSpeed:       3,
		AttackRange:     70,
		AttackSpeed:     0.8,
	})
	w.spawnUnit("minion:red-ranged-1", EntityKindRangedMinion, TeamRed, w.width/2+430, w.height/2, 13, Stats{
		HP:              300,
		MaxHP:           300,
		Attack:          38,
		PhysicalDefense: 5,
		MagicDefense:    5,
		MoveSpeed:       3,
		AttackRange:     360,
		AttackSpeed:     0.7,
	})
	w.spawnUnit("minion:red-siege-1", EntityKindSiegeMinion, TeamRed, w.width/2+500, w.height/2+80, 18, Stats{
		HP:              680,
		MaxHP:           680,
		Attack:          62,
		PhysicalDefense: 14,
		MagicDefense:    8,
		MoveSpeed:       2.4,
		AttackRange:     430,
		AttackSpeed:     0.55,
	})
	w.spawnUnit("structure:red-tower-1", EntityKindTower, TeamRed, w.width/2+700, w.height/2-240, 34, Stats{
		HP:              2600,
		MaxHP:           2600,
		Attack:          180,
		PhysicalDefense: 80,
		MagicDefense:    60,
		AttackRange:     620,
		AttackSpeed:     0.75,
	})
	w.spawnUnit("structure:red-barracks-1", EntityKindBarracks, TeamRed, w.width/2+760, w.height/2+80, 40, Stats{
		HP:              3200,
		MaxHP:           3200,
		PhysicalDefense: 55,
		MagicDefense:    45,
	})
	w.spawnUnit("structure:red-crystal", EntityKindCrystal, TeamRed, w.width/2+900, w.height/2+260, 48, Stats{
		HP:              4500,
		MaxHP:           4500,
		PhysicalDefense: 70,
		MagicDefense:    70,
	})
}

func (w *World) SpawnTrainingDummy() {
	w.spawnDummy("dummy:training-1", w.width/2+180, w.height/2)
	w.spawnDummy("dummy:training-2", w.width/2+180, w.height/2+200)
}

func (w *World) spawnDummy(id string, x float64, y float64) {
	w.spawnUnit(id, EntityKindDummy, TeamNeutral, x, y, 28, Stats{
		HP:              3000,
		MaxHP:           3000,
		PhysicalDefense: 10,
		MagicDefense:    10,
	})
}

func (w *World) SpawnObject(kind EntityKind, team Team, x float64, y float64) (string, bool) {
	if team != TeamBlue && team != TeamRed && team != TeamNeutral {
		team = TeamNeutral
	}
	stats, radius, ok := unitTemplate(kind)
	if !ok {
		return "", false
	}
	if kind == EntityKindDummy {
		team = TeamNeutral
	}
	w.nextObjectID++
	id := "spawn:" + string(kind) + ":" + strconv.Itoa(w.nextObjectID)
	w.spawnUnit(id, kind, team, clamp(x, 0, w.width), clamp(y, 0, w.height), radius, stats)
	return id, true
}

func (w *World) spawnUnit(id string, kind EntityKind, team Team, x float64, y float64, radius float64, stats Stats) {
	if _, ok := w.entities[id]; ok {
		return
	}
	entity := &Entity{
		ID:     id,
		Kind:   kind,
		Team:   team,
		Stats:  stats,
		Radius: radius,
		Position: Vector2{
			X: x,
			Y: y,
		},
	}
	if kind == EntityKindEnemyHero {
		entity.Level = MinHeroLevel
		entity.NextLevelExp = w.nextLevelExp(entity.Level)
	}
	w.entities[id] = entity
}

func unitTemplate(kind EntityKind) (Stats, float64, bool) {
	switch kind {
	case EntityKindDummy:
		return Stats{HP: 3000, MaxHP: 3000, PhysicalDefense: 10, MagicDefense: 10}, 28, true
	case EntityKindEnemyHero:
		return Stats{HP: 1200, MaxHP: 1200, MP: 500, MaxMP: 500, Attack: 82, PhysicalDefense: 26, MagicDefense: 18, MoveSpeed: 4.2, AttackRange: 150, AttackSpeed: 1}, 18, true
	case EntityKindMeleeMinion:
		return Stats{HP: 420, MaxHP: 420, Attack: 32, PhysicalDefense: 8, MagicDefense: 4, MoveSpeed: 3, AttackRange: 70, AttackSpeed: 0.8}, 14, true
	case EntityKindRangedMinion:
		return Stats{HP: 300, MaxHP: 300, Attack: 38, PhysicalDefense: 5, MagicDefense: 5, MoveSpeed: 3, AttackRange: 360, AttackSpeed: 0.7}, 13, true
	case EntityKindSiegeMinion:
		return Stats{HP: 680, MaxHP: 680, Attack: 62, PhysicalDefense: 14, MagicDefense: 8, MoveSpeed: 2.4, AttackRange: 430, AttackSpeed: 0.55}, 18, true
	case EntityKindTower:
		return Stats{HP: 2600, MaxHP: 2600, Attack: 180, PhysicalDefense: 80, MagicDefense: 60, AttackRange: 620, AttackSpeed: 0.75}, 34, true
	case EntityKindBarracks:
		return Stats{HP: 3200, MaxHP: 3200, PhysicalDefense: 55, MagicDefense: 45}, 40, true
	case EntityKindCrystal:
		return Stats{HP: 4500, MaxHP: 4500, PhysicalDefense: 70, MagicDefense: 70}, 48, true
	default:
		return Stats{}, 0, false
	}
}

func (w *World) RemovePlayer(playerID string) {
	delete(w.entities, playerID)
	delete(w.entities, playerEntityID(playerID))
}

func (w *World) ApplyInput(playerID string, input protocol.PlayerInput, tick uint64, skills *config.SkillStore, tickRate int) {
	entity := w.entities[playerEntityID(playerID)]
	if entity == nil {
		entity = w.entities[playerID]
	}
	if entity == nil {
		return
	}
	if entity.Death.Dead {
		return
	}
	if input.DebugLevelUp {
		w.debugLevelUp(entity)
	}
	if input.UpgradeSkill != nil {
		w.upgradeSkill(entity, input.UpgradeSkill.Slot)
	}
	if tick < entity.Control.AirborneUntilTick || tick < entity.Control.ActionLockedUntilTick {
		return
	}
	if input.Move != nil {
		w.cancelTankRPreparedCast(entity)
		target := Vector2{
			X: clamp(input.Move.TargetX, 0, w.width),
			Y: clamp(input.Move.TargetY, 0, w.height),
		}
		entity.Intent.MoveTarget = &target
		entity.Intent.AttackPausedTill = tick + uint64(tickRate*3)
	}
	if input.Move == nil && (input.MoveX != 0 || input.MoveY != 0) {
		w.cancelTankRPreparedCast(entity)
		dx, dy := normalize(input.MoveX, input.MoveY)
		before := entity.Position
		step := movementStepAtTick(entity, tickRate, tick)
		entity.Position.X += dx * step
		entity.Position.Y += dy * step
		entity.Position.X = clamp(entity.Position.X, 0, w.width)
		entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
		w.chargeSwordIntent(entity, distance(before, entity.Position))
	}
	if input.Attack != nil {
		if input.Attack.Clear {
			entity.Intent.AttackTargetID = ""
		} else if input.Attack.TargetID != "" && tick >= entity.Control.DashUntilTick {
			entity.Intent.AttackTargetID = input.Attack.TargetID
			entity.Intent.AttackPausedTill = 0
			entity.Intent.MoveTarget = nil
		}
	}
	if input.Cast != nil {
		if skills == nil {
			skills = w.skills
		}
		w.applyCast(entity, *input.Cast, tick, skills, tickRate)
	}
}

func (w *World) Tick(tick uint64, tickRate int) {
	w.expireWindWalls(tick)
	w.expireSkillEffects(tick)
	w.tickProjectiles(tick, tickRate)
	for _, entity := range w.entities {
		w.tickPhysicalDefenseShred(entity, tick)
		w.tickSwordShield(entity, tick)
		w.tickTankGraniteShield(entity, tick, tickRate)
		w.refreshTankWPassive(entity)
		if entity.Kind != EntityKindPlayer {
			continue
		}
		w.tickDashMovement(entity, tick, tickRate)
		w.tickRespawn(entity, tick)
		if entity.Death.Dead || entity.Stats.HP <= 0 {
			continue
		}
		w.tickWarriorToughness(entity, tick, tickRate)
		w.tickPlayer(entity, tick, tickRate)
		w.tickWarriorJudgment(entity, tick, tickRate)
	}
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

func (w *World) tickPhysicalDefenseShred(entity *Entity, tick uint64) {
	if entity == nil || entity.Combat.PhysicalDefenseShredAmount <= 0 || tick < entity.Combat.PhysicalDefenseShredUntil {
		return
	}
	entity.Stats.PhysicalDefense += entity.Combat.PhysicalDefenseShredAmount
	entity.Combat.PhysicalDefenseShredAmount = 0
	entity.Combat.PhysicalDefenseShredUntil = 0
}

func (w *World) expireWindWalls(tick uint64) {
	for id, wall := range w.windWalls {
		if tick >= wall.ExpiresAt {
			delete(w.windWalls, id)
		}
	}
}

func (w *World) expireSkillEffects(tick uint64) {
	for id, effect := range w.skillEffects {
		if tick >= effect.ExpiresAt {
			delete(w.skillEffects, id)
		}
	}
}

func (w *World) tickProjectiles(tick uint64, tickRate int) {
	for id, projectile := range w.projectiles {
		if tick >= projectile.ExpiresAt || projectile.Traveled >= projectile.Range {
			delete(w.projectiles, id)
			continue
		}
		step := projectile.SpeedPerTick
		remaining := projectile.Range - projectile.Traveled
		if step > remaining {
			step = remaining
		}
		if projectile.SkillID == tankQSkillID {
			updateTrackingProjectileDir(projectile, w.entities[projectile.TargetID])
		}
		projectile.Position.X = clamp(projectile.Position.X+projectile.Dir.X*step, 0, w.width)
		projectile.Position.Y = clamp(projectile.Position.Y+projectile.Dir.Y*step, 0, w.height)
		projectile.Traveled += step
		source := w.entities[projectile.SourceID]
		for _, target := range w.entities {
			if projectile.SkillID == tankQSkillID && target.ID != projectile.TargetID {
				continue
			}
			if projectile.HitIDs[target.ID] || !canAttackTarget(source, target) {
				continue
			}
			if distance(projectile.Position, target.Position) > projectile.Radius+target.Radius {
				continue
			}
			projectile.HitIDs[target.ID] = true
			damage := projectile.Damage
			if projectile.SkillID == swordQSkillID && source != nil {
				damage = w.swordQDamage(source, target, w.skillConfig(projectile.SkillID), tick)
			} else if projectile.SkillID == tankQSkillID && source != nil {
				damage = tankQDamage(source, target, w.skillConfig(projectile.SkillID), projectile.Damage, tick)
			}
			target.Combat.LastHitTick = tick
			if target.Kind != EntityKindDummy {
				wasAlive := target.Stats.HP > 0
				if projectile.SkillID == tankQSkillID {
					w.applyMagicDamage(source, target, damage, tickRate)
					applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
					delete(w.projectiles, id)
				} else {
					w.applyDamage(source, target, damage, tickRate)
				}
				if projectile.KnockupTicks > 0 {
					target.Control.AirborneUntilTick = tick + projectile.KnockupTicks
				}
				if wasAlive && target.Stats.HP == 0 {
					w.applyKillReward(source, target)
					w.killPlayer(target, tick, tickRate)
					w.removeDeadUnit(target)
				}
			} else {
				target.Combat.LastDamage = damage
				target.Combat.LastDamageType = projectileDamageType(projectile.SkillID)
				if projectile.SkillID == tankQSkillID {
					applyTankQMoveSpeedSteal(source, target, projectile.EffectRatio, tick+projectile.EffectTicks)
					delete(w.projectiles, id)
				}
			}
		}
		if projectile.Traveled >= projectile.Range {
			delete(w.projectiles, id)
		}
	}
}

func (w *World) tickRespawn(entity *Entity, tick uint64) {
	if !entity.Death.Dead || tick < entity.Death.RespawnTick {
		return
	}
	entity.Death = DeathState{}
	entity.Position = w.spawnPosition(entity.Team)
	entity.Stats.HP = entity.Stats.MaxHP
	entity.Stats.MP = entity.Stats.MaxMP
	entity.Intent = IntentState{}
	entity.Control = ControlState{}
	entity.Warrior = WarriorState{}
	entity.Tank = TankState{}
	entity.Passive.Shield = 0
	entity.Passive.MaxShield = 0
	entity.Passive.LastRegenBreakTick = tick
	entity.Passive.NextRegenTick = 0
	w.refreshTankGraniteShield(entity)
	w.refreshTankWPassive(entity)
	entity.Combat.NextAttackTick = tick
}

func (w *World) tickPlayer(entity *Entity, tick uint64, tickRate int) {
	if tick < entity.Control.AirborneUntilTick || tick < entity.Control.ActionLockedUntilTick {
		return
	}
	if tick < entity.Control.DashUntilTick {
		return
	}
	target := w.entities[entity.Intent.AttackTargetID]
	attackPaused := tick < entity.Intent.AttackPausedTill
	if !attackPaused && canAttackTarget(entity, target) {
		if distance(entity.Position, target.Position) <= w.attackReachAtTick(entity, target, tick) {
			w.applyAttack(entity, target, tick, tickRate)
			return
		}
		w.moveToward(entity, target.Position, movementStepAtTick(entity, tickRate, tick), 0)
		return
	}
	if entity.Intent.MoveTarget != nil {
		if w.moveToward(entity, *entity.Intent.MoveTarget, movementStepAtTick(entity, tickRate, tick), 8) {
			entity.Intent.MoveTarget = nil
			if entity.HeroID == tankHeroID && entity.Tank.UnstoppableCastPending {
				w.releasePreparedTankR(entity, tick, tickRate)
			}
		}
	}
}

func movementStep(entity *Entity, tickRate int) float64 {
	return movementStepAtTick(entity, tickRate, 0)
}

func movementStepAtTick(entity *Entity, tickRate int, tick uint64) float64 {
	moveSpeed := EffectiveMoveSpeedAtTick(entity, tick)
	if tickRate <= 0 {
		return moveSpeed
	}
	return moveSpeed / float64(tickRate)
}

func EffectiveMoveSpeedAtTick(entity *Entity, tick uint64) float64 {
	if entity == nil {
		return 0
	}
	moveSpeed := entity.Stats.MoveSpeed
	if entity.HeroID == warriorHeroID && tick > 0 && tick < entity.Warrior.DecisiveStrikeSpeedUntilTick {
		moveSpeed *= 1 + entity.Warrior.DecisiveStrikeMoveSpeedBonus
	}
	if entity.Control.MoveSpeedBonusUntil > 0 && (tick == 0 || tick < entity.Control.MoveSpeedBonusUntil) {
		moveSpeed += entity.Control.MoveSpeedBonusFlat
	}
	if entity.Control.MoveSpeedSlowUntil > 0 && (tick == 0 || tick < entity.Control.MoveSpeedSlowUntil) {
		moveSpeed *= 1 - clamp(entity.Control.MoveSpeedSlow, 0, 1)
	}
	return moveSpeed
}

func EffectiveAttackSpeedAtTick(entity *Entity, tick uint64) float64 {
	if entity == nil {
		return 0
	}
	attackSpeed := entity.Stats.AttackSpeed
	if entity.Control.AttackSpeedSlowUntil > 0 && (tick == 0 || tick < entity.Control.AttackSpeedSlowUntil) {
		attackSpeed *= 1 - clamp(entity.Control.AttackSpeedSlow, 0, 1)
	}
	return clamp(attackSpeed, 0, 2.5)
}

func (w *World) moveToward(entity *Entity, destination Vector2, step float64, stopDistance float64) bool {
	dx := destination.X - entity.Position.X
	dy := destination.Y - entity.Position.Y
	dist := math.Hypot(dx, dy)
	if dist <= stopDistance {
		return true
	}
	if dist <= step+stopDistance {
		ratio := math.Max(0, dist-stopDistance) / dist
		before := entity.Position
		entity.Position.X += dx * ratio
		entity.Position.Y += dy * ratio
		w.chargeSwordIntent(entity, distance(before, entity.Position))
		return true
	}
	before := entity.Position
	entity.Position.X += dx / dist * step
	entity.Position.Y += dy / dist * step
	entity.Position.X = clamp(entity.Position.X, 0, w.width)
	entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
	w.chargeSwordIntent(entity, distance(before, entity.Position))
	return false
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

func (w *World) WindWalls() []WindWall {
	walls := make([]WindWall, 0, len(w.windWalls))
	for _, wall := range w.windWalls {
		walls = append(walls, wall)
	}
	return walls
}

func (w *World) SkillEffects() []SkillEffect {
	effects := make([]SkillEffect, 0, len(w.projectiles)+len(w.skillEffects))
	for _, effect := range w.skillEffects {
		effects = append(effects, effect)
	}
	for _, projectile := range w.projectiles {
		start := projectile.Start
		createdAt := projectile.CreatedAt
		if projectile.SkillID == tankQSkillID {
			start = projectile.Position
			createdAt = 0
		}
		effects = append(effects, SkillEffect{
			ID:        projectile.ID,
			Kind:      projectile.Kind,
			Team:      projectile.Team,
			Start:     start,
			Dir:       projectile.Dir,
			Range:     projectile.Range,
			Radius:    projectile.Radius,
			Speed:     projectile.SpeedPerTick,
			CreatedAt: createdAt,
			ExpiresAt: projectile.ExpiresAt,
		})
	}
	return effects
}

func updateTrackingProjectileDir(projectile *Projectile, target *Entity) {
	if projectile == nil || target == nil || target.Stats.HP <= 0 {
		return
	}
	dx, dy := normalize(target.Position.X-projectile.Position.X, target.Position.Y-projectile.Position.Y)
	if dx == 0 && dy == 0 {
		return
	}
	projectile.Dir = Vector2{X: dx, Y: dy}
}

func projectileDamageType(skillID string) string {
	if skillID == tankQSkillID {
		return "magic"
	}
	return "physical"
}

func (w *World) BlocksProjectile(team Team, from Vector2, to Vector2) bool {
	for _, wall := range w.windWalls {
		if wall.Team == team {
			continue
		}
		if segmentsIntersect(from, to, windWallStart(wall), windWallEnd(wall)) {
			return true
		}
	}
	return false
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

func normalize(x float64, y float64) (float64, float64) {
	length := math.Hypot(x, y)
	if length == 0 {
		return 0, 0
	}
	return x / length, y / length
}

func clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func playerEntityID(playerID string) string {
	return "player:" + playerID
}

func heroStatsAtLevel(hero config.HeroConfig, level int) Stats {
	level = clampInt(level, MinHeroLevel, MaxHeroLevel)
	growthSteps := level - MinHeroLevel
	growthStepValue := float64(growthSteps)
	bonusHP := hero.Base.BonusHP + hero.Growth.BonusHP*growthSteps
	hp := hero.Base.HP + hero.Growth.HP*growthSteps + bonusHP
	mp := hero.Base.MP + hero.Growth.MP*growthStepValue
	baseAttackSpeed := hero.Base.AttackSpeed * (1 + hero.Growth.AttackSpeed*growthStepValue)
	attackSpeedRatio := hero.Base.AttackSpeedRatio
	attackSpeedBonus := hero.Base.BonusAttackSpeed
	attackSpeedSlow := hero.Base.AttackSpeedSlow
	attackSpeed := finalAttackSpeed(baseAttackSpeed, attackSpeedBonus, attackSpeedRatio, attackSpeedSlow)
	return Stats{
		HP:                   hp,
		MaxHP:                hp,
		BonusHP:              bonusHP,
		MP:                   mp,
		MaxMP:                mp,
		HPRegen5:             hero.Base.HPRegen5 + hero.Growth.HPRegen5*growthStepValue,
		MPRegen5:             hero.Base.MPRegen5 + hero.Growth.MPRegen5*growthStepValue,
		Attack:               hero.Base.Attack + hero.Growth.Attack*growthStepValue,
		BonusAttack:          hero.Base.BonusAttack + hero.Growth.BonusAttack*growthStepValue,
		AbilityPower:         hero.Base.AbilityPower + hero.Growth.AbilityPower*growthSteps,
		DamageReduce:         hero.Base.DamageReduce + hero.Growth.DamageReduce*growthStepValue,
		PhysicalDefense:      hero.Base.PhysicalDefense + hero.Growth.PhysicalDefense*growthStepValue,
		BonusPhysicalDefense: hero.Base.BonusPhysicalDefense + hero.Growth.BonusPhysicalDefense*growthStepValue,
		PhysicalPenPercent:   hero.Base.PhysicalPenPercent + hero.Growth.PhysicalPenPercent*growthStepValue,
		PhysicalPenFlat:      hero.Base.PhysicalPenFlat + hero.Growth.PhysicalPenFlat*growthStepValue,
		PhysicalDamageReduce: hero.Base.PhysicalDamageReduce + hero.Growth.PhysicalDamageReduce*growthStepValue,
		MagicDefense:         hero.Base.MagicDefense + hero.Growth.MagicDefense*growthStepValue,
		BonusMagicDefense:    hero.Base.BonusMagicDefense + hero.Growth.BonusMagicDefense*growthStepValue,
		MagicPenPercent:      hero.Base.MagicPenPercent + hero.Growth.MagicPenPercent*growthStepValue,
		MagicPenFlat:         hero.Base.MagicPenFlat + hero.Growth.MagicPenFlat*growthStepValue,
		MagicDamageReduce:    hero.Base.MagicDamageReduce + hero.Growth.MagicDamageReduce*growthStepValue,
		MoveSpeed:            hero.Base.MoveSpeed + hero.Growth.MoveSpeed*growthStepValue,
		AttackRange:          hero.Base.AttackRange + hero.Growth.AttackRange*growthStepValue,
		AttackSpeed:          attackSpeed,
		BaseAttackSpeed:      baseAttackSpeed,
		AttackSpeedBonus:     attackSpeedBonus,
		AttackSpeedRatio:     attackSpeedRatio,
		AttackSpeedSlow:      attackSpeedSlow,
		CritChance:           hero.Base.CritChance + hero.Growth.CritChance*growthStepValue,
	}
}

func finalAttackSpeed(baseAttackSpeed float64, attackSpeedBonus float64, attackSpeedRatio float64, attackSpeedSlow float64) float64 {
	if baseAttackSpeed < 0 {
		baseAttackSpeed = 0
	}
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	if attackSpeedRatio < 0 {
		attackSpeedRatio = 0
	}
	if attackSpeedSlow < 0 {
		attackSpeedSlow = 0
	}
	if attackSpeedSlow > 1 {
		attackSpeedSlow = 1
	}
	attackSpeed := baseAttackSpeed * (1 + attackSpeedBonus*attackSpeedRatio) * (1 - attackSpeedSlow)
	return clamp(attackSpeed, 0, 2.5)
}

func (w *World) passiveStateForHero(hero config.HeroConfig) PassiveState {
	if hero.HeroID == swordHeroID {
		skill := w.skillConfig(hero.Skills.Passive)
		return PassiveState{
			MaxSwordIntent: skillMetaRange(skill, "intentMax", 100),
		}
	}
	if hero.HeroID == tankHeroID {
		stats := heroStatsAtLevel(hero, MinHeroLevel)
		shield := tankGraniteShieldValue(stats.MaxHP, w.skillConfig(hero.Skills.Passive))
		return PassiveState{
			Shield:    shield,
			MaxShield: shield,
		}
	}
	return PassiveState{}
}

func swordStateForHero(heroID string) SwordState {
	if heroID != swordHeroID {
		return SwordState{}
	}
	return SwordState{
		SweepingBladeTargetUntil: make(map[string]uint64),
	}
}

func (w *World) chargeSwordIntent(entity *Entity, moved float64) {
	if entity == nil || entity.HeroID != swordHeroID || moved <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	if entity.Passive.MaxSwordIntent <= 0 {
		entity.Passive.MaxSwordIntent = skillMetaRange(skill, "intentMax", 100)
	}
	if entity.Passive.SwordIntent >= entity.Passive.MaxSwordIntent {
		return
	}
	moveUnitsPerPercent := skillMetaCurveByLevel(skill, "intentMoveUnitsPerPercent", "intentMoveUnitLevels", entity.Level, 59)
	if moveUnitsPerPercent <= 0 {
		moveUnitsPerPercent = 59
	}
	entity.Passive.SwordIntent += moved / moveUnitsPerPercent
	if entity.Passive.SwordIntent > entity.Passive.MaxSwordIntent {
		entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	}
}

func (w *World) tickSwordShield(entity *Entity, tick uint64) {
	if entity == nil || entity.Passive.Shield <= 0 || entity.Passive.ShieldExpireTick == 0 {
		return
	}
	if tick < entity.Passive.ShieldExpireTick {
		return
	}
	entity.Passive.Shield = 0
	entity.Passive.MaxShield = 0
	entity.Passive.ShieldExpireTick = 0
}

func (w *World) tickTankGraniteShield(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || entity.Stats.HP <= 0 {
		return
	}
	skill := w.heroPassiveSkill(entity)
	maxShield := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	if maxShield <= 0 || entity.Passive.Shield >= maxShield {
		entity.Passive.MaxShield = maxShield
		return
	}
	resetTicks := secondsToTicks(skillMetaRange(skill, "resetSeconds", 10), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+resetTicks {
		return
	}
	entity.Passive.MaxShield = maxShield
	entity.Passive.Shield = maxShield
}

func tankGraniteShieldValue(maxHP int, skill config.SkillConfig) int {
	return int(math.Round(float64(maxHP) * skillMetaRange(skill, "shieldMaxHPRatio", 0.1)))
}

func (w *World) tickWarriorToughness(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || entity.Stats.HP <= 0 || entity.Stats.HP >= entity.Stats.MaxHP {
		return
	}
	skill := w.heroPassiveSkill(entity)
	outOfCombatTicks := secondsToTicks(skillMetaRange(skill, "outOfCombatSeconds", 8), tickRate)
	if tick < entity.Passive.LastRegenBreakTick+outOfCombatTicks {
		return
	}
	intervalTicks := secondsToTicks(skillMetaRange(skill, "regenIntervalSeconds", 5), tickRate)
	if intervalTicks == 0 {
		intervalTicks = uint64(tickRate * 5)
	}
	if entity.Passive.NextRegenTick == 0 {
		entity.Passive.NextRegenTick = tick
	}
	if tick < entity.Passive.NextRegenTick {
		return
	}
	ratio := warriorToughnessRegenRatio(entity.Level, skill)
	heal := int(math.Round(float64(entity.Stats.MaxHP) * ratio))
	if heal < 1 {
		heal = 1
	}
	entity.Stats.HP += heal
	if entity.Stats.HP > entity.Stats.MaxHP {
		entity.Stats.HP = entity.Stats.MaxHP
	}
	entity.Passive.NextRegenTick = tick + intervalTicks
}

func warriorToughnessRegenRatio(level int, skill config.SkillConfig) float64 {
	return skillMetaListByLevel(skill, "regenMaxHPRatio", level, []float64{
		0.015, 0.0198, 0.0246, 0.0294, 0.0342, 0.039,
		0.0438, 0.0486, 0.0534, 0.0582, 0.063, 0.0678,
		0.0726, 0.0774, 0.0822, 0.087, 0.0918, 0.101,
	})
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (w *World) spawnPosition(team Team) Vector2 {
	if team == TeamRed {
		return Vector2{
			X: w.width/2 + 160,
			Y: w.height/2 - 160,
		}
	}
	return Vector2{
		X: w.width/2 - 160,
		Y: w.height/2 + 160,
	}
}

func (w *World) applyCast(entity *Entity, cast protocol.CastInput, tick uint64, skills *config.SkillStore, tickRate int) {
	if tick < entity.Control.SilencedUntilTick {
		return
	}
	state, ok := entity.Skills[cast.SkillID]
	if !ok {
		return
	}
	if state.Level <= 0 {
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorESkillID && tick < entity.Warrior.JudgmentUntilTick {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.stopWarriorE(entity, state, skill, tick, tickRate)
		return
	}
	if tick < state.CooldownUntilTick {
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordQSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordQ(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordWSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordW(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordESkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordE(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == swordHeroID && cast.SkillID == swordRSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		w.applySwordR(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorQSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorQ(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorWSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorW(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorESkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorE(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == warriorHeroID && cast.SkillID == warriorRSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyWarriorR(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankQSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankQ(entity, cast, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankWSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankW(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankESkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankE(entity, state, skill, tick, tickRate)
		return
	}
	if entity.HeroID == tankHeroID && cast.SkillID == tankRSkillID {
		var skill config.SkillConfig
		if skills != nil {
			skill, _ = skills.Get(cast.SkillID)
		}
		if skill.SkillID == "" {
			skill = w.skillConfig(cast.SkillID)
		}
		w.applyTankR(entity, cast, state, skill, tick, tickRate)
		return
	}
	if skills == nil {
		return
	}
	skill, ok := skills.Get(cast.SkillID)
	if !ok {
		return
	}
	w.lockAttackAfterCast(entity, tick, tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skill.CooldownMS, tickRate)
	entity.Skills[cast.SkillID] = state
}

func (w *World) applyWarriorQ(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.DecisiveStrikeUntilTick = tick + secondsToTicks(skillMetaRange(skill, "empowerDurationSeconds", 4.5), tickRate)
	entity.Warrior.DecisiveStrikeSpeedUntilTick = tick + secondsToTicks(skillMetaListByLevel(skill, "moveSpeedDurationSeconds", state.Level, []float64{1.5, 2, 2.5, 3, 3.5}), tickRate)
	entity.Warrior.DecisiveStrikeLevel = state.Level
	entity.Warrior.DecisiveStrikeMoveSpeedBonus = skillMetaRange(skill, "moveSpeedBonus", 0.3)
	entity.Combat.NextAttackTick = tick
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000}), tickRate)
	entity.Skills[warriorQSkillID] = state
}

func (w *World) applyTankQ(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	target := w.entities[cast.TargetID]
	if target == nil {
		target = w.tankQTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	}
	if !canAttackTarget(entity, target) {
		return
	}
	if distance(entity.Position, target.Position) > skillRange(skill, 625)+target.Radius {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{70, 75, 80, 85, 90})
	if entity.Stats.MP < manaCost {
		return
	}
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	entity.Stats.MP -= manaCost
	qRange := skillRange(skill, 625)
	speedPerSecond := skillMetaRange(skill, "projectileSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:tank_q:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "tank_q",
		Team:         entity.Team,
		SourceID:     entity.ID,
		TargetID:     target.ID,
		SkillID:      tankQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       skillMetaRange(skill, "projectileRadius", 45),
		Damage:       state.Level,
		EffectRatio:  skillMetaListByLevel(skill, "moveSpeedSteal", state.Level, []float64{0.2, 0.25, 0.3, 0.35, 0.4}),
		EffectTicks:  secondsToTicks(skillMetaRange(skill, "moveSpeedStealSeconds", 3), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{8000, 8000, 8000, 8000, 8000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[tankQSkillID] = state
}

func (w *World) tankQTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 625)
	pickPadding := skillMetaRange(skill, "targetPickPadding", 90)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) > castRange+target.Radius {
			continue
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+pickPadding {
			continue
		}
		if distToPoint < bestDistance {
			bestDistance = distToPoint
			best = target
		}
	}
	return best
}

func (w *World) applyTankW(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{30, 35, 40, 45, 50})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	entity.Tank.ThunderclapEmpowerUntil = tick + secondsToTicks(skillMetaRange(skill, "aftershockDurationSeconds", 5), tickRate)
	entity.Tank.ThunderclapAftershockUntil = entity.Tank.ThunderclapEmpowerUntil
	entity.Tank.ThunderclapLevel = state.Level
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{10000, 9500, 9000, 8500, 8000}), tickRate)
	entity.Skills[tankWSkillID] = state
}

func (w *World) refreshTankWPassive(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	if entity.Tank.ThunderclapArmorBonus != 0 {
		entity.Stats.PhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Stats.BonusPhysicalDefense -= entity.Tank.ThunderclapArmorBonus
		entity.Tank.ThunderclapArmorBonus = 0
	}
	state, ok := entity.Skills[tankWSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.skillConfig(tankWSkillID)
	ratio := skillMetaListByLevel(skill, "passiveArmorRatio", state.Level, []float64{0.1, 0.15, 0.2, 0.25, 0.3})
	if entity.Passive.Shield > 0 {
		ratio = skillMetaRange(skill, "shieldArmorRatio", 0.3)
	}
	baseArmor := entity.Stats.PhysicalDefense - entity.Stats.BonusPhysicalDefense
	bonus := baseArmor * ratio
	entity.Tank.ThunderclapArmorBonus = bonus
	entity.Stats.PhysicalDefense += bonus
	entity.Stats.BonusPhysicalDefense += bonus
}

func (w *World) applyTankE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaListByLevel(skill, "manaCost", state.Level, []float64{50, 55, 60, 65, 70})
	if entity.Stats.MP < manaCost {
		return
	}
	entity.Stats.MP -= manaCost
	damage := tankEDamage(entity, skill, state.Level)
	slow := skillMetaListByLevel(skill, "attackSpeedSlow", state.Level, []float64{0.3, 0.35, 0.4, 0.45, 0.5})
	slowUntil := tick + secondsToTicks(skillMetaRange(skill, "attackSpeedSlowSeconds", 3), tickRate)
	for _, target := range w.targetsInRadius(entity, entity.Position, skillRange(skill, 400)) {
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyMagicDamage(entity, target, magicDamageAfterResistance(entity, target, damage, tick), tickRate)
			applyAttackSpeedSlow(target, slow, slowUntil)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			applyAttackSpeedSlow(target, slow, slowUntil)
		}
	}
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{7000, 7000, 7000, 7000, 7000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[tankESkillID] = state
}

func tankEDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{60, 95, 130, 165, 200}) +
		entity.Stats.PhysicalDefense*skillMetaRange(skill, "armorRatio", 0.4) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
}

func applyAttackSpeedSlow(target *Entity, slow float64, until uint64) {
	if target == nil || slow <= 0 || until == 0 {
		return
	}
	target.Control.AttackSpeedSlow = clamp(slow, 0, 1)
	target.Control.AttackSpeedSlowUntil = until
}

func (w *World) applyTankR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	start := entity.Position
	rRange := skillRange(skill, 1000)
	targetPoint := Vector2{
		X: clamp(cast.TargetX, 0, w.width),
		Y: clamp(cast.TargetY, 0, w.height),
	}
	if distance(start, targetPoint) > rRange {
		dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
		if dx == 0 && dy == 0 {
			return
		}
		castPosition := Vector2{
			X: clamp(targetPoint.X-dx*rRange, 0, w.width),
			Y: clamp(targetPoint.Y-dy*rRange, 0, w.height),
		}
		entity.Tank.UnstoppableCastPending = true
		entity.Tank.UnstoppableCastTarget = targetPoint
		entity.Tank.UnstoppableCastLevel = state.Level
		entity.Intent.MoveTarget = &castPosition
		entity.Intent.AttackTargetID = ""
		entity.Intent.AttackPausedTill = 0
		return
	}
	w.startTankRDash(entity, targetPoint, state, skill, tick, tickRate)
}

func (w *World) startTankRDash(entity *Entity, targetPoint Vector2, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	start := entity.Position
	dx, dy := normalize(targetPoint.X-start.X, targetPoint.Y-start.Y)
	if dx == 0 && dy == 0 {
		return
	}
	manaCost := skillMetaRange(skill, "manaCost", 100)
	if entity.Stats.MP < manaCost {
		return
	}
	rRange := skillRange(skill, 1000)
	if distance(start, targetPoint) > rRange {
		targetPoint = Vector2{
			X: clamp(start.X+dx*rRange, 0, w.width),
			Y: clamp(start.Y+dy*rRange, 0, w.height),
		}
	}
	entity.Stats.MP -= manaCost
	entity.Intent = IntentState{}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
	landingRadius := skillMetaRange(skill, "landingRadius", 250)
	dashSpeed := skillMetaRange(skill, "dashSpeed", 1600)
	if dashSpeed <= 0 {
		dashSpeed = rRange
	}
	travelTicks := uint64(math.Ceil(distance(start, targetPoint) / dashSpeed * float64(tickRate)))
	if travelTicks < 1 {
		travelTicks = 1
	}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = start
	entity.Control.DashEnd = targetPoint
	entity.Control.DashUntilTick = tick + travelTicks
	entity.Control.ActionLockedUntilTick = entity.Control.DashUntilTick
	entity.Tank.UnstoppableImpactPending = true
	entity.Tank.UnstoppableImpactTick = entity.Control.DashUntilTick
	entity.Tank.UnstoppableImpactLevel = state.Level
	entity.Tank.UnstoppableImpactRadius = landingRadius
	entity.Tank.UnstoppableKnockupTicks = secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1.5), tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{130000, 105000, 80000}), tickRate)
	entity.Skills[tankRSkillID] = state
}

func (w *World) releasePreparedTankR(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	state := entity.Skills[tankRSkillID]
	if state.Level <= 0 {
		w.cancelTankRPreparedCast(entity)
		return
	}
	if state.CooldownUntilTick > tick {
		w.cancelTankRPreparedCast(entity)
		return
	}
	state.Level = entity.Tank.UnstoppableCastLevel
	if state.Level <= 0 {
		state.Level = 1
	}
	w.startTankRDash(entity, entity.Tank.UnstoppableCastTarget, state, w.skillConfig(tankRSkillID), tick, tickRate)
}

func (w *World) cancelTankRPreparedCast(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableCastPending {
		return
	}
	entity.Tank.UnstoppableCastPending = false
	entity.Tank.UnstoppableCastTarget = Vector2{}
	entity.Tank.UnstoppableCastLevel = 0
}

func (w *World) resolveTankRImpact(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != tankHeroID || !entity.Tank.UnstoppableImpactPending {
		return
	}
	if tick < entity.Tank.UnstoppableImpactTick {
		return
	}
	skill := w.skillConfig(tankRSkillID)
	level := entity.Tank.UnstoppableImpactLevel
	if level <= 0 {
		level = 1
	}
	radius := entity.Tank.UnstoppableImpactRadius
	if radius <= 0 {
		radius = skillMetaRange(skill, "landingRadius", 250)
	}
	knockupTicks := entity.Tank.UnstoppableKnockupTicks
	if knockupTicks == 0 {
		knockupTicks = secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1.5), tickRate)
	}
	damage := tankRDamage(entity, skill, level)
	w.addTankRImpactEffect(entity, entity.Position, radius, tick, tickRate)
	for _, target := range w.targetsInRadius(entity, entity.Position, radius) {
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyMagicDamage(entity, target, magicDamageAfterResistance(entity, target, damage, tick), tickRate)
			target.Control.AirborneUntilTick = tick + controlTicksAfterTenacity(target, knockupTicks, tick)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = magicDamageAfterResistance(entity, target, damage, tick)
			target.Combat.LastDamageType = "magic"
			target.Control.AirborneUntilTick = tick + knockupTicks
		}
	}
	entity.Tank.UnstoppableImpactPending = false
	entity.Tank.UnstoppableImpactTick = 0
	entity.Tank.UnstoppableImpactLevel = 0
	entity.Tank.UnstoppableImpactRadius = 0
	entity.Tank.UnstoppableKnockupTicks = 0
}

func tankRDamage(entity *Entity, skill config.SkillConfig, level int) float64 {
	return skillMetaListByLevel(skill, "baseDamage", level, []float64{200, 300, 400}) +
		float64(entity.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.8)
}

func (w *World) addTankRImpactEffect(entity *Entity, center Vector2, radius float64, tick uint64, tickRate int) {
	if entity == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.35))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.nextEffectID++
	id := "effect:tank_r_impact:" + strconv.Itoa(w.nextEffectID)
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "tank_r_impact",
		Team:      entity.Team,
		Start:     center,
		Radius:    radius,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	}
}

func (w *World) applyWarriorW(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	entity.Warrior.CourageUntilTick = tick + secondsToTicks(skillMetaRange(skill, "durationSeconds", 4), tickRate)
	entity.Warrior.CourageFrontUntilTick = tick + secondsToTicks(skillMetaRange(skill, "frontDurationSeconds", 0.75), tickRate)
	entity.Warrior.CourageFrontDamageReduce = skillMetaRange(skill, "frontDamageReduce", 0.6)
	entity.Warrior.CourageFrontTenacity = skillMetaRange(skill, "frontTenacity", 0.6)
	entity.Warrior.CourageBackDamageReduce = skillMetaRange(skill, "backDamageReduce", 0.3)
	entity.Control.TenacityUntilTick = entity.Warrior.CourageFrontUntilTick
	entity.Passive.MaxShield = warriorWShieldValue(entity, skill, state.Level)
	entity.Passive.Shield = entity.Passive.MaxShield
	entity.Passive.ShieldExpireTick = entity.Warrior.CourageUntilTick
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{24000, 22000, 20000, 18000, 16000}), tickRate)
	entity.Skills[warriorWSkillID] = state
}

func (w *World) applyWarriorE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	durationTicks := secondsToTicks(skillMetaRange(skill, "durationSeconds", 3), tickRate)
	spins := warriorESpinCount(entity, skill)
	if durationTicks == 0 || spins <= 0 {
		return
	}
	intervalTicks := durationTicks / uint64(spins)
	if intervalTicks < 1 {
		intervalTicks = 1
	}
	entity.Warrior.JudgmentUntilTick = tick + durationTicks
	entity.Warrior.JudgmentNextSpinTick = tick
	entity.Warrior.JudgmentSpinIntervalTicks = intervalTicks
	entity.Warrior.JudgmentSpinsRemaining = spins
	entity.Warrior.JudgmentLevel = state.Level
	entity.Warrior.JudgmentHits = make(map[string]int)
	state.CooldownUntilTick = 0
	entity.Skills[warriorESkillID] = state
}

func (w *World) stopWarriorE(entity *Entity, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || tick >= entity.Warrior.JudgmentUntilTick {
		return
	}
	remainingTicks := entity.Warrior.JudgmentUntilTick - tick
	cooldownTicks := cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000}), tickRate)
	if cooldownTicks > remainingTicks {
		state.CooldownUntilTick = tick + cooldownTicks - remainingTicks
	} else {
		state.CooldownUntilTick = tick
	}
	clearWarriorE(entity)
	entity.Skills[warriorESkillID] = state
}

func (w *World) tickWarriorJudgment(entity *Entity, tick uint64, tickRate int) {
	if entity == nil || entity.HeroID != warriorHeroID || entity.Warrior.JudgmentUntilTick == 0 {
		return
	}
	skill := w.skillConfig(warriorESkillID)
	if tick >= entity.Warrior.JudgmentUntilTick {
		w.finishWarriorE(entity, skill, tick, tickRate)
		return
	}
	if entity.Warrior.JudgmentSpinsRemaining <= 0 {
		return
	}
	if tick < entity.Warrior.JudgmentNextSpinTick {
		return
	}
	if skill.SkillID == "" {
		return
	}
	w.applyWarriorESpin(entity, skill, tick, tickRate)
	entity.Warrior.JudgmentSpinsRemaining--
	if entity.Warrior.JudgmentSpinsRemaining <= 0 {
		entity.Warrior.JudgmentNextSpinTick = 0
		return
	}
	entity.Warrior.JudgmentNextSpinTick += entity.Warrior.JudgmentSpinIntervalTicks
}

func (w *World) finishWarriorE(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	state := entity.Skills[warriorESkillID]
	if skill.SkillID != "" {
		state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{9000, 8250, 7500, 6750, 6000}), tickRate)
		entity.Skills[warriorESkillID] = state
	}
	clearWarriorE(entity)
}

func clearWarriorE(entity *Entity) {
	entity.Warrior.JudgmentUntilTick = 0
	entity.Warrior.JudgmentNextSpinTick = 0
	entity.Warrior.JudgmentSpinIntervalTicks = 0
	entity.Warrior.JudgmentSpinsRemaining = 0
	entity.Warrior.JudgmentLevel = 0
	entity.Warrior.JudgmentHits = nil
}

func (w *World) applyWarriorESpin(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	targets := w.warriorETargets(entity, skill)
	if len(targets) == 0 {
		return
	}
	nearest := nearestEntity(entity, targets)
	for _, target := range targets {
		damage := w.warriorEDamage(entity, target, skill, entity.Warrior.JudgmentLevel, target == nearest, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyDamage(entity, target, damage, tickRate)
			w.recordWarriorEHit(entity, target, skill, tick, tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
			w.recordWarriorEHit(entity, target, skill, tick, tickRate)
		}
	}
}

func (w *World) warriorETargets(entity *Entity, skill config.SkillConfig) []*Entity {
	targets := []*Entity{}
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) <= skillRange(skill, 180)+target.Radius {
			targets = append(targets, target)
		}
	}
	return targets
}

func nearestEntity(source *Entity, targets []*Entity) *Entity {
	var nearest *Entity
	nearestDistance := math.MaxFloat64
	for _, target := range targets {
		dist := distance(source.Position, target.Position)
		if dist < nearestDistance {
			nearestDistance = dist
			nearest = target
		}
	}
	return nearest
}

func (w *World) warriorEDamage(attacker *Entity, target *Entity, skill config.SkillConfig, level int, nearest bool, tick uint64) int {
	if level <= 0 {
		level = 1
	}
	rawDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{4, 7, 10, 13, 16}) + attacker.Stats.Attack*skillMetaListByLevel(skill, "adRatio", level, []float64{0.4, 0.43, 0.46, 0.49, 0.52})
	if w.attackCrits(attacker, target, tick) {
		rawDamage = skillMetaListByLevel(skill, "critBaseDamage", level, []float64{5.2, 9.1, 13, 16.9, 20.8}) + attacker.Stats.Attack*skillMetaListByLevel(skill, "critAdRatio", level, []float64{0.52, 0.559, 0.598, 0.637, 0.676})
	}
	if nearest {
		rawDamage *= 1 + skillMetaRange(skill, "nearestDamageBonus", 0.25)
	}
	return physicalDamageAfterResistance(attacker, target, rawDamage, tick)
}

func warriorESpinCount(entity *Entity, skill config.SkillConfig) int {
	baseSpins := int(skillMetaRange(skill, "baseSpins", 7))
	if entity == nil {
		return baseSpins
	}
	bonusPerSpin := skillMetaRange(skill, "attackSpeedBonusPerExtraSpin", 0.25)
	if bonusPerSpin <= 0 {
		return baseSpins
	}
	attackSpeedBonus := entity.Stats.AttackSpeedBonus
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	return baseSpins + int(math.Floor(attackSpeedBonus/bonusPerSpin))
}

func (w *World) recordWarriorEHit(attacker *Entity, target *Entity, skill config.SkillConfig, tick uint64, tickRate int) {
	if target == nil || target.Kind != EntityKindPlayer && target.Kind != EntityKindEnemyHero {
		return
	}
	if attacker.Warrior.JudgmentHits == nil {
		attacker.Warrior.JudgmentHits = make(map[string]int)
	}
	attacker.Warrior.JudgmentHits[target.ID]++
	if attacker.Warrior.JudgmentHits[target.ID] != int(skillMetaRange(skill, "armorShredHitCount", 6)) {
		return
	}
	w.applyPhysicalDefenseShred(target, skillMetaRange(skill, "armorShredPercent", 0.25), tick+secondsToTicks(skillMetaRange(skill, "armorShredSeconds", 6), tickRate))
}

func (w *World) applyPhysicalDefenseShred(target *Entity, percent float64, untilTick uint64) {
	if target == nil || percent <= 0 {
		return
	}
	if target.Combat.PhysicalDefenseShredAmount > 0 {
		target.Stats.PhysicalDefense += target.Combat.PhysicalDefenseShredAmount
	}
	shred := target.Stats.PhysicalDefense * clamp(percent, 0, 1)
	target.Stats.PhysicalDefense -= shred
	if target.Stats.PhysicalDefense < 0 {
		target.Stats.PhysicalDefense = 0
	}
	target.Combat.PhysicalDefenseShredAmount = shred
	target.Combat.PhysicalDefenseShredUntil = untilTick
}

func (w *World) applyWarriorR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity == nil || state.Level <= 0 {
		return
	}
	target := w.warriorRTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	damage := warriorRDamage(target, skill, state.Level)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyTrueDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(entity, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = trueDamageAfterReduction(target, damage, tick)
		target.Combat.LastDamageType = "true"
	}
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{120000, 100000, 80000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[warriorRSkillID] = state
}

func (w *World) warriorRTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	castRange := skillRange(skill, 400)
	pickPadding := skillMetaRange(skill, "targetPickPadding", 80)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) > castRange+target.Radius {
			continue
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+pickPadding {
			continue
		}
		if distToPoint < bestDistance {
			bestDistance = distToPoint
			best = target
		}
	}
	return best
}

func warriorRDamage(target *Entity, skill config.SkillConfig, level int) float64 {
	if target == nil {
		return 0
	}
	baseDamage := skillMetaListByLevel(skill, "baseDamage", level, []float64{150, 250, 350})
	missingHPRatio := skillMetaListByLevel(skill, "missingHPRatio", level, []float64{0.25, 0.3, 0.35})
	missingHP := target.Stats.MaxHP - target.Stats.HP
	if missingHP < 0 {
		missingHP = 0
	}
	return baseDamage + float64(missingHP)*missingHPRatio
}

func (w *World) lockAttackAfterCast(entity *Entity, tick uint64, tickRate int) {
	nextAttackTick := tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(entity, tick), tickRate)
	if entity.Combat.NextAttackTick < nextAttackTick {
		entity.Combat.NextAttackTick = nextAttackTick
	}
}

func (w *World) applySwordQ(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if tick > state.StacksExpireTick {
		state.Stacks = 0
	}
	hasWhirlwindStack := state.Stacks >= 2
	form := "line"
	qRange := skillRange(skill, 475)
	if swordEQWindowActive(entity, skill, tick, tickRate) {
		form = "circle"
		qRange = skillMetaRange(skill, "eqRadius", 375)
	} else if hasWhirlwindStack {
		form = "whirlwind"
		qRange = skillMetaRange(skill, "whirlwindRange", 900)
	}
	if form == "whirlwind" {
		w.spawnSwordWhirlwind(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, qRange, skill, tick, tickRate)
		w.lockAttackAfterCast(entity, tick, tickRate)
		state.Stacks = 0
		state.StacksExpireTick = 0
		state.CooldownUntilTick = tick + w.swordQCooldownTicks(entity, skill, state.Level, tickRate)
		entity.Skills[swordQSkillID] = state
		return
	}
	targets := w.swordQTargets(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, qRange, form, skill)
	for _, target := range targets {
		damage := w.swordQDamage(entity, target, skill, tick)
		target.Combat.LastHitTick = tick
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyDamage(entity, target, damage, tickRate)
			if form == "circle" && hasWhirlwindStack {
				target.Control.AirborneUntilTick = tick + secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1), tickRate)
			}
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(entity, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = damage
			target.Combat.LastDamageType = "physical"
		}
	}
	if len(targets) > 0 {
		if form == "circle" && hasWhirlwindStack {
			state.Stacks = 0
			state.StacksExpireTick = 0
		} else {
			state.Stacks++
			if state.Stacks > 2 {
				state.Stacks = 2
			}
			state.StacksExpireTick = tick + secondsToTicks(skillMetaRange(skill, "stackDurationSeconds", swordQStackTicks), tickRate)
		}
	}
	state.CooldownUntilTick = tick + w.swordQCooldownTicks(entity, skill, state.Level, tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordQSkillID] = state
}

func swordEQWindowActive(entity *Entity, skill config.SkillConfig, tick uint64, tickRate int) bool {
	if entity == nil || tick >= entity.Control.DashUntilTick {
		return false
	}
	windowTicks := secondsToTicks(skillMetaRange(skill, "eqWindowSeconds", 0.15), tickRate)
	if windowTicks == 0 {
		return false
	}
	return entity.Control.DashUntilTick-tick <= windowTicks
}

func (w *World) applySwordW(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(cast.TargetX-entity.Position.X, cast.TargetY-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	w.nextWallID++
	id := "wind_wall:" + strconv.Itoa(w.nextWallID)
	width := skillMetaListByLevel(skill, "width", state.Level, []float64{300, 350, 400, 450, 500})
	placeDistance := skillMetaRange(skill, "placeDistance", 180)
	center := Vector2{
		X: clamp(entity.Position.X+dx*placeDistance, 0, w.width),
		Y: clamp(entity.Position.Y+dy*placeDistance, 0, w.height),
	}
	w.windWalls[id] = WindWall{
		ID:        id,
		Team:      entity.Team,
		Center:    center,
		Dir:       Vector2{X: -dy, Y: dx},
		Width:     width,
		ExpiresAt: tick + secondsToTicks(skillMetaRange(skill, "durationSeconds", windWallDuration), tickRate),
	}
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{26000, 24000, 22000, 20000, 18000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordWSkillID] = state
}

func (w *World) spawnSwordWhirlwind(entity *Entity, targetPoint Vector2, qRange float64, skill config.SkillConfig, tick uint64, tickRate int) {
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	radius := skillMetaRange(skill, "whirlwindRadius", 70)
	speedPerSecond := skillMetaRange(skill, "whirlwindSpeed", 1200)
	speedPerTick := speedPerSecond / float64(tickRate)
	if speedPerTick <= 0 {
		speedPerTick = qRange
	}
	lifeTicks := uint64(math.Ceil(qRange / speedPerTick))
	if lifeTicks == 0 {
		lifeTicks = 1
	}
	w.nextProjectileID++
	id := "projectile:sword_whirlwind:" + strconv.Itoa(w.nextProjectileID)
	w.projectiles[id] = &Projectile{
		ID:           id,
		Kind:         "sword_whirlwind",
		Team:         entity.Team,
		SourceID:     entity.ID,
		SkillID:      swordQSkillID,
		Position:     entity.Position,
		Start:        entity.Position,
		Dir:          Vector2{X: dx, Y: dy},
		SpeedPerTick: speedPerTick,
		Range:        qRange,
		Radius:       radius,
		Damage:       w.swordQDamage(entity, &Entity{ID: id}, skill, tick),
		KnockupTicks: secondsToTicks(skillMetaRange(skill, "knockupSeconds", 1), tickRate),
		CreatedAt:    tick,
		ExpiresAt:    tick + lifeTicks + 1,
		HitIDs:       make(map[string]bool),
	}
}

func (w *World) applySwordE(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	if entity.Sword.SweepingBladeTargetUntil == nil {
		entity.Sword.SweepingBladeTargetUntil = make(map[string]uint64)
	}
	target := w.swordETarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill)
	if target == nil {
		return
	}
	if tick < entity.Sword.SweepingBladeTargetUntil[target.ID] {
		return
	}
	damage := swordEDamage(entity, target, skill, state.Level, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyMagicDamage(entity, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(entity, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "magic"
	}
	entity.Sword.SweepingBladeStacks++
	maxStacks := int(skillMetaRange(skill, "maxStacks", 4))
	if entity.Sword.SweepingBladeStacks > maxStacks {
		entity.Sword.SweepingBladeStacks = maxStacks
	}
	targetCooldownMS := skillMetaListByLevelMS(skill, "targetCooldownMs", state.Level, []float64{10000, 9000, 8000, 7000, 6000})
	entity.Sword.SweepingBladeTargetUntil[target.ID] = tick + cooldownTicks(targetCooldownMS, tickRate)
	dx, dy := normalize(target.Position.X-entity.Position.X, target.Position.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	dashEnd := Vector2{
		X: clamp(target.Position.X+dx*(target.Radius+entity.Radius+skillMetaRange(skill, "dashThroughDistance", 34)), 0, w.width),
		Y: clamp(target.Position.Y+dy*(target.Radius+entity.Radius+skillMetaRange(skill, "dashThroughDistance", 34)), 0, w.height),
	}
	entity.Intent = IntentState{}
	entity.Control.DashStartTick = tick
	entity.Control.DashStart = entity.Position
	entity.Control.DashEnd = dashEnd
	entity.Control.DashUntilTick = tick + secondsToTicks(skillMetaRange(skill, "dashDurationSeconds", 0.35), tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{500, 400, 300, 200, 100}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordESkillID] = state
}

func (w *World) applySwordR(entity *Entity, cast protocol.CastInput, state SkillState, skill config.SkillConfig, tick uint64, tickRate int) {
	target := w.swordRTarget(entity, Vector2{X: cast.TargetX, Y: cast.TargetY}, skill, tick)
	if target == nil {
		return
	}
	entity.Position = Vector2{
		X: clamp(target.Position.X-entity.Radius-target.Radius-18, 0, w.width),
		Y: target.Position.Y,
	}
	entity.Intent = IntentState{}
	hits := w.swordRTargets(entity, target.Position, skill, tick)
	for _, hit := range hits {
		damage := swordRDamage(entity, hit, skill, state.Level, tick)
		hit.Combat.LastHitTick = tick
		if hit.Kind != EntityKindDummy {
			wasAlive := hit.Stats.HP > 0
			w.applyDamage(entity, hit, damage, tickRate)
			hit.Control.AirborneUntilTick += secondsToTicks(skillMetaRange(skill, "airborneExtendSeconds", 1), tickRate)
			if wasAlive && hit.Stats.HP == 0 {
				w.applyKillReward(entity, hit)
				w.killPlayer(hit, tick, tickRate)
				w.removeDeadUnit(hit)
			}
		} else {
			hit.Combat.LastDamage = damage
			hit.Combat.LastDamageType = "physical"
		}
	}
	entity.Passive.SwordIntent = entity.Passive.MaxSwordIntent
	entity.Passive.MaxShield = w.swordShieldValue(entity)
	entity.Passive.Shield = entity.Passive.MaxShield
	qState := entity.Skills[swordQSkillID]
	qState.Stacks = int(skillMetaRange(skill, "qStacksAfterCast", 2))
	if qState.Stacks > 0 {
		qState.StacksExpireTick = tick + secondsToTicks(skillMetaRange(skill, "qStackDurationSeconds", swordQStackTicks), tickRate)
	} else {
		qState.StacksExpireTick = 0
	}
	entity.Skills[swordQSkillID] = qState
	entity.Sword.LastBreathUntilTick = tick + secondsToTicks(skillMetaRange(skill, "lastBreathDurationSeconds", 15), tickRate)
	entity.Control.ActionLockedUntilTick = tick + secondsToTicks(skillMetaRange(skill, "selfActionLockSeconds", 1), tickRate)
	state.CooldownUntilTick = tick + cooldownTicks(skillMetaListByLevelMS(skill, "cooldownMs", state.Level, []float64{80000, 55000, 30000}), tickRate)
	w.lockAttackAfterCast(entity, tick, tickRate)
	entity.Skills[swordRSkillID] = state
}

func (w *World) swordRTarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig, tick uint64) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	pickPadding := skillMetaRange(skill, "targetPickPadding", 80)
	for _, target := range w.entities {
		if !isAirborneEnemyHero(entity, target, tick) {
			continue
		}
		dist := distance(targetPoint, target.Position)
		if dist > target.Radius+pickPadding {
			continue
		}
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	}
	if best != nil {
		return best
	}
	castRange := skillRange(skill, 1200)
	for _, target := range w.entities {
		if !isAirborneEnemyHero(entity, target, tick) {
			continue
		}
		dist := distance(entity.Position, target.Position)
		if dist > castRange+target.Radius {
			continue
		}
		if dist < bestDistance {
			best = target
			bestDistance = dist
		}
	}
	return best
}

func (w *World) swordRTargets(entity *Entity, center Vector2, skill config.SkillConfig, tick uint64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !isAirborneEnemyHero(entity, target, tick) {
			continue
		}
		if distance(center, target.Position) <= skillMetaRange(skill, "hitRadius", 450)+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func isAirborneEnemyHero(attacker *Entity, target *Entity, tick uint64) bool {
	if !canAttackTarget(attacker, target) {
		return false
	}
	if target.Control.AirborneUntilTick <= tick {
		return false
	}
	return target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero
}

func (w *World) swordETarget(entity *Entity, targetPoint Vector2, skill config.SkillConfig) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(entity.Position, target.Position) > skillRange(skill, 475)+target.Radius {
			continue
		}
		distToPoint := distance(targetPoint, target.Position)
		if distToPoint > target.Radius+skillMetaRange(skill, "targetPickPadding", 48) {
			continue
		}
		if distToPoint < bestDistance {
			best = target
			bestDistance = distToPoint
		}
	}
	return best
}

func swordEDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{60, 70, 80, 90, 100})
	damageValue := baseDamage + attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 0.2) + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	damageValue *= 1 + float64(attacker.Sword.SweepingBladeStacks)*skillMetaRange(skill, "stackDamageBonus", 0.25)
	return magicDamageAfterResistance(attacker, target, damageValue, tick)
}

func swordRDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{200, 300, 400})
	return physicalDamageAfterResistance(attacker, target, baseDamage+attacker.Stats.BonusAttack*skillMetaRange(skill, "bonusAdRatio", 1.5), tick)
}

func (w *World) swordQTargets(entity *Entity, targetPoint Vector2, qRange float64, form string, skill config.SkillConfig) []*Entity {
	if form == "circle" {
		return w.targetsInRadius(entity, entity.Position, qRange)
	}
	dx, dy := normalize(targetPoint.X-entity.Position.X, targetPoint.Y-entity.Position.Y)
	if dx == 0 && dy == 0 {
		dx = 1
	}
	if form == "whirlwind" {
		return w.targetsAlongMovingCircle(entity, entity.Position, Vector2{X: dx, Y: dy}, qRange, skillMetaRange(skill, "whirlwindRadius", 70))
	}
	width := skillMetaRange(skill, "lineWidth", 55)
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(entity.Position, Vector2{X: dx, Y: dy}, target.Position)
		if along < 0 || along > qRange+target.Radius {
			continue
		}
		if perpendicular > width+target.Radius {
			continue
		}
		hits = append(hits, target)
	}
	return hits
}

func (w *World) targetsAlongMovingCircle(entity *Entity, origin Vector2, direction Vector2, travelRange float64, radius float64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		along, perpendicular := projectPoint(origin, direction, target.Position)
		if along < -target.Radius || along > travelRange+target.Radius {
			continue
		}
		if perpendicular <= radius+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func (w *World) targetsInRadius(entity *Entity, center Vector2, radius float64) []*Entity {
	hits := make([]*Entity, 0)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		if distance(center, target.Position) <= radius+target.Radius {
			hits = append(hits, target)
		}
	}
	return hits
}

func (w *World) targetsInCone(entity *Entity, direction Vector2, coneRange float64, angleDegrees float64) []*Entity {
	hits := make([]*Entity, 0)
	if direction.X == 0 && direction.Y == 0 {
		direction = Vector2{X: 1, Y: 0}
	}
	cosLimit := math.Cos((angleDegrees / 2) * math.Pi / 180)
	for _, target := range w.entities {
		if !canAttackTarget(entity, target) {
			continue
		}
		toTarget := Vector2{X: target.Position.X - entity.Position.X, Y: target.Position.Y - entity.Position.Y}
		dist := math.Hypot(toTarget.X, toTarget.Y)
		if dist > coneRange+target.Radius || dist == 0 {
			continue
		}
		dot := (toTarget.X*direction.X + toTarget.Y*direction.Y) / dist
		if dot >= cosLimit {
			hits = append(hits, target)
		}
	}
	return hits
}

func isMonster(entity *Entity) bool {
	if entity == nil {
		return false
	}
	switch entity.Kind {
	case EntityKindPlayer, EntityKindEnemyHero, EntityKindTower, EntityKindBarracks, EntityKindCrystal, EntityKindDummy:
		return false
	default:
		return entity.Team == TeamNeutral
	}
}

func cooldownTicks(cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 {
		return 0
	}
	ticks := math.Ceil(float64(cooldownMS) / 1000 * float64(tickRate))
	return uint64(ticks)
}

func secondsToTicks(seconds float64, tickRate int) uint64 {
	if seconds <= 0 {
		return 0
	}
	return uint64(math.Ceil(seconds * float64(tickRate)))
}

func skillRange(skill config.SkillConfig, fallback float64) float64 {
	if skill.Range > 0 {
		return skill.Range
	}
	return fallback
}

func skillMetaRange(skill config.SkillConfig, key string, fallback float64) float64 {
	if skill.Meta == nil {
		return fallback
	}
	value, ok := skill.Meta[key]
	if !ok {
		return fallback
	}
	return value
}

func skillMetaListByLevel(skill config.SkillConfig, key string, level int, fallback []float64) float64 {
	values := fallback
	if skill.MetaLists != nil && len(skill.MetaLists[key]) > 0 {
		values = skill.MetaLists[key]
	}
	rank := skillRank(level, len(values))
	return values[rank-1]
}

func skillMetaCurveByLevel(skill config.SkillConfig, valueKey string, levelKey string, level int, fallback float64) float64 {
	if skill.MetaLists == nil {
		return fallback
	}
	values := skill.MetaLists[valueKey]
	levels := skill.MetaLists[levelKey]
	if len(values) == 0 || len(values) != len(levels) {
		return fallback
	}
	currentLevel := float64(clampInt(level, MinHeroLevel, MaxHeroLevel))
	if currentLevel <= levels[0] {
		return values[0]
	}
	last := len(values) - 1
	if currentLevel >= levels[last] {
		return values[last]
	}
	for i := 1; i < len(values); i++ {
		if currentLevel > levels[i] {
			continue
		}
		fromLevel := levels[i-1]
		toLevel := levels[i]
		if toLevel <= fromLevel {
			return values[i]
		}
		t := (currentLevel - fromLevel) / (toLevel - fromLevel)
		return values[i-1] + (values[i]-values[i-1])*t
	}
	return values[last]
}

func skillMetaListByLevelMS(skill config.SkillConfig, key string, level int, fallback []float64) int {
	return int(math.Round(skillMetaListByLevel(skill, key, level, fallback)))
}

func skillRank(level int, count int) int {
	return clampInt(level, 1, count)
}

func (w *World) swordQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, tick uint64) int {
	state := attacker.Skills[swordQSkillID]
	baseDamage := skillMetaListByLevel(skill, "baseDamage", state.Level, []float64{20, 45, 70, 95, 120})
	attack := baseDamage + attacker.Stats.Attack*skillMetaRange(skill, "adRatio", 1)
	if w.attackCrits(attacker, target, tick) {
		attack *= w.critDamageMultiplier(attacker)
	}
	return physicalDamageAfterResistance(attacker, target, attack, tick)
}

func (w *World) swordQCooldownTicks(entity *Entity, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	attackSpeedBonus := 0.0
	if entity != nil {
		attackSpeedBonus = entity.Stats.AttackSpeedBonus
	}
	return swordQCooldownTicksByBonus(attackSpeedBonus, skill, skillLevel, tickRate)
}

func swordQCooldownTicksByBonus(attackSpeedBonus float64, skill config.SkillConfig, skillLevel int, tickRate int) uint64 {
	baseCooldownMS := skillMetaListByLevelMS(skill, "cooldownMs", skillLevel, []float64{6000, 5500, 5000, 4500, 4000})
	if attackSpeedBonus < 0 {
		attackSpeedBonus = 0
	}
	seconds := float64(baseCooldownMS) / 1000 * (1 - attackSpeedBonus*0.6)
	minSeconds := skillMetaRange(skill, "minCooldownSeconds", 1.33)
	if seconds < minSeconds {
		seconds = minSeconds
	}
	return uint64(math.Ceil(seconds*float64(tickRate) - 1e-6))
}

func (w *World) applyAttack(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker.Kind != EntityKindPlayer || tick < attacker.Combat.NextAttackTick {
		return
	}
	if attacker.HeroID == warriorHeroID && tick < attacker.Warrior.JudgmentUntilTick {
		return
	}
	if !canAttackTarget(attacker, target) {
		return
	}
	if distance(attacker.Position, target.Position) > w.attackReachAtTick(attacker, target, tick) {
		return
	}

	damage := w.attackDamage(attacker, target, tick)
	target.Combat.LastHitTick = tick
	if target.Kind != EntityKindDummy {
		wasAlive := target.Stats.HP > 0
		w.applyDamage(attacker, target, damage, tickRate)
		if wasAlive && target.Stats.HP == 0 {
			w.applyKillReward(attacker, target)
			w.killPlayer(target, tick, tickRate)
			w.removeDeadUnit(target)
		}
	} else {
		target.Combat.LastDamage = damage
		target.Combat.LastDamageType = "physical"
	}
	attacker.Combat.NextAttackTick = tick + attackCooldownTicks(EffectiveAttackSpeedAtTick(attacker, tick), tickRate)
	w.applyTankWAftershock(attacker, target, tick, tickRate)
	w.consumeWarriorQ(attacker, target, tick, tickRate)
}

func (w *World) attackDamage(attacker *Entity, target *Entity, tick uint64) int {
	attack := attacker.Stats.Attack
	if w.attackCrits(attacker, target, tick) {
		attack *= w.critDamageMultiplier(attacker)
	}
	return physicalDamageAfterResistance(attacker, target, attack+w.warriorQBonusDamage(attacker, tick)+w.tankWBonusDamage(attacker, tick), tick)
}

func (w *World) warriorQBonusDamage(attacker *Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != warriorHeroID || tick >= attacker.Warrior.DecisiveStrikeUntilTick {
		return 0
	}
	skill := w.skillConfig(warriorQSkillID)
	level := attacker.Warrior.DecisiveStrikeLevel
	if level <= 0 {
		level = 1
	}
	baseDamage := skillMetaListByLevel(skill, "bonusDamage", level, []float64{30, 60, 90, 120, 150})
	return baseDamage + attacker.Stats.Attack*skillMetaRange(skill, "totalAdRatio", 1.4)
}

func (w *World) consumeWarriorQ(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.HeroID != warriorHeroID || tick >= attacker.Warrior.DecisiveStrikeUntilTick {
		return
	}
	skill := w.skillConfig(warriorQSkillID)
	if target != nil {
		silenceTicks := secondsToTicks(skillMetaRange(skill, "silenceSeconds", 1.5), tickRate)
		target.Control.SilencedUntilTick = tick + controlTicksAfterTenacity(target, silenceTicks, tick)
	}
	attacker.Warrior.DecisiveStrikeUntilTick = 0
	attacker.Warrior.DecisiveStrikeLevel = 0
}

func (w *World) tankWBonusDamage(attacker *Entity, tick uint64) float64 {
	if attacker == nil || attacker.HeroID != tankHeroID || tick >= attacker.Tank.ThunderclapEmpowerUntil {
		return 0
	}
	skill := w.skillConfig(tankWSkillID)
	level := attacker.Tank.ThunderclapLevel
	if level <= 0 {
		level = 1
	}
	return skillMetaListByLevel(skill, "bonusDamage", level, []float64{30, 40, 50, 60, 70}) +
		float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.2) +
		attacker.Stats.PhysicalDefense*skillMetaRange(skill, "armorRatio", 0.15)
}

func (w *World) applyTankWAftershock(attacker *Entity, primary *Entity, tick uint64, tickRate int) {
	if attacker == nil || attacker.HeroID != tankHeroID || tick >= attacker.Tank.ThunderclapAftershockUntil {
		return
	}
	skill := w.skillConfig(tankWSkillID)
	level := attacker.Tank.ThunderclapLevel
	if level <= 0 {
		level = 1
	}
	damage := skillMetaListByLevel(skill, "aftershockDamage", level, []float64{15, 25, 35, 45, 55}) +
		float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "aftershockAPRatio", 0.3) +
		attacker.Stats.PhysicalDefense*skillMetaRange(skill, "aftershockArmorRatio", 0.15)
	direction := Vector2{X: 1, Y: 0}
	if primary != nil {
		dx, dy := normalize(primary.Position.X-attacker.Position.X, primary.Position.Y-attacker.Position.Y)
		if dx != 0 || dy != 0 {
			direction = Vector2{X: dx, Y: dy}
		}
	}
	coneRange := skillMetaRange(skill, "aftershockConeRange", 300)
	coneAngle := skillMetaRange(skill, "aftershockConeAngleDegrees", 70)
	w.addTankWAftershockEffect(attacker, direction, coneRange, coneAngle, tick, tickRate)
	for _, target := range w.targetsInCone(attacker, direction, coneRange, coneAngle) {
		target.Combat.LastHitTick = tick
		previousDamage := 0
		if primary != nil && target.ID == primary.ID {
			previousDamage = target.Combat.LastDamage
		}
		aftershockDamage := damage
		if isMonster(target) {
			aftershockDamage *= skillMetaRange(skill, "monsterDamageMultiplier", 1.8)
		}
		if target.Kind != EntityKindDummy {
			wasAlive := target.Stats.HP > 0
			w.applyDamage(attacker, target, physicalDamageAfterResistance(attacker, target, aftershockDamage, tick), tickRate)
			if wasAlive && target.Stats.HP == 0 {
				w.applyKillReward(attacker, target)
				w.killPlayer(target, tick, tickRate)
				w.removeDeadUnit(target)
			}
		} else {
			target.Combat.LastDamage = physicalDamageAfterResistance(attacker, target, aftershockDamage, tick)
			target.Combat.LastDamageType = "physical"
		}
		if previousDamage > 0 {
			target.Combat.LastDamage += previousDamage
		}
	}
	if tick < attacker.Tank.ThunderclapEmpowerUntil {
		attacker.Tank.ThunderclapEmpowerUntil = 0
	}
}

func (w *World) addTankWAftershockEffect(attacker *Entity, direction Vector2, coneRange float64, coneAngle float64, tick uint64, tickRate int) {
	if attacker == nil {
		return
	}
	lifeTicks := uint64(math.Ceil(float64(tickRate) * 0.25))
	if lifeTicks < 1 {
		lifeTicks = 1
	}
	w.nextEffectID++
	id := "effect:tank_w_aftershock:" + strconv.Itoa(w.nextEffectID)
	w.skillEffects[id] = SkillEffect{
		ID:        id,
		Kind:      "tank_w_aftershock",
		Team:      attacker.Team,
		Start:     attacker.Position,
		Dir:       direction,
		Range:     coneRange,
		Radius:    coneAngle,
		CreatedAt: tick,
		ExpiresAt: tick + lifeTicks,
	}
}

func physicalDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	damageReduce := target.damageReductionForType("physical", tick)
	return damageAfterResistance(rawDamage, effectiveResistance(target.Stats.PhysicalDefense, attacker.Stats.PhysicalPenPercent, attacker.Stats.PhysicalPenFlat), damageReduce)
}

func magicDamageAfterResistance(attacker *Entity, target *Entity, rawDamage float64, tick uint64) int {
	damageReduce := target.damageReductionForType("magic", tick)
	return damageAfterResistance(rawDamage, effectiveResistance(target.Stats.MagicDefense, attacker.Stats.MagicPenPercent, attacker.Stats.MagicPenFlat), damageReduce)
}

func tankQDamage(attacker *Entity, target *Entity, skill config.SkillConfig, skillLevel int, tick uint64) int {
	baseDamage := skillMetaListByLevel(skill, "baseDamage", skillLevel, []float64{70, 120, 170, 220, 270})
	rawDamage := baseDamage + float64(attacker.Stats.AbilityPower)*skillMetaRange(skill, "apRatio", 0.6)
	return magicDamageAfterResistance(attacker, target, rawDamage, tick)
}

func applyTankQMoveSpeedSteal(source *Entity, target *Entity, ratio float64, until uint64) {
	if source == nil || target == nil || ratio <= 0 || until == 0 {
		return
	}
	ratio = clamp(ratio, 0, 1)
	stolen := EffectiveMoveSpeedAtTick(target, 0) * ratio
	source.Control.MoveSpeedBonusFlat = stolen
	source.Control.MoveSpeedBonusUntil = until
	target.Control.MoveSpeedSlow = ratio
	target.Control.MoveSpeedSlowUntil = until
}

func trueDamageAfterReduction(target *Entity, rawDamage float64, tick uint64) int {
	return damageAfterResistance(rawDamage, 0, target.damageReductionForType("true", tick))
}

func (entity *Entity) damageReductionForType(damageType string, tick uint64) float64 {
	reductions := []float64{entity.Stats.DamageReduce}
	switch damageType {
	case "physical":
		reductions = append(reductions, entity.Stats.PhysicalDamageReduce)
	case "magic":
		reductions = append(reductions, entity.Stats.MagicDamageReduce)
	}
	if entity.HeroID == warriorHeroID && entity.Warrior.CourageUntilTick > 0 {
		reductions = append(reductions, entity.Warrior.courageDamageReductionAtTick(tick))
	}
	return stackDamageReduction(reductions...)
}

func (entity *Entity) tenacityAtTick(tick uint64) float64 {
	tenacity := []float64{}
	if entity.HeroID == warriorHeroID && entity.Warrior.CourageFrontUntilTick > 0 && (tick == 0 || tick < entity.Warrior.CourageFrontUntilTick) {
		tenacity = append(tenacity, entity.Warrior.CourageFrontTenacity)
	}
	return stackTenacity(tenacity...)
}

func controlTicksAfterTenacity(target *Entity, ticks uint64, tick uint64) uint64 {
	if target == nil || ticks == 0 {
		return ticks
	}
	remainingRatio := 1 - target.tenacityAtTick(tick)
	adjusted := uint64(math.Ceil(float64(ticks) * remainingRatio))
	if adjusted < 1 {
		return 1
	}
	return adjusted
}

func (state WarriorState) courageDamageReductionAtTick(tick uint64) float64 {
	if state.CourageUntilTick == 0 {
		return 0
	}
	if tick > 0 && tick >= state.CourageUntilTick {
		return 0
	}
	if tick == 0 || tick < state.CourageFrontUntilTick {
		return state.CourageFrontDamageReduce
	}
	return state.CourageBackDamageReduce
}

func warriorWShieldValue(entity *Entity, skill config.SkillConfig, skillLevel int) int {
	baseShield := skillMetaListByLevel(skill, "shieldValue", skillLevel, []float64{70, 95, 120, 145, 170})
	return int(math.Round(baseShield + float64(entity.Stats.BonusHP)*skillMetaRange(skill, "bonusHealthRatio", 0.2)))
}

func effectiveResistance(resistance float64, percentPen float64, flatPen float64) float64 {
	if resistance < 0 {
		return resistance
	}
	if percentPen < 0 {
		percentPen = 0
	}
	if percentPen > 1 {
		percentPen = 1
	}
	if flatPen < 0 {
		flatPen = 0
	}
	effective := resistance*(1-percentPen) - flatPen
	if effective < 0 {
		return 0
	}
	return effective
}

func damageAfterResistance(rawDamage float64, resistance float64, damageReduce float64) int {
	if rawDamage <= 0 {
		return 0
	}
	multiplier := 100 / (resistance + 100)
	if resistance < 0 {
		denominator := 100 + resistance
		if denominator < 1 {
			denominator = 1
		}
		multiplier = 100 / denominator
	}
	damageReduce = clamp(damageReduce, 0, 1)
	damage := int(math.Round(rawDamage * multiplier * (1 - damageReduce)))
	if damage < 1 {
		return 1
	}
	return damage
}

func stackDamageReduction(reductions ...float64) float64 {
	multiplier := 1.0
	for _, reduction := range reductions {
		reduction = clamp(reduction, 0, 1)
		multiplier *= 1 - reduction
	}
	return 1 - multiplier
}

func stackTenacity(tenacityValues ...float64) float64 {
	multiplier := 1.0
	for _, tenacity := range tenacityValues {
		tenacity = clamp(tenacity, 0, 1)
		multiplier *= 1 - tenacity
	}
	return 1 - multiplier
}

func (w *World) attackCrits(attacker *Entity, target *Entity, tick uint64) bool {
	chance := w.critChance(attacker)
	if chance <= 0 {
		return false
	}
	if chance >= 1 {
		return true
	}
	roll := deterministicCritRoll(attacker.ID, target.ID, tick)
	return roll < chance
}

func (w *World) critChance(attacker *Entity) float64 {
	chance := attacker.Stats.CritChance
	if attacker.HeroID == swordHeroID {
		chance *= skillMetaRange(w.heroPassiveSkill(attacker), "critChanceMultiplier", 2)
	}
	if chance > 1 {
		return 1
	}
	if chance < 0 {
		return 0
	}
	return chance
}

func (w *World) critDamageMultiplier(attacker *Entity) float64 {
	if attacker.HeroID == swordHeroID {
		return skillMetaRange(w.heroPassiveSkill(attacker), "critDamageMultiplier", 1.9)
	}
	return 2
}

func deterministicCritRoll(attackerID string, targetID string, tick uint64) float64 {
	hash := uint64(1469598103934665603)
	for _, value := range []string{attackerID, targetID, strconv.FormatUint(tick, 10)} {
		for i := 0; i < len(value); i++ {
			hash ^= uint64(value[i])
			hash *= 1099511628211
		}
	}
	return float64(hash%critRollModulo) / critRollModulo
}

func (w *World) applyDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "physical", tickRate)
}

func (w *World) applyMagicDamage(source *Entity, target *Entity, damage int, tickRate int) {
	w.applyResolvedDamage(source, target, damage, "magic", tickRate)
}

func (w *World) applyTrueDamage(source *Entity, target *Entity, rawDamage float64, tickRate int) {
	w.applyResolvedDamage(source, target, trueDamageAfterReduction(target, rawDamage, target.Combat.LastHitTick), "true", tickRate)
}

func (w *World) applyResolvedDamage(source *Entity, target *Entity, damage int, damageType string, tickRate int) {
	if damage <= 0 {
		target.Combat.LastDamage = 0
		target.Combat.LastDamageType = ""
		return
	}
	damage = w.applyShield(source, target, damage, tickRate)
	target.Combat.LastDamage = damage
	target.Combat.LastDamageType = damageType
	w.breakTankGraniteShield(target, target.Combat.LastHitTick)
	if damage <= 0 {
		return
	}
	target.Stats.HP -= damage
	if target.Stats.HP < 0 {
		target.Stats.HP = 0
	}
	w.breakWarriorToughness(source, target, target.Combat.LastHitTick)
}

func (w *World) skillConfig(skillID string) config.SkillConfig {
	if w == nil || w.skills == nil || skillID == "" {
		return config.SkillConfig{}
	}
	skill, _ := w.skills.Get(skillID)
	return skill
}

func (w *World) heroPassiveSkill(entity *Entity) config.SkillConfig {
	if entity == nil || w == nil || w.heroes == nil {
		return config.SkillConfig{}
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return config.SkillConfig{}
	}
	return w.skillConfig(hero.Skills.Passive)
}

func (w *World) breakWarriorToughness(source *Entity, target *Entity, tick uint64) {
	if target == nil || target.HeroID != warriorHeroID || !warriorToughnessBreaksRegen(source) {
		return
	}
	target.Passive.LastRegenBreakTick = tick
	target.Passive.NextRegenTick = 0
}

func (w *World) breakTankGraniteShield(target *Entity, tick uint64) {
	if target == nil || target.HeroID != tankHeroID {
		return
	}
	target.Passive.LastRegenBreakTick = tick
	target.Passive.NextRegenTick = 0
}

func warriorToughnessBreaksRegen(source *Entity) bool {
	if source == nil {
		return false
	}
	switch source.Kind {
	case EntityKindPlayer, EntityKindEnemyHero, EntityKindTower:
		return true
	default:
		return false
	}
}

func (w *World) killPlayer(target *Entity, tick uint64, tickRate int) {
	if target.Kind != EntityKindPlayer || target.Death.Dead {
		return
	}
	target.Death = DeathState{
		Dead:              true,
		RespawnTick:       tick + uint64(respawnSeconds*tickRate),
		RespawnTickRate:   tickRate,
		RespawnSeconds:    respawnSeconds,
		LastDeathPosition: target.Position,
	}
	target.Intent = IntentState{}
	target.Warrior = WarriorState{}
	target.Passive.Shield = 0
	target.Passive.MaxShield = 0
	target.Passive.ShieldExpireTick = 0
	for _, entity := range w.entities {
		if entity.Intent.AttackTargetID == target.ID {
			entity.Intent.AttackTargetID = ""
		}
	}
}

func (w *World) applyShield(source *Entity, target *Entity, damage int, tickRate int) int {
	if target == nil {
		return damage
	}
	if target.HeroID == swordHeroID && target.Passive.Shield <= 0 && target.Passive.SwordIntent >= target.Passive.MaxSwordIntent && swordShieldTriggers(source) {
		skill := w.heroPassiveSkill(target)
		target.Passive.MaxShield = w.swordShieldValue(target)
		target.Passive.Shield = target.Passive.MaxShield
		target.Passive.ShieldExpireTick = target.Combat.LastHitTick + secondsToTicks(skillMetaRange(skill, "shieldDurationSeconds", 1), tickRate)
		target.Passive.SwordIntent = 0
	}
	if target.Passive.Shield <= 0 {
		return damage
	}
	absorbed := damage
	if absorbed > target.Passive.Shield {
		absorbed = target.Passive.Shield
	}
	target.Passive.Shield -= absorbed
	return damage - absorbed
}

func swordShieldTriggers(source *Entity) bool {
	if source == nil {
		return false
	}
	return source.Kind == EntityKindPlayer || source.Kind == EntityKindEnemyHero
}

func (w *World) swordShieldValue(entity *Entity) int {
	level := clampInt(entity.Level, MinHeroLevel, MaxHeroLevel)
	skill := w.heroPassiveSkill(entity)
	return int(math.Round(skillMetaCurveByLevel(skill, "shieldValue", "shieldValueLevels", level, 125)))
}

func (w *World) removeDeadUnit(target *Entity) {
	if target.Kind == EntityKindPlayer || target.Kind == EntityKindDummy {
		return
	}
	delete(w.entities, target.ID)
	for _, entity := range w.entities {
		if entity.Intent.AttackTargetID == target.ID {
			entity.Intent.AttackTargetID = ""
		}
	}
}

func (w *World) applyKillReward(killer *Entity, target *Entity) {
	if killer == nil || target == nil || w.rewards == nil {
		return
	}
	w.applyWarriorWPassiveKill(killer, target)
	switch target.Kind {
	case EntityKindMeleeMinion, EntityKindRangedMinion, EntityKindSiegeMinion:
		if exp, ok := w.rewards.MinionExp(string(target.Kind), 1); ok {
			w.addExperience(killer, exp)
		}
	case EntityKindTower:
		if exp, ok := w.rewards.StructureTeamExp(string(target.Kind)); ok {
			w.addTeamExperience(killer.Team, float64(exp))
		}
	case EntityKindEnemyHero, EntityKindPlayer:
		targetNextExp := w.nextLevelExp(target.Level)
		if targetNextExp > 0 {
			w.addExperience(killer, w.rewards.HeroKillExp(int(targetNextExp), killer.Level, target.Level))
		}
	}
}

func (w *World) applyWarriorWPassiveKill(killer *Entity, target *Entity) {
	if killer == nil || target == nil || killer.HeroID != warriorHeroID {
		return
	}
	state, ok := killer.Skills[warriorWSkillID]
	if !ok || state.Level <= 0 {
		return
	}
	skill := w.skillConfig(warriorWSkillID)
	gain := skillMetaRange(skill, "passiveMinionResistGain", 0.2)
	if target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero {
		gain = skillMetaRange(skill, "passiveHeroResistGain", 1)
	}
	maxGain := skillMetaRange(skill, "passiveMaxResistGain", 40)
	before := killer.Warrior.CouragePassiveResistGain
	after := before + gain
	if after > maxGain {
		after = maxGain
	}
	if after <= before {
		return
	}
	delta := after - before
	killer.Warrior.CouragePassiveResistGain = after
	killer.Stats.PhysicalDefense += delta
	killer.Stats.BonusPhysicalDefense += delta
	killer.Stats.MagicDefense += delta
	killer.Stats.BonusMagicDefense += delta
	killer.Skills[warriorWSkillID] = state
}

func (w *World) addTeamExperience(team Team, exp float64) {
	for _, entity := range w.entities {
		if entity.Kind == EntityKindPlayer && entity.Team == team && entity.Stats.HP > 0 {
			w.addExperience(entity, exp)
		}
	}
}

func (w *World) addExperience(entity *Entity, exp float64) {
	if entity == nil || entity.Kind != EntityKindPlayer || exp <= 0 || entity.Level >= MaxHeroLevel {
		return
	}
	entity.Exp += exp
	entity.TotalExp += exp
	for entity.Level < MaxHeroLevel {
		nextExp := w.nextLevelExp(entity.Level)
		if nextExp <= 0 || entity.Exp < nextExp {
			break
		}
		entity.Exp -= nextExp
		w.levelUp(entity)
	}
	entity.NextLevelExp = w.nextLevelExp(entity.Level)
}

func (w *World) levelUp(entity *Entity) {
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return
	}
	oldMaxHP := entity.Stats.MaxHP
	oldMaxMP := entity.Stats.MaxMP
	entity.Level = clampInt(entity.Level+1, MinHeroLevel, MaxHeroLevel)
	nextStats := heroStatsAtLevel(hero, entity.Level)
	hpGain := nextStats.MaxHP - oldMaxHP
	mpGain := nextStats.MaxMP - oldMaxMP
	nextStats.HP = entity.Stats.HP + hpGain
	nextStats.MP = entity.Stats.MP + mpGain
	if nextStats.HP > nextStats.MaxHP {
		nextStats.HP = nextStats.MaxHP
	}
	if nextStats.MP > nextStats.MaxMP {
		nextStats.MP = nextStats.MaxMP
	}
	entity.Tank.ThunderclapArmorBonus = 0
	entity.Stats = nextStats
	w.refreshTankGraniteShieldMax(entity)
	w.refreshTankWPassive(entity)
	entity.SkillPoints++
}

func (w *World) refreshTankGraniteShield(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	skill := w.heroPassiveSkill(entity)
	shield := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	entity.Passive.MaxShield = shield
	entity.Passive.Shield = shield
	entity.Passive.ShieldExpireTick = 0
}

func (w *World) refreshTankGraniteShieldMax(entity *Entity) {
	if entity == nil || entity.HeroID != tankHeroID {
		return
	}
	oldMax := entity.Passive.MaxShield
	skill := w.heroPassiveSkill(entity)
	nextMax := tankGraniteShieldValue(entity.Stats.MaxHP, skill)
	entity.Passive.MaxShield = nextMax
	if entity.Passive.Shield >= oldMax {
		entity.Passive.Shield = nextMax
	}
	if entity.Passive.Shield > nextMax {
		entity.Passive.Shield = nextMax
	}
}

func (w *World) debugLevelUp(entity *Entity) {
	if entity == nil || entity.Kind != EntityKindPlayer || entity.Level >= MaxHeroLevel {
		return
	}
	w.levelUp(entity)
	entity.NextLevelExp = w.nextLevelExp(entity.Level)
	if entity.Level >= MaxHeroLevel {
		entity.Exp = 0
	}
}

func (w *World) nextLevelExp(level int) float64 {
	if w.levels == nil {
		return 0
	}
	nextExp, ok := w.levels.NextExp(level)
	if !ok {
		return 0
	}
	return float64(nextExp)
}

func (w *World) upgradeSkill(entity *Entity, slot string) {
	if entity == nil || entity.SkillPoints <= 0 {
		return
	}
	hero, ok := w.heroes.Get(entity.HeroID)
	if !ok {
		return
	}
	skillID := hero.Skills.SkillIDForSlot(slot)
	if skillID == "" {
		return
	}
	state, ok := entity.Skills[skillID]
	if !ok {
		return
	}
	maxLevel := maxSkillLevel(slot)
	if maxLevel <= 0 || state.Level >= maxLevel {
		return
	}
	state.Level++
	entity.SkillPoints--
	entity.Skills[skillID] = state
	w.refreshTankWPassive(entity)
}

func maxSkillLevel(slot string) int {
	switch slot {
	case "q", "w", "e":
		return MaxBasicSkillLevel
	case "r":
		return MaxUltSkillLevel
	default:
		return 0
	}
}

func canAttackTarget(attacker *Entity, target *Entity) bool {
	if attacker == nil || target == nil {
		return false
	}
	if target.Stats.HP <= 0 {
		return false
	}
	if target.Kind == EntityKindPlayer && target.Death.Dead {
		return false
	}
	if target.ID == attacker.ID || target.Team == attacker.Team {
		return false
	}
	return true
}

func attackReach(attacker *Entity, target *Entity) float64 {
	return attacker.Stats.AttackRange + attacker.Radius + target.Radius
}

func (w *World) attackReachAtTick(attacker *Entity, target *Entity, tick uint64) float64 {
	attackRange := attacker.Stats.AttackRange
	if attacker.HeroID == warriorHeroID && tick < attacker.Warrior.DecisiveStrikeUntilTick {
		attackRange = math.Max(attackRange, skillRange(w.skillConfig(warriorQSkillID), 300))
	}
	return attackRange + attacker.Radius + target.Radius
}

func attackCooldownTicks(attackSpeed float64, tickRate int) uint64 {
	if attackSpeed <= 0 {
		return uint64(tickRate)
	}
	ticks := math.Ceil(float64(tickRate) / attackSpeed)
	if ticks < 1 {
		return 1
	}
	return uint64(ticks)
}

func distance(a Vector2, b Vector2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func projectPoint(origin Vector2, direction Vector2, point Vector2) (float64, float64) {
	dx := point.X - origin.X
	dy := point.Y - origin.Y
	along := dx*direction.X + dy*direction.Y
	perpX := dx - along*direction.X
	perpY := dy - along*direction.Y
	return along, math.Hypot(perpX, perpY)
}

func windWallStart(wall WindWall) Vector2 {
	half := wall.Width / 2
	return Vector2{
		X: wall.Center.X - wall.Dir.X*half,
		Y: wall.Center.Y - wall.Dir.Y*half,
	}
}

func windWallEnd(wall WindWall) Vector2 {
	half := wall.Width / 2
	return Vector2{
		X: wall.Center.X + wall.Dir.X*half,
		Y: wall.Center.Y + wall.Dir.Y*half,
	}
}

func segmentsIntersect(a Vector2, b Vector2, c Vector2, d Vector2) bool {
	ab1 := orientation(a, b, c)
	ab2 := orientation(a, b, d)
	cd1 := orientation(c, d, a)
	cd2 := orientation(c, d, b)
	return ab1*ab2 <= 0 && cd1*cd2 <= 0
}

func orientation(a Vector2, b Vector2, c Vector2) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}
