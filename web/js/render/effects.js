function drawEffects(frame) {
    skillLayer.clear();
    drawActiveSkillRanges(frame);
    drawSwordETargetCooldowns(frame);
    drawNinjaPassiveCooldowns(frame);
    drawFireMageBlazeExplosions(frame);
    drawCastWindups(frame);
    drawSkillPreview(frame);
    drawShadowAssassinQReadyEffects(frame);

    const visibleServants = new Set();
    for (const effect of state.effects) {
        if (effect.kind === 'crossbowman_final_hour') {
            drawCrossbowmanFinalHourEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'crossbowman_tumble') {
            drawCrossbowmanTumbleEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'crossbowman_silver_bolts') {
            drawCrossbowmanSilverBoltsEffect(effect, frame, false);
            continue;
        }
        if (effect.kind === 'crossbowman_silver_bolts_proc') {
            drawCrossbowmanSilverBoltsEffect(effect, frame, true);
            continue;
        }
        if (effect.kind === 'crossbowman_condemn') {
            drawCrossbowmanCondemnProjectile(effect, frame);
            continue;
        }
        if (effect.kind === 'crossbowman_condemn_knockback') {
            drawCrossbowmanCondemnKnockback(effect, frame);
            continue;
        }
        if (effect.kind === 'crossbowman_condemn_impact') {
            drawCrossbowmanCondemnImpact(effect, frame);
            continue;
        }
        if (effect.kind === 'warrior_q_light') {
            drawWarriorQLightEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'warrior_w_shields') {
            drawWarriorWShieldsEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'warrior_r_sword') {
            drawWarriorRSwordEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'sword_whirlwind') {
            drawSwordWhirlwindEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'sword_q') {
            drawSwordQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'sword_q_circle') {
            drawSwordQCircleEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'sword_e') {
            drawSwordEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'sword_r') {
            drawSwordREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'blade_q_heal') {
            drawBladeQHealEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'blade_e_whirlwind') {
            drawBladeEWhirlwindEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'blade_r_rage') {
            drawBladeRRageEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_q') {
            drawMonkQSonicWaveEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_q_mark') {
            drawMonkQMarkEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_q_echo') {
            drawMonkQEchoEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_w_safeguard') {
            drawMonkWSafeguardEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_w_iron_will') {
            drawMonkWIronWillEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_e_tempest') {
            drawMonkETempestEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_e_reveal') {
            drawMonkERevealEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_q_cast') {
            drawShadowAssassinQCastEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_q_hit') {
            drawShadowAssassinQHitEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_q_reveal') {
            drawShadowAssassinQRevealEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_w') {
            drawShadowAssassinWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_e') {
            drawShadowAssassinEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_e_mark') {
            drawShadowAssassinEMarkEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'shadow_assassin_r') {
            drawShadowAssassinREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'monk_e_cripple') {
            drawMonkECrippleEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'berserker_q_range') {
            drawBerserkerQRangeEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'berserker_q') {
            drawBerserkerQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'berserker_w') {
            drawBerserkerWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'berserker_e') {
            drawBerserkerEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'berserker_r') {
            drawBerserkerREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'tank_q') {
            drawTankShardEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'tank_w_aftershock') {
            drawTankAftershockEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'tank_r_impact') {
            drawTankImpactEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'basic_arrow') {
            drawBasicArrowEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'siege_cannonball') {
            drawSiegeCannonballEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'fire_mage_q') {
            drawFireMageQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_q') {
            drawKillerQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_q_cast_range') {
            drawKillerQCastRangeEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_q_dagger_airborne') {
            drawKillerQAirborneDaggerEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_w_dagger_airborne') {
            drawKillerWAirborneDaggerEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_q_dagger' || effect.kind === 'killer_w_dagger') {
            drawKillerQDaggerEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_dagger_slash') {
            drawKillerDaggerSlashEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_e') {
            drawKillerEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_r_channel') {
            drawKillerRChannelEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'killer_r') {
            drawKillerRProjectileEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'fire_mage_w_range') {
            drawFireMageWRangeEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'fire_mage_w') {
            drawFireMageWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'fire_mage_e') {
            drawFireMageEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'fire_mage_r') {
            drawFireMageREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'doctor_q') {
            drawDoctorQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'doctor_w') {
            drawDoctorWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'doctor_r') {
            drawDoctorREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'frostmage_w') {
            drawFrostMageWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'frostmage_q' || effect.kind === 'frostmage_q_shard') {
            drawFrostMageQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'frostmage_e') {
            drawFrostMageEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'frostmage_r_enemy' || effect.kind === 'frostmage_r_self') {
            drawFrostMageREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'frostmage_servant') {
            visibleServants.add(servantEffectID(effect));
            drawFrostMageServantEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'robot_q') {
            drawRobotHookProjectile(effect, frame);
            continue;
        }
        if (effect.kind === 'butcher_q') {
            drawRobotHookProjectile(effect, frame);
            continue;
        }
        if (effect.kind === 'robot_q_pull') {
            drawRobotHookPullEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'butcher_q_pull') {
            drawRobotHookPullEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'butcher_w') {
            drawButcherRotEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'butcher_e') {
            drawButcherMeatShieldEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'butcher_r') {
            drawButcherDismemberEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'robot_r') {
            drawRobotRRangeEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'gunner_q') {
            drawGunnerQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'gunner_q_muzzle') {
            drawGunnerQMuzzleEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'gunner_w') {
            drawGunnerWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'gunner_r') {
            drawGunnerREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'gunner_e') {
            drawGunnerEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'explorer_q') {
            drawExplorerQEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'explorer_w') {
            drawExplorerWEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'explorer_e') {
            drawExplorerEEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'explorer_r') {
            drawExplorerREffect(effect, frame);
            continue;
        }
        if (effect.kind === 'archer_volley_arrow') {
            drawVolleyArrowEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'archer_hawk') {
            drawArcherHawkEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'archer_crystal_arrow') {
            drawCrystalArrowEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'mage_light_binding') {
            drawMageLightBindingEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'ninja_shuriken') {
            drawNinjaShurikenEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'ninja_e') {
            drawNinjaERangeEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'mage_prismatic_barrier') {
            drawMagePrismaticBarrierEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'mage_lucent_singularity_orb') {
            drawMageLucentSingularityOrbEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'mage_lucent_singularity') {
            drawMageLucentSingularityEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'mage_lucent_singularity_burst') {
            drawMageLucentSingularityBurstEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'mage_final_spark') {
            drawMageFinalSparkEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'fountain_shot') {
            drawFountainShotEffect(effect, frame);
            continue;
        }
        if (effect.kind === 'ninja_shadow') {
            drawNinjaShadowEffect(effect, frame);
            continue;
        }
        if (effect.kind !== 'wind_wall') {
            continue;
        }
        drawWindWallEffect(effect, frame);
    }
    for (const id of state.servantEffectPositions.keys()) {
        if (!visibleServants.has(id)) {
            state.servantEffectPositions.delete(id);
        }
    }
}

function drawCrossbowmanFinalHourEffect(effect, frame) {
    const source = effectSourcePosition(effect) || effect;
    const x = frame.offsetX + source.x * frame.scale;
    const y = frame.offsetY + source.y * frame.scale;
    const baseRadius = Math.max(24, ((source.radius || effect.radius || 16) + 18) * frame.scale);
    const fade = Math.min(1, effectAlpha(effect) * 8);
    const pulse = 0.5 + 0.5 * Math.sin(performance.now() / 170);
    const rotation = performance.now() / 850;

    skillLayer.circle(x, y, baseRadius * (1.05 + pulse * 0.08));
    skillLayer.fill({ color: 0x111827, alpha: 0.16 * fade });
    skillLayer.stroke({ color: 0xdc2626, width: Math.max(2, 3 * frame.scale), alpha: (0.56 + pulse * 0.18) * fade });
    for (let index = 0; index < 4; index++) {
        const angle = rotation + (Math.PI * 2 * index) / 4;
        const inner = baseRadius * 0.72;
        const outer = baseRadius * 1.28;
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
        skillLayer.stroke({
            color: index % 2 ? 0xf8fafc : 0x991b1b,
            width: Math.max(2, 3 * frame.scale),
            alpha: 0.72 * fade,
        });
    }
}

function drawCrossbowmanTumbleEffect(effect, frame) {
    const sx = frame.offsetX + effect.x * frame.scale;
    const sy = frame.offsetY + effect.y * frame.scale;
    const tx = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const ty = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const dx = tx - sx;
    const dy = ty - sy;
    const length = Math.hypot(dx, dy) || 1;
    const nx = -dy / length;
    const ny = dx / length;
    const alpha = Math.min(1, effectAlpha(effect) * 2.2);
    const progress = 1 - effectAlpha(effect);

    for (let index = 0; index < 7; index++) {
        const t = clamp(index / 6 + progress * 0.12, 0, 1);
        const offset = (index % 2 ? 1 : -1) * (4 + index) * frame.scale;
        const radius = Math.max(2, (9 - index * 0.8) * frame.scale);
        skillLayer.circle(sx + dx * t + nx * offset, sy + dy * t + ny * offset, radius);
        skillLayer.fill({ color: index % 2 ? 0x94a3b8 : 0xd6d3d1, alpha: 0.26 * alpha * (1 - t * 0.45) });
    }
    skillLayer.moveTo(sx, sy);
    skillLayer.lineTo(tx, ty);
    skillLayer.stroke({ color: 0x64748b, width: Math.max(2, 5 * frame.scale), alpha: 0.34 * alpha });
    skillLayer.circle(tx, ty, Math.max(8, (effect.radius || 16) * frame.scale * (0.7 + progress * 0.5)));
    skillLayer.stroke({ color: 0xe2e8f0, width: Math.max(2, 3 * frame.scale), alpha: 0.7 * alpha });
}

function drawCrossbowmanSilverBoltsEffect(effect, frame, proc) {
    const x = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const y = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const baseRadius = Math.max(18, ((effect.radius || 18) + 16) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * (proc ? 2.4 : 10));
    const progress = 1 - effectAlpha(effect);
    const marks = proc ? 3 : Math.max(1, Math.min(2, effect.count || 1));
    const rotation = -Math.PI / 2 + progress * (proc ? 1.1 : 0.3);

    skillLayer.circle(x, y, baseRadius * (proc ? 1.1 + progress * 0.9 : 1));
    skillLayer.stroke({
        color: proc ? 0xffffff : 0xcbd5e1,
        width: Math.max(2, (proc ? 4 : 2.5) * frame.scale),
        alpha: (proc ? 0.82 : 0.68) * alpha,
    });

    for (let index = 0; index < marks; index++) {
        const angle = rotation + (Math.PI * 2 * index) / 3;
        const inner = baseRadius * (proc ? 0.2 : 0.58);
        const outer = baseRadius * (proc ? 1.45 : 1.08);
        const sx = x + Math.cos(angle) * inner;
        const sy = y + Math.sin(angle) * inner;
        const ex = x + Math.cos(angle) * outer;
        const ey = y + Math.sin(angle) * outer;
        skillLayer.moveTo(sx, sy);
        skillLayer.lineTo(ex, ey);
        skillLayer.stroke({
            color: index % 2 ? 0xe2e8f0 : 0x94a3b8,
            width: Math.max(3, 5 * frame.scale),
            alpha: 0.78 * alpha,
        });
        const wing = Math.max(4, 7 * frame.scale);
        skillLayer.moveTo(ex, ey);
        skillLayer.lineTo(ex - Math.cos(angle - 0.55) * wing, ey - Math.sin(angle - 0.55) * wing);
        skillLayer.moveTo(ex, ey);
        skillLayer.lineTo(ex - Math.cos(angle + 0.55) * wing, ey - Math.sin(angle + 0.55) * wing);
        skillLayer.stroke({ color: 0xf8fafc, width: Math.max(2, 3 * frame.scale), alpha: 0.9 * alpha });
    }

    if (proc) {
        skillLayer.circle(x, y, baseRadius * (0.25 + progress * 0.45));
        skillLayer.fill({ color: 0xffffff, alpha: 0.38 * alpha });
        skillLayer.circle(x, y, baseRadius * (0.7 + progress * 1.4));
        skillLayer.stroke({ color: 0x94a3b8, width: Math.max(2, 3 * frame.scale), alpha: 0.5 * alpha });
    }
}

function drawCrossbowmanCondemnProjectile(effect, frame) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const normalX = -dirY;
    const normalY = dirX;
    const length = Math.max(24, 38 * frame.scale);
    const width = Math.max(5, 8 * frame.scale);
    const tailX = x - dirX * length;
    const tailY = y - dirY * length;

    skillLayer.moveTo(tailX, tailY);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0x7f1d1d, width: width * 2.2, alpha: 0.22 });
    skillLayer.moveTo(tailX, tailY);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0xf8fafc, width, alpha: 0.92 });
    skillLayer.moveTo(tailX + normalX * width, tailY + normalY * width);
    skillLayer.lineTo(x, y);
    skillLayer.moveTo(tailX - normalX * width, tailY - normalY * width);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0xdc2626, width: Math.max(2, 2.5 * frame.scale), alpha: 0.78 });
    skillLayer.circle(x, y, Math.max(4, 6 * frame.scale));
    skillLayer.fill({ color: 0xffffff, alpha: 0.9 });
}

