const state = {
    socket: null,
    roomId: 'demo-room',
    playerId: `p-${Math.floor(Math.random() * 10000)}`,
    team: 'blue',
    seq: 0,
    map: { width: 8000, height: 8000 },
    players: new Map(),
    units: new Map(),
    effects: [],
    seenEffectIds: new Set(),
    hiddenEffectIds: new Set(),
    servantEffectPositions: new Map(),
    lastDamageByTarget: new Map(),
    sprites: new Map(),
    unitSprites: new Map(),
    damageTexts: [],
    moveTarget: null,
    selectedTargetId: '',
    attackTargetId: '',
    selectedEquipmentSlot: 0,
    selectedShopItem: '',
    shopCategory: 'all',
    attackMoveArmed: false,
    attackFlash: null,
    aimPoint: null,
    cursorScreenPoint: null,
    skillPreview: null,
    castWindups: [],
    pendingSnapshot: null,
    snapshotFrameScheduled: false,
    snapshotTick: 0,
    snapshotAtMs: 0,
    tickRate: 20,
    cameraScale: 0.4,
    frame: { scale: 1, offsetX: 0, offsetY: 0 },
};

const maxDamageEventsPerTarget = 12;
const maxActiveFloatingTexts = 120;

const els = {
    serverUrl: document.querySelector('#serverUrl'),
    roomId: document.querySelector('#roomId'),
    playerId: document.querySelector('#playerId'),
    heroId: document.querySelector('#heroId'),
    team: document.querySelector('#team'),
    spawnKind: document.querySelector('#spawnKind'),
    spawnTeam: document.querySelector('#spawnTeam'),
    spawnBtn: document.querySelector('#spawnBtn'),
    shopItem: document.querySelector('#shopItem'),
    shopStatus: document.querySelector('#shopStatus'),
    shopToggleBtn: document.querySelector('#shopToggleBtn'),
    shopOverlay: document.querySelector('#shopOverlay'),
    shopCloseBtn: document.querySelector('#shopCloseBtn'),
    shopSearch: document.querySelector('#shopSearch'),
    shopGrid: document.querySelector('#shopGrid'),
    shopDetail: document.querySelector('#shopDetail'),
    shopGold: document.querySelector('#shopGold'),
    shopEquipmentSlots: document.querySelector('#shopEquipmentSlots'),
    buyEquipmentBtn: document.querySelector('#buyEquipmentBtn'),
    sellEquipmentBtn: document.querySelector('#sellEquipmentBtn'),
    levelUpBtn: document.querySelector('#levelUpBtn'),
    abilityHasteBtn: document.querySelector('#abilityHasteBtn'),
    goldBtn: document.querySelector('#goldBtn'),
    connectBtn: document.querySelector('#connectBtn'),
    leaveBtn: document.querySelector('#leaveBtn'),
    status: document.querySelector('#status'),
    tick: document.querySelector('#tick'),
    playerCount: document.querySelector('#playerCount'),
    teamLabel: document.querySelector('#teamLabel'),
    position: document.querySelector('#position'),
    statLevel: document.querySelector('#statLevel'),
    statExp: document.querySelector('#statExp'),
    statSkillPoints: document.querySelector('#statSkillPoints'),
    equipGold: document.querySelector('#equipGold'),
    equipmentSlots: [
        document.querySelector('#equipment1'),
        document.querySelector('#equipment2'),
        document.querySelector('#equipment3'),
        document.querySelector('#equipment4'),
        document.querySelector('#equipment5'),
        document.querySelector('#equipment6'),
    ],
    equipmentTips: [
        document.querySelector('#equipmentTip1'),
        document.querySelector('#equipmentTip2'),
        document.querySelector('#equipmentTip3'),
        document.querySelector('#equipmentTip4'),
        document.querySelector('#equipmentTip5'),
        document.querySelector('#equipmentTip6'),
    ],
    statResourceLabel: document.querySelector('#statResourceLabel'),
    statResource: document.querySelector('#statResource'),
    statMpLabel: document.querySelector('#statMpLabel'),
    statHp: document.querySelector('#statHp'),
    statMp: document.querySelector('#statMp'),
    statHpRegen5: document.querySelector('#statHpRegen5'),
    statMpRegen5Label: document.querySelector('#statMpRegen5Label'),
    statMpRegen5: document.querySelector('#statMpRegen5'),
    statAttack: document.querySelector('#statAttack'),
    statAbilityPower: document.querySelector('#statAbilityPower'),
    statAbilityHasteTip: document.querySelector('#statAbilityHasteTip'),
    statAbilityHaste: document.querySelector('#statAbilityHaste'),
    statPhysicalDefenseTip: document.querySelector('#statPhysicalDefenseTip'),
    statPhysicalDefense: document.querySelector('#statPhysicalDefense'),
    statMagicDefenseTip: document.querySelector('#statMagicDefenseTip'),
    statMagicDefense: document.querySelector('#statMagicDefense'),
    statMoveSpeed: document.querySelector('#statMoveSpeed'),
    statAttackRange: document.querySelector('#statAttackRange'),
    statAttackSpeed: document.querySelector('#statAttackSpeed'),
    statCritChanceTip: document.querySelector('#statCritChanceTip'),
    statCritChance: document.querySelector('#statCritChance'),
    statOmnivamp: document.querySelector('#statOmnivamp'),
    statLifeSteal: document.querySelector('#statLifeSteal'),
    statHealingPower: document.querySelector('#statHealingPower'),
    target: document.querySelector('#target'),
    buffs: document.querySelector('#buffs'),
    skills: document.querySelector('#skills'),
    heroPortrait: document.querySelector('#heroPortrait'),
    hudHpFill: document.querySelector('#hudHpFill'),
    hudResourceFill: document.querySelector('#hudResourceFill'),
    minimap: document.querySelector('#minimap'),
    hud: document.querySelector('.hud'),
    stage: document.querySelector('#stage'),
};

