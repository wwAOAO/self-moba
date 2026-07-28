function websocketURL() {
    const scheme = location.protocol === 'https:' ? 'wss' : 'ws';
    return `${scheme}://${location.host || 'localhost:6969'}/ws`;
}

function screenPointFromEvent(event) {
    const rect = app.canvas.getBoundingClientRect();
    return {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
    };
}

function screenToWorld(event) {
    const canvasPoint = screenPointFromEvent(event);
    return {
        x: (canvasPoint.x - state.frame.offsetX) / state.frame.scale,
        y: (canvasPoint.y - state.frame.offsetY) / state.frame.scale,
    };
}

function clamp(value, min, max) {
    return Math.max(min, Math.min(max, value));
}

function formatNumber(value) {
    return Number.isInteger(value) ? String(value) : String(Math.round(value * 1000) / 1000);
}

function formatInteger(value) {
    return String(Math.floor(value || 0));
}

function shieldValue(entity) {
    return Math.max(0, entity?.passive?.shield || 0);
}

function formatHpWithShield(entity) {
    const stats = entity?.stats || {};
    const shield = shieldValue(entity);
    if (shield <= 0) {
        return `${formatInteger(stats.hp)}/${formatInteger(stats.maxHp)}`;
    }
    return `${formatInteger(stats.hp)} + ${formatInteger(shield)}/${formatInteger(stats.maxHp)}`;
}

function formatHpRegen5(entity) {
    const stats = entity?.stats || {};
    const base = (stats.hpRegen5 || 0) + equipmentPercentRegen5(entity, 'hp');
    const passive = warriorToughnessRegen5(entity);
    if (passive <= 0) {
        return formatNumber(base);
    }
    return `${formatNumber(base)} + ${formatNumber(passive)}`;
}

function equipmentPercentRegen5(entity, resource) {
    if (!entity?.equipment || !entity?.stats) {
        return 0;
    }
    const outOfCombat = (state.snapshotTick || 0) >= (entity.lastHitTick || 0) + 5 * (state.tickRate || 20);
    return entity.equipment.reduce((total, equipment) => {
        const effects = equipmentConfig(equipment)?.effects || {};
        const ratio = outOfCombat
            ? effects[`outOfCombat${resource === 'hp' ? 'Hp' : 'Mp'}RegenMax${resource === 'hp' ? 'Hp' : 'Mp'}Ratio5`]
            : effects[`combat${resource === 'hp' ? 'Hp' : 'Mp'}RegenMax${resource === 'hp' ? 'Hp' : 'Mp'}Ratio5`];
        return total + (entity.stats[resource === 'hp' ? 'maxHp' : 'maxMp'] || 0) * (ratio || 0);
    }, 0);
}

function warriorToughnessRegen5(entity) {
    if ((entity?.heroId || '') !== 'warrior') {
        return 0;
    }
    const ratios = skillClientConfig.warrior_toughness?.metaLists?.regenMaxHPRatio || [];
    if (ratios.length === 0) {
        return 0;
    }
    const level = clamp(Math.max(1, entity.level || 1), 1, ratios.length);
    const ratio = ratios[level - 1] || 0;
    return (entity.stats?.maxHp || 0) * ratio;
}

function hpShieldRatio(entity) {
    const stats = entity?.stats || {};
    return ratio((stats.hp || 0) + shieldValue(entity), stats.maxHp || 0);
}

function formatAttack(stats) {
    return formatBasePlusBonus(stats.attack || 0, stats.bonusAttack || 0);
}

function formatPhysicalDefense(stats) {
    return formatBasePlusBonus(stats.physicalDefense || 0, stats.bonusPhysicalDefense || 0);
}

function formatMagicDefense(stats) {
    return formatBasePlusBonus(stats.magicDefense || 0, stats.bonusMagicDefense || 0);
}

function formatDefenseTip(resistance, typeLabel) {
    return `<span class="stat-tip" data-tip="${escapeHtml(formatResistanceTip(resistance, typeLabel))}">?</span>`;
}

