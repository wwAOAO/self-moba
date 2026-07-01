function draw(ticker) {
  const frame = calculateFrame();
  drawMap(frame);
  drawEffects(frame);
  syncUnits(frame);
  syncSprites(frame, ticker.deltaMS);
  syncDamageTexts(frame, ticker.deltaMS);
}

function calculateFrame() {
  const padding = 36;
  const scale = Math.min(
    (app.renderer.width - padding * 2) / state.map.width,
    (app.renderer.height - padding * 2) / state.map.height,
  );
  state.frame = {
    scale,
    offsetX: (app.renderer.width - state.map.width * scale) / 2,
    offsetY: (app.renderer.height - state.map.height * scale) / 2,
  };
  return state.frame;
}

function drawMap(frame) {
  gridLayer.clear();
  gridLayer.rect(
    frame.offsetX,
    frame.offsetY,
    state.map.width * frame.scale,
    state.map.height * frame.scale,
  );
  gridLayer.fill(0xbfd1bb);
  gridLayer.stroke({ color: 0x35594b, width: 3 });

  if (state.moveTarget) {
    gridLayer.circle(
      frame.offsetX + state.moveTarget.x * frame.scale,
      frame.offsetY + state.moveTarget.y * frame.scale,
      5,
    );
    gridLayer.fill(0x22c55e);
  }

  drawActiveSkillRanges(frame);
  drawSwordETargetCooldowns(frame);
  drawCastWindups(frame);
  drawSkillPreview(frame);

  const selectedTarget = state.selectedTargetId
    ? targetMap().get(state.selectedTargetId)
    : null;
  if (selectedTarget) {
    gridLayer.circle(
      frame.offsetX + selectedTarget.x * frame.scale,
      frame.offsetY + selectedTarget.y * frame.scale,
      targetSelectRadius(selectedTarget, frame),
    );
    gridLayer.stroke({ color: 0xf6d365, width: 3 });
  }

  drawAttackFlash(frame);

  if (state.attackTargetId) {
    const target = targetMap().get(state.attackTargetId);
    if (target && target.id !== state.selectedTargetId) {
      gridLayer.circle(
        frame.offsetX + target.x * frame.scale,
        frame.offsetY + target.y * frame.scale,
        targetSelectRadius(target, frame),
      );
      gridLayer.stroke({ color: 0xf6d365, width: 3 });
    }
  }
}

function drawAttackFlash(frame) {
  const flash = state.attackFlash;
  if (!flash) {
    return;
  }
  if (performance.now() >= flash.until) {
    state.attackFlash = null;
    return;
  }
  gridLayer.circle(
    frame.offsetX + flash.x * frame.scale,
    frame.offsetY + flash.y * frame.scale,
    (flash.radius || 0) * frame.scale,
  );
  gridLayer.stroke({ color: 0x2f6fdd, width: 2, alpha: 0.75 });
}

function drawEffects(frame) {
  for (const effect of state.effects) {
    if (effect.kind === "sword_whirlwind") {
      drawSwordWhirlwindEffect(effect, frame);
      continue;
    }
    if (effect.kind === "tank_q") {
      drawTankShardEffect(effect, frame);
      continue;
    }
    if (effect.kind === "tank_w_aftershock") {
      drawTankAftershockEffect(effect, frame);
      continue;
    }
    if (effect.kind === "tank_r_impact") {
      drawTankImpactEffect(effect, frame);
      continue;
    }
    if (effect.kind === "basic_arrow") {
      drawBasicArrowEffect(effect, frame);
      continue;
    }
    if (effect.kind === "archer_volley_arrow") {
      drawVolleyArrowEffect(effect, frame);
      continue;
    }
    if (effect.kind === "archer_hawk") {
      drawArcherHawkEffect(effect, frame);
      continue;
    }
    if (effect.kind === "archer_crystal_arrow") {
      drawCrystalArrowEffect(effect, frame);
      continue;
    }
    if (effect.kind !== "wind_wall") {
      continue;
    }
    const half = effect.width / 2;
    const startX =
      frame.offsetX + (effect.x - effect.dirX * half) * frame.scale;
    const startY =
      frame.offsetY + (effect.y - effect.dirY * half) * frame.scale;
    const endX = frame.offsetX + (effect.x + effect.dirX * half) * frame.scale;
    const endY = frame.offsetY + (effect.y + effect.dirY * half) * frame.scale;
    gridLayer.moveTo(startX, startY);
    gridLayer.lineTo(endX, endY);
    gridLayer.stroke({ color: 0x67e8f9, width: 10, alpha: 0.45 });
    gridLayer.moveTo(startX, startY);
    gridLayer.lineTo(endX, endY);
    gridLayer.stroke({ color: 0x0e7490, width: 2, alpha: 0.9 });
  }
}