let heroClientConfig = {};
let heroSkillSlots = {
    sword: {
        passive: 'sword_edge',
        q: 'sword_cut',
        w: 'sword_wind_wall',
        e: 'sword_sweeping_blade',
        r: 'sword_storm',
    },
    warrior: {
        passive: 'warrior_toughness',
        q: 'slash',
        w: 'dash',
        e: 'judgment',
        r: 'justice',
    },
    archer: {
        passive: 'archer_focus',
        q: 'shot',
        w: 'roll',
        e: 'trap',
        r: 'arrow_rain',
    },
    tank: {
        passive: 'tank_armor',
        q: 'slam',
        w: 'guard',
        e: 'taunt',
        r: 'earthquake',
    },
    mage: {
        passive: 'mage_arcane_flow',
        q: 'mage_q',
        w: 'mage_w',
        e: 'mage_e',
        r: 'mage_r',
    },
    gunner: {},
    blade: {
        passive: 'blade_passive',
        q: 'blade_q',
        w: 'blade_w',
        e: 'blade_e',
        r: 'blade_r',
    },
    berserker: {
        passive: 'berserker_passive',
        q: 'berserker_q',
        w: 'berserker_w',
        e: 'berserker_e',
        r: 'berserker_r',
    },
    ninja: {
        passive: 'ninja_passive',
        q: 'ninja_q',
        w: 'ninja_w',
        e: 'ninja_e',
        r: 'ninja_r',
    },
};

let skillClientConfig = {
    sword_cut: {
        range: 475,
        whirlwindRange: 900,
        eqRadius: 375,
        eqWindowSeconds: 0.35,
        previewMs: 450,
    },
};
let levelClientConfig = { maxLevel: 18, levels: [] };
let rewardClientConfig = {};
let equipmentClientConfig = {};

els.playerId.value = state.playerId;

