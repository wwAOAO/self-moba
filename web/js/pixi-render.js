function draw(ticker) {
  const frame = calculateFrame();
  drawMap(frame);
  drawEffects(frame);
  syncUnits(frame, ticker.deltaMS);
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
    return;
  }
  if (windup.skillId === "mage_r") {
    drawMageFinalSparkWindup(windup, frame, alpha);
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

function syncUnits(frame, deltaMS) {
  const smoothing = 1 - Math.exp(-deltaMS / 80);

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
    sprite.targetX = unit.x;
    sprite.targetY = unit.y;
    updateUnitBars(sprite, unit);
    updateUnitCollisionCircle(sprite, unit, frame);
    updateStatusLabel(sprite, unit, -(unitModelDisplayRadius(unit) + 30));
    sprite.x += (sprite.targetX - sprite.x) * smoothing;
    sprite.y += (sprite.targetY - sprite.y) * smoothing;
    sprite.node.x = frame.offsetX + sprite.x * frame.scale;
    sprite.node.y = frame.offsetY + sprite.y * frame.scale;
  }
}

function updateUnitCollisionCircle(sprite, unit, frame) {
  if (!sprite.collision) {
    return;
  }
  sprite.collision.clear();
  if (unit.kind !== "enemy_hero") {
    return;
  }
  const radius = (unit.radius || 18) * frame.scale;
  if (radius < 1) {
    return;
  }
  sprite.collision.circle(0, 0, radius);
  sprite.collision.stroke({ color: 0x172026, width: 1, alpha: 0.65 });
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
  drawBar(hpBack, 0x24312b, 1, -34);
  drawHealthBar(hpFill, player, -34);
  drawBar(resourceBack, 0x24312b, 1, -28);
  drawBar(
    resourceFill,
    playerResourceColor(player),
    playerResourceRatio(player),
    -28,
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
  } else if (shape === "mage") {
    drawMageIcon(sprite.body, radius);
  } else if (shape === "gunner") {
    drawGunnerIcon(sprite.body, radius);
  } else if (shape === "ninja") {
    drawNinjaIcon(sprite.body, radius);
  } else if (shape === "blade") {
    drawBladeIcon(sprite.body, radius);
  } else if (shape === "berserker") {
    drawBerserkerIcon(sprite.body, radius);
  } else {
    sprite.body.circle(0, 0, radius);
  }
  sprite.body.fill(player.dead ? 0x6b7280 : colorForTeam(player.team));
  if (shape !== "archer" && shape !== "mage" && shape !== "ninja") {
    sprite.body.stroke({
      color: player.dead
        ? 0x111827
        : shape === "gunner" || shape === "berserker" || shape === "blade"
          ? 0x000000
          : isSelf
            ? 0xffffff
            : 0x172026,
      width: shape === "gunner" || shape === "berserker" || shape === "blade" ? 1 : isSelf ? 2 : 1,
      alpha: player.dead ? 0.45 : 1,
    });
  }
}

function createUnit(unit) {
  const node = new PIXI.Container();
  const body = new PIXI.Graphics();
  const collision = new PIXI.Graphics();
  const hpFill = new PIXI.Graphics();
  let visual = unitVisual(unit.kind);
  if (unit.kind === "fountain" || isMinionKind(unit.kind)) {
    visual = { ...visual, color: colorForTeam(unit.team) };
  }
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
  if (unit.kind !== "fountain") {
    body.stroke({ color: 0xf2f7f3, width: 2 });
  }
  if (unit.kind !== "fountain") {
    drawHealthBar(hpFill, unit, -(modelRadius + 16));
  }
  label.anchor.set(0.5, 0);
  label.y = modelRadius + 6;
  label.visible = unitLabelVisible(unit.kind);
  node.addChild(statusLabel, hpFill, collision, body, label);
  return { node, hpFill, collision, statusLabel, x: unit.x || 0, y: unit.y || 0 };
}

function unitLabelVisible(kind) {
  return (
    kind !== "fountain" &&
    kind !== "tower" &&
    kind !== "crystal" &&
    !isMinionKind(kind)
  );
}

