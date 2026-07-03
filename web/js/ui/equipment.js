function setEquipmentCard(player) {
  els.equipGold.textContent = player ? Math.floor(player.gold || 0) : "-";
  const equipments = Array.isArray(player?.equipment) ? player.equipment : [];
  els.equipmentSlots.forEach((slot, index) => {
    const equipment = equipments[index];
    const name = equipmentName(equipment);
    slot.textContent = name;
    slot.disabled = !player || name === "-";
    slot.classList.toggle("selected", state.selectedEquipmentSlot === index + 1);
    setEquipmentTip(els.equipmentTips[index], equipment);
  });
}

function setEquipmentTip(tip, equipment) {
  if (!tip) {
    return;
  }
  const text = formatEquipmentTip(equipment);
  if (!text) {
    tip.innerHTML = "";
    return;
  }
  tip.innerHTML = `<span class="stat-tip" data-tip="${escapeHtml(text)}">?</span>`;
}

function formatEquipmentTip(equipment) {
  const config = equipmentConfig(equipment);
  if (!config) {
    return "";
  }
  if (Array.isArray(config.description) && config.description.length) {
    return config.description.join("\n");
  }
  const parts = [];
  const stats = config.stats || {};
  addTipStat(parts, stats.attack, "攻击力");
  addTipStat(parts, stats.abilityPower, "法术强度");
  addTipStat(parts, stats.abilityHaste, "技能急速");
  addTipStat(parts, stats.hp, "生命");
  addTipStat(parts, stats.mp, "法力");
  addTipStat(parts, stats.physicalDefense, "物理防御");
  addTipStat(parts, stats.magicDefense, "魔法防御");
  addTipStat(parts, stats.moveSpeed, "移动速度");
  addTipStat(parts, stats.hpRegen5, "生命/5秒");
  addTipStat(parts, stats.mpRegen5, "法力/5秒");
  addTipPercent(parts, stats.attackSpeedBonus, "攻击速度");
  addTipPercent(parts, stats.critChance, "暴击率");
  addTipPercent(parts, stats.moveSpeedPercent, "移动速度");
  addTipPercent(parts, stats.omnivamp, "全能吸血");
  addTipPercent(parts, stats.lifeSteal, "生命偷取");
  addTipPercent(parts, stats.healingPower, "治疗加成");
  addTipPercent(parts, stats.grievousWounds, "重伤");
  const effects = config.effects || {};
  if (effects.basicAttackBonusDamage) {
    parts.push(
      `普攻命中 +${formatNumber(effects.basicAttackBonusDamage)} ${formatDamageTypeName(effects.basicAttackBonusDamageType)}`,
    );
  }
  if (effects.minionBasicAttackBonusDamage) {
    parts.push(
      `普攻小兵 +${formatNumber(effects.minionBasicAttackBonusDamage)} ${formatDamageTypeName(effects.minionBasicAttackBonusDamageType)}`,
    );
  }
  if (effects.heroHitSmallHeal) {
    parts.push(`被英雄命中回血 +${formatNumber(effects.heroHitHeal || 0)}`);
  }
  if (effects.levelUpRestoreHpRatio || effects.levelUpRestoreMpRatio) {
    parts.push(
      `升级回复生命 ${formatPercent(effects.levelUpRestoreHpRatio || 0)} / 法力 ${formatPercent(effects.levelUpRestoreMpRatio || 0)}`,
    );
  }
  if (effects.outOfCombatMoveSpeed) {
    parts.push(`脱战移动速度 +${formatNumber(effects.outOfCombatMoveSpeed)}`);
  }
  if (effects.unitKillPhysicalDefenseGain || effects.unitKillAbilityPowerGain) {
    parts.push(
      `击杀单位 +${formatNumber(effects.unitKillPhysicalDefenseGain || 0)} 物理防御 / +${formatNumber(effects.unitKillAbilityPowerGain || 0)} 法术强度，最多 +${formatNumber(effects.unitKillMaxGain || 0)}`,
    );
  }
  if (effects.critDamageBonus) {
    parts.push(`暴击伤害 +${formatPercent(effects.critDamageBonus)}`);
  }
  if (effects.lowHealthShieldMax) {
    parts.push(
      `低生命护盾 ${formatNumber(effects.lowHealthShieldMin || 0)}-${formatNumber(effects.lowHealthShieldMax)} / 减伤 ${formatPercent(effects.lowHealthDamageReduce || 0)}`,
    );
  }
  return parts.join("\n");
}

function formatDamageTypeName(type) {
  if (type === "physical") {
    return "物理伤害";
  }
  if (type === "magic") {
    return "魔法伤害";
  }
  if (type === "true") {
    return "真实伤害";
  }
  return "伤害";
}

function equipmentConfig(equipment) {
  if (!equipment) {
    return null;
  }
  if (typeof equipment === "string") {
    return equipmentClientConfig[equipment] || null;
  }
  const equipmentId = equipment.equipmentId || equipment.id || "";
  if (equipmentId && equipmentClientConfig[equipmentId]) {
    return equipmentClientConfig[equipmentId];
  }
  return equipment.stats || equipment.effects ? equipment : null;
}

function addTipStat(parts, value, label) {
  if (!value || value <= 0) {
    return;
  }
  parts.push(`${label} +${formatNumber(value)}`);
}

function addTipPercent(parts, value, label) {
  if (!value || value <= 0) {
    return;
  }
  parts.push(`${label} +${formatPercent(value)}`);
}

function equipmentName(equipment) {
  if (!equipment) {
    return "-";
  }
  if (typeof equipment === "string") {
    return equipment || "-";
  }
  if (typeof equipment.name === "string" && equipment.name) {
    return equipment.name;
  }
  if (typeof equipment.equipmentId === "string" && equipment.equipmentId) {
    return equipment.equipmentId;
  }
  return "-";
}

function setStatPairVisible(label, value, visible) {
  label.style.display = visible ? "" : "none";
  value.style.display = visible ? "" : "none";
}
