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
}

/** 绘制技能蓄力提示；全图技能按当前地图尺寸展示世界范围。 */
function drawCastWindup(windup, frame, now) {
    const progress = clamp((now - windup.startedAt) / Math.max(1, windup.durationMs || 1), 0, 1);
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
    skillLayer.moveTo(x + Math.cos(angleStart) * (pulseRadius + 6), y + Math.sin(angleStart) * (pulseRadius + 6));
    skillLayer.arc(x, y, pulseRadius + 6, angleStart, angleStart + Math.PI * 2 * progress);
    skillLayer.stroke({ color, width: 4, alpha: 0.85 });

    if (windup.skillId === 'taunt') {
        drawCircleWindup(windup, frame, color, alpha, windup.range || 400);
        return;
    }
    if (windup.skillId === 'justice') {
        drawTargetLockWindup(windup, frame, color, alpha);
        return;
    }
    if (windup.skillId === 'killer_q') {
        drawTargetLockWindup(windup, frame, color, alpha);
        return;
    }
    if (windup.skillId === 'ninja_q') {
        drawNinjaQWindup(windup, frame, color, alpha);
        return;
    }
    if (windup.skillId === 'arrow_rain') {
        drawDirectionalWindup(
            { ...windup, range: Math.hypot(state.map.width || 8000, state.map.height || 8000) },
            frame,
            color,
            alpha,
            10,
        );
        return;
    }
    if (windup.skillId === 'explorer_e') {
        drawDirectionalWindup(windup, frame, color, alpha, 14);
        return;
    }
    if (windup.skillId === 'explorer_q') {
        drawDirectionalWindup(windup, frame, color, alpha, 10);
        return;
    }
    if (windup.skillId === 'explorer_r') {
        drawDirectionalWindup(
            { ...windup, range: Math.hypot(state.map.width || 8000, state.map.height || 8000) },
            frame,
            color,
            alpha,
            34,
        );
        return;
    }
    if (windup.skillId === 'mage_q') {
        drawMageLinearWindup(windup, frame, alpha, 0xfacc15, 0x67e8f9, 34);
        return;
    }
    if (windup.skillId === 'mage_w') {
        drawMageLinearWindup(windup, frame, alpha, 0x67e8f9, 0xf9a8d4, 42);
        return;
    }
    if (windup.skillId === 'mage_e') {
        drawMageLucentSingularityWindup(windup, frame, alpha);
        return;
    }
    if (windup.skillId === 'mage_r') {
        drawMageFinalSparkWindup(windup, frame, alpha);
        return;
    }
}

function drawMageLinearWindup(windup, frame, alpha, primary, accent, widthWorld) {
    const range = windup.range || 1100;
    const startX = frame.offsetX + windup.x * frame.scale;
    const startY = frame.offsetY + windup.y * frame.scale;
    const endX = frame.offsetX + (windup.x + (windup.dirX || 1) * range) * frame.scale;
    const endY = frame.offsetY + (windup.y + (windup.dirY || 0) * range) * frame.scale;
    const normalX = -(windup.dirY || 0);
    const normalY = windup.dirX || 1;
    const width = Math.max(7, widthWorld * frame.scale);
    const progress = clamp((performance.now() - windup.startedAt) / Math.max(1, windup.durationMs || 1), 0, 1);

    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: primary, width, alpha: 0.12 * alpha });
    for (const side of [-1, 1]) {
        const offset = side * width * 0.48;
        skillLayer.moveTo(startX + normalX * offset, startY + normalY * offset);
        skillLayer.lineTo(endX + normalX * offset, endY + normalY * offset);
        skillLayer.stroke({
            color: side < 0 ? accent : primary,
            width: Math.max(1, width * 0.12),
            alpha: 0.58 * alpha,
        });
    }
    const chargeX = startX + (endX - startX) * progress;
    const chargeY = startY + (endY - startY) * progress;
    skillLayer.circle(chargeX, chargeY, Math.max(4, width * 0.26));
    skillLayer.fill({ color: 0xffffff, alpha: 0.68 * alpha });
}

