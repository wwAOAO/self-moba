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
  if (player.heroId === "mage" || player.heroId === "fire_mage") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "gunner") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "blade") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "killer") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "berserker") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "ninja") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "explorer") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "robot") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "doctor") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "monk") {
    return player.playerId === state.playerId ? 13 : 11;
  }
  if (player.heroId === "butcher") {
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
  if (player.heroId === "fire_mage") {
    return "fire";
  }
  if (player.heroId === "frostmage") {
    return "snowflake";
  }
  if (player.heroId === "gunner") {
    return "gunner";
  }
  if (player.heroId === "blade") {
    return "blade";
  }
  if (player.heroId === "killer") {
    return "killer";
  }
  if (player.heroId === "berserker") {
    return "berserker";
  }
  if (player.heroId === "ninja") {
    return "ninja";
  }
  if (player.heroId === "explorer") {
    return "explorer";
  }
  if (player.heroId === "robot") {
    return "robot";
  }
  if (player.heroId === "doctor") {
    return "doctor";
  }
  if (player.heroId === "monk") {
    return "monk";
  }
  if (player.heroId === "butcher") {
    return "butcher";
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

function drawKillerIcon(graphics, radius) {
  drawKillerDagger(graphics, radius, -1);
  drawKillerDagger(graphics, radius, 1);

  graphics.moveTo(0, -radius * 0.34);
  graphics.lineTo(radius * 0.34, 0);
  graphics.lineTo(0, radius * 0.34);
  graphics.lineTo(-radius * 0.34, 0);
  graphics.closePath();
}

function drawKillerDagger(graphics, radius, side) {
  const mirrorX = (value) => value * side;

  graphics.moveTo(mirrorX(radius * 1.34), -radius * 1.16);
  graphics.quadraticCurveTo(
    mirrorX(radius * 0.72),
    -radius * 0.84,
    mirrorX(radius * 0.2),
    -radius * 0.12,
  );
  graphics.lineTo(mirrorX(radius * 0.48), radius * 0.12);
  graphics.quadraticCurveTo(
    mirrorX(radius * 0.94),
    -radius * 0.46,
    mirrorX(radius * 1.34),
    -radius * 1.16,
  );
  graphics.closePath();

  graphics.moveTo(mirrorX(radius * 0.1), -radius * 0.18);
  graphics.lineTo(mirrorX(radius * 0.6), radius * 0.28);
  graphics.lineTo(mirrorX(radius * 0.46), radius * 0.44);
  graphics.lineTo(mirrorX(radius * 0.18), radius * 0.22);
  graphics.lineTo(mirrorX(radius * 0.42), radius * 0.52);
  graphics.lineTo(mirrorX(-radius * 0.48), radius * 1.42);
  graphics.lineTo(mirrorX(-radius * 0.72), radius * 1.18);
  graphics.lineTo(mirrorX(radius * 0.18), radius * 0.28);
  graphics.lineTo(mirrorX(-radius * 0.08), radius * 0.02);
  graphics.closePath();

  graphics.circle(mirrorX(-radius * 0.66), radius * 1.36, radius * 0.2);
}

function drawExplorerHatIcon(graphics, radius) {
  graphics.ellipse(0, radius * 0.38, radius * 1.48, radius * 0.34);
  graphics.moveTo(-radius * 0.88, radius * 0.28);
  graphics.quadraticCurveTo(-radius * 0.64, -radius * 0.74, 0, -radius * 0.86);
  graphics.quadraticCurveTo(radius * 0.64, -radius * 0.74, radius * 0.88, radius * 0.28);
  graphics.closePath();
  graphics.rect(-radius * 0.78, -radius * 0.02, radius * 1.56, radius * 0.26);
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
  graphics.moveTo(-radius * 1.42, radius * 0.08);
  graphics.quadraticCurveTo(-radius * 0.82, -radius * 0.12, 0, -radius * 0.16);
  graphics.quadraticCurveTo(radius * 0.82, -radius * 0.12, radius * 1.42, radius * 0.08);
  graphics.quadraticCurveTo(radius * 0.78, radius * 0.42, 0, radius * 0.46);
  graphics.quadraticCurveTo(-radius * 0.78, radius * 0.42, -radius * 1.42, radius * 0.08);
  graphics.closePath();

  graphics.moveTo(-radius * 0.82, -radius * 0.04);
  graphics.quadraticCurveTo(-radius * 0.42, -radius * 0.78, 0, -radius * 1.3);
  graphics.quadraticCurveTo(radius * 0.42, -radius * 0.78, radius * 0.82, -radius * 0.04);
  graphics.quadraticCurveTo(radius * 0.36, radius * 0.12, 0, radius * 0.14);
  graphics.quadraticCurveTo(-radius * 0.36, radius * 0.12, -radius * 0.82, -radius * 0.04);
  graphics.closePath();

  graphics.rect(-radius * 0.72, -radius * 0.08, radius * 1.44, radius * 0.18);

  graphics.moveTo(-radius * 0.58, radius * 0.32);
  graphics.quadraticCurveTo(-radius * 0.5, radius * 0.88, -radius * 0.72, radius * 1.36);
  graphics.lineTo(-radius * 0.5, radius * 1.45);
  graphics.quadraticCurveTo(-radius * 0.24, radius * 0.84, -radius * 0.36, radius * 0.3);
  graphics.closePath();

  graphics.moveTo(radius * 0.58, radius * 0.32);
  graphics.quadraticCurveTo(radius * 0.5, radius * 0.88, radius * 0.72, radius * 1.36);
  graphics.lineTo(radius * 0.5, radius * 1.45);
  graphics.quadraticCurveTo(radius * 0.24, radius * 0.84, radius * 0.36, radius * 0.3);
  graphics.closePath();

  graphics.moveTo(-radius * 0.16, radius * 0.4);
  graphics.lineTo(0, radius * 0.72);
  graphics.lineTo(radius * 0.16, radius * 0.4);
  graphics.lineTo(0, radius * 1.12);
  graphics.closePath();
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

function drawRobotIcon(graphics, radius) {
  graphics.roundRect(
    -radius * 0.72,
    -radius * 0.62,
    radius * 1.44,
    radius * 1.18,
    radius * 0.16,
  );
  graphics.rect(-radius * 0.5, radius * 0.56, radius * 1, radius * 0.42);
  graphics.rect(-radius * 0.14, -radius * 1.04, radius * 0.28, radius * 0.42);
  graphics.circle(0, -radius * 1.14, radius * 0.16);
  graphics.circle(-radius * 0.36, -radius * 0.2, radius * 0.16);
  graphics.circle(radius * 0.36, -radius * 0.2, radius * 0.16);
  graphics.rect(-radius * 0.32, radius * 0.2, radius * 0.64, radius * 0.12);
  graphics.rect(-radius * 1.12, -radius * 0.22, radius * 0.36, radius * 0.72);
  graphics.rect(radius * 0.76, -radius * 0.22, radius * 0.36, radius * 0.72);
}

function drawDoctorIcon(graphics, radius) {
  graphics.roundRect(
    -radius * 0.58,
    -radius * 1.12,
    radius * 1.16,
    radius * 2.24,
    radius * 0.26,
  );
  graphics.rect(-radius * 0.34, -radius * 1.38, radius * 0.68, radius * 0.34);
  graphics.roundRect(
    -radius * 0.76,
    -radius * 0.42,
    radius * 1.52,
    radius * 0.86,
    radius * 0.12,
  );
  graphics.rect(-radius * 0.14, -radius * 0.76, radius * 0.28, radius * 1.52);
  graphics.rect(-radius * 0.58, -radius * 0.14, radius * 1.16, radius * 0.28);
  graphics.circle(radius * 0.42, radius * 0.74, radius * 0.14);
  graphics.circle(-radius * 0.34, radius * 0.78, radius * 0.1);
}

function drawMonkIcon(graphics, radius) {
  graphics.circle(0, -radius * 0.58, radius * 0.48);
  graphics.circle(-radius * 0.46, -radius * 0.54, radius * 0.12);
  graphics.circle(radius * 0.46, -radius * 0.54, radius * 0.12);

  graphics.rect(-radius * 0.54, -radius * 0.72, radius * 1.08, radius * 0.22);
  graphics.moveTo(radius * 0.5, -radius * 0.68);
  graphics.lineTo(radius * 1.28, -radius * 0.42);
  graphics.lineTo(radius * 0.92, -radius * 0.14);
  graphics.lineTo(radius * 0.48, -radius * 0.48);
  graphics.closePath();

  graphics.moveTo(-radius * 1.12, radius * 0.98);
  graphics.quadraticCurveTo(-radius * 0.98, radius * 0.1, -radius * 0.42, -radius * 0.02);
  graphics.lineTo(radius * 0.42, -radius * 0.02);
  graphics.quadraticCurveTo(radius * 0.98, radius * 0.1, radius * 1.12, radius * 0.98);
  graphics.quadraticCurveTo(radius * 0.56, radius * 1.34, 0, radius * 1.36);
  graphics.quadraticCurveTo(-radius * 0.56, radius * 1.34, -radius * 1.12, radius * 0.98);
  graphics.closePath();

  graphics.moveTo(0, radius * 0.02);
  graphics.lineTo(radius * 0.24, radius * 0.28);
  graphics.lineTo(radius * 0.16, radius * 1.08);
  graphics.lineTo(0, radius * 1.3);
  graphics.lineTo(-radius * 0.16, radius * 1.08);
  graphics.lineTo(-radius * 0.24, radius * 0.28);
  graphics.closePath();
}

function drawButcherIcon(graphics, radius) {
  graphics.moveTo(-radius * 1.34, -radius * 1.08);
  graphics.lineTo(radius * 0.62, -radius * 0.88);
  graphics.quadraticCurveTo(radius * 0.84, -radius * 0.28, radius * 0.54, radius * 0.3);
  graphics.lineTo(-radius * 0.22, radius * 0.46);
  graphics.lineTo(-radius * 0.5, radius * 0.78);
  graphics.lineTo(-radius * 0.94, radius * 0.5);
  graphics.lineTo(-radius * 0.76, radius * 0.16);
  graphics.lineTo(-radius * 1.34, radius * 0.02);
  graphics.closePath();

  graphics.moveTo(radius * 0.1, radius * 0.3);
  graphics.lineTo(radius * 0.52, radius * 0.2);
  graphics.lineTo(radius * 1.16, radius * 1.42);
  graphics.lineTo(radius * 0.72, radius * 1.66);
  graphics.closePath();

  graphics.moveTo(radius * 0.54, radius * 1.22);
  graphics.lineTo(radius * 1.26, radius * 0.84);
  graphics.lineTo(radius * 1.42, radius * 1.12);
  graphics.lineTo(radius * 0.72, radius * 1.52);
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

function drawFireIcon(graphics, radius, dead, teamColor) {
  const outer = dead ? 0x6b7280 : teamColor;
  const mid = dead ? 0x9ca3af : 0xf97316;
  const inner = dead ? 0xd1d5db : 0xfef3c7;

  graphics.moveTo(0, radius * 1.18);
  graphics.quadraticCurveTo(-radius * 1.15, radius * 0.5, -radius * 0.74, -radius * 0.18);
  graphics.quadraticCurveTo(-radius * 0.52, -radius * 0.54, -radius * 0.38, -radius * 1.1);
  graphics.quadraticCurveTo(-radius * 0.08, -radius * 0.7, radius * 0.04, -radius * 1.55);
  graphics.quadraticCurveTo(radius * 0.78, -radius * 0.82, radius * 0.62, -radius * 0.22);
  graphics.quadraticCurveTo(radius * 1.08, radius * 0.02, radius * 0.96, radius * 0.58);
  graphics.quadraticCurveTo(radius * 0.72, radius * 1.14, 0, radius * 1.18);
  graphics.closePath();
  graphics.fill(outer);
  graphics.stroke({ color: dead ? 0x111827 : 0x7f1d1d, width: 1, alpha: dead ? 0.45 : 0.95 });

  graphics.moveTo(0, radius * 0.92);
  graphics.quadraticCurveTo(-radius * 0.58, radius * 0.46, -radius * 0.34, -radius * 0.12);
  graphics.quadraticCurveTo(-radius * 0.18, -radius * 0.48, radius * 0.02, -radius * 0.86);
  graphics.quadraticCurveTo(radius * 0.42, -radius * 0.36, radius * 0.34, radius * 0.02);
  graphics.quadraticCurveTo(radius * 0.68, radius * 0.38, 0, radius * 0.92);
  graphics.closePath();
  graphics.fill(mid);

  graphics.moveTo(0, radius * 0.68);
  graphics.quadraticCurveTo(-radius * 0.24, radius * 0.28, radius * 0.02, -radius * 0.26);
  graphics.quadraticCurveTo(radius * 0.28, radius * 0.14, 0, radius * 0.68);
  graphics.closePath();
  graphics.fill(inner);
}

function drawSnowflakeIcon(graphics, radius, dead, teamColor) {
  const color = dead ? 0x9ca3af : teamColor;
  const ice = dead ? 0xd1d5db : 0xe0faff;
  const edge = dead ? 0x374151 : 0xa5f3fc;
  const point = (angle, forward, side, scale) => ({
    x: Math.cos(angle) * radius * forward * scale + Math.cos(angle + Math.PI / 2) * radius * side * scale,
    y: Math.sin(angle) * radius * forward * scale + Math.sin(angle + Math.PI / 2) * radius * side * scale,
  });
  const drawCrystal = (offset = 0, scale = 1) => {
    for (let i = 0; i < 6; i++) {
      const angle = -Math.PI / 2 + offset + (Math.PI * i) / 3;
      const points = [
        point(angle, 0.05, -0.04, scale),
        point(angle, 0.5, -0.18, scale),
        point(angle, 0.8, -0.38, scale),
        point(angle, 0.73, -0.18, scale),
        point(angle, 1.26, -0.22, scale),
        point(angle, 1.58, 0, scale),
        point(angle, 1.26, 0.22, scale),
        point(angle, 0.73, 0.18, scale),
        point(angle, 0.8, 0.38, scale),
        point(angle, 0.5, 0.18, scale),
        point(angle, 0.05, 0.04, scale),
      ];
      graphics.moveTo(points[0].x, points[0].y);
      for (const p of points.slice(1)) {
        graphics.lineTo(p.x, p.y);
      }
      graphics.closePath();
    }
  };

  graphics.circle(0, 0, radius * 1.55);
  graphics.fill({ color: edge, alpha: dead ? 0.1 : 0.18 });
  drawCrystal();
  graphics.fill({ color, alpha: dead ? 0.72 : 0.96 });
  graphics.stroke({ color: edge, width: Math.max(1, radius * 0.06), alpha: dead ? 0.45 : 0.85 });
  drawCrystal(Math.PI / 6, 0.54);
  graphics.fill({ color: ice, alpha: dead ? 0.38 : 0.58 });
  graphics.circle(0, 0, radius * 0.12);
  graphics.fill({ color: ice, alpha: dead ? 0.65 : 0.9 });
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