function drawCrossbowmanCondemnKnockback(effect, frame) {
    const sx = frame.offsetX + effect.x * frame.scale;
    const sy = frame.offsetY + effect.y * frame.scale;
    const ex = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const ey = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const dx = ex - sx;
    const dy = ey - sy;
    const length = Math.hypot(dx, dy) || 1;
    const dirX = dx / length;
    const dirY = dy / length;
    const normalX = -dirY;
    const normalY = dirX;
    const alpha = Math.min(1, effectAlpha(effect) * 3);

    for (let index = 0; index < 3; index++) {
        const side = (index - 1) * Math.max(5, 8 * frame.scale);
        skillLayer.moveTo(sx + normalX * side, sy + normalY * side);
        skillLayer.lineTo(ex + normalX * side * 0.25, ey + normalY * side * 0.25);
        skillLayer.stroke({
            color: index === 1 ? 0xf8fafc : 0x991b1b,
            width: Math.max(2, (4 - index * 0.5) * frame.scale),
            alpha: (0.58 - index * 0.08) * alpha,
        });
    }
    const arrowSize = Math.max(10, 16 * frame.scale);
    skillLayer.moveTo(ex, ey);
    skillLayer.lineTo(
        ex - dirX * arrowSize + normalX * arrowSize * 0.55,
        ey - dirY * arrowSize + normalY * arrowSize * 0.55,
    );
    skillLayer.moveTo(ex, ey);
    skillLayer.lineTo(
        ex - dirX * arrowSize - normalX * arrowSize * 0.55,
        ey - dirY * arrowSize - normalY * arrowSize * 0.55,
    );
    skillLayer.stroke({ color: 0xe2e8f0, width: Math.max(2, 3 * frame.scale), alpha: 0.8 * alpha });
}

function drawCrossbowmanCondemnImpact(effect, frame) {
    const x = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const y = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const radius = Math.max(22, ((effect.radius || 16) + 24) * frame.scale);
    const remaining = effectAlpha(effect);
    const progress = 1 - remaining;
    const alpha = Math.min(1, remaining * 3.5);

    skillLayer.circle(x, y, radius * (0.55 + progress * 1.3));
    skillLayer.stroke({ color: 0xffffff, width: Math.max(3, 5 * frame.scale), alpha: 0.72 * alpha });
    skillLayer.circle(x, y, radius * (0.35 + progress * 0.75));
    skillLayer.fill({ color: 0x991b1b, alpha: 0.2 * alpha });
    for (let index = 0; index < 8; index++) {
        const angle = (Math.PI * 2 * index) / 8 + progress * 0.35;
        const inner = radius * 0.22;
        const outer = radius * (0.7 + (index % 2) * 0.38);
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
        skillLayer.stroke({
            color: index % 2 ? 0xf8fafc : 0xdc2626,
            width: Math.max(2, 3 * frame.scale),
            alpha: 0.7 * alpha,
        });
    }
}

/**
 * 绘制由多层横向风带、翻卷气流和漂浮风尘组成的剑客风墙。
 * @param {object} effect 服务端同步的风墙中心、方向、宽度与持续时间。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawWindWallEffect(effect, frame) {
    const half = effect.width / 2;
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const nx = -dirY;
    const ny = dirX;
    const startX = frame.offsetX + (effect.x - dirX * half) * frame.scale;
    const startY = frame.offsetY + (effect.y - dirY * half) * frame.scale;
    const endX = frame.offsetX + (effect.x + dirX * half) * frame.scale;
    const endY = frame.offsetY + (effect.y + dirY * half) * frame.scale;
    const remainingTicks = Math.max(0, (effect.expiresAt || 0) - interpolatedTick());
    const alpha = effect.createdAt
        ? Math.min(1, effectAlpha(effect) * 3.4)
        : Math.min(1, remainingTicks / Math.max(1, state.tickRate * 0.3));
    const time = performance.now();
    const wallDepth = Math.max(30, 56 * frame.scale);
    const length = Math.hypot(endX - startX, endY - startY) || 1;

    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x7dd3fc, width: wallDepth * 1.22, alpha: 0.16 * alpha });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0xf0f9ff, width: wallDepth * 0.42, alpha: 0.18 * alpha });

    for (let i = -3; i <= 3; i++) {
        const layer = i + 3;
        const sway = Math.sin(time / (115 + layer * 24) + layer * 1.7) * wallDepth * 0.18;
        const offset = i * wallDepth * 0.19 + sway;
        skillLayer.moveTo(startX + nx * offset, startY + ny * offset);
        skillLayer.lineTo(endX + nx * offset, endY + ny * offset);
        skillLayer.stroke({
            color: layer === 3 ? 0xffffff : layer % 2 ? 0x7dd3fc : 0x22d3ee,
            width: Math.max(3, wallDepth * (layer === 3 ? 0.2 : 0.12)),
            alpha: (layer === 3 ? 0.9 : 0.48) * alpha,
        });
    }

    skillLayer.moveTo(startX - nx * wallDepth * 0.55, startY - ny * wallDepth * 0.55);
    skillLayer.lineTo(endX - nx * wallDepth * 0.55, endY - ny * wallDepth * 0.55);
    skillLayer.stroke({ color: 0x0ea5e9, width: 4, alpha: 0.72 * alpha });
    skillLayer.moveTo(startX + nx * wallDepth * 0.55, startY + ny * wallDepth * 0.55);
    skillLayer.lineTo(endX + nx * wallDepth * 0.55, endY + ny * wallDepth * 0.55);
    skillLayer.stroke({ color: 0xf0f9ff, width: 4, alpha: 0.66 * alpha });

    for (let i = 0; i < 18; i++) {
        const t = (i / 18 + ((time / 780) % 1)) % 1;
        const wave = Math.sin(t * Math.PI * 4 + time / 120 + i) * wallDepth * 0.38;
        const side = ((i % 3) - 1) * wallDepth * 0.28 + wave;
        const px = startX + (endX - startX) * t + nx * side;
        const py = startY + (endY - startY) * t + ny * side;
        const streak = Math.max(8, length * 0.035);
        skillLayer.moveTo(px - dirX * streak * 0.5, py - dirY * streak * 0.5);
        skillLayer.lineTo(px + dirX * streak * 0.5, py + dirY * streak * 0.5);
        skillLayer.stroke({ color: i % 2 ? 0xffffff : 0x7dd3fc, width: 3, alpha: 0.58 * alpha });
        skillLayer.circle(px, py, 3 + (i % 3));
        skillLayer.fill({ color: 0xf8fafc, alpha: 0.42 * alpha });
    }

    for (let i = 0; i < 6; i++) {
        const t = (i + 0.5) / 6;
        const px = startX + (endX - startX) * t;
        const py = startY + (endY - startY) * t;
        const curlRadius = wallDepth * (0.2 + (i % 2) * 0.08);
        const curlStart = time / 180 + i * 1.4;
        skillLayer.moveTo(px + Math.cos(curlStart) * curlRadius, py + Math.sin(curlStart) * curlRadius);
        skillLayer.arc(px, py, curlRadius, curlStart, curlStart + Math.PI * 1.55);
        skillLayer.stroke({ color: i % 2 ? 0xffffff : 0xbae6fd, width: 3, alpha: 0.48 * alpha });
    }
}

function drawBerserkerQEffect(effect, frame) {
    const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
    const x = frame.offsetX + source.x * frame.scale;
    const y = frame.offsetY + source.y * frame.scale;
    const outer = Math.max(42, (effect.range || 425) * frame.scale);
    const inner = Math.max(24, (effect.radius || 300) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * 2.6);
    const progress = 1 - effectAlpha(effect);
    const rotation = performance.now() / 130;

    skillLayer.circle(x, y, outer);
    skillLayer.fill({ color: 0x7f1d1d, alpha: 0.07 * alpha });
    skillLayer.circle(x, y, outer);
    skillLayer.stroke({ color: 0xef4444, width: Math.max(2, outer * 0.012), alpha: 0.58 * alpha });
    skillLayer.circle(x, y, inner);
    skillLayer.stroke({ color: 0xfef2f2, width: Math.max(2, inner * 0.009), alpha: 0.34 * alpha });
    for (let i = 0; i < 4; i++) {
        const start = rotation + (Math.PI * 2 * i) / 4 + progress * 0.9;
        const radius = inner + (outer - inner) * (0.45 + (i % 2) * 0.34);
        skillLayer.moveTo(x + Math.cos(start) * radius * 0.55, y + Math.sin(start) * radius * 0.55);
        skillLayer.arc(x, y, radius, start, start + Math.PI * 0.92);
        skillLayer.stroke({
            color: i % 2 ? 0xfef2f2 : 0xef4444,
            width: Math.max(5, outer * (i % 2 ? 0.03 : 0.045)),
            alpha: (0.9 - i * 0.09) * alpha,
        });
    }
    skillLayer.circle(x, y, inner * (0.42 + progress * 0.28));
    skillLayer.fill({ color: 0x7f1d1d, alpha: 0.12 * alpha });
    if ((effect.count || 0) > 0) {
        skillLayer.circle(x, y, Math.max(14, inner * 0.12));
        skillLayer.fill({ color: 0x22c55e, alpha: 0.26 * alpha });
    }
}

function drawBerserkerQRangeEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const inner = (effect.radius || 300) * frame.scale;
    const outer = (effect.range || 425) * frame.scale;
    const alpha = Math.min(1, effectAlpha(effect) * 1.8);
    skillLayer.circle(x, y, outer);
    skillLayer.fill({ color: 0xf97316, alpha: 0.08 * alpha });
    skillLayer.circle(x, y, inner);
    skillLayer.fill({ color: 0xdc2626, alpha: 0.11 * alpha });
    skillLayer.circle(x, y, inner);
    skillLayer.stroke({ color: 0xdc2626, width: 2, alpha: 0.72 * alpha });
    skillLayer.circle(x, y, outer);
    skillLayer.stroke({ color: 0xf59e0b, width: 3, alpha: 0.82 * alpha });
}

function drawBerserkerWEffect(effect, frame) {
    const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
    const sx = frame.offsetX + source.x * frame.scale;
    const sy = frame.offsetY + source.y * frame.scale;
    const tx = frame.offsetX + (effect.endX || effect.x) * frame.scale;
    const ty = frame.offsetY + (effect.endY || effect.y) * frame.scale;
    const dx = tx - sx;
    const dy = ty - sy;
    const len = Math.hypot(dx, dy) || 1;
    const ux = dx / len;
    const uy = dy / len;
    const nx = -uy;
    const ny = ux;
    const alpha = Math.min(1, effectAlpha(effect) * 2.8);
    const radius = Math.max(18, ((effect.radius || 20) + 18) * frame.scale);

    skillLayer.moveTo(sx + nx * 8, sy + ny * 8);
    skillLayer.lineTo(tx + nx * radius * 0.4, ty + ny * radius * 0.4);
    skillLayer.stroke({ color: 0xfef2f2, width: Math.max(4, radius * 0.26), alpha: 0.82 * alpha });
    skillLayer.moveTo(sx - nx * 10, sy - ny * 10);
    skillLayer.lineTo(tx - nx * radius * 0.25, ty - ny * radius * 0.25);
    skillLayer.stroke({ color: 0xdc2626, width: Math.max(5, radius * 0.34), alpha: 0.5 * alpha });
    skillLayer.circle(tx, ty, radius * 0.82);
    skillLayer.fill({ color: 0x7f1d1d, alpha: 0.18 * alpha });
    skillLayer.circle(tx + ux * radius * 0.2, ty + uy * radius * 0.2, radius * 0.32);
    skillLayer.fill({ color: 0xfca5a5, alpha: 0.34 * alpha });
}

function drawBerserkerEEffect(effect, frame) {
    const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
    const x = frame.offsetX + source.x * frame.scale;
    const y = frame.offsetY + source.y * frame.scale;
    const dir = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const range = (effect.range || 535) * frame.scale;
    const spread = ((effect.width || 50) * Math.PI) / 360;
    const alpha = Math.min(1, effectAlpha(effect) * 2.6);
    const progress = 1 - effectAlpha(effect);

    for (let i = -1; i <= 1; i++) {
        const angle = dir + i * spread;
        const pull = range * (0.35 + 0.5 * (1 - progress));
        const hookX = x + Math.cos(angle) * pull;
        const hookY = y + Math.sin(angle) * pull;
        skillLayer.moveTo(x, y);
        skillLayer.lineTo(hookX, hookY);
        skillLayer.stroke({ color: 0x475569, width: 6, alpha: 0.42 * alpha });
        skillLayer.moveTo(x, y);
        skillLayer.lineTo(hookX, hookY);
        skillLayer.stroke({ color: 0xe5e7eb, width: 2, alpha: 0.78 * alpha });
        skillLayer.moveTo(hookX, hookY);
        skillLayer.lineTo(hookX - Math.cos(angle - 0.7) * range * 0.08, hookY - Math.sin(angle - 0.7) * range * 0.08);
        skillLayer.lineTo(hookX - Math.cos(angle + 0.7) * range * 0.08, hookY - Math.sin(angle + 0.7) * range * 0.08);
        skillLayer.stroke({ color: 0xdc2626, width: 4, alpha: 0.75 * alpha });
    }
    skillLayer.circle(x, y, Math.max(10, 34 * frame.scale));
    skillLayer.fill({ color: 0x7f1d1d, alpha: (effect.count || 0) > 0 ? 0.26 * alpha : 0.12 * alpha });
}

function drawBerserkerREffect(effect, frame) {
    const sx = frame.offsetX + effect.x * frame.scale;
    const sy = frame.offsetY + effect.y * frame.scale;
    const tx = frame.offsetX + (effect.endX || effect.x) * frame.scale;
    const ty = frame.offsetY + (effect.endY || effect.y) * frame.scale;
    const radius = Math.max(24, ((effect.radius || 22) + 30) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * 2.4);
    const progress = 1 - effectAlpha(effect);
    const dropY = ty - radius * 3.2 * (1 - clamp(progress / 0.42, 0, 1));
    const angle = Math.atan2(effect.dirY || ty - sy, effect.dirX || tx - sx);

    skillLayer.moveTo(sx, sy);
    skillLayer.lineTo(tx, ty);
    skillLayer.stroke({ color: 0x7f1d1d, width: 5, alpha: 0.22 * alpha });
    skillLayer.moveTo(tx, dropY - radius * 0.85);
    skillLayer.lineTo(tx - radius * 0.7, dropY + radius * 0.15);
    skillLayer.lineTo(tx + radius * 0.7, dropY + radius * 0.15);
    skillLayer.closePath();
    skillLayer.fill({ color: 0xf8fafc, alpha: 0.92 * alpha });
    skillLayer.stroke({ color: 0x991b1b, width: 4, alpha: 0.9 * alpha });
    skillLayer.moveTo(tx - Math.cos(angle) * radius, ty - Math.sin(angle) * radius);
    skillLayer.lineTo(tx + Math.cos(angle) * radius, ty + Math.sin(angle) * radius);
    skillLayer.stroke({ color: 0xdc2626, width: Math.max(6, radius * 0.22), alpha: 0.74 * alpha });
    skillLayer.circle(tx, ty, radius * (0.45 + progress * 0.9));
    skillLayer.stroke({ color: 0xef4444, width: 4, alpha: (1 - progress * 0.45) * alpha });
    for (let i = 0; i < Math.max(5, 5 + (effect.count || 0)); i++) {
        const a = (Math.PI * 2 * i) / Math.max(5, 5 + (effect.count || 0)) + progress;
        skillLayer.circle(
            tx + Math.cos(a) * radius * 0.72,
            ty + Math.sin(a) * radius * 0.5,
            Math.max(3, radius * 0.07),
        );
        skillLayer.fill({ color: i % 2 ? 0xfca5a5 : 0x991b1b, alpha: 0.48 * alpha });
    }
}

function drawActiveSkillRanges(frame) {
    const tick = interpolatedTick();
    for (const player of state.players.values()) {
        if (player.dead) {
            continue;
        }
        if (player.heroId !== 'warrior') {
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
    skillLayer.lineTo(
        centerX - forwardX * length * 0.48 + sideX * width,
        centerY - forwardY * length * 0.48 + sideY * width,
    );
    skillLayer.lineTo(
        centerX - forwardX * length * 0.48 - sideX * width,
        centerY - forwardY * length * 0.48 - sideY * width,
    );
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

/**
 * 绘制石头人 W 强化攻击命中后的扇形岩层震波。
 * @param {object} effect 服务端同步的震波范围与方向。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
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
    const progress = 1 - alpha;
    skillLayer.moveTo(x, y);
    skillLayer.arc(x, y, radius, startAngle, endAngle);
    skillLayer.closePath();
    skillLayer.fill({ color: 0x92400e, alpha: 0.16 * alpha });
    skillLayer.moveTo(x, y);
    skillLayer.arc(x, y, radius * (0.72 + progress * 0.28), startAngle, endAngle);
    skillLayer.closePath();
    skillLayer.stroke({ color: 0xf59e0b, width: 4, alpha: 0.82 * alpha });

    for (let index = 0; index < 7; index++) {
        const crackAngle = startAngle + (angle * (index + 0.5)) / 7;
        const length = radius * (0.58 + (index % 3) * 0.16);
        drawTankRockCrack(x, y, crackAngle, length, alpha, index);
        const rockX = x + Math.cos(crackAngle) * length * 0.82;
        const rockY = y + Math.sin(crackAngle) * length * 0.82;
        const rockSize = Math.max(4, radius * (0.026 + (index % 2) * 0.012));
        skillLayer
            .moveTo(rockX, rockY - rockSize * (1.3 - progress * 0.4))
            .lineTo(rockX + rockSize, rockY)
            .lineTo(rockX + rockSize * 0.2, rockY + rockSize * 0.72)
            .lineTo(rockX - rockSize, rockY + rockSize * 0.2)
            .closePath();
        skillLayer.fill({ color: index % 2 ? 0x78716c : 0xa16207, alpha: 0.9 * alpha });
    }
}

/**
 * 绘制石头人 R 落地时的环形冲击、放射地裂与飞散岩块。
 * @param {object} effect 服务端同步的落点效果。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawTankImpactEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || 250) * frame.scale;
    const alpha = effectAlpha(effect);
    const progress = 1 - alpha;
    skillLayer.circle(x, y, radius * (0.34 + progress * 0.66));
    skillLayer.fill({ color: 0x292524, alpha: 0.2 * alpha });
    for (let ring = 0; ring < 3; ring++) {
        skillLayer.circle(x, y, radius * (0.24 + progress * 0.62 + ring * 0.08));
        skillLayer.stroke({
            color: ring === 1 ? 0xf59e0b : 0x57534e,
            width: Math.max(2, radius * (0.035 - ring * 0.006)),
            alpha: (0.88 - ring * 0.18) * alpha,
        });
    }
    for (let index = 0; index < 12; index++) {
        const angle = (Math.PI * 2 * index) / 12 + (index % 2) * 0.13;
        drawTankRockCrack(x, y, angle, radius * (0.62 + (index % 3) * 0.15), alpha, index);
        const distance = radius * (0.32 + progress * 0.62);
        const rockSize = Math.max(4, radius * (0.025 + (index % 3) * 0.008));
        const rockX = x + Math.cos(angle) * distance;
        const rockY = y + Math.sin(angle) * distance - Math.sin(progress * Math.PI) * radius * 0.22;
        skillLayer.rect(rockX - rockSize / 2, rockY - rockSize / 2, rockSize, rockSize);
        skillLayer.fill({ color: index % 2 ? 0x78716c : 0xa8a29e, alpha: 0.88 * alpha });
    }
    skillLayer.circle(x, y, Math.max(10, radius * 0.13));
    skillLayer.fill({ color: 0xfbbf24, alpha: 0.3 * alpha });
}

/**
 * 绘制一条带分叉的石质地裂，供石头人 W、E、R 复用。
 * @param {number} x 裂缝起点屏幕横坐标。
 * @param {number} y 裂缝起点屏幕纵坐标。
 * @param {number} angle 裂缝延伸弧度。
 * @param {number} length 裂缝屏幕长度。
 * @param {number} alpha 当前透明度。
 * @param {number} seed 用于稳定控制折线方向的序号。
 */
