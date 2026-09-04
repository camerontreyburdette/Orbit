import { MAXIMUM_HISTORY_SNAPSHOT_COUNT } from "./constants.js";
import { state } from "./state.js";

const undoHistorySnapshots = [];
const redoHistorySnapshots = [];
let lastBoardSnapshotJSON = null;

function cloneCardForSnapshot(currentCard) {
  const cardCopy = {
    ...currentCard,
    tags: [...(currentCard.tags || [])],
    checklists: (currentCard.checklists || []).map((currentChecklist) => ({
      ...currentChecklist,
      items: currentChecklist.items.map((currentItem) => ({ ...currentItem })),
    })),
    attachments: (currentCard.attachments || []).map((currentAttachment) => ({ ...currentAttachment })),
  };
  delete cardCopy.list_id;
  return cardCopy;
}

function createBoardDocumentObject() {
  if (!state.currentBoard) {
    return null;
  }
  return {
    ...state.currentBoard.board,
    lists: state.currentBoard.lists.map((currentList) => ({
      id: currentList.id,
      name: currentList.name,
      cards: currentList.cards.map(cloneCardForSnapshot),
    })),
  };
}

function captureCurrentBoardSnapshotJSON() {
  const documentObject = createBoardDocumentObject();
  return documentObject ? JSON.stringify(documentObject) : null;
}

export function pushCurrentSnapshotToUndoHistory() {
  if (!lastBoardSnapshotJSON) {
    return;
  }
  undoHistorySnapshots.push(lastBoardSnapshotJSON);
  if (undoHistorySnapshots.length > MAXIMUM_HISTORY_SNAPSHOT_COUNT) {
    undoHistorySnapshots.shift();
  }
  redoHistorySnapshots.length = 0;
}

export function recordBaselineSnapshot() {
  lastBoardSnapshotJSON = captureCurrentBoardSnapshotJSON();
}

export function saveLocalBoardSnapshotState() {
  pushCurrentSnapshotToUndoHistory();
  recordBaselineSnapshot();
}

export function clearHistorySnapshots() {
  undoHistorySnapshots.length = 0;
  redoHistorySnapshots.length = 0;
  lastBoardSnapshotJSON = null;
}

export function hasUndoSnapshots() {
  return undoHistorySnapshots.length > 0;
}

export function hasRedoSnapshots() {
  return redoHistorySnapshots.length > 0;
}

export function takeUndoSnapshot() {
  const snapshotString = undoHistorySnapshots.pop();
  if (lastBoardSnapshotJSON) {
    redoHistorySnapshots.push(lastBoardSnapshotJSON);
  }
  return snapshotString;
}

export function takeRedoSnapshot() {
  const snapshotString = redoHistorySnapshots.pop();
  if (lastBoardSnapshotJSON) {
    undoHistorySnapshots.push(lastBoardSnapshotJSON);
  }
  return snapshotString;
}