function drawActiveSkillRanges(frame) {
  const tick = interpolatedTick();
  for (const player of state.players.values()) {
    if (player.dead) {
      continue;
    }
    if (player.heroId !== "warrior") {
      continue;
    }
    drawWarriorJudgmentRange(player, frame, tick);
  }
}

function drawWarriorJudgmentRange(player, frame, tick) {
  if ((player.warrior?.judgmentUntilTick || 0) <= tick) {
    return;
  }
  const config = skillClientConfig.judgment || {};
  const radius = config.range || 180;
  const hitRadius = radius + unitCollisionRadius({ radius: 18 });
  const x = frame.offsetX + player.x * frame.scale;
  const y = frame.offsetY + player.y * frame.scale;
  gridLayer.circle(x, y, radius * frame.scale);
  gridLayer.fill({ color: 0xf59e0b, alpha: 0.08 });
  gridLayer.circle(x, y, radius * frame.scale);
  gridLayer.stroke({ color: 0xf97316, width: 3, alpha: 0.75 });
  gridLayer.circle(x, y, hitRadius * frame.scale);
  gridLayer.stroke({ color: 0xf97316, width: 1, alpha: 0.35 });
}

function drawTankAftershockEffect(effect, frame) {
  const range = effect.range || 300;
  const angle = ((effect.radius || 70) * Math.PI) / 180;
  const center = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const startAngle = center - angle / 2;
  const endAngle = center + angle / 2;
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = range * frame.scale;
  const alpha = effectAlpha(effect);
  gridLayer.moveTo(x, y);
  gridLayer.arc(x, y, radius, startAngle, endAngle);
  gridLayer.closePath();
  gridLayer.fill({ color: 0xfacc15, alpha: 0.14 * alpha });
  gridLayer.moveTo(x, y);
  gridLayer.arc(x, y, radius, startAngle, endAngle);
  gridLayer.closePath();
  gridLayer.stroke({ color: 0xd97706, width: 2, alpha: 0.8 * alpha });
}

function drawTankImpactEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 250) * frame.scale;
  const alpha = effectAlpha(effect);
  gridLayer.circle(x, y, radius);
  gridLayer.fill({ color: 0x94a3b8, alpha: 0.14 * alpha });
  gridLayer.circle(x, y, radius);
  gridLayer.stroke({ color: 0x475569, width: 3, alpha: 0.85 * alpha });
  gridLayer.circle(x, y, Math.max(8, radius * 0.12));
  gridLayer.fill({ color: 0xe2e8f0, alpha: 0.22 * alpha });
}

function drawBasicArrowEffect(effect, frame) {
  if ((effect.count || 1) >= 3) {
    drawTripleArrowProjectile(effect, frame, 0xf8d36a, 0xf59e0b);
    return;
  }
  drawArrowProjectile(effect, frame, 0xf8d36a, 0xf59e0b);
}

function drawVolleyArrowEffect(effect, frame) {
  if (state.hiddenEffectIds.has(effect.id)) {
    return;
  }
  drawArrowProjectile(effect, frame, 0xbae6fd, 0x38bdf8, {
    fromSnapshot: true,
    hideOnEnemyHit: true,
  });
}

function drawCrystalArrowEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = (effect.radius || 130) * frame.scale;
  gridLayer.circle(x, y, radius);
  gridLayer.fill({ color: 0xa78bfa, alpha: 0.08 });
  gridLayer.circle(x, y, radius);
  gridLayer.stroke({ color: 0x8b5cf6, width: 2, alpha: 0.7 });
  drawArrowProjectile(effect, frame, 0xc4b5fd, 0x7c3aed, {
    fromSnapshot: true,
  });
}

function drawArcherHawkEffect(effect, frame) {
  const tick = interpolatedTick();
  const arriveTick = effect.height || effect.createdAt || tick;
  const arrived = tick >= arriveTick;
  const progress = arrived
    ? 1
    : clamp(
        (tick - (effect.createdAt || tick)) /
          Math.max(1, arriveTick - (effect.createdAt || tick)),
        0,
        1,
      );
  const worldX =
    (effect.x || 0) +
    ((effect.endX || effect.x || 0) - (effect.x || 0)) * progress;
  const worldY =
    (effect.y || 0) +
    ((effect.endY || effect.y || 0) - (effect.y || 0)) * progress;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(14, (effect.radius || 80) * frame.scale);
  if (arrived) {
    const alpha = effectAlpha(effect);
    gridLayer.circle(x, y, radius);
    gridLayer.fill({ color: 0x38bdf8, alpha: 0.08 * alpha });
    gridLayer.circle(x, y, radius);
    gridLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.7 * alpha });
  }
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const size = arrived ? 10 : 14;
  gridLayer
    .moveTo(x + Math.cos(angle) * size, y + Math.sin(angle) * size)
    .lineTo(
      x + Math.cos(angle + 2.45) * size,
      y + Math.sin(angle + 2.45) * size,
    )
    .lineTo(
      x + Math.cos(angle + Math.PI) * size * 0.35,
      y + Math.sin(angle + Math.PI) * size * 0.35,
    )
    .lineTo(
      x + Math.cos(angle - 2.45) * size,
      y + Math.sin(angle - 2.45) * size,
    )
    .closePath();
  gridLayer.fill({ color: 0x0ea5e9, alpha: arrived ? 0.8 : 0.95 });
}

