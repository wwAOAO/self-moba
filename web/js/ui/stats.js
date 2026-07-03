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
