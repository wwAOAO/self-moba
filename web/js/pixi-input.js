function websocketURL() {
  const scheme = location.protocol === "https:" ? "wss" : "ws";
  return `${scheme}://${location.host || "localhost:6969"}/ws`;
}

function handlePointerDown(event) {
  updateAimPoint(event);
  if (event.button === 0) {
    const targetId = pickTargetUnit(event);
    if (targetId) {
      state.selectedTargetId = targetId;
      setTargetCard(currentTarget());
      if (state.attackMoveArmed) {
        attackTarget(targetId);
        state.attackMoveArmed = false;
      }
      return;
    }
    state.selectedTargetId = "";
    state.attackMoveArmed = false;
    setTargetCard(null);
    return;
  }
  if (event.button !== 2) {
    return;
  }
  const point = screenToWorld(event);
  const targetId = pickTargetUnit(event);
  if (targetId) {
    state.selectedTargetId = targetId;
    setTargetCard(currentTarget());
    attackTarget(targetId);
    return;
  }
  moveToPoint(point);
}

function updateAimPoint(event) {
  const point = screenToWorld(event);
  state.aimPoint = {
    x: clamp(point.x, 0, state.map.width),
    y: clamp(point.y, 0, state.map.height),
  };
}

function attackTarget(targetId) {
  const self = state.players.get(state.playerId);
  if (self?.dead) {
    return;
  }
  state.attackTargetId = targetId;
  state.moveTarget = null;
  sendPacket("input", {
    attack: {
      targetId,
    },
    clientSeq: state.seq,
  });
}

function moveToPoint(point) {
  const self = state.players.get(state.playerId);
  if (self?.dead) {
    return;
  }
  state.attackMoveArmed = false;
  state.attackTargetId = "";
  state.moveTarget = {
    x: clamp(point.x, 0, state.map.width),
    y: clamp(point.y, 0, state.map.height),
  };
  sendPacket("input", {
    move: {
      targetX: state.moveTarget.x,
      targetY: state.moveTarget.y,
    },
    clientSeq: state.seq,
  });
  clearAttackTarget();
}

function clearAttackTarget() {
  sendPacket("input", {
    attack: {
      clear: true,
    },
    clientSeq: state.seq,
  });
}

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
    .filter((target) => target.id !== self?.id)
    .filter((target) => !target.dead)
    .filter(
      (target) =>
        !self || target.team !== self.team || target.team === "neutral",
    )
    .sort(
      (a, b) =>
        screenDistanceToTarget(screenPoint, a) -
        screenDistanceToTarget(screenPoint, b),
    );
  for (const target of targets) {
    if (hitTestUnit(screenPoint, point, target)) {
      return target.id;
    }
  }
  return "";
}

function hitTestUnit(screenPoint, worldPoint, unit) {
  const visual = unitVisual(unit.kind || "dummy");
  const isPlayerModel = unit.kind === "player";
  const isEnemyHeroModel = unit.kind === "enemy_hero";
  const sx = state.frame.offsetX + unit.x * state.frame.scale;
  const sy = state.frame.offsetY + unit.y * state.frame.scale;
  if (isPlayerModel) {
    return hitTestPlayerModel(screenPoint.x - sx, screenPoint.y - sy, unit);
  }
  if (isEnemyHeroModel) {
    return (
      Math.hypot(screenPoint.x - sx, screenPoint.y - sy) <= unitHitRadius(unit)
    );
  }
  const size = Math.max(18, unit.radius);
  const padding = 10;
  const dx = Math.abs(worldPoint.x - unit.x);
  const dy = Math.abs(worldPoint.y - unit.y);
  if (visual.shape === "circle") {
    return Math.hypot(dx, dy) <= size + padding;
  }
  if (visual.shape === "tower") {
    return dx <= size + padding && dy <= size * 1.3 + padding;
  }
  return dx <= size + padding && dy <= size + padding;
}

function screenDistanceToTarget(point, unit) {
  const sx = state.frame.offsetX + unit.x * state.frame.scale;
  const sy = state.frame.offsetY + unit.y * state.frame.scale;
  return Math.hypot(point.x - sx, point.y - sy);
}

function targetMap() {
  const targets = new Map(state.units);
  for (const player of state.players.values()) {
    targets.set(player.id, player);
  }
  return targets;
}

function targetLabel(target) {
  if (target.kind === "player") {
    return `Player ${target.playerId || ""}`.trim();
  }
  return unitVisual(target.kind).label;
}

function targetSelectRadius(target, frame) {
  if (target.kind === "player") {
    return playerModelRadius(target) + 4;
  }
  if (target.kind === "enemy_hero") {
    return unitModelDisplayRadius(target) + 4;
  }
  return Math.max(14, (target.radius || 18) * frame.scale + 6);
}

