/** 平滑同步玩家位置、权威碰撞范围和界面状态。 */
// 英雄图标最大绘制外延约为基础半径的 1.66 倍，覆盖层使用更保守的定位系数。
const modelOverlayExtentRatio = 1.75;

/** 在镜头计算前推进玩家平滑坐标，确保镜头与模型使用同一帧的位置。 */
function syncSpritePositions(deltaMS) {
    const smoothing = 1 - Math.exp(-deltaMS / 80);

    for (const [playerId, sprite] of state.sprites) {
        if (!state.players.has(playerId)) {
            playerLayer.removeChild(sprite.node);
            state.sprites.delete(playerId);
        }
    }

    for (const player of state.players.values()) {
        let sprite = state.sprites.get(player.playerId);
        if (!sprite) {
            sprite = createPlayer(player);
            state.sprites.set(player.playerId, sprite);
            playerLayer.addChild(sprite.node);
        }

        sprite.targetX = player.x;
        sprite.targetY = player.y;
        sprite.x += (sprite.targetX - sprite.x) * smoothing;
        sprite.y += (sprite.targetY - sprite.y) * smoothing;
    }
}

/** 使用已推进的玩家坐标同步权威碰撞范围和界面状态。 */
function syncSprites(frame) {
    for (const player of state.players.values()) {
        const sprite = state.sprites.get(player.playerId);
        if (!sprite) {
            continue;
        }

        redrawPlayerBody(sprite, player);
        updatePlayerCollisionCircle(sprite, player);
        const barY = -(playerModelRadius(player) * modelOverlayExtentRatio + 10);
        updateBars(sprite, player, barY);
        updatePlayerLabel(sprite, player);
        updateStatusLabel(sprite, player, barY + 5, 24);
        sprite.node.x = frame.offsetX + sprite.x * frame.scale;
        sprite.node.y = frame.offsetY + sprite.y * frame.scale;
        sprite.node.alpha = (player.control?.invisibleUntilTick || 0) > interpolatedTick() ? 0.45 : 1;
    }
}

/** 绘制与服务端碰撞半径一致的玩家边界提示。 */
function updatePlayerCollisionCircle(sprite, player) {
    sprite.collision.clear();
    const radius = playerModelRadius(player);
    sprite.collision.circle(0, 0, radius);
    sprite.collision.stroke({ color: 0x172026, width: 1, alpha: 0.45 });
}

/** 平滑同步单位位置，并按当前镜头比例刷新模型和固定尺寸覆盖层。 */
function syncUnits(frame, deltaMS) {
    const smoothing = 1 - Math.exp(-deltaMS / 80);

    for (const [unitId, sprite] of state.unitSprites) {
        if (!state.units.has(unitId)) {
            unitLayer.removeChild(sprite.node);
            state.unitSprites.delete(unitId);
        }
    }

    for (const unit of state.units.values()) {
        let sprite = state.unitSprites.get(unit.id);
        if (!sprite) {
            sprite = createUnit(unit);
            state.unitSprites.set(unit.id, sprite);
            unitLayer.addChild(sprite.node);
        }
        sprite.targetX = unit.x;
        sprite.targetY = unit.y;
        redrawUnitBody(sprite, unit);
        const overlayRadius = unitModelDisplayRadius(unit) * modelOverlayExtentRatio;
        updateUnitBars(sprite, unit, -(overlayRadius + 16));
        updateUnitCollisionCircle(sprite, unit, frame);
        updateStatusLabel(sprite, unit, -(overlayRadius + 30));
        sprite.x += (sprite.targetX - sprite.x) * smoothing;
        sprite.y += (sprite.targetY - sprite.y) * smoothing;
        sprite.node.x = frame.offsetX + sprite.x * frame.scale;
        sprite.node.y = frame.offsetY + sprite.y * frame.scale;
    }
}

function updateUnitCollisionCircle(sprite, unit, frame) {
    if (!sprite.collision) {
        return;
    }
    sprite.collision.clear();
    if (unit.kind !== 'enemy_hero') {
        return;
    }
    const radius = (unit.radius || 18) * frame.scale;
    if (radius < 1) {
        return;
    }
    sprite.collision.circle(0, 0, radius);
    sprite.collision.stroke({ color: 0x172026, width: 1, alpha: 0.65 });
}

