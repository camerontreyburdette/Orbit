import { createElement, querySelector } from "./dom.js";
import { createIconElement } from "./icons.js";
import { state, renderFlags, selectedCardIdentifiers, selectedListIdentifiers, clearEnteringRenderFlags } from "./state.js";
import { hasRedoSnapshots, hasUndoSnapshots } from "./history.js";
import { syncDiscordPresence, updateWindowTitleText } from "./presence.js";
import { renderTopbarHeader } from "./topbar.js";
import { applyCardSearchHighlights } from "./card_search.js";
import { renderHomeScreen } from "./home.js";
import { createAddListComposerElement, createCardItemElement, createListColumnElement } from "./board_view.js";
import { findListByIdentifier } from "./board_queries.js";
import { handleListsDragOver, handleListsDrop } from "./drag_and_drop.js";
import { createMassEditListPanelElement, createMassEditPanelElement } from "./mass_edit.js";
import { redoBoardAction, undoBoardAction } from "./board_actions.js";
import { hasAnySelection, clearSelectionAndRender } from "./selection.js";

let previousViewName = null;

const SELECTION_PRESERVING_SELECTORS = [".card", ".list-header", ".mass-edit-bar", ".history-buttons", ".composer"];

function currentViewName() {
  return state.boardIdentifier == null || !state.currentBoard ? "home" : "board:" + state.boardIdentifier;
}

function captureScrollPositions(rootElement) {
  const previousListsElement = rootElement.querySelector("#lists");
  const listCardScrollOffsets = {};
  rootElement.querySelectorAll(".cards").forEach((cardsElement) => {
    listCardScrollOffsets[cardsElement.dataset.listIdentifier] = cardsElement.scrollTop;
  });
  return {
    listsScrollLeft: previousListsElement ? previousListsElement.scrollLeft : 0,
    listCardScrollOffsets,
  };
}

function restoreScrollPositions(listsContainerElement, scrollPositions) {
  listsContainerElement.scrollLeft = scrollPositions.listsScrollLeft;
  listsContainerElement.querySelectorAll(".cards").forEach((cardsElement) => {
    const savedOffset = scrollPositions.listCardScrollOffsets[cardsElement.dataset.listIdentifier];
    if (savedOffset != null) {
      cardsElement.scrollTop = savedOffset;
    }
  });
}

function isClickInsideSelectionPreservingArea(targetElement) {
  return SELECTION_PRESERVING_SELECTORS.some((selector) => targetElement.closest(selector));
}

function handleBoardBackgroundClick(clickEvent) {
  if (isClickInsideSelectionPreservingArea(clickEvent.target)) {
    return;
  }
  if (hasAnySelection()) {
    clearSelectionAndRender();
  }
}

function createHistoryButtons() {
  return createElement(
    "div",
    { class: "history-buttons" },
    createElement(
      "button",
      { class: "icon-button", dataset: { tooltip: "Undo" }, disabled: !hasUndoSnapshots(), onclick: undoBoardAction },
      createIconElement("arrowleft", 18)
    ),
    createElement(
      "button",
      { class: "icon-button", dataset: { tooltip: "Redo" }, disabled: !hasRedoSnapshots(), onclick: redoBoardAction },
      createIconElement("arrowright", 18)
    )
  );
}

function createActiveMassEditPanel() {
  if (selectedCardIdentifiers.size >= 1) {
    return createMassEditPanelElement();
  }
  if (selectedListIdentifiers.size >= 1) {
    return createMassEditListPanelElement();
  }
  return null;
}

function createListsContainerElement() {
  const listsContainerElement = createElement("div", {
    id: "lists",
    class: renderFlags.isViewEntering ? "enter" : null,
    ondragover: handleListsDragOver,
    ondrop: handleListsDrop,
  });
  for (const currentList of state.currentBoard.lists) {
    listsContainerElement.append(createListColumnElement(currentList));
  }
  listsContainerElement.append(createAddListComposerElement());
  return listsContainerElement;
}

function focusActiveComposer(rootElement) {
  if (!state.composerState) {
    return;
  }
  const targetComposerElement = rootElement.querySelector(".composer textarea, .composer input");
  if (targetComposerElement) {
    targetComposerElement.focus();
    targetComposerElement.scrollIntoView({ block: "nearest" });
  }
}

function renderBoardScreen(rootElement, scrollPositions) {
  const listsContainerElement = createListsContainerElement();
  rootElement.append(listsContainerElement);
  rootElement.onclick = handleBoardBackgroundClick;
  rootElement.append(
    createElement("div", { class: "bottom-actions-container" }, createActiveMassEditPanel(), createHistoryButtons())
  );

  restoreScrollPositions(listsContainerElement, scrollPositions);
  focusActiveComposer(rootElement);
  applyCardSearchHighlights();
}

export function renderCardOnBoard(card) {
  const existingCardElement = document.querySelector(`#lists .card[data-card-identifier="${card.id}"]`);
  const parentList = findListByIdentifier(card.list_id);
  if (!existingCardElement || !parentList) {
    renderMainApplicationView();
    return;
  }
  existingCardElement.replaceWith(createCardItemElement(card, parentList));
  if (state.cardSearchQuery.trim()) {
    applyCardSearchHighlights();
  }
}

export function renderMainApplicationView() {
  renderTopbarHeader();
  updateWindowTitleText();
  syncDiscordPresence();

  const rootElement = querySelector("#board-root");
  const viewName = currentViewName();
  renderFlags.isViewEntering = viewName !== previousViewName;
  previousViewName = viewName;
  const scrollPositions = captureScrollPositions(rootElement);

  rootElement.innerHTML = "";
  if (viewName === "home") {
    renderHomeScreen(rootElement);
  } else {
    renderBoardScreen(rootElement, scrollPositions);
  }
  clearEnteringRenderFlags();
}
