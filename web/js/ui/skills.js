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
  if (!els.skills.querySelector(".stat-tip:hover")) {
    els.skills.innerHTML = formatSkillCooldowns(self, tick);
  }
}

function formatSkillCooldowns(player, tick) {
  const slots =
    heroSkillSlots[player.heroId || els.heroId.value] ||
    heroSkillSlots[els.heroId.value];
  if (!slots) {
    return "-";
  }
  const canSpend = (player.skillPoints || 0) > 0;
  const rows = [];
  const passiveRow = formatPassiveSkillRow(player, slots.passive, tick);
  if (passiveRow) {
    rows.push(passiveRow);
  }
  rows.push(
    ...["q", "w", "e", "r"].map((slot) => {
      const skill = skillState(player, slots[slot]);
      const remainTicks = Math.max(0, (skill?.cooldownUntilTick || 0) - tick);
      const remainSeconds = (remainTicks / state.tickRate).toFixed(1);
      const level = skill?.level || 0;
      const maxLevel = maxSkillLevel(slot);
      const disabled = !canSpend || !canUpgradeSkill(player, slot, level) ? "disabled" : "";
      const chargeText = formatSkillChargeText(player, slot, skill);
      const timeText = formatSkillTimeText(player, slot, skill, remainSeconds, tick);
      const tip = formatSkillTip(slots[slot], slot, level || 1);
      return `<div class="skill-row${chargeText ? " has-charge" : ""}">
                      <strong>${slot.toUpperCase()}</strong>
                      <span>${level}/${maxLevel}</span>
                      ${chargeText ? `<span>${chargeText}</span>` : ""}
                      <span>${timeText}</span>
                      ${tip}
                      <button type="button" class="icon-button" data-skill-upgrade="${slot}" ${disabled}>+</button>
                  </div>`;
    }),
  );
  return `<div class="skill-list">${rows.join("")}</div>`;
}

function formatPassiveSkillRow(player, skillId, tick) {
  if (!skillId) {
    return "";
  }
  const skill = skillState(player, skillId);
  const remainTicks = Math.max(0, (skill?.cooldownUntilTick || 0) - tick);
  const timeText =
    remainTicks > 0 ? `${(remainTicks / state.tickRate).toFixed(1)}s` : "-";
  const tip = formatSkillTip(skillId, "p", player.level || 1);
  return `<div class="skill-row">
                      <strong>P</strong>
                      <span>-</span>
                      <span>${timeText}</span>
                      ${tip}
                      <span></span>
                  </div>`;
}

function formatSkillTip(skillId, slot, level) {
  const config = skillClientConfig[skillId];
  if (!config) {
    return "";
  }
  const text = skillTipText(config, slot, level);
  return text ? `<span class="stat-tip" data-tip="${escapeHtml(text)}">?</span>` : "";
}

function skillTipText(config, slot, level) {
  if (Array.isArray(config.description) && config.description.length) {
    return config.description.join("\n");
  }
  const parts = [`${slot.toUpperCase()} ${config.name || config.skillId}`];
  if (config.type) {
    parts.push(`类型 ${formatSkillType(config.type)}`);
  }
  if (config.range > 0) {
    parts.push(`范围 ${formatNumber(config.range)}`);
  }
  const cooldownMs = skillTipListValue(config, "cooldownMs", level, config.cooldownMs || 0);
  if (cooldownMs > 0) {
    parts.push(`冷却 ${(cooldownMs / 1000).toFixed(1)}秒`);
  }
  const manaCost = skillTipListValue(config, "manaCost", level, 0);
  if (manaCost > 0) {
    parts.push(`耗蓝 ${formatNumber(manaCost)}`);
  }
  const usedListKeys = new Set(["cooldownMs", "manaCost", "baseDamage", "damagePerSecond", "shieldValue", "waves"]);
  const usedMetaKeys = new Set(["durationSeconds", "radius", "coneAngleDegrees"]);
  addSkillTipList(parts, config, "baseDamage", level, "基础伤害");
  addSkillTipList(parts, config, "damagePerSecond", level, "每秒伤害");
  addSkillTipList(parts, config, "shieldValue", level, "护盾");
  addSkillTipList(parts, config, "waves", level, "弹幕波数");
  if (config.durationSeconds) {
    parts.push(`持续 ${formatNumber(config.durationSeconds)}秒`);
  }
  if (config.radius) {
    parts.push(`半径 ${formatNumber(config.radius)}`);
  }
  if (config.coneAngleDegrees) {
    parts.push(`角度 ${formatNumber(config.coneAngleDegrees)}度`);
  }
  addSkillTipMetaDetails(parts, config, usedMetaKeys);
  addSkillTipListDetails(parts, config, level, usedListKeys);
  return parts.join("\n");
}

function addSkillTipList(parts, config, key, level, label) {
  const value = skillTipListValue(config, key, level, 0);
  if (value > 0) {
    parts.push(skillTipListLine(label, value, config.metaLists?.[key]));
  }
}

function addSkillTipMetaDetails(parts, config, usedKeys) {
  const meta = config.meta || {};
  for (const [key, value] of Object.entries(meta)) {
    if (usedKeys.has(key) || value === undefined || value === null || value === "" || value === 0) {
      continue;
    }
    parts.push(`${formatSkillKeyName(key)} ${formatSkillTipValue(key, value)}`);
  }
}

