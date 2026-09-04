import { invokeBackendMethod } from "./backend.js";
import {
  clearHistorySnapshots,
  hasRedoSnapshots,
  hasUndoSnapshots,
  pushCurrentSnapshotToUndoHistory,
  recordBaselineSnapshot,
  saveLocalBoardSnapshotState,
  takeRedoSnapshot,
  takeUndoSnapshot,
} from "./history.js";
import { state, resetBoardViewState } from "./state.js";
import { clearMultiSelectionState } from "./selection.js";
import { renderMainApplicationView } from "./render.js";
import { renderCardModalDialog } from "./card_modal.js";
import { confirmDeletion } from "./overlays.js";

function reportActiveBoard(boardIdentifier) {
  invokeBackendMethod("set_active_board", boardIdentifier == null ? 0 : boardIdentifier).catch(() => {});
}

export async function refreshBoardsList() {
  state.boards = await invokeBackendMethod("get_boards");
}

export async function refreshCurrentBoard() {
  if (state.boardIdentifier == null) {
    state.currentBoard = null;
  } else {
    try {
      state.currentBoard = await invokeBackendMethod("get_board", state.boardIdentifier);
    } catch {
      state.boardIdentifier = null;
      state.currentBoard = null;
    }
  }
  recordBaselineSnapshot();
  renderMainApplicationView();
  renderCardModalDialog();
}

export function switchActiveBoard(targetBoardIdentifier) {
  state.boardIdentifier = targetBoardIdentifier;
  reportActiveBoard(targetBoardIdentifier);
  resetBoardViewState();
  clearHistorySnapshots();
  clearMultiSelectionState();
  refreshCurrentBoard();
}

export async function navigateToHomeScreen() {
  state.boardIdentifier = null;
  reportActiveBoard(null);
  state.currentBoard = null;
  resetBoardViewState();
  clearHistorySnapshots();
  clearMultiSelectionState();
  renderCardModalDialog();
  try {
    await refreshBoardsList();
  } catch {}
  renderMainApplicationView();
}

export async function createNewBoard(boardName, boardDescription = "") {
  try {
    const response = await invokeBackendMethod("create_board", boardName, boardDescription);
    if (response && response.boards) {
      state.boards = response.boards;
    } else {
      await refreshBoardsList();
    }
  } catch {
    return;
  }
  renderMainApplicationView();
}

async function applyBoardListResponse(requestPromise) {
  try {
    const response = await requestPromise;
    if (response && response.boards) {
      state.boards = response.boards;
      renderMainApplicationView();
    }
  } catch {
    await refreshBoardsList();
    renderMainApplicationView();
  }
}

export async function deleteExistingBoard(boardSummary) {
  const isConfirmed = await confirmDeletion(boardSummary.name);
  if (!isConfirmed) {
    return;
  }
  state.boards = state.boards.filter((item) => item.id !== boardSummary.id);
  renderMainApplicationView();
  await applyBoardListResponse(invokeBackendMethod("delete_board", boardSummary.id));
}

export async function updateBoardFields(boardIdentifier, updatedFields) {
  const targetBoardSummary = state.boards.find((item) => item.id === boardIdentifier);
  if (targetBoardSummary) {
    Object.assign(targetBoardSummary, updatedFields);
    renderMainApplicationView();
  }
  await applyBoardListResponse(invokeBackendMethod("update_board", boardIdentifier, updatedFields));
}

async function restoreBoardSnapshot(snapshotString) {
  try {
    await invokeBackendMethod("restore_board", state.boardIdentifier, JSON.parse(snapshotString));
  } catch {
    return;
  }
  refreshCurrentBoard();
}

export async function undoBoardAction() {
  if (!hasUndoSnapshots() || !state.currentBoard) {
    return;
  }
  await restoreBoardSnapshot(takeUndoSnapshot());
}

export async function redoBoardAction() {
  if (!hasRedoSnapshots() || !state.currentBoard) {
    return;
  }
  await restoreBoardSnapshot(takeRedoSnapshot());
}

export function performBackendMutationThenRefresh(methodName, ...methodArguments) {
  return invokeBackendMethod(methodName, ...methodArguments)
    .then(() => {
      pushCurrentSnapshotToUndoHistory();
      return refreshCurrentBoard();
    })
    .catch(() => {});
}

export function persistLocalChange(methodName, ...methodArguments) {
  invokeBackendMethod(methodName, ...methodArguments).catch(refreshCurrentBoard);
  saveLocalBoardSnapshotState();
}
