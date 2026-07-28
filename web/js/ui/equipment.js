function setEquipmentCard(player) {
    const gold = player ? Math.floor(player.gold || 0) : '-';
    els.equipGold.textContent = gold;
    els.shopGold.textContent = gold;
    els.buyEquipmentBtn.disabled = !player || !state.selectedShopItem;
    const equipments = Array.isArray(player?.equipment) ? player.equipment : [];
    const selectedEquipment = equipments[state.selectedEquipmentSlot - 1];
    const hasSelectedEquipment = equipmentName(selectedEquipment) !== '-';
    if (!hasSelectedEquipment) {
        state.selectedEquipmentSlot = 0;
    }
    els.sellEquipmentBtn.disabled = !hasSelectedEquipment;
    els.sellEquipmentBtn.textContent = hasSelectedEquipment
        ? `出售 ${equipmentSellPrice(selectedEquipment)} G`
        : '出售选中';
    els.equipmentSlots.forEach((slot, index) => {
        const equipment = equipments[index];
        const name = equipmentName(equipment);
        slot.textContent = name;
        slot.disabled = !player || name === '-';
        slot.classList.toggle('selected', state.selectedEquipmentSlot === index + 1);
        setEquipmentTip(els.equipmentTips[index], equipment);
    });
    renderShopEquipment(equipments, Boolean(player));
}

function renderShopEquipment(equipments, hasPlayer) {
    els.shopEquipmentSlots.innerHTML = Array.from({ length: 6 }, (_, index) => {
        const equipment = equipments[index];
        const name = equipmentName(equipment);
        const selected = state.selectedEquipmentSlot === index + 1;
        const tip = formatEquipmentTip(equipment);
        const title =
            name === '-'
                ? `装备槽 ${index + 1}`
                : `${name}\n出售 ${equipmentSellPrice(equipment)} G${tip ? `\n${tip}` : ''}`;
        return `<button class="shop-equipment-slot${selected ? ' selected' : ''}" type="button" data-shop-equipment-slot="${index + 1}" title="${escapeHtml(title)}"${!hasPlayer || name === '-' ? ' disabled' : ''}>${escapeHtml(name)}</button>`;
    }).join('');
}

function equipmentSellPrice(equipment) {
    const config = equipmentConfig(equipment);
    return Math.floor((config?.price || 0) * (config?.sellRatio || 0.5));
}

function openShop() {
    els.shopOverlay.hidden = false;
    renderShopItems();
    els.shopSearch.focus();
}

function closeShop() {
    els.shopOverlay.hidden = true;
}

function toggleShop() {
    if (els.shopOverlay.hidden) {
        openShop();
        return;
    }
    closeShop();
}

function renderShopItems(equipment = Object.values(equipmentClientConfig)) {
    const term = els.shopSearch.value.trim().toLowerCase();
    const items = equipment
        .filter(item => state.shopCategory === 'all' || equipmentCategory(item) === state.shopCategory)
        .filter(item => {
            if (!term) {
                return true;
            }
            return [item.name, item.equipmentId, ...(item.description || [])].join(' ').toLowerCase().includes(term);
        })
        .sort(
            (left, right) =>
                equipmentTier(left) - equipmentTier(right) ||
                (left.price || 0) - (right.price || 0) ||
                String(left.name || left.equipmentId).localeCompare(String(right.name || right.equipmentId)),
        );

    els.shopGrid.innerHTML = items.length
        ? items.map(renderShopItem).join('')
        : '<div class="shop-empty">没有匹配的物品</div>';
    if (items.length && !items.some(item => item.equipmentId === state.selectedShopItem)) {
        selectShopItem(items[0].equipmentId);
        return;
    }
    updateShopSelection();
}

function renderShopItem(item) {
    const id = escapeHtml(item.equipmentId);
    const name = escapeHtml(item.name || item.equipmentId);
    const category = equipmentCategory(item);
    return `<button class="shop-item" type="button" data-shop-item="${id}" title="${name}">
    <span class="shop-item-icon" data-category="${category}">${shopItemGlyph(category)}</span>
    <strong>${name}</strong>
    <span class="shop-item-price">${item.price || 0} G</span>
  </button>`;
}

function shopItemGlyph(category) {
    return { physical: '⚔', magic: '✦', defense: '⬟', shoes: '➤' }[category] || '◆';
}