function drawArrowProjectile(effect, frame, shaftColor, headColor, options = {}) {
  const position = projectileDrawPosition(effect, options);
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const length = 26;
  const width = 5;
  if (options.hideOnEnemyHit && volleyArrowHitsEnemy(effect, frame, x, y, angle, length)) {
    state.hiddenEffectIds.add(effect.id);
    return;
  }
  gridLayer
    .moveTo(
      x + Math.cos(angle) * length * 0.5,
      y + Math.sin(angle) * length * 0.5,
    )
    .lineTo(
      x - Math.cos(angle) * length * 0.5,
      y - Math.sin(angle) * length * 0.5,
    );
  gridLayer.stroke({ color: shaftColor, width: 3, alpha: 0.95 });
  gridLayer
    .moveTo(
      x + Math.cos(angle) * length * 0.5,
      y + Math.sin(angle) * length * 0.5,
    )
    .lineTo(
      x + Math.cos(angle + Math.PI * 0.82) * width,
      y + Math.sin(angle + Math.PI * 0.82) * width,
    )
    .lineTo(
      x + Math.cos(angle - Math.PI * 0.82) * width,
      y + Math.sin(angle - Math.PI * 0.82) * width,
    )
    .closePath();
  gridLayer.fill({ color: headColor, alpha: 0.95 });
}

function projectileDrawPosition(effect, options = {}) {
  const tick = interpolatedTick();
  const baseTick = options.fromSnapshot
    ? state.snapshotTick
    : (effect.createdAt ?? tick);
  const traveled = Math.max(0, tick - baseTick) * (effect.speed || 0);
  return {
    x: (effect.x || 0) + (effect.dirX || 1) * traveled,
    y: (effect.y || 0) + (effect.dirY || 0) * traveled,
  };
}

function volleyArrowHitsEnemy(effect, frame, x, y, angle, length) {
  const halfLength = length * 0.5;
  const start = {
    x: x - Math.cos(angle) * halfLength,
    y: y - Math.sin(angle) * halfLength,
  };
  const end = {
    x: x + Math.cos(angle) * halfLength,
    y: y + Math.sin(angle) * halfLength,
  };
  const arrowRadius = 5;
  for (const target of targetMap().values()) {
    if (!target || target.dead || target.team === effect.team) {
      continue;
    }
    const targetX = frame.offsetX + target.x * frame.scale;
    const targetY = frame.offsetY + target.y * frame.scale;
    const radius = targetScreenRadius(target, frame);
    if (distancePointToSegment({ x: targetX, y: targetY }, start, end) <= arrowRadius + radius) {
      return true;
    }
  }
  return false;
}

function distancePointToSegment(point, start, end) {
  const dx = end.x - start.x;
  const dy = end.y - start.y;
  const lengthSquared = dx * dx + dy * dy;
  if (lengthSquared <= 0) {
    return Math.hypot(point.x - start.x, point.y - start.y);
  }
  const t = Math.max(
    0,
    Math.min(1, ((point.x - start.x) * dx + (point.y - start.y) * dy) / lengthSquared),
  );
  const closestX = start.x + dx * t;
  const closestY = start.y + dy * t;
  return Math.hypot(point.x - closestX, point.y - closestY);
}

function targetScreenRadius(target, frame) {
  if (typeof targetSelectRadius === "function") {
    return targetSelectRadius(target, frame);
  }
  return Math.max(14, (target.radius || 18) * frame.scale + 6);
}

function drawTripleArrowProjectile(effect, frame, shaftColor, headColor) {
  const arrows = [
    { forward: -82, side: 0 },
    { forward: 0, side: -25 },
    { forward: 82, side: 25 },
  ];
  for (const arrow of arrows) {
    drawArrowProjectile(
      {
        ...effect,
        x:
          (effect.x || 0) +
          (effect.dirX || 1) * arrow.forward -
          (effect.dirY || 0) * arrow.side,
        y:
          (effect.y || 0) +
          (effect.dirY || 0) * arrow.forward +
          (effect.dirX || 1) * arrow.side,
      },
      frame,
      shaftColor,
      headColor,
    );
  }
}

function effectAlpha(effect) {
  const createdAt = effect.createdAt || 0;
  const expiresAt = effect.expiresAt || 0;
  const duration = Math.max(1, expiresAt - createdAt);
  return clamp((expiresAt - interpolatedTick()) / duration, 0, 1);
}

