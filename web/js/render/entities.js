/** 平滑同步玩家位置、权威碰撞范围和界面状态。 */
// 2D 英雄模型相对于碰撞半径的完整边长比例。
const heroModelSizeRatio = 6;
// 英雄血条、资源条和名字使用模型半边尺寸作为安全外延。
const heroModelOverlayExtentRatio = heroModelSizeRatio / 2;
// 矢量单位最大绘制外延约为基础半径的 1.66 倍，覆盖层使用更保守的定位系数。
const unitModelOverlayExtentRatio = 1.75;
// 英雄待机动画每帧持续时间，单位为毫秒。
const heroIdleFrameMs = 320;
// 英雄走路动画每帧持续时间，单位为毫秒。
const heroWalkFrameMs = 60;
// 英雄走路动画的基准移动速度，单位为世界单位/秒；剑客与刀客均为 345。
const heroWalkReferenceSpeed = 345;
// 英雄走路动画允许的最短单帧时长，单位为毫秒，避免高速时动作闪烁。
const heroWalkMinFrameMs = 40;
// 英雄走路动画允许的最长单帧时长，单位为毫秒，避免减速时动作停滞。
const heroWalkMaxFrameMs = 95;
// 网络位置变化小于该世界距离时不更新人物朝向。
const playerFacingMovementThreshold = 0.5;
// 四方向切换边界的迟滞角，单位为弧度，避免斜向移动时反复闪烁。
const playerFacingDirectionHysteresis = Math.PI / 18;
// 四方向对应的朝向中心角，角度约定为上 0、右 π/2、下 π、左 -π/2。
const playerFacingDirectionAngles = Object.freeze({
    up: 0,
    right: Math.PI / 2,
    down: Math.PI,
    left: -Math.PI / 2,
});
// 到达目标后的短暂宽限时间，避免网络快照间隔造成走路循环闪回待机。
const playerWalkingGraceMs = 140;
// 服务端动作名称与普通攻击、技能动作精灵表的键保持一一对应。
const heroActionNames = Object.freeze(['attack', 'q', 'w', 'e', 'r']);
// 圣骑士审判旋转的单帧持续时间，单位为毫秒。
const warriorJudgmentFrameMs = 90;

/** 判断快照中的英雄普通攻击或技能动作是否存在对应素材且仍在服务端动作窗口内。 */
function activeHeroAction(player, tick) {
    const animations = heroModelAnimations.get(player?.heroId);
    if (!animations?.up?.[player?.action] || !heroActionNames.includes(player.action)) {
        return '';
    }
    const startedAtTick = Number(player.actionStartedAtTick);
    const endsAtTick = Number(player.actionEndsAtTick);
    if (!Number.isFinite(startedAtTick) || !Number.isFinite(endsAtTick) || endsAtTick <= startedAtTick) {
        return '';
    }
    return tick >= startedAtTick && tick < endsAtTick ? player.action : '';
}

/** 将服务端动作时间窗口映射为指定英雄动作序列的当前帧。 */
function heroActionFrameIndex(player, tick, frameCount) {
    const startedAtTick = Number(player.actionStartedAtTick);
    const endsAtTick = Number(player.actionEndsAtTick);
    const duration = Math.max(1, endsAtTick - startedAtTick);
    const progress = Math.max(0, Math.min(0.999999, (tick - startedAtTick) / duration));
    return Math.min(frameCount - 1, Math.floor(progress * frameCount));
}