function playerModelRadius(player) {
  if (player.heroId === "sword") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "warrior") {
    return player.playerId === state.playerId ? 12 : 10;
  }
  if (player.heroId === "archer") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "tank") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  return player.playerId === state.playerId ? 10 : 8;
}

function playerModelShape(player) {
  if (player.heroId === "sword") {
    return "katana";
  }
  if (player.heroId === "warrior") {
    return "warrior";
  }
  if (player.heroId === "archer") {
    return "archer";
  }
  if (player.heroId === "tank") {
    return "octagon";
  }
  return "circle";
}

function hitTestPlayerModel(dx, dy, player) {
  const radius = playerModelRadius(player) + 4;
  if (playerModelShape(player) === "square") {
    return Math.abs(dx) <= radius && Math.abs(dy) <= radius;
  }
  return Math.hypot(dx, dy) <= radius;
}

function drawSwordIcon(graphics, radius) {
  const bladeHalf = radius * 0.26;
  const guardHalf = radius * 0.78;
  const guardY = radius * 0.34;
  const gripHalf = radius * 0.18;
  graphics.moveTo(0, -radius * 1.25);
  graphics.lineTo(bladeHalf, -radius * 0.75);
  graphics.lineTo(bladeHalf, guardY);
  graphics.lineTo(guardHalf, guardY);
  graphics.lineTo(guardHalf, radius * 0.58);
  graphics.lineTo(gripHalf, radius * 0.58);
  graphics.lineTo(gripHalf, radius * 1.08);
  graphics.lineTo(-gripHalf, radius * 1.08);
  graphics.lineTo(-gripHalf, radius * 0.58);
  graphics.lineTo(-guardHalf, radius * 0.58);
  graphics.lineTo(-guardHalf, guardY);
  graphics.lineTo(-bladeHalf, guardY);
  graphics.lineTo(-bladeHalf, -radius * 0.75);
  graphics.closePath();
}

function drawWarriorIcon(graphics, radius) {
  const bladeLeft = -radius * 1.08;
  const bladeRight = -radius * 0.62;
  const bladeTop = -radius * 1.24;
  const bladeBottom = radius * 0.48;
  graphics.moveTo(bladeLeft, bladeTop);
  graphics.lineTo(bladeRight, bladeTop);
  graphics.lineTo(bladeRight, bladeBottom);
  graphics.lineTo(bladeLeft, bladeBottom);
  graphics.closePath();
  graphics.moveTo(-radius * 1.38, radius * 0.2);
  graphics.lineTo(-radius * 0.34, radius * 0.2);
  graphics.lineTo(-radius * 0.34, radius * 0.48);
  graphics.lineTo(-radius * 1.38, radius * 0.48);
  graphics.closePath();
  graphics.moveTo(-radius * 1.02, radius * 0.48);
  graphics.lineTo(-radius * 0.68, radius * 0.48);
  graphics.lineTo(-radius * 0.68, radius * 1.14);
  graphics.lineTo(-radius * 1.02, radius * 1.14);
  graphics.closePath();
  graphics.circle(radius * 0.52, radius * 0.06, radius * 0.72);
  graphics.moveTo(radius * 0.18, -radius * 0.28);
  graphics.lineTo(radius * 0.86, -radius * 0.28);
  graphics.lineTo(radius * 0.86, radius * 0.4);
  graphics.lineTo(radius * 0.18, radius * 0.4);
  graphics.closePath();
}

function drawKatanaIcon(graphics, radius) {
  graphics.moveTo(radius * 0.1, -radius * 0.05);
  graphics.quadraticCurveTo(
    -radius * 0.32,
    radius * 0.48,
    -radius * 0.86,
    radius * 0.82,
  );
  graphics.quadraticCurveTo(
    -radius * 1.12,
    radius * 0.98,
    -radius * 1.24,
    radius * 0.86,
  );
  graphics.quadraticCurveTo(
    -radius * 1.12,
    radius * 1.18,
    -radius * 0.74,
    radius * 1.12,
  );
  graphics.quadraticCurveTo(
    -radius * 0.08,
    radius * 0.86,
    radius * 0.55,
    radius * 0.08,
  );
  graphics.lineTo(radius * 0.36, -radius * 0.08);
  graphics.closePath();
  graphics.moveTo(-radius * 0.1, -radius * 0.34);
  graphics.lineTo(radius * 0.78, radius * 0.14);
  graphics.lineTo(radius * 0.64, radius * 0.42);
  graphics.lineTo(-radius * 0.24, -radius * 0.06);
  graphics.closePath();
  graphics.moveTo(radius * 0.45, -radius * 1.04);
  graphics.lineTo(radius * 1.12, -radius * 0.66);
  graphics.lineTo(radius * 0.76, radius * 0.03);
  graphics.lineTo(radius * 0.1, -radius * 0.34);
  graphics.closePath();
  graphics.moveTo(-radius * 0.96, -radius * 0.5);
  graphics.lineTo(-radius * 0.46, -radius * 0.5);
  graphics.lineTo(-radius * 0.28, -radius * 0.24);
  graphics.lineTo(-radius * 0.94, -radius * 0.24);
  graphics.closePath();
}

