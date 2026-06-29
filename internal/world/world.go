package world

import (
	"math"

	"l-battle/internal/config"
	"l-battle/internal/protocol"
)

const (
	DefaultMapWidth  = 2000.0
	DefaultMapHeight = 2000.0
	MinHeroLevel     = 1
	MaxHeroLevel     = 18
)

type World struct {
	width    float64
	height   float64
	entities map[string]*Entity
}

func NewWorld() *World {
	w := &World{
		width:    DefaultMapWidth,
		height:   DefaultMapHeight,
		entities: make(map[string]*Entity),
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
	if entity := w.entities[entityID]; entity != nil {
		entity.Team = team
		entity.HeroID = hero.HeroID
		entity.Level = level
		entity.Stats = stats
		entity.Radius = hero.Radius
		entity.Skills = skills
		entity.Position = position
		entity.Combat = CombatState{}
		return
	}
	w.entities[entityID] = &Entity{
		ID:       entityID,
		Kind:     EntityKindPlayer,
		Team:     team,
		PlayerID: playerID,
		HeroID:   hero.HeroID,
		Level:    level,
		Stats:    stats,
		Radius:   hero.Radius,
		Skills:   skills,
		Position: position,
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

func (w *World) spawnUnit(id string, kind EntityKind, team Team, x float64, y float64, radius float64, stats Stats) {
	if _, ok := w.entities[id]; ok {
		return
	}
	w.entities[id] = &Entity{
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
	if input.Move != nil {
		target := Vector2{
			X: clamp(input.Move.TargetX, 0, w.width),
			Y: clamp(input.Move.TargetY, 0, w.height),
		}
		entity.Intent.MoveTarget = &target
		entity.Intent.AttackPausedTill = tick + uint64(tickRate*3)
	}
	if input.Move == nil && (input.MoveX != 0 || input.MoveY != 0) {
		dx, dy := normalize(input.MoveX, input.MoveY)
		entity.Position.X += dx * entity.Stats.MoveSpeed
		entity.Position.Y += dy * entity.Stats.MoveSpeed
		entity.Position.X = clamp(entity.Position.X, 0, w.width)
		entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
	}
	if input.Attack != nil {
		if input.Attack.Clear {
			entity.Intent.AttackTargetID = ""
		} else if input.Attack.TargetID != "" {
			entity.Intent.AttackTargetID = input.Attack.TargetID
			entity.Intent.AttackPausedTill = 0
			entity.Intent.MoveTarget = nil
		}
	}
	if input.Cast != nil {
		w.applyCast(entity, *input.Cast, tick, skills, tickRate)
	}
}

func (w *World) Tick(tick uint64, tickRate int) {
	for _, entity := range w.entities {
		if entity.Kind != EntityKindPlayer || entity.Stats.HP <= 0 {
			continue
		}
		w.tickPlayer(entity, tick, tickRate)
	}
}

func (w *World) tickPlayer(entity *Entity, tick uint64, tickRate int) {
	target := w.entities[entity.Intent.AttackTargetID]
	attackPaused := tick < entity.Intent.AttackPausedTill
	if !attackPaused && canAttackTarget(entity, target) {
		if distance(entity.Position, target.Position) <= attackReach(entity, target) {
			w.applyAttack(entity, target, tick, tickRate)
			return
		}
		w.moveToward(entity, target.Position, 0)
		return
	}
	if entity.Intent.MoveTarget != nil {
		if w.moveToward(entity, *entity.Intent.MoveTarget, 8) {
			entity.Intent.MoveTarget = nil
		}
	}
}

func (w *World) moveToward(entity *Entity, destination Vector2, stopDistance float64) bool {
	dx := destination.X - entity.Position.X
	dy := destination.Y - entity.Position.Y
	dist := math.Hypot(dx, dy)
	if dist <= stopDistance {
		return true
	}
	step := entity.Stats.MoveSpeed
	if dist <= step+stopDistance {
		ratio := math.Max(0, dist-stopDistance) / dist
		entity.Position.X += dx * ratio
		entity.Position.Y += dy * ratio
		return true
	}
	entity.Position.X += dx / dist * step
	entity.Position.Y += dy / dist * step
	entity.Position.X = clamp(entity.Position.X, 0, w.width)
	entity.Position.Y = clamp(entity.Position.Y, 0, w.height)
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
	hp := hero.Base.HP + hero.Growth.HP*growthSteps
	mp := hero.Base.MP + hero.Growth.MP*growthSteps
	return Stats{
		HP:              hp,
		MaxHP:           hp,
		MP:              mp,
		MaxMP:           mp,
		Attack:          hero.Base.Attack + hero.Growth.Attack*growthSteps,
		PhysicalDefense: hero.Base.PhysicalDefense + hero.Growth.PhysicalDefense*growthSteps,
		MagicDefense:    hero.Base.MagicDefense + hero.Growth.MagicDefense*growthSteps,
		MoveSpeed:       hero.Base.MoveSpeed + hero.Growth.MoveSpeed*float64(growthSteps),
		AttackRange:     hero.Base.AttackRange + hero.Growth.AttackRange*float64(growthSteps),
		AttackSpeed:     hero.Base.AttackSpeed + hero.Growth.AttackSpeed*float64(growthSteps),
	}
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
	state, ok := entity.Skills[cast.SkillID]
	if !ok {
		return
	}
	if tick < state.CooldownUntilTick {
		return
	}
	skill, ok := skills.Get(cast.SkillID)
	if !ok {
		return
	}
	state.CooldownUntilTick = tick + cooldownTicks(skill.CooldownMS, tickRate)
	entity.Skills[cast.SkillID] = state
}

func cooldownTicks(cooldownMS int, tickRate int) uint64 {
	if cooldownMS <= 0 {
		return 0
	}
	ticks := math.Ceil(float64(cooldownMS) / 1000 * float64(tickRate))
	return uint64(ticks)
}

func (w *World) applyAttack(attacker *Entity, target *Entity, tick uint64, tickRate int) {
	if attacker.Kind != EntityKindPlayer || tick < attacker.Combat.NextAttackTick {
		return
	}
	if !canAttackTarget(attacker, target) {
		return
	}
	if distance(attacker.Position, target.Position) > attackReach(attacker, target) {
		return
	}

	damage := attacker.Stats.Attack - target.Stats.PhysicalDefense
	if damage < 1 {
		damage = 1
	}
	target.Combat.LastHitTick = tick
	target.Combat.LastDamage = damage
	if target.Kind != EntityKindDummy {
		target.Stats.HP -= damage
		if target.Stats.HP < 0 {
			target.Stats.HP = 0
		}
	}
	attacker.Combat.NextAttackTick = tick + attackCooldownTicks(attacker.Stats.AttackSpeed, tickRate)
}

func canAttackTarget(attacker *Entity, target *Entity) bool {
	if attacker == nil || target == nil {
		return false
	}
	if target.Stats.HP <= 0 {
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