function drawTankRockCrack(x, y, angle, length, alpha, seed) {
    const sideAngle = angle + Math.PI / 2;
    const bend = ((seed % 3) - 1) * length * 0.08;
    const midX = x + Math.cos(angle) * length * 0.48 + Math.cos(sideAngle) * bend;
    const midY = y + Math.sin(angle) * length * 0.48 + Math.sin(sideAngle) * bend;
    const endX = x + Math.cos(angle) * length;
    const endY = y + Math.sin(angle) * length;
    skillLayer.moveTo(x, y).lineTo(midX, midY).lineTo(endX, endY);
    skillLayer.stroke({ color: 0x292524, width: Math.max(2, length * 0.018), alpha: 0.9 * alpha });
    skillLayer
        .moveTo(midX, midY)
        .lineTo(
            midX + Math.cos(angle + (seed % 2 ? 0.72 : -0.72)) * length * 0.22,
            midY + Math.sin(angle + (seed % 2 ? 0.72 : -0.72)) * length * 0.22,
        );
    skillLayer.stroke({ color: 0xf59e0b, width: Math.max(1, length * 0.008), alpha: 0.58 * alpha });
}

function drawBasicArrowEffect(effect, frame) {
    if (effect.sourceHeroId === 'mage') {
        drawMageBasicStarEffect(effect, frame);
        return;
    }
    if (effect.sourceHeroId === 'explorer') {
        drawExplorerBasicEffect(effect, frame);
        return;
    }
    if (effect.sourceHeroId === 'frostmage') {
        drawArrowProjectile(effect, frame, 0xbae6fd, 0x38bdf8, {
            fromSnapshot: true,
        });
        return;
    }
    if (effect.sourceHeroId === 'fire_mage') {
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
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const forwardX = Math.cos(angle);
    const forwardY = Math.sin(angle);
    const sideX = -forwardY;
    const sideY = forwardX;
    const length = Math.max(22, (effect.radius || 16) * frame.scale * 4.2);
    const width = Math.max(5, length * 0.16);
    const pulse = 0.82 + Math.sin(performance.now() / 42) * 0.18;

    skillLayer.moveTo(x - forwardX * length * 1.8, y - forwardY * length * 1.8);
    skillLayer.lineTo(x + forwardX * length * 0.15, y + forwardY * length * 0.15);
    skillLayer.stroke({ color: 0xf97316, width: width * 2.8, alpha: 0.1 * pulse });
    skillLayer.moveTo(x - forwardX * length * 1.35, y - forwardY * length * 1.35);
    skillLayer.lineTo(x + forwardX * length * 0.35, y + forwardY * length * 0.35);
    skillLayer.stroke({ color: 0xfbbf24, width: width, alpha: 0.68 * pulse });
    skillLayer.moveTo(x - forwardX * length * 0.85, y - forwardY * length * 0.85);
    skillLayer.lineTo(x + forwardX * length * 0.48, y + forwardY * length * 0.48);
    skillLayer.stroke({ color: 0xfffbeb, width: Math.max(2, width * 0.38), alpha: 0.95 });

    skillLayer
        .moveTo(x + forwardX * length * 0.68, y + forwardY * length * 0.68)
        .lineTo(x - forwardX * length * 0.22 + sideX * width, y - forwardY * length * 0.22 + sideY * width)
        .lineTo(x - forwardX * length * 0.48, y - forwardY * length * 0.48)
        .lineTo(x - forwardX * length * 0.22 - sideX * width, y - forwardY * length * 0.22 - sideY * width)
        .closePath();
    skillLayer.fill({ color: 0xfff7d6, alpha: 0.98 });
    skillLayer.stroke({ color: 0xea580c, width: 1.5, alpha: 0.92 });

    skillLayer.circle(x - forwardX * length * 0.35, y - forwardY * length * 0.35, width * 0.72);
    skillLayer.stroke({ color: 0xfde68a, width: 1.5, alpha: 0.7 });
}

function drawGunnerQMuzzleEffect(effect, frame) {
    const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
    const x = frame.offsetX + source.x * frame.scale;
    const y = frame.offsetY + source.y * frame.scale;
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const forwardX = Math.cos(angle);
    const forwardY = Math.sin(angle);
    const sideX = -forwardY;
    const sideY = forwardX;
    const alpha = Math.min(1, effectAlpha(effect) * 2.4);
    const progress = 1 - effectAlpha(effect);
    const length = Math.max(34, (effect.range || 150) * frame.scale * (0.78 + progress * 0.22));
    const width = Math.max(12, (effect.width || 44) * frame.scale * (1.15 - progress * 0.25));
    const muzzleX = x + forwardX * Math.max(12, 44 * frame.scale);
    const muzzleY = y + forwardY * Math.max(12, 44 * frame.scale);

    skillLayer
        .moveTo(muzzleX, muzzleY)
        .lineTo(muzzleX + forwardX * length + sideX * width * 0.12, muzzleY + forwardY * length + sideY * width * 0.12)
        .lineTo(muzzleX + forwardX * length * 0.46 + sideX * width, muzzleY + forwardY * length * 0.46 + sideY * width)
        .lineTo(muzzleX + forwardX * length * 0.18, muzzleY + forwardY * length * 0.18)
        .lineTo(muzzleX + forwardX * length * 0.46 - sideX * width, muzzleY + forwardY * length * 0.46 - sideY * width)
        .closePath();
    skillLayer.fill({ color: 0xf59e0b, alpha: 0.42 * alpha });
    skillLayer.stroke({ color: 0xfff7d6, width: 2, alpha: 0.88 * alpha });

    skillLayer.circle(muzzleX, muzzleY, width * (0.48 + progress * 0.35));
    skillLayer.fill({ color: 0xffffff, alpha: 0.9 * alpha });
    skillLayer.circle(muzzleX, muzzleY, width * (1.1 + progress * 0.65));
    skillLayer.stroke({ color: 0xfbbf24, width: Math.max(2, width * 0.16), alpha: 0.64 * alpha });
    for (let i = -2; i <= 2; i++) {
        const rayAngle = angle + i * 0.22;
        const rayLength = length * (0.52 + (2 - Math.abs(i)) * 0.13);
        skillLayer.moveTo(muzzleX, muzzleY);
        skillLayer.lineTo(muzzleX + Math.cos(rayAngle) * rayLength, muzzleY + Math.sin(rayAngle) * rayLength);
    }
    skillLayer.stroke({ color: 0xfffbeb, width: 1.5, alpha: 0.62 * alpha });
}

function drawGunnerWEffect(effect, frame) {
    const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
    const x = frame.offsetX + source.x * frame.scale;
    const y = frame.offsetY + source.y * frame.scale;
    const radius = Math.max(30, (effect.radius || 105) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * 5);
    const time = performance.now();
    const pulse = 0.5 + Math.sin(time / 105) * 0.5;
    const rotation = time / 360;

    skillLayer.circle(x, y, radius * (0.88 + pulse * 0.08));
    skillLayer.fill({ color: 0x0891b2, alpha: 0.07 * alpha });
    skillLayer.circle(x, y, radius * (0.92 + pulse * 0.06));
    skillLayer.stroke({ color: 0x67e8f9, width: 2.5, alpha: (0.44 + pulse * 0.25) * alpha });

    for (let i = 0; i < 3; i++) {
        const start = rotation + (Math.PI * 2 * i) / 3;
        const orbit = radius * (0.62 + i * 0.12);
        skillLayer.moveTo(x + Math.cos(start) * orbit, y + Math.sin(start) * orbit);
        skillLayer.arc(x, y, orbit, start, start + 0.72);
        skillLayer.stroke({
            color: i === 1 ? 0xfffbeb : 0x22d3ee,
            width: Math.max(2, radius * 0.08),
            alpha: (0.78 - i * 0.12) * alpha,
        });
    }

    for (let i = 0; i < 4; i++) {
        const phase = (time / 620 + i / 4) % 1;
        const side = i % 2 ? 1 : -1;
        const px = x + side * radius * (0.28 + phase * 0.55);
        const py = y + radius * (0.72 - phase * 1.44);
        const wing = Math.max(5, radius * 0.18 * (1 - phase * 0.35));
        skillLayer.moveTo(px - wing, py + wing * 0.45);
        skillLayer.lineTo(px, py - wing * 0.45);
        skillLayer.lineTo(px + wing, py + wing * 0.45);
    }
    skillLayer.stroke({ color: 0xecfeff, width: 2, alpha: 0.58 * alpha });
}

function drawFireMageQEffect(effect, frame) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const radius = Math.max(8, (effect.radius || 28) * frame.scale * 0.42);
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const pulse = 0.5 + Math.sin(performance.now() / 70) * 0.5;
    skillLayer.circle(x, y, radius * (1.35 + pulse * 0.12));
    skillLayer.fill({ color: 0xdc2626, alpha: 0.22 });
    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0xf97316, alpha: 0.92 });
    skillLayer.circle(x + Math.cos(angle) * radius * 0.32, y + Math.sin(angle) * radius * 0.32, radius * 0.5);
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
    const startX = frame.offsetX + (effect.x || position.x) * frame.scale;
    const startY = frame.offsetY + (effect.y || position.y) * frame.scale;
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const radius = Math.max(24, (effect.radius || 70) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * 2.2);
    const progress = 1 - effectAlpha(effect);
    const dir = Math.atan2(
        effect.dirY || position.y - (effect.y || position.y),
        effect.dirX || position.x - (effect.x || position.x) || 1,
    );
    const nx = -Math.sin(dir);
    const ny = Math.cos(dir);
    const rotation = performance.now() / 70 + progress * 1.6;

    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0x0f172a, width: Math.max(22, radius * 0.72), alpha: 0.14 * alpha });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0x38bdf8, width: Math.max(12, radius * 0.42), alpha: 0.2 * alpha });
    for (let index = -1; index <= 1; index++) {
        const offset = index * radius * 0.28;
        skillLayer.moveTo(startX + nx * offset, startY + ny * offset);
        skillLayer.lineTo(x + nx * offset * 0.35, y + ny * offset * 0.35);
        skillLayer.stroke({
            color: index === 0 ? 0xf8fafc : 0x67e8f9,
            width: index === 0 ? Math.max(3, radius * 0.1) : Math.max(2, radius * 0.055),
            alpha: (index === 0 ? 0.78 : 0.5) * alpha,
        });
    }

    skillLayer.circle(x, y, radius * 0.86);
    skillLayer.fill({ color: 0x082f49, alpha: 0.12 * alpha });
    for (let index = 0; index < 4; index++) {
        const start = rotation + (Math.PI * 2 * index) / 4;
        const arcRadius = radius * (0.48 + (index % 2) * 0.22);
        skillLayer.moveTo(x + Math.cos(start) * arcRadius * 0.45, y + Math.sin(start) * arcRadius * 0.45);
        skillLayer.arc(x, y, arcRadius, start, start + Math.PI * 1.05);
        skillLayer.stroke({
            color: index % 2 ? 0xfef3c7 : 0xe0f2fe,
            width: Math.max(4, radius * (index % 2 ? 0.09 : 0.12)),
            alpha: (0.72 - index * 0.06) * alpha,
        });
    }
    for (let index = 0; index < 3; index++) {
        const side = index - 1;
        const sx = x - Math.cos(dir) * radius * (0.8 + index * 0.1) + nx * side * radius * 0.32;
        const sy = y - Math.sin(dir) * radius * (0.8 + index * 0.1) + ny * side * radius * 0.32;
        const ex = x + Math.cos(dir) * radius * 0.68 + nx * side * radius * 0.16;
        const ey = y + Math.sin(dir) * radius * 0.68 + ny * side * radius * 0.16;
        skillLayer.moveTo(sx, sy);
        skillLayer.lineTo(ex, ey);
        skillLayer.stroke({
            color: index === 1 ? 0xffffff : 0xfacc15,
            width: Math.max(3, radius * 0.075),
            alpha: (0.7 - index * 0.08) * alpha,
        });
    }
    skillLayer.circle(x, y, Math.max(7, radius * 0.16));
    skillLayer.fill({ color: 0xf8fafc, alpha: 0.3 * alpha });
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
        skillLayer.stroke({
            color: index === 0 ? 0xe0f2fe : 0x38bdf8,
            width: 3 - index * 0.5,
            alpha: 0.9 - index * 0.2,
        });
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
            const distance = (length * step) / 4;
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

