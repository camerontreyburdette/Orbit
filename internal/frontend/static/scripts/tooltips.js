import { TOOLTIP_SHOW_DELAY_MILLISECONDS } from "./constants.js";
import { createElement } from "./dom.js";

const TOOLTIP_GAP_PIXELS = 6;
const TOOLTIP_VIEWPORT_MARGIN_PIXELS = 8;
const TOOLTIP_TARGET_SELECTOR = "[data-tooltip]";

let isTooltipsEnabled = true;
let activeTargetElement = null;
let tooltipElement = null;
let showTimerIdentifier = 0;

function clearShowTimer() {
  if (showTimerIdentifier) {
    clearTimeout(showTimerIdentifier);
    showTimerIdentifier = 0;
  }
}

function hideTooltip() {
  clearShowTimer();
  activeTargetElement = null;
  if (tooltipElement) {
    tooltipElement.remove();
    tooltipElement = null;
  }
}

function computeTooltipPosition(targetRectangle, tooltipRectangle) {
  const centeredLeft = targetRectangle.left + targetRectangle.width / 2 - tooltipRectangle.width / 2;
  const maximumLeft = window.innerWidth - tooltipRectangle.width - TOOLTIP_VIEWPORT_MARGIN_PIXELS;
  const left = Math.max(TOOLTIP_VIEWPORT_MARGIN_PIXELS, Math.min(centeredLeft, maximumLeft));
  const belowTop = targetRectangle.bottom + TOOLTIP_GAP_PIXELS;
  const fitsBelow = belowTop + tooltipRectangle.height <= window.innerHeight - TOOLTIP_VIEWPORT_MARGIN_PIXELS;
  const top = fitsBelow ? belowTop : targetRectangle.top - TOOLTIP_GAP_PIXELS - tooltipRectangle.height;
  return { left, top };
}

function showTooltipForTarget(targetElement) {
  showTimerIdentifier = 0;
  if (!targetElement.isConnected || targetElement !== activeTargetElement) {
    return;
  }
  tooltipElement = createElement("div", { class: "tooltip" }, targetElement.dataset.tooltip);
  document.body.append(tooltipElement);
  const { left, top } = computeTooltipPosition(targetElement.getBoundingClientRect(), tooltipElement.getBoundingClientRect());
  tooltipElement.style.left = left + "px";
  tooltipElement.style.top = top + "px";
}

function findTooltipTarget(eventTarget) {
  if (!eventTarget || typeof eventTarget.closest !== "function") {
    return null;
  }
  const targetElement = eventTarget.closest(TOOLTIP_TARGET_SELECTOR);
  if (!targetElement || targetElement.disabled || !targetElement.dataset.tooltip) {
    return null;
  }
  return targetElement;
}

function handlePointerOver(mouseEvent) {
  if (!isTooltipsEnabled) {
    return;
  }
  const targetElement = findTooltipTarget(mouseEvent.target);
  if (targetElement === activeTargetElement) {
    return;
  }
  hideTooltip();
  if (!targetElement) {
    return;
  }
  activeTargetElement = targetElement;
  showTimerIdentifier = setTimeout(() => showTooltipForTarget(targetElement), TOOLTIP_SHOW_DELAY_MILLISECONDS);
}

function handlePointerOut(mouseEvent) {
  if (!activeTargetElement) {
    return;
  }
  const nextElement = mouseEvent.relatedTarget;
  if (nextElement && activeTargetElement.contains(nextElement)) {
    return;
  }
  hideTooltip();
}

export function setTooltipsEnabled(enabled) {
  isTooltipsEnabled = Boolean(enabled);
  if (!isTooltipsEnabled) {
    hideTooltip();
  }
}

export function installButtonTooltips() {
  document.addEventListener("mouseover", handlePointerOver);
  document.addEventListener("mouseout", handlePointerOut);
  document.addEventListener("mousedown", hideTooltip, true);
  document.addEventListener("keydown", hideTooltip, true);
  document.addEventListener("scroll", hideTooltip, true);
  window.addEventListener("blur", hideTooltip);
}
