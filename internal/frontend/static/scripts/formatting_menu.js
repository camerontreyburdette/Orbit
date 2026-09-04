import { COLOR_HEX_DEFINITIONS } from "./constants.js";
import { createElement } from "./dom.js";
import { createColorSwatchRow } from "./color_swatches.js";
import {
  WRAP_MARKERS,
  parseColorWrapper,
  replaceSelectedText,
  selectedTextOf,
  toggleColorWrapper,
  toggleWrapMarker,
} from "./text_formatting.js";

const MENU_VIEWPORT_MARGIN = 8;

const WRAP_FORMAT_BUTTONS = [
  { name: "bold", glyph: "B", label: "Bold" },
  { name: "italic", glyph: "I", label: "Italic" },
  { name: "underline", glyph: "U", label: "Underline" },
  { name: "strikethrough", glyph: "S", label: "Strikethrough" },
];

const COLOR_NAME_BY_HEX = new Map(Object.entries(COLOR_HEX_DEFINITIONS).map(([colorName, colorHex]) => [colorHex, colorName]));

let openMenu = null;

function isFormattableTextField(element) {
  if (!element || !element.tagName) {
    return false;
  }
  const isTextarea = element.tagName === "TEXTAREA";
  const isTextInput = element.tagName === "INPUT" && element.type === "text";
  if (!isTextarea && !isTextInput) {
    return false;
  }
  return !element.closest(".search-box, .tag-search-wrap");
}

function hasTextSelection(inputElement) {
  return inputElement.selectionStart !== inputElement.selectionEnd;
}

function closeFormattingMenu() {
  if (!openMenu) {
    return;
  }
  openMenu.dispose();
  openMenu = null;
}

function selectedColorHexOf(inputElement) {
  const colorWrapper = parseColorWrapper(selectedTextOf(inputElement));
  return colorWrapper ? COLOR_HEX_DEFINITIONS[colorWrapper.colorName] : "";
}

function createWrapFormatButton(definition, applyFormat) {
  return createElement(
    "button",
    {
      class: "icon-button format-menu-button " + definition.name,
      dataset: { tooltip: definition.label },
      onclick: () => applyFormat((selectedText) => toggleWrapMarker(selectedText, WRAP_MARKERS[definition.name])),
    },
    definition.glyph
  );
}

function createMenuElement(inputElement) {
  const applyFormat = (transformSelectedText) => {
    replaceSelectedText(inputElement, transformSelectedText);
    closeFormattingMenu();
    inputElement.focus();
  };
  return createElement(
    "div",
    { class: "format-menu", onmousedown: (mouseEvent) => mouseEvent.preventDefault() },
    WRAP_FORMAT_BUTTONS.map((definition) => createWrapFormatButton(definition, applyFormat)),
    createElement("span", { class: "format-menu-divider" }),
    createColorSwatchRow("format-menu-colors", selectedColorHexOf(inputElement), (colorHex) =>
      applyFormat((selectedText) => toggleColorWrapper(selectedText, COLOR_NAME_BY_HEX.get(colorHex) || ""))
    )
  );
}

function positionMenuElement(menuElement, clientX, clientY) {
  const { width, height } = menuElement.getBoundingClientRect();
  const left = Math.max(MENU_VIEWPORT_MARGIN, Math.min(clientX, window.innerWidth - width - MENU_VIEWPORT_MARGIN));
  const top = Math.max(MENU_VIEWPORT_MARGIN, Math.min(clientY, window.innerHeight - height - MENU_VIEWPORT_MARGIN));
  menuElement.style.left = left + "px";
  menuElement.style.top = top + "px";
}

function createDismissBindings(menuElement, inputElement) {
  const handleOutsideMouseDown = (mouseEvent) => {
    if (!menuElement.contains(mouseEvent.target)) {
      closeFormattingMenu();
    }
  };
  const handleEscapeKey = (keyboardEvent) => {
    if (keyboardEvent.key === "Escape") {
      keyboardEvent.stopPropagation();
      closeFormattingMenu();
    }
  };
  const handleSelectionChange = () => {
    if (!hasTextSelection(inputElement)) {
      closeFormattingMenu();
    }
  };
  return [
    [document, "mousedown", handleOutsideMouseDown, true],
    [window, "keydown", handleEscapeKey, true],
    [document, "selectionchange", handleSelectionChange, false],
    [document, "scroll", closeFormattingMenu, true],
    [window, "resize", closeFormattingMenu, false],
    [window, "blur", closeFormattingMenu, false],
    [inputElement, "blur", closeFormattingMenu, false],
  ];
}

function openFormattingMenu(inputElement, clientX, clientY) {
  closeFormattingMenu();
  const menuElement = createMenuElement(inputElement);
  document.body.append(menuElement);
  positionMenuElement(menuElement, clientX, clientY);

  const dismissBindings = createDismissBindings(menuElement, inputElement);
  for (const [target, eventName, listener, useCapture] of dismissBindings) {
    target.addEventListener(eventName, listener, useCapture);
  }
  openMenu = {
    dispose() {
      for (const [target, eventName, listener, useCapture] of dismissBindings) {
        target.removeEventListener(eventName, listener, useCapture);
      }
      menuElement.remove();
    },
  };
}

function handleContextMenu(contextMenuEvent) {
  const targetElement = contextMenuEvent.target;
  if (!isFormattableTextField(targetElement) || !hasTextSelection(targetElement)) {
    closeFormattingMenu();
    return;
  }
  contextMenuEvent.preventDefault();
  openFormattingMenu(targetElement, contextMenuEvent.clientX, contextMenuEvent.clientY);
}

export function installFormattingContextMenu() {
  document.addEventListener("contextmenu", handleContextMenu);
}
