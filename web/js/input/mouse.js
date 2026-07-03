function handlePointerDown(event) {
  updateAimPoint(event);
  if (event.button === 0) {
    const targetId = pickTargetUnit(event);
    if (targetId) {
      state.selectedTargetId = targetId;
      setTargetCard(currentTarget());
      if (state.attackMoveArmed) {
        attackTarget(targetId);
        state.attackMoveArmed = false;
      }
      return;
    }
    state.selectedTargetId = "";
    state.attackMoveArmed = false;
    setTargetCard(null);
    return;
  }
  if (event.button !== 2) {
    return;
  }
  const point = screenToWorld(event);
  const targetId = pickTargetUnit(event);
  if (targetId) {
    state.selectedTargetId = targetId;
    setTargetCard(currentTarget());
    attackTarget(targetId);
    return;
  }
  moveToPoint(point);
}

function updateAimPoint(event) {
  const point = screenToWorld(event);
  state.aimPoint = {
    x: clamp(point.x, 0, state.map.width),
    y: clamp(point.y, 0, state.map.height),
  };
}

function attackTarget(targetId) {
  const self = state.players.get(state.playerId);
  if (self?.dead) {
    return;
  }
  state.attackTargetId = targetId;
  state.moveTarget = null;
  sendPacket("input", {
    attack: {
      targetId,
    },
    clientSeq: state.seq,
  });
}

function moveToPoint(point) {
  const self = state.players.get(state.playerId);
  if (self?.dead) {
    return;
  }
  state.attackMoveArmed = false;
  state.attackTargetId = "";
  state.moveTarget = {
    x: clamp(point.x, 0, state.map.width),
    y: clamp(point.y, 0, state.map.height),
  };
  sendPacket("input", {
    move: {
      targetX: state.moveTarget.x,
      targetY: state.moveTarget.y,
    },
    clientSeq: state.seq,
  });
  clearAttackTarget();
}

function clearAttackTarget() {
  sendPacket("input", {
    attack: {
      clear: true,
    },
    clientSeq: state.seq,
  });
}