function drawBowArrowIcon(graphics, radius) {
  graphics.moveTo(-radius * 1.22, -radius * 0.14);
  graphics.lineTo(radius * 1.05, -radius * 0.14);
  graphics.lineTo(radius * 1.05, radius * 0.14);
  graphics.lineTo(-radius * 1.22, radius * 0.14);
  graphics.closePath();
  graphics.moveTo(radius * 1.05, -radius * 0.42);
  graphics.lineTo(radius * 1.42, 0);
  graphics.lineTo(radius * 1.05, radius * 0.42);
  graphics.closePath();
  graphics.moveTo(-radius * 1.22, -radius * 0.14);
  graphics.lineTo(-radius * 1.58, -radius * 0.44);
  graphics.lineTo(-radius * 1.38, 0);
  graphics.lineTo(-radius * 1.58, radius * 0.44);
  graphics.lineTo(-radius * 1.22, radius * 0.14);
  graphics.closePath();
  graphics.moveTo(-radius * 0.32, -radius * 1.25);
  graphics.lineTo(-radius * 0.2, -radius * 1.25);
  graphics.lineTo(-radius * 0.2, -radius * 0.24);
  graphics.lineTo(-radius * 0.32, -radius * 0.24);
  graphics.closePath();
  graphics.moveTo(-radius * 0.32, radius * 0.24);
  graphics.lineTo(-radius * 0.2, radius * 0.24);
  graphics.lineTo(-radius * 0.2, radius * 1.25);
  graphics.lineTo(-radius * 0.32, radius * 1.25);
  graphics.closePath();
  graphics.moveTo(-radius * 0.18, -radius * 1.25);
  graphics.quadraticCurveTo(
    radius * 0.76,
    -radius * 1.05,
    radius * 0.78,
    -radius * 0.14,
  );
  graphics.lineTo(radius * 0.5, -radius * 0.14);
  graphics.quadraticCurveTo(
    radius * 0.42,
    -radius * 0.82,
    -radius * 0.18,
    -radius * 1.25,
  );
  graphics.closePath();
  graphics.moveTo(radius * 0.5, radius * 0.14);
  graphics.lineTo(radius * 0.78, radius * 0.14);
  graphics.quadraticCurveTo(
    radius * 0.76,
    radius * 1.05,
    -radius * 0.18,
    radius * 1.25,
  );
  graphics.quadraticCurveTo(
    radius * 0.42,
    radius * 0.82,
    radius * 0.5,
    radius * 0.14,
  );
  graphics.closePath();
}

function drawChamferedOctagon(graphics, radius) {
  const width = radius * 1.9;
  const height = radius * 1.65;
  const halfW = width / 2;
  const halfH = height / 2;
  const cornerX = radius * 0.36;
  const sideInsetY = radius * 0.52;
  graphics.moveTo(-halfW + cornerX, -halfH);
  graphics.lineTo(halfW - cornerX, -halfH);
  graphics.lineTo(halfW, -halfH + sideInsetY);
  graphics.lineTo(halfW, halfH - sideInsetY);
  graphics.lineTo(halfW - cornerX, halfH);
  graphics.lineTo(-halfW + cornerX, halfH);
  graphics.lineTo(-halfW, halfH - sideInsetY);
  graphics.lineTo(-halfW, -halfH + sideInsetY);
  graphics.closePath();
}

function unitModelRadius(unit) {
  return Math.max(18, unit.radius || 0);
}

function unitModelDisplayRadius(unit) {
  return Math.max(18, unit.radius || 0);
}

function unitCollisionRadius(unit) {
  return unitModelRadius(unit);
}

function unitHitRadius(unit) {
  return unitModelDisplayRadius(unit) + 8;
}

function screenPointFromEvent(event) {
  const rect = app.canvas.getBoundingClientRect();
  return {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top,
  };
}

function screenToWorld(event) {
  const canvasPoint = screenPointFromEvent(event);
  return {
    x: (canvasPoint.x - state.frame.offsetX) / state.frame.scale,
    y: (canvasPoint.y - state.frame.offsetY) / state.frame.scale,
  };
}

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function formatNumber(value) {
  return Number.isInteger(value)
    ? String(value)
    : String(Math.round(value * 1000) / 1000);
}

function shieldValue(entity) {
  return Math.max(0, entity?.passive?.shield || 0);
}

