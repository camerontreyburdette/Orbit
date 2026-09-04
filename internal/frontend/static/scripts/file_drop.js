import { pushCurrentSnapshotToUndoHistory } from "./history.js";
import { refreshCurrentBoard } from "./board_actions.js";
import { attachmentUploadUniformResourceLocator } from "./attachments.js";

function isFileDragEvent(dragEvent) {
  return Boolean(dragEvent.dataTransfer) && [...(dragEvent.dataTransfer.types || [])].includes("Files");
}

function adjustFileHoverCount(targetElement, delta) {
  targetElement.fileHoverCount = Math.max(0, (targetElement.fileHoverCount || 0) + delta);
  targetElement.classList.toggle("file-target", targetElement.fileHoverCount > 0);
}

export function createFileDropZoneProperties(onFileDrop) {
  return {
    ondragenter: (dragEvent) => {
      if (!isFileDragEvent(dragEvent)) {
        return;
      }
      adjustFileHoverCount(dragEvent.currentTarget, 1);
    },
    ondragover: (dragEvent) => {
      if (!isFileDragEvent(dragEvent)) {
        return;
      }
      dragEvent.preventDefault();
      dragEvent.stopPropagation();
      dragEvent.dataTransfer.dropEffect = "copy";
    },
    ondragleave: (dragEvent) => {
      if (!isFileDragEvent(dragEvent)) {
        return;
      }
      adjustFileHoverCount(dragEvent.currentTarget, -1);
    },
    ondrop: (dragEvent) => {
      if (!isFileDragEvent(dragEvent)) {
        return;
      }
      dragEvent.preventDefault();
      dragEvent.stopPropagation();
      const currentTargetElement = dragEvent.currentTarget;
      currentTargetElement.fileHoverCount = 0;
      currentTargetElement.classList.remove("file-target");
      onFileDrop(dragEvent);
    },
  };
}

export function createFileDropProperties(getCardIdentifierCallback) {
  return createFileDropZoneProperties((dragEvent) => uploadFiles(getCardIdentifierCallback(), dragEvent.dataTransfer.files));
}

async function uploadSingleFile(cardIdentifier, fileObject) {
  const formData = new FormData();
  formData.append("file", fileObject, fileObject.name);
  const response = await fetch(attachmentUploadUniformResourceLocator(cardIdentifier), { method: "POST", body: formData });
  if (!response.ok) {
    throw new Error("Upload failed for " + fileObject.name + ": " + (await response.text()));
  }
}

async function uploadFiles(cardIdentifier, fileList) {
  const fileArray = [...fileList];
  if (!fileArray.length || cardIdentifier == null) {
    return;
  }
  for (const fileObject of fileArray) {
    try {
      await uploadSingleFile(cardIdentifier, fileObject);
    } catch (uploadError) {
      console.error(uploadError);
    }
  }
  pushCurrentSnapshotToUndoHistory();
  refreshCurrentBoard();
}

function preventDefaultForFileDrag(dragEvent) {
  if (isFileDragEvent(dragEvent)) {
    dragEvent.preventDefault();
  }
}

export function installDocumentFileDropGuards() {
  document.addEventListener("dragover", preventDefaultForFileDrag);
  document.addEventListener("drop", preventDefaultForFileDrag);
}