function drawSwordETargetCooldowns(frame) {
  const self = state.players.get(state.playerId);
  if (!self || self.heroId !== "sword") {
    return;
  }
  const targetUntil = self.sword?.sweepingBladeTargetUntil || {};
  const tick = interpolatedTick();
  const targets = targetMap();
  const cooldownTicks = swordETargetCooldownTicks(self);
  for (const [targetId, untilTick] of Object.entries(targetUntil)) {
    const remainingTicks = (untilTick || 0) - tick;
    if (remainingTicks <= 0) {
      continue;
    }
    const target = targets.get(targetId);
    if (!target || target.dead) {
      continue;
    }
    const x = frame.offsetX + target.x * frame.scale;
    const y = frame.offsetY + target.y * frame.scale;
    const radius = targetSelectRadius(target, frame) + 5;
    const progress = ratio(remainingTicks, cooldownTicks);
    const startAngle = -Math.PI / 2;
    const endAngle = startAngle + Math.PI * 2 * progress;
    const startX = x + Math.cos(startAngle) * radius;
    const startY = y + Math.sin(startAngle) * radius;
    gridLayer.circle(x, y, radius);
    gridLayer.stroke({ color: 0x7dd3fc, width: 4, alpha: 0.25 });
    gridLayer.moveTo(startX, startY);
    gridLayer.arc(x, y, radius, startAngle, endAngle);
    gridLayer.stroke({ color: 0x38bdf8, width: 4, alpha: 0.9 });
  }
}

function drawSwordWhirlwindEffect(effect, frame) {
  const tick = interpolatedTick();
  const ageTicks = Math.max(0, tick - (effect.createdAt || tick));
  const traveled = clamp(ageTicks * (effect.speed || 0), 0, effect.range || 0);
  const x = effect.x + (effect.dirX || 0) * traveled;
  const y = effect.y + (effect.dirY || 0) * traveled;
  const sx = frame.offsetX + x * frame.scale;
  const sy = frame.offsetY + y * frame.scale;
  const radius = (effect.radius || 70) * frame.scale;
  gridLayer.circle(sx, sy, radius);
  gridLayer.stroke({ color: 0x0284c7, width: 3, alpha: 0.9 });
  gridLayer.circle(sx, sy, Math.max(6, radius * 0.35));
  gridLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.8 });
  gridLayer.moveTo(sx - radius * 0.55, sy);
  gridLayer.lineTo(sx + radius * 0.55, sy);
  gridLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.45 });
}

function drawTankShardEffect(effect, frame) {
  const tickDelta = clamp(
    interpolatedTick() - Number(els.tick.textContent || 0),
    0,
    1,
  );
  const smoothX =
    effect.x + (effect.dirX || 0) * (effect.speed || 0) * tickDelta;
  const smoothY =
    effect.y + (effect.dirY || 0) * (effect.speed || 0) * tickDelta;
  const sx = frame.offsetX + smoothX * frame.scale;
  const sy = frame.offsetY + smoothY * frame.scale;
  const radius = Math.max(5, (effect.radius || 45) * frame.scale);
  gridLayer.circle(sx, sy, radius * 0.65);
  gridLayer.fill({ color: 0x8b5e34, alpha: 0.85 });
  gridLayer.circle(sx, sy, radius);
  gridLayer.stroke({ color: 0x5c4033, width: 2, alpha: 0.75 });
}

function drawCastWindups(frame) {
  const now = performance.now();
  const activeWindups = [];
  for (const windup of state.castWindups) {
    if (now <= windup.expiresAt) {
      activeWindups.push(windup);
      continue;
    }
    finishCastWindup(windup);
  }
  state.castWindups = activeWindups;
  for (const windup of state.castWindups) {
    drawCastWindup(windup, frame, now);
  }
}

function finishCastWindup(windup) {
  if (windup.finished) {
    return;
  }
  windup.finished = true;
  if (windup.skillId !== "sword_cut") {
    return;
  }
  showSwordQReleasePreview(windup);
}

function showSwordQReleasePreview(windup) {
  const self = state.players.get(state.playerId);
  if (!self || self.dead) {
    return;
  }
  if (windup.preview) {
    showSwordQPreviewFromData(windup.preview);
    return;
  }
  showSwordQPreview(self, {
    x: windup.targetX,
    y: windup.targetY,
  });
}

