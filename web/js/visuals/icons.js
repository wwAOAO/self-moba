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
  if (player.heroId === "mage") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "gunner") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "blade") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "berserker") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "ninja") {
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
  if (player.heroId === "mage") {
    return "mage";
  }
  if (player.heroId === "gunner") {
    return "gunner";
  }
  if (player.heroId === "blade") {
    return "blade";
  }
  if (player.heroId === "berserker") {
    return "berserker";
  }
  if (player.heroId === "ninja") {
    return "ninja";
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

function drawGunnerIcon(graphics, radius) {
  graphics.roundRect(
    -radius * 1.5,
    -radius * 0.95,
    radius * 3.15,
    radius * 0.7,
    radius * 0.08,
  );

  graphics.rect(-radius * 1.42, -radius * 0.28, radius * 2.95, radius * 0.26);

  graphics.moveTo(-radius * 1.12, -radius * 1.06);
  graphics.lineTo(-radius * 1.02, -radius * 1.24);
  graphics.lineTo(-radius * 0.88, -radius * 1.24);
  graphics.lineTo(-radius * 0.82, -radius * 1.06);
  graphics.closePath();

  graphics.moveTo(radius * 1.42, -radius * 1.08);
  graphics.lineTo(radius * 1.6, -radius * 1.08);
  graphics.lineTo(radius * 1.6, -radius * 0.18);
  graphics.lineTo(radius * 1.42, -radius * 0.18);
  graphics.closePath();

  graphics.moveTo(-radius * 0.52, -radius * 0.02);
  graphics.lineTo(radius * 0.42, -radius * 0.02);
  graphics.quadraticCurveTo(radius * 0.22, radius * 0.58, -radius * 0.38, radius * 0.58);
  graphics.lineTo(-radius * 0.66, radius * 0.58);
  graphics.quadraticCurveTo(-radius * 0.98, radius * 0.32, -radius * 0.86, radius * 0.02);
  graphics.closePath();

  graphics.moveTo(-radius * 0.68, -radius * 0.02);
  graphics.lineTo(-radius * 0.18, -radius * 0.02);
  graphics.lineTo(-radius * 0.56, radius * 1.48);
  graphics.lineTo(-radius * 1.34, radius * 1.48);
  graphics.lineTo(-radius * 1.16, radius * 0.74);
  graphics.quadraticCurveTo(-radius * 1.02, radius * 0.34, -radius * 1.28, radius * 0.12);
  graphics.lineTo(-radius * 1.44, -radius * 0.02);
  graphics.closePath();

  graphics.moveTo(-radius * 0.58, radius * 0.1);
  graphics.quadraticCurveTo(-radius * 0.2, radius * 0.18, -radius * 0.36, radius * 0.46);
  graphics.quadraticCurveTo(-radius * 0.64, radius * 0.38, -radius * 0.58, radius * 0.1);
  graphics.closePath();

  graphics.moveTo(-radius * 0.18, -radius * 1.02);
  graphics.lineTo(-radius * 0.02, -radius * 0.82);
  graphics.lineTo(radius * 0.54, -radius * 0.82);
  graphics.lineTo(radius * 0.7, -radius * 1.02);
  graphics.closePath();
}

function drawNinjaIcon(graphics, radius) {
  const iconRadius = radius * 0.9;
  drawShurikenBlade(graphics, iconRadius, -Math.PI / 2);
  drawShurikenBlade(graphics, iconRadius, 0);
  drawShurikenBlade(graphics, iconRadius, Math.PI / 2);
  drawShurikenBlade(graphics, iconRadius, Math.PI);
  graphics.circle(0, 0, iconRadius * 0.3);
}

function drawShurikenBlade(graphics, radius, angle) {
  const tip = pointAt(angle, radius * 1.65);
  const leftWing = pointAt(angle - 0.55, radius * 0.88);
  const leftInner = pointAt(angle - 1.15, radius * 0.42);
  const rightInner = pointAt(angle + 1.15, radius * 0.42);
  const rightWing = pointAt(angle + 0.55, radius * 0.88);
  graphics.moveTo(tip.x, tip.y);
  graphics.lineTo(leftWing.x, leftWing.y);
  graphics.quadraticCurveTo(
    leftInner.x,
    leftInner.y,
    pointAt(angle, radius * 0.5).x,
    pointAt(angle, radius * 0.5).y,
  );
  graphics.quadraticCurveTo(rightInner.x, rightInner.y, rightWing.x, rightWing.y);
  graphics.closePath();
}

function pointAt(angle, length) {
  return {
    x: Math.cos(angle) * length,
    y: Math.sin(angle) * length,
  };
}

function drawBladeIcon(graphics, radius) {
  const scale = radius * 1.2;
  const paths = [
    [
      [0.896, -1],
      [0.857, -0.826],
      [0.796, -0.664],
      [0.465, -0.193],
      [0.224, 0.098],
      [-0.034, 0.361],
      [-0.347, 0.613],
      [-0.291, 0.765],
      [-0.05, 0.546],
      [0.101, 0.395],
      [0.353, 0.143],
      [0.633, -0.249],
      [0.891, -0.664],
      [0.963, -0.854],
      [1, -1],
      [0.963, -0.955],
      [0.935, -0.955],
      [0.913, -0.983],
    ],
    [
      [0.146, 0.249],
      [0.448, 0.507],
      [0.538, 0.462],
      [0.616, 0.445],
      [0.599, 0.524],
      [0.549, 0.613],
      [0.487, 0.681],
      [0.403, 0.737],
      [0.314, 0.765],
      [0.101, 0.574],
      [-0.028, 0.439],
    ],
    [
      [-0.616, 0.647],
      [-0.532, 0.737],
      [-0.476, 0.77],
      [-0.751, 0.983],
      [-0.796, 1],
      [-0.868, 0.983],
      [-0.902, 0.938],
      [-0.891, 0.86],
    ],
    [
      [0.611, 0.647],
      [0.891, 0.86],
      [0.902, 0.938],
      [0.868, 0.983],
      [0.824, 1],
      [0.751, 0.983],
      [0.476, 0.77],
      [0.538, 0.731],
    ],
    [
      [-0.902, -1],
      [-0.667, -0.737],
      [-0.291, -0.216],
      [-0.039, 0.059],
      [-0.14, 0.176],
      [-0.426, -0.143],
      [-0.796, -0.664],
    ],
  ];
  for (const path of paths) {
    for (let i = 0; i < path.length; i++) {
      const x = path[i][0] * scale;
      const y = path[i][1] * scale;
      if (i === 0) {
        graphics.moveTo(x, y);
      } else {
        graphics.lineTo(x, y);
      }
    }
    graphics.closePath();
  }
}

function drawBerserkerIcon(graphics, radius) {
  graphics.rect(-radius * 0.12, -radius * 1.18, radius * 0.24, radius * 3.05);

  graphics.moveTo(-radius * 0.12, -radius * 1.28);
  graphics.lineTo(radius * 0.78, -radius * 1.42);
  graphics.quadraticCurveTo(
    radius * 1.28,
    -radius * 0.9,
    radius * 1.16,
    -radius * 0.12,
  );
  graphics.lineTo(radius * 0.52, -radius * 0.22);
  graphics.lineTo(radius * 0.2, -radius * 0.54);
  graphics.lineTo(-radius * 0.12, -radius * 0.54);
  graphics.closePath();

  graphics.moveTo(-radius * 0.12, -radius * 1.26);
  graphics.lineTo(-radius * 0.62, -radius * 1.08);
  graphics.lineTo(-radius * 0.48, -radius * 0.54);
  graphics.lineTo(-radius * 0.12, -radius * 0.54);
  graphics.closePath();

  graphics.rect(-radius * 0.46, -radius * 0.56, radius * 0.92, radius * 0.22);
  graphics.moveTo(-radius * 0.34, radius * 1.68);
  graphics.lineTo(radius * 0.34, radius * 1.68);
  graphics.lineTo(radius * 0.22, radius * 2.04);
  graphics.lineTo(-radius * 0.22, radius * 2.04);
  graphics.closePath();
}

function drawMageIcon(graphics, radius) {
  graphics.circle(0, -radius * 0.38, radius * 0.48);
  graphics.circle(0, -radius * 0.38, radius * 0.24);

  graphics.rect(-radius * 0.13, -radius * 0.2, radius * 0.26, radius * 1.45);
  graphics.moveTo(-radius * 0.26, -radius * 0.02);
  graphics.lineTo(radius * 0.26, -radius * 0.02);
  graphics.lineTo(radius * 0.2, radius * 0.18);
  graphics.lineTo(-radius * 0.2, radius * 0.18);
  graphics.closePath();
  graphics.moveTo(0, radius * 1.35);
  graphics.lineTo(radius * 0.16, radius * 1.12);
  graphics.lineTo(-radius * 0.16, radius * 1.12);
  graphics.closePath();

  drawMageWing(graphics, radius, -1);
  drawMageWing(graphics, radius, 1);
  drawMageStar(graphics, 0, -radius * 1.34, radius * 0.25);
  drawMageStar(graphics, -radius * 0.48, -radius * 1.05, radius * 0.14);
  drawMageStar(graphics, radius * 0.5, -radius * 1.04, radius * 0.13);
}

function drawMageWing(graphics, radius, side) {
  graphics.moveTo(side * radius * 0.34, -radius * 0.44);
  graphics.quadraticCurveTo(side * radius * 0.96, -radius * 0.94, side * radius * 1.45, -radius * 1.02);
  graphics.quadraticCurveTo(side * radius * 1.18, -radius * 0.66, side * radius * 0.74, -radius * 0.44);
  graphics.quadraticCurveTo(side * radius * 1.2, -radius * 0.36, side * radius * 1.38, -radius * 0.12);
  graphics.quadraticCurveTo(side * radius * 0.9, -radius * 0.12, side * radius * 0.52, -radius * 0.28);
  graphics.quadraticCurveTo(side * radius * 0.88, -radius * 0.02, side * radius * 0.98, radius * 0.24);
  graphics.quadraticCurveTo(side * radius * 0.52, radius * 0.08, side * radius * 0.3, -radius * 0.2);
  graphics.closePath();
}

function drawMageStar(graphics, x, y, size) {
  const points = 5;
  graphics.moveTo(x, y - size);
  for (let i = 1; i < points * 2; i++) {
    const angle = -Math.PI / 2 + (Math.PI * i) / points;
    const r = i % 2 ? size * 0.42 : size;
    graphics.lineTo(x + Math.cos(angle) * r, y + Math.sin(angle) * r);
  }
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
  if (isMinionKind(unit?.kind)) {
    return Math.max(8, (unit.radius || 0) * 0.5);
  }
  return Math.max(18, unit.radius || 0);
}

function unitCollisionRadius(unit) {
  return unitModelRadius(unit);
}

function unitHitRadius(unit) {
  return unitModelDisplayRadius(unit) + 8;
}