/** 创建玩家模型及其状态、资源和碰撞提示图层。 */
function createPlayer(player) {
    const node = new PIXI.Container();
    const body = new PIXI.Graphics();
    const collision = new PIXI.Graphics();
    const hpBack = new PIXI.Graphics();
    const hpFill = new PIXI.Graphics();
    const resourceBack = new PIXI.Graphics();
    const resourceFill = new PIXI.Graphics();
    const statusLabel = createStatusLabel();
    const label = new PIXI.Text({
        text: player.playerId,
        style: {
            fill: 0x172026,
            fontFamily: 'Arial',
            fontSize: 13,
            fontWeight: '700',
        },
    });

    label.anchor.set(0.5, 0);
    node.addChild(statusLabel, hpBack, hpFill, resourceBack, resourceFill, collision, body, label);
    return {
        node,
        body,
        collision,
        x: player.x,
        y: player.y,
        targetX: player.x,
        targetY: player.y,
        hpBack,
        hpFill,
        resourceBack,
        resourceFill,
        label,
        statusLabel,
    };
}

/** 更新玩家名字内容，并保持文字位于世界模型边缘之外。 */
function updatePlayerLabel(sprite, player) {
    if (!sprite.label) {
        return;
    }
    sprite.label.text = player.dead ? `${player.playerId} ${Math.ceil(player.respawnIn || 0)}s` : player.playerId;
    sprite.label.y = playerModelRadius(player) * modelOverlayExtentRatio + 6;
    sprite.label.alpha = player.dead ? 0.65 : 1;
}

function redrawPlayerBody(sprite, player) {
    const isSelf = player.playerId === state.playerId;
    const radius = playerModelRadius(player);
    sprite.body.clear();
    const shape = playerModelShape(player);
    if (shape === 'triangle') {
        sprite.body.moveTo(0, -radius);
        sprite.body.lineTo(radius * 0.92, radius * 0.7);
        sprite.body.lineTo(-radius * 0.92, radius * 0.7);
        sprite.body.closePath();
    } else if (shape === 'archer') {
        drawBowArrowIcon(sprite.body, radius);
    } else if (shape === 'crossbow') {
        drawCrossbowIcon(sprite.body, radius);
    } else if (shape === 'square') {
        sprite.body.rect(-radius, -radius, radius * 2, radius * 2);
    } else if (shape === 'octagon') {
        drawChamferedOctagon(sprite.body, radius);
    } else if (shape === 'katana') {
        drawKatanaIcon(sprite.body, radius);
    } else if (shape === 'warrior') {
        drawWarriorIcon(sprite.body, radius);
    } else if (shape === 'sword') {
        drawSwordIcon(sprite.body, radius);
    } else if (shape === 'mage') {
        drawMageIcon(sprite.body, radius);
    } else if (shape === 'fire') {
        drawFireIcon(sprite.body, radius, player.dead, colorForTeam(player.team));
        return;
    } else if (shape === 'snowflake') {
        drawSnowflakeIcon(sprite.body, radius, player.dead, colorForTeam(player.team));
        return;
    } else if (shape === 'gunner') {
        drawGunnerIcon(sprite.body, radius);
    } else if (shape === 'ninja') {
        drawNinjaIcon(sprite.body, radius);
    } else if (shape === 'explorer') {
        drawExplorerHatIcon(sprite.body, radius);
    } else if (shape === 'blade') {
        drawBladeIcon(sprite.body, radius);
    } else if (shape === 'killer') {
        drawKillerIcon(sprite.body, radius);
    } else if (shape === 'shadow_assassin') {
        drawShadowAssassinIcon(sprite.body, radius);
    } else if (shape === 'berserker') {
        drawBerserkerIcon(sprite.body, radius);
    } else if (shape === 'robot') {
        drawRobotIcon(sprite.body, radius);
    } else if (shape === 'doctor') {
        drawDoctorIcon(sprite.body, radius);
    } else if (shape === 'monk') {
        drawMonkIcon(sprite.body, radius);
    } else if (shape === 'butcher') {
        drawButcherIcon(sprite.body, radius);
    } else {
        sprite.body.circle(0, 0, radius);
    }
    sprite.body.fill(player.dead ? 0x6b7280 : colorForTeam(player.team));
    if (shape !== 'archer' && shape !== 'mage' && shape !== 'ninja') {
        sprite.body.stroke({
            color: player.dead
                ? 0x111827
                : shape === 'gunner' ||
                    shape === 'berserker' ||
                    shape === 'blade' ||
                    shape === 'killer' ||
                    shape === 'shadow_assassin' ||
                    shape === 'robot' ||
                    shape === 'explorer' ||
                    shape === 'doctor' ||
                    shape === 'monk' ||
                    shape === 'butcher'
                  ? 0x000000
                  : isSelf
                    ? 0xffffff
                    : 0x172026,
            width:
                shape === 'gunner' ||
                shape === 'berserker' ||
                shape === 'blade' ||
                shape === 'killer' ||
                shape === 'shadow_assassin' ||
                shape === 'robot' ||
                shape === 'explorer' ||
                shape === 'doctor' ||
                shape === 'monk' ||
                shape === 'butcher'
                    ? 1
                    : isSelf
                      ? 2
                      : 1,
            alpha: player.dead ? 0.45 : 1,
        });
    }
}

