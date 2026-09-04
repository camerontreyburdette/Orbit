import { createElement, stopPropagationThen } from "./dom.js";
import { createIconElement } from "./icons.js";
import { renderBlockMarkup, renderInlineMarkup } from "./markup.js";
import { formatDateString, formatTimeSpent } from "./formatting.js";
import { state, renderFlags } from "./state.js";
import { focusInputElementAtEnd } from "./inline_editing.js";
import { openModalOverlay } from "./overlays.js";
import { createNewBoard, deleteExistingBoard, switchActiveBoard, updateBoardFields } from "./board_actions.js";
import { openSettingsModalDialog } from "./settings.js";
import { openBoardImportModalDialog } from "./board_import_dialog.js";

let isSettingsCornerEntering = true;

function submitOnEnter(submitCallback, allowsShiftNewline) {
  return (keyboardEvent) => {
    if (keyboardEvent.key === "Enter" && (!allowsShiftNewline || !keyboardEvent.shiftKey)) {
      keyboardEvent.preventDefault();
      submitCallback();
    }
  };
}

function createBoardTileButtons(boardSummary) {
  return createElement(
    "div",
    { class: "tile-buttons" },
    createElement(
      "button",
      {
        class: "icon-button" + (boardSummary.pinned ? " active" : ""),
        dataset: { tooltip: boardSummary.pinned ? "Unpin board" : "Pin board" },
        onclick: stopPropagationThen(() => updateBoardFields(boardSummary.id, { pinned: boardSummary.pinned ? 0 : 1 })),
      },
      createIconElement("pin", 18, Boolean(boardSummary.pinned))
    ),
    createElement(
      "button",
      {
        class: "icon-button",
        dataset: { tooltip: "Edit board" },
        onclick: stopPropagationThen(() => openEditBoardModalDialog(boardSummary)),
      },
      createIconElement("pencil", 18)
    ),
    createElement(
      "button",
      {
        class: "icon-button danger",
        dataset: { tooltip: "Delete board" },
        onclick: stopPropagationThen(() => deleteExistingBoard(boardSummary)),
      },
      createIconElement("trash", 18)
    )
  );
}

function createBoardTileElement(boardSummary) {
  const nameElement = createElement("div", { class: "tile-name" });
  nameElement.innerHTML = renderInlineMarkup(boardSummary.name);
  let descriptionElement = null;
  if (boardSummary.description) {
    descriptionElement = createElement("div", { class: "tile-description" });
    descriptionElement.innerHTML = renderBlockMarkup(boardSummary.description);
  }
  return createElement(
    "div",
    {
      class: "board-tile" + (boardSummary.pinned ? " pinned" : ""),
      onclick: () => switchActiveBoard(boardSummary.id),
    },
    nameElement,
    descriptionElement,
    createElement(
      "div",
      { class: "tile-footer" },
      createBoardTileButtons(boardSummary),
      createElement(
        "span",
        { class: "tile-meta" },
        createElement("span", { class: "tile-time", dataset: { tooltip: "Time spent in this board" } }, formatTimeSpent(boardSummary.time_spent_seconds)),
        createElement("span", { class: "tile-separator" }, "·"),
        createElement("span", { class: "tile-date" }, formatDateString(boardSummary.created_at))
      )
    )
  );
}

function matchesBoardSearch(boardSummary, query) {
  return (
    !query ||
    boardSummary.name.toLowerCase().includes(query) ||
    (boardSummary.description || "").toLowerCase().includes(query)
  );
}

function populateBoardGrid(boardGroupsContainer) {
  boardGroupsContainer.innerHTML = "";
  const query = state.searchQuery.trim().toLowerCase();
  const filteredBoards = state.boards.filter((boardSummary) => matchesBoardSearch(boardSummary, query));
  const pinnedBoards = filteredBoards.filter((boardSummary) => boardSummary.pinned);
  const regularBoards = filteredBoards.filter((boardSummary) => !boardSummary.pinned);

  if (pinnedBoards.length) {
    boardGroupsContainer.append(createElement("div", { class: "board-grid" }, pinnedBoards.map(createBoardTileElement)));
  }
  if (pinnedBoards.length && regularBoards.length) {
    boardGroupsContainer.append(createElement("div", { class: "board-separator" }));
  }
  if (regularBoards.length) {
    boardGroupsContainer.append(createElement("div", { class: "board-grid" }, regularBoards.map(createBoardTileElement)));
  }
}

