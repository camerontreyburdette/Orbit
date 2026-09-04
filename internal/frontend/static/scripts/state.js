export const state = {
  boards: [],
  boardIdentifier: null,
  currentBoard: null,
  openCardIdentifier: null,
  composerState: null,
  editingListIdentifier: null,
  searchQuery: "",
  cardSearchQuery: "",
  isSearchOpen: false,
  isEditingDescription: false,
  isEditingTitle: false,
};

export const selectedCardIdentifiers = new Set();
export const selectedListIdentifiers = new Set();

export const renderFlags = {
  isViewEntering: false,
  isComposerCreatingNewItem: false,
  enteringCardIdentifier: null,
  enteringListIdentifier: null,
};

export const cardModalEditing = {
  isChecklistCreating: false,
  editingChecklistTitleIdentifier: null,
  editingChecklistItemIdentifier: null,
  refocusChecklistAddIdentifier: null,
  editingTagIndex: null,
  renamingAttachmentIdentifier: null,
  descriptionEditFromHeight: 0,
};

export function resetCardModalEditingState() {
  state.isEditingDescription = false;
  state.isEditingTitle = false;
  cardModalEditing.isChecklistCreating = false;
  cardModalEditing.editingChecklistTitleIdentifier = null;
  cardModalEditing.editingChecklistItemIdentifier = null;
  cardModalEditing.refocusChecklistAddIdentifier = null;
  cardModalEditing.editingTagIndex = null;
}

export function resetBoardViewState() {
  state.composerState = null;
  state.openCardIdentifier = null;
  state.editingListIdentifier = null;
  state.cardSearchQuery = "";
  state.isSearchOpen = false;
}

export function clearEnteringRenderFlags() {
  renderFlags.isComposerCreatingNewItem = false;
  renderFlags.enteringCardIdentifier = null;
  renderFlags.enteringListIdentifier = null;
}

export function isInBoardView() {
  return state.boardIdentifier != null && Boolean(state.currentBoard);
}
