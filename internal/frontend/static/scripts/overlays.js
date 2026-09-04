import { createElement, querySelector } from "./dom.js";
import { renderInlineMarkup } from "./markup.js";

export function bindOverlayCloseHandler(overlayElement, closeCallback) {
  let isArmed = false;
  overlayElement.addEventListener("mousedown", (mouseEvent) => {
    isArmed = mouseEvent.target === overlayElement;
  });
  overlayElement.addEventListener("click", (mouseEvent) => {
    if (mouseEvent.target === overlayElement && isArmed) {
      closeCallback();
    }
    isArmed = false;
  });
}

export function closeOverlayElement(targetElement) {
  const overlay =
    targetElement && targetElement.classList && targetElement.classList.contains("overlay")
      ? targetElement
      : targetElement && targetElement.firstElementChild;
  if (overlay && overlay.parentNode) {
    overlay.remove();
  }
}

export function hasOpenConfirmOverlay() {
  return Boolean(querySelector("#confirm-root").firstElementChild);
}

function captureEscapeKey(onEscape) {
  const handleKeyDown = (keyboardEvent) => {
    if (keyboardEvent.key === "Escape") {
      keyboardEvent.stopPropagation();
      onEscape();
    }
  };
  document.addEventListener("keydown", handleKeyDown, true);
  return () => document.removeEventListener("keydown", handleKeyDown, true);
}

export function openModalOverlay(contentElement) {
  const overlayElement = createElement("div", { class: "overlay" }, contentElement);
  const closeModal = () => {
    closeOverlayElement(overlayElement);
    releaseEscapeKey();
  };
  const releaseEscapeKey = captureEscapeKey(closeModal);
  bindOverlayCloseHandler(overlayElement, closeModal);
  querySelector("#confirm-root").append(overlayElement);
  return closeModal;
}

function createDeleteConfirmationElement(entityName) {
  const containerElement = createElement("span", {}, "Are you sure you want to delete ");
  const nameElement = createElement("strong");
  nameElement.innerHTML = renderInlineMarkup(entityName);
  containerElement.append(nameElement, "?");
  return containerElement;
}

function createConfirmDialog(messageElementOrString, confirmLabel, isDestructive) {
  return new Promise((resolve) => {
    const rootElement = querySelector("#confirm-root");
    const complete = (resultValue) => {
      closeOverlayElement(overlayElement);
      document.removeEventListener("keydown", handleKeyDown, true);
      resolve(resultValue);
    };
    const handleKeyDown = (keyboardEvent) => {
      if (keyboardEvent.key === "Escape") {
        keyboardEvent.stopPropagation();
        complete(false);
      } else if (keyboardEvent.key === "Enter") {
        keyboardEvent.stopPropagation();
        keyboardEvent.preventDefault();
        complete(true);
      }
    };
    document.addEventListener("keydown", handleKeyDown, true);
    const confirmButton = createElement(
      "button",
      {
        class: isDestructive ? "button button-danger" : "button button-primary",
        onclick: () => complete(true),
      },
      confirmLabel
    );
    const overlayElement = createElement(
      "div",
      { class: "overlay" },
      createElement(
        "div",
        { class: "confirm-box" },
        createElement("p", {}, messageElementOrString),
        createElement(
          "div",
          { class: "confirm-actions" },
          createElement("button", { class: "button", onclick: () => complete(false) }, "Cancel"),
          confirmButton
        )
      )
    );
    bindOverlayCloseHandler(overlayElement, () => complete(false));
    rootElement.append(overlayElement);
    confirmButton.focus();
  });
}

export function confirmDeletion(entityName) {
  return createConfirmDialog(createDeleteConfirmationElement(entityName), "Delete", true);
}

export function confirmDestructiveAction(messageElementOrString, confirmLabel) {
  return createConfirmDialog(messageElementOrString, confirmLabel, true);
}