function drawShadowAssassinQReadyEffects(frame) {
    const rotation = performance.now() / 260;
    for (const player of state.players.values()) {
        if (player.dead || player.heroId !== 'shadow_assassin') {
            continue;
        }
        if (!(player.buffs || []).some(buff => buff.id === 'shadow_assassin_q_ready')) {
            continue;
        }
        const x = frame.offsetX + player.x * frame.scale;
        const y = frame.offsetY + player.y * frame.scale;
        const radius = playerModelRadius(player) + 8;
        for (let index = 0; index < 2; index++) {
            const start = rotation + index * Math.PI;
            skillLayer.moveTo(x + Math.cos(start) * radius, y + Math.sin(start) * radius);
            skillLayer.arc(x, y, radius, start, start + Math.PI * 0.72);
            skillLayer.stroke({ color: index === 0 ? 0xef4444 : 0xf8fafc, width: 3, alpha: 0.82 });
            const tip = start + Math.PI * 0.72;
            const tipX = x + Math.cos(tip) * radius;
            const tipY = y + Math.sin(tip) * radius;
            skillLayer
                .moveTo(tipX + Math.cos(tip) * 7, tipY + Math.sin(tip) * 7)
                .lineTo(tipX - Math.sin(tip) * 4, tipY + Math.cos(tip) * 4)
                .lineTo(tipX + Math.sin(tip) * 3, tipY - Math.cos(tip) * 3)
                .closePath();
            skillLayer.fill({ color: 0x991b1b, alpha: 0.88 });
        }
    }
}

function drawShadowAssassinQCastEffect(effect, frame) {
    const alpha = Math.min(1, effectAlpha(effect) * 2.4);
    const progress = 1 - effectAlpha(effect);
    const worldX = effect.endX ?? effect.x;
    const worldY = effect.endY ?? effect.y;
    const x = frame.offsetX + worldX * frame.scale;
    const y = frame.offsetY + worldY * frame.scale;
    const radius = Math.max(22, ((effect.radius || 18) + 40) * frame.scale);
    const rotation = progress * Math.PI * 1.35;

    for (let index = 0; index < 2; index++) {
        const angle = rotation + index * Math.PI;
        const inner = radius * (0.24 + progress * 0.2);
        const outer = radius * (1.2 - progress * 0.25);
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle + 0.38) * outer, y + Math.sin(angle + 0.38) * outer);
        skillLayer.lineTo(x + Math.cos(angle - 0.16) * radius * 0.72, y + Math.sin(angle - 0.16) * radius * 0.72);
        skillLayer.closePath();
        skillLayer.fill({ color: index === 0 ? 0xdc2626 : 0xf8fafc, alpha: 0.72 * alpha });
    }
    skillLayer.circle(x, y, radius * (0.35 + progress * 0.65));
    skillLayer.stroke({ color: 0x991b1b, width: 3, alpha: 0.66 * alpha });
}

function drawShadowAssassinQHitEffect(effect, frame) {
    const alpha = Math.min(1, effectAlpha(effect) * 3);
    const progress = 1 - effectAlpha(effect);
    const startX = frame.offsetX + effect.x * frame.scale;
    const startY = frame.offsetY + effect.y * frame.scale;
    const endX = frame.offsetX + effect.endX * frame.scale;
    const endY = frame.offsetY + effect.endY * frame.scale;
    const dx = endX - startX;
    const dy = endY - startY;
    const length = Math.hypot(dx, dy) || 1;
    const dirX = dx / length;
    const dirY = dy / length;
    const sideX = -dirY;
    const sideY = dirX;
    const radius = Math.max(18, ((effect.radius || 18) + 20) * frame.scale);

    skillLayer.moveTo(startX, startY).lineTo(endX, endY);
    skillLayer.stroke({ color: 0x7f1d1d, width: 8, alpha: 0.22 * alpha });
    skillLayer.moveTo(startX + dirX * length * progress, startY + dirY * length * progress).lineTo(endX, endY);
    skillLayer.stroke({ color: 0xf8fafc, width: 2, alpha: 0.86 * alpha });
    for (const side of [-1, 1]) {
        skillLayer
            .moveTo(
                endX - dirX * radius + sideX * radius * 0.62 * side,
                endY - dirY * radius + sideY * radius * 0.62 * side,
            )
            .lineTo(
                endX + dirX * radius - sideX * radius * 0.62 * side,
                endY + dirY * radius - sideY * radius * 0.62 * side,
            );
        skillLayer.stroke({ color: side < 0 ? 0xef4444 : 0xf8fafc, width: 4, alpha: 0.88 * alpha });
    }
    skillLayer.circle(endX, endY, radius * (0.35 + progress * 0.9));
    skillLayer.stroke({ color: 0x991b1b, width: 3, alpha: (1 - progress) * 0.76 });
}

function drawShadowAssassinQRevealEffect(effect, frame) {
    const worldX = effect.endX ?? effect.x;
    const worldY = effect.endY ?? effect.y;
    const x = frame.offsetX + worldX * frame.scale;
    const y = frame.offsetY + worldY * frame.scale;
    const radius = Math.max(16, ((effect.radius || 18) + 18) * frame.scale);
    const alpha = Math.max(0.22, effectAlpha(effect));
    const pulse = 0.88 + Math.sin(performance.now() / 120) * 0.1;

    skillLayer.ellipse(x, y, radius * pulse, radius * 0.46 * pulse);
    skillLayer.stroke({ color: 0xdc2626, width: 2, alpha: 0.7 * alpha });
    skillLayer.circle(x, y, radius * 0.16);
    skillLayer.fill({ color: 0xfef2f2, alpha: 0.82 * alpha });
    for (let index = -1; index <= 1; index++) {
        const offset = index * radius * 0.3;
        skillLayer
            .moveTo(x - radius * 0.4 + offset, y - radius * 0.44)
            .lineTo(x + radius * 0.12 + offset, y + radius * 0.38);
        skillLayer.stroke({
            color: index === 0 ? 0xfef2f2 : 0x991b1b,
            width: index === 0 ? 2 : 3,
            alpha: 0.72 * alpha,
        });
    }
    const fall = (performance.now() / 520) % 1;
    for (let index = 0; index < 3; index++) {
        const phase = (fall + index / 3) % 1;
        const dropX = x + (index - 1) * radius * 0.28;
        const dropY = y + radius * (0.15 + phase * 0.8);
        skillLayer.circle(dropX, dropY, Math.max(2, radius * 0.06 * (1 - phase * 0.35)));
        skillLayer.fill({ color: 0xb91c1c, alpha: (1 - phase) * 0.72 * alpha });
    }
}

/**
 * 绘制影刃 W 的银红飞刀，突出高速投掷与锋利刃口。
 * @param {object} effect 服务端同步的投射物效果。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawShadowAssassinWEffect(effect, frame) {
    drawShadowBladeProjectile(effect, frame, 0xf8fafc, 0xb91c1c, 0x450a0a, false);
}

/**
 * 绘制影刃专属飞刀轮廓与运动残影，供 W 和 R 共用一致的武器造型。
 * @param {object} effect 服务端同步的投射物效果。
 * @param {object} frame 当前世界到屏幕的变换参数。
 * @param {number} bladeColor 刀身填充颜色。
 * @param {number} edgeColor 刃口与刀脊颜色。
 * @param {number} trailColor 运动残影颜色。
 * @param {boolean} spectral 是否绘制 R 的幽影外层。
 */
function drawShadowBladeProjectile(effect, frame, bladeColor, edgeColor, trailColor, spectral) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const forwardX = Math.cos(angle);
    const forwardY = Math.sin(angle);
    const sideX = -forwardY;
    const sideY = forwardX;
    const length = Math.max(14, (effect.radius || 18) * frame.scale * 0.82);
    const width = Math.max(5, length * 0.34);
    const trailLength = length * (spectral ? 1.9 : 1.45);

    // 先绘制沿飞行方向收窄的残影，避免刀身被半透明轨迹覆盖。
    skillLayer
        .moveTo(x - forwardX * trailLength + sideX * width * 0.55, y - forwardY * trailLength + sideY * width * 0.55)
        .lineTo(x + forwardX * length * 0.18, y + forwardY * length * 0.18)
        .lineTo(x - forwardX * trailLength - sideX * width * 0.55, y - forwardY * trailLength - sideY * width * 0.55)
        .closePath();
    skillLayer.fill({ color: trailColor, alpha: spectral ? 0.2 : 0.16 });

    if (spectral) {
        skillLayer
            .moveTo(x + forwardX * length * 1.16, y + forwardY * length * 1.16)
            .lineTo(
                x - forwardX * length * 0.72 + sideX * width * 1.25,
                y - forwardY * length * 0.72 + sideY * width * 1.25,
            )
            .lineTo(x - forwardX * length * 0.48, y - forwardY * length * 0.48)
            .lineTo(
                x - forwardX * length * 0.72 - sideX * width * 1.25,
                y - forwardY * length * 0.72 - sideY * width * 1.25,
            )
            .closePath();
        skillLayer.stroke({ color: edgeColor, width: 4, alpha: 0.24 });
    }

    skillLayer
        .moveTo(x + forwardX * length * 1.18, y + forwardY * length * 1.18)
        .lineTo(x - forwardX * length * 0.06 + sideX * width, y - forwardY * length * 0.06 + sideY * width)
        .lineTo(
            x - forwardX * length * 0.38 + sideX * width * 0.42,
            y - forwardY * length * 0.38 + sideY * width * 0.42,
        )
        .lineTo(
            x - forwardX * length * 0.92 + sideX * width * 0.72,
            y - forwardY * length * 0.92 + sideY * width * 0.72,
        )
        .lineTo(x - forwardX * length * 0.66, y - forwardY * length * 0.66)
        .lineTo(
            x - forwardX * length * 0.92 - sideX * width * 0.72,
            y - forwardY * length * 0.92 - sideY * width * 0.72,
        )
        .lineTo(
            x - forwardX * length * 0.38 - sideX * width * 0.42,
            y - forwardY * length * 0.38 - sideY * width * 0.42,
        )
        .lineTo(x - forwardX * length * 0.06 - sideX * width, y - forwardY * length * 0.06 - sideY * width)
        .closePath();
    skillLayer.fill({ color: bladeColor, alpha: 0.97 });
    skillLayer.stroke({ color: edgeColor, width: spectral ? 2.4 : 2, alpha: 0.94 });

    skillLayer
        .moveTo(x + forwardX * length * 0.88, y + forwardY * length * 0.88)
        .lineTo(x - forwardX * length * 0.55, y - forwardY * length * 0.55);
    skillLayer.stroke({ color: edgeColor, width: 1.5, alpha: 0.82 });
    skillLayer
        .moveTo(x + sideX * width * 0.28, y + sideY * width * 0.28)
        .lineTo(x + forwardX * length * 0.22, y + forwardY * length * 0.22)
        .lineTo(x - sideX * width * 0.28, y - sideY * width * 0.28)
        .lineTo(x - forwardX * length * 0.18, y - forwardY * length * 0.18)
        .closePath();
    skillLayer.fill({ color: edgeColor, alpha: 0.9 });
}

