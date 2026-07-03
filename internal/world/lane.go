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
	switch kind {
	case EntityKindMeleeMinion:
		stats.MaxHP += steps * 20
		stats.HP = stats.MaxHP
		stats.Attack += float64(steps)
		stats.PhysicalDefense += float64(steps * 2)
		stats.MagicDefense += float64(steps) * 1.25
	case EntityKindRangedMinion:
		stats.MaxHP += steps * 14
		if stats.MaxHP > 425 {
			stats.MaxHP = 425
		}
		stats.HP = stats.MaxHP
		stats.Attack += float64(steps) * 2
	case EntityKindSiegeMinion:
		stats.MaxHP += steps * 27
		stats.HP = stats.MaxHP
		stats.Attack += float64(steps) * 1.5
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
		w.moveToward(minion, laneMoveTarget(minion.Position, routeStart, routeEnd), movementStepAtTick(minion, tickRate, tick), 8)
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
		w.moveToward(minion, target.Position, movementStepAtTick(minion, tickRate, tick), 0)
		return
	}
	w.moveToward(minion, laneMoveTarget(minion.Position, routeStart, routeEnd), movementStepAtTick(minion, tickRate, tick), 8)
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
	return entity != nil && (entity.Kind == EntityKindPlayer || entity.Kind == EntityKindEnemyHero || isMinion(entity))
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