function drawCastWindup(windup, frame, now) {
  const progress = clamp(
    (now - windup.startedAt) / Math.max(1, windup.durationMs || 1),
    0,
    1,
  );
  const alpha = 1 - progress * 0.35;
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  const pulseRadius = (20 + 18 * progress) * frame.scale;
  const color = castWindupColor(windup.skillId);
  gridLayer.circle(x, y, pulseRadius);
  gridLayer.stroke({ color, width: 3, alpha: 0.7 * alpha });
  gridLayer.circle(x, y, Math.max(5, pulseRadius * 0.18));
  gridLayer.fill({ color, alpha: 0.18 * alpha });
  const angleStart = -Math.PI / 2;
  gridLayer.moveTo(
    x + Math.cos(angleStart) * (pulseRadius + 6),
    y + Math.sin(angleStart) * (pulseRadius + 6),
  );
  gridLayer.arc(
    x,
    y,
    pulseRadius + 6,
    angleStart,
    angleStart + Math.PI * 2 * progress,
  );
  gridLayer.stroke({ color, width: 4, alpha: 0.85 });

  if (windup.skillId === "sword_cut") {
    drawSwordQWindup(windup, frame, color, alpha);
    return;
  }
  if (windup.skillId === "slam") {
    drawDirectionalWindup(windup, frame, color, alpha, 18);
    return;
  }
  if (windup.skillId === "taunt") {
    drawCircleWindup(windup, frame, color, alpha, windup.range || 400);
    return;
  }
  if (windup.skillId === "justice") {
    drawTargetLockWindup(windup, frame, color, alpha);
    return;
  }
  if (windup.skillId === "arrow_rain") {
    drawDirectionalWindup(windup, frame, color, alpha, 10);
  }
}

function drawSwordQWindup(windup, frame, color, alpha) {
  return;
}

function drawDirectionalWindup(windup, frame, color, alpha, width) {
  const range = windup.range || 475;
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  const endX =
    frame.offsetX + (windup.x + (windup.dirX || 1) * range) * frame.scale;
  const endY =
    frame.offsetY + (windup.y + (windup.dirY || 0) * range) * frame.scale;
  gridLayer.moveTo(x, y);
  gridLayer.lineTo(endX, endY);
  gridLayer.stroke({ color, width, alpha: 0.16 * alpha });
  gridLayer.moveTo(x, y);
  gridLayer.lineTo(endX, endY);
  gridLayer.stroke({ color, width: 2, alpha: 0.72 * alpha });
}

function drawCircleWindup(windup, frame, color, alpha, range) {
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  gridLayer.circle(x, y, range * frame.scale);
  gridLayer.fill({ color, alpha: 0.06 * alpha });
  gridLayer.circle(x, y, range * frame.scale);
  gridLayer.stroke({ color, width: 3, alpha: 0.55 * alpha });
}

function drawTargetLockWindup(windup, frame, color, alpha) {
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  const tx = frame.offsetX + windup.targetX * frame.scale;
  const ty = frame.offsetY + windup.targetY * frame.scale;
  gridLayer.moveTo(x, y);
  gridLayer.lineTo(tx, ty);
  gridLayer.stroke({ color, width: 3, alpha: 0.55 * alpha });
  gridLayer.circle(tx, ty, 26);
  gridLayer.stroke({ color, width: 3, alpha: 0.8 * alpha });
  gridLayer.circle(tx, ty, 8);
  gridLayer.fill({ color, alpha: 0.18 * alpha });
}

function castWindupColor(skillId) {
  if (skillId === "justice") {
    return 0xf97316;
  }
  if (skillId === "slam") {
    return 0x8b5e34;
  }
  if (skillId === "taunt") {
    return 0x64748b;
  }
  if (skillId === "arrow_rain") {
    return 0xa78bfa;
  }
  return 0x38bdf8;
}

function interpolatedTick() {
  if (!state.snapshotAtMs) {
    return Number(els.tick.textContent || 0);
  }
  return (
    state.snapshotTick +
    ((performance.now() - state.snapshotAtMs) / 1000) * state.tickRate
  );
}

function showSwordQPreview(self, target) {
  const preview = swordQPreviewData(self, target);
  if (!preview) {
    return;
  }
  showSwordQPreviewFromData(preview);
}

function swordQPreviewData(self, target) {
  const tick = Number(els.tick.textContent || 0);
  const qState = skillState(self, "sword_cut");
  if ((qState?.level || 0) <= 0) {
    return null;
  }
  const config = skillClientConfig.sword_cut || {};
  let form = "line";
  let range = config.range || 475;
  if (swordEQWindowActive(self, config, tick)) {
    form = "circle";
    range = config.eqRadius || 375;
  } else if ((qState?.stacks || 0) >= 2) {
    form = "whirlwind";
    range = config.whirlwindRange || 900;
  }
  const dx = target.x - self.x;
  const dy = target.y - self.y;
  const len = Math.hypot(dx, dy) || 1;
  return {
    kind: "sword_q",
    form,
    x: self.x,
    y: self.y,
    dirX: dx / len,
    dirY: dy / len,
    range,
    radius: form === "whirlwind" ? config.whirlwindRadius || 70 : 0,
    previewMs: config.previewMs || 450,
  };
}

function showSwordQPreviewFromData(preview) {
  const previewMs = preview.previewMs || 450;
  state.skillPreview = {
    ...preview,
    previewMs,
    expiresAt: performance.now() + previewMs,
  };
}

