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
  const origin = castWindupOrigin(windup);
  const x = frame.offsetX + origin.x * frame.scale;
  const y = frame.offsetY + origin.y * frame.scale;
  const pulseRadius = (20 + 18 * progress) * frame.scale;
  const color = castWindupColor(windup.skillId);
  skillLayer.circle(x, y, pulseRadius);
  skillLayer.stroke({ color, width: 3, alpha: 0.7 * alpha });
  skillLayer.circle(x, y, Math.max(5, pulseRadius * 0.18));
  skillLayer.fill({ color, alpha: 0.18 * alpha });
  const angleStart = -Math.PI / 2;
  skillLayer.moveTo(
    x + Math.cos(angleStart) * (pulseRadius + 6),
    y + Math.sin(angleStart) * (pulseRadius + 6),
  );
  skillLayer.arc(
    x,
    y,
    pulseRadius + 6,
    angleStart,
    angleStart + Math.PI * 2 * progress,
  );
  skillLayer.stroke({ color, width: 4, alpha: 0.85 });

  if (windup.skillId === "sword_cut") {
    drawSwordQWindup(windup, frame, color, alpha);
    return;
  }
  if (windup.skillId === "berserker_q") {
    drawBerserkerQWindup(windup, frame, alpha);
    return;
  }
  if (windup.skillId === "berserker_e") {
    drawBerserkerEWindup(windup, frame, alpha);
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
  if (windup.skillId === "berserker_r") {
    drawBerserkerRWindup(windup, frame, color, alpha);
    return;
  }
  if (windup.skillId === "fire_mage_q") {
    drawDirectionalWindup(windup, frame, color, alpha, 10);
    return;
  }
  if (windup.skillId === "fire_mage_r") {
    drawFireMageRWindup(windup, frame, color, alpha);
    return;
  }
  if (windup.skillId === "ninja_q") {
    drawNinjaQWindup(windup, frame, color, alpha);
    return;
  }
  if (windup.skillId === "arrow_rain") {
    drawDirectionalWindup(windup, frame, color, alpha, 10);
    return;
  }
  if (windup.skillId === "explorer_e") {
    drawDirectionalWindup(windup, frame, color, alpha, 14);
    return;
  }
  if (windup.skillId === "explorer_q") {
    drawDirectionalWindup(windup, frame, color, alpha, 10);
    return;
  }
  if (windup.skillId === "explorer_r") {
    drawDirectionalWindup(
      { ...windup, range: Math.hypot(state.map.width || 6000, state.map.height || 6000) },
      frame,
      color,
      alpha,
      34,
    );
    return;
  }
  if (windup.skillId === "mage_r") {
    drawMageFinalSparkWindup(windup, frame, alpha);
    return;
  }
}

function drawSwordQWindup(windup, frame, color, alpha) {
  return;
}

function drawBerserkerQWindup(windup, frame, alpha) {
  const config = skillClientConfig.berserker_q || {};
  const origin = castWindupOrigin(windup);
  drawBerserkerQRange(
    origin.x,
    origin.y,
    config.innerRadius || 300,
    config.range || windup.range || 425,
    frame,
    alpha,
  );
}

function drawBerserkerEWindup(windup, frame, alpha) {
  const config = skillClientConfig.berserker_e || {};
  const range = config.range || windup.range || 535;
  const angle = ((config.coneAngleDegrees || 50) * Math.PI) / 180;
  const center = Math.atan2(windup.dirY || 0, windup.dirX || 1);
  const startAngle = center - angle / 2;
  const endAngle = center + angle / 2;
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  const radius = range * frame.scale;
  skillLayer.moveTo(x, y);
  skillLayer.arc(x, y, radius, startAngle, endAngle);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xdc2626, alpha: 0.1 * alpha });
  skillLayer.moveTo(x, y);
  skillLayer.arc(x, y, radius, startAngle, endAngle);
  skillLayer.closePath();
  skillLayer.stroke({ color: 0xf97316, width: 3, alpha: 0.75 * alpha });
}

function drawBerserkerRWindup(windup, frame, color, alpha) {
  const config = skillClientConfig.berserker_r || {};
  drawCircleWindup(windup, frame, color, alpha, config.range || windup.range || 460);
  drawTargetLockWindup(windup, frame, color, alpha);
}

function drawFireMageRWindup(windup, frame, color, alpha) {
  drawTargetLockWindup(windup, frame, color, alpha);
}

function drawNinjaQWindup(windup, frame, color, alpha) {
  drawDirectionalWindup(windup, frame, color, alpha, 10);
  const self = state.players.get(state.playerId);
  if (!self) {
    return;
  }
  const tick = interpolatedTick();
  if ((self.ninja?.shadowReadyTick || 0) <= tick) {
    drawNinjaQShadowWindup(self.ninja?.shadowX, self.ninja?.shadowY, self.ninja?.shadowExpiresAt, windup, frame, color, alpha, tick);
  }
  drawNinjaQShadowWindup(self.ninja?.rShadowX, self.ninja?.rShadowY, self.ninja?.rShadowExpiresAt, windup, frame, color, alpha, tick);
}

