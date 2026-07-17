function drawEffects(frame) {
  skillLayer.clear();
  drawActiveSkillRanges(frame);
  drawSwordETargetCooldowns(frame);
  drawNinjaPassiveCooldowns(frame);
  drawFireMageBlazeExplosions(frame);
  drawCastWindups(frame);
  drawSkillPreview(frame);

  const visibleServants = new Set();
  for (const effect of state.effects) {
    if (effect.kind === "warrior_q_light") {
      drawWarriorQLightEffect(effect, frame);
      continue;
    }
    if (effect.kind === "warrior_w_shields") {
      drawWarriorWShieldsEffect(effect, frame);
      continue;
    }
    if (effect.kind === "warrior_r_sword") {
      drawWarriorRSwordEffect(effect, frame);
      continue;
    }
    if (effect.kind === "sword_whirlwind") {
      drawSwordWhirlwindEffect(effect, frame);
      continue;
    }
    if (effect.kind === "blade_q_heal") {
      drawBladeQHealEffect(effect, frame);
      continue;
    }
    if (effect.kind === "blade_e_whirlwind") {
      drawBladeEWhirlwindEffect(effect, frame);
      continue;
    }
    if (effect.kind === "blade_r_rage") {
      drawBladeRRageEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_q") {
      drawMonkQSonicWaveEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_q_mark") {
      drawMonkQMarkEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_q_echo") {
      drawMonkQEchoEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_w_safeguard") {
      drawMonkWSafeguardEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_w_iron_will") {
      drawMonkWIronWillEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_e_tempest") {
      drawMonkETempestEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_e_reveal") {
      drawMonkERevealEffect(effect, frame);
      continue;
    }
    if (effect.kind === "monk_e_cripple") {
      drawMonkECrippleEffect(effect, frame);
      continue;
    }
    if (effect.kind === "berserker_q") {
      if (hasLocalBerserkerQWindup()) {
        continue;
      }
      drawBerserkerQRange(
        effect.x,
        effect.y,
        effect.radius,
        effect.range,
        frame,
        effectAlpha(effect),
      );
      continue;
    }
    if (effect.kind === "berserker_r") {
      if (hasLocalBerserkerRWindup()) {
        continue;
      }
      drawBerserkerRRangeEffect(effect, frame);
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
    if (effect.kind === "siege_cannonball") {
      drawSiegeCannonballEffect(effect, frame);
      continue;
    }
    if (effect.kind === "fire_mage_q") {
      drawFireMageQEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_q") {
      drawKillerQEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_q_cast_range") {
      drawKillerQCastRangeEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_q_dagger_airborne") {
      drawKillerQAirborneDaggerEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_w_dagger_airborne") {
      drawKillerWAirborneDaggerEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_q_dagger" || effect.kind === "killer_w_dagger") {
      drawKillerQDaggerEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_dagger_slash") {
      drawKillerDaggerSlashEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_e") {
      drawKillerEEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_r_channel") {
      drawKillerRChannelEffect(effect, frame);
      continue;
    }
    if (effect.kind === "killer_r") {
      drawKillerRProjectileEffect(effect, frame);
      continue;
    }
    if (effect.kind === "fire_mage_w") {
      drawFireMageWEffect(effect, frame);
      continue;
    }
    if (effect.kind === "fire_mage_e") {
      drawFireMageEEffect(effect, frame);
      continue;
    }
    if (effect.kind === "fire_mage_r") {
      drawFireMageREffect(effect, frame);
      continue;
    }
    if (effect.kind === "doctor_q") {
      drawDoctorQEffect(effect, frame);
      continue;
    }
    if (effect.kind === "doctor_w") {
      drawDoctorWEffect(effect, frame);
      continue;
    }
    if (effect.kind === "doctor_r") {
      drawDoctorREffect(effect, frame);
      continue;
    }
    if (effect.kind === "frostmage_w") {
      drawFrostMageWEffect(effect, frame);
      continue;
    }
    if (effect.kind === "frostmage_q" || effect.kind === "frostmage_q_shard") {
      drawFrostMageQEffect(effect, frame);
      continue;
    }
    if (effect.kind === "frostmage_e") {
      drawFrostMageEEffect(effect, frame);
      continue;
    }
    if (effect.kind === "frostmage_r_enemy" || effect.kind === "frostmage_r_self") {
      drawFrostMageREffect(effect, frame);
      continue;
    }
    if (effect.kind === "frostmage_servant") {
      visibleServants.add(servantEffectID(effect));
      drawFrostMageServantEffect(effect, frame);
      continue;
    }
    if (effect.kind === "robot_q") {
      drawRobotHookProjectile(effect, frame);
      continue;
    }
    if (effect.kind === "butcher_q") {
      drawRobotHookProjectile(effect, frame);
      continue;
    }
    if (effect.kind === "robot_q_pull") {
      drawRobotHookPullEffect(effect, frame);
      continue;
    }
    if (effect.kind === "butcher_q_pull") {
      drawRobotHookPullEffect(effect, frame);
      continue;
    }
    if (effect.kind === "butcher_w") {
      drawButcherRotEffect(effect, frame);
      continue;
    }
    if (effect.kind === "butcher_e") {
      drawButcherMeatShieldEffect(effect, frame);
      continue;
    }
    if (effect.kind === "butcher_r") {
      drawButcherDismemberEffect(effect, frame);
      continue;
    }
    if (effect.kind === "robot_r") {
      drawRobotRRangeEffect(effect, frame);
      continue;
    }
    if (effect.kind === "gunner_q") {
      drawGunnerQEffect(effect, frame);
      continue;
    }
    if (effect.kind === "gunner_r") {
      drawGunnerREffect(effect, frame);
      continue;
    }
    if (effect.kind === "gunner_e") {
      drawGunnerEEffect(effect, frame);
      continue;
    }
    if (effect.kind === "explorer_q") {
      drawExplorerQEffect(effect, frame);
      continue;
    }
    if (effect.kind === "explorer_w") {
      drawExplorerWEffect(effect, frame);
      continue;
    }
    if (effect.kind === "explorer_e") {
      drawExplorerEEffect(effect, frame);
      continue;
    }
    if (effect.kind === "explorer_r") {
      drawExplorerREffect(effect, frame);
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
    if (effect.kind === "ninja_shuriken") {
      drawNinjaShurikenEffect(effect, frame);
      continue;
    }
    if (effect.kind === "ninja_e") {
      drawNinjaERangeEffect(effect, frame);
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
    if (effect.kind === "ninja_shadow") {
      drawNinjaShadowEffect(effect, frame);
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
  for (const id of state.servantEffectPositions.keys()) {
    if (!visibleServants.has(id)) {
      state.servantEffectPositions.delete(id);
    }
  }
}

function hasLocalBerserkerQWindup() {
  return state.castWindups.some((windup) => windup.skillId === "berserker_q");
}

function hasLocalBerserkerRWindup() {
  return state.castWindups.some((windup) => windup.skillId === "berserker_r");
}

function drawBerserkerRRangeEffect(effect, frame) {
  const alpha = effectAlpha(effect);
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const tx = frame.offsetX + (effect.endX || effect.x) * frame.scale;
  const ty = frame.offsetY + (effect.endY || effect.y) * frame.scale;
  skillLayer.circle(x, y, (effect.range || 460) * frame.scale);
  skillLayer.fill({ color: 0xef4444, alpha: 0.06 * alpha });
  skillLayer.circle(x, y, (effect.range || 460) * frame.scale);
  skillLayer.stroke({ color: 0xef4444, width: 3, alpha: 0.55 * alpha });
  skillLayer.moveTo(x, y);
  skillLayer.lineTo(tx, ty);
  skillLayer.stroke({ color: 0xef4444, width: 3, alpha: 0.55 * alpha });
  skillLayer.circle(tx, ty, 26);
  skillLayer.stroke({ color: 0xef4444, width: 3, alpha: 0.8 * alpha });
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
  const scaledRadius = radius * frame.scale;
  const rotation = performance.now() / 110;
  skillLayer.circle(x, y, radius * frame.scale);
  skillLayer.fill({ color: 0xfacc15, alpha: 0.1 });
  for (let index = 0; index < 3; index++) {
    const start = rotation + (Math.PI * 2 * index) / 3;
    const arcRadius = scaledRadius * (0.55 + index * 0.16);
    skillLayer.moveTo(x + Math.cos(start) * arcRadius, y + Math.sin(start) * arcRadius);
    skillLayer.arc(x, y, arcRadius, start, start + Math.PI * 1.12);
    skillLayer.stroke({
      color: index === 1 ? 0xfef3c7 : 0xfbbf24,
      width: Math.max(3, scaledRadius * (0.1 - index * 0.018)),
      alpha: 0.72 - index * 0.1,
    });
  }
  for (let index = 0; index < 2; index++) {
    const angle = rotation + Math.PI * index;
    drawWarriorSpinningSword(x, y, scaledRadius * 0.78, angle);
  }
  skillLayer.circle(x, y, hitRadius * frame.scale);
  skillLayer.stroke({ color: 0xf59e0b, width: 1, alpha: 0.28 });
}

function drawWarriorQLightEffect(effect, frame) {
  const source = effectSourcePosition(effect);
  const worldX = source?.x ?? effect.x;
  const worldY = source?.y ?? effect.y;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(50, (effect.radius || 120) * frame.scale);
  const alpha = effectAlpha(effect);
  const progress = 1 - alpha;
  const count = Math.max(8, effect.count || 18);

  skillLayer.circle(x, y, radius * (0.35 + progress * 0.65));
  skillLayer.stroke({ color: 0xfacc15, width: 4, alpha: 0.75 * alpha });
  for (let index = 0; index < count; index++) {
    const angle = (Math.PI * 2 * index) / count + progress * 0.8;
    const distance = radius * (0.18 + progress * (0.55 + (index % 4) * 0.06));
    const px = x + Math.cos(angle) * distance;
    const py = y + Math.sin(angle) * distance - radius * progress * 0.35;
    const size = Math.max(2, radius * (0.025 + (index % 3) * 0.008));
    skillLayer.circle(px, py, size);
    skillLayer.fill({
      color: index % 3 === 0 ? 0xffffff : index % 2 === 0 ? 0xfef08a : 0xfbbf24,
      alpha: alpha * 0.9,
    });
  }
}

function drawWarriorWShieldsEffect(effect, frame) {
  const source = effectSourcePosition(effect);
  const worldX = source?.x ?? effect.x;
  const worldY = source?.y ?? effect.y;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(38, (effect.radius || 100) * frame.scale);
  const count = Math.max(1, effect.count || 3);
  const rotation = performance.now() / 700;
  const alpha = Math.min(1, effectAlpha(effect) * 4);

  skillLayer.circle(x, y, radius * 0.72);
  skillLayer.stroke({ color: 0xfde68a, width: 2, alpha: 0.3 * alpha });
  for (let index = 0; index < count; index++) {
    const angle = rotation + (Math.PI * 2 * index) / count;
    const shieldX = x + Math.cos(angle) * radius;
    const shieldY = y + Math.sin(angle) * radius;
    drawWarriorShield(shieldX, shieldY, Math.max(9, radius * 0.25), angle, alpha);
  }
}

function drawWarriorShield(x, y, size, angle, alpha) {
  const radialX = Math.cos(angle);
  const radialY = Math.sin(angle);
  const sideX = -radialY;
  const sideY = radialX;
  skillLayer.moveTo(x - sideX * size * 0.7 - radialX * size * 0.55, y - sideY * size * 0.7 - radialY * size * 0.55);
  skillLayer.lineTo(x + sideX * size * 0.7 - radialX * size * 0.55, y + sideY * size * 0.7 - radialY * size * 0.55);
  skillLayer.lineTo(x + sideX * size * 0.55 + radialX * size * 0.35, y + sideY * size * 0.55 + radialY * size * 0.35);
  skillLayer.lineTo(x + radialX * size, y + radialY * size);
  skillLayer.lineTo(x - sideX * size * 0.55 + radialX * size * 0.35, y - sideY * size * 0.55 + radialY * size * 0.35);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xf59e0b, alpha: 0.72 * alpha });
  skillLayer.stroke({ color: 0xfffbeb, width: 2, alpha: 0.92 * alpha });
}

function drawWarriorSpinningSword(x, y, distance, angle) {
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const centerX = x + forwardX * distance;
  const centerY = y + forwardY * distance;
  const length = Math.max(10, distance * 0.42);
  const width = Math.max(4, length * 0.16);
  skillLayer.moveTo(centerX + forwardX * length * 0.58, centerY + forwardY * length * 0.58);
  skillLayer.lineTo(centerX - forwardX * length * 0.48 + sideX * width, centerY - forwardY * length * 0.48 + sideY * width);
  skillLayer.lineTo(centerX - forwardX * length * 0.48 - sideX * width, centerY - forwardY * length * 0.48 - sideY * width);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xfffbeb, alpha: 0.9 });
  skillLayer.moveTo(centerX - sideX * width * 1.8, centerY - sideY * width * 1.8);
  skillLayer.lineTo(centerX + sideX * width * 1.8, centerY + sideY * width * 1.8);
  skillLayer.stroke({ color: 0xf59e0b, width: 3, alpha: 0.9 });
}

function drawWarriorRSwordEffect(effect, frame) {
  const worldX = effect.endX ?? effect.x;
  const worldY = effect.endY ?? effect.y;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(60, (effect.radius || 108) * frame.scale);
  const tick = interpolatedTick();
  const duration = Math.max(1, (effect.expiresAt || tick + 1) - (effect.createdAt || tick));
  const progress = clamp((tick - (effect.createdAt || tick)) / duration, 0, 1);
  const drop = clamp(progress / 0.34, 0, 1);
  const alpha = Math.min(1, effectAlpha(effect) * 3);
  const swordLength = radius * 1.25;
  const tipY = y - radius * 2.2 * (1 - drop);
  const bladeTop = tipY - swordLength;
  const bladeWidth = Math.max(8, radius * 0.16);

  skillLayer.moveTo(x, tipY);
  skillLayer.lineTo(x - bladeWidth, bladeTop + swordLength * 0.18);
  skillLayer.lineTo(x - bladeWidth * 0.7, bladeTop);
  skillLayer.lineTo(x + bladeWidth * 0.7, bladeTop);
  skillLayer.lineTo(x + bladeWidth, bladeTop + swordLength * 0.18);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xfffbeb, alpha: 0.94 * alpha });
  skillLayer.stroke({ color: 0xf59e0b, width: 3, alpha });
  skillLayer.moveTo(x - bladeWidth * 2.4, bladeTop - 3);
  skillLayer.lineTo(x + bladeWidth * 2.4, bladeTop - 3);
  skillLayer.stroke({ color: 0xfbbf24, width: 6, alpha });
  skillLayer.rect(x - bladeWidth * 0.35, bladeTop - radius * 0.42, bladeWidth * 0.7, radius * 0.42);
  skillLayer.fill({ color: 0x92400e, alpha });

  if (drop >= 1) {
    const impact = clamp((progress - 0.34) / 0.66, 0, 1);
    skillLayer.circle(x, y, radius * (0.28 + impact * 0.9));
    skillLayer.stroke({ color: 0xfacc15, width: Math.max(3, radius * 0.08), alpha: (1 - impact) * alpha });
    skillLayer.circle(x, y, radius * 0.32);
    skillLayer.fill({ color: 0xfef3c7, alpha: 0.2 * (1 - impact) });
  }
}