function createHomeControls(boardGroupsContainer) {
  return createElement(
    "div",
    { class: "home-controls" },
    createElement(
      "button",
      {
        class: "button import-board-button",
        "aria-label": "Import boards",
        dataset: { tooltip: "Import boards" },
        onclick: openBoardImportModalDialog,
      },
      createIconElement("upload", 18)
    ),
    createElement(
      "div",
      { class: "search-box" },
      createIconElement("search", 16),
      createElement("input", {
        class: "search-input",
        value: state.searchQuery,
        oninput: (inputEvent) => {
          state.searchQuery = inputEvent.target.value;
          populateBoardGrid(boardGroupsContainer);
        },
      })
    ),
    createElement(
      "button",
      { class: "button button-primary add-board-button", dataset: { tooltip: "New board" }, onclick: openCreateBoardModalDialog },
      createIconElement("plus", 18)
    )
  );
}

function createSettingsCornerContainer() {
  const container = createElement(
    "div",
    { class: "settings-corner" + (isSettingsCornerEntering ? " enter" : "") },
    createElement(
      "button",
      {
        class: "icon-button",
        "aria-label": "Settings",
        dataset: { tooltip: "Settings" },
        onclick: openSettingsModalDialog,
      },
      createIconElement("gear", 18)
    )
  );
  isSettingsCornerEntering = false;
  return container;
}

export function renderHomeScreen(rootElement) {
  const boardGroupsContainer = createElement("div", { class: "board-groups" });
  populateBoardGrid(boardGroupsContainer);

  rootElement.append(
    createElement(
      "div",
      { class: "home" + (renderFlags.isViewEntering ? " enter" : "") },
      createHomeControls(boardGroupsContainer),
      boardGroupsContainer
    )
  );
  rootElement.append(createSettingsCornerContainer());
}

function openBoardFormModalDialog({ name, description, submitLabel, onSubmit }) {
  const nameInputElement = createElement("input", {
    placeholder: "Name",
    value: name,
    onkeydown: submitOnEnter(() => submitForm(), false),
  });
  const descriptionInputElement = createElement("textarea", {
    rows: 3,
    placeholder: "Description",
    value: description,
    onkeydown: submitOnEnter(() => submitForm(), true),
  });

  function submitForm() {
    const trimmedName = nameInputElement.value.trim();
    if (!trimmedName) {
      nameInputElement.focus();
      return;
    }
    const descriptionValue = descriptionInputElement.value;
    closeModal();
    onSubmit(trimmedName, descriptionValue);
  }

  const closeModal = openModalOverlay(
    createElement(
      "div",
      { class: "confirm-box dialog-box" },
      nameInputElement,
      descriptionInputElement,
      createElement(
        "div",
        { class: "confirm-actions" },
        createElement("button", { class: "button", onclick: () => closeModal() }, "Cancel"),
        createElement("button", { class: "button button-primary", onclick: submitForm }, submitLabel)
      )
    )
  );
  focusInputElementAtEnd(nameInputElement);
}

function openCreateBoardModalDialog() {
  openBoardFormModalDialog({
    name: "",
    description: "",
    submitLabel: "Create",
    onSubmit: (name, description) => createNewBoard(name, description),
  });
}

function openEditBoardModalDialog(boardSummary) {
  openBoardFormModalDialog({
    name: boardSummary.name,
    description: boardSummary.description || "",
    submitLabel: "Save",
    onSubmit: (name, description) => updateBoardFields(boardSummary.id, { name, description }),
  });
}