const app = new PIXI.Application();
const worldLayer = new PIXI.Container();
const gridLayer = new PIXI.Graphics();
const unitLayer = new PIXI.Container();
const playerLayer = new PIXI.Container();
const skillLayer = new PIXI.Graphics();
const effectLayer = new PIXI.Container();
// 首批英雄 2D 模型资源路径，后续英雄沿用相同注册方式逐步迁移。
const heroModelAssetPaths = Object.freeze({
    sword: '/assets/heroes/sword/model.png?v=20260729-2',
    blade: '/assets/heroes/blade/model.png?v=20260729-2',
    mage: '/assets/heroes/mage/model.png?v=20260730-1',
    warrior: '/assets/heroes/warrior/model.png?v=20260730-2',
});
// 已制作动画的英雄资源路径；方向表首行是四帧待机，其余十二帧是走路循环。
const heroAnimationAssetPaths = Object.freeze({
    sword: {
        directions: {
            up: '/assets/heroes/sword/animation-up.png?v=20260730-1',
            right: '/assets/heroes/sword/animation-right.png?v=20260730-1',
            down: '/assets/heroes/sword/animation-down.png?v=20260730-1',
            left: '/assets/heroes/sword/animation-left.png?v=20260730-1',
        },
        actions: {
            attack: '/assets/heroes/sword/action-attack.png?v=20260730-2',
            q: '/assets/heroes/sword/action-q.png?v=20260729-1',
            w: '/assets/heroes/sword/action-w.png?v=20260729-1',
            e: '/assets/heroes/sword/action-e.png?v=20260729-1',
            r: '/assets/heroes/sword/action-r.png?v=20260729-1',
        },
    },
    blade: {
        directions: {
            up: '/assets/heroes/blade/animation-up.png?v=20260729-1',
            right: '/assets/heroes/blade/animation-right.png?v=20260729-1',
            down: '/assets/heroes/blade/animation-down.png?v=20260729-1',
            left: '/assets/heroes/blade/animation-left.png?v=20260729-1',
        },
        actions: {
            attack: '/assets/heroes/blade/action-attack.png?v=20260730-2',
            q: '/assets/heroes/blade/action-q.png?v=20260729-1',
            w: '/assets/heroes/blade/action-w.png?v=20260729-1',
            e: '/assets/heroes/blade/action-e.png?v=20260729-1',
            r: '/assets/heroes/blade/action-r.png?v=20260729-1',
        },
    },
    // 光明法师使用四方向待机、走路、普通攻击与 Q/W/E/R 施法动作，光效仍由特效层绘制。
    mage: {
        directions: {
            up: '/assets/heroes/mage/animation-up.png?v=20260730-2',
            right: '/assets/heroes/mage/animation-right.png?v=20260730-2',
            down: '/assets/heroes/mage/animation-down.png?v=20260730-2',
            left: '/assets/heroes/mage/animation-left.png?v=20260730-2',
        },
        actions: {
            attack: '/assets/heroes/mage/action-attack.png?v=20260730-1',
            q: '/assets/heroes/mage/action-q.png?v=20260730-1',
            w: '/assets/heroes/mage/action-w.png?v=20260730-1',
            e: '/assets/heroes/mage/action-e.png?v=20260730-1',
            r: '/assets/heroes/mage/action-r.png?v=20260730-1',
        },
    },
    // 圣骑士人物动作只包含身体与大剑，护盾、旋风和圣剑仍由特效层绘制。
    warrior: {
        directions: {
            up: '/assets/heroes/warrior/animation-up.png?v=20260731-1',
            right: '/assets/heroes/warrior/animation-right.png?v=20260731-1',
            down: '/assets/heroes/warrior/animation-down.png?v=20260731-1',
            left: '/assets/heroes/warrior/animation-left.png?v=20260731-1',
        },
        actions: {
            attack: '/assets/heroes/warrior/action-attack.png?v=20260731-1',
            q: '/assets/heroes/warrior/action-q.png?v=20260731-1',
            w: '/assets/heroes/warrior/action-w.png?v=20260731-1',
            e: '/assets/heroes/warrior/action-e.png?v=20260731-1',
            r: '/assets/heroes/warrior/action-r.png?v=20260731-1',
        },
    },
});
// 已成功加载的英雄模型纹理；缺失项继续使用矢量图标。
const heroModelTextures = new Map();
// 已切分的英雄动画帧，按英雄和四方向保存待机、走路、普通攻击与技能序列。
const heroModelAnimations = new Map();
