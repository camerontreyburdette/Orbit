import { COLOR_HEX_DEFINITIONS } from "./constants.js";
import { createElement } from "./dom.js";

function capitalizeColorName(colorName) {
  return colorName.charAt(0).toUpperCase() + colorName.slice(1);
}

export function createColorSwatchRow(containerClass, selectedColor, onSelectColor) {
  return createElement(
    "div",
    { class: containerClass },
    createElement("button", {
      class: "color-swatch none" + (selectedColor === "" ? " selected" : ""),
      dataset: { tooltip: "No color" },
      onclick: () => onSelectColor(""),
    }),
    Object.entries(COLOR_HEX_DEFINITIONS).map(([colorName, colorHex]) =>
      createElement("button", {
        class: "color-swatch" + (selectedColor === colorHex ? " selected" : ""),
        style: { background: colorHex },
        dataset: { tooltip: capitalizeColorName(colorName) },
        onclick: () => onSelectColor(colorHex),
      })
    )
  );
}