function isMinionKind(kind) {
  return kind === "melee_minion" || kind === "ranged_minion" || kind === "siege_minion" || kind === "super_minion";
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
    statuses.push("击飞");
  }
  if ((target.control?.actionLockedUntilTick || 0) > tick) {
    statuses.push("Lock");
  }
  if ((target.control?.stunnedUntilTick || 0) > tick) {
    statuses.push("眩晕");
  }
  if ((target.control?.silencedUntilTick || 0) > tick) {
    statuses.push("沉默");
  }
  if ((target.control?.rootedUntilTick || 0) > tick) {
    statuses.push("禁锢");
  }
  if ((target.control?.tenacityUntilTick || 0) > tick) {
    statuses.push("韧性");
  }
  if ((target.control?.moveSpeedSlowUntil || 0) > tick) {
    statuses.push("减速");
  }
  if ((target.control?.mageIlluminationUntil || 0) > tick) {
    statuses.push("启明");
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
    return kind !== "dummy" && String(unit.id || "").startsWith("spawn:");
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
    siege_minion: { label: "Cannon", color: 0x6b7280, shape: "cannon_minion" },
    melee_minion: { label: "Melee", color: 0xf97316, shape: "melee_minion" },
    ranged_minion: { label: "Ranged", color: 0xfacc15, shape: "ranged_minion" },
    tower: { label: "Tower", color: 0x475569, shape: "tower" },
    crystal: { label: "Crystal", color: 0xa855f7, shape: "crystal" },
    barracks: { label: "Barracks", color: 0x7c2d12, shape: "rect" },
    fountain: { label: "Fountain", color: 0x38bdf8, shape: "fountain" },
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
  if (visual.shape === "crystal") {
    body.rect(-size * 0.9, size * 0.42, size * 1.8, size * 0.38);
    body.fill(0x475569);
    body.rect(-size * 0.62, size * 0.18, size * 1.24, size * 0.42);
    body.fill(0x64748b);
    body.moveTo(0, -size * 1.05);
    body.lineTo(size * 0.42, -size * 0.22);
    body.lineTo(size * 0.22, size * 0.38);
    body.lineTo(-size * 0.22, size * 0.38);
    body.lineTo(-size * 0.42, -size * 0.22);
    body.closePath();
    body.fill(visual.color);
    body.moveTo(0, -size * 0.78);
    body.lineTo(size * 0.18, -size * 0.18);
    body.lineTo(0, size * 0.18);
    body.lineTo(-size * 0.18, -size * 0.18);
    body.closePath();
    body.fill({ color: 0xf5d0fe, alpha: 0.72 });
    return;
  }
  if (visual.shape === "melee_minion") {
    drawMeleeMinion(body, size, visual.color);
    return;
  }
  if (visual.shape === "ranged_minion") {
    drawRangedMinion(body, size, visual.color);
    return;
  }
  if (visual.shape === "cannon_minion") {
    drawCannonMinion(body, size, visual.color);
    return;
  }
  if (visual.shape === "tower") {
    drawTowerBuilding(body, size, visual.color);
    return;
  }
  if (visual.shape === "fountain") {
    body.circle(0, 0, size);
    body.fill({ color: visual.color, alpha: 0.2 });
    drawFountainCore(body, size, visual.color);
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

function drawTowerBuilding(body, size, color) {
  body.rect(-size * 0.9, size * 0.58, size * 1.8, size * 0.38);
  body.fill(0x334155);
  body.rect(-size * 0.68, size * 0.28, size * 1.36, size * 0.42);
  body.fill(0x475569);

  body.moveTo(-size * 0.48, size * 0.32);
  body.lineTo(-size * 0.34, -size * 0.78);
  body.lineTo(size * 0.34, -size * 0.78);
  body.lineTo(size * 0.48, size * 0.32);
  body.closePath();
  body.fill(visualTowerStone(color));

  body.rect(-size * 0.56, -size * 0.92, size * 1.12, size * 0.3);
  body.fill(0x64748b);
  for (let i = -1; i <= 1; i++) {
    body.rect(i * size * 0.36 - size * 0.1, -size * 1.14, size * 0.2, size * 0.28);
    body.fill(0x475569);
  }

  body.moveTo(0, -size * 1.48);
  body.lineTo(size * 0.34, -size * 0.92);
  body.lineTo(-size * 0.34, -size * 0.92);
  body.closePath();
  body.fill(color);

  body.circle(0, -size * 0.32, size * 0.16);
  body.fill({ color: 0xf8fafc, alpha: 0.78 });
  body.rect(-size * 0.1, size * 0.02, size * 0.2, size * 0.46);
  body.fill({ color: 0x1f2937, alpha: 0.7 });
}

function visualTowerStone(color) {
  if (color === 0xef4444) {
    return 0x7f1d1d;
  }
  if (color === 0x2563eb) {
    return 0x1e3a8a;
  }
  return 0x475569;
}

function drawMeleeMinion(body, size, color) {
  const dark = 0x1f2937;
  body.circle(-size * 0.12, -size * 0.45, size * 0.28);
  body.fill(0x94a3b8);
  body.roundRect(-size * 0.48, -size * 0.22, size * 0.7, size * 0.82, size * 0.12);
  body.fill(color);
  body.moveTo(size * 0.34, -size * 0.38);
  body.lineTo(size * 0.86, -size * 0.12);
  body.lineTo(size * 0.66, size * 0.44);
  body.lineTo(size * 0.2, size * 0.18);
  body.closePath();
  body.fill(0x64748b);
  body.moveTo(-size * 0.52, size * 0.1);
  body.lineTo(-size * 0.92, size * 0.58);
  body.lineTo(-size * 0.7, size * 0.68);
  body.lineTo(-size * 0.32, size * 0.2);
  body.closePath();
  body.fill(dark);
  body.circle(-size * 0.12, -size * 0.5, size * 0.12);
  body.fill(0xe5e7eb);
}

function drawRangedMinion(body, size, color) {
  body.moveTo(0, -size * 0.9);
  body.lineTo(size * 0.5, -size * 0.42);
  body.lineTo(size * 0.62, size * 0.62);
  body.lineTo(-size * 0.62, size * 0.62);
  body.lineTo(-size * 0.5, -size * 0.42);
  body.closePath();
  body.fill(color);
  body.circle(0, -size * 0.48, size * 0.22);
  body.fill(0x111827);
  body.moveTo(-size * 0.52, -size * 0.18);
  body.lineTo(-size * 0.94, size * 0.18);
  body.lineTo(-size * 0.58, size * 0.38);
  body.lineTo(-size * 0.24, size * 0.02);
  body.closePath();
  body.fill(0x8b5cf6);
  body.moveTo(size * 0.35, size * 0.04);
  body.lineTo(size * 0.9, -size * 0.18);
  body.lineTo(size * 0.72, size * 0.3);
  body.lineTo(size * 0.42, size * 0.38);
  body.closePath();
  body.fill({ color: 0xf59e0b, alpha: 0.9 });
}

function drawCannonMinion(body, size, color) {
  const dark = 0x1f2937;
  body.rect(-size * 0.92, -size * 0.46, size * 1.08, size * 0.28);
  body.fill(0x475569);
  body.circle(-size * 0.92, -size * 0.32, size * 0.2);
  body.fill(0x94a3b8);
  body.circle(-size * 0.92, -size * 0.32, size * 0.11);
  body.fill(dark);
  body.moveTo(-size * 0.24, -size * 0.38);
  body.quadraticCurveTo(size * 0.16, -size * 0.92, size * 0.62, -size * 0.36);
  body.lineTo(size * 0.78, size * 0.46);
  body.lineTo(-size * 0.42, size * 0.46);
  body.lineTo(-size * 0.42, -size * 0.08);
  body.lineTo(-size * 0.24, -size * 0.08);
  body.closePath();
  body.fill(color);
  body.rect(-size * 0.64, -size * 0.02, size * 0.28, size * 0.76);
  body.fill(0x475569);
  body.circle(-size * 0.28, size * 0.62, size * 0.27);
  body.fill(dark);
  body.circle(size * 0.48, size * 0.62, size * 0.27);
  body.fill(dark);
  body.circle(-size * 0.28, size * 0.62, size * 0.13);
  body.fill(0xe5e7eb);
  body.circle(size * 0.48, size * 0.62, size * 0.13);
  body.fill(0xe5e7eb);
}

function drawFountainCore(body, size, color) {
  const s = size / 90;
  body.circle(0, 0, 62 * s);
  body.fill({ color: 0x1f2937, alpha: 0.94 });
  body.circle(0, 0, 52 * s);
  body.fill({ color: 0x8b6a3a, alpha: 0.96 });
  body.circle(0, 0, 40 * s);
  body.fill({ color: 0x374151, alpha: 0.96 });
  body.circle(0, 0, 30 * s);
  body.fill({ color, alpha: 0.34 });

  for (let i = 0; i < 8; i++) {
    const angle = (Math.PI * 2 * i) / 8;
    const cx = Math.cos(angle) * 57 * s;
    const cy = Math.sin(angle) * 57 * s;
    body.circle(cx, cy, 8 * s);
    body.fill({ color: 0xa17842, alpha: 0.98 });
  }

  body.circle(0, 0, 24 * s);
  body.fill({ color: 0x111827, alpha: 0.98 });
  body.circle(0, 0, 18 * s);
  body.fill({ color: 0xa17842, alpha: 0.98 });
  body.circle(0, 0, 12 * s);
  body.fill({ color, alpha: 0.9 });

  const guards = [
    [-45, -14, -28, -35, -15, -27, -31, -7],
    [15, -27, 28, -35, 45, -14, 31, -7],
    [-45, 14, -31, 7, -15, 27, -28, 35],
    [15, 27, 31, 7, 45, 14, 28, 35],
  ];
  for (const g of guards) {
    body.moveTo(g[0] * s, g[1] * s);
    body.lineTo(g[2] * s, g[3] * s);
    body.lineTo(g[4] * s, g[5] * s);
    body.lineTo(g[6] * s, g[7] * s);
    body.closePath();
    body.fill({ color: 0x475569, alpha: 0.95 });
  }

  body.circle(0, -18 * s, 34 * s);
  body.stroke({ color, width: Math.max(1, 2 * s), alpha: 0.45 });
  body.rect(-8 * s, -78 * s, 16 * s, 48 * s);
  body.fill({ color, alpha: 0.13 });
}

function spawnDamageText(target, damage, damageType) {
  if (!damage) {
    return;
  }
  const sameFrameOffset = state.damageTexts.filter(
    (effect) => effect.targetId === (target.id || target.playerId),
  ).length;
  const text = new PIXI.Text({
    text: `-${damage}`,
    style: {
      fill: damageTextColor(damageType),
      fontFamily: "Arial",
      fontSize: 15,
      fontWeight: "900",
      stroke: { color: damageTextStrokeColor(damageType), width: 2 },
    },
  });
  text.anchor.set(0.5, 0.5);
  effectLayer.addChild(text);
  state.damageTexts.push({
    node: text,
    targetId: target.id || target.playerId,
    x: target.x,
    y: target.y - 42 - sameFrameOffset * 100,
    age: 0,
    lifetime: 720,
  });
}

function damageTextColor(damageType) {
  if (damageType === "magic") {
    return 0x8b5cf6;
  }
  if (damageType === "true") {
    return 0xe5e7eb;
  }
  return 0xff3333;
}

function damageTextStrokeColor(damageType) {
  if (damageType === "true") {
    return 0x111827;
  }
  return 0xffffff;
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
  drawHealthBar(sprite.hpFill, target, -34);
  if (sprite.resourceFill) {
    drawBar(
      sprite.resourceFill,
      playerResourceColor(target),
      playerResourceRatio(target),
      -28,
    );
  }
}

function updateUnitBars(sprite, unit) {
  if (unit.kind === "fountain") {
    sprite.hpFill.clear();
    return;
  }
  drawHealthBar(sprite.hpFill, unit, -(unitModelDisplayRadius(unit) + 16));
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
  if ((player?.heroId || els.heroId.value) === "blade") {
    return 0xef4444;
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
  const width = 36 * value;
  graphics.roundRect(-18, y, width, 4, 1);
  graphics.fill(color);
  graphics.roundRect(-18, y, 36, 4, 1);
  graphics.stroke({ color: 0x172026, width: 1, alpha: 0.85 });
}

function drawHealthBar(graphics, entity, y) {
  const stats = entity?.stats || entity || {};
  const maxHp = stats.maxHp || 0;
  const hp = Math.max(0, stats.hp || 0);
  const shield = shieldValue(entity);
  const total = hp + shield;
  const scale = maxHp > 0 ? 36 / Math.max(maxHp, total) : 0;
  const hpWidth = Math.min(36, hp * scale);
  const shieldWidth = Math.min(36 - hpWidth, shield * scale);
  graphics.clear();
  if (hpWidth > 0) {
    graphics.rect(-18, y, hpWidth, 4);
    graphics.fill(0x22c55e);
  }
  if (shieldWidth > 0) {
    graphics.rect(-18 + hpWidth, y, shieldWidth, 4);
    graphics.fill(0xf8fafc);
  }
  graphics.roundRect(-18, y, 36, 4, 1);
  graphics.stroke({ color: 0x172026, width: 1, alpha: 0.85 });
}

