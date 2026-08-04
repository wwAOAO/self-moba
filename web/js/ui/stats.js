// HUD 图标按英雄和阵营缓存，避免属性刷新时重复读取 GPU 像素。
const heroPortraitCache = new Map();

/** 将统一的 Pixi 英雄图标渲染为 HUD 可复用的透明位图。 */
async function renderHeroPortrait(player) {
    const graphics = new PIXI.Graphics();
    drawHeroModelIcon(graphics, { ...player, dead: false }, 20, true);
    try {
        const canvas = await app.renderer.extract.canvas({ target: graphics, resolution: 2, antialias: true });
        return canvas.toDataURL('image/png');
    } finally {
        graphics.destroy();
    }
}

/** 更新固定屏幕尺寸的 HUD 英雄图标，并在渲染失败时退回名称首字。 */
function setHeroPortrait(player, heroName) {
    els.heroPortrait.title = heroName;
    els.heroPortrait.setAttribute('aria-label', heroName);
    if (!player?.heroId) {
        els.heroPortrait.dataset.iconKey = '';
        els.heroPortrait.classList.remove('has-icon');
        els.heroPortrait.style.removeProperty('--hero-icon');
        els.heroPortrait.textContent = heroName.slice(0, 1);
        return;
    }

    const modelAssetPath = heroModelAssetPaths[player.heroId];
    if (modelAssetPath) {
        const modelIconKey = `model:${player.heroId}`;
        if (els.heroPortrait.dataset.iconKey !== modelIconKey) {
            els.heroPortrait.dataset.iconKey = modelIconKey;
            els.heroPortrait.textContent = '';
            els.heroPortrait.style.setProperty('--hero-icon', `url("${modelAssetPath}")`);
            els.heroPortrait.classList.add('has-icon');
        }
        return;
    }

    const iconKey = `${player.heroId}:${player.team || 'neutral'}`;
    if (
        els.heroPortrait.dataset.iconKey === iconKey &&
        (els.heroPortrait.classList.contains('has-icon') || heroPortraitCache.has(iconKey))
    ) {
        return;
    }
    els.heroPortrait.dataset.iconKey = iconKey;
    els.heroPortrait.textContent = '';
    let iconPromise = heroPortraitCache.get(iconKey);
    if (!iconPromise) {
        iconPromise = renderHeroPortrait(player).catch(error => {
            heroPortraitCache.delete(iconKey);
            throw error;
        });
        heroPortraitCache.set(iconKey, iconPromise);
    }
    iconPromise
        .then(dataUrl => {
            if (els.heroPortrait.dataset.iconKey !== iconKey) {
                return;
            }
            els.heroPortrait.style.setProperty('--hero-icon', `url("${dataUrl}")`);
            els.heroPortrait.classList.add('has-icon');
        })
        .catch(() => {
            if (els.heroPortrait.dataset.iconKey !== iconKey) {
                return;
            }
            els.heroPortrait.classList.remove('has-icon');
            els.heroPortrait.textContent = heroName.slice(0, 1);
        });
}

