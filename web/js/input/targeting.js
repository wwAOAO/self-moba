function currentTarget() {
    if (state.selectedTargetId) {
        return targetMap().get(state.selectedTargetId) || null;
    }
    if (state.attackTargetId) {
        return targetMap().get(state.attackTargetId) || null;
    }
    return null;
}

function pickTargetUnit(event) {
    const point = screenToWorld(event);
    const screenPoint = screenPointFromEvent(event);
    const self = state.players.get(state.playerId);
    const targets = [...targetMap().values()]
        .filter(target => target.id !== self?.id)
        .filter(target => !target.dead)
        .sort((a, b) => screenDistanceToTarget(screenPoint, a) - screenDistanceToTarget(screenPoint, b));
    for (const target of targets) {
        if (hitTestUnit(screenPoint, point, target)) {
            return target.id;
        }
    }
    return '';
}

function pickAttackTargetUnit(event) {
    const targetId = pickTargetUnit(event);
    if (!targetId) {
        return '';
    }
    const self = state.players.get(state.playerId);
    const target = targetMap().get(targetId);
    if (!self || !target || target.team !== self.team || target.team === 'neutral') {
        return targetId;
    }
    return '';
}

function hoveredEnemyHero() {
    const self = state.players.get(state.playerId);
    const screenPoint = state.cursorScreenPoint;
    if (!self || !screenPoint) {
        return null;
    }
    const worldPoint = {
        x: (screenPoint.x - state.frame.offsetX) / state.frame.scale,
        y: (screenPoint.y - state.frame.offsetY) / state.frame.scale,
    };
    const heroes = [...targetMap().values()]
        .filter(target => target.id !== self.id && !target.dead)
        .filter(target => target.kind === 'player' || target.kind === 'enemy_hero')
        .filter(target => target.team !== self.team)
        .sort((a, b) => screenDistanceToTarget(screenPoint, a) - screenDistanceToTarget(screenPoint, b));
    return heroes.find(target => hitTestUnit(screenPoint, worldPoint, target)) || null;
}

function nearestAttackTarget(point, range) {
    const self = state.players.get(state.playerId);
    if (!self) {
        return '';
    }
    let best = null;
    let bestDistance = Infinity;
    for (const target of targetMap().values()) {
        if (target.id === self.id || target.dead) {
            continue;
        }
        if (target.team === self.team && target.team !== 'neutral') {
            continue;
        }
        const distance = Math.hypot(target.x - point.x, target.y - point.y);
        if (distance <= range && distance < bestDistance) {
            best = target;
            bestDistance = distance;
        }
    }
    return best?.id || '';
}

/** 按模型形状检测单位，并为建筑保留最小屏幕点击区域。 */
function hitTestUnit(screenPoint, worldPoint, unit) {
    const visual = unitVisual(unit.kind || 'dummy');
    const isPlayerModel = unit.kind === 'player';
    const isEnemyHeroModel = unit.kind === 'enemy_hero';
    const sx = state.frame.offsetX + unit.x * state.frame.scale;
    const sy = state.frame.offsetY + unit.y * state.frame.scale;
    if (isPlayerModel) {
        return hitTestPlayerModel(screenPoint.x - sx, screenPoint.y - sy, unit);
    }
    if (isEnemyHeroModel) {
        return Math.hypot(screenPoint.x - sx, screenPoint.y - sy) <= unitHitRadius(unit);
    }
    const size = Math.max(18, unit.radius);
    const padding = 10;
    const dx = Math.abs(screenPoint.x - sx);
    const dy = Math.abs(screenPoint.y - sy);
    const minimumRadius = unit.kind === 'tower' || unit.kind === 'barracks' || unit.kind === 'crystal' ? 22 : 18;
    if (visual.shape === 'circle') {
        return Math.hypot(dx, dy) <= Math.max(minimumRadius, (size + padding) * state.frame.scale);
    }
    if (visual.shape === 'tower') {
        return (
            dx <= Math.max(minimumRadius, (size + padding) * state.frame.scale) &&
            dy <= Math.max(minimumRadius, (size * 1.3 + padding) * state.frame.scale)
        );
    }
    const hitRadius = Math.max(minimumRadius, (size + padding) * state.frame.scale);
    return dx <= hitRadius && dy <= hitRadius;
}

function screenDistanceToTarget(point, unit) {
    const sx = state.frame.offsetX + unit.x * state.frame.scale;
    const sy = state.frame.offsetY + unit.y * state.frame.scale;
    return Math.hypot(point.x - sx, point.y - sy);
}

function targetMap() {
    const targets = new Map();
    for (const [id, unit] of state.units) {
        if (unit.kind !== 'fountain') {
            targets.set(id, unit);
        }
    }
    for (const player of state.players.values()) {
        targets.set(player.id, player);
    }
    return targets;
}

function targetLabel(target) {
    if (target.kind === 'player') {
        return `Player ${target.playerId || ''}`.trim();
    }
    return unitVisual(target.kind).label;
}

function targetSelectRadius(target, frame) {
    if (target.kind === 'player') {
        return playerModelRadius(target) + 4;
    }
    if (target.kind === 'enemy_hero') {
        return unitModelDisplayRadius(target) + 4;
    }
    return Math.max(14, (target.radius || 18) * frame.scale + 6);
}
