const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { runInNewContext } = require('node:vm');
const test = require('node:test');

/** 创建足以执行渲染主循环的最小浏览器上下文。 */
function createRenderContext() {
    const sprite = { x: 3990, y: 4000 };
    const context = {
        app: {
            renderer: { width: 1000, height: 800 },
        },
        window: { devicePixelRatio: 1 },
        els: {
            hud: { getBoundingClientRect: () => ({ top: 700 }) },
            minimap: { clientWidth: 0 },
        },
        state: {
            playerId: 'self',
            sprites: new Map([['self', sprite]]),
            players: new Map(),
            units: new Map(),
            map: { width: 8000, height: 8000 },
            cameraScale: 0.4,
            moveTarget: null,
            selectedTargetId: '',
            attackTargetId: '',
            attackFlash: null,
        },
        gridLayer: {
            clear() {},
            rect() {},
            fill() {},
            stroke() {},
        },
        drawEffects() {},
        syncUnits() {},
        syncDamageTexts() {},
        targetMap: () => new Map(),
        renderedX: 0,
    };
    context.syncSpritePositions = () => {
        sprite.x = 4000;
    };
    context.syncSprites = frame => {
        // 兼容旧执行顺序：若帧开始时未推进坐标，则在渲染阶段推进并暴露中心偏差。
        if (sprite.x === 3990) {
            sprite.x = 4000;
        }
        context.renderedX = frame.offsetX + sprite.x * frame.scale;
    };
    return context;
}

/** 记录指定英雄图标产生的 Pixi 绘图命令，供轮廓回归测试使用。 */
function recordHeroIcon(heroId) {
    const commands = [];
    const graphics = {};
    for (const method of [
        'circle',
        'closePath',
        'ellipse',
        'fill',
        'lineTo',
        'moveTo',
        'quadraticCurveTo',
        'rect',
        'roundRect',
        'stroke',
    ]) {
        graphics[method] = (...args) => {
            commands.push([method, ...args]);
            return graphics;
        };
    }
    const context = {
        colorForTeam: () => 0x2563eb,
        graphics,
        heroId,
        result: null,
    };
    const source = readFileSync('web/js/visuals/icons.js', 'utf8');
    runInNewContext(
        `${source}\nresult = playerModelShape({ heroId }); drawHeroModelIcon(graphics, { heroId, team: 'blue', dead: false }, 20, true);`,
        context,
    );
    return { commands, shape: context.result };
}

/** 记录指定技能特效函数产生的 Pixi 绘图命令。 */
function recordSkillEffect(functionName, effect) {
    const commands = [];
    const skillLayer = {};
    for (const method of ['arc', 'circle', 'closePath', 'fill', 'lineTo', 'moveTo', 'rect', 'stroke']) {
        skillLayer[method] = (...args) => {
            commands.push([method, ...args]);
            return skillLayer;
        };
    }
    const context = {
        clamp: (value, min, max) => Math.max(min, Math.min(max, value)),
        effect,
        els: { tick: { textContent: '10' } },
        frame: { offsetX: 0, offsetY: 0, scale: 1 },
        performance: { now: () => 1000 },
        skillLayer,
        state: { snapshotAtMs: 0, tickRate: 20 },
    };
    const source = `${readFileSync('web/js/render/effects.js', 'utf8')}\n${readFileSync('web/js/render/windups.js', 'utf8')}`;
    const alpha = functionName === 'drawTankGroundSlamWindup' ? ', 1' : '';
    runInNewContext(`${source}\n${functionName}(effect, frame${alpha});`, context);
    return commands;
}

test('本地玩家插值后仍保持在镜头中心', () => {
    const context = createRenderContext();
    const source = readFileSync('web/js/pixi-render.js', 'utf8');
    runInNewContext(`${source}\ndraw({ deltaMS: 16 });`, context);

    assert.equal(context.renderedX, context.app.renderer.width / 2);
});

test('优化英雄使用不同轮廓且医生保留高对比十字', () => {
    const heroIds = ['archer', 'blade', 'shadow_assassin', 'tank', 'doctor'];
    const results = heroIds.map(recordHeroIcon);
    const signatures = results.map(result => JSON.stringify(result.commands));

    assert.equal(new Set(signatures).size, heroIds.length);
    assert.equal(results[3].shape, 'stone_golem');
    assert.ok(results[4].commands.some(command => command[0] === 'fill' && command[1] === 0xfff4d6));
});

