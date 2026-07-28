package world

import (
	"math"
	"strconv"
)

const (
	// minionWaveIntervalSeconds 是两波小兵之间的秒数。
	minionWaveIntervalSeconds = 30
	// minionWaveCount 是保留的标准波次单位数量。
	minionWaveCount = 7
	// minionSpawnGapSeconds 是同波小兵依次出生的秒数。
	minionSpawnGapSeconds = 0.55
	// laneMinionAggroRange 是小兵的基础索敌与仇恨响应范围，单位为世界距离。
	laneMinionAggroRange = 450
	// laneMinionReturnDistance 是允许小兵偏离兵线的最大世界距离。
	laneMinionReturnDistance = 500
	// laneMinionReturnSeconds 是偏离兵线后进入强制归线的秒数。
	laneMinionReturnSeconds = 5
	// laneMinionAggroSeconds 是英雄伤害触发的小兵强制仇恨秒数。
	laneMinionAggroSeconds = 2.5
	// laneMinionMoveSpeed 是兵线小兵的固定世界移动速度。
	laneMinionMoveSpeed = 260
	// laneMinionAvoidLookahead 是普通局部避障的前视世界距离。
	laneMinionAvoidLookahead = 120
	// laneMinionReturnArrival 是完成强制归线的世界距离阈值。
	laneMinionReturnArrival = 32
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

// tickLaneMinion 推进单个小兵的归线、索敌、攻击和路线移动行为。
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
	routeDistance := distancePointToSegment(minion.Position, routeStart, routeEnd)
	if routeDistance <= laneMinionReturnDistance {
		minion.Lane.LastOnLaneTick = tick
	}
	if routeDistance > laneMinionReturnDistance && tick >= minion.Lane.LastOnLaneTick+secondsToTicks(laneMinionReturnSeconds, tickRate) {
		minion.Lane.Returning = true
	}
	if minion.Lane.Returning {
		minion.Intent.AttackTargetID = ""
		minion.Lane.AggroTargetID = ""
		minion.Lane.AggroUntilTick = 0
		minion.Combat.PendingAttackTargetID = ""
		minion.Combat.AttackReleaseTick = 0
		destination := laneMoveTarget(minion.Position, routeStart, routeEnd)
		if routeDistance <= laneMinionReturnArrival {
			minion.Lane.Returning = false
			minion.Lane.LastOnLaneTick = tick
		} else {
			w.moveToward(minion, w.laneMoveTargetAvoidingBlockers(minion, destination), movementStepAtTick(minion, tickRate, tick), 8)
			return
		}
	}

	target := w.selectLaneTarget(minion, tick)
	if target == nil {
		minion.Intent.AttackTargetID = ""
		minion.Combat.PendingAttackTargetID = ""
		minion.Combat.AttackReleaseTick = 0
		destination := laneMoveTarget(minion.Position, routeStart, routeEnd)
		w.moveToward(minion, w.laneMoveTargetAvoidingBlockers(minion, destination), movementStepAtTick(minion, tickRate, tick), 8)
		return
	}
	if minion.Intent.AttackTargetID != target.ID {
		minion.Intent.AttackTargetID = target.ID
		minion.Combat.PendingAttackTargetID = ""
		minion.Combat.AttackReleaseTick = 0
	}
	if distance(minion.Position, target.Position) <= w.attackReachAtTick(minion, target, tick) {
		w.applyAttack(minion, target, tick, tickRate)
		return
	}
	w.moveToward(minion, w.laneMoveTargetAvoidingBlockers(minion, target.Position), movementStepAtTick(minion, tickRate, tick), 0)
}

// selectLaneTarget 保留同级当前目标，并在更高优先级目标出现时稳定转火。
func (w *World) selectLaneTarget(minion *Entity, tick uint64) *Entity {
	current := w.entities[minion.Intent.AttackTargetID]
	if !w.isLaneTargetInDetectionRange(minion, current, tick) {
		current = nil
	}
	best := w.nearestLaneTarget(minion, tick)
	if current == nil || best == nil || w.laneTargetPriority(minion, best, tick) < w.laneTargetPriority(minion, current, tick) {
		return best
	}
	return current
}

// nearestLaneTarget 按玩法优先级、距离和实体 ID 选择确定性的最佳目标。
func (w *World) nearestLaneTarget(minion *Entity, tick uint64) *Entity {
	var best *Entity
	bestPriority := math.MaxInt
	bestDistance := math.MaxFloat64
	for _, target := range w.entitiesInStableOrder() {
		if !w.isLaneTargetInDetectionRange(minion, target, tick) {
			continue
		}
		priority := w.laneTargetPriority(minion, target, tick)
		d := distance(minion.Position, target.Position)
		if priority < bestPriority || priority == bestPriority && d < bestDistance {
			best = target
			bestPriority = priority
			bestDistance = d
		}
	}
	return best
}

