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
  worldLayer.addChild(skillLayer);
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

async function fetchConfigList(path, fallbackPath) {
  try {
    const manifest = await fetchConfig(`${path}/manifest.json`);
    const parts = await Promise.all(
      manifest.map((file) => fetchConfig(`${path}/${file}`)),
    );
    return parts.flat();
  } catch (error) {
    if (!fallbackPath) {
      throw error;
    }
    return fetchConfig(fallbackPath);
  }
}

async function loadHeroConfig() {
  try {
    const heroes = await fetchConfigList("/configs/heroes");
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
    const skills = await fetchConfigList("/configs/skills");
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
    const equipment = await fetchConfigList("/configs/equipment");
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
  const groups = [
    ["物理装备", "physical"],
    ["魔法装备", "magic"],
    ["防护装备", "defense"],
    ["鞋子", "shoes"],
  ];
  els.shopItem.innerHTML = groups
    .flatMap(([label, category]) => renderEquipmentTierGroups(label, category, equipment))
    .join("");
}

function renderEquipmentTierGroups(label, category, equipment) {
  return [1, 2, 3]
    .map((tier) => [
      `${label} - ${["", "一级装备", "二级装备", "三级装备"][tier]}`,
      equipment.filter(
        (item) => equipmentCategory(item) === category && equipmentTier(item) === tier,
      ),
    ])
    .filter(([, items]) => items.length)
    .map(
      ([groupLabel, items]) =>
        `<optgroup label="${groupLabel}">${items.map(renderEquipmentOption).join("")}</optgroup>`,
    );
}

function renderEquipmentOption(item) {
  const tip = typeof formatEquipmentTip === "function" ? formatEquipmentTip(item) : "";
  const summary = tip ? ` - ${tip.replaceAll("\n", "；")}` : "";
  return `<option value="${escapeHtml(item.equipmentId)}">${escapeHtml(item.name || item.equipmentId)} ${item.price || 0}G${escapeHtml(summary)}</option>`;
}

function equipmentTier(item) {
  return item.tier || (item.components?.length ? 2 : 1);
}

function equipmentCategory(item) {
  if (item.category) {
    return item.category;
  }
  const stats = item.stats || {};
  if (stats.attack || stats.attackSpeedBonus || stats.critChance) {
    return "physical";
  }
  if (stats.abilityPower || stats.mp || stats.mpRegen5) {
    return "magic";
  }
  return "defense";
}

function renderHeroOptions(heroes) {
  if (!heroes.length) {
    return;
  }
  const selected = els.heroId.value;
  els.heroId.innerHTML = heroes
    .map(
      (hero) =>
        `<option value="${escapeHtml(hero.heroId)}">${escapeHtml(heroDisplayName(hero))}</option>`,
    )
    .join("");
  if (heroClientConfig[selected]) {
    els.heroId.value = selected;
  }
}

function heroDisplayName(hero) {
  const names = {
    sword: "剑客",
    warrior: "战士",
    warriors: "战士",
    archer: "弓手",
    mage: "法师",
    tank: "坦克",
    gunner: "枪手",
    blade: "刀客",
    berserker: "狂战士",
    ninja: "忍者",
    frostmage: "冰霜法师",
    robot: "机器人",
    explorer: "探险者",
  };
  return names[hero.heroId] || hero.name || hero.heroId;
}