function drawMageLucentSingularityWindup(windup, frame, alpha) {
    const config = skillClientConfig.mage_e || {};
    const x = frame.offsetX + windup.targetX * frame.scale;
    const y = frame.offsetY + windup.targetY * frame.scale;
    const radius = (config.radius || 310) * frame.scale;
    const progress = clamp((performance.now() - windup.startedAt) / Math.max(1, windup.durationMs || 1), 0, 1);
    const rotation = performance.now() / 220;

    skillLayer.circle(x, y, radius);
    skillLayer.fill({ color: 0xfacc15, alpha: 0.05 * alpha });
    skillLayer.circle(x, y, radius);
    skillLayer.stroke({ color: 0xf59e0b, width: 2, alpha: 0.64 * alpha });
    for (let i = 0; i < 3; i++) {
        const start = rotation + (Math.PI * 2 * i) / 3;
        const ringRadius = radius * (0.34 + progress * 0.48 + i * 0.05);
        skillLayer.moveTo(x + Math.cos(start) * ringRadius, y + Math.sin(start) * ringRadius);
        skillLayer.arc(x, y, ringRadius, start, start + Math.PI * 1.05);
        skillLayer.stroke({
            color: i === 1 ? 0x67e8f9 : 0xfacc15,
            width: Math.max(2, radius * 0.012),
            alpha: (0.7 - i * 0.12) * alpha,
        });
    }
    skillLayer.circle(x, y, Math.max(7, radius * 0.06));
    skillLayer.fill({ color: 0xffffff, alpha: (0.25 + progress * 0.42) * alpha });
}

function drawNinjaQWindup(windup, frame, color, alpha) {
    drawDirectionalWindup(windup, frame, color, alpha, 10);
    const self = state.players.get(state.playerId);
    if (!self) {
        return;
    }
    const tick = interpolatedTick();
    if ((self.ninja?.shadowReadyTick || 0) <= tick) {
        drawNinjaQShadowWindup(
            self.ninja?.shadowX,
            self.ninja?.shadowY,
            self.ninja?.shadowExpiresAt,
            windup,
            frame,
            color,
            alpha,
            tick,
        );
    }
    drawNinjaQShadowWindup(
        self.ninja?.rShadowX,
        self.ninja?.rShadowY,
        self.ninja?.rShadowExpiresAt,
        windup,
        frame,
        color,
        alpha,
        tick,
    );
}

function drawNinjaQShadowWindup(x, y, expiresAt, windup, frame, color, alpha, tick) {
    if (!expiresAt || expiresAt <= tick) {
        return;
    }
    const dx = windup.targetX - x;
    const dy = windup.targetY - y;
    const len = Math.hypot(dx, dy) || 1;
    drawDirectionalWindup({ ...windup, x, y, dirX: dx / len, dirY: dy / len }, frame, color, alpha, 10);
}

function castWindupOrigin(windup) {
    return windup;
}

function drawDirectionalWindup(windup, frame, color, alpha, width) {
    const range = windup.range || 475;
    const x = frame.offsetX + windup.x * frame.scale;
    const y = frame.offsetY + windup.y * frame.scale;
    const endX = frame.offsetX + (windup.x + (windup.dirX || 1) * range) * frame.scale;
    const endY = frame.offsetY + (windup.y + (windup.dirY || 0) * range) * frame.scale;
    skillLayer.moveTo(x, y);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color, width, alpha: 0.16 * alpha });
    skillLayer.moveTo(x, y);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color, width: 2, alpha: 0.72 * alpha });
}

