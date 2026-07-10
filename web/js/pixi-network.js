function connect() {
  if (state.socket?.readyState === WebSocket.OPEN) {
    return;
  }

  state.roomId = els.roomId.value.trim() || "demo-room";
  state.playerId = Array.from(els.playerId.value.trim() || state.playerId).slice(0, 16).join("");
  state.team = els.team.value;
  els.serverUrl.value = els.serverUrl.value.trim() || websocketURL();
  els.playerId.value = state.playerId;

  const socket = new WebSocket(els.serverUrl.value.trim());
  state.socket = socket;
  setStatus("连接中");

  socket.addEventListener("open", () => {
    setStatus("已连接");
    sendPacket("join_room", {
      roomId: state.roomId,
      playerId: state.playerId,
      heroId: els.heroId.value,
      team: state.team,
    });
  });

  socket.addEventListener("message", (event) => {
    const packet = JSON.parse(event.data);
    if (packet.type === "snapshot") {
      queueSnapshot(packet.payload);
    }
    if (packet.type === "error") {
      setStatus(packet.payload?.message || "服务器错误");
    }
  });

  socket.addEventListener("close", () => {
    setStatus("未连接");
    resetClientState();
  });
  socket.addEventListener("error", () => setStatus("连接错误"));
}

function queueSnapshot(snapshot) {
  state.pendingSnapshot = snapshot;
  if (document.hidden || state.snapshotFrameScheduled) {
    return;
  }
  state.snapshotFrameScheduled = true;
  requestAnimationFrame(flushPendingSnapshot);
}

function flushPendingSnapshot() {
  state.snapshotFrameScheduled = false;
  const snapshot = state.pendingSnapshot;
  state.pendingSnapshot = null;
  if (snapshot) {
    applySnapshot(snapshot);
  }
}

document.addEventListener("visibilitychange", () => {
  if (!document.hidden && state.pendingSnapshot) {
    flushPendingSnapshot();
  }
});

function applySnapshot(snapshot) {
  const previousTargets = targetMap();
  const previousPlayers = state.players;
  state.map = snapshot.map;
  state.players = new Map(
    snapshot.players.map((player) => [
      player.playerId,
      normalizePlayer(player),
    ]),
  );
  const units = visibleUnits(snapshot);
  state.units = new Map(units.map((unit) => [unit.id, normalizeUnit(unit)]));
  updateEffectFlashes(snapshot.effects || []);
  cleanupHiddenEffects(snapshot.effects || []);
  state.effects = snapshot.effects || [];
  const currentTargets = targetMap();
  if (
    state.selectedTargetId &&
    (!currentTargets.has(state.selectedTargetId) ||
      currentTargets.get(state.selectedTargetId)?.dead)
  ) {
    state.selectedTargetId = "";
  }
  if (
    state.attackTargetId &&
    (!currentTargets.has(state.attackTargetId) ||
      currentTargets.get(state.attackTargetId)?.dead)
  ) {
    state.attackTargetId = "";
  }
  updateRewardTexts(previousPlayers, state.players);
  updateDamageEffects(previousTargets, currentTargets, snapshot.tick);
  state.snapshotTick = snapshot.tick;
  state.snapshotAtMs = performance.now();
  els.tick.textContent = snapshot.tick;
  els.playerCount.textContent = snapshot.players.length;
  const self = state.players.get(state.playerId);
  if (self?.message && self.messageTick === snapshot.tick) {
    setShopStatus(self.message);
  }
  updatePositionLabel();
}

function cleanupHiddenEffects(effects) {
  const visibleEffectIds = new Set(effects.map((effect) => effect.id).filter(Boolean));
  for (const id of state.hiddenEffectIds) {
    if (!visibleEffectIds.has(id)) {
      state.hiddenEffectIds.delete(id);
    }
  }
}

function updateRewardTexts(previousPlayers, currentPlayers) {
  for (const [playerId, current] of currentPlayers) {
    const previous = previousPlayers.get(playerId);
    if (!previous) {
      continue;
    }
    const expGain = (current.totalExp || 0) - (previous.totalExp || 0);
    const goldGain = (current.gold || 0) - (previous.gold || 0);
    if (expGain > 0) {
      spawnRewardText(current, `+${formatRewardAmount(expGain)} exp`, "exp");
    }
    if (goldGain > 0) {
      spawnRewardText(current, `+${formatRewardAmount(goldGain)} G`, "gold");
    }
  }
}

