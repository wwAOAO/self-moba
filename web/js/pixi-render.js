function draw(ticker) {
    syncSpritePositions(ticker.deltaMS);
    const frame = calculateFrame();
    drawMap(frame);
    drawEffects(frame);
    syncUnits(frame, ticker.deltaMS);
    syncSprites(frame);
    syncDamageTexts(frame, ticker.deltaMS);
    drawMinimap();
}

function drawMinimap() {
    const canvas = els.minimap;
    const size = Math.floor(canvas.clientWidth * window.devicePixelRatio);
    if (!size) {
        return;
    }
    if (canvas.width !== size || canvas.height !== size) {
        canvas.width = size;
        canvas.height = size;
    }
    const context = canvas.getContext('2d');
    const scale = size / Math.max(state.map.width, state.map.height);
    const point = (entity, radius, color) => {
        context.beginPath();
        context.arc(entity.x * scale, entity.y * scale, radius, 0, Math.PI * 2);
        context.fillStyle = color;
        context.fill();
    };

    context.clearRect(0, 0, size, size);
    context.fillStyle = '#15251d';
    context.fillRect(0, 0, size, size);
    context.strokeStyle = 'rgba(188, 211, 192, .3)';
    context.lineWidth = Math.max(2, size * 0.014);
    context.beginPath();
    context.moveTo(size * 0.08, size * 0.92);
    context.lineTo(size * 0.92, size * 0.08);
    context.moveTo(size * 0.08, size * 0.92);
    context.lineTo(size * 0.08, size * 0.08);
    context.lineTo(size * 0.92, size * 0.08);
    context.stroke();

    for (const unit of state.units.values()) {
        const color = unit.team === 'red' ? '#dc4b4b' : unit.team === 'blue' ? '#3e82dc' : '#c6b65d';
        point(unit, unit.kind === 'tower' || unit.kind === 'crystal' ? size * 0.015 : size * 0.008, color);
    }
    for (const player of state.players.values()) {
        const self = player.playerId === state.playerId;
        point(
            player,
            self ? size * 0.022 : size * 0.016,
            self ? '#ffffff' : player.team === 'red' ? '#ff5b5b' : '#55a0ff',
        );
    }
}

function calculateFrame() {
    const scale = state.cameraScale;
    const self = state.sprites.get(state.playerId) || state.players.get(state.playerId);
    const focus = self || { x: state.map.width / 2, y: state.map.height / 2 };
    const hudTop = els.hud?.getBoundingClientRect().top || app.renderer.height;
    const viewHeight = hudTop > app.renderer.height * 0.45 ? hudTop : app.renderer.height;
    const minOffsetX = Math.min(0, app.renderer.width - state.map.width * scale);
    const minOffsetY = Math.min(0, viewHeight - state.map.height * scale);
    const offsetX = app.renderer.width / 2 - focus.x * scale;
    const offsetY = viewHeight / 2 - focus.y * scale;
    state.frame = {
        scale,
        offsetX: Math.max(minOffsetX, Math.min(0, offsetX)),
        offsetY: Math.max(minOffsetY, Math.min(0, offsetY)),
    };
    return state.frame;
}

function drawMap(frame) {
    gridLayer.clear();
    gridLayer.rect(frame.offsetX, frame.offsetY, state.map.width * frame.scale, state.map.height * frame.scale);
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

    const selectedTarget = state.selectedTargetId ? targetMap().get(state.selectedTargetId) : null;
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