function formatResistanceTip(resistance, typeLabel) {
    if (resistance >= 0) {
        const reduce = resistance / (resistance + 100);
        return `${typeLabel}伤害减免 ${formatPercent(reduce)}`;
    }
    const multiplier = 100 / Math.max(1, 100 + resistance);
    return `${typeLabel}伤害放大 ${formatNumber(multiplier)} 倍`;
}

function formatPercent(value) {
    return `${formatNumber(value * 100)}%`;
}

function formatCritChance(entity) {
    return `${Math.round((entity?.stats?.critChance || 0) * 1000) / 10}%`;
}

function formatCritChanceTip(entity) {
    return `<span class="stat-tip" data-tip="${escapeHtml(`以${formatCritChance(entity)}的概率造成${formatPercent(2 + entityCritDamageBonus(entity))}伤害`)}">?</span>`;
}

function entityCritDamageBonus(entity) {
    const equipment = Array.isArray(entity?.equipment) ? entity.equipment : [];
    return equipment.reduce((bonus, item) => Math.max(bonus, equipmentConfig(item)?.effects?.critDamageBonus || 0), 0);
}

function formatBasePlusBonus(base, bonus) {
    if (bonus <= 0) {
        return formatNumber(base);
    }
    return `${formatNumber(base)} + ${formatNumber(bonus)}`;
}

function formatSwordIntent(passive) {
    return passive?.maxSwordIntent > 0
        ? `${Math.floor(passive.swordIntent || 0)}/${Math.floor(passive.maxSwordIntent)}`
        : '-';
}

function entityResourceKind(entity, heroConfig) {
    if (heroConfig?.resource) {
        return heroConfig.resource;
    }
    if (entity?.heroId === 'sword') {
        return 'sword_intent';
    }
    if (entity?.heroId === 'blade') {
        return 'rage';
    }
    if (entity?.heroId === 'ninja') {
        return 'energy';
    }
    if (entity?.stats?.maxMp > 0) {
        return 'mp';
    }
    return 'none';
}

function formatEntityResourceValue(entity, heroConfig) {
    const kind = entityResourceKind(entity, heroConfig);
    const stats = entity?.stats || {};
    if (kind === 'none') {
        return '-';
    }
    if (kind === 'sword_intent') {
        return formatSwordIntent(entity?.passive || {});
    }
    return `${formatInteger(stats.mp)}/${formatInteger(stats.maxMp)}`;
}

function formatTargetResource(target) {
    const heroConfig = heroClientConfig[target?.heroId] || {};
    const kind = entityResourceKind(target, heroConfig);
    if (kind === 'none') {
        return '';
    }
    return `<div>${formatResource(kind)} ${formatEntityResourceValue(target, heroConfig)}</div>`;
}

function formatTargetMpRegen(target) {
    const stats = target?.stats || {};
    if (!stats?.maxMp || stats.maxMp <= 0) {
        return '';
    }
    return `<div>法力/5秒 ${formatNumber((stats.mpRegen5 || 0) + equipmentPercentRegen5(target, 'mp'))}</div>`;
}

function formatResource(resource) {
    if (resource === 'sword_intent') {
        return '剑意';
    }
    if (resource === 'rage') {
        return '怒气';
    }
    if (resource === 'energy') {
        return '能量';
    }
    if (!resource || resource === 'mp') {
        return '法力';
    }
    if (resource === 'none') {
        return '';
    }
    return resource;
}

