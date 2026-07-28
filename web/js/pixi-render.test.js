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

test('本地玩家插值后仍保持在镜头中心', () => {
    const context = createRenderContext();
    const source = readFileSync('web/js/pixi-render.js', 'utf8');
    runInNewContext(`${source}\ndraw({ deltaMS: 16 });`, context);

    assert.equal(context.renderedX, context.app.renderer.width / 2);
});
