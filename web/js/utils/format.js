function websocketURL() {
  const scheme = location.protocol === "https:" ? "wss" : "ws";
  return `${scheme}://${location.host || "localhost:6969"}/ws`;
}


function screenPointFromEvent(event) {
  const rect = app.canvas.getBoundingClientRect();
  return {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top,
  };
}

function screenToWorld(event) {
  const canvasPoint = screenPointFromEvent(event);
  return {
    x: (canvasPoint.x - state.frame.offsetX) / state.frame.scale,
    y: (canvasPoint.y - state.frame.offsetY) / state.frame.scale,
  };
}

function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function formatNumber(value) {
  return Number.isInteger(value)
    ? String(value)
    : String(Math.round(value * 1000) / 1000);
}

function formatInteger(value) {
  return String(Math.floor(value || 0));
}

function shieldValue(entity) {
  return Math.max(0, entity?.passive?.shield || 0);
}

function formatHpWithShield(entity) {
  const stats = entity?.stats || {};
  const shield = shieldValue(entity);
  if (shield <= 0) {
    return `${formatInteger(stats.hp)}/${formatInteger(stats.maxHp)}`;
  }
  return `${formatInteger(stats.hp)} + ${formatInteger(shield)}/${formatInteger(stats.maxHp)}`;
}

function formatHpRegen5(entity) {
  const stats = entity?.stats || {};
  const base = (stats.hpRegen5 || 0) + equipmentPercentRegen5(entity, "hp");
  const passive = warriorToughnessRegen5(entity);
  if (passive <= 0) {
    return formatNumber(base);
  }
  return `${formatNumber(base)} + ${formatNumber(passive)}`;
}

function equipmentPercentRegen5(entity, resource) {
  if (!entity?.equipment || !entity?.stats) {
    return 0;
  }
  const outOfCombat =
    (state.snapshotTick || 0) >=
    (entity.lastHitTick || 0) + 5 * (state.tickRate || 20);
  return entity.equipment.reduce((total, equipment) => {
    const effects = equipmentConfig(equipment)?.effects || {};
    const ratio = outOfCombat
      ? effects[`outOfCombat${resource === "hp" ? "Hp" : "Mp"}RegenMax${resource === "hp" ? "Hp" : "Mp"}Ratio5`]
      : effects[`combat${resource === "hp" ? "Hp" : "Mp"}RegenMax${resource === "hp" ? "Hp" : "Mp"}Ratio5`];
    return total + (entity.stats[resource === "hp" ? "maxHp" : "maxMp"] || 0) * (ratio || 0);
  }, 0);
}

function warriorToughnessRegen5(entity) {
  if ((entity?.heroId || "") !== "warrior") {
    return 0;
  }
  const ratios =
    skillClientConfig.warrior_toughness?.metaLists?.regenMaxHPRatio || [];
  if (ratios.length === 0) {
    return 0;
  }
  const level = clamp(Math.max(1, entity.level || 1), 1, ratios.length);
  const ratio = ratios[level - 1] || 0;
  return (entity.stats?.maxHp || 0) * ratio;
}

function hpShieldRatio(entity) {
  const stats = entity?.stats || {};
  return ratio((stats.hp || 0) + shieldValue(entity), stats.maxHp || 0);
}

function formatAttack(stats) {
  return formatBasePlusBonus(stats.attack || 0, stats.bonusAttack || 0);
}

function formatPhysicalDefense(stats) {
  return formatBasePlusBonus(
    stats.physicalDefense || 0,
    stats.bonusPhysicalDefense || 0,
  );
}

function formatMagicDefense(stats) {
  return formatBasePlusBonus(
    stats.magicDefense || 0,
    stats.bonusMagicDefense || 0,
  );
}

function formatDefenseTip(resistance, typeLabel) {
  return `<span class="stat-tip" data-tip="${escapeHtml(formatResistanceTip(resistance, typeLabel))}">?</span>`;
}

function formatResistanceTip(resistance, typeLabel) {
  if (resistance >= 0) {
    const reduce = resistance / (resistance + 100);
    return `${typeLabel}伤害减免 ${formatPercent(reduce)}`;
  }
  const multiplier = 100 / Math.max(1, 100 + resistance);
  return `${typeLabel}伤害放大 ${formatNumber(multiplier)} 倍`;
}

function formatPercent(value) {
  return `${formatNumber(value * 100)}%`;
}

function formatBasePlusBonus(base, bonus) {
  if (bonus <= 0) {
    return formatNumber(base);
  }
  return `${formatNumber(base)} + ${formatNumber(bonus)}`;
}

function formatSwordIntent(passive) {
  return passive?.maxSwordIntent > 0
    ? `${Math.floor(passive.swordIntent || 0)}/${Math.floor(passive.maxSwordIntent)}`
    : "-";
}

function entityResourceKind(entity, heroConfig) {
  if (heroConfig?.resource) {
    return heroConfig.resource;
  }
  if (entity?.heroId === "sword") {
    return "sword_intent";
  }
  if (entity?.heroId === "blade") {
    return "rage";
  }
  if (entity?.heroId === "ninja") {
    return "energy";
  }
  if (entity?.stats?.maxMp > 0) {
    return "mp";
  }
  return "none";
}

function formatEntityResourceValue(entity, heroConfig) {
  const kind = entityResourceKind(entity, heroConfig);
  const stats = entity?.stats || {};
  if (kind === "none") {
    return "-";
  }
  if (kind === "sword_intent") {
    return formatSwordIntent(entity?.passive || {});
  }
  return `${formatInteger(stats.mp)}/${formatInteger(stats.maxMp)}`;
}

function formatTargetResource(target) {
  const heroConfig = heroClientConfig[target?.heroId] || {};
  const kind = entityResourceKind(target, heroConfig);
  if (kind === "none") {
    return "";
  }
  return `<div>${formatResource(kind)} ${formatEntityResourceValue(target, heroConfig)}</div>`;
}

function formatTargetMpRegen(target) {
  const stats = target?.stats || {};
  if (!stats?.maxMp || stats.maxMp <= 0) {
    return "";
  }
  return `<div>法力/5秒 ${formatNumber((stats.mpRegen5 || 0) + equipmentPercentRegen5(target, "mp"))}</div>`;
}

function formatResource(resource) {
  if (resource === "sword_intent") {
    return "剑意";
  }
  if (resource === "rage") {
    return "怒气";
  }
  if (resource === "energy") {
    return "能量";
  }
  if (!resource || resource === "mp") {
    return "法力";
  }
  if (resource === "none") {
    return "";
  }
  return resource;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function ratio(value, max) {
  if (!max || max <= 0) {
    return 0;
  }
  return clamp(value / max, 0, 1);
}
