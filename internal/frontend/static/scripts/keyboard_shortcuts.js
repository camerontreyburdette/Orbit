import { state } from "./state.js";
import { isInputElementFocused } from "./inline_editing.js";
import { hasOpenConfirmOverlay } from "./overlays.js";
import { closeLightboxModal, isLightboxModalOpen } from "./lightbox.js";
import { hasAnySelection, clearSelectionAndRender } from "./selection.js";
import { closeCardModalDialog, deleteOpenCardAction } from "./card_modal.js";
import { renderMainApplicationView } from "./render.js";

function handleEscapeKey() {
  if (isLightboxModalOpen()) {
    closeLightboxModal();
  } else if (hasAnySelection()) {
    clearSelectionAndRender();
  } else if (state.openCardIdentifier != null) {
    closeCardModalDialog();
  } else if (state.composerState) {
    state.composerState = null;
    renderMainApplicationView();
  }
}

function canDeleteOpenCardWithKeyboard() {
  return state.openCardIdentifier != null && !isInputElementFocused() && !isLightboxModalOpen();
}

function handleGlobalKeyDown(keyboardEvent) {
  if (hasOpenConfirmOverlay()) {
    return;
  }
  if (keyboardEvent.key === "Escape") {
    handleEscapeKey();
  } else if (keyboardEvent.key === "Delete" && canDeleteOpenCardWithKeyboard()) {
    deleteOpenCardAction();
  }
}

export function installGlobalKeyboardShortcuts() {
  document.addEventListener("keydown", handleGlobalKeyDown);
}