function drawShadowAssassinEEffect(effect, frame) {
    const startX = frame.offsetX + effect.x * frame.scale;
    const startY = frame.offsetY + effect.y * frame.scale;
    const endX = frame.offsetX + effect.endX * frame.scale;
    const endY = frame.offsetY + effect.endY * frame.scale;
    const alpha = effectAlpha(effect);

    skillLayer.moveTo(startX, startY).lineTo(endX, endY);
    skillLayer.stroke({ color: 0xdc2626, width: 5, alpha: 0.55 * alpha });
    skillLayer.circle(endX, endY, Math.max(16, (effect.radius || 18) * frame.scale + 8));
    skillLayer.stroke({ color: 0xfca5a5, width: 2, alpha: 0.72 * alpha });
}

function drawShadowAssassinEMarkEffect(effect, frame) {
    const x = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const y = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const radius = Math.max(18, ((effect.radius || 18) + 12) * frame.scale);
    const pulse = 0.9 + Math.sin(performance.now() / 140) * 0.08;

    skillLayer.circle(x, y, radius * pulse);
    skillLayer.stroke({ color: 0x991b1b, width: 2, alpha: 0.72 });
    skillLayer.moveTo(x - radius * 0.28, y).lineTo(x + radius * 0.28, y);
    skillLayer.stroke({ color: 0xfef2f2, width: 2, alpha: 0.82 });
}