function formatRewardAmount(value) {
  return Number.isInteger(value)
    ? String(value)
    : String(Math.round(value * 100) / 100);
}

function updateEffectFlashes(effects) {
  const visibleEffectIds = new Set();
  for (const effect of effects) {
    if (!effect.id) {
      continue;
    }
    visibleEffectIds.add(effect.id);
    if (state.seenEffectIds.has(effect.id)) {
      continue;
    }
    state.seenEffectIds.add(effect.id);
    const self = state.players.get(state.playerId);
    if (effect.kind === "basic_arrow" && effect.sourceId === self?.id) {
      state.attackFlash = {
        x: self.x,
        y: self.y,
        radius: effect.width || self?.stats?.attackRange || 600,
        until: performance.now() + 220,
      };
    }
  }
  for (const id of state.seenEffectIds) {
    if (!visibleEffectIds.has(id)) {
      state.seenEffectIds.delete(id);
    }
  }
}

function updateDamageEffects(previousTargets, currentTargets, tick) {
  const seenDamageIds = new Set();
  for (const [id, current] of currentTargets) {
    const previous = previousTargets.get(id);
    if (!previous || current.lastHitTick <= previous.lastHitTick) {
      continue;
    }
    seenDamageIds.add(id);
    showTargetDamage(id, current);
    rememberTargetDamage(id, current);
  }
  for (const [id, previous] of previousTargets) {
    if (currentTargets.has(id) || seenDamageIds.has(id)) {
      continue;
    }
    const remembered = state.lastDamageByTarget.get(id);
    if (!remembered || remembered.lastHitTick !== tick - 1) {
      continue;
    }
    showTargetDamage(id, remembered);
    state.lastDamageByTarget.delete(id);
  }
}

function showTargetDamage(id, target) {
  const self = state.players.get(state.playerId);
  const damageEvents = target.damageEvents?.length
    ? target.damageEvents
    : [{ damage: target.lastDamage || 0, damageType: target.lastDamageType || "physical" }];
  if (
    id === state.attackTargetId &&
    self?.heroId !== "archer" &&
    damageEvents.some((event) => event.basicAttack && event.sourceId === self?.id)
  ) {
    state.attackFlash = {
      x: self.x,
      y: self.y,
      radius: self.stats?.attackRange || 0,
      until: performance.now() + 180,
    };
  }
  for (const event of damageEvents.slice(-maxDamageEventsPerTarget)) {
    spawnDamageText(
      target,
      event.damage || 0,
      event.damageType || "physical",
      isLocalPlayerDamage(target, event, self),
    );
  }
}

function isLocalPlayerDamage(target, event, self) {
  if (!self) {
    return false;
  }
  return target.id === self.id || event.sourceId === self.id;
}

function rememberTargetDamage(id, target) {
  state.lastDamageByTarget.set(id, {
    x: target.x,
    y: target.y,
    lastHitTick: target.lastHitTick,
    lastDamage: target.lastDamage,
    lastDamageType: target.lastDamageType,
    damageEvents: target.damageEvents || [],
  });
  if (state.lastDamageByTarget.size > 128) {
    state.lastDamageByTarget.delete(state.lastDamageByTarget.keys().next().value);
  }
}

function leave() {
  sendPacket("leave", null);
  state.socket?.close();
  resetClientState();
}

function resetClientState() {
  state.players = new Map();
  state.units = new Map();
  state.effects = [];
  state.seenEffectIds.clear();
  state.hiddenEffectIds.clear();
  state.servantEffectPositions.clear();
  state.lastDamageByTarget.clear();
  state.moveTarget = null;
  state.selectedTargetId = "";
  state.attackTargetId = "";
  state.selectedEquipmentSlot = 0;
  state.attackMoveArmed = false;
  state.attackFlash = null;
  state.skillPreview = null;
  state.castWindups = [];
  state.snapshotTick = 0;
  state.snapshotAtMs = 0;
  state.pendingSnapshot = null;
  state.snapshotFrameScheduled = false;

  for (const effect of state.damageTexts) {
    effectLayer.removeChild(effect.node);
  }
  state.damageTexts = [];

  els.tick.textContent = "0";
  els.playerCount.textContent = "0";
  setShopStatus("-");
  els.position.textContent = "-";
  els.teamLabel.textContent = "-";
  els.buffs.innerHTML = "-";
  els.skills.innerHTML = "-";
  setStatsCard(null);
  setTargetCard(null);
}

