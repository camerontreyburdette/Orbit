import { invokeBackendMethod } from "./backend.js";
import { createElement, createInlineElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { renderBlockMarkup, renderInlineMarkup } from "./markup.js";
import { state, renderFlags, selectedCardIdentifiers, selectedListIdentifiers } from "./state.js";
import { createInlineEditInput, scheduleFocusAtEnd } from "./inline_editing.js";
import { pushCurrentSnapshotToUndoHistory } from "./history.js";
import { persistLocalChange, refreshCurrentBoard } from "./board_actions.js";
import { renderMainApplicationView } from "./render.js";
import { canActivateShiftSelectionMode, clearMultiSelectionState, toggleCardSelection, toggleListSelection } from "./selection.js";
import { cleanupDragState, handleCardDragStart, handleCardsDragOver, handleCardsDrop, handleListDragStart } from "./drag_and_drop.js";
import { createFileDropProperties } from "./file_drop.js";
import { createAttachmentImageElement } from "./attachments.js";
import { openCardModalDialog } from "./card_modal.js";

function closeComposer() {
  state.composerState = null;
  renderMainApplicationView();
}

function openComposer(composerState) {
  state.composerState = composerState;
  renderFlags.isComposerCreatingNewItem = true;
  renderMainApplicationView();
}

function createComposerInput(tagName, placeholder, submit, submitsOnShiftEnter) {
  return createElement(tagName, {
    placeholder,
    onkeydown: (keyboardEvent) => {
      if (keyboardEvent.key === "Enter" && (submitsOnShiftEnter || !keyboardEvent.shiftKey)) {
        keyboardEvent.preventDefault();
        submit();
      }
      if (keyboardEvent.key === "Escape") {
        closeComposer();
      }
    },
  });
}

function createComposerElement(inputElement, submitLabel, submit) {
  return createElement(
    "div",
    { class: "composer" + (renderFlags.isComposerCreatingNewItem ? " enter" : "") },
    inputElement,
    createElement(
      "div",
      { class: "composer-actions" },
      createElement("button", { class: "button button-primary", onclick: submit }, submitLabel),
      createElement("button", { class: "icon-button", dataset: { tooltip: "Cancel" }, onclick: closeComposer }, createIconElement("x", 16))
    )
  );
}

function createComposerOpenButton(label, composerState) {
  return createElement("button", { class: "composer-button", onclick: () => openComposer(composerState) }, label);
}

async function submitComposerCreation(inputElement, createEntity, markEntering) {
  const trimmedValue = inputElement.value.trim();
  if (!trimmedValue) {
    return;
  }
  inputElement.value = "";
  let creationResponse;
  try {
    creationResponse = await createEntity(trimmedValue);
  } catch {
    return;
  }
  markEntering(creationResponse.id);
  state.composerState = null;
  pushCurrentSnapshotToUndoHistory();
  refreshCurrentBoard();
}

export function createAddListComposerElement() {
  const wrapperElement = createElement("div", { class: "add-list" });
  if (!(state.composerState && state.composerState.type === "list")) {
    wrapperElement.append(createComposerOpenButton("+ Add list", { type: "list" }));
    return wrapperElement;
  }

  const submitNewList = () =>
    submitComposerCreation(
      inputElement,
      (name) => invokeBackendMethod("create_list", state.boardIdentifier, name),
      (identifier) => {
        renderFlags.enteringListIdentifier = identifier;
      }
    );
  const inputElement = createComposerInput("input", "List name…", submitNewList, true);
  wrapperElement.append(createComposerElement(inputElement, "Add list", submitNewList));
  return wrapperElement;
}

function createAddCardComposerElement(listIdentifier) {
  const isComposerOpen =
    state.composerState && state.composerState.type === "card" && state.composerState.listIdentifier === listIdentifier;
  if (!isComposerOpen) {
    return createComposerOpenButton("+ Add a card", { type: "card", listIdentifier });
  }

  const submitNewCard = () =>
    submitComposerCreation(
      textareaElement,
      (title) => invokeBackendMethod("create_card", listIdentifier, title),
      (identifier) => {
        renderFlags.enteringCardIdentifier = identifier;
      }
    );
  const textareaElement = createComposerInput("textarea", "Card title…", submitNewCard, false);
  return createComposerElement(textareaElement, "Add card", submitNewCard);
}

function createListNameEditor(list) {
  const inputElement = createInlineEditInput({
    value: list.name,
    shouldStopEscapePropagation: false,
    onCancel: () => {
      state.editingListIdentifier = null;
      renderMainApplicationView();
    },
    onCommit: (rawValue) => {
      const trimmedValue = rawValue.trim();
      state.editingListIdentifier = null;
      if (trimmedValue && trimmedValue !== list.name) {
        list.name = trimmedValue;
        persistLocalChange("rename_list", list.id, trimmedValue);
      }
      renderMainApplicationView();
    },
  });
  scheduleFocusAtEnd(inputElement);
  return inputElement;
}

function createListNameElement(list) {
  const nameElement = createElement("h2", {
    onclick: (clickEvent) => {
      if (clickEvent.shiftKey && canActivateShiftSelectionMode()) {
        clickEvent.preventDefault();
        clickEvent.stopPropagation();
        toggleListSelection(list.id);
        return;
      }
      if (selectedListIdentifiers.size > 0) {
        return;
      }
      state.editingListIdentifier = list.id;
      renderMainApplicationView();
    },
  });
  nameElement.innerHTML = renderInlineMarkup(list.name);
  return nameElement;
}

function createListHeaderElement(list) {
  const isEditing = state.editingListIdentifier === list.id;
  return createElement(
    "div",
    {
      class: "list-header",
      draggable: !isEditing && state.currentBoard.lists.length > 1,
      onclick: (clickEvent) => {
        if (clickEvent.shiftKey && canActivateShiftSelectionMode()) {
          clickEvent.preventDefault();
          clickEvent.stopPropagation();
          toggleListSelection(list.id);
        }
      },
      ondragstart: (dragEvent) => handleListDragStart(dragEvent, list.id),
      ondragend: cleanupDragState,
    },
    isEditing ? createListNameEditor(list) : createListNameElement(list)
  );
}

export function createListColumnElement(list) {
  const isListSelected = selectedListIdentifiers.has(list.id);
  const cardsContainerElement = createElement("div", { class: "cards", dataset: { listIdentifier: list.id } });
  for (const currentCard of list.cards) {
    cardsContainerElement.append(createCardItemElement(currentCard, list));
  }

  return createElement(
    "section",
    {
      class:
        "list" + (list.id === renderFlags.enteringListIdentifier ? " enter" : "") + (isListSelected ? " selected" : ""),
      dataset: { listIdentifier: list.id },
      ondragover: (dragEvent) => handleCardsDragOver(dragEvent, cardsContainerElement),
      ondrop: (dragEvent) => handleCardsDrop(dragEvent, cardsContainerElement, list.id),
    },
    createListHeaderElement(list),
    cardsContainerElement,
    createAddCardComposerElement(list.id)
  );
}

function countChecklistProgress(card) {
  let completedCount = 0;
  let totalCount = 0;
  for (const currentChecklist of card.checklists || []) {
    for (const currentItem of currentChecklist.items) {
      totalCount++;
      if (currentItem.done) {
        completedCount++;
      }
    }
  }
  return { completedCount, totalCount };
}

function createCardBadgeElements(card) {
  const badgeElements = [];
  if (card.attachments.length) {
    badgeElements.push(createElement("span", { class: "badge" }, createIconElement("clip", 13), String(card.attachments.length)));
  }
  const { completedCount, totalCount } = countChecklistProgress(card);
  if (totalCount) {
    badgeElements.push(
      createElement(
        "span",
        { class: "badge checklist-badge" + (completedCount === totalCount ? " checklist-complete" : "") },
        createIconElement("check", 13),
        completedCount + "/" + totalCount
      )
    );
  }
  return badgeElements;
}

function createCardDescriptionElement(card) {
  if (!card.description) {
    return null;
  }
  const descriptionElement = createElement("div", { class: "card-description" });
  descriptionElement.innerHTML = renderBlockMarkup(card.description);
  return descriptionElement;
}

function createCardTagsElement(card) {
  if (!(card.tags && card.tags.length)) {
    return null;
  }
  return createElement(
    "div",
    { class: "card-tags" },
    card.tags.map((tagText) =>
      createElement("span", { class: "tag-chip" }, createInlineElement("span", { class: "tag-text" }, tagText))
    )
  );
}

function findCoverAttachment(card) {
  if (!card.cover_id) {
    return null;
  }
  return card.attachments.find((attachmentItem) => attachmentItem.id === card.cover_id && attachmentItem.kind === "image");
}

function handleCardClick(clickEvent, card) {
  if (clickEvent.shiftKey) {
    clickEvent.preventDefault();
    clickEvent.stopPropagation();
    toggleCardSelection(card.id);
    return;
  }
  if (selectedCardIdentifiers.size > 0) {
    clearMultiSelectionState();
    renderMainApplicationView();
  }
  openCardModalDialog(card.id);
}

function handleCardContextMenu(contextMenuEvent, card) {
  contextMenuEvent.preventDefault();
  contextMenuEvent.stopPropagation();
  if (contextMenuEvent.shiftKey) {
    toggleCardSelection(card.id);
  } else if (!selectedCardIdentifiers.has(card.id)) {
    selectedCardIdentifiers.clear();
    selectedCardIdentifiers.add(card.id);
  }
  renderMainApplicationView();
}

export function createCardItemElement(card, list) {
  const coverAttachment = findCoverAttachment(card);
  const isDragAllowed = state.currentBoard.lists.length > 1 || list.cards.length > 1;
  const isCardSelected = selectedCardIdentifiers.has(card.id);
  const badgeElements = createCardBadgeElements(card);

  const cardArticleElement = createElement(
    "article",
    {
      class:
        "card" + (card.id === renderFlags.enteringCardIdentifier ? " enter" : "") + (isCardSelected ? " selected" : ""),
      draggable: isDragAllowed,
      dataset: { cardIdentifier: card.id },
      onclick: (clickEvent) => handleCardClick(clickEvent, card),
      oncontextmenu: (contextMenuEvent) => handleCardContextMenu(contextMenuEvent, card),
      ondragstart: (dragEvent) => handleCardDragStart(dragEvent, card.id),
      ondragend: cleanupDragState,
      ...createFileDropProperties(() => card.id),
    },
    coverAttachment
      ? createAttachmentImageElement(coverAttachment, { class: "card-cover", alt: "" }, true)
      : null,
    createElement(
      "div",
      { class: "card-body" },
      createInlineElement("div", { class: "card-title" }, card.title),
      createCardDescriptionElement(card),
      createCardTagsElement(card),
      badgeElements.length ? createElement("div", { class: "card-badges" }, badgeElements) : null
    )
  );
  if (card.color) {
    cardArticleElement.style.borderColor = card.color;
  }
  return cardArticleElement;
}
