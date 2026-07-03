function updatePositionLabel() {
  const self = state.players.get(state.playerId);
  if (!self) {
    els.position.textContent = "-";
    els.buffs.innerHTML = "-";
    setStatsCard(null);
    return;
  }
  els.position.textContent = `${self.x.toFixed(1)}, ${self.y.toFixed(1)}`;
  els.teamLabel.textContent = self.team || state.team;
  setStatsCard(self);
  setTargetCard(currentTarget());
  const tick = Number(els.tick.textContent || 0);
  els.buffs.innerHTML = formatPlayerBuffs(self, tick) || "-";
  if (!self.skills || self.skills.length === 0) {
    els.skills.innerHTML = "-";
    return;
  }
  els.skills.innerHTML = formatSkillCooldowns(self, tick);
}

function formatSkillCooldowns(player, tick) {
  const slots =
    heroSkillSlots[player.heroId || els.heroId.value] ||
    heroSkillSlots[els.heroId.value];
  if (!slots) {
    return "-";
  }
  const canSpend = (player.skillPoints || 0) > 0;
  return `<div class="skill-list">${["q", "w", "e", "r"]
    .map((slot) => {
      const skill = skillState(player, slots[slot]);
      const remainTicks = Math.max(0, (skill?.cooldownUntilTick || 0) - tick);
      const remainSeconds = (remainTicks / state.tickRate).toFixed(1);
      const level = skill?.level || 0;
      const maxLevel = maxSkillLevel(slot);
      const disabled = !canSpend || level >= maxLevel ? "disabled" : "";
      const chargeText = formatSkillChargeText(player, slot, skill);
      const timeText = formatSkillTimeText(player, slot, skill, remainSeconds, tick);
      return `<div class="skill-row${chargeText ? " has-charge" : ""}">
                      <strong>${slot.toUpperCase()}</strong>
                      <span>${level}/${maxLevel}</span>
                      ${chargeText ? `<span>${chargeText}</span>` : ""}
                      <span>${timeText}</span>
                      <button type="button" class="icon-button" data-skill-upgrade="${slot}" ${disabled}>+</button>
                  </div>`;
    })
    .join("")}</div>`;
}

function formatSkillChargeText(player, slot, skill) {
  const heroId = player.heroId || els.heroId.value;
  if (heroId === "archer" && slot === "e") {
    return `${skill?.stacks || 0}/2`;
  }
  return "";
}

function formatSkillTimeText(player, slot, skill, remainSeconds, tick) {
  const heroId = player.heroId || els.heroId.value;
  if (heroId === "archer" && slot === "e") {
    const rechargeTicks = Math.max(0, (skill?.stacksExpireTick || 0) - tick);
    return `${(rechargeTicks / state.tickRate).toFixed(1)}s`;
  }
  return `${remainSeconds}s`;
}

function maxSkillLevel(slot) {
  return slot === "r" ? 3 : 5;
}

function formatHeroSkillState(player, tick) {
  const heroId = player.heroId || els.heroId.value;
  if (heroId === "sword") {
    return formatSwordSkillState(player, tick);
  }
  if (heroId === "archer") {
    return formatArcherSkillState(player, tick);
  }
  return "";
}

function formatPlayerBuffs(player, tick) {
  const rows = [];
  for (const buff of player.buffs || []) {
    rows.push(formatBuffRow(buff, tick));
  }
  rows.push(...formatControlBuffRows(player, tick));
  const heroBuffs = formatHeroSkillState(player, tick);
  if (heroBuffs) {
    rows.push(heroBuffs);
  }
  return rows.length ? `<div class="skill-list">${rows.join("")}</div>` : "";
}

function formatControlBuffRows(player, tick) {
  const control = player.control || {};
  return [
    controlBuffRow("击飞", control.airborneUntilTick, tick),
    controlBuffRow("眩晕", control.stunnedUntilTick, tick),
    controlBuffRow("沉默", control.silencedUntilTick, tick),
    controlBuffRow("禁锢", control.rootedUntilTick, tick),
    controlBuffRow("韧性", control.tenacityUntilTick, tick),
    controlBuffRow("减速", control.moveSpeedSlowUntil, tick),
    controlBuffRow("启明", control.mageIlluminationUntil, tick),
  ].filter(Boolean);
}

function controlBuffRow(name, untilTick, tick) {
  if (!untilTick || untilTick <= tick) {
    return "";
  }
  return `<div class="buff-row"><strong>${name}</strong><span>${((untilTick - tick) / state.tickRate).toFixed(1)}s</span></div>`;
}

function formatBuffRow(buff, tick) {
  const name = escapeHtml(formatBuffName(buff));
  const remain =
    buff.expiresAtTick > 0
      ? `${Math.max(0, (buff.expiresAtTick - tick) / state.tickRate).toFixed(1)}s`
      : "∞";
  return `<div class="buff-row"><strong>${name}</strong><span>${remain}</span></div>`;
}

