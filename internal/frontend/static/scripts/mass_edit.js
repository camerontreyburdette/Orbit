import { invokeBackendMethod } from "./backend.js";
import { createElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { selectedCardIdentifiers, selectedListIdentifiers } from "./state.js";
import { collectSelectedCards, findCardByIdentifier, findListByIdentifier } from "./board_queries.js";
import { pushCurrentSnapshotToUndoHistory, saveLocalBoardSnapshotState } from "./history.js";
import { refreshCurrentBoard } from "./board_actions.js";
import { renderMainApplicationView } from "./render.js";
import { clearMultiSelectionState } from "./selection.js";
import { confirmDeletion } from "./overlays.js";
import { createColorSwatchRow } from "./color_swatches.js";
import { openMassEditTagsModalDialog } from "./mass_edit_tags.js";

function determineSharedColor(selectedCards) {
  let sharedColor = null;
  let isFirstCard = true;
  for (const currentCard of selectedCards) {
    const currentCardColor = currentCard.color || "";
    if (isFirstCard) {
      sharedColor = currentCardColor;
      isFirstCard = false;
    } else if (sharedColor !== currentCardColor) {
      return null;
    }
  }
  return sharedColor;
}

function applyColorToSelectedCards(chosenColorHex) {
  pushCurrentSnapshotToUndoHistory();
  invokeBackendMethod("batch_update_cards", [...selectedCardIdentifiers], { color: chosenColorHex })
    .then(() => {
      for (const currentCard of collectSelectedCards()) {
        currentCard.color = chosenColorHex;
      }
      saveLocalBoardSnapshotState();
      renderMainApplicationView();
    })
    .catch(refreshCurrentBoard);
}

async function duplicateSelection(identifierSet, duplicateMethodName) {
  pushCurrentSnapshotToUndoHistory();
  await invokeBackendMethod(duplicateMethodName, [...identifierSet]);
  clearMultiSelectionState();
  await refreshCurrentBoard();
}

async function deleteSelection(identifierSet, deleteMethodName, describeSelection) {
  const isConfirmed = await confirmDeletion(describeSelection());
  if (!isConfirmed) {
    return;
  }
  pushCurrentSnapshotToUndoHistory();
  const identifierList = [...identifierSet];
  clearMultiSelectionState();
  await invokeBackendMethod(deleteMethodName, identifierList);
  await refreshCurrentBoard();
}

function describeSelectedCards() {
  const cardCount = selectedCardIdentifiers.size;
  if (cardCount !== 1) {
    return `${cardCount} selected cards`;
  }
  const [singleCardIdentifier] = selectedCardIdentifiers;
  const targetCard = findCardByIdentifier(singleCardIdentifier);
  return targetCard ? targetCard.title : "card";
}

function describeSelectedLists() {
  const listCount = selectedListIdentifiers.size;
  if (listCount !== 1) {
    return `${listCount} selected lists`;
  }
  const [singleListIdentifier] = selectedListIdentifiers;
  const targetList = findListByIdentifier(singleListIdentifier);
  return targetList ? targetList.name : "list";
}

function createSelectionActionButtons(onDuplicate, onDelete) {
  return [
    createElement("button", { class: "icon-button", dataset: { tooltip: "Duplicate" }, onclick: onDuplicate }, createIconElement("copy", 18)),
    createElement("button", { class: "icon-button danger", dataset: { tooltip: "Delete" }, onclick: onDelete }, createIconElement("trash", 18)),
  ];
}

export function createMassEditPanelElement() {
  const sharedColor = determineSharedColor(collectSelectedCards());
  const tagIconButton = createElement(
    "button",
    {
      class: "icon-button",
      dataset: { tooltip: "Edit tags" },
      onclick: (clickEvent) => {
        clickEvent.stopPropagation();
        openMassEditTagsModalDialog();
      },
    },
    createIconElement("tag", 18)
  );

  return createElement(
    "div",
    { class: "mass-edit-bar" },
    createColorSwatchRow("mass-edit-colors", sharedColor, applyColorToSelectedCards),
    tagIconButton,
    createSelectionActionButtons(
      () => duplicateSelection(selectedCardIdentifiers, "duplicate_cards"),
      () => deleteSelection(selectedCardIdentifiers, "batch_delete_cards", describeSelectedCards)
    )
  );
}

export function createMassEditListPanelElement() {
  return createElement(
    "div",
    { class: "mass-edit-bar" },
    createSelectionActionButtons(
      () => duplicateSelection(selectedListIdentifiers, "duplicate_lists"),
      () => deleteSelection(selectedListIdentifiers, "batch_delete_lists", describeSelectedLists)
    )
  );
}
