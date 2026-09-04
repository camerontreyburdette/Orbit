import { createElement } from "./dom.js";
import { persistLocalChange } from "./board_actions.js";
import { renderCardModalDialog, rerenderCardViews } from "./card_modal.js";
import { countItemsBeforePlaceholder, findVerticalSiblingAfter, placePlaceholderElement } from "./drag_and_drop.js";

let checklistDragIdentifier = null;
let checklistItemDragState = null;
const checklistDragPlaceholderElement = createElement("div", { class: "checklist-placeholder" });
const checklistItemDragPlaceholderElement = createElement("div", { class: "checklist-item-placeholder" });

function isReorderableChecklistBlock(element) {
  return (
    element.classList.contains("checklist-block") &&
    !element.classList.contains("drag-src") &&
    !element.classList.contains("checklist-naming")
  );
}

function isReorderableChecklistItem(element) {
  return element.classList.contains("checklist-item") && !element.classList.contains("dragging");
}

function moveWithinArray(items, fromIndex, toIndex) {
  const [movedItem] = items.splice(fromIndex, 1);
  items.splice(Math.min(toIndex, items.length), 0, movedItem);
}

export function handleChecklistDragStart(dragEvent, checklistIdentifier) {
  checklistDragIdentifier = checklistIdentifier;
  dragEvent.dataTransfer.effectAllowed = "move";
  dragEvent.dataTransfer.setData("text/plain", "checklist:" + checklistIdentifier);
  const blockElement = dragEvent.currentTarget.closest(".checklist-block");
  requestAnimationFrame(() => blockElement.classList.add("drag-src"));
  dragEvent.stopPropagation();
}

export function cleanupChecklistDragState() {
  checklistDragIdentifier = null;
  checklistDragPlaceholderElement.remove();
  document.querySelectorAll(".checklist-block.drag-src").forEach((element) => {
    element.classList.remove("drag-src");
  });
}

export function handleChecklistWrapDragOver(dragEvent) {
  if (checklistDragIdentifier == null) {
    return;
  }
  dragEvent.preventDefault();
  dragEvent.stopPropagation();
  dragEvent.dataTransfer.dropEffect = "move";
  const wrapContainer = dragEvent.currentTarget;
  const siblingAfter = findVerticalSiblingAfter(
    wrapContainer,
    ":scope > .checklist-block:not(.drag-src):not(.checklist-naming)",
    dragEvent.clientY
  );
  placePlaceholderElement(checklistDragPlaceholderElement, wrapContainer, siblingAfter || null);
}

export function handleChecklistWrapDrop(dropEvent, card) {
  if (checklistDragIdentifier == null) {
    return;
  }
  dropEvent.preventDefault();
  dropEvent.stopPropagation();
  const dropIndex = countItemsBeforePlaceholder(dropEvent.currentTarget, checklistDragPlaceholderElement, isReorderableChecklistBlock);
  const draggedIdentifier = checklistDragIdentifier;
  cleanupChecklistDragState();

  const checklists = card.checklists || [];
  const fromIndex = checklists.findIndex((item) => item.id === draggedIdentifier);
  if (fromIndex === -1) {
    return;
  }
  moveWithinArray(checklists, fromIndex, dropIndex);
  persistLocalChange("move_checklist", draggedIdentifier, dropIndex);
  renderCardModalDialog();
}

export function handleChecklistItemDragStart(dragEvent, checklistIdentifier, itemIdentifier) {
  checklistItemDragState = { checklistIdentifier, itemIdentifier };
  dragEvent.dataTransfer.effectAllowed = "move";
  dragEvent.dataTransfer.setData("text/plain", "checklist-item:" + itemIdentifier);
  const rowElement = dragEvent.currentTarget;
  requestAnimationFrame(() => rowElement.classList.add("dragging"));
  dragEvent.stopPropagation();
}

export function cleanupChecklistItemDragState() {
  checklistItemDragState = null;
  checklistItemDragPlaceholderElement.remove();
  document.querySelectorAll(".checklist-item.dragging").forEach((element) => {
    element.classList.remove("dragging");
  });
}

export function handleChecklistBlockDragOver(dragEvent) {
  if (!checklistItemDragState) {
    return;
  }
  dragEvent.preventDefault();
  dragEvent.stopPropagation();
  dragEvent.dataTransfer.dropEffect = "move";
  const itemsContainer = dragEvent.currentTarget.querySelector(":scope > .checklist-items");
  if (!itemsContainer) {
    return;
  }
  const siblingAfter = findVerticalSiblingAfter(itemsContainer, ":scope > .checklist-item:not(.dragging)", dragEvent.clientY);
  placePlaceholderElement(checklistItemDragPlaceholderElement, itemsContainer, siblingAfter || null);
}

export function handleChecklistBlockDrop(dropEvent, card, targetChecklist) {
  if (!checklistItemDragState) {
    return;
  }
  dropEvent.preventDefault();
  dropEvent.stopPropagation();
  const itemsContainer = dropEvent.currentTarget.querySelector(":scope > .checklist-items");
  const dropIndex = itemsContainer
    ? countItemsBeforePlaceholder(itemsContainer, checklistItemDragPlaceholderElement, isReorderableChecklistItem)
    : 0;
  const { checklistIdentifier, itemIdentifier } = checklistItemDragState;
  cleanupChecklistItemDragState();

  const sourceChecklist = (card.checklists || []).find((item) => item.id === checklistIdentifier);
  if (!sourceChecklist) {
    return;
  }
  const fromIndex = sourceChecklist.items.findIndex((item) => item.id === itemIdentifier);
  if (fromIndex === -1) {
    return;
  }
  const [movedItem] = sourceChecklist.items.splice(fromIndex, 1);
  targetChecklist.items.splice(Math.min(dropIndex, targetChecklist.items.length), 0, movedItem);
  persistLocalChange("move_checklist_item", itemIdentifier, targetChecklist.id, dropIndex);
  rerenderCardViews(card);
}
