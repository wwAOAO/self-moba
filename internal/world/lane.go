package world

import (
	"math"
	"strconv"
)

const (
	minionWaveIntervalSeconds = 30
	minionWaveCount           = 7
	minionSpawnGapSeconds     = 0.55
	laneMinionAggroRange      = 450
	laneMinionReturnDistance  = 500
	laneMinionReturnSeconds   = 5
	laneMinionMoveSpeed       = 260
	laneMinionAvoidLookahead  = 120
)

func (w *World) tickMinionWaves(tick uint64, tickRate int) {
	if tickRate <= 0 {
		return
	}
	interval := secondsToTicks(minionWaveIntervalSeconds, tickRate)
	if w.nextMinionWaveTick == 0 {
		w.nextMinionWaveTick = interval
	}
	if tick >= w.nextMinionWaveTick {
		w.queueMinionWave(TeamBlue, tick, tickRate)
		w.queueMinionWave(TeamRed, tick, tickRate)
		w.nextMinionWaveTick = tick + interval
	}
	w.spawnDueMinions(tick)
}

func (w *World) queueMinionWave(team Team, tick uint64, tickRate int) {
	w.minionWaveNumber++
	gap := secondsToTicks(minionSpawnGapSeconds, tickRate)
	for i, kind := range minionWaveKinds() {
		w.pendingMinionSpawns = append(w.pendingMinionSpawns, PendingMinionSpawn{
			Team: team, Kind: kind, Index: i, WaveNumber: w.minionWaveNumber, SpawnTick: tick + uint64(i)*gap,
		})
	}
}

func (w *World) spawnMinionWave(team Team, tick uint64) {
	w.minionWaveNumber++
	for i, kind := range minionWaveKinds() {
		w.spawnLaneMinion(team, kind, i, w.minionWaveNumber, tick)
	}
}

func minionWaveKinds() []EntityKind {
	return []EntityKind{
		EntityKindMeleeMinion,
		EntityKindMeleeMinion,
		EntityKindMeleeMinion,
		EntityKindRangedMinion,
		EntityKindRangedMinion,
		EntityKindRangedMinion,
		EntityKindSiegeMinion,
	}
}

func (w *World) spawnDueMinions(tick uint64) {
	kept := make([]PendingMinionSpawn, 0, len(w.pendingMinionSpawns))
	for _, spawn := range w.pendingMinionSpawns {
		if tick < spawn.SpawnTick {
			kept = append(kept, spawn)
			continue
		}
		w.spawnLaneMinion(spawn.Team, spawn.Kind, spawn.Index, spawn.WaveNumber, tick)
	}
	w.pendingMinionSpawns = kept
}

func (w *World) spawnLaneMinion(team Team, kind EntityKind, index int, waveNumber int, tick uint64) {
	start := w.spawnPosition(team)
	target := w.spawnPosition(oppositeTeam(team))
	dx, dy := normalize(target.X-start.X, target.Y-start.Y)
	stats, radius, ok := unitTemplate(kind)
	if !ok {
		return
	}
	applyMinionGrowth(&stats, kind, tick)
	stats.MoveSpeed = laneMinionMoveSpeed
	id := "spawn:lane_minion:" + string(team) + ":" + strconv.Itoa(waveNumber) + ":" + strconv.Itoa(index+1)
	forwardOffset := 190 - float64(index)*12
	pos := Vector2{X: clamp(start.X+dx*forwardOffset, 0, w.width), Y: clamp(start.Y+dy*forwardOffset, 0, w.height)}
	w.spawnUnit(id, kind, team, pos.X, pos.Y, radius, stats)
	if entity := w.entities[id]; entity != nil {
		entity.Lane = LaneState{Active: true, RouteTarget: target, LastOnLaneTick: tick}
	}
}

func applyMinionGrowth(stats *Stats, kind EntityKind, tick uint64) {
	steps := int(tick / (uint64(minionWaveIntervalSeconds) * 6))
	if steps <= 0 {
		return
	}
	growthSteps := float64(steps)
	switch kind {
	case EntityKindMeleeMinion:
		stats.MaxHP = min(stats.MaxHP+growthSteps*20, 3000)
		stats.HP = stats.MaxHP
		stats.Attack = min(stats.Attack+float64(steps), 160)
		stats.PhysicalDefense = min(stats.PhysicalDefense+float64(steps*2), 40)
		stats.MagicDefense = min(stats.MagicDefense+float64(steps)*1.25, 25)
	case EntityKindRangedMinion:
		stats.MaxHP = min(stats.MaxHP+growthSteps*14, 1200)
		stats.HP = stats.MaxHP
		stats.Attack = min(stats.Attack+float64(steps)*2, 250)
	case EntityKindSiegeMinion:
		stats.MaxHP = min(stats.MaxHP+growthSteps*27, 5400)
		stats.HP = stats.MaxHP
		stats.Attack = min(stats.Attack+float64(steps)*1.5, 261)
		stats.PhysicalDefense = min(stats.PhysicalDefense+float64(steps*2), 120)
		stats.MagicDefense = min(stats.MagicDefense+float64(steps)*1.25, 100)
	}
}