function castSkill(slot) {
  const self = state.players.get(state.playerId);
  if (!self) {
    return;
  }
  const skillId = skillIdForSlot(els.heroId.value, slot);
  if (!skillId) {
    return;
  }
  if (!isSkillLearned(self, skillId)) {
    return;
  }
  if (
    isSkillOnCooldown(self, skillId) &&
    !isNinjaShadowRecast(self, skillId) &&
    !isFrostMageERecast(self, skillId)
  ) {
    return;
  }
  if (!canCastDuringSwordEDash(self)) {
    return;
  }
  const selected = currentTarget();
  const useAimPointFirst =
    skillId.startsWith("mage_") ||
    skillId === "fire_mage_q" ||
    skillId === "fire_mage_w" ||
    skillId === "robot_q" ||
    skillId.startsWith("explorer_") ||
    skillId === "berserker_e" ||
    skillId === "berserker_r" ||
    skillId === "ninja_q" ||
    skillId === "ninja_w" ||
    skillId === "ninja_r" ||
    skillId === "blade_e" ||
    skillId === "sword_cut" ||
    skillId === "sword_sweeping_blade" ||
    skillId === "trap" ||
    skillId === "earthquake";
  const fallbackTarget = state.moveTarget || { x: self.x + 1, y: self.y };
  const target = useAimPointFirst
    ? state.aimPoint || selected || fallbackTarget
    : selected || state.aimPoint || fallbackTarget;
  if (slot === "e" && skillId === "taunt") {
    showTankEPreview(self);
  }
  if (skillId === "frostmage_r") {
    showFrostMageRPreview(self);
  }
  addCastWindup(self, skillId, target, selected);
  sendPacket("input", {
    cast: {
      skillId,
      targetId: skillId === "slam" || skillId === "trap" ? "" : selected?.id || "",
      targetX: target.x,
      targetY: target.y,
    },
    clientSeq: state.seq,
  });
}

function isNinjaShadowRecast(player, skillId) {
  const tick = Number(els.tick.textContent || 0);
  return (
    player?.heroId === "ninja" &&
    ((skillId === "ninja_w" &&
      player.ninja?.shadowRecastSkillId === skillId &&
      (player.ninja?.shadowReadyTick || 0) <= tick &&
      (player.ninja?.shadowExpiresAt || 0) > tick) ||
      (skillId === "ninja_r" &&
        (player.ninja?.rShadowRecastUntil || 0) > tick &&
        (player.ninja?.rShadowExpiresAt || 0) > tick))
  );
}

function isFrostMageERecast(player, skillId) {
  if (player?.heroId !== "frostmage" || skillId !== "frostmage_e") {
    return false;
  }
  const tick = Number(els.tick.textContent || 0);
  const recastTicks = (skillClientConfig.frostmage_e?.recastDelaySeconds || 0.5) * state.tickRate;
  return state.effects.some(
    (effect) =>
      effect.kind === "frostmage_e" &&
      effect.sourceId === player.id &&
      (effect.expiresAt || 0) > tick &&
      (effect.createdAt || 0) + recastTicks <= tick,
  );
}

function canCastDuringSwordEDash(player) {
  if (player?.heroId !== "sword") {
    return true;
  }
  const tick = Number(els.tick.textContent || 0);
  const dashUntilTick = player.control?.dashUntilTick || 0;
  if (dashUntilTick <= tick) {
    return true;
  }
  return dashUntilTick - tick <= 0.2 * state.tickRate;
}

