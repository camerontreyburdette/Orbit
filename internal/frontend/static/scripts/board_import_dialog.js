import { invokeBackendMethod } from "./backend.js";
import { createElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { state } from "./state.js";
import { openModalOverlay } from "./overlays.js";
import { createFileDropZoneProperties } from "./file_drop.js";
import { refreshBoardsList } from "./board_actions.js";
import { renderMainApplicationView } from "./render.js";
import { collectDroppedFiles, groupFilesIntoBoardPackages } from "./board_import_entries.js";
import { uploadBoardPackage } from "./board_import_upload.js";

const IMPORT_OUTCOME_DISPLAY_MILLISECONDS = 1800;
const IMPORT_ZONE_STATE_CLASSES = ["importing", "success", "failure"];

function createDropZoneStateController(zoneElement) {
  let clearTimer = null;
  function applyState(stateName) {
    clearTimeout(clearTimer);
    zoneElement.classList.remove(...IMPORT_ZONE_STATE_CLASSES);
    if (stateName) {
      zoneElement.classList.add(stateName);
    }
  }
  return {
    showImporting: () => applyState("importing"),
    clear: () => applyState(null),
    showOutcome(stateName) {
      applyState(stateName);
      clearTimer = setTimeout(() => applyState(null), IMPORT_OUTCOME_DISPLAY_MILLISECONDS);
    },
  };
}

async function applyImportedBoardList(importResponse) {
  if (importResponse && importResponse.boards) {
    state.boards = importResponse.boards;
  } else {
    try {
      await refreshBoardsList();
    } catch {}
  }
  renderMainApplicationView();
}

async function importBoardPackages(boardPackages) {
  let hasFailure = false;
  for (const boardPackage of boardPackages) {
    try {
      await applyImportedBoardList(await uploadBoardPackage(boardPackage));
    } catch (importError) {
      console.error(importError);
      hasFailure = true;
    }
  }
  return !hasFailure;
}

async function importThroughFolderDialog() {
  const importResponse = await invokeBackendMethod("import_boards_dialog");
  if (importResponse && importResponse.cancelled) {
    return null;
  }
  await applyImportedBoardList(importResponse);
  return importResponse.imported_count > 0 && importResponse.failed_count === 0;
}

function createImportQueue(zoneState) {
  let queuePromise = Promise.resolve();
  return (runImport) => {
    queuePromise = queuePromise.then(async () => {
      zoneState.showImporting();
      let isSuccessful = false;
      try {
        isSuccessful = await runImport();
      } catch (importError) {
        console.error(importError);
      }
      if (isSuccessful === null) {
        zoneState.clear();
        return;
      }
      zoneState.showOutcome(isSuccessful ? "success" : "failure");
    });
  };
}

function createImportDropZone(onFolderClick, onFilesDropped) {
  return createElement(
    "div",
    {
      class: "import-drop-zone",
      onclick: onFolderClick,
      ...createFileDropZoneProperties(async (dragEvent) => onFilesDropped(await collectDroppedFiles(dragEvent.dataTransfer))),
    },
    createIconElement("upload", 22),
    createElement("span", { class: "import-drop-title" }, "Drop board folders here"),
    createElement("span", { class: "import-drop-hint" }, "or click to choose a folder")
  );
}

function createImportDialogHead(closeModal) {
  return createElement(
    "div",
    { class: "dialog-head" },
    createElement("h3", {}, "Import boards"),
    createElement("button", { class: "icon-button", dataset: { tooltip: "Close" }, onclick: () => closeModal() }, createIconElement("x", 16))
  );
}

export function openBoardImportModalDialog() {
  let enqueueImport = null;

  const handleFolderClick = () => enqueueImport(importThroughFolderDialog);
  const handleDroppedFiles = (collectedFiles) => {
    const boardPackages = groupFilesIntoBoardPackages(collectedFiles);
    enqueueImport(() => (boardPackages.length ? importBoardPackages(boardPackages) : false));
  };

  const zoneElement = createImportDropZone(handleFolderClick, handleDroppedFiles);
  enqueueImport = createImportQueue(createDropZoneStateController(zoneElement));

  const dialogBoxElement = createElement("div", { class: "confirm-box dialog-box import-box" });
  const closeModal = openModalOverlay(dialogBoxElement);
  dialogBoxElement.append(createImportDialogHead(closeModal), zoneElement);
}