/** 创建采用世界尺寸模型和固定屏幕尺寸覆盖层的单位节点。 */
function createUnit(unit) {
    const node = new PIXI.Container();
    const body = new PIXI.Graphics();
    const collision = new PIXI.Graphics();
    const hpFill = new PIXI.Graphics();
    let visual = unitVisual(unit.kind);
    if (
        unit.kind === 'fountain' ||
        unit.kind === 'tower' ||
        unit.kind === 'crystal' ||
        unit.kind === 'barracks' ||
        isMinionKind(unit.kind)
    ) {
        visual = { ...visual, color: colorForTeam(unit.team) };
    }
    const statusLabel = createStatusLabel();
    const label = new PIXI.Text({
        text: visual.label,
        style: {
            fill: 0x172026,
            fontFamily: 'Arial',
            fontSize: 13,
            fontWeight: '700',
        },
    });

    label.anchor.set(0.5, 0);
    label.visible = unitLabelVisible(unit.kind);
    node.addChild(statusLabel, hpFill, collision, body, label);
    return {
        node,
        body,
        visual,
        label,
        hpFill,
        collision,
        statusLabel,
        x: unit.x || 0,
        y: unit.y || 0,
    };
}

/** 按当前镜头比例重绘单位模型，不缩放名字、血条和状态图标。 */
function redrawUnitBody(sprite, unit) {
    const modelRadius = unitModelDisplayRadius(unit);
    if (sprite.modelRadius === modelRadius) {
        return;
    }
    sprite.modelRadius = modelRadius;
    sprite.body.clear();
    drawUnitBody(sprite.body, sprite.visual, modelRadius);
    if (unit.kind !== 'fountain') {
        sprite.body.stroke({ color: 0xf2f7f3, width: 2 });
    }
    sprite.label.y = modelRadius * modelOverlayExtentRatio + 6;
}

function unitLabelVisible(kind) {
    return kind !== 'fountain' && kind !== 'tower' && kind !== 'crystal' && !isMinionKind(kind);
}

function isMinionKind(kind) {
    return kind === 'melee_minion' || kind === 'ranged_minion' || kind === 'siege_minion' || kind === 'super_minion';
}

function createStatusLabel() {
    const label = new PIXI.Text({
        text: '',
        style: {
            fill: 0xffd15c,
            fontFamily: 'Arial',
            fontSize: 14,
            fontWeight: '900',
            stroke: { color: 0xffffff, width: 2 },
        },
    });
    label.anchor.set(0.5, 0.5);
    label.visible = false;
    return label;
}

/** 在场上仅显示最高优先级硬控及其倒计时。 */
function updateStatusLabel(sprite, target, y, x = 0) {
    if (!sprite.statusLabel) {
        return;
    }
    if (target.kind !== 'player' && target.kind !== 'enemy_hero') {
        sprite.statusLabel.visible = false;
        return;
    }
    const tick = Number(els.tick.textContent || 0);
    const statuses = displayStatuses(target, tick).filter(status => status.kind === 'control');
    if (!statuses.length) {
        sprite.statusLabel.visible = false;
        return;
    }
    const primary = statuses[0];
    const additional = statuses.length > 1 ? ` +${statuses.length - 1}` : '';
    sprite.statusLabel.text = `${primary.icon} ${formatStatusDuration(primary, tick)}${additional}`;
    sprite.statusLabel.x = x;
    sprite.statusLabel.y = y;
    sprite.statusLabel.visible = true;
}

function normalizeUnit(unit) {
    return {
        ...unit,
        kind: unit.kind || 'dummy',
        team: unit.team || 'neutral',
    };
}