/**
 * 绘制影刃 R 的冷白紫影飞刀，幽影外层用于区分大招刀阵。
 * @param {object} effect 服务端同步的投射物效果。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawShadowAssassinREffect(effect, frame) {
    drawShadowBladeProjectile(effect, frame, 0xede9fe, 0x7e22ce, 0x2e1065, true);
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

    skillLayer.moveTo(startX, startY).lineTo(groundX, groundY);
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
    skillLayer.moveTo(x + Math.cos(rotation) * radius * 0.92, y + Math.sin(rotation) * radius * 0.92);
    skillLayer.arc(x, y, radius * 0.92, rotation, rotation + Math.PI * 1.45);
    skillLayer.stroke({ color: 0xc4b5fd, width: 5, alpha: 0.82 * alpha });
    const innerStart = rotation + Math.PI;
    skillLayer.moveTo(x + Math.cos(innerStart) * radius * 0.68, y + Math.sin(innerStart) * radius * 0.68);
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
        skillLayer.lineTo(
            bladeX - Math.cos(angle) * length + sideX * length * 0.35,
            bladeY - Math.sin(angle) * length + sideY * length * 0.35,
        );
        skillLayer.lineTo(
            bladeX - Math.cos(angle) * length - sideX * length * 0.35,
            bladeY - Math.sin(angle) * length - sideY * length * 0.35,
        );
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
    const radius = Math.max(42, (effect.radius || 260) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * 2.4);
    const progress = 1 - effectAlpha(effect);
    const height = radius * (1.5 - progress * 0.55);
    const core = Math.max(16, radius * 0.2);
    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0x7f1d1d, alpha: 0.07 * alpha });
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0xef4444, width: Math.max(3, radius * 0.018), alpha: 0.68 * alpha });
    skillLayer.circle(x, y, radius * (0.2 + progress * 0.38));
    skillLayer.fill({ color: 0x7f1d1d, alpha: 0.18 * alpha });
    for (let i = 0; i < 7; i++) {
        const angle = (Math.PI * 2 * i) / 7 + performance.now() / 180;
        const base = radius * (0.16 + (i % 3) * 0.055);
        const px = x + Math.cos(angle) * radius * 0.18;
        const py = y + Math.sin(angle) * radius * 0.08;
        skillLayer.moveTo(px, py);
        skillLayer.quadraticCurveTo(
            x + Math.cos(angle) * radius * 0.32,
            y - height * (0.45 + (i % 2) * 0.16),
            x + Math.cos(angle) * base,
            y - height,
        );
        skillLayer.stroke({
            color: i % 2 ? 0xfef3c7 : 0xf97316,
            width: Math.max(4, radius * 0.04),
            alpha: (0.72 - i * 0.035) * alpha,
        });
    }
    skillLayer.circle(x, y, core * (1.1 + progress * 0.35));
    skillLayer.fill({ color: 0xfef3c7, alpha: 0.44 * alpha });
    if ((effect.count || 0) > 0) {
        skillLayer.circle(x, y, radius * 0.36);
        skillLayer.stroke({ color: 0xfca5a5, width: Math.max(3, radius * 0.035), alpha: 0.46 * alpha });
    }
}

function drawFireMageWRangeEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || 260) * frame.scale;
    const alpha = Math.min(1, effectAlpha(effect) * 1.8);
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
    skillLayer.lineTo(
        x - Math.cos(angle) * radius * 1.1 - Math.sin(angle) * radius * 0.5,
        y - Math.sin(angle) * radius * 1.1 + Math.cos(angle) * radius * 0.5,
    );
    skillLayer.lineTo(x - Math.cos(angle) * radius * 0.6, y - Math.sin(angle) * radius * 0.6);
    skillLayer.lineTo(
        x - Math.cos(angle) * radius * 1.1 + Math.sin(angle) * radius * 0.5,
        y - Math.sin(angle) * radius * 1.1 - Math.cos(angle) * radius * 0.5,
    );
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
    skillLayer.lineTo(
        x - forwardX * radius * 0.45 + sideX * radius * 0.95,
        y - forwardY * radius * 0.45 + sideY * radius * 0.95,
    );
    skillLayer.lineTo(x - forwardX * radius * 0.15, y - forwardY * radius * 0.15);
    skillLayer.lineTo(
        x - forwardX * radius * 0.45 - sideX * radius * 0.95,
        y - forwardY * radius * 0.45 - sideY * radius * 0.95,
    );
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
    return effect.id || `${effect.sourceId || 'frostmage_servant'}:${effect.createdAt || 0}`;
}

function drawFireMageEEffect(effect, frame) {
    const x = frame.offsetX + (effect.endX || effect.x) * frame.scale;
    const y = frame.offsetY + (effect.endY || effect.y) * frame.scale;
    const radius = Math.max(20, ((effect.radius || 18) + 28) * frame.scale);
    const alpha = Math.min(1, effectAlpha(effect) * 2.6);
    const progress = 1 - effectAlpha(effect);
    skillLayer.circle(x, y, radius * (0.62 + progress * 0.44));
    skillLayer.fill({ color: 0x7f1d1d, alpha: 0.2 * alpha });
    for (let i = 0; i < 6; i++) {
        const angle = (Math.PI * 2 * i) / 6 + progress * 0.8;
        const inner = radius * 0.35;
        const outer = radius * (1.05 + (i % 2) * 0.25);
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
        skillLayer.stroke({
            color: i % 2 ? 0xfef3c7 : 0xf97316,
            width: Math.max(3, radius * 0.11),
            alpha: 0.78 * alpha,
        });
    }
    if ((effect.count || 0) > 0) {
        const spread = Math.max(radius * 1.6, (effect.width || 600) * frame.scale * 0.42);
        for (let i = 0; i < 5; i++) {
            const angle = (Math.PI * 2 * i) / 5 + performance.now() / 250;
            skillLayer.moveTo(x, y);
            skillLayer.quadraticCurveTo(
                x + Math.cos(angle + 0.28) * spread * 0.45,
                y + Math.sin(angle + 0.28) * spread * 0.45,
                x + Math.cos(angle) * spread,
                y + Math.sin(angle) * spread,
            );
            skillLayer.stroke({ color: 0xfb923c, width: Math.max(3, radius * 0.13), alpha: 0.34 * alpha });
        }
    }
    skillLayer.circle(x, y, radius * 0.38);
    skillLayer.fill({ color: 0xfef3c7, alpha: 0.5 * alpha });
}

function drawFireMageREffect(effect, frame) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const radius = Math.max(12, (effect.radius || 36) * frame.scale * 0.5);
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const pulse = 0.5 + Math.sin(performance.now() / 55) * 0.5;
    skillLayer.circle(x, y, radius * (1.45 + pulse * 0.14));
    skillLayer.fill({ color: 0xdc2626, alpha: 0.24 });
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

    skillLayer.moveTo(x - forwardX * tail, y - forwardY * tail).lineTo(x, y);
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
        const drift = (performance.now() / 850 + i * 0.2) % 1;
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
        drawGunnerRChannelEffect(effect, frame);
        return;
    }
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const count = Math.max(1, effect.count || 1);
    const center = (count - 1) / 2;
    const length = Math.max(17, (effect.radius || 18) * frame.scale * 2.8);
    const spread = Math.max(4, length * 0.12);
    const halfAngle = (((effect.width || 45) * Math.PI) / 180) * 0.5;
    const origin = effect.endX || effect.endY ? { x: effect.endX, y: effect.endY } : effectSourcePosition(effect);
    const traveled = origin ? Math.hypot(position.x - origin.x, position.y - origin.y) : 0;
    for (let i = 0; i < count; i++) {
        const bulletAngle = angle + (center ? ((i - center) / center) * halfAngle : 0);
        const forwardX = Math.cos(bulletAngle);
        const forwardY = Math.sin(bulletAngle);
        const x = frame.offsetX + (origin ? origin.x + forwardX * traveled : position.x) * frame.scale;
        const y = frame.offsetY + (origin ? origin.y + forwardY * traveled : position.y) * frame.scale;
        const waveAlpha = 0.56 + (1 - Math.abs(i - center) / Math.max(1, center)) * 0.28;
        skillLayer.moveTo(x - forwardX * length * 1.25, y - forwardY * length * 1.25);
        skillLayer.lineTo(x + forwardX * length * 0.35, y + forwardY * length * 0.35);
        skillLayer.stroke({ color: 0xea580c, width: spread * 2.8, alpha: 0.11 * waveAlpha });
        skillLayer.moveTo(x - forwardX * length, y - forwardY * length);
        skillLayer.lineTo(x + forwardX * length * 0.48, y + forwardY * length * 0.48);
        skillLayer.stroke({ color: 0xfbbf24, width: spread * 0.75, alpha: 0.68 * waveAlpha });
        skillLayer.moveTo(x + forwardX * length, y + forwardY * length);
        skillLayer.lineTo(x + Math.cos(bulletAngle + 2.55) * spread, y + Math.sin(bulletAngle + 2.55) * spread);
        skillLayer.lineTo(x + Math.cos(bulletAngle - 2.55) * spread, y + Math.sin(bulletAngle - 2.55) * spread);
        skillLayer.closePath();
        skillLayer.fill({ color: 0xfffbeb, alpha: 0.92 * waveAlpha });
        skillLayer.stroke({ color: 0xf97316, width: 1.25, alpha: 0.86 * waveAlpha });
    }
}

function drawGunnerRChannelEffect(effect, frame) {
    const source = effectSourcePosition(effect) || { x: effect.x, y: effect.y };
    const x = frame.offsetX + source.x * frame.scale;
    const y = frame.offsetY + source.y * frame.scale;
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const range = Math.max(160, (effect.range || 1400) * frame.scale);
    const halfAngle = (((effect.width || 45) * Math.PI) / 180) * 0.5;
    const alpha = Math.min(1, effectAlpha(effect) * 5);
    const elapsed = 1 - effectAlpha(effect);
    const scan = (performance.now() / 310) % 1;

    skillLayer.moveTo(x, y);
    skillLayer.arc(x, y, range, angle - halfAngle, angle + halfAngle);
    skillLayer.closePath();
    skillLayer.fill({ color: 0x7c2d12, alpha: 0.055 * alpha });

    for (const side of [-1, 1]) {
        const edgeAngle = angle + halfAngle * side;
        skillLayer.moveTo(x, y);
        skillLayer.lineTo(x + Math.cos(edgeAngle) * range, y + Math.sin(edgeAngle) * range);
    }
    skillLayer.stroke({ color: 0xf59e0b, width: 2, alpha: 0.46 * alpha });
    skillLayer.moveTo(x + Math.cos(angle - halfAngle) * range, y + Math.sin(angle - halfAngle) * range);
    skillLayer.arc(x, y, range, angle - halfAngle, angle + halfAngle);
    skillLayer.stroke({ color: 0xfbbf24, width: 2.5, alpha: 0.58 * alpha });

    for (let i = 0; i < 5; i++) {
        const lane = -1 + (i / 4) * 2;
        const laneAngle = angle + halfAngle * lane * 0.88;
        const start = range * (0.08 + ((scan + i * 0.17) % 1) * 0.22);
        const end = range * (0.52 + ((scan + i * 0.13) % 1) * 0.44);
        skillLayer.moveTo(x + Math.cos(laneAngle) * start, y + Math.sin(laneAngle) * start);
        skillLayer.lineTo(x + Math.cos(laneAngle) * end, y + Math.sin(laneAngle) * end);
    }
    skillLayer.stroke({ color: 0xfff7d6, width: 1.5, alpha: (0.18 + elapsed * 0.16) * alpha });

    const shockRadius = Math.max(22, 64 * frame.scale) * (0.9 + Math.sin(performance.now() / 70) * 0.08);
    skillLayer.circle(x, y, shockRadius);
    skillLayer.fill({ color: 0xf97316, alpha: 0.12 * alpha });
    skillLayer.circle(x, y, shockRadius * 0.72);
    skillLayer.stroke({ color: 0xfffbeb, width: 3, alpha: 0.68 * alpha });
}

function drawGunnerEEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || 300) * frame.scale;
    const alpha = Math.min(1, effectAlpha(effect) * 4);
    const time = performance.now();
    const progress = 1 - effectAlpha(effect);
    const pulse = (time / 430) % 1;

    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0x082f49, alpha: 0.16 * alpha });
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0x38bdf8, width: 3, alpha: 0.82 * alpha });
    skillLayer.circle(x, y, radius * (0.22 + pulse * 0.76));
    skillLayer.stroke({ color: 0x7dd3fc, width: 2.5, alpha: (1 - pulse) * 0.62 * alpha });
    skillLayer.circle(x, y, radius * (0.86 + Math.sin(time / 115) * 0.035));
    skillLayer.stroke({ color: 0xe0f2fe, width: 1.5, alpha: 0.34 * alpha });

    const streakLength = Math.max(14, radius * 0.14);
    for (let i = 0; i < 16; i++) {
        const seed = ((i * 47) % 101) / 101;
        const orbit = radius * (0.16 + (((i * 29) % 79) / 79) * 0.72);
        const angle = i * 2.39996 + progress * 1.8;
        const fall = (time / (360 + (i % 4) * 55) + seed) % 1;
        const px = x + Math.cos(angle) * orbit + streakLength * 0.22;
        const py = y + Math.sin(angle) * orbit - streakLength * (0.7 - fall);
        skillLayer.moveTo(px + streakLength * 0.38, py - streakLength);
        skillLayer.lineTo(px, py);
        skillLayer.stroke({
            color: i % 3 === 0 ? 0xfef3c7 : 0x7dd3fc,
            width: i % 3 === 0 ? 2.2 : 1.5,
            alpha: (0.36 + fall * 0.42) * alpha,
        });
        if (fall > 0.82) {
            const impact = (fall - 0.82) / 0.18;
            skillLayer.circle(px, py, Math.max(2, radius * 0.018) * (1 + impact * 1.8));
            skillLayer.stroke({ color: 0xbae6fd, width: 1.5, alpha: (1 - impact) * 0.68 * alpha });
        }
    }

    for (let i = 0; i < 4; i++) {
        const angle = Math.PI * 0.25 + (Math.PI * 2 * i) / 4;
        const inner = radius * 0.89;
        const outer = radius * 1.04;
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
    }
    skillLayer.stroke({ color: 0xe0f2fe, width: 2, alpha: 0.7 * alpha });
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

    skillLayer.moveTo(outerX + Math.cos(angle - spread) * outerRadius, outerY + Math.sin(angle - spread) * outerRadius);
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
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const normalX = -dirY;
    const normalY = dirX;
    const radius = Math.max(15, (effect.radius || 45) * frame.scale * 0.88);
    const tail = Math.max(44, radius * 4.8);
    const rotation = performance.now() / 115;

    skillLayer.moveTo(x - dirX * tail, y - dirY * tail);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0xfacc15, width: radius * 1.35, alpha: 0.18 });
    for (let i = -1; i <= 1; i++) {
        const offset = i * radius * 0.34;
        skillLayer.moveTo(
            x - dirX * tail * (0.72 + Math.abs(i) * 0.08) + normalX * offset,
            y - dirY * tail * (0.72 + Math.abs(i) * 0.08) + normalY * offset,
        );
        skillLayer.lineTo(
            x - dirX * radius * 0.18 + normalX * offset * 0.25,
            y - dirY * radius * 0.18 + normalY * offset * 0.25,
        );
        skillLayer.stroke({
            color: i === 0 ? 0xffffff : i < 0 ? 0x67e8f9 : 0xfde68a,
            width: Math.max(2, radius * (i === 0 ? 0.22 : 0.15)),
            alpha: i === 0 ? 0.94 : 0.68,
        });
    }

    skillLayer.circle(x, y, radius * 1.42);
    skillLayer.fill({ color: 0xfacc15, alpha: 0.16 });
    skillLayer.circle(x, y, radius * 1.4);
    skillLayer.stroke({ color: 0x67e8f9, width: Math.max(2, radius * 0.11), alpha: 0.56 });
    for (let i = 0; i < 3; i++) {
        const start = rotation + (Math.PI * 2 * i) / 3;
        skillLayer.moveTo(
            x + Math.cos(start) * radius * (0.78 + i * 0.08),
            y + Math.sin(start) * radius * (0.78 + i * 0.08),
        );
        skillLayer.arc(x, y, radius * (0.78 + i * 0.08), start, start + Math.PI * 1.22);
        skillLayer.stroke({
            color: i === 1 ? 0x67e8f9 : 0xfacc15,
            width: Math.max(2, radius * 0.16),
            alpha: 0.9 - i * 0.1,
        });
    }
    drawStarPath(skillLayer, x, y, radius * 0.82, radius * 0.34);
    skillLayer.fill({ color: 0xffffff, alpha: 0.96 });
    skillLayer.stroke({ color: 0xf59e0b, width: 2, alpha: 0.92 });
    skillLayer.moveTo(x + dirX * radius * 0.45, y + dirY * radius * 0.45);
    skillLayer.lineTo(
        x + dirX * radius * 1.65 + normalX * radius * 0.48,
        y + dirY * radius * 1.65 + normalY * radius * 0.48,
    );
    skillLayer.lineTo(x + dirX * radius * 1.28, y + dirY * radius * 1.28);
    skillLayer.lineTo(
        x + dirX * radius * 1.65 - normalX * radius * 0.48,
        y + dirY * radius * 1.65 - normalY * radius * 0.48,
    );
    skillLayer.closePath();
    skillLayer.fill({ color: 0xfef3c7, alpha: 0.82 });
}

function drawNinjaShurikenEffect(effect, frame) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    const size = Math.max(9, (effect.radius || 35) * frame.scale * 0.42);
    const tail = Math.max(28, size * 4.2);
    const spin = performance.now() / 48;
    skillLayer.moveTo(x - Math.cos(angle) * tail, y - Math.sin(angle) * tail).lineTo(x, y);
    skillLayer.stroke({ color: 0x7c3aed, width: Math.max(4, size * 0.82), alpha: 0.18 });
    skillLayer.moveTo(x - Math.cos(angle) * tail * 0.55, y - Math.sin(angle) * tail * 0.55).lineTo(x, y);
    skillLayer.stroke({ color: 0xc4b5fd, width: Math.max(2, size * 0.34), alpha: 0.58 });
    skillLayer.circle(x, y, size * 1.18);
    skillLayer.fill({ color: 0x8b5cf6, alpha: 0.1 });
    for (let i = 0; i < 4; i++) {
        const bladeAngle = angle + spin + Math.PI / 4 + (Math.PI / 2) * i;
        skillLayer.moveTo(x, y).lineTo(x + Math.cos(bladeAngle) * size, y + Math.sin(bladeAngle) * size);
    }
    skillLayer.stroke({ color: 0xf8fafc, width: Math.max(2, size * 0.4), alpha: 0.98 });
    for (let i = 0; i < 4; i++) {
        const bladeAngle = angle + spin + (Math.PI / 2) * i;
        skillLayer.moveTo(x, y).lineTo(x + Math.cos(bladeAngle) * size * 0.72, y + Math.sin(bladeAngle) * size * 0.72);
    }
    skillLayer.stroke({ color: 0x312e81, width: Math.max(1, size * 0.22), alpha: 0.8 });
    skillLayer.circle(x, y, Math.max(2, size * 0.24));
    skillLayer.fill({ color: 0x111827, alpha: 0.95 });
}

function drawNinjaERangeEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || 290) * frame.scale;
    const alpha = effectAlpha(effect);
    const progress = 1 - alpha;
    const rotation = performance.now() / 180;
    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0x312e81, alpha: 0.1 * alpha });
    skillLayer.circle(x, y, radius * (0.55 + progress * 0.45));
    skillLayer.stroke({ color: 0xf8fafc, width: Math.max(2, radius * 0.02), alpha: 0.34 * alpha });
    for (let i = 0; i < 6; i++) {
        const angle = rotation + (Math.PI * 2 * i) / 6;
        drawNinjaSlash(x, y, radius, angle, i % 2 === 0 ? 0xc4b5fd : 0x7c3aed, alpha);
    }
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0xc4b5fd, width: 3, alpha: 0.78 * alpha });
    skillLayer.circle(x, y, Math.max(7, radius * 0.08));
    skillLayer.fill({ color: 0xf8fafc, alpha: 0.16 * alpha });
}

function drawNinjaSlash(x, y, radius, angle, color, alpha) {
    const inner = radius * 0.28;
    const outer = radius * 0.92;
    const curve = 0.62;
    skillLayer.moveTo(x + Math.cos(angle - curve) * inner, y + Math.sin(angle - curve) * inner);
    skillLayer.quadraticCurveTo(
        x + Math.cos(angle) * outer,
        y + Math.sin(angle) * outer,
        x + Math.cos(angle + curve) * inner,
        y + Math.sin(angle + curve) * inner,
    );
    skillLayer.stroke({ color, width: Math.max(2, radius * 0.025), alpha: 0.72 * alpha });
}

function drawMagePrismaticBarrierEffect(effect, frame) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const normalX = -dirY;
    const normalY = dirX;
    const size = Math.max(22, (effect.radius || 55) * frame.scale * 0.82);
    const tail = Math.max(42, size * 3.4);
    const time = performance.now();

    skillLayer.moveTo(x - dirX * tail, y - dirY * tail);
    skillLayer.lineTo(x - dirX * size * 0.15, y - dirY * size * 0.15);
    skillLayer.stroke({ color: 0x67e8f9, width: size * 1.15, alpha: 0.18 });
    for (let i = -1; i <= 1; i++) {
        const offset = i * size * 0.28;
        skillLayer.moveTo(
            x - dirX * tail * (0.66 + Math.abs(i) * 0.1) + normalX * offset,
            y - dirY * tail * (0.66 + Math.abs(i) * 0.1) + normalY * offset,
        );
        skillLayer.lineTo(x - dirX * size * 0.35, y - dirY * size * 0.35);
        skillLayer.stroke({
            color: i < 0 ? 0x67e8f9 : i > 0 ? 0xf9a8d4 : 0xffffff,
            width: Math.max(2, size * 0.18),
            alpha: i === 0 ? 0.94 : 0.7,
        });
    }

    skillLayer.circle(x, y, size * 1.35);
    skillLayer.fill({ color: 0x22d3ee, alpha: 0.13 });
    skillLayer.moveTo(x - dirX * size * 0.95, y - dirY * size * 0.95);
    skillLayer.lineTo(x + dirX * size * 0.7, y + dirY * size * 0.7);
    skillLayer.stroke({ color: 0xfef3c7, width: Math.max(3, size * 0.16), alpha: 0.92 });
    skillLayer.moveTo(x + dirX * size * 1.35, y + dirY * size * 1.35);
    skillLayer.lineTo(x + normalX * size * 0.72, y + normalY * size * 0.72);
    skillLayer.lineTo(x - dirX * size * 0.25, y - dirY * size * 0.25);
    skillLayer.lineTo(x - normalX * size * 0.72, y - normalY * size * 0.72);
    skillLayer.closePath();
    skillLayer.fill({ color: 0xffffff, alpha: 0.9 });
    skillLayer.stroke({ color: 0x06b6d4, width: Math.max(3, size * 0.14), alpha: 0.98 });
    skillLayer.moveTo(x - normalX * size * 0.58, y - normalY * size * 0.58);
    skillLayer.lineTo(x + normalX * size * 0.58, y + normalY * size * 0.58);
    skillLayer.stroke({ color: 0xf472b6, width: Math.max(2, size * 0.11), alpha: 0.9 });

    for (let i = 0; i < 4; i++) {
        const angle = time / 170 + (Math.PI * 2 * i) / 4;
        const px = x + Math.cos(angle) * size * 1.12;
        const py = y + Math.sin(angle) * size * 0.62;
        skillLayer.circle(px, py, Math.max(3, size * 0.12));
        skillLayer.fill({ color: i % 2 ? 0xfde68a : 0xa5f3fc, alpha: 0.9 });
    }
    skillLayer.circle(x, y, size * 1.08);
    skillLayer.stroke({ color: 0xfacc15, width: 3, alpha: 0.72 });
}

function drawMageLucentSingularityOrbEffect(effect, frame) {
    const position = projectileDrawPosition(effect, { fromSnapshot: true });
    const x = frame.offsetX + position.x * frame.scale;
    const y = frame.offsetY + position.y * frame.scale;
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const normalX = -dirY;
    const normalY = dirX;
    const radius = Math.max(15, (effect.radius || 34) * frame.scale);
    const tail = Math.max(36, radius * 3.7);
    const rotation = performance.now() / 135;

    skillLayer.moveTo(x - dirX * tail, y - dirY * tail);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0xfacc15, width: radius * 1.2, alpha: 0.18 });
    skillLayer.moveTo(
        x - dirX * tail * 0.76 + normalX * radius * 0.32,
        y - dirY * tail * 0.76 + normalY * radius * 0.32,
    );
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0x67e8f9, width: Math.max(2, radius * 0.18), alpha: 0.58 });
    skillLayer.moveTo(x - dirX * tail * 0.62 - normalX * radius * 0.3, y - dirY * tail * 0.62 - normalY * radius * 0.3);
    skillLayer.lineTo(x, y);
    skillLayer.stroke({ color: 0xfde68a, width: Math.max(2, radius * 0.16), alpha: 0.62 });

    skillLayer.circle(x, y, radius * 1.3);
    skillLayer.fill({ color: 0xfacc15, alpha: 0.18 });
    for (let i = 0; i < 3; i++) {
        const angle = rotation + (Math.PI * 2 * i) / 3;
        const px = x + Math.cos(angle) * radius * 1.08;
        const py = y + Math.sin(angle) * radius * 0.68;
        skillLayer.circle(px, py, Math.max(2, radius * 0.13));
        skillLayer.fill({ color: i === 1 ? 0x67e8f9 : 0xfef3c7, alpha: 0.88 });
    }
    skillLayer.circle(x, y, radius * 0.72);
    skillLayer.fill({ color: 0xffffff, alpha: 0.94 });
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0xf59e0b, width: Math.max(2, radius * 0.13), alpha: 0.9 });
}

function drawMageLucentSingularityEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || 300) * frame.scale;
    const remaining = effectAlpha(effect);
    const alpha = Math.min(1, remaining * 5);
    const progress = 1 - remaining;
    const rotation = performance.now() / 360;
    const pulse = 0.5 + Math.sin(performance.now() / 190) * 0.5;

    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0xfacc15, alpha: 0.11 * alpha });
    skillLayer.circle(x, y, radius * 0.72);
    skillLayer.fill({ color: 0x67e8f9, alpha: 0.035 * alpha });

    for (let i = 0; i < 8; i++) {
        const angle = rotation + (Math.PI * 2 * i) / 8;
        const inner = radius * 0.2;
        const outer = radius * (0.78 + (i % 2) * 0.12);
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
        skillLayer.stroke({
            color: i % 2 ? 0xfde68a : 0x67e8f9,
            width: Math.max(2, radius * 0.012),
            alpha: 0.44 * alpha,
        });
    }

    for (let i = 0; i < 3; i++) {
        const start = -rotation * (i % 2 ? 1 : -1) + (Math.PI * 2 * i) / 3;
        const ringRadius = radius * (0.48 + i * 0.19);
        skillLayer.moveTo(x + Math.cos(start) * ringRadius, y + Math.sin(start) * ringRadius);
        skillLayer.arc(x, y, ringRadius, start, start + Math.PI * (0.92 + i * 0.08));
        skillLayer.stroke({
            color: i === 1 ? 0xffffff : i === 2 ? 0x67e8f9 : 0xfacc15,
            width: Math.max(3, radius * (0.016 + i * 0.003)),
            alpha: (0.82 - i * 0.1) * alpha,
        });
    }

    const warningRadius = radius * (0.92 - ((progress * 2.4) % 1) * 0.58);
    skillLayer.circle(x, y, warningRadius);
    skillLayer.stroke({ color: 0xffffff, width: Math.max(2, radius * 0.014), alpha: (0.34 + pulse * 0.3) * alpha });
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0xf59e0b, width: Math.max(4, radius * 0.018), alpha: 0.94 * alpha });
    drawStarPath(skillLayer, x, y, Math.max(16, radius * (0.12 + pulse * 0.018)), Math.max(7, radius * 0.052));
    skillLayer.fill({ color: 0xffffff, alpha: 0.62 * alpha });
    skillLayer.stroke({ color: 0x67e8f9, width: Math.max(2, radius * 0.012), alpha: 0.82 * alpha });
}

function drawMageLucentSingularityBurstEffect(effect, frame) {
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || 300) * frame.scale;
    const remaining = effectAlpha(effect);
    const progress = 1 - remaining;
    const alpha = Math.min(1, remaining * 2.8);
    const rotation = performance.now() / 150;

    skillLayer.circle(x, y, radius * (0.18 + progress * 0.86));
    skillLayer.fill({ color: 0xffffff, alpha: 0.18 * alpha });
    skillLayer.circle(x, y, radius * (0.28 + progress * 0.74));
    skillLayer.stroke({ color: 0xfacc15, width: Math.max(5, radius * 0.032), alpha: 0.98 * alpha });
    skillLayer.circle(x, y, radius * (0.12 + progress * 0.9));
    skillLayer.stroke({ color: 0x06b6d4, width: Math.max(3, radius * 0.018), alpha: 0.78 * alpha });
    drawMageRayBurst(x, y, radius * (0.52 + progress * 0.72), rotation, alpha);
    drawMageRayBurst(x, y, radius * (0.38 + progress * 0.58), -rotation * 0.7, alpha * 0.58);
    drawStarPath(skillLayer, x, y, radius * (0.18 + progress * 0.12), radius * (0.07 + progress * 0.05));
    skillLayer.fill({ color: 0xffffff, alpha: 0.86 * alpha });
}

function drawMageFinalSparkEffect(effect, frame) {
    const startX = frame.offsetX + effect.x * frame.scale;
    const startY = frame.offsetY + effect.y * frame.scale;
    const endX = frame.offsetX + (effect.endX || effect.x) * frame.scale;
    const endY = frame.offsetY + (effect.endY || effect.y) * frame.scale;
    const dx = endX - startX;
    const dy = endY - startY;
    const length = Math.hypot(dx, dy) || 1;
    const dirX = dx / length;
    const dirY = dy / length;
    const normalX = -dirY;
    const normalY = dirX;
    const width = Math.max(8, (effect.width || 200) * frame.scale);
    const remaining = effectAlpha(effect);
    const progress = 1 - remaining;
    const alpha = Math.min(1, remaining * 2.5);
    const time = performance.now();

    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0xfacc15, width: width * 1.48, alpha: 0.11 * alpha });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x67e8f9, width: width, alpha: 0.24 * alpha });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0xfef3c7, width: Math.max(5, width * 0.5), alpha: 0.8 * alpha });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0xffffff, width: Math.max(3, width * 0.19), alpha: 0.98 * alpha });

    for (const side of [-1, 1]) {
        const offset = side * width * 0.46;
        skillLayer.moveTo(startX + normalX * offset, startY + normalY * offset);
        skillLayer.lineTo(endX + normalX * offset, endY + normalY * offset);
        skillLayer.stroke({
            color: side < 0 ? 0x22d3ee : 0xfacc15,
            width: Math.max(2, width * 0.045),
            alpha: 0.82 * alpha,
        });
    }

    for (let i = 0; i < 13; i++) {
        const travel = (i / 13 + ((time / 420) % 1)) % 1;
        const side = ((i % 3) - 1) * width * 0.2;
        const segment = Math.max(18, width * (0.55 + (i % 2) * 0.2));
        const px = startX + dx * travel + normalX * side;
        const py = startY + dy * travel + normalY * side;
        skillLayer.moveTo(px - dirX * segment * 0.5, py - dirY * segment * 0.5);
        skillLayer.lineTo(px + dirX * segment * 0.5, py + dirY * segment * 0.5);
        skillLayer.stroke({
            color: i % 3 === 0 ? 0x67e8f9 : 0xffffff,
            width: Math.max(2, width * 0.035),
            alpha: 0.58 * alpha,
        });
    }

    const startRadius = width * (0.48 + progress * 0.18);
    skillLayer.circle(startX, startY, startRadius);
    skillLayer.fill({ color: 0xffffff, alpha: 0.24 * alpha });
    skillLayer.circle(startX, startY, startRadius * 1.22);
    skillLayer.stroke({ color: 0xfacc15, width: Math.max(3, width * 0.04), alpha: 0.74 * alpha });
    drawMageRayBurst(endX, endY, width * (0.72 + progress * 0.45), time / 180, alpha);
    skillLayer.circle(endX, endY, width * (0.18 + progress * 0.25));
    skillLayer.stroke({ color: 0xffffff, width: Math.max(3, width * 0.05), alpha: 0.82 * alpha });
}

function drawMageRayBurst(x, y, radius, rotation, alpha) {
    for (let i = 0; i < 12; i++) {
        const angle = rotation + (Math.PI * 2 * i) / 12;
        const inner = radius * (i % 2 ? 0.12 : 0.2);
        const outer = radius * (i % 3 === 0 ? 1 : 0.72);
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle) * outer, y + Math.sin(angle) * outer);
        skillLayer.stroke({
            color: i % 2 ? 0xfacc15 : 0x67e8f9,
            width: Math.max(2, radius * (i % 3 === 0 ? 0.035 : 0.02)),
            alpha: (i % 3 === 0 ? 0.72 : 0.44) * alpha,
        });
    }
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
    const pulse = 0.5 + Math.sin(performance.now() / 170) * 0.5;
    const angle = Math.atan2(effect.dirY || 0, effect.dirX || 1);
    skillLayer.circle(x, y, radius * (1.45 + pulse * 0.18));
    skillLayer.fill({ color: 0x111827, alpha: 0.2 });
    drawNinjaSmoke(x, y, radius, pulse);
    skillLayer.circle(x, y, radius * 0.96);
    skillLayer.fill({ color: 0x111827, alpha: 0.86 });
    skillLayer.circle(x, y, radius * 1.42);
    skillLayer.stroke({ color: 0x7c3aed, width: 2, alpha: 0.76 });
    skillLayer
        .moveTo(x + Math.cos(angle) * radius * 1.45, y + Math.sin(angle) * radius * 1.45)
        .lineTo(x + Math.cos(angle + 2.35) * radius, y + Math.sin(angle + 2.35) * radius)
        .lineTo(x - Math.cos(angle) * radius * 0.55, y - Math.sin(angle) * radius * 0.55)
        .lineTo(x + Math.cos(angle - 2.35) * radius, y + Math.sin(angle - 2.35) * radius)
        .closePath();
    skillLayer.stroke({ color: 0xe9d5ff, width: 2, alpha: 0.62 });
    drawNinjaShadowTimer(effect, x, y, radius * 1.75);
}

function drawNinjaSmoke(x, y, radius, pulse) {
    for (let i = 0; i < 5; i++) {
        const angle = performance.now() / 550 + (Math.PI * 2 * i) / 5;
        const distance = radius * (0.9 + (i % 2) * 0.28 + pulse * 0.12);
        skillLayer.circle(
            x + Math.cos(angle) * distance,
            y + Math.sin(angle) * distance * 0.72,
            radius * (0.2 + (i % 3) * 0.05),
        );
        skillLayer.fill({ color: 0x4c1d95, alpha: 0.22 });
    }
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
    const traveled = clamp(Math.max(0, interpolatedTick() - (effect.createdAt || 0)) * effect.speed, 0, length);
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
        : clamp((tick - (effect.createdAt || tick)) / Math.max(1, arriveTick - (effect.createdAt || tick)), 0, 1);
    const worldX = (effect.x || 0) + ((effect.endX || effect.x || 0) - (effect.x || 0)) * progress;
    const worldY = (effect.y || 0) + ((effect.endY || effect.y || 0) - (effect.y || 0)) * progress;
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
        .lineTo(x + Math.cos(angle + 2.45) * size, y + Math.sin(angle + 2.45) * size)
        .lineTo(x + Math.cos(angle + Math.PI) * size * 0.35, y + Math.sin(angle + Math.PI) * size * 0.35)
        .lineTo(x + Math.cos(angle - 2.45) * size, y + Math.sin(angle - 2.45) * size)
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
        .moveTo(x + Math.cos(angle) * length * 0.5, y + Math.sin(angle) * length * 0.5)
        .lineTo(x - Math.cos(angle) * length * 0.5, y - Math.sin(angle) * length * 0.5);
    skillLayer.stroke({ color: shaftColor, width: 3, alpha: 0.95 });
    skillLayer
        .moveTo(x + Math.cos(angle) * length * 0.5, y + Math.sin(angle) * length * 0.5)
        .lineTo(x + Math.cos(angle + Math.PI * 0.82) * width, y + Math.sin(angle + Math.PI * 0.82) * width)
        .lineTo(x + Math.cos(angle - Math.PI * 0.82) * width, y + Math.sin(angle - Math.PI * 0.82) * width)
        .closePath();
    skillLayer.fill({ color: headColor, alpha: 0.95 });
}

function projectileDrawPosition(effect, options = {}) {
    const tick = interpolatedTick();
    const baseTick = options.fromSnapshot ? state.snapshotTick : (effect.createdAt ?? tick);
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
    const t = Math.max(0, Math.min(1, ((point.x - start.x) * dx + (point.y - start.y) * dy) / lengthSquared));
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
                x: (effect.x || 0) + (effect.dirX || 1) * arrow.forward - (effect.dirY || 0) * arrow.side,
                y: (effect.y || 0) + (effect.dirY || 0) * arrow.forward + (effect.dirX || 1) * arrow.side,
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
    if (!self || self.heroId !== 'sword') {
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
        skillLayer.stroke({ color: 0x0f172a, width: 5, alpha: 0.18 });
        skillLayer.moveTo(startX, startY);
        skillLayer.arc(x, y, radius, startAngle, endAngle);
        skillLayer.stroke({ color: 0x38bdf8, width: 4, alpha: 0.9 });
        skillLayer.moveTo(x - radius * 0.42, y + radius * 0.42);
        skillLayer.lineTo(x + radius * 0.42, y - radius * 0.42);
        skillLayer.stroke({ color: 0xe0f2fe, width: 2, alpha: 0.46 });
    }
}

function drawNinjaPassiveCooldowns(frame) {
    const self = state.players.get(state.playerId);
    if (!self || self.heroId !== 'ninja') {
        return;
    }
    const targetUntil = self.passive?.ninjaSoulCooldowns || {};
    const tick = interpolatedTick();
    const targets = targetMap();
    const cooldownTicks = (skillClientConfig.ninja_passive?.heroCooldownSeconds || 10) * state.tickRate;
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
    const durationTicks = (skillClientConfig.fire_mage_passive?.explosionDelaySeconds || 2) * state.tickRate;
    for (const target of targetMap().values()) {
        if (!target || target.dead) {
            continue;
        }
        const burn = (target.buffs || []).find(
            buff => buff.id?.startsWith('fire_mage_blaze:') && (buff.explosionAtTick || 0) > tick,
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

/**
 * 绘制剑客第三段 Q 向前推进的龙卷风，包含漏斗风环、旋转剑气和卷入的风尘。
 * @param {object} effect 服务端同步的龙卷投射物位置、方向与范围。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawSwordWhirlwindEffect(effect, frame) {
    const tick = interpolatedTick();
    const ageTicks = Math.max(0, tick - (effect.createdAt || tick));
    const traveled = clamp(ageTicks * (effect.speed || 0), 0, effect.range || 0);
    const x = effect.x + (effect.dirX || 0) * traveled;
    const y = effect.y + (effect.dirY || 0) * traveled;
    const sx = frame.offsetX + x * frame.scale;
    const sy = frame.offsetY + y * frame.scale;
    const radius = (effect.radius || 70) * frame.scale;
    const rotation = performance.now() / 75;
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const sideX = -dirY;
    const sideY = dirX;
    const tail = Math.max(36, radius * 3.6);
    skillLayer
        .moveTo(sx - dirX * tail, sy - dirY * tail)
        .lineTo(sx + sideX * radius * 0.9, sy + sideY * radius * 0.9)
        .lineTo(sx + dirX * radius * 0.75, sy + dirY * radius * 0.75)
        .lineTo(sx - sideX * radius * 0.9, sy - sideY * radius * 0.9)
        .closePath();
    skillLayer.fill({ color: 0x38bdf8, alpha: 0.1 });

    for (let i = 0; i < 5; i++) {
        const depth = i / 5;
        const cx = sx - dirX * tail * depth * 0.72;
        const cy = sy - dirY * tail * depth * 0.72;
        const ringRadius = radius * (1 - depth * 0.62);
        const start = rotation + i * 1.1;
        skillLayer.moveTo(cx + Math.cos(start) * ringRadius, cy + Math.sin(start) * ringRadius);
        skillLayer.arc(cx, cy, ringRadius, start, start + Math.PI * 1.45);
        skillLayer.stroke({
            color: i % 2 ? 0xffffff : 0x7dd3fc,
            width: Math.max(3, radius * (0.16 - i * 0.018)),
            alpha: 0.9 - i * 0.1,
        });
    }

    for (let i = 0; i < 12; i++) {
        const particleDepth = ((i / 12 + performance.now() / 900) % 1) * tail;
        const spread = Math.sin(rotation + i * 1.8) * radius * (0.25 + particleDepth / tail);
        const px = sx - dirX * particleDepth + sideX * spread;
        const py = sy - dirY * particleDepth + sideY * spread;
        skillLayer.circle(px, py, Math.max(2, radius * (0.055 + (i % 3) * 0.018)));
        skillLayer.fill({ color: i % 2 ? 0xffffff : 0xbae6fd, alpha: 0.5 });
    }
    skillLayer.circle(sx, sy, Math.max(8, radius * 0.3));
    skillLayer.fill({ color: 0xf8fafc, alpha: 0.42 });
}

/**
 * 绘制剑客前两段 Q 向前扫出的月牙剑气，横向覆盖技能宽度且锋尖走完整段伤害范围。
 * @param {object} effect 服务端同步的剑气起终点、宽度与持续时间。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawSwordQEffect(effect, frame) {
    const alpha = Math.min(1, effectAlpha(effect) * 3.2);
    const progress = 1 - effectAlpha(effect);
    const startX = frame.offsetX + effect.x * frame.scale;
    const startY = frame.offsetY + effect.y * frame.scale;
    const endX = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const endY = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const dx = endX - startX;
    const dy = endY - startY;
    const len = Math.hypot(dx, dy) || 1;
    const halfWidth = Math.max(9, (effect.width || 55) * frame.scale);
    const travel = progress * (2 - progress);
    const angle = Math.atan2(dy, dx);
    const sweep = 1.08;
    const radius = halfWidth / Math.sin(sweep);
    const frontX = startX + dx * travel;
    const frontY = startY + dy * travel;
    const x = frontX - Math.cos(angle) * radius;
    const y = frontY - Math.sin(angle) * radius;

    skillLayer.moveTo(x + Math.cos(angle - sweep) * radius * 1.04, y + Math.sin(angle - sweep) * radius * 1.04);
    skillLayer.arc(x, y, radius * 1.04, angle - sweep, angle + sweep);
    skillLayer.stroke({ color: 0xcbd5e1, width: Math.max(5, radius * 0.28), alpha: 0.24 * alpha });
    skillLayer.moveTo(x + Math.cos(angle - sweep) * radius, y + Math.sin(angle - sweep) * radius);
    skillLayer.arc(x, y, radius, angle - sweep, angle + sweep);
    skillLayer.stroke({ color: 0xf8fafc, width: Math.max(3, radius * 0.15), alpha: 0.96 * alpha });
    skillLayer.moveTo(x + Math.cos(angle - sweep) * radius * 0.78, y + Math.sin(angle - sweep) * radius * 0.78);
    skillLayer.arc(x, y, radius * 0.78, angle - sweep, angle + sweep);
    skillLayer.stroke({ color: 0xffffff, width: Math.max(2, radius * 0.055), alpha: 0.78 * alpha });
}

/**
 * 绘制剑客 EQ 连招触发的环形剑气，表现贴身旋转出刀。
 * @param {object} effect 服务端同步的圆斩中心与范围。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawSwordQCircleEffect(effect, frame) {
    const alpha = Math.min(1, effectAlpha(effect) * 2.2);
    const progress = 1 - effectAlpha(effect);
    const x = frame.offsetX + effect.x * frame.scale;
    const y = frame.offsetY + effect.y * frame.scale;
    const radius = (effect.radius || effect.range || 375) * frame.scale;
    const rotation = performance.now() / 95;
    skillLayer.circle(x, y, radius * (0.6 + progress * 0.4));
    skillLayer.fill({ color: 0x0ea5e9, alpha: 0.06 * alpha });
    for (let i = 0; i < 4; i++) {
        const start = rotation + (Math.PI * 2 * i) / 4;
        skillLayer.moveTo(x + Math.cos(start) * radius * 0.28, y + Math.sin(start) * radius * 0.28);
        skillLayer.arc(x, y, radius * (0.46 + i * 0.12), start, start + Math.PI * 0.82);
        skillLayer.stroke({
            color: i % 2 ? 0xe0f2fe : 0x38bdf8,
            width: Math.max(3, radius * 0.055),
            alpha: (0.78 - i * 0.1) * alpha,
        });
    }
    skillLayer.circle(x, y, Math.max(8, radius * 0.08));
    skillLayer.fill({ color: 0xf0f9ff, alpha: 0.18 * alpha });
}

/**
 * 绘制剑客 E 位移期间贴地掠过的风痕、流线和终点风环。
 * @param {object} effect 服务端同步的位移起终点与持续时间。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawSwordEEffect(effect, frame) {
    const alpha = effectAlpha(effect);
    const startX = frame.offsetX + effect.x * frame.scale;
    const startY = frame.offsetY + effect.y * frame.scale;
    const endX = frame.offsetX + (effect.endX || effect.x) * frame.scale;
    const endY = frame.offsetY + (effect.endY || effect.y) * frame.scale;
    const dx = endX - startX;
    const dy = endY - startY;
    const len = Math.hypot(dx, dy) || 1;
    const nx = -dy / len;
    const ny = dx / len;
    const progress = 1 - alpha;
    const gustWidth = Math.max(12, (effect.radius || 36) * frame.scale * 0.7);
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x38bdf8, width: gustWidth * 1.7, alpha: 0.1 * alpha });
    for (let i = -3; i <= 3; i++) {
        const offset = i * gustWidth * 0.42;
        const wave = Math.sin(progress * Math.PI * 4 + i * 1.3) * gustWidth * 0.32;
        skillLayer.moveTo(startX + nx * offset, startY + ny * offset);
        skillLayer.lineTo(endX + nx * (offset * 0.2 + wave), endY + ny * (offset * 0.2 + wave));
        skillLayer.stroke({
            color: i === 0 ? 0xffffff : i % 2 ? 0x7dd3fc : 0x38bdf8,
            width: Math.max(2, gustWidth * (i === 0 ? 0.2 : 0.11)),
            alpha: (0.86 - Math.abs(i) * 0.12) * alpha,
        });
    }
    for (let i = 0; i < 8; i++) {
        const t = (i / 8 + progress) % 1;
        const px = startX + dx * t + nx * Math.sin(i * 2.1 + progress * 7) * gustWidth;
        const py = startY + dy * t + ny * Math.sin(i * 2.1 + progress * 7) * gustWidth;
        skillLayer.circle(px, py, Math.max(2, gustWidth * (0.12 + (i % 2) * 0.05)));
        skillLayer.fill({ color: i % 2 ? 0xffffff : 0xbae6fd, alpha: 0.46 * alpha });
    }
    skillLayer.circle(endX, endY, gustWidth * (0.75 + progress * 0.45));
    skillLayer.stroke({ color: 0xe0f2fe, width: 3, alpha: 0.72 * alpha });
}

/**
 * 绘制剑客 R 围绕单个目标收束的旋风斩、短剑光与中心风眼。
 * @param {object} effect 服务端同步的大招中心、范围与命中数量。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawSwordREffect(effect, frame) {
    const alpha = Math.min(1, effectAlpha(effect) * 2);
    const x = frame.offsetX + (effect.endX ?? effect.x) * frame.scale;
    const y = frame.offsetY + (effect.endY ?? effect.y) * frame.scale;
    const radius = Math.min(effect.radius || 450, 180) * frame.scale;
    const progress = 1 - effectAlpha(effect);
    const rotation = performance.now() / 68;
    for (let ring = 0; ring < 3; ring++) {
        const ringRadius = radius * (0.42 + ring * 0.22) * (0.82 + progress * 0.18);
        const start = rotation * (ring % 2 ? -1 : 1) + ring * 1.1;
        skillLayer.moveTo(x + Math.cos(start) * ringRadius, y + Math.sin(start) * ringRadius);
        skillLayer.arc(x, y, ringRadius, start, start + Math.PI * (1.05 + ring * 0.08));
        skillLayer.stroke({
            color: ring % 2 ? 0xffffff : 0x38bdf8,
            width: Math.max(3, radius * (0.04 - ring * 0.006)),
            alpha: (0.82 - ring * 0.12) * alpha,
        });
    }
    const blades = 6;
    for (let i = 0; i < blades; i++) {
        const angle = rotation + (Math.PI * 2 * i) / blades + progress * 1.4;
        const inner = radius * (0.18 + (i % 2) * 0.1);
        const outer = radius * (0.56 + (i % 3) * 0.1);
        skillLayer.moveTo(x + Math.cos(angle) * inner, y + Math.sin(angle) * inner);
        skillLayer.lineTo(x + Math.cos(angle + 0.2) * outer, y + Math.sin(angle + 0.2) * outer);
        skillLayer.stroke({
            color: i % 2 ? 0xe0f2fe : 0x38bdf8,
            width: Math.max(2, radius * 0.022),
            alpha: (0.76 - (i % 3) * 0.1) * alpha,
        });
    }
    const finisherAngle = rotation + 0.6;
    skillLayer.moveTo(x - Math.cos(finisherAngle) * radius * 0.68, y - Math.sin(finisherAngle) * radius * 0.68);
    skillLayer.lineTo(x + Math.cos(finisherAngle) * radius * 0.68, y + Math.sin(finisherAngle) * radius * 0.68);
    skillLayer.stroke({ color: 0xffffff, width: Math.max(4, radius * 0.03), alpha: 0.86 * alpha });
    skillLayer.circle(x, y, Math.max(12, radius * 0.1));
    skillLayer.fill({ color: 0xf8fafc, alpha: 0.28 * alpha });
}

/**
 * 绘制石头人 Q 的高速棱角岩片与碎石尾迹。
 * @param {object} effect 服务端同步的投射物效果。
 * @param {object} frame 当前世界到屏幕的变换参数。
 */