function drawMageFinalSparkWindup(windup, frame, alpha) {
    const config = skillClientConfig.mage_r || {};
    const range = windup.range || 3400;
    const startX = frame.offsetX + windup.x * frame.scale;
    const startY = frame.offsetY + windup.y * frame.scale;
    const endX = frame.offsetX + (windup.x + (windup.dirX || 1) * range) * frame.scale;
    const endY = frame.offsetY + (windup.y + (windup.dirY || 0) * range) * frame.scale;
    const dx = endX - startX;
    const dy = endY - startY;
    const length = Math.hypot(dx, dy) || 1;
    const normalX = -dy / length;
    const normalY = dx / length;
    const width = Math.max(12, (config.beamWidth || 200) * frame.scale);
    const progress = clamp((performance.now() - windup.startedAt) / Math.max(1, windup.durationMs || 1), 0, 1);
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0xfacc15, width, alpha: (0.08 + progress * 0.08) * alpha });
    skillLayer.moveTo(startX, startY);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0xfef3c7, width: Math.max(3, width * 0.16), alpha: 0.5 * alpha });
    for (const side of [-1, 1]) {
        const offset = side * width * 0.5;
        skillLayer.moveTo(startX + normalX * offset, startY + normalY * offset);
        skillLayer.lineTo(endX + normalX * offset, endY + normalY * offset);
        skillLayer.stroke({
            color: side < 0 ? 0x67e8f9 : 0xfacc15,
            width: Math.max(2, width * 0.025),
            alpha: 0.66 * alpha,
        });
    }
    skillLayer.circle(startX, startY, width * (0.2 + progress * 0.3));
    skillLayer.stroke({ color: 0xffffff, width: Math.max(2, width * 0.035), alpha: 0.76 * alpha });
    skillLayer.circle(startX, startY, Math.max(5, width * 0.08));
    skillLayer.fill({ color: 0xffffff, alpha: (0.22 + progress * 0.52) * alpha });
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
    if (skillId === 'justice') {
        return 0xf97316;
    }
    if (skillId === 'slam') {
        return 0x8b5e34;
    }
    if (skillId === 'taunt') {
        return 0x64748b;
    }
    if (skillId === 'arrow_rain') {
        return 0xa78bfa;
    }
    if (skillId === 'explorer_q' || skillId === 'explorer_e' || skillId === 'explorer_r') {
        return 0x38bdf8;
    }
    if (skillId === 'killer_q') {
        return 0x8b5cf6;
    }
    return 0x38bdf8;
}

function interpolatedTick() {
    if (!state.snapshotAtMs) {
        return Number(els.tick.textContent || 0);
    }
    return state.snapshotTick + ((performance.now() - state.snapshotAtMs) / 1000) * state.tickRate;
}

function showTankEPreview(self) {
    const config = skillClientConfig.taunt || {};
    const previewMs = config.previewMs || 450;
    state.skillPreview = {
        kind: 'tank_e',
        form: 'circle',
        x: self.x,
        y: self.y,
        range: config.range || 400,
        previewMs,
        expiresAt: performance.now() + previewMs,
    };
}

function showFrostMageRPreview(self) {
    const config = skillClientConfig.frostmage_r || {};
    const previewMs = config.previewMs || 450;
    state.skillPreview = {
        kind: 'frostmage_r',
        form: 'circle',
        x: self.x,
        y: self.y,
        range: config.range || 550,
        previewMs,
        expiresAt: performance.now() + previewMs,
    };
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
    const alpha = Math.max(0, (preview.expiresAt - performance.now()) / (preview.previewMs || 450));
    const x = frame.offsetX + preview.x * frame.scale;
    const y = frame.offsetY + preview.y * frame.scale;
    if (preview.form === 'circle') {
        skillLayer.circle(x, y, preview.range * frame.scale);
        skillLayer.stroke({ color: 0x38bdf8, width: 3, alpha: 0.65 * alpha });
        skillLayer.circle(x, y, 12);
        skillLayer.fill({ color: 0x38bdf8, alpha: 0.18 * alpha });
        return;
    }
    const endX = frame.offsetX + (preview.x + preview.dirX * preview.range) * frame.scale;
    const endY = frame.offsetY + (preview.y + preview.dirY * preview.range) * frame.scale;
    if (preview.form === 'whirlwind') {
        skillLayer.moveTo(x, y);
        skillLayer.lineTo(endX, endY);
        skillLayer.stroke({ color: 0x38bdf8, width: 2, alpha: 0.55 * alpha });
        skillLayer.circle(endX, endY, (preview.radius || 70) * frame.scale);
        skillLayer.stroke({ color: 0x0284c7, width: 3, alpha: 0.85 * alpha });
        skillLayer.circle(endX, endY, Math.max(6, (preview.radius || 70) * frame.scale * 0.28));
        skillLayer.fill({ color: 0x38bdf8, alpha: 0.16 * alpha });
        return;
    }
    const width = preview.form === 'whirlwind' ? 18 : 12;
    skillLayer.moveTo(x, y);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x38bdf8, width, alpha: 0.28 * alpha });
    skillLayer.moveTo(x, y);
    skillLayer.lineTo(endX, endY);
    skillLayer.stroke({ color: 0x0284c7, width: 2, alpha: 0.8 * alpha });
}