function showTankEPreview(self) {
  const config = skillClientConfig.taunt || {};
  const previewMs = config.previewMs || 450;
  state.skillPreview = {
    kind: "tank_e",
    form: "circle",
    x: self.x,
    y: self.y,
    range: config.range || 400,
    previewMs,
    expiresAt: performance.now() + previewMs,
  };
}

function swordEQWindowActive(player, config, tick) {
  const dashUntilTick = player.control?.dashUntilTick || 0;
  if (dashUntilTick <= tick) {
    return false;
  }
  const windowTicks = (config.eqWindowSeconds || 0.15) * state.tickRate;
  return dashUntilTick - tick <= windowTicks;
}

function drawSkillPreview(frame) {
  const preview = state.skillPreview;
  if (!preview) {
    return;
  }
  if (performance.now() > preview.expiresAt) {
    state.skillPreview = null;
    return;
  }
  const alpha = Math.max(
    0,
    (preview.expiresAt - performance.now()) / (preview.previewMs || 450),
  );
  const x = frame.offsetX + preview.x * frame.scale;
  const y = frame.offsetY + preview.y * frame.scale;
  if (preview.form === "circle") {
    gridLayer.circle(x, y, preview.range * frame.scale);
    gridLayer.stroke({ color: 0x38bdf8, width: 3, alpha: 0.65 * alpha });
    gridLayer.circle(x, y, 12);
    gridLayer.fill({ color: 0x38bdf8, alpha: 0.18 * alpha });
    return;
  }
  const endX =
    frame.offsetX + (preview.x + preview.dirX * preview.range) * frame.scale;
  const endY =
    frame.offsetY + (preview.y + preview.dirY * preview.range) * frame.scale;
  if (preview.form === "whirlwind") {
    gridLayer.moveTo(x, y);
    gridLayer.lineTo(endX, endY);
    gridLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.55 * alpha });
    gridLayer.circle(endX, endY, (preview.radius || 70) * frame.scale);
    gridLayer.stroke({ color: 0x0284c7, width: 3, alpha: 0.85 * alpha });
    gridLayer.circle(
      endX,
      endY,
      Math.max(6, (preview.radius || 70) * frame.scale * 0.28),
    );
    gridLayer.fill({ color: 0x38bdf8, alpha: 0.16 * alpha });
    return;
  }
  const width = preview.form === "whirlwind" ? 18 : 12;
  gridLayer.moveTo(x, y);
  gridLayer.lineTo(endX, endY);
  gridLayer.stroke({ color: 0x38bdf8, width, alpha: 0.28 * alpha });
  gridLayer.moveTo(x, y);
  gridLayer.lineTo(endX, endY);
  gridLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.8 * alpha });
}

function syncSprites(frame, deltaMS) {
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
    redrawPlayerBody(sprite, player);
    updateBars(sprite, player);
    updatePlayerLabel(sprite, player);
    updateStatusLabel(sprite, player, -43);
    sprite.x += (sprite.targetX - sprite.x) * smoothing;
    sprite.y += (sprite.targetY - sprite.y) * smoothing;
    sprite.node.x = frame.offsetX + sprite.x * frame.scale;
    sprite.node.y = frame.offsetY + sprite.y * frame.scale;
  }
}

function syncUnits(frame) {
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
    updateUnitBars(sprite, unit);
    updateStatusLabel(sprite, unit, -(unitModelDisplayRadius(unit) + 30));
    sprite.node.x = frame.offsetX + unit.x * frame.scale;
    sprite.node.y = frame.offsetY + unit.y * frame.scale;
  }
}

function createPlayer(player) {
  const node = new PIXI.Container();
  const body = new PIXI.Graphics();
  const hpBack = new PIXI.Graphics();
  const hpFill = new PIXI.Graphics();
  const resourceBack = new PIXI.Graphics();
  const resourceFill = new PIXI.Graphics();
  const statusLabel = createStatusLabel();
  const label = new PIXI.Text({
    text: player.playerId,
    style: {
      fill: 0x172026,
      fontFamily: "Arial",
      fontSize: 13,
      fontWeight: "700",
    },
  });

  redrawPlayerBody({ body }, player);
  drawBar(hpBack, 0x24312b, 1, -29);
  drawBar(hpFill, 0xd94948, 1, -29);
  drawBar(resourceBack, 0x24312b, 1, -23);
  drawBar(
    resourceFill,
    playerResourceColor(player),
    playerResourceRatio(player),
    -23,
  );
  label.anchor.set(0.5, 0);
  label.y = 16;
  node.addChild(
    statusLabel,
    hpBack,
    hpFill,
    resourceBack,
    resourceFill,
    body,
    label,
  );
  return {
    node,
    body,
    x: player.x,
    y: player.y,
    targetX: player.x,
    targetY: player.y,
    hpFill,
    resourceFill,
    label,
    statusLabel,
  };
}

