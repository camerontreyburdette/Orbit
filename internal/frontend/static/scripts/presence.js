import { invokeBackendMethod, isBackendAvailable } from "./backend.js";
import { stripMarkupFormatting } from "./markup.js";
import { state, isInBoardView } from "./state.js";
import { countBoardCards, findCardByIdentifier } from "./board_queries.js";

const MAXIMUM_PRESENCE_TEXT_LENGTH = 100;

let lastWindowTitle = "";
let lastPresenceJSON = "";

function truncatePresenceText(rawText) {
  return String(stripMarkupFormatting(rawText)).slice(0, MAXIMUM_PRESENCE_TEXT_LENGTH);
}

export function updateWindowTitleText() {
  const currentTitle = state.currentBoard
    ? "Orbit - " + stripMarkupFormatting(state.currentBoard.board.name)
    : "Orbit";
  if (currentTitle === lastWindowTitle) {
    return;
  }
  lastWindowTitle = currentTitle;
  document.title = currentTitle;
  invokeBackendMethod("set_title", currentTitle).catch(() => {});
}

function buildPresenceSnapshot() {
  const inBoardView = isInBoardView();
  const presenceSnapshot = {
    view: inBoardView ? "board" : "home",
    board_name: "",
    card_title: "",
    editing: false,
    lists: 0,
    cards: 0,
    boards: Array.isArray(state.boards) ? state.boards.length : 0,
  };

  if (!inBoardView) {
    return presenceSnapshot;
  }

  const currentBoard = state.currentBoard;
  if (currentBoard.board && currentBoard.board.name) {
    presenceSnapshot.board_name = truncatePresenceText(currentBoard.board.name);
  }
  if (Array.isArray(currentBoard.lists)) {
    presenceSnapshot.lists = currentBoard.lists.length;
    presenceSnapshot.cards = countBoardCards();

    if (state.openCardIdentifier != null) {
      const openCard = findCardByIdentifier(state.openCardIdentifier);
      if (openCard) {
        presenceSnapshot.card_title = truncatePresenceText(openCard.title);
      }
      presenceSnapshot.editing = Boolean(state.isEditingTitle || state.isEditingDescription);
    }
  }
  return presenceSnapshot;
}

export function syncDiscordPresence(overrideCardTitle, overrideEditing) {
  if (!isBackendAvailable()) {
    return;
  }

  const presenceSnapshot = buildPresenceSnapshot();
  if (overrideCardTitle != null) {
    presenceSnapshot.card_title = truncatePresenceText(overrideCardTitle);
  }
  if (overrideEditing != null) {
    presenceSnapshot.editing = Boolean(overrideEditing);
  }

  const jsonString = JSON.stringify(presenceSnapshot);
  if (jsonString === lastPresenceJSON) {
    return;
  }
  lastPresenceJSON = jsonString;

  invokeBackendMethod("set_presence_context", presenceSnapshot).catch(() => {});
}