test('剑客、刀客、光明法师与圣骑士 2D 模型均为 RGBA PNG 且保留矢量降级', () => {
    const modelAssets = [
        ['sword', 'web/assets/heroes/sword/model.png'],
        ['blade', 'web/assets/heroes/blade/model.png'],
        ['mage', 'web/assets/heroes/mage/model.png'],
        ['warrior', 'web/assets/heroes/warrior/model.png'],
    ];
    const registry = readFileSync('web/js/pixi-state.js', 'utf8');

    for (const [heroId, path] of modelAssets) {
        const image = readFileSync(path);
        assert.deepEqual([...image.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
        assert.equal(image.readUInt32BE(16), 512);
        assert.equal(image.readUInt32BE(20), 512);
        assert.equal(image[25], 6);
        assert.match(registry, new RegExp(`${heroId}: '/assets/heroes/${heroId}/model\\.png`));
        assert.ok(recordHeroIcon(heroId).commands.length > 0);
    }

    for (const heroId of ['sword', 'blade', 'mage', 'warrior']) {
        for (const direction of ['up', 'right', 'down', 'left']) {
            const animation = readFileSync(`web/assets/heroes/${heroId}/animation-${direction}.png`);
            assert.deepEqual([...animation.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
            assert.equal(animation.readUInt32BE(16), 1024);
            assert.equal(animation.readUInt32BE(20), 1024);
            assert.equal(animation[25], 6);
            assert.match(registry, new RegExp(`${direction}: '/assets/heroes/${heroId}/animation-${direction}\\.png`));
        }

        const actions = ['attack', 'q', 'w', 'e', 'r'];
        for (const action of actions) {
            const animation = readFileSync(`web/assets/heroes/${heroId}/action-${action}.png`);
            assert.deepEqual([...animation.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
            assert.equal(
                animation.readUInt32BE(16),
                heroId === 'warrior' || (action === 'attack' && ['sword', 'blade'].includes(heroId)) ? 1536 : 1024,
            );
            assert.equal(animation.readUInt32BE(20), 1024);
            assert.equal(animation[25], 6);
            assert.match(registry, new RegExp(`${action}: '/assets/heroes/${heroId}/action-${action}\\.png`));
        }
    }
});

test('剑客、刀客与光明法师使用 AnimatedSprite 切换四方向待机、走路动画并稳定选择移动朝向', () => {
    const context = {
        PIXI: {
            Rectangle: class Rectangle {
                constructor(x, y, width, height) {
                    Object.assign(this, { x, y, width, height });
                }
            },
            Texture: class Texture {
                constructor(options) {
                    Object.assign(this, options);
                }
            },
        },
        result: null,
    };
    const configSource = readFileSync('web/js/pixi-config.js', 'utf8');
    const entitySource = readFileSync('web/js/render/entities.js', 'utf8');
    runInNewContext(
        `${configSource}\n${entitySource}\nconst frames = splitHeroAnimationFrames({ width: 1024, height: 1024, source: 'sheet' });
        const wideFrames = splitHeroAnimationFrames({ width: 1536, height: 1024, source: 'wide-sheet' });
        result = {
            frames,
            wideFrames,
            angles: [playerFacingAngle(0, -1), playerFacingAngle(1, 0), playerFacingAngle(0, 1), playerFacingAngle(-1, 0)],
            directions: [
                playerFacingDirection(0, 'up'),
                playerFacingDirection(Math.PI / 2, 'up'),
                playerFacingDirection(Math.PI, 'right'),
                playerFacingDirection(-Math.PI / 2, 'down'),
            ],
            hysteresis: [
                playerFacingDirection((Math.PI * 50) / 180, 'up'),
                playerFacingDirection((Math.PI * 56) / 180, 'up'),
                playerFacingDirection((Math.PI * 40) / 180, 'right'),
                playerFacingDirection((Math.PI * 34) / 180, 'right'),
            ],
            walkFrameDurations: [
                heroWalkFrameDuration(345),
                heroWalkFrameDuration(1000),
                heroWalkFrameDuration(100),
            ],
            modelRatios: [heroModelSizeRatio, heroModelOverlayExtentRatio, unitModelOverlayExtentRatio],
        };`,
        context,
    );

    assert.equal(context.result.frames.length, 16);
    assert.deepEqual(
        Array.from(context.result.frames, texture => [
            texture.frame.x,
            texture.frame.y,
            texture.frame.width,
            texture.frame.height,
        ]),
        Array.from({ length: 16 }, (_, index) => [(index % 4) * 256, Math.floor(index / 4) * 256, 256, 256]),
    );
    assert.deepEqual(
        Array.from(context.result.wideFrames, texture => [
            texture.frame.x,
            texture.frame.y,
            texture.frame.width,
            texture.frame.height,
        ]),
        Array.from({ length: 16 }, (_, index) => [(index % 4) * 384, Math.floor(index / 4) * 256, 384, 256]),
    );
    assert.match(entitySource, /new PIXI\.AnimatedSprite/);
    assert.match(entitySource, /sprite\.model\.scale\.set/);
    assert.deepEqual(
        Array.from(context.result.angles, value => Math.round(value * 1000)),
        [0, 1571, 3142, -1571],
    );
    assert.deepEqual(Array.from(context.result.directions), ['up', 'right', 'down', 'left']);
    assert.deepEqual(Array.from(context.result.hysteresis), ['up', 'right', 'right', 'up']);
    assert.deepEqual(Array.from(context.result.walkFrameDurations), [60, 40, 95]);
    assert.deepEqual(Array.from(context.result.modelRatios), [6, 3, 1.75]);
});

test('刀客普通攻击动作按服务端 tick 优先于走路并在窗口结束后恢复', () => {
    const tickState = { value: 13 };
    const actionFrames = ['q0', 'q1', 'q2', 'q3'];
    const walkFrames = ['w0', 'w1'];
    const model = {
        visible: true,
        textures: [],
        loop: true,
        currentFrame: -1,
        gotoAndStop(frame) {
            this.currentFrame = frame;
            this.playing = false;
        },
        gotoAndPlay(frame) {
            this.currentFrame = frame;
            this.playing = true;
        },
    };
    const context = {
        heroModelAnimations: new Map([
            [
                'blade',
                {
                    up: { idle: ['i0'], walk: walkFrames, attack: actionFrames },
                },
            ],
        ]),
        interpolatedTick: () => tickState.value,
        model,
        result: null,
    };
    runInNewContext(
        `${readFileSync('web/js/render/entities.js', 'utf8')}
        var player = {
            heroId: 'blade',
            action: 'attack',
            actionStartedAtTick: 10,
            actionEndsAtTick: 18,
            dead: false,
            stats: { moveSpeed: 345 },
        };
        var sprite = {
            model,
            facingDirection: 'up',
            moving: true,
            animationName: '',
        };
        updatePlayerModelAnimation(sprite, player);
        result = {
            actionTextures: sprite.model.textures,
            actionFrame: sprite.model.currentFrame,
            actionLoop: sprite.model.loop,
        };`,
        context,
    );
    assert.deepEqual(context.result.actionTextures, actionFrames);
    assert.equal(context.result.actionFrame, 1);
    assert.equal(context.result.actionLoop, false);

    tickState.value = 18;
    runInNewContext(
        `sprite.animationName = 'up:attack:10';
        updatePlayerModelAnimation(sprite, player);
        result = { textures: sprite.model.textures, playing: sprite.model.playing, loop: sprite.model.loop };`,
        context,
    );
    assert.deepEqual(context.result.textures, walkFrames);
    assert.equal(context.result.playing, true);
    assert.equal(context.result.loop, true);
});

test('圣骑士 E 在服务端动作窗口内循环播放四帧旋转动作', () => {
    const actionFrames = ['e0', 'e1', 'e2', 'e3'];
    const model = {
        visible: true,
        textures: [],
        loop: false,
        gotoAndStop(frame) {
            this.currentFrame = frame;
            this.playing = false;
        },
        gotoAndPlay(frame) {
            this.currentFrame = frame;
            this.playing = true;
        },
    };
    const context = {
        heroModelAnimations: new Map([['warrior', { up: { idle: ['i0'], walk: ['w0'], e: actionFrames } }]]),
        interpolatedTick: () => 30,
        model,
        result: null,
    };
    runInNewContext(
        `${readFileSync('web/js/render/entities.js', 'utf8')}
        updatePlayerModelAnimation(
            { model, facingDirection: 'up', moving: false, animationName: '' },
            {
                heroId: 'warrior',
                action: 'e',
                actionStartedAtTick: 20,
                actionEndsAtTick: 80,
                dead: false,
                stats: { moveSpeed: 340 },
            },
        );
        result = {
            textures: model.textures,
            loop: model.loop,
            playing: model.playing,
            animationSpeed: model.animationSpeed,
        };`,
        context,
    );

    assert.deepEqual(context.result.textures, actionFrames);
    assert.equal(context.result.loop, true);
    assert.equal(context.result.playing, true);
    assert.ok(context.result.animationSpeed > 0);
});

test('石头人 QWER 使用棱角岩片、地裂和多层冲击波', () => {
    const base = { x: 100, y: 100, dirX: 1, dirY: 0, radius: 70, createdAt: 8, expiresAt: 14 };
    const q = recordSkillEffect('drawTankShardEffect', { ...base, speed: 0 });
    const w = recordSkillEffect('drawTankAftershockEffect', { ...base, range: 500 });
    const e = recordSkillEffect('drawTankGroundSlamWindup', {
        ...base,
        range: 400,
        startedAt: 800,
        durationMs: 400,
    });
    const r = recordSkillEffect('drawTankImpactEffect', { ...base, radius: 300 });

    assert.ok(q.filter(command => command[0] === 'lineTo').length >= 5);
    assert.ok(w.filter(command => command[0] === 'lineTo').length >= 20);
    assert.ok(e.filter(command => command[0] === 'lineTo').length >= 20);
    assert.ok(r.filter(command => command[0] === 'circle').length >= 4);
    assert.ok(r.filter(command => command[0] === 'rect').length >= 10);
});

test('剑客 QWER 使用剑斩、龙卷、风墙、位移风痕与旋风斩', () => {
    const base = { x: 100, y: 100, dirX: 1, dirY: 0, createdAt: 8, expiresAt: 14 };
    const q = recordSkillEffect('drawSwordQEffect', { ...base, endX: 700, endY: 100, width: 55 });
    const qHalfRange = recordSkillEffect('drawSwordQEffect', { ...base, endX: 400, endY: 100, width: 55 });
    const qDoubleWidth = recordSkillEffect('drawSwordQEffect', { ...base, endX: 700, endY: 100, width: 110 });
    const q3 = recordSkillEffect('drawSwordWhirlwindEffect', { ...base, range: 900, radius: 80, speed: 20 });
    const w = recordSkillEffect('drawWindWallEffect', { ...base, createdAt: 0, width: 500 });
    const e = recordSkillEffect('drawSwordEEffect', { ...base, endX: 500, endY: 100, radius: 72 });
    const r = recordSkillEffect('drawSwordREffect', { ...base, endX: 300, endY: 100, radius: 450, count: 3 });

    assert.equal(q.filter(command => command[0] === 'arc').length, 3);
    assert.equal(q.filter(command => command[0] === 'lineTo').length, 0);
    assert.equal(q.filter(command => command[0] === 'fill').length, 0);
    const qMainArc = q.filter(command => command[0] === 'arc')[1];
    const qHalfRangeMainArc = qHalfRange.filter(command => command[0] === 'arc')[1];
    const qDoubleWidthMainArc = qDoubleWidth.filter(command => command[0] === 'arc')[1];
    assert.equal(Math.round(qMainArc[3] * 1000), Math.round(qDoubleWidthMainArc[3] * 500));
    assert.equal(
        Math.round((qMainArc[1] + qMainArc[3] - base.x) * 1000),
        Math.round((qHalfRangeMainArc[1] + qHalfRangeMainArc[3] - base.x) * 2000),
    );
    assert.ok(q3.filter(command => command[0] === 'arc').length >= 5);
    assert.ok(w.filter(command => command[0] === 'arc').length >= 6);
    assert.ok(w.some(command => command[0] === 'stroke' && command[1]?.alpha >= 0.4));
    assert.ok(e.filter(command => command[0] === 'lineTo').length >= 8);
    const rArcs = r.filter(command => command[0] === 'arc');
    assert.equal(rArcs.length, 3);
    assert.equal(r.filter(command => command[0] === 'lineTo').length, 7);
    assert.ok(rArcs.every(command => command[3] <= 180));
});
