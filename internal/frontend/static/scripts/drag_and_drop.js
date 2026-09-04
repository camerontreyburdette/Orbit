import { createElement } from "./dom.js";
import { renderFlags, selectedCardIdentifiers, selectedListIdentifiers } from "./state.js";
import { currentBoardLists, collectSelectedCardIdentifiersInBoardOrder } from "./board_queries.js";
import { performBackendMutationThenRefresh } from "./board_actions.js";
import { applyStackedDragImage } from "./drag_ghost.js";

let activeDragState = null;

const cardDragPlaceholderElement = createElement("div", { class: "card-placeholder" });
const listDragPlaceholderElement = createElement("div", { class: "list-placeholder" });

function selectedListIdentifiersInBoardOrder() {
  return currentBoardLists()
    .filter((currentList) => selectedListIdentifiers.has(currentList.id))
    .map((currentList) => currentList.id);
}

function markDraggedElements(selectorTemplate, identifiers, className) {
  requestAnimationFrame(() => {
    for (const identifier of identifiers) {
      const element = document.querySelector(selectorTemplate(identifier));
      if (element) {
        element.classList.add(className);
      }
    }
  });
}

function beginDrag(dragEvent, type, identifier, identifiers) {
  activeDragState = { type, identifier, identifiers };
  dragEvent.dataTransfer.effectAllowed = "move";
  dragEvent.dataTransfer.setData("text/plain", type + ":" + identifier);
}

export function handleCardDragStart(dragEvent, cardIdentifier) {
  let cardIdentifiersToDrag = [cardIdentifier];
  if (selectedCardIdentifiers.has(cardIdentifier)) {
    cardIdentifiersToDrag = collectSelectedCardIdentifiersInBoardOrder();
  } else {
    selectedCardIdentifiers.clear();
  }

  beginDrag(dragEvent, "card", cardIdentifier, cardIdentifiersToDrag);
  applyStackedDragImage(dragEvent, dragEvent.currentTarget, cardIdentifiersToDrag.length);
  markDraggedElements((identifier) => `.card[data-card-identifier="${identifier}"]`, cardIdentifiersToDrag, "dragging");
  dragEvent.stopPropagation();
}

export function handleListDragStart(dragEvent, listIdentifier) {
  let listIdentifiersToDrag = [listIdentifier];
  if (selectedListIdentifiers.has(listIdentifier)) {
    listIdentifiersToDrag = selectedListIdentifiersInBoardOrder();
  } else {
    selectedListIdentifiers.clear();
  }

  beginDrag(dragEvent, "list", listIdentifier, listIdentifiersToDrag);
  applyStackedDragImage(dragEvent, dragEvent.currentTarget, listIdentifiersToDrag.length);
  markDraggedElements((identifier) => `.list[data-list-identifier="${identifier}"]`, listIdentifiersToDrag, "drag-src");
}

export function cleanupDragState() {
  activeDragState = null;
  cardDragPlaceholderElement.remove();
  listDragPlaceholderElement.remove();
  document.querySelectorAll(".dragging, .drag-src").forEach((element) => {
    element.classList.remove("dragging", "drag-src");
  });
}

export function placePlaceholderElement(placeholderElement, containerElement, beforeElement) {
  if (placeholderElement.parentNode === containerElement && placeholderElement.nextSibling === (beforeElement || null)) {
    return;
  }
  containerElement.insertBefore(placeholderElement, beforeElement || null);
}

export function findVerticalSiblingAfter(containerElement, selector, clientY) {
  return [...containerElement.querySelectorAll(selector)].find(
    (element) => clientY < element.getBoundingClientRect().top + element.offsetHeight / 2
  );
}

function findHorizontalSiblingAfter(containerElement, selector, clientX) {
  return [...containerElement.querySelectorAll(selector)].find(
    (element) => clientX < element.getBoundingClientRect().left + element.offsetWidth / 2
  );
}

export function countItemsBeforePlaceholder(containerElement, placeholderElement, isCountedItem) {
  let dropIndex = 0;
  for (const childElement of containerElement.children) {
    if (childElement === placeholderElement) {
      break;
    }
    if (isCountedItem(childElement)) {
      dropIndex++;
    }
  }
  return dropIndex;
}

function isActiveDragOfType(type) {
  return Boolean(activeDragState) && activeDragState.type === type;
}

function takeDraggedIdentifiers() {
  const identifiers =
    activeDragState.identifiers && activeDragState.identifiers.length
      ? activeDragState.identifiers
      : [activeDragState.identifier];
  cleanupDragState();
  return identifiers;
}

export function handleCardsDragOver(dragEvent, containerElement) {
  if (!isActiveDragOfType("card")) {
    return;
  }
  dragEvent.preventDefault();
  dragEvent.dataTransfer.dropEffect = "move";
  const siblingAfter = findVerticalSiblingAfter(containerElement, ":scope > .card:not(.dragging)", dragEvent.clientY);
  placePlaceholderElement(cardDragPlaceholderElement, containerElement, siblingAfter || null);
}

export function handleCardsDrop(dragEvent, containerElement, targetListIdentifier) {
  if (!isActiveDragOfType("card")) {
    return;
  }
  dragEvent.preventDefault();
  dragEvent.stopPropagation();
  const dropIndex = countItemsBeforePlaceholder(
    containerElement,
    cardDragPlaceholderElement,
    (childElement) => childElement.classList.contains("card") && !childElement.classList.contains("dragging")
  );

  const cardIdentifiers = takeDraggedIdentifiers();
  renderFlags.enteringCardIdentifier = cardIdentifiers[0];

  if (cardIdentifiers.length === 1) {
    performBackendMutationThenRefresh("move_card", cardIdentifiers[0], targetListIdentifier, dropIndex);
  } else {
    performBackendMutationThenRefresh("move_cards", cardIdentifiers, targetListIdentifier, dropIndex);
  }
}

export function handleListsDragOver(dragEvent) {
  if (!isActiveDragOfType("list")) {
    return;
  }
  dragEvent.preventDefault();
  dragEvent.dataTransfer.dropEffect = "move";
  const containerElement = dragEvent.currentTarget;
  const siblingAfter = findHorizontalSiblingAfter(containerElement, ":scope > .list:not(.drag-src)", dragEvent.clientX);
  placePlaceholderElement(
    listDragPlaceholderElement,
    containerElement,
    siblingAfter || containerElement.querySelector(":scope > .add-list")
  );
}

export function handleListsDrop(dragEvent) {
  if (!isActiveDragOfType("list")) {
    return;
  }
  dragEvent.preventDefault();
  const containerElement = dragEvent.currentTarget;
  const dropIndex = countItemsBeforePlaceholder(
    containerElement,
    listDragPlaceholderElement,
    (childElement) => childElement.classList.contains("list") && !childElement.classList.contains("drag-src")
  );

  const listIdentifiers = takeDraggedIdentifiers();
  renderFlags.enteringListIdentifier = listIdentifiers[0];

  if (listIdentifiers.length === 1) {
    performBackendMutationThenRefresh("move_list", listIdentifiers[0], dropIndex);
  } else {
    performBackendMutationThenRefresh("move_lists", listIdentifiers, dropIndex);
  }
}