function drawBerserkerQRange(
  worldX,
  worldY,
  innerRadius,
  outerRadius,
  frame,
  alpha = 1,
) {
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const inner = (innerRadius || 300) * frame.scale;
  const outer = (outerRadius || 425) * frame.scale;
  skillLayer.circle(x, y, outer);
  skillLayer.fill({ color: 0xf97316, alpha: 0.08 * alpha });
  skillLayer.circle(x, y, inner);
  skillLayer.fill({ color: 0xdc2626, alpha: 0.11 * alpha });
  skillLayer.circle(x, y, inner);
  skillLayer.stroke({ color: 0xdc2626, width: 2, alpha: 0.72 * alpha });
  skillLayer.circle(x, y, outer);
  skillLayer.stroke({ color: 0xf59e0b, width: 3, alpha: 0.82 * alpha });
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
  if (effect.sourceHeroId === "explorer") {
    drawExplorerBasicEffect(effect, frame);
    return;
  }
  if (effect.sourceHeroId === "frostmage") {
    drawArrowProjectile(effect, frame, 0xbae6fd, 0x38bdf8, {
      fromSnapshot: true,
    });
    return;
  }
  if (effect.sourceHeroId === "fire_mage") {
    drawArrowProjectile(effect, frame, 0xf97316, 0xef4444, {
      fromSnapshot: true,
    });
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

function drawGunnerQEffect(effect, frame) {
  drawArrowProjectile(effect, frame, 0xfde68a, 0xf97316, {
    fromSnapshot: true,
  });
}

function drawFireMageQEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(8, (effect.radius || 28) * frame.scale * 0.55);
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const tail = radius * 3;
  skillLayer.moveTo(x - Math.cos(angle) * tail, y - Math.sin(angle) * tail);
  skillLayer.lineTo(x, y);
  skillLayer.stroke({ color: 0xf97316, width: Math.max(3, radius), alpha: 0.42 });
  skillLayer.circle(x, y, radius * 1.45);
  skillLayer.fill({ color: 0xef4444, alpha: 0.28 });
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0xf97316, alpha: 0.86 });
  skillLayer.circle(x + Math.cos(angle) * radius * 0.28, y + Math.sin(angle) * radius * 0.28, radius * 0.42);
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.92 });
}

function drawKillerQEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const length = Math.max(15, (effect.radius || 24) * frame.scale * 1.25);
  const width = Math.max(5, length * 0.28);

  skillLayer.moveTo(x + forwardX * length, y + forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.75 + sideX * width, y - forwardY * length * 0.75 + sideY * width);
  skillLayer.lineTo(x - forwardX * length, y - forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.75 - sideX * width, y - forwardY * length * 0.75 - sideY * width);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xe5e7eb, alpha: 0.94 });
  skillLayer.stroke({ color: 0x7c3aed, width: 2, alpha: 0.9 });
}

function drawKillerQCastRangeEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const targetX = frame.offsetX + effect.endX * frame.scale;
  const targetY = frame.offsetY + effect.endY * frame.scale;
  const range = (effect.range || 625) * frame.scale;
  const targetRadius = Math.max(18, (effect.radius || 18) * frame.scale + 8);
  const alpha = effectAlpha(effect);

  skillLayer.circle(x, y, range);
  skillLayer.fill({ color: 0x7c3aed, alpha: 0.035 * alpha });
  skillLayer.circle(x, y, range);
  skillLayer.stroke({ color: 0xa78bfa, width: 2, alpha: 0.48 * alpha });
  skillLayer.circle(targetX, targetY, targetRadius);
  skillLayer.stroke({ color: 0xe9d5ff, width: 3, alpha: 0.9 * alpha });
}

function drawBladeQHealEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = Math.max(22, (effect.radius || 90) * frame.scale);
  const tick = interpolatedTick();
  const duration = Math.max(1, (effect.expiresAt || tick + 1) - (effect.createdAt || tick));
  const progress = clamp((tick - (effect.createdAt || tick)) / duration, 0, 1);
  const alpha = 1 - progress;

  skillLayer.circle(x, y, radius * (0.42 + progress * 0.58));
  skillLayer.stroke({ color: 0x4ade80, width: 4, alpha: 0.72 * alpha });
  skillLayer.circle(x, y, radius * (0.22 + progress * 0.36));
  skillLayer.fill({ color: 0x86efac, alpha: 0.14 * alpha });
  for (let index = 0; index < 8; index++) {
    const angle = (Math.PI * 2 * index) / 8 + progress * 0.7;
    const distance = radius * (0.18 + progress * (0.5 + (index % 3) * 0.08));
    const px = x + Math.cos(angle) * distance;
    const py = y + Math.sin(angle) * distance - radius * progress * 0.42;
    const size = Math.max(3, radius * 0.055 * (1 - progress * 0.35));
    skillLayer
      .moveTo(px - size, py)
      .lineTo(px + size, py)
      .moveTo(px, py - size)
      .lineTo(px, py + size);
  }
  skillLayer.stroke({ color: 0xbbf7d0, width: 2.5, alpha: 0.88 * alpha });
}

function drawBladeEWhirlwindEffect(effect, frame) {
  const position = movingEffectPosition(effect);
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(20, (effect.radius || 70) * frame.scale);
  const rotation = performance.now() / 95;
  const alpha = Math.max(0.35, effectAlpha(effect));

  skillLayer.circle(x, y, radius * 0.72);
  skillLayer.fill({ color: 0xdbeafe, alpha: 0.11 * alpha });
  for (let index = 0; index < 3; index++) {
    const start = rotation + (Math.PI * 2 * index) / 3;
    const arcRadius = radius * (0.48 + index * 0.18);
    skillLayer.moveTo(
      x + Math.cos(start) * arcRadius,
      y + Math.sin(start) * arcRadius,
    );
    skillLayer.arc(x, y, arcRadius, start, start + Math.PI * 1.28);
    skillLayer.stroke({
      color: index === 1 ? 0xe0f2fe : 0x93c5fd,
      width: Math.max(2, radius * (0.12 - index * 0.018)),
      alpha: (0.82 - index * 0.12) * alpha,
    });
  }
  for (let index = 0; index < 4; index++) {
    const angle = rotation + (Math.PI * 2 * index) / 4;
    const bladeX = x + Math.cos(angle) * radius * 0.72;
    const bladeY = y + Math.sin(angle) * radius * 0.72;
    const sideX = -Math.sin(angle);
    const sideY = Math.cos(angle);
    const length = radius * 0.24;
    skillLayer
      .moveTo(bladeX + Math.cos(angle) * length, bladeY + Math.sin(angle) * length)
      .lineTo(bladeX - Math.cos(angle) * length + sideX * length * 0.36, bladeY - Math.sin(angle) * length + sideY * length * 0.36)
      .lineTo(bladeX - Math.cos(angle) * length - sideX * length * 0.36, bladeY - Math.sin(angle) * length - sideY * length * 0.36)
      .closePath();
  }
  skillLayer.fill({ color: 0xf8fafc, alpha: 0.86 * alpha });
}

function drawBladeRRageEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = Math.max(28, (effect.radius || 110) * frame.scale);
  const alpha = Math.max(0.3, effectAlpha(effect));
  const time = performance.now() / 1000;

  skillLayer.circle(x, y, radius * (0.78 + Math.sin(time * 8) * 0.05));
  skillLayer.fill({ color: 0x991b1b, alpha: 0.09 * alpha });
  skillLayer.circle(x, y, radius * 0.9);
  skillLayer.stroke({ color: 0xef4444, width: 3, alpha: 0.58 * alpha });
  for (let index = 0; index < 14; index++) {
    const phase = (time * (0.55 + (index % 4) * 0.08) + index / 14) % 1;
    const angle = (Math.PI * 2 * index) / 14 + Math.sin(time * 1.8 + index) * 0.22;
    const distance = radius * (0.34 + phase * 0.72);
    const px = x + Math.cos(angle) * distance;
    const py = y + Math.sin(angle) * distance - radius * phase * 0.45;
    const size = Math.max(3, radius * 0.065 * (1 - phase * 0.45));
    const particleAlpha = alpha * (1 - phase) * 0.9;
    skillLayer.circle(px, py, size);
    skillLayer.fill({
      color: index % 3 === 0 ? 0xfca5a5 : index % 2 === 0 ? 0xef4444 : 0xb91c1c,
      alpha: particleAlpha,
    });
  }
}

function drawMonkQSonicWaveEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const radius = Math.max(10, (effect.radius || 35) * frame.scale);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;

  for (let index = 0; index < 3; index++) {
    const distance = index * radius * 0.42;
    const centerX = x - forwardX * distance;
    const centerY = y - forwardY * distance;
    const width = radius * (0.78 - index * 0.16);
    skillLayer
      .moveTo(centerX + sideX * width, centerY + sideY * width)
      .quadraticCurveTo(
        centerX + forwardX * radius * 0.5,
        centerY + forwardY * radius * 0.5,
        centerX - sideX * width,
        centerY - sideY * width,
      );
    skillLayer.stroke({ color: index === 0 ? 0xe0f2fe : 0x38bdf8, width: 3 - index * 0.5, alpha: 0.9 - index * 0.2 });
  }
  skillLayer.circle(x, y, radius * 0.26);
  skillLayer.fill({ color: 0xecfeff, alpha: 0.95 });
}

function drawMonkQMarkEffect(effect, frame) {
  const worldX = effect.endX ?? effect.x;
  const worldY = effect.endY ?? effect.y;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(20, ((effect.radius || 18) + 34) * frame.scale);
  const pulse = 0.86 + Math.sin(performance.now() / 95) * 0.1;
  const alpha = Math.max(0.35, effectAlpha(effect));

  skillLayer.circle(x, y, radius * pulse);
  skillLayer.stroke({ color: 0x22d3ee, width: 3, alpha: 0.82 * alpha });
  for (let index = 0; index < 4; index++) {
    const angle = Math.PI / 4 + (Math.PI * index) / 2;
    const inner = radius * 0.62;
    const outer = radius * 0.94;
    skillLayer
      .moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner)
      .lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
  }
  skillLayer.stroke({ color: 0xa5f3fc, width: 2, alpha: 0.76 * alpha });
}

function drawMonkQEchoEffect(effect, frame) {
  const position = movingEffectPosition(effect);
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(16, (effect.radius || 55) * frame.scale);
  const angle = Math.atan2(effect.endY - effect.y, effect.endX - effect.x);
  const backX = Math.cos(angle + Math.PI);
  const backY = Math.sin(angle + Math.PI);

  for (let index = 0; index < 4; index++) {
    const offset = radius * index * 0.42;
    skillLayer.circle(x + backX * offset, y + backY * offset, radius * (0.72 - index * 0.1));
    skillLayer.fill({ color: index === 0 ? 0x67e8f9 : 0x0ea5e9, alpha: 0.2 - index * 0.035 });
  }
  skillLayer.circle(x, y, radius * 0.68);
  skillLayer.stroke({ color: 0xecfeff, width: 3, alpha: 0.85 });
}

