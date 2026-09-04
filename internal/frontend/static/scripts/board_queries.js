import { state, selectedCardIdentifiers } from "./state.js";

let indexedBoard = null;
let cardsByIdentifier = new Map();

function rebuildCardIndex() {
  indexedBoard = state.currentBoard;
  cardsByIdentifier = new Map();
  if (!indexedBoard) {
    return;
  }
  for (const currentList of indexedBoard.lists) {
    for (const currentCard of currentList.cards) {
      if (!cardsByIdentifier.has(currentCard.id)) {
        cardsByIdentifier.set(currentCard.id, currentCard);
      }
    }
  }
}

export function currentBoardLists() {
  return state.currentBoard ? state.currentBoard.lists : [];
}

function* iterateBoardCards() {
  for (const currentList of currentBoardLists()) {
    for (const currentCard of currentList.cards) {
      yield currentCard;
    }
  }
}

export function findCardByIdentifier(cardIdentifier) {
  if (!state.currentBoard) {
    return null;
  }
  if (indexedBoard !== state.currentBoard) {
    rebuildCardIndex();
  }
  return cardsByIdentifier.get(cardIdentifier) || null;
}

export function findListByIdentifier(listIdentifier) {
  return currentBoardLists().find((currentList) => currentList.id === listIdentifier) || null;
}

export function collectSelectedCards() {
  const selectedCards = [];
  for (const currentCard of iterateBoardCards()) {
    if (selectedCardIdentifiers.has(currentCard.id)) {
      selectedCards.push(currentCard);
    }
  }
  return selectedCards;
}

export function collectSelectedCardIdentifiersInBoardOrder() {
  return collectSelectedCards().map((currentCard) => currentCard.id);
}

function compareTagsAlphabetically(firstTag, secondTag) {
  return firstTag.localeCompare(secondTag, undefined, { sensitivity: "base" });
}

export function collectAllBoardTags() {
  const seenTagMap = new Map();
  for (const currentCard of iterateBoardCards()) {
    for (const tagText of currentCard.tags || []) {
      const lowerCaseTag = tagText.toLowerCase();
      if (!seenTagMap.has(lowerCaseTag)) {
        seenTagMap.set(lowerCaseTag, tagText);
      }
    }
  }
  return [...seenTagMap.values()].sort(compareTagsAlphabetically);
}

export function countBoardCards() {
  let total = 0;
  for (const currentList of currentBoardLists()) {
    total += Array.isArray(currentList.cards) ? currentList.cards.length : 0;
  }
  return total;
}
