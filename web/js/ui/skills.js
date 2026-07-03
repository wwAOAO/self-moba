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