function selectShopItem(equipmentId) {
    const item = equipmentClientConfig[equipmentId];
    if (!item) {
        return;
    }
    state.selectedShopItem = equipmentId;
    els.shopItem.value = equipmentId;
    updateShopSelection();
    renderShopDetail(item);
    els.buyEquipmentBtn.disabled = !state.players.has(state.playerId);
}

function updateShopSelection() {
    for (const button of els.shopGrid.querySelectorAll('[data-shop-item]')) {
        button.classList.toggle('selected', button.dataset.shopItem === state.selectedShopItem);
    }
    for (const button of document.querySelectorAll('[data-shop-category]')) {
        button.classList.toggle('selected', button.dataset.shopCategory === state.shopCategory);
    }
}

function renderShopDetail(item) {
    const category = equipmentCategory(item);
    const tip = formatEquipmentTip(item).split('\n').map(escapeHtml).join('<br>');
    const components = (item.components || [])
        .map(component => equipmentClientConfig[typeof component === 'string' ? component : component.equipmentId])
        .filter(Boolean)
        .map(component => escapeHtml(component.name || component.equipmentId))
        .join(' + ');
    els.shopDetail.innerHTML = `
    <div class="shop-detail-title">
      <span class="shop-item-icon" data-category="${category}">${shopItemGlyph(category)}</span>
      <div><strong>${escapeHtml(item.name || item.equipmentId)}</strong><div class="shop-detail-price">${item.price || 0} G</div></div>
    </div>
    <div>${tip || '暂无属性说明'}</div>
    ${components ? `<div class="shop-components">合成：${components}</div>` : ''}
  `;
}

function setEquipmentTip(tip, equipment) {
    if (!tip) {
        return;
    }
    const text = formatEquipmentTip(equipment);
    if (!text) {
        setHtmlIfChanged(tip, '');
        return;
    }
    setHtmlIfChanged(tip, `<span class="stat-tip" data-tip="${escapeHtml(text)}">?</span>`);
}

function formatEquipmentTip(equipment) {
    const config = equipmentConfig(equipment);
    if (!config) {
        return '';
    }
    if (Array.isArray(config.description) && config.description.length) {
        return config.description.join('\n');
    }
    const parts = [];
    const stats = config.stats || {};
    addTipStat(parts, stats.attack, '攻击力');
    addTipStat(parts, stats.abilityPower, '法术强度');
    addTipStat(parts, stats.abilityHaste, '技能急速');
    addTipStat(parts, stats.hp, '生命');
    addTipStat(parts, stats.mp, '法力');
    addTipStat(parts, stats.physicalDefense, '物理防御');
    addTipStat(parts, stats.magicDefense, '魔法防御');
    addTipStat(parts, stats.moveSpeed, '移动速度');
    addTipStat(parts, stats.hpRegen5, '生命/5秒');
    addTipStat(parts, stats.mpRegen5, '法力/5秒');
    addTipPercent(parts, stats.attackSpeedBonus, '攻击速度');
    addTipPercent(parts, stats.critChance, '暴击率');
    addTipPercent(parts, stats.moveSpeedPercent, '移动速度');
    addTipPercent(parts, stats.omnivamp, '全能吸血');
    addTipPercent(parts, stats.lifeSteal, '生命偷取');
    addTipPercent(parts, stats.healingPower, '治疗加成');
    addTipPercent(parts, stats.grievousWounds, '重伤');
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
    return parts.join('\n');
}

function formatDamageTypeName(type) {
    if (type === 'physical') {
        return '物理伤害';
    }
    if (type === 'magic') {
        return '魔法伤害';
    }
    if (type === 'true') {
        return '真实伤害';
    }
    return '伤害';
}

function equipmentConfig(equipment) {
    if (!equipment) {
        return null;
    }
    if (typeof equipment === 'string') {
        return equipmentClientConfig[equipment] || null;
    }
    const equipmentId = equipment.equipmentId || equipment.id || '';
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
        return '-';
    }
    if (typeof equipment === 'string') {
        return equipment || '-';
    }
    if (typeof equipment.name === 'string' && equipment.name) {
        return equipment.name;
    }
    if (typeof equipment.equipmentId === 'string' && equipment.equipmentId) {
        return equipment.equipmentId;
    }
    return '-';
}

function setStatPairVisible(label, value, visible) {
    label.style.display = visible ? '' : 'none';
    value.style.display = visible ? '' : 'none';
}