function updatePlayerLabel(sprite, player) {
  if (!sprite.label) {
    return;
  }
  sprite.label.text = player.dead
    ? `${player.playerId} ${Math.ceil(player.respawnIn || 0)}s`
    : player.playerId;
  sprite.label.alpha = player.dead ? 0.65 : 1;
}

function redrawPlayerBody(sprite, player) {
  const isSelf = player.playerId === state.playerId;
  const radius = playerModelRadius(player);
  sprite.body.clear();
  const shape = playerModelShape(player);
  if (shape === "triangle") {
    sprite.body.moveTo(0, -radius);
    sprite.body.lineTo(radius * 0.92, radius * 0.7);
    sprite.body.lineTo(-radius * 0.92, radius * 0.7);
    sprite.body.closePath();
  } else if (shape === "archer") {
    drawBowArrowIcon(sprite.body, radius);
  } else if (shape === "square") {
    sprite.body.rect(-radius, -radius, radius * 2, radius * 2);
  } else if (shape === "octagon") {
    drawChamferedOctagon(sprite.body, radius);
  } else if (shape === "katana") {
    drawKatanaIcon(sprite.body, radius);
  } else if (shape === "warrior") {
    drawWarriorIcon(sprite.body, radius);
  } else if (shape === "sword") {
    drawSwordIcon(sprite.body, radius);
  } else {
    sprite.body.circle(0, 0, radius);
  }
  sprite.body.fill(player.dead ? 0x6b7280 : colorForTeam(player.team));
  if (shape !== "archer") {
    sprite.body.stroke({
      color: player.dead ? 0x111827 : isSelf ? 0xffffff : 0x172026,
      width: isSelf ? 2 : 1,
      alpha: player.dead ? 0.45 : 1,
    });
  }
}

function createUnit(unit) {
  const node = new PIXI.Container();
  const body = new PIXI.Graphics();
  const hpFill = new PIXI.Graphics();
  const visual = unitVisual(unit.kind);
  const statusLabel = createStatusLabel();
  const label = new PIXI.Text({
    text: visual.label,
    style: {
      fill: 0x172026,
      fontFamily: "Arial",
      fontSize: 13,
      fontWeight: "700",
    },
  });

  const modelRadius = unitModelDisplayRadius(unit);
  drawUnitBody(body, visual, modelRadius);
  body.stroke({ color: 0xf2f7f3, width: 2 });
  drawBar(hpFill, 0xd94948, 1, -(modelRadius + 16));
  label.anchor.set(0.5, 0);
  label.y = modelRadius + 6;
  node.addChild(statusLabel, hpFill, body, label);
  return { node, hpFill, statusLabel };
}

function createStatusLabel() {
  const label = new PIXI.Text({
    text: "",
    style: {
      fill: 0xf97316,
      fontFamily: "Arial",
      fontSize: 12,
      fontWeight: "900",
      stroke: { color: 0xffffff, width: 2 },
    },
  });
  label.anchor.set(0.5, 0.5);
  label.visible = false;
  return label;
}

function updateStatusLabel(sprite, target, y) {
  if (!sprite.statusLabel) {
    return;
  }
  if (target.kind !== "player" && target.kind !== "enemy_hero") {
    sprite.statusLabel.visible = false;
    return;
  }
  const statuses = abnormalStatuses(target);
  if (!statuses.length) {
    sprite.statusLabel.visible = false;
    return;
  }
  sprite.statusLabel.text = statuses.join(" ");
  sprite.statusLabel.y = y;
  sprite.statusLabel.visible = true;
}

function abnormalStatuses(target) {
  const tick = Number(els.tick.textContent || 0);
  const statuses = [];
  if ((target.control?.airborneUntilTick || 0) > tick) {
    statuses.push("Airborne");
  }
  if ((target.control?.actionLockedUntilTick || 0) > tick) {
    statuses.push("Locked");
  }
  if ((target.control?.stunnedUntilTick || 0) > tick) {
    statuses.push("Stun");
  }
  if ((target.control?.silencedUntilTick || 0) > tick) {
    statuses.push("Silence");
  }
  if ((target.control?.tenacityUntilTick || 0) > tick) {
    statuses.push("Tenacity");
  }
  if ((target.control?.moveSpeedSlowUntil || 0) > tick) {
    statuses.push("Slow");
  }
  return statuses;
}

function normalizeUnit(unit) {
  return {
    ...unit,
    kind: unit.kind || "dummy",
    team: unit.team || "neutral",
  };
}

function visibleUnits(snapshot) {
  const units = snapshot.units || snapshot.dummies || [];
  return units.filter((unit) => {
    const kind = unit.kind || "dummy";
    if (kind === "dummy") {
      return state.showDummies || String(unit.id || "").startsWith("spawn:");
    }
    return String(unit.id || "").startsWith("spawn:");
  });
}