function formatBuffName(buff) {
  if (buff.id === "debug_ability_haste") {
    return `+${formatNumber(buff.abilityHaste || 0)}技能急速`;
  }
  return buff.name || buff.id || "buff";
}

function formatArcherSkillState(player, tick) {
  const archer = player.archer || {};
  if ((archer.focusActiveUntil || 0) > tick) {
    const remain = ((archer.focusActiveUntil - tick) / state.tickRate).toFixed(
      1,
    );
    return `<div class="skill-list"><div class="skill-row"><strong>foc</strong><span>生效中</span><span>${remain}s</span><span></span></div></div>`;
  }
  const stacks = archer.focusStacks || 0;
  if (stacks <= 0) {
    return "";
  }
  const remain = Math.max(0, (archer.focusExpireTick || 0) - tick);
  return `<div class="skill-list"><div class="skill-row"><strong>foc</strong><span>${stacks}/4</span><span>${(remain / state.tickRate).toFixed(1)}s</span><span></span></div></div>`;
}

function formatSwordSkillState(player, tick) {
  const qState = skillState(player, "sword_cut");
  if (!qState || (qState.level || 0) <= 0 || (qState.stacks || 0) <= 0) {
    return "";
  }
  const remain = Math.max(0, (qState.stacksExpireTick || 0) - tick);
  if (remain <= 0) {
    return "";
  }
  return `<div class="skill-list"><div class="skill-row"><strong>buf</strong><span>${qState.stacks}/2</span><span>${(remain / state.tickRate).toFixed(1)}s</span><span></span></div></div>`;
}

function skillIdForSlot(heroId, slot) {
  return (
    heroSkillSlots[heroId]?.[slot] ||
    heroClientConfig[heroId]?.skills?.[slot] ||
    ""
  );
}

function skillState(player, skillId) {
  return (
    (player.skills || []).find((skill) => skill.skillId === skillId) || null
  );
}

function skillMetaListValue(skillId, key, level, fallback) {
  const values = skillClientConfig[skillId]?.metaLists?.[key] || fallback;
  const index = clamp(Math.max(1, level || 1), 1, values.length) - 1;
  return values[index] || 0;
}

function swordETargetCooldownTicks(player) {
  const level = skillState(player, "sword_sweeping_blade")?.level || 1;
  const cooldownMs = skillMetaListValue(
    "sword_sweeping_blade",
    "targetCooldownMs",
    level,
    [10000, 9000, 8000, 7000, 6000],
  );
  return (cooldownMs / 1000) * state.tickRate;
}

function isSkillOnCooldown(player, skillId) {
  const tick = Number(els.tick.textContent || 0);
  const skill = skillState(player, skillId);
  return (skill?.cooldownUntilTick || 0) > tick;
}

function isSkillLearned(player, skillId) {
  return (skillState(player, skillId)?.level || 0) > 0;
}

function setStatus(value) {
  els.status.textContent = value;
}

function setShopStatus(value) {
  els.shopStatus.textContent = value || "-";
}

