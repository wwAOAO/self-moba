window.addEventListener('keydown', event => {
    const slot = event.key.toLowerCase();
    if (slot === 'escape' && !els.shopOverlay.hidden) {
        closeShop();
        return;
    }
    if (['input', 'select', 'textarea'].includes(event.target.tagName.toLowerCase())) {
        return;
    }
    if (slot === 'p') {
        event.preventDefault();
        toggleShop();
        return;
    }
    if (slot === 'a') {
        event.preventDefault();
        attackMoveAtPoint(state.aimPoint || state.moveTarget || { x: state.map.width / 2, y: state.map.height / 2 });
        return;
    }
    if (!['q', 'w', 'e', 'r'].includes(slot)) {
        return;
    }
    event.preventDefault();
    if (event.shiftKey) {
        upgradeSkill(slot);
        return;
    }
    castSkill(slot);
});

els.skills.addEventListener('pointerdown', event => {
    const button = event.target.closest('[data-skill-upgrade]');
    if (!button) {
        return;
    }
    event.preventDefault();
    event.stopPropagation();
    if (button.disabled) {
        return;
    }
    upgradeSkill(button.dataset.skillUpgrade);
});

els.connectBtn.addEventListener('click', connect);
els.leaveBtn.addEventListener('click', leave);
els.spawnBtn.addEventListener('click', spawnObject);
els.shopToggleBtn.addEventListener('click', toggleShop);
els.shopCloseBtn.addEventListener('click', closeShop);
els.shopOverlay.addEventListener('click', event => {
    if (event.target === els.shopOverlay) {
        closeShop();
    }
});
els.shopSearch.addEventListener('input', () => renderShopItems());
els.shopGrid.addEventListener('click', event => {
    const item = event.target.closest('[data-shop-item]');
    if (item) {
        selectShopItem(item.dataset.shopItem);
    }
});
document.querySelector('.shop-categories').addEventListener('click', event => {
    const category = event.target.closest('[data-shop-category]');
    if (!category) {
        return;
    }
    state.shopCategory = category.dataset.shopCategory;
    renderShopItems();
});
els.buyEquipmentBtn.addEventListener('click', buyEquipment);
els.sellEquipmentBtn.addEventListener('click', sellSelectedEquipment);
els.shopEquipmentSlots.addEventListener('click', event => {
    const slot = event.target.closest('[data-shop-equipment-slot]');
    if (!slot || slot.disabled) {
        return;
    }
    state.selectedEquipmentSlot = Number(slot.dataset.shopEquipmentSlot);
    setEquipmentCard(state.players.get(state.playerId));
});
els.equipmentSlots.forEach((slot, index) => {
    slot.addEventListener('click', () => {
        state.selectedEquipmentSlot = index + 1;
        setEquipmentCard(state.players.get(state.playerId));
    });
});
els.levelUpBtn.addEventListener('click', debugLevelUp);
els.abilityHasteBtn.addEventListener('click', toggleDebugAbilityHaste);
els.goldBtn.addEventListener('click', debugAddGold);

els.serverUrl.value = websocketURL();
bootPixi();
