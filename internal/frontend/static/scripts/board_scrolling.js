import { isInBoardView } from "./state.js";

const LINE_DELTA_MODE = 1;
const PAGE_DELTA_MODE = 2;
const LINE_SCROLL_PIXELS = 40;

const USER_INTERFACE_SELECTORS = [
  ".list",
  ".add-list",
  "#topbar",
  "#app-footer",
  ".bottom-actions-container",
  ".overlay",
  ".modal",
  "#modal-root",
  "#confirm-root",
  "#settings-root",
  "#lightbox-root",
  ".tag-suggest",
];

function isEventInsideListOrUserInterface(targetElement) {
  if (!targetElement || !(targetElement instanceof Element)) {
    return false;
  }
  return USER_INTERFACE_SELECTORS.some((selector) => targetElement.closest(selector));
}

function normalizeWheelDelta(wheelEvent, listsContainerElement) {
  if (wheelEvent.deltaMode === LINE_DELTA_MODE) {
    return wheelEvent.deltaY * LINE_SCROLL_PIXELS;
  }
  if (wheelEvent.deltaMode === PAGE_DELTA_MODE) {
    return wheelEvent.deltaY * listsContainerElement.clientWidth;
  }
  return wheelEvent.deltaY;
}

function handleBoardSpaceWheelScroll(wheelEvent) {
  if (!isInBoardView() || wheelEvent.deltaY === 0) {
    return;
  }
  const listsContainerElement = document.getElementById("lists");
  if (!listsContainerElement || listsContainerElement.scrollWidth <= listsContainerElement.clientWidth) {
    return;
  }
  if (isEventInsideListOrUserInterface(wheelEvent.target)) {
    return;
  }
  wheelEvent.preventDefault();
  listsContainerElement.scrollLeft += normalizeWheelDelta(wheelEvent, listsContainerElement);
}

export function installBoardSpaceWheelScrolling() {
  window.addEventListener("wheel", handleBoardSpaceWheelScroll, { passive: false });
}
