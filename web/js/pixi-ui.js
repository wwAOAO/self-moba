function updatePositionLabel() {
  const self = state.players.get(state.playerId);
  if (!self) {
    els.position.textContent = "-";
    setStatsCard(null);
    return;
  }
  els.position.textContent = `${self.x.toFixed(1)}, ${self.y.toFixed(1)}`;
  els.teamLabel.textContent = self.team || state.team;
  setStatsCard(self);
  setTargetCard(currentTarget());
  if (!self.skills || self.skills.length === 0) {
    els.skills.innerHTML = "-";
    return;
  }
  const tick = Number(els.tick.textContent || 0);
  els.skills.innerHTML =
    formatSkillCooldowns(self, tick) + formatHeroSkillState(self, tick);
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
      const remainSeconds = (remainTicks / 20).toFixed(1);
      const level = skill?.level || 0;
      const maxLevel = maxSkillLevel(slot);
      const disabled = !canSpend || level >= maxLevel ? "disabled" : "";
      const chargeText = formatSkillRowState(
        player,
        slot,
        skill,
        remainSeconds,
        tick,
      );
      return `<div class="skill-row">
                      <strong>${slot.toUpperCase()}</strong>
                      <span>${level}/${maxLevel}</span>
                      <span>${chargeText}</span>
                      <button type="button" class="icon-button" data-skill-upgrade="${slot}" ${disabled}>+</button>
                  </div>`;
    })
    .join("")}</div>`;
}

function formatSkillRowState(player, slot, skill, remainSeconds, tick) {
  const heroId = player.heroId || els.heroId.value;
  if (heroId === "archer" && slot === "e") {
    return `${skill?.stacks || 0}/2`;
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

function formatArcherSkillState(player, tick) {
  const archer = player.archer || {};
  if ((archer.focusActiveUntil || 0) > tick) {
    const remain = ((archer.focusActiveUntil - tick) / state.tickRate).toFixed(
      1,
    );
    return `<div class="skill-list"><div class="skill-row"><strong>foc</strong><span>Active</span><span>${remain}s</span><span></span></div></div>`;
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

function setStatsCard(player) {
  if (!player?.stats) {
    els.statLevel.textContent = "-";
    els.statExp.textContent = "-";
    els.statSkillPoints.textContent = "-";
    setStatPairVisible(els.statResourceLabel, els.statResource, false);
    els.statResource.textContent = "-";
    els.statMpLabel.textContent = "MP";
    els.statHp.textContent = "-";
    els.statMp.textContent = "-";
    els.statHpRegen5.textContent = "-";
    setStatPairVisible(els.statMpRegen5Label, els.statMpRegen5, false);
    els.statMpRegen5.textContent = "-";
    els.statAttack.textContent = "-";
    els.statAbilityPower.textContent = "-";
    els.statAbilityHaste.textContent = "-";
    els.statPhysicalDefense.textContent = "-";
    els.statMagicDefense.textContent = "-";
    els.statMoveSpeed.textContent = "-";
    els.statAttackRange.textContent = "-";
    els.statAttackSpeed.textContent = "-";
    els.statCritChance.textContent = "-";
    els.abilityHasteBtn.textContent = "+200 Haste";
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
      : "MAX";
  els.statSkillPoints.textContent = player.skillPoints || 0;
  const resourceLabel = formatResource(heroConfig.resource);
  const hasResource = resourceLabel !== "";
  setStatPairVisible(els.statResourceLabel, els.statResource, hasResource);
  els.statResource.textContent = hasResource ? resourceLabel : "-";
  els.statHp.textContent = formatHpWithShield(player);
  els.statMpLabel.textContent = isSword ? "Sword Intent" : "MP";
  els.statMp.textContent = isSword
    ? formatSwordIntent(passive)
    : stats.maxMp > 0
      ? `${formatNumber(stats.mp)}/${formatNumber(stats.maxMp)}`
      : "-";
  els.statHpRegen5.textContent = formatHpRegen5(player);
  const showMpRegen = !isSword && stats.maxMp > 0;
  setStatPairVisible(els.statMpRegen5Label, els.statMpRegen5, showMpRegen);
  els.statMpRegen5.textContent = showMpRegen
    ? formatNumber(stats.mpRegen5 || 0)
    : "-";
  els.statAttack.textContent = formatAttack(stats);
  els.statAbilityPower.textContent = stats.abilityPower || 0;
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
  els.abilityHasteBtn.textContent =
    (stats.abilityHaste || 0) >= 200 ? "Close 200 Haste" : "+200 Haste";
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
    <div>Team ${target.team || "-"}</div>
    ${airborneTicks > 0 ? `<div>Airborne ${(airborneTicks / state.tickRate).toFixed(1)}s</div>` : ""}
    <div>HP ${formatHpWithShield(target)}</div>
    ${formatTargetResource(target)}
    <div>HP/5s ${formatHpRegen5(target)}</div>
    ${formatTargetMpRegen(stats)}
    <div>ATK ${formatAttack(stats)}</div>
    <div>AP ${stats.abilityPower || 0}</div>
    <div>CD ${formatNumber(stats.abilityHaste || 0)}</div>
    <div>Phys DEF ${formatDefenseTip(stats.physicalDefense || 0, "物理")} ${formatPhysicalDefense(stats)}</div>
    <div>Magic DEF ${formatDefenseTip(stats.magicDefense || 0, "魔法")} ${formatMagicDefense(stats)}</div>
    <div>Move SPD ${stats.moveSpeed}</div>
    <div>ATK Range ${stats.attackRange}</div>
    <div>ATK SPD ${stats.attackSpeed}</div>
    <div>Crit ${Math.round((stats.critChance || 0) * 1000) / 10}%</div>
  `;
}