// isLaneTargetInDetectionRange 判断目标是否有效且位于小兵的索敌或攻击距离内。
func (w *World) isLaneTargetInDetectionRange(minion *Entity, target *Entity, tick uint64) bool {
	if !canAttackTarget(minion, target) || !isLaneTarget(target) {
		return false
	}
	detectionRange := math.Max(laneMinionAggroRange+target.Radius, w.attackReachAtTick(minion, target, tick))
	return distance(minion.Position, target.Position) <= detectionRange
}

// laneTargetPriority 返回越小越优先的目标级别，临时英雄仇恨高于常规索敌。
func (w *World) laneTargetPriority(minion *Entity, target *Entity, tick uint64) int {
	if minion.Lane.AggroTargetID == target.ID && tick < minion.Lane.AggroUntilTick {
		return 0
	}
	if attacked := w.entities[target.Intent.AttackTargetID]; attacked != nil && attacked.Team == minion.Team {
		if IsHeroUnit(attacked) {
			return 1
		}
		if isMinion(attacked) {
			return 2
		}
	}
	if isMinion(target) {
		return 3
	}
	if IsHeroUnit(target) {
		return 4
	}
	return 5
}

// provokeLaneMinions 让英雄伤害事件触发目标友方小兵的限时仇恨。
func (w *World) provokeLaneMinions(source *Entity, target *Entity, tick uint64, tickRate int) {
	if source == nil || target == nil || !IsHeroUnit(source) || source.Team == target.Team || tickRate <= 0 {
		return
	}
	if !IsHeroUnit(target) && !isMinion(target) {
		return
	}
	for _, minion := range w.entitiesInStableOrder() {
		if !minion.Lane.Active || minion.Lane.Returning || minion.Team != target.Team || !canAttackTarget(minion, source) {
			continue
		}
		if distance(minion.Position, target.Position) > laneMinionAggroRange+target.Radius {
			continue
		}
		minion.Lane.AggroTargetID = source.ID
		minion.Lane.AggroUntilTick = tick + secondsToTicks(laneMinionAggroSeconds, tickRate)
	}
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

// nearestTowerTarget 按距离和稳定实体顺序选择防御塔目标。
func (w *World) nearestTowerTarget(tower *Entity, tick uint64) *Entity {
	var best *Entity
	bestDistance := math.MaxFloat64
	for _, target := range w.entitiesInStableOrder() {
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

// laneMoveTargetAvoidingBlockers 为路线上的友军和友方建筑生成稳定的局部绕行点。
func (w *World) laneMoveTargetAvoidingBlockers(minion *Entity, target Vector2) Vector2 {
	dx, dy := normalize(target.X-minion.Position.X, target.Y-minion.Position.Y)
	if dx == 0 && dy == 0 {
		return target
	}
	perpX, perpY := -dy, dx
	bestForward := math.MaxFloat64
	bestClearance := 0.0
	bestIsStructure := false
	for _, other := range w.entitiesInStableOrder() {
		if other == nil || other.ID == minion.ID || other.Team != minion.Team || !isCollisionEntity(other) {
			continue
		}
		rx := other.Position.X - minion.Position.X
		ry := other.Position.Y - minion.Position.Y
		forward := rx*dx + ry*dy
		clearance := minion.Radius + other.Radius + 8
		lookahead := float64(laneMinionAvoidLookahead)
		if isStructure(other) {
			lookahead += clearance
		}
		if forward <= 0 || forward > lookahead || forward >= bestForward {
			continue
		}
		side := rx*perpX + ry*perpY
		if math.Abs(side) >= clearance {
			continue
		}
		bestForward = forward
		bestClearance = clearance
		bestIsStructure = isStructure(other)
	}
	if bestForward == math.MaxFloat64 {
		return target
	}
	forwardStep := float64(laneMinionAvoidLookahead)
	if bestIsStructure {
		forwardStep = bestForward + bestClearance
	}
	sideStep := bestClearance * laneMinionAvoidSide(minion)
	return Vector2{
		X: clamp(minion.Position.X+dx*forwardStep+perpX*sideStep, 0, w.width),
		Y: clamp(minion.Position.Y+dy*forwardStep+perpY*sideStep, 0, w.height),
	}
}

func laneMinionAvoidSide(minion *Entity) float64 {
	if len(minion.ID) > 0 && minion.ID[len(minion.ID)-1]%2 == 0 {
		return 1
	}
	return -1
}