function drawNinjaQShadowWindup(x, y, expiresAt, windup, frame, color, alpha, tick) {
  if (!expiresAt || expiresAt <= tick) {
    return;
  }
  const dx = windup.targetX - x;
  const dy = windup.targetY - y;
  const len = Math.hypot(dx, dy) || 1;
  drawDirectionalWindup(
    { ...windup, x, y, dirX: dx / len, dirY: dy / len },
    frame,
    color,
    alpha,
    10,
  );
}

function castWindupOrigin(windup) {
  if (windup.skillId === "berserker_q") {
    const self = state.players.get(state.playerId);
    if (self && !self.dead) {
      return self;
    }
  }
  return windup;
}

function drawDirectionalWindup(windup, frame, color, alpha, width) {
  const range = windup.range || 475;
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  const endX =
    frame.offsetX + (windup.x + (windup.dirX || 1) * range) * frame.scale;
  const endY =
    frame.offsetY + (windup.y + (windup.dirY || 0) * range) * frame.scale;
  skillLayer.moveTo(x, y);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color, width, alpha: 0.16 * alpha });
  skillLayer.moveTo(x, y);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color, width: 2, alpha: 0.72 * alpha });
}

function drawMageFinalSparkWindup(windup, frame, alpha) {
  const range = windup.range || 3400;
  const startX = frame.offsetX + windup.x * frame.scale;
  const startY = frame.offsetY + windup.y * frame.scale;
  const endX =
    frame.offsetX + (windup.x + (windup.dirX || 1) * range) * frame.scale;
  const endY =
    frame.offsetY + (windup.y + (windup.dirY || 0) * range) * frame.scale;
  const width = Math.max(4, 36 * frame.scale);
  skillLayer.moveTo(startX, startY);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0xfef3c7, width, alpha: 0.2 * alpha });
  skillLayer.moveTo(startX, startY);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0xfacc15, width: Math.max(2, width * 0.35), alpha: 0.7 * alpha });
  skillLayer.moveTo(startX, startY);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0xffffff, width: 1, alpha: 0.85 * alpha });
}

function drawCircleWindup(windup, frame, color, alpha, range) {
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  skillLayer.circle(x, y, range * frame.scale);
  skillLayer.fill({ color, alpha: 0.06 * alpha });
  skillLayer.circle(x, y, range * frame.scale);
  skillLayer.stroke({ color, width: 3, alpha: 0.55 * alpha });
}

function drawTargetLockWindup(windup, frame, color, alpha) {
  const x = frame.offsetX + windup.x * frame.scale;
  const y = frame.offsetY + windup.y * frame.scale;
  const tx = frame.offsetX + windup.targetX * frame.scale;
  const ty = frame.offsetY + windup.targetY * frame.scale;
  skillLayer.moveTo(x, y);
  skillLayer.lineTo(tx, ty);
  skillLayer.stroke({ color, width: 3, alpha: 0.55 * alpha });
  skillLayer.circle(tx, ty, 26);
  skillLayer.stroke({ color, width: 3, alpha: 0.8 * alpha });
  skillLayer.circle(tx, ty, 8);
  skillLayer.fill({ color, alpha: 0.18 * alpha });
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
  if (skillId === "explorer_q" || skillId === "explorer_e" || skillId === "explorer_r") {
    return 0x38bdf8;
  }
  if (skillId === "berserker_q" || skillId === "fire_mage_r") {
    return 0xf97316;
  }
  if (skillId === "fire_mage_q" || skillId === "fire_mage_w") {
    return 0xf97316;
  }
  if (skillId === "berserker_e") {
    return 0xdc2626;
  }
  if (skillId === "berserker_r") {
    return 0xef4444;
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
    skillLayer.circle(x, y, preview.range * frame.scale);
    skillLayer.stroke({ color: 0x38bdf8, width: 3, alpha: 0.65 * alpha });
    skillLayer.circle(x, y, 12);
    skillLayer.fill({ color: 0x38bdf8, alpha: 0.18 * alpha });
    return;
  }
  const endX =
    frame.offsetX + (preview.x + preview.dirX * preview.range) * frame.scale;
  const endY =
    frame.offsetY + (preview.y + preview.dirY * preview.range) * frame.scale;
  if (preview.form === "whirlwind") {
    skillLayer.moveTo(x, y);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.55 * alpha });
    skillLayer.circle(endX, endY, (preview.radius || 70) * frame.scale);
    skillLayer.stroke({ color: 0x0284c7, width: 3, alpha: 0.85 * alpha });
    skillLayer.circle(
      endX,
      endY,
      Math.max(6, (preview.radius || 70) * frame.scale * 0.28),
    );
    skillLayer.fill({ color: 0x38bdf8, alpha: 0.16 * alpha });
    return;
  }
  const width = preview.form === "whirlwind" ? 18 : 12;
  skillLayer.moveTo(x, y);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0x38bdf8, width, alpha: 0.28 * alpha });
  skillLayer.moveTo(x, y);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.8 * alpha });
}