function drawMonkWSafeguardEffect(effect, frame) {
  const tick = interpolatedTick();
  const duration = Math.max(1, (effect.expiresAt || tick + 1) - (effect.createdAt || tick));
  const progress = clamp((tick - (effect.createdAt || tick)) / duration, 0, 1);
  const worldX = effect.x + (effect.endX - effect.x) * progress;
  const worldY = effect.y + (effect.endY - effect.y) * progress;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(22, (effect.radius || 85) * frame.scale);
  const alpha = 1 - progress * 0.35;

  skillLayer.circle(x, y, radius * (0.72 + progress * 0.18));
  skillLayer.fill({ color: 0xfacc15, alpha: 0.1 * alpha });
  skillLayer.circle(x, y, radius * 0.82);
  skillLayer.stroke({ color: 0xfef08a, width: 4, alpha: 0.82 * alpha });
  skillLayer
    .moveTo(x, y - radius * 0.56)
    .lineTo(x + radius * 0.48, y - radius * 0.2)
    .lineTo(x + radius * 0.34, y + radius * 0.48)
    .lineTo(x, y + radius * 0.68)
    .lineTo(x - radius * 0.34, y + radius * 0.48)
    .lineTo(x - radius * 0.48, y - radius * 0.2)
    .closePath();
  skillLayer.stroke({ color: 0xfde68a, width: 2, alpha: 0.72 * alpha });
}

function drawMonkWIronWillEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = Math.max(20, (effect.radius || 75) * frame.scale);
  const rotation = performance.now() / 240;
  const alpha = Math.max(0.3, effectAlpha(effect));

  skillLayer.circle(x, y, radius * 0.76);
  skillLayer.fill({ color: 0xf59e0b, alpha: 0.08 * alpha });
  for (let index = 0; index < 3; index++) {
    const start = rotation + (Math.PI * 2 * index) / 3;
    const arcRadius = radius * (0.62 + index * 0.09);
    skillLayer.moveTo(x + Math.cos(start) * arcRadius, y + Math.sin(start) * arcRadius);
    skillLayer.arc(x, y, arcRadius, start, start + Math.PI * 1.18);
    skillLayer.stroke({ color: index === 0 ? 0xfef3c7 : 0xfbbf24, width: 3, alpha: (0.78 - index * 0.14) * alpha });
  }
}

function drawMonkETempestEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 350) * frame.scale;
  const tick = interpolatedTick();
  const duration = Math.max(1, (effect.expiresAt || tick + 1) - (effect.createdAt || tick));
  const progress = clamp((tick - (effect.createdAt || tick)) / duration, 0, 1);
  const alpha = Math.max(0, 1 - progress);
  const reach = radius * Math.min(1, progress * 2.8);

  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x082f49, alpha: 0.16 * alpha });

  for (let index = 0; index < 8; index++) {
    const angle = (Math.PI * 2 * index) / 8 + 0.18 * Math.sin(index * 2.7);
    const length = reach * (0.68 + 0.3 * Math.sin(index * 4.1) ** 2);
    skillLayer.moveTo(x, y);
    for (let step = 1; step <= 4; step++) {
      const distance = length * step / 4;
      const bend = Math.sin(index * 3.4 + step * 4.7) * radius * 0.035;
      skillLayer.lineTo(
        x + Math.cos(angle) * distance - Math.sin(angle) * bend,
        y + Math.sin(angle) * distance + Math.cos(angle) * bend,
      );
    }
    skillLayer.stroke({ color: 0x0284c7, width: 7, alpha: 0.3 * alpha });
    skillLayer.stroke({ color: 0xe0f2fe, width: 2.5, alpha: 0.9 * alpha });
  }

  for (let index = 0; index < 2; index++) {
    const ringProgress = clamp((progress - index * 0.14) / (1 - index * 0.14), 0, 1);
    skillLayer.circle(x, y, radius * (0.08 + ringProgress * 0.92));
    skillLayer.stroke({
      color: index === 0 ? 0x7dd3fc : 0xffffff,
      width: index === 0 ? 6 : 3,
      alpha: (0.75 - index * 0.18) * (1 - ringProgress),
    });
  }

  skillLayer.circle(x, y, radius * (0.2 - progress * 0.12));
  skillLayer.fill({ color: 0xf0f9ff, alpha: 0.72 * alpha });
  skillLayer.circle(x, y, radius * (0.32 + progress * 0.16));
  skillLayer.stroke({ color: 0x38bdf8, width: 4, alpha: 0.45 * alpha });
}

function drawMonkERevealEffect(effect, frame) {
  const worldX = effect.endX ?? effect.x;
  const worldY = effect.endY ?? effect.y;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(16, ((effect.radius || 18) + 18) * frame.scale);
  const alpha = Math.max(0.2, effectAlpha(effect));
  const pulse = 0.86 + Math.sin(performance.now() / 130) * 0.1;

  skillLayer.ellipse(x, y, radius * pulse, radius * 0.48 * pulse);
  skillLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.62 * alpha });
  skillLayer.circle(x, y, radius * 0.18);
  skillLayer.fill({ color: 0xe0f2fe, alpha: 0.78 * alpha });
}

function drawMonkECrippleEffect(effect, frame) {
  const worldX = effect.endX ?? effect.x;
  const worldY = effect.endY ?? effect.y;
  const x = frame.offsetX + worldX * frame.scale;
  const y = frame.offsetY + worldY * frame.scale;
  const radius = Math.max(24, (effect.radius || 70) * frame.scale);
  const tick = interpolatedTick();
  const duration = Math.max(1, (effect.expiresAt || tick + 1) - (effect.createdAt || tick));
  const progress = clamp((tick - (effect.createdAt || tick)) / duration, 0, 1);
  const alpha = 1 - progress;

  for (let index = 0; index < 3; index++) {
    const ring = radius * (1.12 - progress * 0.58 - index * 0.16);
    if (ring <= 0) {
      continue;
    }
    skillLayer.circle(x, y, ring);
    skillLayer.stroke({ color: index === 0 ? 0xf97316 : 0xfbbf24, width: 3, alpha: (0.78 - index * 0.16) * alpha });
  }
}

function drawKillerWAirborneDaggerEffect(effect, frame) {
  const tick = interpolatedTick();
  const createdAt = effect.createdAt ?? tick;
  const expiresAt = effect.expiresAt ?? createdAt + 1;
  const duration = Math.max(1, expiresAt - createdAt);
  const progress = clamp((tick - createdAt) / duration, 0, 1);
  const groundX = frame.offsetX + effect.x * frame.scale;
  const groundY = frame.offsetY + effect.y * frame.scale;
  const height = Math.sin(Math.PI * progress) * 95 * frame.scale;
  const x = groundX;
  const y = groundY - height;
  const angle = progress * Math.PI * 4;
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const length = Math.max(16, (effect.radius || 32) * frame.scale * 0.72);
  const width = Math.max(5, length * 0.28);

  skillLayer.ellipse(groundX, groundY, length * (0.8 - progress * 0.25), length * 0.28);
  skillLayer.fill({ color: 0x312e81, alpha: 0.2 });
  skillLayer.moveTo(x + forwardX * length, y + forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.5 + sideX * width, y - forwardY * length * 0.5 + sideY * width);
  skillLayer.lineTo(x - forwardX * length, y - forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.5 - sideX * width, y - forwardY * length * 0.5 - sideY * width);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xf3f4f6, alpha: 0.94 });
  skillLayer.stroke({ color: 0x8b5cf6, width: 2, alpha: 0.92 });
}

function drawKillerQAirborneDaggerEffect(effect, frame) {
  const tick = interpolatedTick();
  const createdAt = effect.createdAt ?? tick;
  const expiresAt = effect.expiresAt ?? createdAt + 1;
  const duration = Math.max(1, expiresAt - createdAt);
  const progress = clamp((tick - createdAt) / duration, 0, 1);
  const startX = frame.offsetX + effect.x * frame.scale;
  const startY = frame.offsetY + effect.y * frame.scale;
  const endX = frame.offsetX + effect.endX * frame.scale;
  const endY = frame.offsetY + effect.endY * frame.scale;
  const groundX = startX + (endX - startX) * progress;
  const groundY = startY + (endY - startY) * progress;
  const height = Math.sin(Math.PI * progress) * 72 * frame.scale;
  const x = groundX;
  const y = groundY - height;
  const angle = Math.atan2(endY - startY, endX - startX) + progress * Math.PI * 5;
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const length = Math.max(15, (effect.radius || 32) * frame.scale * 0.76);
  const width = Math.max(5, length * 0.28);

  skillLayer
    .moveTo(startX, startY)
    .lineTo(groundX, groundY);
  skillLayer.stroke({ color: 0xa78bfa, width: 2, alpha: 0.28 });
  skillLayer.ellipse(endX, endY, length * 0.72, length * 0.22);
  skillLayer.fill({ color: 0x312e81, alpha: 0.18 + progress * 0.12 });
  skillLayer.moveTo(x + forwardX * length, y + forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.55 + sideX * width, y - forwardY * length * 0.55 + sideY * width);
  skillLayer.lineTo(x - forwardX * length, y - forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.55 - sideX * width, y - forwardY * length * 0.55 - sideY * width);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xf5f3ff, alpha: 0.96 });
  skillLayer.stroke({ color: 0x7c3aed, width: 2, alpha: 0.94 });
}

function drawKillerQDaggerEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1) + Math.PI / 2;
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const alpha = effectAlpha(effect);
  const length = Math.max(18, (effect.radius || 32) * frame.scale * 0.9);
  const width = Math.max(6, length * 0.3);

  skillLayer.circle(x, y, length * 0.85);
  skillLayer.fill({ color: 0x7c3aed, alpha: 0.12 * alpha });
  skillLayer.circle(x, y, length * 0.85);
  skillLayer.stroke({ color: 0xa78bfa, width: 2, alpha: 0.55 * alpha });
  skillLayer.moveTo(x + forwardX * length, y + forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.5 + sideX * width, y - forwardY * length * 0.5 + sideY * width);
  skillLayer.lineTo(x - forwardX * length, y - forwardY * length);
  skillLayer.lineTo(x - forwardX * length * 0.5 - sideX * width, y - forwardY * length * 0.5 - sideY * width);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xf3f4f6, alpha: 0.9 * alpha });
  skillLayer.stroke({ color: 0x6d28d9, width: 2, alpha: 0.92 * alpha });
}

function drawKillerDaggerSlashEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 340) * frame.scale;
  const alpha = effectAlpha(effect);
  const rotation = performance.now() / 160;

  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x7c3aed, alpha: 0.08 * alpha });
  skillLayer.moveTo(
    x + Math.cos(rotation) * radius * 0.92,
    y + Math.sin(rotation) * radius * 0.92,
  );
  skillLayer.arc(x, y, radius * 0.92, rotation, rotation + Math.PI * 1.45);
  skillLayer.stroke({ color: 0xc4b5fd, width: 5, alpha: 0.82 * alpha });
  const innerStart = rotation + Math.PI;
  skillLayer.moveTo(
    x + Math.cos(innerStart) * radius * 0.68,
    y + Math.sin(innerStart) * radius * 0.68,
  );
  skillLayer.arc(x, y, radius * 0.68, innerStart, rotation + Math.PI * 2.3);
  skillLayer.stroke({ color: 0x8b5cf6, width: 3, alpha: 0.72 * alpha });
  for (let index = 0; index < 4; index++) {
    const angle = rotation + (Math.PI * 2 * index) / 4;
    const bladeX = x + Math.cos(angle) * radius * 0.78;
    const bladeY = y + Math.sin(angle) * radius * 0.78;
    const sideX = -Math.sin(angle);
    const sideY = Math.cos(angle);
    const length = Math.max(12, radius * 0.1);
    skillLayer.moveTo(bladeX + Math.cos(angle) * length, bladeY + Math.sin(angle) * length);
    skillLayer.lineTo(bladeX - Math.cos(angle) * length + sideX * length * 0.35, bladeY - Math.sin(angle) * length + sideY * length * 0.35);
    skillLayer.lineTo(bladeX - Math.cos(angle) * length - sideX * length * 0.35, bladeY - Math.sin(angle) * length - sideY * length * 0.35);
    skillLayer.closePath();
  }
  skillLayer.fill({ color: 0xf3f4f6, alpha: 0.88 * alpha });
  skillLayer.stroke({ color: 0x6d28d9, width: 1.5, alpha: 0.9 * alpha });
}

function drawKillerEEffect(effect, frame) {
  const startX = frame.offsetX + effect.x * frame.scale;
  const startY = frame.offsetY + effect.y * frame.scale;
  const endX = frame.offsetX + effect.endX * frame.scale;
  const endY = frame.offsetY + effect.endY * frame.scale;
  const alpha = effectAlpha(effect);
  const distance = Math.hypot(endX - startX, endY - startY);
  const segments = Math.max(3, Math.ceil(distance / 70));

  for (let index = 0; index <= segments; index++) {
    const progress = index / segments;
    const x = startX + (endX - startX) * progress;
    const y = startY + (endY - startY) * progress;
    const radius = 16 + Math.sin(Math.PI * progress) * 8;
    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0x8b5cf6, alpha: 0.08 * alpha * (1 - progress * 0.45) });
    skillLayer.circle(x, y, radius * 0.45);
    skillLayer.fill({ color: 0xe9d5ff, alpha: 0.22 * alpha });
  }
  skillLayer.circle(endX, endY, 30);
  skillLayer.stroke({ color: 0xc4b5fd, width: 3, alpha: 0.82 * alpha });
}

function drawKillerRChannelEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 550) * frame.scale;
  const alpha = Math.max(0.3, effectAlpha(effect));
  const rotation = performance.now() / 210;

  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x581c87, alpha: 0.055 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xc084fc, width: 2, alpha: 0.42 * alpha });
  for (let index = 0; index < 10; index++) {
    const angle = rotation + (Math.PI * 2 * index) / 10;
    const inner = radius * 0.22;
    const outer = radius * 0.38;
    skillLayer
      .moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner)
      .lineTo(x + Math.cos(angle + 0.15) * outer, y + Math.sin(angle + 0.15) * outer);
  }
  skillLayer.stroke({ color: 0xe9d5ff, width: 3, alpha: 0.62 * alpha });
}

function drawKillerRProjectileEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const length = Math.max(13, (effect.radius || 18) * frame.scale * 1.45);
  const width = Math.max(4, length * 0.28);

  skillLayer
    .moveTo(x + forwardX * length, y + forwardY * length)
    .lineTo(x - forwardX * length + sideX * width, y - forwardY * length + sideY * width)
    .lineTo(x - forwardX * length - sideX * width, y - forwardY * length - sideY * width)
    .closePath();
  skillLayer.fill({ color: 0xf5f3ff, alpha: 0.95 });
  skillLayer.stroke({ color: 0x9333ea, width: 2, alpha: 0.92 });
}

function drawFireMageWEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 260) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0xf97316, alpha: 0.08 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xef4444, width: 3, alpha: 0.75 * alpha });
  skillLayer.circle(x, y, Math.max(8, radius * 0.12));
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.22 * alpha });
}

function drawFrostMageWEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 450) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x7dd3fc, alpha: 0.08 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x38bdf8, width: 3, alpha: 0.78 * alpha });
  skillLayer.circle(x, y, Math.max(8, radius * 0.08));
  skillLayer.fill({ color: 0xe0f2fe, alpha: 0.35 * alpha });
}

function drawFrostMageQEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(7, (effect.radius || 75) * frame.scale * 0.38);
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const tail = radius * 3.8;
  skillLayer.moveTo(x - Math.cos(angle) * tail, y - Math.sin(angle) * tail);
  skillLayer.lineTo(x, y);
  skillLayer.stroke({ color: 0x7dd3fc, width: Math.max(4, radius * 0.9), alpha: 0.42 });
  skillLayer.moveTo(x + Math.cos(angle) * radius * 1.7, y + Math.sin(angle) * radius * 1.7);
  skillLayer.lineTo(x - Math.cos(angle) * radius * 1.1 - Math.sin(angle) * radius * 0.5, y - Math.sin(angle) * radius * 1.1 + Math.cos(angle) * radius * 0.5);
  skillLayer.lineTo(x - Math.cos(angle) * radius * 0.6, y - Math.sin(angle) * radius * 0.6);
  skillLayer.lineTo(x - Math.cos(angle) * radius * 1.1 + Math.sin(angle) * radius * 0.5, y - Math.sin(angle) * radius * 1.1 - Math.cos(angle) * radius * 0.5);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xbfdbfe, alpha: 0.9 });
  skillLayer.stroke({ color: 0x0ea5e9, width: 2, alpha: 0.85 });
}

function drawFrostMageEEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(10, (effect.radius || 90) * frame.scale * 0.45);
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  drawProjectileSweepArea(effect, frame, position, radius, 0x7dd3fc, 0x38bdf8);
  skillLayer.moveTo(x + forwardX * radius * 1.35, y + forwardY * radius * 1.35);
  skillLayer.lineTo(x - forwardX * radius * 0.45 + sideX * radius * 0.95, y - forwardY * radius * 0.45 + sideY * radius * 0.95);
  skillLayer.lineTo(x - forwardX * radius * 0.15, y - forwardY * radius * 0.15);
  skillLayer.lineTo(x - forwardX * radius * 0.45 - sideX * radius * 0.95, y - forwardY * radius * 0.45 - sideY * radius * 0.95);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xbae6fd, alpha: 0.86 });
  skillLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.9 });
}

function drawFrostMageREffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 550) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x93c5fd, alpha: 0.1 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x60a5fa, width: 4, alpha: 0.82 * alpha });
  skillLayer.circle(x, y, Math.max(12, radius * 0.11));
  skillLayer.fill({ color: 0xe0f2fe, alpha: 0.45 * alpha });
}