/** 在镜头计算前推进玩家平滑坐标，确保镜头与模型使用同一帧的位置。 */
function syncSpritePositions(deltaMS) {
    const smoothing = 1 - Math.exp(-deltaMS / 80);
    const now = performance.now();

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

        const movementX = player.x - sprite.targetX;
        const movementY = player.y - sprite.targetY;
        const serverFacingX = Number(player.facingX);
        const serverFacingY = Number(player.facingY);
        const hasServerActionFacing =
            activeHeroAction(player, interpolatedTick()) && Math.hypot(serverFacingX, serverFacingY) >= 0.01;
        if (hasServerActionFacing) {
            sprite.facingAngle = playerFacingAngle(serverFacingX, serverFacingY, sprite.facingAngle);
            sprite.facingDirection = playerFacingDirection(sprite.facingAngle, sprite.facingDirection);
        } else if (Math.hypot(movementX, movementY) >= playerFacingMovementThreshold) {
            sprite.facingAngle = playerFacingAngle(movementX, movementY, sprite.facingAngle);
            sprite.facingDirection = playerFacingDirection(sprite.facingAngle, sprite.facingDirection);
            sprite.lastMoveAt = now;
        }
        sprite.targetX = player.x;
        sprite.targetY = player.y;
        sprite.x += (sprite.targetX - sprite.x) * smoothing;
        sprite.y += (sprite.targetY - sprite.y) * smoothing;
        sprite.moving =
            Math.hypot(sprite.targetX - sprite.x, sprite.targetY - sprite.y) >= playerFacingMovementThreshold ||
            now - sprite.lastMoveAt <= playerWalkingGraceMs;
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
        updatePlayerModelAnimation(sprite, player);
        updatePlayerCollisionCircle(sprite, player);
        const barY = -(playerModelRadius(player) * heroModelOverlayExtentRatio + 10);
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
        const overlayRadius = unitModelDisplayRadius(unit) * unitModelOverlayExtentRatio;
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

/** 创建玩家场景节点，并将世界尺寸模型与固定屏幕尺寸覆盖层分层挂载。 */
function createPlayer(player) {
    const node = new PIXI.Container();
    const body = new PIXI.Graphics();
    const model = new PIXI.AnimatedSprite([PIXI.Texture.EMPTY]);
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
    model.anchor.set(0.5);
    model.visible = false;
    node.addChild(statusLabel, hpBack, hpFill, resourceBack, resourceFill, collision, body, model, label);
    return {
        node,
        body,
        model,
        collision,
        x: player.x,
        y: player.y,
        targetX: player.x,
        targetY: player.y,
        facingAngle: 0,
        facingDirection: 'up',
        lastMoveAt: 0,
        moving: false,
        animationName: '',
        modelAssetKey: '',
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
    sprite.label.y = playerModelRadius(player) * heroModelOverlayExtentRatio + 6;
    sprite.label.alpha = player.dead ? 0.65 : 1;
}

/** 优先绘制世界尺寸的 2D 英雄模型，资源不可用时回退统一矢量图标。 */
function redrawPlayerBody(sprite, player) {
    const isSelf = player.playerId === state.playerId;
    const radius = playerModelRadius(player);
    const modelAnimation = heroModelAnimations.get(player.heroId);
    const directionAnimation = modelAnimation?.[sprite.facingDirection] || modelAnimation?.up || modelAnimation;
    const modelTexture = directionAnimation?.idle?.[0] || heroModelTextures.get(player.heroId);
    sprite.body.clear();
    if (modelTexture) {
        sprite.body.circle(0, 0, radius * 1.48);
        sprite.body.fill({ color: colorForTeam(player.team), alpha: player.dead ? 0.12 : 0.28 });
        sprite.body.stroke({ color: player.dead ? 0x374151 : colorForTeam(player.team), width: 2, alpha: 0.9 });
        if (sprite.modelAssetKey !== player.heroId) {
            sprite.model.textures = directionAnimation?.idle || [modelTexture];
            sprite.model.gotoAndStop(0);
            sprite.modelAssetKey = player.heroId;
            sprite.animationName = '';
        }
        sprite.model.scale.set((radius * heroModelSizeRatio) / Math.max(1, sprite.model.texture.frame.height));
        sprite.model.position.set(0, 0);
        sprite.model.rotation = 0;
        sprite.model.alpha = player.dead ? 0.48 : 1;
        sprite.model.visible = true;
        return;
    }
    sprite.model.visible = false;
    drawHeroModelIcon(sprite.body, player, radius, isSelf);
}

/** 根据移动向量计算人物朝向角，为后续四方向素材选择保留稳定状态。 */
function playerFacingAngle(deltaX, deltaY, previousAngle = 0) {
    if (Math.hypot(deltaX, deltaY) < playerFacingMovementThreshold) {
        return previousAngle;
    }
    const angle = Math.atan2(deltaY, deltaX) + Math.PI / 2;
    return Math.atan2(Math.sin(angle), Math.cos(angle));
}

/**
 * 将连续朝向角归入上、右、下、左四方向，并在当前方向边界保留 10° 迟滞。
 * @param {number} angle 以向上为零点的朝向角，单位为弧度。
 * @param {'up'|'right'|'down'|'left'} previousDirection 上一帧采用的方向。
 * @returns {'up'|'right'|'down'|'left'} 当前稳定方向。
 */
function playerFacingDirection(angle, previousDirection = 'up') {
    const normalizedAngle = Math.atan2(Math.sin(angle), Math.cos(angle));
    let candidate = 'left';
    if (normalizedAngle >= -Math.PI / 4 && normalizedAngle < Math.PI / 4) {
        candidate = 'up';
    } else if (normalizedAngle >= Math.PI / 4 && normalizedAngle < (Math.PI * 3) / 4) {
        candidate = 'right';
    } else if (normalizedAngle >= (Math.PI * 3) / 4 || normalizedAngle < (-Math.PI * 3) / 4) {
        candidate = 'down';
    }
    if (candidate === previousDirection || !(previousDirection in playerFacingDirectionAngles)) {
        return candidate;
    }
    const previousAngle = playerFacingDirectionAngles[previousDirection];
    const angularDistance = Math.abs(
        Math.atan2(Math.sin(normalizedAngle - previousAngle), Math.cos(normalizedAngle - previousAngle)),
    );
    return angularDistance >= Math.PI / 4 + playerFacingDirectionHysteresis ? candidate : previousDirection;
}

/**
 * 根据实际移动速度计算英雄走路动画的单帧时长。
 * @param {number} moveSpeed 当前移动速度，单位为世界单位/秒。
 * @returns {number} 受上下限约束的单帧时长，单位为毫秒。
 */
function heroWalkFrameDuration(moveSpeed) {
    const normalizedSpeed = Number(moveSpeed);
    const effectiveSpeed =
        Number.isFinite(normalizedSpeed) && normalizedSpeed > 0 ? normalizedSpeed : heroWalkReferenceSpeed;
    return Math.max(
        heroWalkMinFrameMs,
        Math.min(heroWalkMaxFrameMs, (heroWalkFrameMs * heroWalkReferenceSpeed) / effectiveSpeed),
    );
}

/** 在英雄四方向待机、走路、普通攻击、技能动作和死亡状态间切换 Pixi 原生逐帧动画。 */
function updatePlayerModelAnimation(sprite, player) {
    const animations = heroModelAnimations.get(player.heroId);
    if (!sprite.model.visible || !animations) {
        return;
    }
    const directionAnimations = animations[sprite.facingDirection] || animations.up || animations;
    const tick = interpolatedTick();
    const actionName = activeHeroAction(player, tick);
    const stateName = player.dead ? 'dead' : actionName || (sprite.moving ? 'walk' : 'idle');
    const actionTextures = actionName ? directionAnimations[actionName] : null;
    const animationName = `${sprite.facingDirection}:${stateName}:${player.actionStartedAtTick || 0}`;
    const moving = stateName === 'walk';
    const action = Boolean(actionTextures);
    const loopingAction = player.heroId === 'warrior' && actionName === 'e';
    const frameDuration = moving ? heroWalkFrameDuration(player.stats?.moveSpeed) : heroIdleFrameMs;
    sprite.model.loop = loopingAction || !action;
    sprite.model.animationSpeed = loopingAction
        ? 1000 / (warriorJudgmentFrameMs * 60)
        : action
          ? 1
          : 1000 / (frameDuration * 60);
    if (sprite.animationName !== animationName) {
        sprite.animationName = animationName;
        sprite.model.textures = action ? actionTextures : moving ? directionAnimations.walk : directionAnimations.idle;
        if (player.dead) {
            sprite.model.gotoAndStop(0);
        } else if (loopingAction) {
            sprite.model.gotoAndPlay(0);
        } else if (action) {
            sprite.model.gotoAndStop(heroActionFrameIndex(player, tick, actionTextures.length));
        } else {
            sprite.model.gotoAndPlay(0);
        }
    } else if (action && !loopingAction) {
        sprite.model.gotoAndStop(heroActionFrameIndex(player, tick, actionTextures.length));
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
    sprite.label.y = modelRadius * unitModelOverlayExtentRatio + 6;
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