function drawTankShardEffect(effect, frame) {
    const tickDelta = clamp(interpolatedTick() - Number(els.tick.textContent || 0), 0, 1);
    const smoothX = effect.x + (effect.dirX || 0) * (effect.speed || 0) * tickDelta;
    const smoothY = effect.y + (effect.dirY || 0) * (effect.speed || 0) * tickDelta;
    const sx = frame.offsetX + smoothX * frame.scale;
    const sy = frame.offsetY + smoothY * frame.scale;
    const radius = Math.max(5, (effect.radius || 45) * frame.scale);
    const dirX = effect.dirX || 1;
    const dirY = effect.dirY || 0;
    const sideX = -dirY;
    const sideY = dirX;
    for (let index = 3; index >= 1; index--) {
        const distance = radius * (1.15 + index * 0.72);
        const debrisSize = Math.max(2, radius * (0.18 - index * 0.025));
        const debrisX = sx - dirX * distance + sideX * (index % 2 ? radius * 0.34 : -radius * 0.28);
        const debrisY = sy - dirY * distance + sideY * (index % 2 ? radius * 0.34 : -radius * 0.28);
        skillLayer.circle(debrisX, debrisY, debrisSize);
        skillLayer.fill({ color: 0xa8a29e, alpha: 0.55 - index * 0.08 });
    }
    skillLayer
        .moveTo(sx + dirX * radius * 1.15, sy + dirY * radius * 1.15)
        .lineTo(sx + sideX * radius * 0.82, sy + sideY * radius * 0.82)
        .lineTo(sx - dirX * radius * 0.95 + sideX * radius * 0.34, sy - dirY * radius * 0.95 + sideY * radius * 0.34)
        .lineTo(sx - dirX * radius * 0.72 - sideX * radius * 0.7, sy - dirY * radius * 0.72 - sideY * radius * 0.7)
        .lineTo(sx + dirX * radius * 0.25 - sideX * radius * 0.92, sy + dirY * radius * 0.25 - sideY * radius * 0.92)
        .closePath();
    skillLayer.fill({ color: 0x78716c, alpha: 0.98 });
    skillLayer.stroke({ color: 0x292524, width: 3, alpha: 0.92 });
    skillLayer
        .moveTo(sx + dirX * radius * 0.7, sy + dirY * radius * 0.7)
        .lineTo(sx - dirX * radius * 0.4 + sideX * radius * 0.2, sy - dirY * radius * 0.4 + sideY * radius * 0.2);
    skillLayer.stroke({ color: 0xfbbf24, width: 2, alpha: 0.78 });
}