/** 更新本地英雄属性与底部生命、资源信息。 */
function setStatsCard(player) {
    if (!player?.stats) {
        els.statLevel.textContent = '-';
        els.statExp.textContent = '-';
        els.statSkillPoints.textContent = '-';
        setEquipmentCard(null);
        setStatPairVisible(els.statResourceLabel, els.statResource, false);
        els.statResource.textContent = '-';
        els.statMpLabel.textContent = '法力';
        els.statHp.textContent = '-';
        els.statMp.textContent = '-';
        els.statHpRegen5.textContent = '-';
        setStatPairVisible(els.statMpRegen5Label, els.statMpRegen5, false);
        els.statMpRegen5.textContent = '-';
        els.statAttack.textContent = '-';
        els.statAbilityPower.textContent = '-';
        setHtmlIfChanged(els.statAbilityHasteTip, '');
        els.statAbilityHaste.textContent = '-';
        els.statPhysicalDefense.textContent = '-';
        els.statMagicDefense.textContent = '-';
        els.statMoveSpeed.textContent = '-';
        els.statAttackRange.textContent = '-';
        els.statAttackSpeed.textContent = '-';
        setHtmlIfChanged(els.statCritChanceTip, '');
        els.statCritChance.textContent = '-';
        els.statOmnivamp.textContent = '-';
        els.statLifeSteal.textContent = '-';
        els.statHealingPower.textContent = '-';
        setHeroPortrait(null, '英雄头像');
        els.hudHpFill.style.width = '0';
        els.hudResourceFill.style.width = '0';
        els.hudResourceFill.parentElement.hidden = true;
        els.abilityHasteBtn.textContent = '+200急速';
        return;
    }
    const stats = player.stats;
    const heroConfig = heroClientConfig[player.heroId || els.heroId.value] || {};
    const resourceKind = entityResourceKind(player, heroConfig);
    const heroName = heroDisplayName(heroConfig) || player.heroId || '英雄';
    setHeroPortrait(player, heroName);
    els.statLevel.textContent = `${player.level || 1}/${player.maxLevel || levelClientConfig.maxLevel || 18}`;
    els.statExp.textContent =
        player.nextLevelExp > 0 ? `${Math.floor(player.exp || 0)}/${Math.floor(player.nextLevelExp)}` : '满级';
    els.statSkillPoints.textContent = player.skillPoints || 0;
    setEquipmentCard(player);
    const resourceLabel = formatResource(resourceKind);
    const hasResource = resourceLabel !== '';
    els.hudResourceFill.parentElement.hidden = !hasResource;
    setStatPairVisible(els.statResourceLabel, els.statResource, hasResource);
    els.statResource.textContent = hasResource ? resourceLabel : '-';
    els.statHp.textContent = formatHpWithShield(player);
    els.statMpLabel.textContent = resourceLabel || '资源';
    els.statMp.textContent = formatEntityResourceValue(player, heroConfig);
    els.hudHpFill.style.width = `${Math.min(100, Math.max(0, ((stats.hp || 0) / (stats.maxHp || 1)) * 100))}%`;
    els.hudResourceFill.style.width = `${Math.min(100, Math.max(0, playerResourceRatio(player) * 100))}%`;
    els.hudResourceFill.style.background = `#${playerResourceColor(player).toString(16).padStart(6, '0')}`;
    els.statHpRegen5.textContent = formatHpRegen5(player);
    const showMpRegen = resourceKind === 'mp' && stats.maxMp > 0;
    setStatPairVisible(els.statMpRegen5Label, els.statMpRegen5, showMpRegen);
    els.statMpRegen5.textContent = showMpRegen
        ? formatNumber((stats.mpRegen5 || 0) + equipmentPercentRegen5(player, 'mp'))
        : '-';
    els.statAttack.textContent = formatInteger(stats.attack);
    els.statAbilityPower.textContent = formatInteger(stats.abilityPower);
    setHtmlIfChanged(els.statAbilityHasteTip, formatAbilityHasteTip(stats.abilityHaste || 0));
    els.statAbilityHaste.textContent = formatInteger(stats.abilityHaste);
    els.statPhysicalDefense.textContent = formatInteger(stats.physicalDefense);
    setHtmlIfChanged(els.statPhysicalDefenseTip, formatDefenseTip(stats.physicalDefense || 0, '物理'));
    els.statMagicDefense.textContent = formatInteger(stats.magicDefense);
    setHtmlIfChanged(els.statMagicDefenseTip, formatDefenseTip(stats.magicDefense || 0, '魔法'));
    els.statMoveSpeed.textContent = formatInteger(stats.moveSpeed);
    els.statAttackRange.textContent = formatNumber(stats.attackRange);
    els.statAttackSpeed.textContent = formatNumber(stats.attackSpeed);
    setHtmlIfChanged(els.statCritChanceTip, formatCritChanceTip(player));
    els.statCritChance.textContent = `${formatInteger((stats.critChance || 0) * 100)}%`;
    els.statOmnivamp.textContent = formatPercent(stats.omnivamp || 0);
    els.statLifeSteal.textContent = formatPercent(stats.lifeSteal || 0);
    els.statHealingPower.textContent = formatPercent(stats.healingPower || 0);
    els.abilityHasteBtn.textContent = (player.buffs || []).some(buff => buff.id === 'debug_ability_haste')
        ? '关闭200急速'
        : '+200急速';
}

function formatAbilityPercent(value) {
    return `${Math.round(value * 1000) / 10}%`;
}

function formatAbilityHasteTip(abilityHaste) {
    const reduction = abilityHaste / (100 + abilityHaste);
    return `<span class="stat-tip" data-tip="实际减少 ${formatAbilityPercent(reduction)} 冷却">?</span>`;
}