function visibleUnits(snapshot) {
    const units = snapshot.units || snapshot.dummies || [];
    return units.filter(unit => {
        const kind = unit.kind || 'dummy';
        return kind !== 'dummy' && String(unit.id || '').startsWith('spawn:');
    });
}

function visiblePlayers(snapshot) {
    const players = snapshot.players || [];
    const viewer = players.find(player => player.playerId === state.playerId);
    const viewerTeam = viewer?.team || state.team;
    const tick = snapshot.tick || 0;
    return players.filter(player => {
        if (player.playerId === state.playerId || player.team === viewerTeam) {
            return true;
        }
        return (player.control?.invisibleUntilTick || 0) <= tick;
    });
}

/** 补充客户端目标系统需要的标识，不覆盖服务端权威碰撞半径。 */
function normalizePlayer(player) {
    const isSelf = player.playerId === state.playerId;
    return {
        ...player,
        id: `player:${player.playerId}`,
        kind: 'player',
        team: isSelf ? player.team || state.team : player.team || 'unknown',
    };
}

function unitVisual(kind) {
    const visuals = {
        enemy_hero: { label: 'Enemy Hero', color: 0xdc2626, shape: 'circle' },
        siege_minion: { label: 'Cannon', color: 0x6b7280, shape: 'cannon_minion' },
        melee_minion: { label: 'Melee', color: 0xf97316, shape: 'melee_minion' },
        ranged_minion: { label: 'Ranged', color: 0xfacc15, shape: 'ranged_minion' },
        tower: { label: 'Tower', color: 0x475569, shape: 'tower' },
        crystal: { label: 'Crystal', color: 0xa855f7, shape: 'crystal' },
        barracks: { label: 'Barracks', color: 0x7c2d12, shape: 'barracks' },
        fountain: { label: 'Fountain', color: 0x38bdf8, shape: 'fountain' },
        dummy: { label: 'Dummy', color: 0x8a5a32, shape: 'rect' },
        fruit: { label: 'Fruit', color: 0x22c55e, shape: 'diamond' },
        ward: { label: 'Ward', color: 0x06b6d4, shape: 'diamond' },
    };
    return visuals[kind] || { label: kind, color: 0x334155, shape: 'circle' };
}

function drawUnitBody(body, visual, radius) {
    const size = Math.max(1, radius);
    if (visual.shape === 'diamond') {
        body.moveTo(0, -size);
        body.lineTo(size, 0);
        body.lineTo(0, size);
        body.lineTo(-size, 0);
        body.closePath();
        body.fill(visual.color);
        return;
    }
    if (visual.shape === 'crystal') {
        body.rect(-size * 0.9, size * 0.42, size * 1.8, size * 0.38);
        body.fill(0x475569);
        body.rect(-size * 0.62, size * 0.18, size * 1.24, size * 0.42);
        body.fill(0x64748b);
        body.moveTo(0, -size * 1.05);
        body.lineTo(size * 0.42, -size * 0.22);
        body.lineTo(size * 0.22, size * 0.38);
        body.lineTo(-size * 0.22, size * 0.38);
        body.lineTo(-size * 0.42, -size * 0.22);
        body.closePath();
        body.fill(visual.color);
        body.moveTo(0, -size * 0.78);
        body.lineTo(size * 0.18, -size * 0.18);
        body.lineTo(0, size * 0.18);
        body.lineTo(-size * 0.18, -size * 0.18);
        body.closePath();
        body.fill({ color: 0xf5d0fe, alpha: 0.72 });
        return;
    }
    if (visual.shape === 'melee_minion') {
        drawMeleeMinion(body, size, visual.color);
        return;
    }
    if (visual.shape === 'ranged_minion') {
        drawRangedMinion(body, size, visual.color);
        return;
    }
    if (visual.shape === 'cannon_minion') {
        drawCannonMinion(body, size, visual.color);
        return;
    }
    if (visual.shape === 'tower') {
        drawTowerBuilding(body, size, visual.color);
        return;
    }
    if (visual.shape === 'barracks') {
        drawBarracksBuilding(body, size, visual.color);
        return;
    }
    if (visual.shape === 'fountain') {
        body.circle(0, 0, size);
        body.fill({ color: visual.color, alpha: 0.2 });
        drawFountainCore(body, size, visual.color);
        return;
    }
    if (visual.shape === 'rect') {
        body.roundRect(-size, -size, size * 2, size * 2, 4);
        body.fill(visual.color);
        return;
    }
    body.circle(0, 0, size);
    body.fill(visual.color);
}