function addCastWindup(self, skillId, target, selectedTarget) {
  const config = skillClientConfig[skillId] || {};
  if (skillId === "berserker_r") {
    return;
  }
  const windupSeconds =
    Number(config.castWindupSeconds || 0) ||
    Number(config.castDelaySeconds || 0);
  if (windupSeconds <= 0) {
    return;
  }
  const now = performance.now();
  if (skillId === "mage_r" && hasActiveCastWindup(skillId, now)) {
    return;
  }
  const durationMs = windupSeconds * 1000;
  const dx = (target?.x ?? self.x + 1) - self.x;
  const dy = (target?.y ?? self.y) - self.y;
  const len = Math.hypot(dx, dy) || 1;
  const preview =
    skillId === "sword_cut" ? swordQPreviewData(self, target) : null;
  state.castWindups.push({
    id: `${skillId}:${now}`,
    skillId,
    heroId: self.heroId,
    x: self.x,
    y: self.y,
    targetX: target?.x ?? self.x + 1,
    targetY: target?.y ?? self.y,
    targetId: selectedTarget?.id || "",
    dirX: dx / len,
    dirY: dy / len,
    range: config.range || 0,
    radius: config.landingRadius || config.whirlwindRadius || 0,
    preview,
    startedAt: now,
    expiresAt: now + durationMs,
    durationMs,
  });
}

function hasActiveCastWindup(skillId, now) {
  return state.castWindups.some(
    (windup) => windup.skillId === skillId && now <= windup.expiresAt,
  );
}

function upgradeSkill(slot) {
  const self = state.players.get(state.playerId);
  if (!self || self.dead || (self.skillPoints || 0) <= 0) {
    return;
  }
  const skillId = skillIdForSlot(self.heroId || els.heroId.value, slot);
  if (!skillId) {
    return;
  }
  const skill = skillState(self, skillId);
  if (!canUpgradeSkill(self, slot, skill?.level || 0)) {
    return;
  }
  sendPacket("input", {
    upgradeSkill: {
      slot,
    },
    clientSeq: state.seq,
  });
}

function debugLevelUp() {
  const self = state.players.get(state.playerId);
  if (
    !self ||
    self.dead ||
    self.level >= (self.maxLevel || levelClientConfig.maxLevel || 18)
  ) {
    return;
  }
  sendPacket("input", {
    debugLevelUp: true,
    clientSeq: state.seq,
  });
}

function toggleDebugAbilityHaste() {
  const self = state.players.get(state.playerId);
  if (!self || self.dead) {
    return;
  }
  const enabled = (self.buffs || []).some(
    (buff) => buff.id === "debug_ability_haste",
  );
  sendPacket("input", {
    debugAbilityHaste: enabled ? 0 : 200,
    clientSeq: state.seq,
  });
}

function debugAddGold() {
  const self = state.players.get(state.playerId);
  if (!self || self.dead) {
    return;
  }
  sendPacket("input", {
    debugGold: 10000,
    clientSeq: state.seq,
  });
}

function buyEquipment() {
  const equipmentId = els.shopItem.value;
  if (!equipmentId) {
    return;
  }
  setShopStatus("-");
  sendPacket("input", {
    buyEquipment: {
      equipmentId,
    },
    clientSeq: state.seq,
  });
}

function sellSelectedEquipment() {
  if (!state.selectedEquipmentSlot) {
    return;
  }
  setShopStatus("-");
  sendPacket("input", {
    sellEquipment: {
      slot: state.selectedEquipmentSlot,
    },
    clientSeq: state.seq,
  });
}

function sendPacket(type, payload) {
  const socket = state.socket;
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    return;
  }
  socket.send(
    JSON.stringify({
      type,
      roomId: state.roomId,
      playerId: state.playerId,
      seq: state.seq++,
      payload,
    }),
  );
}

function spawnObject() {
  const self = state.players.get(state.playerId);
  const x = clamp((self?.x ?? state.map.width / 2) + 120, 0, state.map.width);
  const y = clamp(self?.y ?? state.map.height / 2, 0, state.map.height);
  const kind = els.spawnKind.value;
  let team = els.spawnTeam.value;
  if (kind === "enemy_hero" && self?.team && team === self.team) {
    team = self.team === "blue" ? "red" : "blue";
    els.spawnTeam.value = team;
  }
  sendPacket("spawn_object", {
    kind,
    team,
    x,
    y,
  });
}