function addSkillTipListDetails(parts, config, level, usedKeys) {
  const metaLists = config.metaLists || {};
  for (const [key, values] of Object.entries(metaLists)) {
    if (usedKeys.has(key) || !Array.isArray(values) || values.length === 0) {
      continue;
    }
    parts.push(skillTipListLine(formatSkillKeyName(key), skillTipListValue(config, key, level, 0), values));
  }
}

function skillTipListLine(label, value, values) {
  const current = formatNumber(value);
  if (!Array.isArray(values) || values.length <= 1) {
    return `${label} ${current}`;
  }
  return `${label} ${current}（全等级 ${values.map(formatNumber).join("/")}）`;
}

function skillTipListValue(config, key, level, fallback) {
  const values = config.metaLists?.[key];
  if (!Array.isArray(values) || values.length === 0) {
    return fallback || 0;
  }
  const index = clamp(Math.max(1, level || 1), 1, values.length) - 1;
  return values[index] || 0;
}

function formatSkillType(type) {
  return (
    {
      passive: "被动",
      damage: "伤害",
      dash: "位移",
      wall: "墙体",
      self_buff: "自身增益",
      defense: "防御",
      execute: "斩杀",
      channel_cone: "引导锥形",
      area_dot: "范围持续伤害",
      target_bounce: "目标弹射",
      spin: "旋转",
      empower: "强化",
      projectile: "投射物",
      self_heal: "自我治疗",
    }[type] || type
  );
}

function formatSkillKeyName(key) {
  return (
    {
      adRatio: "攻击力加成",
      totalAdRatio: "总攻击力加成",
      bonusAdRatio: "额外攻击力加成",
      apRatio: "法强加成",
      slowAPRatio: "法强减速加成",
      cooldownMs: "冷却",
      manaCost: "耗蓝",
      baseDamage: "基础伤害",
      bonusDamage: "额外伤害",
      damagePerSecond: "每秒伤害",
      shieldValue: "护盾",
      width: "宽度",
      waves: "弹幕波数",
      moveSpeed: "移动速度",
      enhancedMoveSpeed: "强化移动速度",
      attackSpeedBonus: "攻击速度加成",
      durationSeconds: "持续时间",
      activeDurationSeconds: "主动持续时间",
      castWindupSeconds: "施法前摇",
      minCastWindupSeconds: "最短前摇",
      dashDurationSeconds: "冲刺时间",
      tickSeconds: "跳伤间隔",
      firstWaveSeconds: "第一波时间",
      lastWaveSeconds: "最后一波时间",
      coneAngleDegrees: "锥形角度",
      radius: "半径",
      eqRadius: "EQ半径",
      range: "范围",
      whirlwindRange: "旋风范围",
      whirlwindRadius: "旋风半径",
      whirlwindSpeed: "旋风速度",
      projectileRadius: "弹体半径",
      projectileSpeed: "弹体速度",
      bounceRange: "弹射范围",
      bounceAngleDegrees: "弹射角度",
      targetPickPadding: "选取容差",
      slow: "减速",
      maxStacks: "最大层数",
      stackDurationSeconds: "层数持续时间",
      knockupSeconds: "击飞时间",
      minCooldownSeconds: "最短冷却",
    }[key] || key
  );
}

function formatSkillTipValue(key, value) {
  if (typeof value !== "number") {
    return String(value);
  }
  if (key.endsWith("Seconds")) {
    return `${formatNumber(value)}秒`;
  }
  if (key.endsWith("Ms")) {
    return `${formatNumber(value / 1000)}秒`;
  }
  if (key.includes("Degrees")) {
    return `${formatNumber(value)}度`;
  }
  if (
    key.includes("Ratio") ||
    key.includes("Bonus") ||
    key.includes("Percent") ||
    key === "slow"
  ) {
    return formatPercent(value);
  }
  return formatNumber(value);
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

function canUpgradeSkill(player, slot, level) {
  if (level >= maxSkillLevel(slot)) {
    return false;
  }
  if (slot !== "r") {
    return true;
  }
  return (player.level || 1) >= [6, 11, 16][level];
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
  const tip = buff.tooltip
    ? `<span class="stat-tip" data-tip="${escapeHtml(buff.tooltip)}">?</span>`
    : "";
  const remain =
    buff.expiresAtTick > 0
      ? `${Math.max(0, (buff.expiresAtTick - tick) / state.tickRate).toFixed(1)}s`
      : "∞";
  const status = buff.stacks > 0 ? `${buff.stacks}层 · ${remain}` : remain;
  return `<div class="buff-row"><strong>${name}${tip}</strong><span>${status}</span></div>`;
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
    return `<div class="skill-list"><div class="skill-row"><strong>foc</strong><span>生效中</span><span>${remain}s</span><span></span><span></span></div></div>`;
  }
  const stacks = archer.focusStacks || 0;
  if (stacks <= 0) {
    return "";
  }
  const remain = Math.max(0, (archer.focusExpireTick || 0) - tick);
  return `<div class="skill-list"><div class="skill-row"><strong>foc</strong><span>${stacks}/4</span><span>${(remain / state.tickRate).toFixed(1)}s</span><span></span><span></span></div></div>`;
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
  return `<div class="skill-list"><div class="skill-row"><strong>buf</strong><span>${qState.stacks}/2</span><span>${(remain / state.tickRate).toFixed(1)}s</span><span></span><span></span></div></div>`;
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