function drawTowerBuilding(body, size, color) {
    body.rect(-size * 0.9, size * 0.58, size * 1.8, size * 0.38);
    body.fill(0x334155);
    body.rect(-size * 0.68, size * 0.28, size * 1.36, size * 0.42);
    body.fill(0x475569);

    body.moveTo(-size * 0.48, size * 0.32);
    body.lineTo(-size * 0.34, -size * 0.78);
    body.lineTo(size * 0.34, -size * 0.78);
    body.lineTo(size * 0.48, size * 0.32);
    body.closePath();
    body.fill(visualTowerStone(color));

    body.rect(-size * 0.56, -size * 0.92, size * 1.12, size * 0.3);
    body.fill(0x64748b);
    for (let i = -1; i <= 1; i++) {
        body.rect(i * size * 0.36 - size * 0.1, -size * 1.14, size * 0.2, size * 0.28);
        body.fill(0x475569);
    }

    body.moveTo(0, -size * 1.48);
    body.lineTo(size * 0.34, -size * 0.92);
    body.lineTo(-size * 0.34, -size * 0.92);
    body.closePath();
    body.fill(color);

    body.circle(0, -size * 0.32, size * 0.16);
    body.fill({ color: 0xf8fafc, alpha: 0.78 });
    body.rect(-size * 0.1, size * 0.02, size * 0.2, size * 0.46);
    body.fill({ color: 0x1f2937, alpha: 0.7 });
}

function drawBarracksBuilding(body, size, color) {
    const stone = visualTowerStone(color);
    body.roundRect(-size, size * 0.5, size * 2, size * 0.38, size * 0.1);
    body.fill(0x334155);
    body.rect(-size * 0.82, -size * 0.32, size * 1.64, size * 0.94);
    body.fill(stone);

    for (const side of [-1, 1]) {
        body.rect(side * size * 0.58 - size * 0.2, -size * 0.62, size * 0.4, size * 1.12);
        body.fill(0x475569);
        body.moveTo(side * size * 0.82, -size * 0.58);
        body.lineTo(side * size * 0.58, -size * 1.02);
        body.lineTo(side * size * 0.34, -size * 0.58);
        body.closePath();
        body.fill(color);
    }

    body.moveTo(-size * 0.5, -size * 0.3);
    body.lineTo(0, -size * 0.92);
    body.lineTo(size * 0.5, -size * 0.3);
    body.closePath();
    body.fill(color);

    body.circle(0, size * 0.2, size * 0.3);
    body.fill(0x1f2937);
    body.rect(-size * 0.3, size * 0.2, size * 0.6, size * 0.42);
    body.fill(0x1f2937);
    body.moveTo(0, -size * 0.58);
    body.lineTo(size * 0.16, -size * 0.38);
    body.lineTo(0, -size * 0.18);
    body.lineTo(-size * 0.16, -size * 0.38);
    body.closePath();
    body.fill({ color: 0xf8fafc, alpha: 0.82 });
}

function visualTowerStone(color) {
    if (color === 0xef4444) {
        return 0x7f1d1d;
    }
    if (color === 0x2563eb) {
        return 0x1e3a8a;
    }
    return 0x475569;
}

function drawMeleeMinion(body, size, color) {
    const dark = 0x1f2937;
    body.circle(-size * 0.12, -size * 0.45, size * 0.28);
    body.fill(0x94a3b8);
    body.roundRect(-size * 0.48, -size * 0.22, size * 0.7, size * 0.82, size * 0.12);
    body.fill(color);
    body.moveTo(size * 0.34, -size * 0.38);
    body.lineTo(size * 0.86, -size * 0.12);
    body.lineTo(size * 0.66, size * 0.44);
    body.lineTo(size * 0.2, size * 0.18);
    body.closePath();
    body.fill(0x64748b);
    body.moveTo(-size * 0.52, size * 0.1);
    body.lineTo(-size * 0.92, size * 0.58);
    body.lineTo(-size * 0.7, size * 0.68);
    body.lineTo(-size * 0.32, size * 0.2);
    body.closePath();
    body.fill(dark);
    body.circle(-size * 0.12, -size * 0.5, size * 0.12);
    body.fill(0xe5e7eb);
}

