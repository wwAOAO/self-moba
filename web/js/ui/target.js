function setTargetCard(target) {
  const targetCards =
    els.target.closest(".target-cards") || els.target.parentElement;
  if (!target?.stats) {
    targetCards.style.display = "none";
    setHtmlIfChanged(els.target, "-");
    setHtmlIfChanged(els.targetEquipment, "-");
    return;
  }
  const stats = target.stats;
  const airborneTicks = Math.max(
    0,
    (target.control?.airborneUntilTick || 0) -
      Number(els.tick.textContent || 0),
  );
  const targetID = target.id || target.playerId || "";
  const idRow =
    targetID && !targetID.startsWith("spawn:") ? `<div>${targetID}</div>` : "";
  const showEquipment = target.kind === "player" || target.kind === "enemy_hero";
  targetCards.style.display = "grid";
  els.targetEquipment.parentElement.style.display = showEquipment ? "" : "none";
  setHtmlIfChanged(els.target, `
    <div>${targetLabel(target)}</div>
    ${idRow}
    ${airborneTicks > 0 ? `<div>击飞 ${(airborneTicks / state.tickRate).toFixed(1)}s</div>` : ""}
    <div>生命 ${formatHpWithShield(target)}</div>
    ${formatTargetResource(target)}
    <div class="stats-grid">
      <span>攻击力 ${formatAttack(stats)}</span>
      <span>法术强度 ${stats.abilityPower || 0}</span>
      <span>攻击速度 ${stats.attackSpeed}</span>
      <span>技能急速 ${formatNumber(stats.abilityHaste || 0)}</span>
      <span>物理防御 ${formatDefenseTip(stats.physicalDefense || 0, "物理")} ${formatPhysicalDefense(stats)}</span>
      <span>魔法防御 ${formatDefenseTip(stats.magicDefense || 0, "魔法")} ${formatMagicDefense(stats)}</span>
      <span>移动速度 ${stats.moveSpeed}</span>
      <span>攻击距离 ${stats.attackRange}</span>
      <span>暴击率 ${formatCritChance(target)}${formatCritChanceTip(target)}</span>
    </div>
  `);
  setHtmlIfChanged(
    els.targetEquipment,
    showEquipment ? formatTargetEquipment(target) : "-",
  );
}

function formatTargetEquipment(target) {
  const rows = (Array.isArray(target?.equipment) ? target.equipment : [])
    .map(formatTargetEquipmentRow)
    .filter(Boolean);
  return rows.length ? `<div class="stats-grid">${rows.join("")}</div>` : "-";
}

function formatTargetEquipmentRow(equipment, index) {
  const name = equipmentName(equipment);
  if (name === "-") {
    return "";
  }
  const tip = formatEquipmentTip(equipment);
  const tipIcon = tip
    ? `<span class="stat-tip" data-tip="${escapeHtml(tip)}">?</span>`
    : "<span></span>";
  return `<span>${index + 1}</span><span class="equipment-slot-row">${tipIcon}<span>${escapeHtml(name)}</span></span>`;
}