function formatHpWithShield(entity) {
  const stats = entity?.stats || {};
  const shield = shieldValue(entity);
  if (shield <= 0) {
    return `${formatNumber(stats.hp || 0)}/${formatNumber(stats.maxHp || 0)}`;
  }
  return `${formatNumber(stats.hp || 0)} + ${formatNumber(shield)}/${formatNumber(stats.maxHp || 0)}`;
}

function formatHpRegen5(entity) {
  const stats = entity?.stats || {};
  const base = stats.hpRegen5 || 0;
  const passive = warriorToughnessRegen5(entity);
  if (passive <= 0) {
    return formatNumber(base);
  }
  return `${formatNumber(base)} + ${formatNumber(passive)}`;
}

function warriorToughnessRegen5(entity) {
  if ((entity?.heroId || "") !== "warrior") {
    return 0;
  }
  const ratios =
    skillClientConfig.warrior_toughness?.metaLists?.regenMaxHPRatio || [];
  if (ratios.length === 0) {
    return 0;
  }
  const level = clamp(Math.max(1, entity.level || 1), 1, ratios.length);
  const ratio = ratios[level - 1] || 0;
  return (entity.stats?.maxHp || 0) * ratio;
}

function hpShieldRatio(entity) {
  const stats = entity?.stats || {};
  return ratio((stats.hp || 0) + shieldValue(entity), stats.maxHp || 0);
}

function formatAttack(stats) {
  return formatBasePlusBonus(stats.attack || 0, stats.bonusAttack || 0);
}

function formatPhysicalDefense(stats) {
  return formatBasePlusBonus(
    stats.physicalDefense || 0,
    stats.bonusPhysicalDefense || 0,
  );
}

function formatMagicDefense(stats) {
  return formatBasePlusBonus(
    stats.magicDefense || 0,
    stats.bonusMagicDefense || 0,
  );
}

function formatDefenseTip(resistance, typeLabel) {
  return `<span class="stat-tip" data-tip="${escapeHtml(formatResistanceTip(resistance, typeLabel))}">?</span>`;
}

function formatResistanceTip(resistance, typeLabel) {
  if (resistance >= 0) {
    const reduce = resistance / (resistance + 100);
    return `${typeLabel}伤害减免 ${formatPercent(reduce)}`;
  }
  const multiplier = 100 / Math.max(1, 100 + resistance);
  return `${typeLabel}伤害放大 ${formatNumber(multiplier)} 倍`;
}

function formatPercent(value) {
  return `${formatNumber(value * 100)}%`;
}

function formatBasePlusBonus(base, bonus) {
  if (bonus <= 0) {
    return formatNumber(base);
  }
  return `${formatNumber(base)} + ${formatNumber(bonus)}`;
}

function formatSwordIntent(passive) {
  return passive?.maxSwordIntent > 0
    ? `${Math.floor(passive.swordIntent || 0)}/${Math.floor(passive.maxSwordIntent)}`
    : "-";
}

function formatTargetResource(target) {
  if (target?.heroId === "sword") {
    return `<div>Sword Intent ${formatSwordIntent(target.passive || {})}</div>`;
  }
  const stats = target?.stats || {};
  if (!stats.maxMp || stats.maxMp <= 0) {
    return "";
  }
  return `<div>MP ${formatNumber(stats.mp)}/${formatNumber(stats.maxMp)}</div>`;
}

function formatTargetMpRegen(stats) {
  if (!stats?.maxMp || stats.maxMp <= 0) {
    return "";
  }
  return `<div>MP/5s ${formatNumber(stats.mpRegen5 || 0)}</div>`;
}

function formatResource(resource) {
  if (resource === "sword_intent") {
    return "Sword Intent";
  }
  if (!resource || resource === "mp") {
    return "MP";
  }
  if (resource === "none") {
    return "";
  }
  return resource;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function ratio(value, max) {
  if (!max || max <= 0) {
    return 0;
  }
  return clamp(value / max, 0, 1);
}

window.addEventListener("keydown", (event) => {
  const slot = event.key.toLowerCase();
  if (slot === "a") {
    event.preventDefault();
    state.attackMoveArmed = true;
    return;
  }
  if (!["q", "w", "e", "r"].includes(slot)) {
    return;
  }
  event.preventDefault();
  if (event.shiftKey) {
    upgradeSkill(slot);
    return;
  }
  castSkill(slot);
});

els.skills.addEventListener("pointerdown", (event) => {
  const button = event.target.closest("[data-skill-upgrade]");
  if (!button) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  if (button.disabled) {
    return;
  }
  upgradeSkill(button.dataset.skillUpgrade);
});

els.connectBtn.addEventListener("click", connect);
els.leaveBtn.addEventListener("click", leave);
els.spawnBtn.addEventListener("click", spawnObject);
els.levelUpBtn.addEventListener("click", debugLevelUp);
els.abilityHasteBtn.addEventListener("click", toggleDebugAbilityHaste);

els.serverUrl.value = websocketURL();
bootPixi();