function drawFrostMageServantEffect(effect, frame) {
  const position = smoothedServantEffectPosition(effect);
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = (effect.radius || 450) * frame.scale;
  const alpha = 1;
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x7dd3fc, alpha: 0.06 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.62 * alpha });
  skillLayer.circle(x, y, 16);
  skillLayer.fill({ color: colorForTeam(effect.team), alpha: 0.72 });
  skillLayer.circle(x, y, 10);
  skillLayer.fill({ color: 0xe0f2fe, alpha: 0.9 });
  skillLayer.moveTo(x, y - 18);
  skillLayer.lineTo(x + 10, y);
  skillLayer.lineTo(x, y + 18);
  skillLayer.lineTo(x - 10, y);
  skillLayer.closePath();
  skillLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.9 });
}

function smoothedServantEffectPosition(effect) {
  const id = servantEffectID(effect);
  const now = performance.now();
  const targetX = effect.x || 0;
  const targetY = effect.y || 0;
  let position = state.servantEffectPositions.get(id);
  if (!position) {
    position = { x: targetX, y: targetY, lastMs: now };
    state.servantEffectPositions.set(id, position);
    return position;
  }
  const smoothing = 1 - Math.exp(-(now - position.lastMs) / 80);
  position.x += (targetX - position.x) * smoothing;
  position.y += (targetY - position.y) * smoothing;
  position.lastMs = now;
  return position;
}

function servantEffectID(effect) {
  return effect.id || `${effect.sourceId || "frostmage_servant"}:${effect.createdAt || 0}`;
}

function drawFireMageEEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 600) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xf97316, width: 2, alpha: 0.42 * alpha });
  skillLayer.circle(x, y, Math.max(10, radius * 0.08));
  skillLayer.fill({ color: 0xef4444, alpha: 0.28 * alpha });
  skillLayer.circle(x, y, Math.max(5, radius * 0.035));
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.82 * alpha });
}

function drawFireMageREffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(12, (effect.radius || 36) * frame.scale * 0.62);
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const tail = radius * 3.4;
  skillLayer.moveTo(x - Math.cos(angle) * tail, y - Math.sin(angle) * tail);
  skillLayer.lineTo(x, y);
  skillLayer.stroke({ color: 0xdc2626, width: Math.max(5, radius * 1.1), alpha: 0.46 });
  skillLayer.circle(x, y, radius * 1.75);
  skillLayer.fill({ color: 0xef4444, alpha: 0.26 });
  skillLayer.circle(x, y, radius * 1.08);
  skillLayer.fill({ color: 0xf97316, alpha: 0.9 });
  skillLayer.circle(x + Math.cos(angle) * radius * 0.32, y + Math.sin(angle) * radius * 0.32, radius * 0.48);
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.94 });
}

function drawDoctorQEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(9, (effect.radius || 60) * frame.scale * 0.4);
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const tail = radius * 4;

  skillLayer
    .moveTo(x - forwardX * tail, y - forwardY * tail)
    .lineTo(x, y);
  skillLayer.stroke({ color: 0x22c55e, width: Math.max(4, radius * 0.85), alpha: 0.38 });
  skillLayer
    .moveTo(x - forwardX * radius * 1.8, y - forwardY * radius * 1.8)
    .lineTo(x - forwardX * radius * 0.45, y - forwardY * radius * 0.45);
  skillLayer.stroke({ color: 0xbbf7d0, width: Math.max(3, radius * 0.34), alpha: 0.82 });

  skillLayer
    .moveTo(x + forwardX * radius * 1.65, y + forwardY * radius * 1.65)
    .lineTo(
      x - forwardX * radius * 0.55 + sideX * radius * 0.8,
      y - forwardY * radius * 0.55 + sideY * radius * 0.8,
    )
    .quadraticCurveTo(
      x - forwardX * radius * 1.2,
      y - forwardY * radius * 1.2,
      x - forwardX * radius * 0.55 - sideX * radius * 0.8,
      y - forwardY * radius * 0.55 - sideY * radius * 0.8,
    )
    .closePath();
  skillLayer.fill({ color: 0xd1fae5, alpha: 0.92 });
  skillLayer.stroke({ color: 0x059669, width: 2, alpha: 0.9 });
  skillLayer.circle(x + forwardX * radius * 0.15, y + forwardY * radius * 0.15, radius * 0.22);
  skillLayer.fill({ color: 0x10b981, alpha: 0.95 });
}

function drawDoctorWEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 325) * frame.scale;
  const alpha = Math.max(0.35, effectAlpha(effect));
  const pulse = 0.5 + Math.sin(performance.now() / 120) * 0.5;
  const inner = radius * (0.72 + pulse * 0.08);

  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x14b8a6, alpha: 0.07 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x2dd4bf, width: 3, alpha: 0.78 * alpha });
  skillLayer.circle(x, y, inner);
  skillLayer.stroke({ color: 0xfacc15, width: 2, alpha: 0.45 * alpha });
  for (let i = 0; i < 6; i++) {
    const angle = (Math.PI * 2 * i) / 6 + performance.now() / 420;
    const start = radius * 0.34;
    const end = radius * 0.55;
    skillLayer
      .moveTo(x + Math.cos(angle) * start, y + Math.sin(angle) * start)
      .lineTo(x + Math.cos(angle + 0.15) * end, y + Math.sin(angle + 0.15) * end);
  }
  skillLayer.stroke({ color: 0xfef08a, width: 2, alpha: 0.55 * alpha });
  skillLayer.circle(x, y, Math.max(8, radius * 0.07));
  skillLayer.fill({ color: 0xccfbf1, alpha: 0.34 * alpha });
}

function drawDoctorREffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = Math.max(24, (effect.radius || 26) * frame.scale * 1.35);
  const alpha = Math.max(0.35, effectAlpha(effect));

  for (let i = 0; i < 5; i++) {
    const drift = ((performance.now() / 850 + i * 0.2) % 1);
    const side = (i - 2) * radius * 0.22 + Math.sin(performance.now() / 260 + i) * 3;
    const px = x + side;
    const py = y + radius * 0.65 - drift * radius * 1.55;
    const size = Math.max(4, radius * 0.14) * (1 - drift * 0.3);
    const particleAlpha = alpha * (1 - drift) * 0.8;
    skillLayer
      .moveTo(px - size, py)
      .lineTo(px + size, py)
      .moveTo(px, py - size)
      .lineTo(px, py + size);
    skillLayer.stroke({ color: 0xbbf7d0, width: 3, alpha: particleAlpha });
  }
}

function drawGunnerREffect(effect, frame) {
  if (!effect.speed) {
    return;
  }
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const count = Math.max(1, effect.count || 1);
  const center = (count - 1) / 2;
  const length = Math.max(9, (effect.radius || 18) * frame.scale * 0.75);
  const spread = Math.max(3, (effect.radius || 18) * frame.scale * 0.28);
  const halfAngle = (((effect.width || 45) * Math.PI) / 180) * 0.5;
  const origin =
    effect.endX || effect.endY
      ? { x: effect.endX, y: effect.endY }
      : effectSourcePosition(effect);
  const traveled = origin
    ? Math.hypot(position.x - origin.x, position.y - origin.y)
    : 0;
  for (let i = 0; i < count; i++) {
    const bulletAngle = angle + (center ? ((i - center) / center) * halfAngle : 0);
    const forwardX = Math.cos(bulletAngle);
    const forwardY = Math.sin(bulletAngle);
    const x = frame.offsetX + (origin ? origin.x + forwardX * traveled : position.x) * frame.scale;
    const y = frame.offsetY + (origin ? origin.y + forwardY * traveled : position.y) * frame.scale;
    skillLayer.moveTo(x + forwardX * length, y + forwardY * length);
    skillLayer.lineTo(
      x + Math.cos(bulletAngle + 2.55) * spread,
      y + Math.sin(bulletAngle + 2.55) * spread,
    );
    skillLayer.lineTo(
      x + Math.cos(bulletAngle - 2.55) * spread,
      y + Math.sin(bulletAngle - 2.55) * spread,
    );
    skillLayer.closePath();
    skillLayer.fill({ color: 0xfacc15, alpha: 0.62 });
    skillLayer.stroke({ color: 0xf97316, width: 1.25, alpha: 0.72 });
  }
}

function drawGunnerEEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 300) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x38bdf8, alpha: 0.08 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x0ea5e9, width: 3, alpha: 0.78 * alpha });
}

function drawExplorerBasicEffect(effect, frame) {
  drawExplorerBoltEffect(effect, frame, 0x7dd3fc, 0xfef3c7, 0.68);
}

function drawExplorerQEffect(effect, frame) {
  drawExplorerBoltEffect(effect, frame, 0x38bdf8, 0xffffff, 0.92);
}

function drawExplorerWEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(10, (effect.radius || 80) * frame.scale * 0.45);
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const tail = radius * 2.6;
  skillLayer.moveTo(x - Math.cos(angle) * tail, y - Math.sin(angle) * tail);
  skillLayer.lineTo(x, y);
  skillLayer.stroke({ color: 0x60a5fa, width: Math.max(3, radius * 0.45), alpha: 0.5 });
  skillLayer.circle(x, y, radius * 1.35);
  skillLayer.fill({ color: 0x2563eb, alpha: 0.22 });
  skillLayer.circle(x, y, radius * 0.9);
  skillLayer.fill({ color: 0x93c5fd, alpha: 0.72 });
  skillLayer.circle(x, y, radius * 0.34);
  skillLayer.fill({ color: 0xffffff, alpha: 0.92 });
}

function drawExplorerEEffect(effect, frame) {
  drawExplorerBoltEffect(effect, frame, 0xfbbf24, 0xffffff, 0.95, 1.25);
}

function drawExplorerREffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const radius = Math.max(26, (effect.radius || 160) * frame.scale);
  drawExplorerMoonArcEffect(effect, frame, position, radius);
}

function drawExplorerMoonArcEffect(effect, frame, position, radius) {
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const outerRadius = Math.max(34, radius * 0.86);
  const innerRadius = outerRadius * 0.82;
  const outerX = x - forwardX * outerRadius * 0.18;
  const outerY = y - forwardY * outerRadius * 0.18;
  const innerX = x + forwardX * outerRadius * 0.3;
  const innerY = y + forwardY * outerRadius * 0.3;
  const spread = 1.22;
  const steps = 18;

  for (let i = 0; i <= steps; i += 1) {
    const t = angle - spread + (spread * 2 * i) / steps;
    const px = outerX + Math.cos(t) * outerRadius;
    const py = outerY + Math.sin(t) * outerRadius;
    if (i === 0) {
      skillLayer.moveTo(px, py);
    } else {
      skillLayer.lineTo(px, py);
    }
  }
  for (let i = steps; i >= 0; i -= 1) {
    const t = angle - spread + (spread * 2 * i) / steps;
    skillLayer.lineTo(innerX + Math.cos(t) * innerRadius, innerY + Math.sin(t) * innerRadius);
  }
  skillLayer.closePath();
  skillLayer.fill({ color: 0xfef3c7, alpha: 0.78 });
  skillLayer.stroke({ color: 0x38bdf8, width: Math.max(2, radius * 0.08), alpha: 0.86 });

  skillLayer.moveTo(
    outerX + Math.cos(angle - spread) * outerRadius,
    outerY + Math.sin(angle - spread) * outerRadius,
  );
  skillLayer.arc(outerX, outerY, outerRadius, angle - spread, angle + spread);
  skillLayer.stroke({ color: 0xffffff, width: Math.max(2, radius * 0.04), alpha: 0.92 });
}

function drawExplorerBoltEffect(effect, frame, shaftColor, headColor, alpha, scale = 1) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const length = Math.max(18, (effect.radius || 40) * frame.scale * 0.78 * scale);
  const width = Math.max(5, length * 0.22);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  skillLayer.moveTo(x - forwardX * length * 0.9, y - forwardY * length * 0.9);
  skillLayer.lineTo(x + forwardX * length * 0.38, y + forwardY * length * 0.38);
  skillLayer.stroke({ color: shaftColor, width, alpha: 0.34 * alpha });
  skillLayer.moveTo(x - forwardX * length * 0.52, y - forwardY * length * 0.52);
  skillLayer.lineTo(x + forwardX * length * 0.58, y + forwardY * length * 0.58);
  skillLayer.stroke({ color: shaftColor, width: Math.max(2, width * 0.42), alpha: 0.9 * alpha });
  skillLayer
    .moveTo(x + forwardX * length * 0.78, y + forwardY * length * 0.78)
    .lineTo(x - forwardX * length * 0.1 + sideX * width, y - forwardY * length * 0.1 + sideY * width)
    .lineTo(x + forwardX * length * 0.1, y + forwardY * length * 0.1)
    .lineTo(x - forwardX * length * 0.1 - sideX * width, y - forwardY * length * 0.1 - sideY * width)
    .closePath();
  skillLayer.fill({ color: headColor, alpha: 0.92 * alpha });
  skillLayer.stroke({ color: shaftColor, width: 2, alpha: 0.78 * alpha });
}

function drawRobotHookProjectile(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
  drawRobotChain(source, position, frame, 0.72);
  drawRobotHookHead(position, effect, frame);
}

function drawRobotHookPullEffect(effect, frame) {
  const alpha = effectAlpha(effect);
  const start = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
  const end = { x: effect.endX || effect.x, y: effect.endY || effect.y };
  drawRobotChain(start, end, frame, 0.85 * alpha);
  drawRobotHookHead(end, effect, frame, alpha);
}

function drawButcherRotEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 250) * frame.scale;
  const pulse = 0.5 + Math.sin(performance.now() / 130) * 0.5;

  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x365314, alpha: 0.08 + pulse * 0.025 });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x84cc16, width: 3, alpha: 0.48 + pulse * 0.16 });
  for (let index = 0; index < 10; index++) {
    const phase = (performance.now() / 1200 + index / 10) % 1;
    const angle = (Math.PI * 2 * index) / 10 + phase * 0.8;
    const distance = radius * (0.18 + phase * 0.7);
    skillLayer.circle(
      x + Math.cos(angle) * distance,
      y + Math.sin(angle) * distance - phase * 18 * frame.scale,
      Math.max(3, radius * 0.025 * (1 - phase * 0.4)),
    );
  }
  skillLayer.fill({ color: 0xa3e635, alpha: 0.22 });
}

function drawButcherMeatShieldEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = Math.max(24, ((effect.radius || 22) + 14) * frame.scale);
  const pulse = 0.5 + Math.sin(performance.now() / 170) * 0.5;
  const alpha = effectAlpha(effect);

  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x7f1d1d, alpha: (0.1 + pulse * 0.04) * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xfca5a5, width: 4, alpha: (0.62 + pulse * 0.18) * alpha });
  skillLayer.circle(x, y, radius * 0.72);
  skillLayer.stroke({ color: 0x991b1b, width: 2, alpha: 0.7 * alpha });
}

function drawButcherDismemberEffect(effect, frame) {
  const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
  const target = { x: effect.endX || effect.x, y: effect.endY || effect.y };
  const alpha = effectAlpha(effect);
  drawRobotChain(source, target, frame, alpha);
  const x = frame.offsetX + target.x * frame.scale;
  const y = frame.offsetY + target.y * frame.scale;
  const radius = Math.max(22, ((effect.radius || 22) + 10) * frame.scale);
  const pulse = 0.5 + Math.sin(performance.now() / 110) * 0.5;
  skillLayer.circle(x, y, radius * (0.9 + pulse * 0.08));
  skillLayer.fill({ color: 0x7f1d1d, alpha: 0.14 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xef4444, width: 4, alpha: 0.8 * alpha });
}

function drawRobotChain(start, end, frame, alpha) {
  const sx = frame.offsetX + start.x * frame.scale;
  const sy = frame.offsetY + start.y * frame.scale;
  const ex = frame.offsetX + end.x * frame.scale;
  const ey = frame.offsetY + end.y * frame.scale;
  const dx = ex - sx;
  const dy = ey - sy;
  const length = Math.hypot(dx, dy);
  if (length < 1) {
    return;
  }
  const ux = dx / length;
  const uy = dy / length;
  const nx = -uy;
  const ny = ux;
  skillLayer.moveTo(sx, sy);
  skillLayer.lineTo(ex, ey);
  skillLayer.stroke({ color: 0x94a3b8, width: 5, alpha: 0.28 * alpha });
  const step = Math.max(10, 16 * frame.scale);
  const link = Math.max(4, 5 * frame.scale);
  for (let d = step; d < length - step * 0.5; d += step) {
    const cx = sx + ux * d;
    const cy = sy + uy * d;
    skillLayer.moveTo(cx - nx * link, cy - ny * link);
    skillLayer.lineTo(cx + nx * link, cy + ny * link);
  }
  skillLayer.stroke({ color: 0xe5e7eb, width: 2, alpha: 0.85 * alpha });
}

function drawRobotHookHead(position, effect, frame, alpha = 1) {
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const forwardX = Math.cos(angle);
  const forwardY = Math.sin(angle);
  const sideX = -forwardY;
  const sideY = forwardX;
  const size = Math.max(9, (effect.radius || 70) * frame.scale * 0.22);
  const back = size * 0.65;
  const tipX = x + forwardX * size;
  const tipY = y + forwardY * size;
  const baseX = x - forwardX * back;
  const baseY = y - forwardY * back;
  skillLayer.moveTo(tipX, tipY);
  skillLayer.lineTo(baseX + sideX * size * 0.72, baseY + sideY * size * 0.72);
  skillLayer.lineTo(baseX + sideX * size * 0.22, baseY + sideY * size * 0.1);
  skillLayer.lineTo(baseX, baseY);
  skillLayer.lineTo(baseX - sideX * size * 0.22, baseY - sideY * size * 0.1);
  skillLayer.lineTo(baseX - sideX * size * 0.72, baseY - sideY * size * 0.72);
  skillLayer.closePath();
  skillLayer.fill({ color: 0xd1d5db, alpha: 0.96 * alpha });
  skillLayer.stroke({ color: 0x475569, width: 2, alpha: 0.9 * alpha });
  skillLayer.circle(x, y, size * 0.32);
  skillLayer.fill({ color: 0x38bdf8, alpha: 0.72 * alpha });
}

function drawRobotRRangeEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || effect.range || 600) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x38bdf8, alpha: 0.09 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xe0f2fe, width: 5, alpha: 0.38 * alpha });
  skillLayer.circle(x, y, radius * 0.72);
  skillLayer.stroke({ color: 0x0ea5e9, width: 2, alpha: 0.68 * alpha });
  skillLayer.circle(x, y, Math.max(10, radius * 0.08));
  skillLayer.fill({ color: 0xf8fafc, alpha: 0.32 * alpha });
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

function drawSiegeCannonballEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(5, (effect.radius || 14) * frame.scale * 0.55);
  const tailX = x - (effect.dirX || 1) * radius * 2.2;
  const tailY = y - (effect.dirY || 0) * radius * 2.2;
  skillLayer.moveTo(tailX, tailY);
  skillLayer.lineTo(x, y);
  skillLayer.stroke({ color: 0x475569, width: Math.max(3, radius * 0.8), alpha: 0.45 });
  skillLayer.circle(x, y, radius * 1.45);
  skillLayer.fill({ color: 0x111827, alpha: 0.24 });
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x334155, alpha: 0.96 });
  skillLayer.circle(x - radius * 0.25, y - radius * 0.25, radius * 0.28);
  skillLayer.fill({ color: 0xe5e7eb, alpha: 0.45 });
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

function drawNinjaShurikenEffect(effect, frame) {
  const position = projectileDrawPosition(effect, { fromSnapshot: true });
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
  const size = Math.max(8, (effect.radius || 35) * frame.scale * 0.38);
  const tail = Math.max(20, size * 3.2);
  skillLayer
    .moveTo(x - Math.cos(angle) * tail, y - Math.sin(angle) * tail)
    .lineTo(x, y);
  skillLayer.stroke({ color: 0x64748b, width: Math.max(2, size * 0.55), alpha: 0.45 });
  for (let i = 0; i < 4; i++) {
    const bladeAngle = angle + Math.PI / 4 + (Math.PI / 2) * i;
    skillLayer
      .moveTo(x, y)
      .lineTo(x + Math.cos(bladeAngle) * size, y + Math.sin(bladeAngle) * size);
  }
  skillLayer.stroke({ color: 0xe5e7eb, width: Math.max(2, size * 0.35), alpha: 0.95 });
  skillLayer.circle(x, y, Math.max(2, size * 0.24));
  skillLayer.fill({ color: 0x111827, alpha: 0.95 });
}

function drawNinjaERangeEffect(effect, frame) {
  const x = frame.offsetX + effect.x * frame.scale;
  const y = frame.offsetY + effect.y * frame.scale;
  const radius = (effect.radius || 290) * frame.scale;
  const alpha = effectAlpha(effect);
  skillLayer.circle(x, y, radius);
  skillLayer.fill({ color: 0x8b5cf6, alpha: 0.08 * alpha });
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0xc4b5fd, width: 3, alpha: 0.75 * alpha });
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

function drawNinjaShadowEffect(effect, frame) {
  const position = movingEffectPosition(effect);
  const x = frame.offsetX + position.x * frame.scale;
  const y = frame.offsetY + position.y * frame.scale;
  const radius = Math.max(12, (effect.radius || 16) * frame.scale);
  skillLayer.circle(x, y, radius * 1.25);
  skillLayer.fill({ color: 0x111827, alpha: 0.22 });
  skillLayer.circle(x, y, radius * 0.9);
  skillLayer.fill({ color: 0x1f2937, alpha: 0.82 });
  skillLayer.circle(x, y, radius * 1.35);
  skillLayer.stroke({ color: 0x8b5cf6, width: 2, alpha: 0.72 });
  skillLayer
    .moveTo(x, y - radius * 1.35)
    .lineTo(x + radius * 0.95, y)
    .lineTo(x, y + radius * 1.35)
    .lineTo(x - radius * 0.95, y)
    .closePath();
  skillLayer.stroke({ color: 0xc4b5fd, width: 2, alpha: 0.55 });
  drawNinjaShadowTimer(effect, x, y, radius * 1.75);
}

function drawNinjaShadowTimer(effect, x, y, radius) {
  const tick = interpolatedTick();
  const remainingTicks = (effect.expiresAt || 0) - tick;
  if (remainingTicks <= 0) {
    return;
  }
  const durationTicks = Math.max(1, (effect.expiresAt || 0) - (effect.createdAt || 0));
  const progress = ratio(remainingTicks, durationTicks);
  const startAngle = -Math.PI / 2;
  const endAngle = startAngle + Math.PI * 2 * progress;
  const startX = x + Math.cos(startAngle) * radius;
  const startY = y + Math.sin(startAngle) * radius;
  skillLayer.circle(x, y, radius);
  skillLayer.stroke({ color: 0x64748b, width: 4, alpha: 0.35 });
  skillLayer.moveTo(startX, startY);
  skillLayer.arc(x, y, radius, startAngle, endAngle);
  skillLayer.stroke({ color: 0xc4b5fd, width: 4, alpha: 0.9 });
}

function movingEffectPosition(effect) {
  const startX = effect.x || 0;
  const startY = effect.y || 0;
  const endX = effect.endX ?? startX;
  const endY = effect.endY ?? startY;
  const dx = endX - startX;
  const dy = endY - startY;
  const length = Math.hypot(dx, dy);
  if (!effect.speed || length <= 0) {
    return { x: endX, y: endY };
  }
  const traveled = clamp(
    Math.max(0, interpolatedTick() - (effect.createdAt || 0)) * effect.speed,
    0,
    length,
  );
  return {
    x: startX + (dx / length) * traveled,
    y: startY + (dy / length) * traveled,
  };
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

function effectSourcePosition(effect) {
  if (!effect?.sourceId) {
    return null;
  }
  for (const player of state.players.values()) {
    if (player.id === effect.sourceId) {
      return player;
    }
  }
  return state.units.get(effect.sourceId) || null;
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

function drawNinjaPassiveCooldowns(frame) {
  const self = state.players.get(state.playerId);
  if (!self || self.heroId !== "ninja") {
    return;
  }
  const targetUntil = self.passive?.ninjaSoulCooldowns || {};
  const tick = interpolatedTick();
  const targets = targetMap();
  const cooldownTicks =
    (skillClientConfig.ninja_passive?.heroCooldownSeconds || 10) *
    state.tickRate;
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
    const radius = targetSelectRadius(target, frame) + 9;
    const progress = ratio(remainingTicks, cooldownTicks);
    const startAngle = -Math.PI / 2;
    const endAngle = startAngle + Math.PI * 2 * progress;
    const startX = x + Math.cos(startAngle) * radius;
    const startY = y + Math.sin(startAngle) * radius;
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0x6b7280, width: 4, alpha: 0.25 });
    skillLayer.moveTo(startX, startY);
    skillLayer.arc(x, y, radius, startAngle, endAngle);
    skillLayer.stroke({ color: 0xa855f7, width: 4, alpha: 0.9 });
  }
}

function drawFireMageBlazeExplosions(frame) {
  const tick = interpolatedTick();
  const durationTicks =
    (skillClientConfig.fire_mage_passive?.explosionDelaySeconds || 2) *
    state.tickRate;
  for (const target of targetMap().values()) {
    if (!target || target.dead) {
      continue;
    }
    const burn = (target.buffs || []).find(
      (buff) =>
        buff.id?.startsWith("fire_mage_blaze:") &&
        (buff.explosionAtTick || 0) > tick,
    );
    if (!burn) {
      continue;
    }
    const remainingTicks = burn.explosionAtTick - tick;
    const progress = ratio(remainingTicks, durationTicks);
    const x = frame.offsetX + target.x * frame.scale;
    const y = frame.offsetY + target.y * frame.scale;
    const radius = targetSelectRadius(target, frame) + 10;
    const startAngle = -Math.PI / 2;
    const endAngle = startAngle + Math.PI * 2 * progress;
    const startX = x + Math.cos(startAngle) * radius;
    const startY = y + Math.sin(startAngle) * radius;
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0x7f1d1d, width: 4, alpha: 0.28 });
    skillLayer.moveTo(startX, startY);
    skillLayer.arc(x, y, radius, startAngle, endAngle);
    skillLayer.stroke({ color: 0xef4444, width: 4, alpha: 0.9 });
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
