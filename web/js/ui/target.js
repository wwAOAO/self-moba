function setTargetCard(target) {
  const targetCards =
    els.target.closest(".target-cards") || els.target.parentElement;
  if (!target?.stats) {
    targetCards.style.display = "none";
    setHtmlIfChanged(els.target, "-");
    return;
  }
  const stats = target.stats;
  const heroConfig = heroClientConfig[target.heroId] || {};
  const resourceKind = entityResourceKind(target, heroConfig);
  const hpRatio = Math.min(100, Math.max(0, ((stats.hp || 0) / (stats.maxHp || 1)) * 100));
  const resourceRatio = Math.min(100, Math.max(0, ratio(
    resourceKind === "sword_intent" ? target.passive?.swordIntent || 0 : stats.mp || 0,
    resourceKind === "sword_intent" ? target.passive?.maxSwordIntent || 0 : stats.maxMp || 0,
  ) * 100));
  const resourceColor = { sword_intent: "f8fafc", rage: "ef4444", energy: "facc15" }[resourceKind] || "3b82f6";
  const airborneTicks = Math.max(
    0,
    (target.control?.airborneUntilTick || 0) -
      Number(els.tick.textContent || 0),
  );
  const targetID = target.id || target.playerId || "";
  const idRow =
    targetID && !targetID.startsWith("spawn:") ? `<span>${escapeHtml(targetID)}</span>` : "";
  targetCards.style.display = "block";
  setHtmlIfChanged(els.target, `
    <div class="target-heading"><strong>${targetLabel(target)}</strong>${idRow}</div>
    ${airborneTicks > 0 ? `<div>击飞 ${(airborneTicks / state.tickRate).toFixed(1)}s</div>` : ""}
    <div class="target-vitals">
      <div class="vital hp"><div class="vital-fill" style="width:${hpRatio}%"></div><span>${formatHpWithShield(target)}</span></div>
      ${resourceKind === "none" ? "" : `<div class="vital resource"><div class="vital-fill" style="width:${resourceRatio}%;background:#${resourceColor}"></div><span>${formatEntityResourceValue(target, heroConfig)}</span></div>`}
    </div>
    <div class="target-attributes" aria-label="目标属性">
      <div class="attribute" title="攻击力"><span class="attribute-icon attack-icon">⚔</span><span class="attribute-value">${formatInteger(stats.attack)}</span></div>
      <div class="attribute" title="法术强度"><span class="attribute-icon ability-power-icon">✦</span><span class="attribute-value">${formatInteger(stats.abilityPower)}</span></div>
      <div class="attribute" title="物理防御"><span class="attribute-icon armor-icon" aria-hidden="true"></span><span class="attribute-value">${formatInteger(stats.physicalDefense)}</span></div>
      <div class="attribute" title="魔法防御"><span class="attribute-icon magic-resist-icon" aria-hidden="true"></span><span class="attribute-value">${formatInteger(stats.magicDefense)}</span></div>
      <div class="attribute" title="攻击速度"><span class="attribute-icon">»</span><span class="attribute-value">${formatNumber(stats.attackSpeed)}</span></div>
      <div class="attribute" title="冷却"><span class="attribute-icon">◷</span><span class="attribute-value">${formatInteger(stats.abilityHaste)}</span></div>
      <div class="attribute" title="暴击率"><span class="attribute-icon crit-icon">✹</span><span class="attribute-value">${formatInteger((stats.critChance || 0) * 100)}%</span></div>
      <div class="attribute" title="移动速度"><span class="attribute-icon">➤</span><span class="attribute-value">${formatInteger(stats.moveSpeed)}</span></div>
    </div>
  `);
}