/** 将仍在生效的控制、减益和增益归一化为统一的前端展示状态。 */
function displayStatuses(target, tick) {
    const statuses = [];
    const control = target?.control || {};
    pushTimedStatus(statuses, tick, 'suppressed', '压制', '压', 'control', 700, control.suppressedUntilTick);
    pushTimedStatus(statuses, tick, 'airborne', '击飞', '飞', 'control', 650, control.airborneUntilTick);
    pushTimedStatus(statuses, tick, 'stunned', '眩晕', '晕', 'control', 600, control.stunnedUntilTick);
    pushTimedStatus(statuses, tick, 'taunted', '嘲讽', '嘲', 'control', 550, control.tauntedUntilTick);
    pushTimedStatus(statuses, tick, 'rooted', '禁锢', '锢', 'control', 500, control.rootedUntilTick);
    pushTimedStatus(statuses, tick, 'silenced', '沉默', '默', 'control', 450, control.silencedUntilTick);
    pushTimedStatus(
        statuses,
        tick,
        'move_speed_slow',
        '移动减速',
        '缓',
        'debuff',
        300,
        control.moveSpeedSlowUntil,
        formatStatusPercent(control.moveSpeedSlow),
    );
    pushTimedStatus(
        statuses,
        tick,
        'attack_speed_slow',
        '攻速降低',
        '攻',
        'debuff',
        280,
        control.attackSpeedSlowUntil,
        formatStatusPercent(control.attackSpeedSlow),
    );
    pushTimedStatus(
        statuses,
        tick,
        'grievous_wounds',
        '重伤',
        '伤',
        'debuff',
        260,
        control.grievousWoundsUntil,
        formatStatusPercent(control.grievousWounds),
    );
    pushTimedStatus(
        statuses,
        tick,
        'mage_illumination',
        '启明标记',
        '光',
        'debuff',
        240,
        control.mageIlluminationUntil,
    );

    const knownIDs = new Set(statuses.map(status => status.id));
    for (const buff of target?.buffs || []) {
        if ((buff.expiresAtTick || 0) > 0 && buff.expiresAtTick <= tick) {
            continue;
        }
        if (knownIDs.has(buff.id)) {
            continue;
        }
        statuses.push({
            id: buff.id || 'status',
            name: buff.name || buff.id || '状态',
            icon: statusIconForBuff(buff),
            kind: buff.negative ? 'debuff' : 'buff',
            priority: buff.negative ? 100 : 0,
            expiresAtTick: buff.expiresAtTick || 0,
            stacks: buff.stacks || 0,
            value: '',
            tooltip: buff.tooltip || '',
        });
    }
    return statuses.sort((left, right) => right.priority - left.priority || left.name.localeCompare(right.name));
}

/** 追加一个尚未结束的定时状态。 */
function pushTimedStatus(statuses, tick, id, name, icon, kind, priority, expiresAtTick, value = '') {
    if (!expiresAtTick || expiresAtTick <= tick) {
        return;
    }
    statuses.push({ id, name, icon, kind, priority, expiresAtTick, stacks: 0, value, tooltip: '' });
}

/** 将比例型状态强度格式化为百分比。 */
function formatStatusPercent(value) {
    return value > 0 ? `${Math.round(value * 100)}%` : '';
}

/** 根据 Buff 标识返回便于快速识别的单字图标。 */
function statusIconForBuff(buff) {
    const id = String(buff?.id || '');
    if (id.includes('bleed')) {
        return '血';
    }
    if (id.includes('burn')) {
        return '灼';
    }
    if (id.includes('shred') || id.includes('cleaver')) {
        return '破';
    }
    return buff?.negative ? '减' : '增';
}

/** 格式化状态剩余时间；永久状态不显示倒计时。 */
function formatStatusDuration(status, tick) {
    if (!status?.expiresAtTick) {
        return '';
    }
    const seconds = Math.max(0, (status.expiresAtTick - tick) / state.tickRate);
    return seconds < 1 ? `${seconds.toFixed(1)}s` : `${Math.ceil(seconds)}s`;
}

/** 格式化状态的强度、层数和剩余时间摘要。 */
function formatStatusDetails(status, tick) {
    const details = [];
    if (status?.value) {
        details.push(status.value);
    }
    if ((status?.stacks || 0) > 0) {
        details.push(`${status.stacks}层`);
    }
    const duration = formatStatusDuration(status, tick);
    if (duration) {
        details.push(duration);
    }
    return details.join(' · ');
}

function escapeHtml(value) {
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#039;');
}

function setHtmlIfChanged(element, html) {
    if (element && element.innerHTML !== html) {
        element.innerHTML = html;
    }
}

function ratio(value, max) {
    if (!max || max <= 0) {
        return 0;
    }
    return clamp(value / max, 0, 1);
}
