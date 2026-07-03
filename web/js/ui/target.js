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
