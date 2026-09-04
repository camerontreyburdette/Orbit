import { state, selectedCardIdentifiers, selectedListIdentifiers } from "./state.js";
import { isInputElementFocused } from "./inline_editing.js";
import { hasOpenConfirmOverlay } from "./overlays.js";
import { isLightboxModalOpen } from "./lightbox.js";
import { renderMainApplicationView } from "./render.js";

export function hasAnySelection() {
  return selectedCardIdentifiers.size > 0 || selectedListIdentifiers.size > 0;
}

export function clearMultiSelectionState() {
  selectedCardIdentifiers.clear();
  selectedListIdentifiers.clear();
}

function toggleMembership(identifierSet, identifier) {
  if (identifierSet.has(identifier)) {
    identifierSet.delete(identifier);
  } else {
    identifierSet.add(identifier);
  }
}

export function toggleCardSelection(cardIdentifier) {
  selectedListIdentifiers.clear();
  toggleMembership(selectedCardIdentifiers, cardIdentifier);
  renderMainApplicationView();
}

export function toggleListSelection(listIdentifier) {
  selectedCardIdentifiers.clear();
  toggleMembership(selectedListIdentifiers, listIdentifier);
  renderMainApplicationView();
}

export function clearSelectionAndRender() {
  clearMultiSelectionState();
  renderMainApplicationView();
}

export function canActivateShiftSelectionMode() {
  if (isInputElementFocused()) {
    return false;
  }
  if (state.openCardIdentifier != null || state.composerState != null) {
    return false;
  }
  if (state.editingListIdentifier != null) {
    return false;
  }
  if (isLightboxModalOpen() || hasOpenConfirmOverlay()) {
    return false;
  }
  return !document.querySelector(".modal-backdrop, .modal, dialog[open], .settings-modal, .lightbox");
}
