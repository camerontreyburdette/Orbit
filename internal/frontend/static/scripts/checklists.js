import { invokeBackendMethod } from "./backend.js";
import { createElement, createInlineElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { cardModalEditing } from "./state.js";
import { saveLocalBoardSnapshotState } from "./history.js";
import { createInlineEditInput, scheduleFocus, scheduleFocusAtEnd } from "./inline_editing.js";
import { confirmDeletion } from "./overlays.js";
import { persistLocalChange } from "./board_actions.js";
import { renderCardOnBoard } from "./render.js";
import { renderCardModalDialog, rerenderCardViews } from "./card_modal.js";
import {
  cleanupChecklistDragState,
  cleanupChecklistItemDragState,
  handleChecklistBlockDragOver,
  handleChecklistBlockDrop,
  handleChecklistDragStart,
  handleChecklistItemDragStart,
  handleChecklistWrapDragOver,
  handleChecklistWrapDrop,
} from "./checklist_drag_and_drop.js";

function calculateChecklistProgress(checklist) {
  const totalItemsCount = checklist.items.length;
  const completedItemsCount = checklist.items.filter((item) => item.done).length;
  const percentage = totalItemsCount ? Math.round((completedItemsCount / totalItemsCount) * 100) : 0;
  return { totalItemsCount, completedItemsCount, percentage };
}

function createProgressBar(percentage) {
  return createElement(
    "div",
    { class: "checklist-bar" },
    createElement("div", { class: "checklist-bar-fill", style: { width: percentage + "%" } })
  );
}

function createDeleteIconButton(label, onClick, extraProperties = {}) {
  return createElement(
    "button",
    { class: "icon-button danger", dataset: { tooltip: label }, onclick: onClick, ...extraProperties },
    createIconElement("trash", 14)
  );
}

export function createChecklistSection(card) {
  const checklists = card.checklists || [];
  const sectionElement = createElement("div", { class: "modal-section checklist-section" }, createElement("h3", {}, "Checklists"));
  if (!cardModalEditing.isChecklistCreating) {
    sectionElement.append(
      createElement(
        "div",
        {
          class: "description-preview empty checklist-add-checklist",
          onclick: () => {
            cardModalEditing.isChecklistCreating = true;
            renderCardModalDialog();
          },
        },
        "Add a checklist…"
      )
    );
  }
  if (checklists.length || cardModalEditing.isChecklistCreating) {
    sectionElement.append(
      createElement(
        "div",
        {
          class: "checklist-wrap",
          ondragover: handleChecklistWrapDragOver,
          ondrop: (dropEvent) => handleChecklistWrapDrop(dropEvent, card),
        },
        cardModalEditing.isChecklistCreating ? createChecklistNamingBlock(card) : null,
        checklists.map((checklist) => createChecklistBlockElement(card, checklist))
      )
    );
  }
  return sectionElement;
}

function createChecklistNamingBlock(card) {
  let blockElement;
  const cancelCreation = () => {
    cardModalEditing.isChecklistCreating = false;
    renderCardModalDialog();
  };
  const titleInputElement = createInlineEditInput({
    className: "checklist-title-input",
    placeholder: "Checklist name…",
    onCancel: (keyboardEvent) => {
      keyboardEvent.target.value = "";
      cancelCreation();
    },
    onCommit: (rawValue, blurEvent) => {
      const trimmedValue = rawValue.trim();
      const isSwitchingToItems = Boolean(
        blurEvent.relatedTarget &&
          blockElement &&
          blockElement.contains(blurEvent.relatedTarget) &&
          blurEvent.relatedTarget.classList.contains("checklist-add")
      );
      cardModalEditing.isChecklistCreating = false;
      if (!trimmedValue) {
        renderCardModalDialog();
        return;
      }
      addChecklistToCard(card, trimmedValue, isSwitchingToItems);
    },
  });
  scheduleFocus(titleInputElement);
  blockElement = createElement(
    "div",
    { class: "checklist-block checklist-naming" },
    createElement(
      "div",
      { class: "checklist-header" },
      titleInputElement,
      createDeleteIconButton(
        "Cancel",
        () => {
          titleInputElement.value = "";
          cancelCreation();
        },
        { onmousedown: (mouseEvent) => mouseEvent.preventDefault() }
      )
    ),
    createProgressBar(0),
    createElement("input", { class: "checklist-add", placeholder: "Add an item…" })
  );
  return blockElement;
}

async function addChecklistToCard(card, checklistTitle, shouldFocusItems) {
  let creationResponse;
  try {
    creationResponse = await invokeBackendMethod("add_checklist", card.id, checklistTitle);
  } catch {
    renderCardModalDialog();
    return;
  }
  card.checklists = [creationResponse.checklist, ...(card.checklists || [])];
  if (shouldFocusItems) {
    cardModalEditing.refocusChecklistAddIdentifier = creationResponse.checklist.id;
  }
  saveLocalBoardSnapshotState();
  rerenderCardViews(card);
}

function createChecklistTitleElement(checklist) {
  if (cardModalEditing.editingChecklistTitleIdentifier !== checklist.id) {
    return createInlineElement(
      "span",
      {
        class: "checklist-title",
        onclick: () => {
          cardModalEditing.editingChecklistTitleIdentifier = checklist.id;
          renderCardModalDialog();
        },
      },
      checklist.title
    );
  }
  const titleInputElement = createInlineEditInput({
    className: "checklist-title-input",
    value: checklist.title,
    onCancel: () => {
      cardModalEditing.editingChecklistTitleIdentifier = null;
      renderCardModalDialog();
    },
    onCommit: (rawValue) => {
      const trimmedValue = rawValue.trim();
      cardModalEditing.editingChecklistTitleIdentifier = null;
      if (trimmedValue && trimmedValue !== checklist.title) {
        checklist.title = trimmedValue;
        persistLocalChange("rename_checklist", checklist.id, trimmedValue);
      }
      renderCardModalDialog();
    },
  });
  scheduleFocusAtEnd(titleInputElement);
  return titleInputElement;
}

function createAddItemInput(card, checklist) {
  const submitItem = (inputElement) => {
    const trimmedValue = inputElement.value.trim();
    if (!trimmedValue) {
      return;
    }
    inputElement.value = "";
    addChecklistItemToCard(card, checklist, trimmedValue);
  };
  const addItemInputElement = createElement("input", {
    class: "checklist-add",
    placeholder: "Add an item…",
    onkeydown: (keyboardEvent) => {
      if (keyboardEvent.key === "Escape") {
        keyboardEvent.stopPropagation();
        keyboardEvent.target.value = "";
        keyboardEvent.target.blur();
        return;
      }
      if (keyboardEvent.key !== "Enter") {
        return;
      }
      keyboardEvent.preventDefault();
      submitItem(keyboardEvent.target);
    },
    onblur: (blurEvent) => submitItem(blurEvent.target),
  });
  if (cardModalEditing.refocusChecklistAddIdentifier === checklist.id) {
    cardModalEditing.refocusChecklistAddIdentifier = null;
    scheduleFocus(addItemInputElement);
  }
  return addItemInputElement;
}

function createChecklistBlockElement(card, checklist) {
  const { totalItemsCount, completedItemsCount, percentage } = calculateChecklistProgress(checklist);
  const isDragAllowed =
    cardModalEditing.editingChecklistTitleIdentifier !== checklist.id && (card.checklists || []).length > 1;

  return createElement(
    "div",
    {
      class: "checklist-block",
      dataset: { checklistIdentifier: checklist.id },
      ondragover: handleChecklistBlockDragOver,
      ondrop: (dropEvent) => handleChecklistBlockDrop(dropEvent, card, checklist),
    },
    createElement(
      "div",
      {
        class: "checklist-header",
        draggable: isDragAllowed,
        ondragstart: (dragEvent) => handleChecklistDragStart(dragEvent, checklist.id),
        ondragend: cleanupChecklistDragState,
      },
      createChecklistTitleElement(checklist),
      createElement("span", { class: "checklist-count" }, completedItemsCount + "/" + totalItemsCount),
      createDeleteIconButton("Delete checklist", () => deleteChecklistFromCard(card, checklist))
    ),
    createProgressBar(percentage),
    createElement("div", { class: "checklist-items" }, checklist.items.map((item) => createChecklistItemRow(card, checklist, item))),
    createAddItemInput(card, checklist)
  );
}

async function addChecklistItemToCard(card, checklist, itemText) {
  let creationResponse;
  try {
    creationResponse = await invokeBackendMethod("add_checklist_item", checklist.id, itemText);
  } catch {
    return;
  }
  checklist.items.push(creationResponse.item);
  cardModalEditing.refocusChecklistAddIdentifier = checklist.id;
  saveLocalBoardSnapshotState();
  rerenderCardViews(card);
}

async function deleteChecklistFromCard(card, checklist) {
  const isConfirmed = await confirmDeletion(checklist.title);
  if (!isConfirmed) {
    return;
  }
  card.checklists = (card.checklists || []).filter((item) => item !== checklist);
  persistLocalChange("delete_checklist", checklist.id);
  rerenderCardViews(card);
}

function createChecklistItemTextElement(item) {
  if (cardModalEditing.editingChecklistItemIdentifier !== item.id) {
    return createInlineElement(
      "span",
      {
        class: "checklist-item-text",
        onclick: () => {
          cardModalEditing.editingChecklistItemIdentifier = item.id;
          renderCardModalDialog();
        },
      },
      item.text
    );
  }
  const textInputElement = createInlineEditInput({
    className: "checklist-item-input",
    value: item.text,
    onCancel: () => {
      cardModalEditing.editingChecklistItemIdentifier = null;
      renderCardModalDialog();
    },
    onCommit: (rawValue) => {
      const trimmedValue = rawValue.trim();
      cardModalEditing.editingChecklistItemIdentifier = null;
      if (trimmedValue && trimmedValue !== item.text) {
        item.text = trimmedValue;
        persistLocalChange("edit_checklist_item", item.id, trimmedValue);
      }
      renderCardModalDialog();
    },
  });
  scheduleFocusAtEnd(textInputElement);
  return textInputElement;
}

function createChecklistItemRow(card, checklist, item) {
  const rowElement = createElement(
    "div",
    {
      class: "checklist-item" + (item.done ? " done" : ""),
      draggable: cardModalEditing.editingChecklistItemIdentifier !== item.id,
      ondragstart: (dragEvent) => handleChecklistItemDragStart(dragEvent, checklist.id, item.id),
      ondragend: cleanupChecklistItemDragState,
    },
    createElement("input", {
      type: "checkbox",
      class: "checklist-check",
      checked: Boolean(item.done),
      onchange: (changeEvent) => toggleChecklistItemState(card, checklist, item, changeEvent.target.checked, rowElement),
    }),
    createChecklistItemTextElement(item),
    createElement(
      "button",
      {
        class: "icon-button danger checklist-item-delete",
        dataset: { tooltip: "Delete item" },
        onclick: () => {
          checklist.items = checklist.items.filter((existingItem) => existingItem !== item);
          persistLocalChange("delete_checklist_item", item.id);
          rerenderCardViews(card);
        },
      },
      createIconElement("x", 12)
    )
  );
  return rowElement;
}

function updateChecklistProgressDisplay(blockElement, checklist) {
  const { totalItemsCount, completedItemsCount, percentage } = calculateChecklistProgress(checklist);
  const countElement = blockElement.querySelector(".checklist-count");
  if (countElement) {
    countElement.textContent = completedItemsCount + "/" + totalItemsCount;
  }
  const fillElement = blockElement.querySelector(".checklist-bar-fill");
  if (fillElement) {
    fillElement.style.width = percentage + "%";
  }
}

function toggleChecklistItemState(card, checklist, item, isDone, rowElement) {
  item.done = isDone;
  rowElement.classList.toggle("done", isDone);
  const blockElement = rowElement.closest(".checklist-block");
  if (blockElement) {
    updateChecklistProgressDisplay(blockElement, checklist);
  }
  persistLocalChange("toggle_checklist_item", item.id, isDone);
  renderCardOnBoard(card);
}
