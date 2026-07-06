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
  for (const buff of target.buffs || []) {
    if (!buff.negative || !buff.expiresAtTick || buff.expiresAtTick <= tick) {
      continue;
    }
    const remain = ((buff.expiresAtTick - tick) / state.tickRate).toFixed(1);
    statuses.push(`${buff.name || buff.id} ${remain}s`);
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
