import { ICON_PATH_DEFINITIONS } from "./constants.js";

const iconTemplateCache = new Map();

function buildIconTemplate(iconName, sizePixels, isFilled) {
  const spanElement = document.createElement("span");
  spanElement.className = "icon";
  spanElement.innerHTML =
    `<svg viewBox="0 0 24 24" width="${sizePixels}" height="${sizePixels}" ` +
    `fill="${isFilled ? "currentColor" : "none"}" stroke="currentColor" ` +
    `stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` +
    ICON_PATH_DEFINITIONS[iconName] +
    "</svg>";
  return spanElement;
}

export function createIconElement(iconName, sizePixels = 16, isFilled = false) {
  const cacheKey = iconName + ":" + sizePixels + ":" + isFilled;
  let templateElement = iconTemplateCache.get(cacheKey);
  if (!templateElement) {
    templateElement = buildIconTemplate(iconName, sizePixels, isFilled);
    iconTemplateCache.set(cacheKey, templateElement);
  }
  return templateElement.cloneNode(true);
}
