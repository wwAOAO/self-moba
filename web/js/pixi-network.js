function connect() {
  if (state.socket?.readyState === WebSocket.OPEN) {
    return;
  }

  state.roomId = els.roomId.value.trim() || "demo-room";
  state.playerId = els.playerId.value.trim() || state.playerId;
  state.team = els.team.value;
  els.serverUrl.value = els.serverUrl.value.trim() || websocketURL();
  els.playerId.value = state.playerId;

  const socket = new WebSocket(els.serverUrl.value.trim());
  state.socket = socket;
  setStatus("Connecting");

  socket.addEventListener("open", () => {
    setStatus("Connected");
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
      applySnapshot(packet.payload);
    }
    if (packet.type === "error") {
      setStatus(packet.payload?.message || "Server error");
    }
  });

  socket.addEventListener("close", () => {
    setStatus("Disconnected");
    resetClientState();
  });
  socket.addEventListener("error", () => setStatus("Connection error"));
}

function applySnapshot(snapshot) {
  const previousTargets = targetMap();
  state.map = snapshot.map;
  state.players = new Map(
    snapshot.players.map((player) => [
      player.playerId,
      normalizePlayer(player),
    ]),
  );
  state.showDummies = els.showDummies.checked;
  const units = visibleUnits(snapshot);
  state.units = new Map(units.map((unit) => [unit.id, normalizeUnit(unit)]));
  updateEffectFlashes(snapshot.effects || []);
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
  updateDamageEffects(previousTargets, currentTargets);
  state.snapshotTick = snapshot.tick;
  state.snapshotAtMs = performance.now();
  els.tick.textContent = snapshot.tick;
  els.playerCount.textContent = snapshot.players.length;
  updatePositionLabel();
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
    if (effect.kind === "basic_arrow") {
      const self = state.players.get(state.playerId);
      state.attackFlash = {
        x: effect.x,
        y: effect.y,
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

function updateDamageEffects(previousTargets, currentTargets) {
  for (const [id, current] of currentTargets) {
    const previous = previousTargets.get(id);
    if (!previous || current.lastHitTick <= previous.lastHitTick) {
      continue;
    }
    const self = state.players.get(state.playerId);
    if (id === state.attackTargetId && self?.heroId !== "archer") {
      state.attackFlash = {
        x: self.x,
        y: self.y,
        radius: self.stats?.attackRange || 0,
        until: performance.now() + 180,
      };
    }
    spawnDamageText(
      current,
      current.lastDamage || 0,
      current.lastDamageType || "physical",
    );
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
  state.moveTarget = null;
  state.selectedTargetId = "";
  state.attackTargetId = "";
  state.attackMoveArmed = false;
  state.attackFlash = null;
  state.skillPreview = null;
  state.castWindups = [];
  state.snapshotTick = 0;
  state.snapshotAtMs = 0;

  for (const effect of state.damageTexts) {
    effectLayer.removeChild(effect.node);
  }
  state.damageTexts = [];

  els.tick.textContent = "0";
  els.playerCount.textContent = "0";
  els.position.textContent = "-";
  els.teamLabel.textContent = "-";
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
  if (isSkillOnCooldown(self, skillId)) {
    return;
  }
  const selected = currentTarget();
  const useAimPointFirst =
    skillId === "sword_cut" ||
    skillId === "sword_sweeping_blade" ||
    skillId === "earthquake";
  const fallbackTarget = state.moveTarget || { x: self.x + 1, y: self.y };
  const target = useAimPointFirst
    ? state.aimPoint || selected || fallbackTarget
    : selected || state.aimPoint || fallbackTarget;
  if (slot === "e" && skillId === "taunt") {
    showTankEPreview(self);
  }
  addCastWindup(self, skillId, target, selected);
  sendPacket("input", {
    cast: {
      skillId,
      targetId: skillId === "slam" ? "" : selected?.id || "",
      targetX: target.x,
      targetY: target.y,
    },
    clientSeq: state.seq,
  });
}

function addCastWindup(self, skillId, target, selectedTarget) {
  const config = skillClientConfig[skillId] || {};
  const windupSeconds =
    Number(config.castWindupSeconds || 0) ||
    Number(config.castDelaySeconds || 0);
  if (windupSeconds <= 0) {
    return;
  }
  const now = performance.now();
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
  if ((skill?.level || 0) >= maxSkillLevel(slot)) {
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
