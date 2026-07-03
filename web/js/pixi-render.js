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

