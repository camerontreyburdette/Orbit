import { createElement } from "./dom.js";

const MAXIMUM_STACK_LAYERS = 2;
const STACK_LAYER_OFFSET_PIXELS = 6;

function createStackLayers(width, height, layerCount) {
  const layers = [];
  for (let layerIndex = layerCount; layerIndex >= 1; layerIndex--) {
    const offset = layerIndex * STACK_LAYER_OFFSET_PIXELS;
    layers.push(
      createElement("div", {
        class: "drag-ghost-layer",
        style: { width: width + "px", height: height + "px", transform: `translate(${offset}px, ${offset}px)` },
      })
    );
  }
  return layers;
}

function cloneSourceForGhost(sourceElement, width) {
  const clone = sourceElement.cloneNode(true);
  clone.classList.remove("dragging", "drag-src");
  clone.style.width = width + "px";
  if (sourceElement.classList.contains("list-header")) {
    return createElement("div", { class: "drag-ghost-list", style: { width: width + "px" } }, clone);
  }
  return clone;
}

function createStackedDragGhost(sourceElement, totalCount) {
  const sourceRectangle = sourceElement.getBoundingClientRect();
  const width = Math.round(sourceRectangle.width);
  const height = Math.round(sourceRectangle.height);
  const layerCount = Math.min(MAXIMUM_STACK_LAYERS, totalCount - 1);
  const frontElement = cloneSourceForGhost(sourceElement, width);
  frontElement.classList.add("drag-ghost-front");
  return createElement(
    "div",
    { class: "drag-ghost", style: { width: width + layerCount * STACK_LAYER_OFFSET_PIXELS + "px" } },
    createStackLayers(width, height, layerCount),
    frontElement,
    createElement("span", { class: "drag-ghost-count" }, String(totalCount))
  );
}

export function applyStackedDragImage(dragEvent, sourceElement, totalCount) {
  if (totalCount < 2 || !dragEvent.dataTransfer || typeof dragEvent.dataTransfer.setDragImage !== "function") {
    return;
  }
  const ghostElement = createStackedDragGhost(sourceElement, totalCount);
  document.body.append(ghostElement);
  const sourceRectangle = sourceElement.getBoundingClientRect();
  dragEvent.dataTransfer.setDragImage(
    ghostElement,
    Math.round(dragEvent.clientX - sourceRectangle.left),
    Math.round(dragEvent.clientY - sourceRectangle.top)
  );
  setTimeout(() => ghostElement.remove(), 0);
}
