async function bootPixi() {
  await app.init({
    resizeTo: els.stage,
    antialias: true,
    backgroundColor: 0x9eb89f,
  });
  els.stage.appendChild(app.canvas);
  app.stage.addChild(worldLayer);
  worldLayer.addChild(gridLayer);
  worldLayer.addChild(unitLayer);
  worldLayer.addChild(playerLayer);
  worldLayer.addChild(effectLayer);
  app.ticker.add(draw);
  app.canvas.addEventListener("contextmenu", (event) => {
    event.preventDefault();
  });
  app.canvas.addEventListener("pointermove", (event) => {
    state.aimPoint = screenToWorld(event);
  });
  app.canvas.addEventListener("pointerdown", handlePointerDown);
  loadGameConfigs();
}

async function loadGameConfigs() {
  await Promise.all([
    loadHeroConfig(),
    loadSkillConfig(),
    loadLevelConfig(),
    loadRewardConfig(),
    loadEquipmentConfig(),
  ]);
}

async function fetchConfig(path) {
  const response = await fetch(path, { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`${path} ${response.status}`);
  }
  return response.json();
}

async function loadHeroConfig() {
  try {
    const heroes = await fetchConfig("/configs/heroes.json");
    heroClientConfig = Object.fromEntries(
      heroes.map((hero) => [hero.heroId, hero]),
    );
    heroSkillSlots = Object.fromEntries(
      heroes.map((hero) => [hero.heroId, hero.skills || {}]),
    );
    renderHeroOptions(heroes);
  } catch (error) {
    console.warn("load hero config failed", error);
  }
}

async function loadSkillConfig() {
  try {
    const skills = await fetchConfig("/configs/skills.json");
    skillClientConfig = Object.fromEntries(
      skills.map((skill) => [
        skill.skillId,
        {
          ...skill,
          ...(skill.meta || {}),
          meta: skill.meta || {},
          metaLists: skill.metaLists || {},
        },
      ]),
    );
  } catch (error) {
    console.warn("load skill config failed", error);
  }
}

async function loadLevelConfig() {
  try {
    levelClientConfig = await fetchConfig("/configs/levels.json");
  } catch (error) {
    console.warn("load level config failed", error);
  }
}

async function loadRewardConfig() {
  try {
    rewardClientConfig = await fetchConfig("/configs/rewards.json");
  } catch (error) {
    console.warn("load reward config failed", error);
  }
}

async function loadEquipmentConfig() {
  try {
    const equipment = await fetchConfig("/configs/equipment.json");
    equipmentClientConfig = Object.fromEntries(
      equipment.map((item) => [item.equipmentId, item]),
    );
    renderEquipmentOptions(equipment);
  } catch (error) {
    console.warn("load equipment config failed", error);
  }
}

function renderEquipmentOptions(equipment) {
  if (!equipment.length) {
    return;
  }
  els.shopItem.innerHTML = equipment
    .map(
      (item) =>
        `<option value="${escapeHtml(item.equipmentId)}">${escapeHtml(item.name || item.equipmentId)} ${item.price || 0}G</option>`,
    )
    .join("");
}

function renderHeroOptions(heroes) {
  if (!heroes.length) {
    return;
  }
  const selected = els.heroId.value;
  els.heroId.innerHTML = heroes
    .map(
      (hero) =>
        `<option value="${escapeHtml(hero.heroId)}">${escapeHtml(hero.name || hero.heroId)}</option>`,
    )
    .join("");
  if (heroClientConfig[selected]) {
    els.heroId.value = selected;
  }
}
