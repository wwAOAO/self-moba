window.addEventListener("keydown", (event) => {
  const slot = event.key.toLowerCase();
  if (slot === "a") {
    event.preventDefault();
    attackMoveAtPoint(state.aimPoint || state.moveTarget || { x: state.map.width / 2, y: state.map.height / 2 });
    return;
  }
  if (!["q", "w", "e", "r"].includes(slot)) {
    return;
  }
  event.preventDefault();
  if (event.shiftKey) {
    upgradeSkill(slot);
    return;
  }
  castSkill(slot);
});

els.skills.addEventListener("pointerdown", (event) => {
  const button = event.target.closest("[data-skill-upgrade]");
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

els.connectBtn.addEventListener("click", connect);
els.leaveBtn.addEventListener("click", leave);
els.spawnBtn.addEventListener("click", spawnObject);
els.buyEquipmentBtn.addEventListener("click", buyEquipment);
els.sellEquipmentBtn.addEventListener("click", sellSelectedEquipment);
els.equipmentSlots.forEach((slot, index) => {
  slot.addEventListener("click", () => {
    state.selectedEquipmentSlot = index + 1;
  });
});
els.levelUpBtn.addEventListener("click", debugLevelUp);
els.abilityHasteBtn.addEventListener("click", toggleDebugAbilityHaste);
els.goldBtn.addEventListener("click", debugAddGold);

els.serverUrl.value = websocketURL();
bootPixi();
