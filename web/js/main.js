const version = "20260703-1";

const scripts = [
  "https://cdn.jsdelivr.net/npm/pixi.js@8.2.6/dist/pixi.min.js",
  `/js/pixi-state.js?v=${version}`,
  `/js/pixi-config.js?v=${version}`,
  `/js/pixi-network.js?v=${version}`,
  `/js/utils/format.js?v=${version}`,
  `/js/input/targeting.js?v=${version}`,
  `/js/visuals/icons.js?v=${version}`,
  `/js/pixi-render.js?v=${version}`,
  `/js/render/effects.js?v=${version}`,
  `/js/render/windups.js?v=${version}`,
  `/js/render/entities.js?v=${version}`,
  `/js/render/hud.js?v=${version}`,
  `/js/ui/skills.js?v=${version}`,
  `/js/ui/status.js?v=${version}`,
  `/js/ui/stats.js?v=${version}`,
  `/js/ui/equipment.js?v=${version}`,
  `/js/ui/target.js?v=${version}`,
  `/js/pixi-ui.js?v=${version}`,
  `/js/input/mouse.js?v=${version}`,
  `/js/pixi-input.js?v=${version}`,
];

for (const src of scripts) {
  await loadScript(src);
}

function loadScript(src) {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = src;
    script.async = false;
    script.onload = resolve;
    script.onerror = () => reject(new Error(`Failed to load ${src}`));
    document.head.appendChild(script);
  });
}
