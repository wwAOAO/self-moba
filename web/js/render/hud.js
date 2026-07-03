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