function normalizePlayer(player) {
  const isSelf = player.playerId === state.playerId;
  return {
    ...player,
    id: `player:${player.playerId}`,
    kind: "player",
    radius: 18,
    team: isSelf ? player.team || state.team : player.team || "unknown",
  };
}

function unitVisual(kind) {
  const visuals = {
    enemy_hero: { label: "Enemy Hero", color: 0xdc2626, shape: "circle" },
    siege_minion: { label: "Cannon", color: 0x6b7280, shape: "rect" },
    melee_minion: { label: "Melee", color: 0xf97316, shape: "circle" },
    ranged_minion: { label: "Ranged", color: 0xfacc15, shape: "diamond" },
    tower: { label: "Tower", color: 0x475569, shape: "tower" },
    crystal: { label: "Crystal", color: 0xa855f7, shape: "diamond" },
    barracks: { label: "Barracks", color: 0x7c2d12, shape: "rect" },
    dummy: { label: "Dummy", color: 0x8a5a32, shape: "rect" },
  };
  return visuals[kind] || { label: kind, color: 0x334155, shape: "circle" };
}

function drawUnitBody(body, visual, radius) {
  const size = Math.max(1, radius);
  if (visual.shape === "diamond") {
    body.moveTo(0, -size);
    body.lineTo(size, 0);
    body.lineTo(0, size);
    body.lineTo(-size, 0);
    body.closePath();
    body.fill(visual.color);
    return;
  }
  if (visual.shape === "tower") {
    body.rect(-size * 0.75, -size, size * 1.5, size * 2);
    body.fill(visual.color);
    body.rect(-size, -size * 1.3, size * 2, size * 0.45);
    body.fill(0x64748b);
    return;
  }
  if (visual.shape === "rect") {
    body.roundRect(-size, -size, size * 2, size * 2, 4);
    body.fill(visual.color);
    return;
  }
  body.circle(0, 0, size);
  body.fill(visual.color);
}

function spawnDamageText(target, damage, damageType) {
  if (!damage) {
    return;
  }
  const text = new PIXI.Text({
    text: `-${damage}`,
    style: {
      fill: damageTextColor(damageType),
      fontFamily: "Arial",
      fontSize: 15,
      fontWeight: "900",
      stroke: { color: 0xffffff, width: 2 },
    },
  });
  text.anchor.set(0.5, 0.5);
  effectLayer.addChild(text);
  state.damageTexts.push({
    node: text,
    x: target.x,
    y: target.y - 42,
    age: 0,
    lifetime: 720,
  });
}

function damageTextColor(damageType) {
  if (damageType === "magic") {
    return 0x8b5cf6;
  }
  if (damageType === "true") {
    return 0xffffff;
  }
  return 0xff3333;
}

function syncDamageTexts(frame, deltaMS) {
  for (let i = state.damageTexts.length - 1; i >= 0; i--) {
    const effect = state.damageTexts[i];
    effect.age += deltaMS;
    const progress = effect.age / effect.lifetime;
    if (progress >= 1) {
      effectLayer.removeChild(effect.node);
      state.damageTexts.splice(i, 1);
      continue;
    }
    effect.node.x = frame.offsetX + effect.x * frame.scale;
    effect.node.y = frame.offsetY + (effect.y - progress * 80) * frame.scale;
    effect.node.alpha = 1 - progress;
  }
}

function updateBars(sprite, target) {
  const stats = target?.stats || target;
  if (!stats) {
    return;
  }
  drawBar(sprite.hpFill, 0xd94948, hpShieldRatio(target), -29);
  if (sprite.resourceFill) {
    drawBar(
      sprite.resourceFill,
      playerResourceColor(target),
      playerResourceRatio(target),
      -23,
    );
  }
}

function updateUnitBars(sprite, unit) {
  drawBar(
    sprite.hpFill,
    0xd94948,
    hpShieldRatio(unit),
    -(unitModelDisplayRadius(unit) + 16),
  );
}

function playerResourceRatio(player) {
  const passive = player?.passive || {};
  if ((player?.heroId || els.heroId.value) === "sword") {
    return ratio(passive.swordIntent || 0, passive.maxSwordIntent || 0);
  }
  return ratio(player?.stats?.mp || 0, player?.stats?.maxMp || 0);
}

function playerResourceColor(player) {
  if ((player?.heroId || els.heroId.value) === "sword") {
    return 0xf8fafc;
  }
  return 0x3b82f6;
}

function colorForTeam(team) {
  if (team === "red") {
    return 0xef4444;
  }
  if (team === "blue") {
    return 0x2563eb;
  }
  return 0x94a3b8;
}

function drawBar(graphics, color, value, y) {
  graphics.clear();
  graphics.roundRect(-18, y, 36 * value, 4, 1);
  graphics.fill(color);
}
