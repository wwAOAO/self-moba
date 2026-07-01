const state = {
  socket: null,
  roomId: "demo-room",
  playerId: `p-${Math.floor(Math.random() * 10000)}`,
  team: "blue",
  showDummies: false,
  seq: 0,
  map: { width: 6000, height: 6000 },
  players: new Map(),
  units: new Map(),
  effects: [],
  seenEffectIds: new Set(),
  sprites: new Map(),
  unitSprites: new Map(),
  damageTexts: [],
  moveTarget: null,
  selectedTargetId: "",
  attackTargetId: "",
  attackMoveArmed: false,
  attackFlash: null,
  aimPoint: null,
  skillPreview: null,
  castWindups: [],
  snapshotTick: 0,
  snapshotAtMs: 0,
  tickRate: 20,
  frame: { scale: 1, offsetX: 0, offsetY: 0 },
};

const els = {
  serverUrl: document.querySelector("#serverUrl"),
  roomId: document.querySelector("#roomId"),
  playerId: document.querySelector("#playerId"),
  heroId: document.querySelector("#heroId"),
  team: document.querySelector("#team"),
  showDummies: document.querySelector("#showDummies"),
  spawnKind: document.querySelector("#spawnKind"),
  spawnTeam: document.querySelector("#spawnTeam"),
  spawnBtn: document.querySelector("#spawnBtn"),
  levelUpBtn: document.querySelector("#levelUpBtn"),
  connectBtn: document.querySelector("#connectBtn"),
  leaveBtn: document.querySelector("#leaveBtn"),
  status: document.querySelector("#status"),
  tick: document.querySelector("#tick"),
  playerCount: document.querySelector("#playerCount"),
  teamLabel: document.querySelector("#teamLabel"),
  position: document.querySelector("#position"),
  statLevel: document.querySelector("#statLevel"),
  statExp: document.querySelector("#statExp"),
  statSkillPoints: document.querySelector("#statSkillPoints"),
  statResourceLabel: document.querySelector("#statResourceLabel"),
  statResource: document.querySelector("#statResource"),
  statMpLabel: document.querySelector("#statMpLabel"),
  statHp: document.querySelector("#statHp"),
  statMp: document.querySelector("#statMp"),
  statHpRegen5: document.querySelector("#statHpRegen5"),
  statMpRegen5Label: document.querySelector("#statMpRegen5Label"),
  statMpRegen5: document.querySelector("#statMpRegen5"),
  statAttack: document.querySelector("#statAttack"),
  statAbilityPower: document.querySelector("#statAbilityPower"),
  statAbilityHaste: document.querySelector("#statAbilityHaste"),
  statPhysicalDefenseTip: document.querySelector("#statPhysicalDefenseTip"),
  statPhysicalDefense: document.querySelector("#statPhysicalDefense"),
  statMagicDefenseTip: document.querySelector("#statMagicDefenseTip"),
  statMagicDefense: document.querySelector("#statMagicDefense"),
  statMoveSpeed: document.querySelector("#statMoveSpeed"),
  statAttackRange: document.querySelector("#statAttackRange"),
  statAttackSpeed: document.querySelector("#statAttackSpeed"),
  statCritChance: document.querySelector("#statCritChance"),
  target: document.querySelector("#target"),
  skills: document.querySelector("#skills"),
  stage: document.querySelector("#stage"),
};

let heroClientConfig = {};
let heroSkillSlots = {
  sword: {
    passive: "sword_edge",
    q: "sword_cut",
    w: "sword_wind_wall",
    e: "sword_sweeping_blade",
    r: "sword_storm",
  },
  warrior: {
    passive: "warrior_toughness",
    q: "slash",
    w: "dash",
    e: "judgment",
    r: "justice",
  },
  archer: {
    passive: "archer_focus",
    q: "shot",
    w: "roll",
    e: "trap",
    r: "arrow_rain",
  },
  tank: {
    passive: "tank_armor",
    q: "slam",
    w: "guard",
    e: "taunt",
    r: "earthquake",
  },
};

let skillClientConfig = {
  sword_cut: {
    range: 475,
    whirlwindRange: 900,
    eqRadius: 375,
    previewMs: 450,
  },
};
let levelClientConfig = { maxLevel: 18, levels: [] };
let rewardClientConfig = {};

els.playerId.value = state.playerId;

const app = new PIXI.Application();
const worldLayer = new PIXI.Container();
const gridLayer = new PIXI.Graphics();
const unitLayer = new PIXI.Container();
const playerLayer = new PIXI.Container();
const effectLayer = new PIXI.Container();