func (w *World) tickLaneMinion(minion *Entity, tick uint64, tickRate int) {
	if minion == nil || minion.Stats.HP <= 0 || tickRate <= 0 {
		return
	}
	routeStart := w.spawnPosition(minion.Team)
	routeEnd := minion.Lane.RouteTarget
	if routeEnd == (Vector2{}) {
		routeEnd = w.spawnPosition(oppositeTeam(minion.Team))
		minion.Lane.RouteTarget = routeEnd
	}
	if distancePointToSegment(minion.Position, routeStart, routeEnd) <= laneMinionReturnDistance {
		minion.Lane.LastOnLaneTick = tick
	}
	if minion.Intent.AttackTargetID != "" && tick-minion.Lane.LastOnLaneTick >= secondsToTicks(laneMinionReturnSeconds, tickRate) {
		minion.Intent.AttackTargetID = ""
		minion.Combat.PendingAttackTargetID = ""
		minion.Combat.AttackReleaseTick = 0
		destination := laneMoveTarget(minion.Position, routeStart, routeEnd)
		w.moveToward(minion, w.laneMoveTargetAvoidingAllies(minion, destination), movementStepAtTick(minion, tickRate, tick), 8)
		return
	}

	target := w.entities[minion.Intent.AttackTargetID]
	if !canAttackTarget(minion, target) || distance(minion.Position, target.Position) > laneMinionAggroRange+target.Radius {
		target = w.nearestLaneTarget(minion)
		if target != nil {
			minion.Intent.AttackTargetID = target.ID
		} else {
			minion.Intent.AttackTargetID = ""
		}
	}
	if canAttackTarget(minion, target) {
		if distance(minion.Position, target.Position) <= w.attackReachAtTick(minion, target, tick) {
			w.applyAttack(minion, target, tick, tickRate)
			return
		}
		w.moveToward(minion, w.laneMoveTargetAvoidingAllies(minion, target.Position), movementStepAtTick(minion, tickRate, tick), 0)
		return
	}
	destination := laneMoveTarget(minion.Position, routeStart, routeEnd)
	w.moveToward(minion, w.laneMoveTargetAvoidingAllies(minion, destination), movementStepAtTick(minion, tickRate, tick), 8)
}

func (w *World) nearestLaneTarget(minion *Entity) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	for _, target := range w.entities {
		if !canAttackTarget(minion, target) || !isLaneTarget(target) {
			continue
		}
		d := distance(minion.Position, target.Position)
		if d <= laneMinionAggroRange+target.Radius && d < bestDistance {
			best = target
			bestDistance = d
		}
	}
	return best
}

func isLaneTarget(entity *Entity) bool {
	return entity != nil && (entity.Kind == EntityKindPlayer || entity.Kind == EntityKindEnemyHero || isMinion(entity) || isStructure(entity))
}

func isStructure(entity *Entity) bool {
	return entity != nil && (entity.Kind == EntityKindTower || entity.Kind == EntityKindBarracks || entity.Kind == EntityKindCrystal)
}

func (w *World) tickTower(tower *Entity, tick uint64, tickRate int) {
	if tower == nil || tower.Kind != EntityKindTower || tower.Stats.HP <= 0 || tickRate <= 0 {
		return
	}
	target := w.entities[tower.Intent.AttackTargetID]
	if !isTowerTarget(tower, target) || distance(tower.Position, target.Position) > w.attackReachAtTick(tower, target, tick) {
		target = w.nearestTowerTarget(tower, tick)
		if target == nil {
			tower.Intent.AttackTargetID = ""
			return
		}
		tower.Intent.AttackTargetID = target.ID
	}
	w.applyAttack(tower, target, tick, tickRate)
}

func (w *World) nearestTowerTarget(tower *Entity, tick uint64) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	for _, target := range w.entities {
		if !isTowerTarget(tower, target) {
			continue
		}
		d := distance(tower.Position, target.Position)
		if d <= w.attackReachAtTick(tower, target, tick) && d < bestDistance {
			best = target
			bestDistance = d
		}
	}
	return best
}

func isTowerTarget(tower *Entity, target *Entity) bool {
	return canAttackTarget(tower, target) && (target.Kind == EntityKindPlayer || target.Kind == EntityKindEnemyHero || isMinion(target))
}

func oppositeTeam(team Team) Team {
	if team == TeamRed {
		return TeamBlue
	}
	return TeamRed
}

func laneMoveTarget(position Vector2, routeStart Vector2, routeEnd Vector2) Vector2 {
	if distancePointToSegment(position, routeStart, routeEnd) > laneMinionReturnDistance {
		return closestPointOnSegment(position, routeStart, routeEnd)
	}
	return routeEnd
}

func (w *World) laneMoveTargetAvoidingAllies(minion *Entity, target Vector2) Vector2 {
	dx, dy := normalize(target.X-minion.Position.X, target.Y-minion.Position.Y)
	if dx == 0 && dy == 0 {
		return target
	}
	perpX, perpY := -dy, dx
	bestForward := math.MaxFloat64
	bestClearance := 0.0
	for _, other := range w.entities {
		if other == nil || other.ID == minion.ID || other.Team != minion.Team || !isCollisionEntity(other) {
			continue
		}
		rx := other.Position.X - minion.Position.X
		ry := other.Position.Y - minion.Position.Y
		forward := rx*dx + ry*dy
		clearance := minion.Radius + other.Radius + 8
		if forward <= 0 || forward > laneMinionAvoidLookahead || forward >= bestForward {
			continue
		}
		side := rx*perpX + ry*perpY
		if math.Abs(side) >= clearance {
			continue
		}
		bestForward = forward
		bestClearance = clearance
	}
	if bestForward == math.MaxFloat64 {
		return target
	}
	sideStep := bestClearance * laneMinionAvoidSide(minion)
	return Vector2{
		X: clamp(minion.Position.X+dx*laneMinionAvoidLookahead+perpX*sideStep, 0, w.width),
		Y: clamp(minion.Position.Y+dy*laneMinionAvoidLookahead+perpY*sideStep, 0, w.height),
	}
}

func laneMinionAvoidSide(minion *Entity) float64 {
	if len(minion.ID) > 0 && minion.ID[len(minion.ID)-1]%2 == 0 {
		return 1
	}
	return -1
}