function setStatsCard(player) {
  if (!player?.stats) {
    els.statLevel.textContent = "-";
    els.statExp.textContent = "-";
    els.statSkillPoints.textContent = "-";
    setEquipmentCard(null);
    setStatPairVisible(els.statResourceLabel, els.statResource, false);
    els.statResource.textContent = "-";
    els.statMpLabel.textContent = "法力";
    els.statHp.textContent = "-";
    els.statMp.textContent = "-";
    els.statHpRegen5.textContent = "-";
    setStatPairVisible(els.statMpRegen5Label, els.statMpRegen5, false);
    els.statMpRegen5.textContent = "-";
    els.statAttack.textContent = "-";
    els.statAbilityPower.textContent = "-";
    els.statAbilityHasteTip.innerHTML = "";
    els.statAbilityHaste.textContent = "-";
    els.statPhysicalDefense.textContent = "-";
    els.statMagicDefense.textContent = "-";
    els.statMoveSpeed.textContent = "-";
    els.statAttackRange.textContent = "-";
    els.statAttackSpeed.textContent = "-";
    els.statCritChance.textContent = "-";
    els.statOmnivamp.textContent = "-";
    els.statLifeSteal.textContent = "-";
    els.statHealingPower.textContent = "-";
    els.abilityHasteBtn.textContent = "+200急速";
    return;
  }
  const stats = player.stats;
  const passive = player.passive || {};
  const heroConfig = heroClientConfig[player.heroId || els.heroId.value] || {};
  const isSword = (player.heroId || els.heroId.value) === "sword";
  els.statLevel.textContent = `${player.level || 1}/${player.maxLevel || levelClientConfig.maxLevel || 18}`;
  els.statExp.textContent =
    player.nextLevelExp > 0
      ? `${Math.floor(player.exp || 0)}/${Math.floor(player.nextLevelExp)}`
      : "满级";
  els.statSkillPoints.textContent = player.skillPoints || 0;
  setEquipmentCard(player);
  const resourceLabel = formatResource(heroConfig.resource);
  const hasResource = resourceLabel !== "";
  setStatPairVisible(els.statResourceLabel, els.statResource, hasResource);
  els.statResource.textContent = hasResource ? resourceLabel : "-";
  els.statHp.textContent = formatHpWithShield(player);
  els.statMpLabel.textContent = isSword ? "剑意" : "法力";
  els.statMp.textContent = isSword
    ? formatSwordIntent(passive)
    : stats.maxMp > 0
      ? `${formatInteger(stats.mp)}/${formatInteger(stats.maxMp)}`
      : "-";
  els.statHpRegen5.textContent = formatHpRegen5(player);
  const showMpRegen = !isSword && stats.maxMp > 0;
  setStatPairVisible(els.statMpRegen5Label, els.statMpRegen5, showMpRegen);
  els.statMpRegen5.textContent = showMpRegen
    ? formatNumber((stats.mpRegen5 || 0) + equipmentPercentRegen5(player, "mp"))
    : "-";
  els.statAttack.textContent = formatAttack(stats);
  els.statAbilityPower.textContent = stats.abilityPower || 0;
  els.statAbilityHasteTip.innerHTML = formatAbilityHasteTip(
    stats.abilityHaste || 0,
  );
  els.statAbilityHaste.textContent = formatNumber(stats.abilityHaste || 0);
  els.statPhysicalDefense.textContent = formatPhysicalDefense(stats);
  els.statPhysicalDefenseTip.innerHTML = formatDefenseTip(
    stats.physicalDefense || 0,
    "物理",
  );
  els.statMagicDefense.textContent = formatMagicDefense(stats);
  els.statMagicDefenseTip.innerHTML = formatDefenseTip(
    stats.magicDefense || 0,
    "魔法",
  );
  els.statMoveSpeed.textContent = formatNumber(stats.moveSpeed);
  els.statAttackRange.textContent = formatNumber(stats.attackRange);
  els.statAttackSpeed.textContent = formatNumber(stats.attackSpeed);
  els.statCritChance.textContent = `${Math.round((stats.critChance || 0) * 1000) / 10}%`;
  els.statOmnivamp.textContent = formatPercent(stats.omnivamp || 0);
  els.statLifeSteal.textContent = formatPercent(stats.lifeSteal || 0);
  els.statHealingPower.textContent = formatPercent(stats.healingPower || 0);
  els.abilityHasteBtn.textContent =
    (player.buffs || []).some((buff) => buff.id === "debug_ability_haste")
      ? "关闭200急速"
      : "+200急速";
}

function formatPercent(value) {
  return `${Math.round(value * 1000) / 10}%`;
}

function formatAbilityHasteTip(abilityHaste) {
  const reduction = abilityHaste / (100 + abilityHaste);
  return `<span class="stat-tip" data-tip="实际减少 ${formatPercent(reduction)} 冷却">?</span>`;
}

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

function setTargetCard(target) {
  if (!target?.stats) {
    els.target.parentElement.style.display = "none";
    els.target.innerHTML = "-";
    return;
  }
  const stats = target.stats;
  const airborneTicks = Math.max(
    0,
    (target.control?.airborneUntilTick || 0) -
      Number(els.tick.textContent || 0),
  );
  els.target.parentElement.style.display = "block";
  els.target.innerHTML = `
    <div>${targetLabel(target)}</div>
    <div>${target.id || target.playerId}</div>
    <div>阵营 ${target.team || "-"}</div>
    ${airborneTicks > 0 ? `<div>击飞 ${(airborneTicks / state.tickRate).toFixed(1)}s</div>` : ""}
    <div>生命 ${formatHpWithShield(target)}</div>
    ${formatTargetResource(target)}
    <div>生命/5秒 ${formatHpRegen5(target)}</div>
    ${formatTargetMpRegen(target)}
    <div>攻击力 ${formatAttack(stats)}</div>
    <div>法术强度 ${stats.abilityPower || 0}</div>
    <div>技能急速 ${formatNumber(stats.abilityHaste || 0)}</div>
    <div>物理防御 ${formatDefenseTip(stats.physicalDefense || 0, "物理")} ${formatPhysicalDefense(stats)}</div>
    <div>魔法防御 ${formatDefenseTip(stats.magicDefense || 0, "魔法")} ${formatMagicDefense(stats)}</div>
    <div>移动速度 ${stats.moveSpeed}</div>
    <div>攻击距离 ${stats.attackRange}</div>
    <div>攻击速度 ${stats.attackSpeed}</div>
    <div>暴击率 ${Math.round((stats.critChance || 0) * 1000) / 10}%</div>
  `;
}
