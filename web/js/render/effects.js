function drawEffects(frame) {
  skillLayer.clear();
  drawActiveSkillRanges(frame);
  drawSwordETargetCooldowns(frame);
  drawCastWindups(frame);
  drawSkillPreview(frame);

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
    if (effect.kind === "mage_light_binding") {
      drawMageLightBindingEffect(effect, frame);
      continue;
    }
    if (effect.kind === "mage_prismatic_barrier") {
      drawMagePrismaticBarrierEffect(effect, frame);
      continue;
    }
    if (effect.kind === "mage_lucent_singularity_orb") {
      drawMageLucentSingularityOrbEffect(effect, frame);
      continue;
    }
    if (effect.kind === "mage_lucent_singularity") {
      drawMageLucentSingularityEffect(effect, frame);
      continue;
    }
    if (effect.kind === "mage_final_spark") {
      drawMageFinalSparkEffect(effect, frame);
      continue;
    }
    if (effect.kind === "fountain_shot") {
      drawFountainShotEffect(effect, frame);
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
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x67e8f9, width: 10, alpha: 0.45 });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x0e7490, width: 2, alpha: 0.9 });
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
  skillLayer.circle(x, y, radius * frame.scale);
  skillLayer.fill({ color: 0xf59e0b, alpha: 0.08 });
  skillLayer.circle(x, y, radius * frame.scale);
  skillLayer.stroke({ color: 0xf97316, width: 3, alpha: 0.75 });
  skillLayer.circle(x, y, hitRadius * frame.scale);
  skillLayer.stroke({ color: 0xf97316, width: 1, alpha: 0.35 });
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
  skillLayer.moveTo(x, y);
  skillLayer.arc(x, y, radius, startAngle, endAngle);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xfacc15, alpha: 0.14 * alpha });
  skillLayer.moveTo(x, y);
  skillLayer.arc(x, y, radius, startAngle, endAngle);
  skillLayer.closePath();
  skillLayer.stroke({ color: 0xd97706, width: 2, alpha: 0.8 * alpha });
}

function drawTankImpactEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 250) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x94a3b8, alpha: 0.14 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x475569, width: 3, alpha: 0.85 * alpha });
  skillLayer.circle(x, y, Math.max(8, radius * 0.12));
  skillLayer.fill({ color: 0xe2e8f0, alpha: 0.22 * alpha });
}

function drawBasicArrowEffect(effect, frame) {
  if (effect.sourceHeroId === "mage") {
    drawMageBasicStarEffect(effect, frame);
    return;
  }
  if (!effect.sourceHeroId) {
    drawMinionBasicProjectile(effect, frame);
    return;
  }
  if ((effect.count || 1) >= 3) {
    drawTripleArrowProjectile(effect, frame, 0xf8d36a, 0xf59e0b, {
      fromSnapshot: true,
    });
    return;
  }
  drawArrowProjectile(effect, frame, 0xf8d36a, 0xf59e0b, {
    fromSnapshot: true,
  });
}

function drawMinionBasicProjectile(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(2, (effect.radius || 10) * frame.scale * 0.325);
  const color = colorForTeam(effect.team);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color, alpha: 0.88 });
  skillLayer.circle(x, y, radius + 2);
  skillLayer.stroke({ color, width: 2, alpha: 0.55 });
}

function drawMageBasicStarEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  drawStarPath(skillLayer, x, y, 10, 4.5);
  skillLayer.fill({ color: 0xfacc15, alpha: 0.95 });
}

function drawStarPath(graphics, x, y, outer, inner) {
  graphics.moveTo(x, y - outer);
  for (let i = 1; i < 10; i++) {
    const angle = -Math.PI / 2 + (Math.PI * i) / 5;
    const radius = i % 2 ? inner : outer;
    graphics.lineTo(x + Math.cos(angle) * radius, y + Math.sin(angle) * radius);
  }
  graphics.closePath();
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
  drawProjectileSweepArea(effect, frame, position, radius, 0xa78bfa, 0x8b5cf6);
  drawArrowProjectile(effect, frame, 0xc4b5fd, 0x7c3aed, {
    fromSnapshot: true,
  });
}

function drawMageLightBindingEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(8, (effect.radius || 45) * frame.scale);
  const alpha = 0.9;
  skillLayer.circle(x, y, radius * 0.45);
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.9 });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xfacc15, width: 3, alpha: 0.72 * alpha });
  skillLayer.moveTo(
    x - (effect.dirX || 1) * radius * 0.9,
    y - (effect.dirY || 0) * radius * 0.9,
  );
  skillLayer.lineTo(
    x + (effect.dirX || 1) * radius * 1.2,
    y + (effect.dirY || 0) * radius * 1.2,
  );
  skillLayer.stroke({ color: 0xfbbf24, width: 4, alpha: 0.8 });
}

function drawMagePrismaticBarrierEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const length = 32;
  skillLayer.moveTo(
    x - Math.cos(angle) * length * 0.5,
    y - Math.sin(angle) * length * 0.5,
  );
  skillLayer.lineTo(
    x + Math.cos(angle) * length * 0.5,
    y + Math.sin(angle) * length * 0.5,
  );
  skillLayer.stroke({ color: 0xfacc15, width: 5, alpha: 0.9 });
  skillLayer.circle(x + Math.cos(angle) * length * 0.55, y + Math.sin(angle) * length * 0.55, 7);
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.95 });
  skillLayer.circle(x, y, Math.max(10, (effect.radius || 55) * frame.scale * 0.35));
  skillLayer.stroke({ color: 0xfbbf24, width: 2, alpha: 0.55 });
}

function drawMageLucentSingularityOrbEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(10, (effect.radius || 34) * frame.scale);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0xfef08a, alpha: 0.42 });
  skillLayer.circle(x, y, radius * 0.45);
  skillLayer.fill({ color: 0xffffff, alpha: 0.82 });
}

function drawMageLucentSingularityEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 300) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0xfef08a, alpha: 0.12 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xfacc15, width: 3, alpha: 0.8 * alpha });
  skillLayer.circle(x, y, Math.max(8, radius * 0.08));
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.35 * alpha });
}

function drawMageFinalSparkEffect(effect, frame) {
  const startX = frame.offsetX + effect.x * frame.scale;
  const startY = frame.offsetY + effect.y * frame.scale;
  const endX = frame.offsetX + (effect.endX || effect.x) * frame.scale;
  const endY = frame.offsetY + (effect.endY || effect.y) * frame.scale;
  const width = Math.max(8, (effect.width || 200) * frame.scale);
  const alpha = effectAlpha(effect);
  skillLayer.moveTo(startX, startY);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0xfef3c7, width, alpha: 0.22 * alpha });
  skillLayer.moveTo(startX, startY);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0xfacc15, width: Math.max(3, width * 0.28), alpha: 0.75 * alpha });
  skillLayer.moveTo(startX, startY);
  skillLayer.lineTo(endX, endY);
  skillLayer.stroke({ color: 0xffffff, width: Math.max(2, width * 0.08), alpha: 0.9 * alpha });
}

function drawFountainShotEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(5, (effect.radius || 18) * frame.scale * 0.45);
  const tailX = x - (effect.dirX || 0) * radius * 3.2;
  const tailY = y - (effect.dirY || 0) * radius * 3.2;
  skillLayer.moveTo(tailX, tailY);
  skillLayer.lineTo(x, y);
  skillLayer.stroke({ color: 0x7dd3fc, width: Math.max(2, radius * 0.75), alpha: 0.65 });
  skillLayer.circle(x, y, radius * 1.8);
  skillLayer.fill({ color: 0xbfdbfe, alpha: 0.22 });
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x7dd3fc, alpha: 0.95 });
}

function drawProjectileSweepArea(effect, frame, position, radius, fillColor, strokeColor) {
  const startX = frame.offsetX + (effect.x || 0) * frame.scale;
  const startY = frame.offsetY + (effect.y || 0) * frame.scale;
  const endX = frame.offsetX + position.x * frame.scale;
  const endY = frame.offsetY + position.y * frame.scale;
  const dx = endX - startX;
  const dy = endY - startY;
  const length = Math.hypot(dx, dy);
  if (length > 0.5) {
    const nx = -dy / length;
    const ny = dx / length;
    skillLayer
      .moveTo(startX + nx * radius, startY + ny * radius)
      .lineTo(endX + nx * radius, endY + ny * radius)
      .lineTo(endX - nx * radius, endY - ny * radius)
      .lineTo(startX - nx * radius, startY - ny * radius)
      .closePath();
    skillLayer.fill({ color: fillColor, alpha: 0.06 });
  }
  skillLayer.circle(endX, endY, radius);
  skillLayer.fill({ color: fillColor, alpha: 0.08 });
  if (length > 0.5) {
    skillLayer
      .moveTo(startX, startY)
      .lineTo(endX, endY)
      .stroke({ color: strokeColor, width: Math.max(2, radius * 2), alpha: 0.16 });
  }
  skillLayer.circle(endX, endY, radius);
  skillLayer.stroke({ color: strokeColor, width: 2, alpha: 0.7 });
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
    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0x38bdf8, alpha: 0.08 * alpha });
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.7 * alpha });
  }
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const size = arrived ? 10 : 14;
  skillLayer
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
  skillLayer.fill({ color: 0x0ea5e9, alpha: arrived ? 0.8 : 0.95 });
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
  skillLayer
    .moveTo(
      x + Math.cos(angle) * length * 0.5,
      y + Math.sin(angle) * length * 0.5,
    )
    .lineTo(
      x - Math.cos(angle) * length * 0.5,
      y - Math.sin(angle) * length * 0.5,
    );
  skillLayer.stroke({ color: shaftColor, width: 3, alpha: 0.95 });
  skillLayer
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
  skillLayer.fill({ color: headColor, alpha: 0.95 });
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
  return (target.radius || 18) * frame.scale;
}

function drawTripleArrowProjectile(effect, frame, shaftColor, headColor, options = {}) {
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
      options,
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
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0x7dd3fc, width: 4, alpha: 0.25 });
    skillLayer.moveTo(startX, startY);
    skillLayer.arc(x, y, radius, startAngle, endAngle);
    skillLayer.stroke({ color: 0x38bdf8, width: 4, alpha: 0.9 });
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
  skillLayer.circle(sx, sy, radius);
  skillLayer.stroke({ color: 0x0284c7, width: 3, alpha: 0.9 });
  skillLayer.circle(sx, sy, Math.max(6, radius * 0.35));
  skillLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.8 });
  skillLayer.moveTo(sx - radius * 0.55, sy);
  skillLayer.lineTo(sx + radius * 0.55, sy);
  skillLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.45 });
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
  skillLayer.circle(sx, sy, radius * 0.65);
  skillLayer.fill({ color: 0x8b5e34, alpha: 0.85 });
  skillLayer.circle(sx, sy, radius);
  skillLayer.stroke({ color: 0x5c4033, width: 2, alpha: 0.75 });
}