function drawRangedMinion(body, size, color) {
    body.moveTo(0, -size * 0.9);
    body.lineTo(size * 0.5, -size * 0.42);
    body.lineTo(size * 0.62, size * 0.62);
    body.lineTo(-size * 0.62, size * 0.62);
    body.lineTo(-size * 0.5, -size * 0.42);
    body.closePath();
    body.fill(color);
    body.circle(0, -size * 0.48, size * 0.22);
    body.fill(0x111827);
    body.moveTo(-size * 0.52, -size * 0.18);
    body.lineTo(-size * 0.94, size * 0.18);
    body.lineTo(-size * 0.58, size * 0.38);
    body.lineTo(-size * 0.24, size * 0.02);
    body.closePath();
    body.fill(0x8b5cf6);
    body.moveTo(size * 0.35, size * 0.04);
    body.lineTo(size * 0.9, -size * 0.18);
    body.lineTo(size * 0.72, size * 0.3);
    body.lineTo(size * 0.42, size * 0.38);
    body.closePath();
    body.fill({ color: 0xf59e0b, alpha: 0.9 });
}

function drawCannonMinion(body, size, color) {
    const dark = 0x1f2937;
    body.rect(-size * 0.92, -size * 0.46, size * 1.08, size * 0.28);
    body.fill(0x475569);
    body.circle(-size * 0.92, -size * 0.32, size * 0.2);
    body.fill(0x94a3b8);
    body.circle(-size * 0.92, -size * 0.32, size * 0.11);
    body.fill(dark);
    body.moveTo(-size * 0.24, -size * 0.38);
    body.quadraticCurveTo(size * 0.16, -size * 0.92, size * 0.62, -size * 0.36);
    body.lineTo(size * 0.78, size * 0.46);
    body.lineTo(-size * 0.42, size * 0.46);
    body.lineTo(-size * 0.42, -size * 0.08);
    body.lineTo(-size * 0.24, -size * 0.08);
    body.closePath();
    body.fill(color);
    body.rect(-size * 0.64, -size * 0.02, size * 0.28, size * 0.76);
    body.fill(0x475569);
    body.circle(-size * 0.28, size * 0.62, size * 0.27);
    body.fill(dark);
    body.circle(size * 0.48, size * 0.62, size * 0.27);
    body.fill(dark);
    body.circle(-size * 0.28, size * 0.62, size * 0.13);
    body.fill(0xe5e7eb);
    body.circle(size * 0.48, size * 0.62, size * 0.13);
    body.fill(0xe5e7eb);
}

function drawFountainCore(body, size, color) {
    const s = size / 90;
    body.circle(0, 0, 62 * s);
    body.fill({ color: 0x1f2937, alpha: 0.94 });
    body.circle(0, 0, 52 * s);
    body.fill({ color: 0x8b6a3a, alpha: 0.96 });
    body.circle(0, 0, 40 * s);
    body.fill({ color: 0x374151, alpha: 0.96 });
    body.circle(0, 0, 30 * s);
    body.fill({ color, alpha: 0.34 });

    for (let i = 0; i < 8; i++) {
        const angle = (Math.PI * 2 * i) / 8;
        const cx = Math.cos(angle) * 57 * s;
        const cy = Math.sin(angle) * 57 * s;
        body.circle(cx, cy, 8 * s);
        body.fill({ color: 0xa17842, alpha: 0.98 });
    }

    body.circle(0, 0, 24 * s);
    body.fill({ color: 0x111827, alpha: 0.98 });
    body.circle(0, 0, 18 * s);
    body.fill({ color: 0xa17842, alpha: 0.98 });
    body.circle(0, 0, 12 * s);
    body.fill({ color, alpha: 0.9 });

    const guards = [
        [-45, -14, -28, -35, -15, -27, -31, -7],
        [15, -27, 28, -35, 45, -14, 31, -7],
        [-45, 14, -31, 7, -15, 27, -28, 35],
        [15, 27, 31, 7, 45, 14, 28, 35],
    ];
    for (const g of guards) {
        body.moveTo(g[0] * s, g[1] * s);
        body.lineTo(g[2] * s, g[3] * s);
        body.lineTo(g[4] * s, g[5] * s);
        body.lineTo(g[6] * s, g[7] * s);
        body.closePath();
        body.fill({ color: 0x475569, alpha: 0.95 });
    }

    body.circle(0, -18 * s, 34 * s);
    body.stroke({ color, width: Math.max(1, 2 * s), alpha: 0.45 });
    body.rect(-8 * s, -78 * s, 16 * s, 48 * s);
    body.fill({ color, alpha: 0.13 });
}
