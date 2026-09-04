import { ATTACHMENT_KIND_ICON_NAMES } from "./constants.js";
import { invokeBackendMethod } from "./backend.js";
import { createElement } from "./dom.js";
import { createAttachmentImageElement } from "./attachments.js";
import { createIconElement } from "./icons.js";
import { formatByteSize } from "./formatting.js";
import { cardModalEditing } from "./state.js";
import { pushCurrentSnapshotToUndoHistory } from "./history.js";
import { createInlineEditInput, scheduleFocusAtEnd } from "./inline_editing.js";
import { confirmDeletion } from "./overlays.js";
import { performBackendMutationThenRefresh, persistLocalChange, refreshCurrentBoard } from "./board_actions.js";
import { renderCardModalDialog, rerenderCardViews } from "./card_modal.js";
import { openLightboxModal } from "./lightbox.js";

const PREVIEWABLE_KINDS = new Set(["image", "video", "audio"]);
const PLAYABLE_KINDS = new Set(["video", "audio"]);

function openAttachmentExternally(attachment) {
  invokeBackendMethod("open_attachment", attachment.id).catch(() => {});
}

export function saveAttachmentAs(attachment) {
  invokeBackendMethod("save_attachment_as", attachment.id).catch(() => {});
}

function previewAttachmentElement(attachment) {
  if (PREVIEWABLE_KINDS.has(attachment.kind)) {
    openLightboxModal(attachment);
  } else {
    openAttachmentExternally(attachment);
  }
}

function addAttachmentsThroughDialog(card) {
  invokeBackendMethod("add_attachments_dialog", card.id)
    .then((response) => {
      if (response.attachments.length) {
        pushCurrentSnapshotToUndoHistory();
      }
      return refreshCurrentBoard();
    })
    .catch(() => {});
}

function createAttachmentPreviewElement(card, attachment, isCover) {
  const previewElement = createElement("div", {
    class: "attachment-preview",
    onclick: () => previewAttachmentElement(attachment),
  });
  if (attachment.kind === "image") {
    previewElement.append(createAttachmentImageElement(attachment, { alt: attachment.name }, false));
    if (isCover) {
      previewElement.append(createElement("div", { class: "cover-badge" }, "Cover"));
    }
    previewElement.addEventListener("contextmenu", (contextMenuEvent) => {
      contextMenuEvent.preventDefault();
      card.cover_id = isCover ? null : attachment.id;
      persistLocalChange("update_card", card.id, { cover_id: card.cover_id });
      rerenderCardViews(card);
    });
    return previewElement;
  }

  previewElement.append(
    createElement("div", { class: "attachment-icon" }, createIconElement(ATTACHMENT_KIND_ICON_NAMES[attachment.kind] || "file", 30))
  );
  if (PLAYABLE_KINDS.has(attachment.kind)) {
    previewElement.append(createElement("div", { class: "play-badge" }, createIconElement("play", 22, true)));
  }
  return previewElement;
}

function createAttachmentNameElement(attachment) {
  if (cardModalEditing.renamingAttachmentIdentifier !== attachment.id) {
    return createElement(
      "div",
      {
        class: "attachment-name",
        onclick: () => {
          cardModalEditing.renamingAttachmentIdentifier = attachment.id;
          renderCardModalDialog();
        },
      },
      attachment.name
    );
  }
  const nameInputElement = createInlineEditInput({
    className: "attachment-name-input",
    value: attachment.name,
    onCancel: () => {
      cardModalEditing.renamingAttachmentIdentifier = null;
      renderCardModalDialog();
    },
    onCommit: (rawValue) => {
      const trimmedValue = rawValue.trim();
      cardModalEditing.renamingAttachmentIdentifier = null;
      if (trimmedValue && trimmedValue !== attachment.name) {
        performBackendMutationThenRefresh("rename_attachment", attachment.id, trimmedValue);
      } else {
        renderCardModalDialog();
      }
    },
  });
  scheduleFocusAtEnd(nameInputElement);
  return nameInputElement;
}

export async function deleteAttachmentWithConfirmation(attachment) {
  const isConfirmed = await confirmDeletion(attachment.name);
  if (!isConfirmed) {
    return false;
  }
  performBackendMutationThenRefresh("delete_attachment", attachment.id);
  return true;
}

function createAttachmentItemElement(card, attachment) {
  const isCover = card.cover_id === attachment.id;
  return createElement(
    "div",
    { class: "attachment-item" + (isCover ? " cover" : "") },
    createAttachmentPreviewElement(card, attachment, isCover),
    createElement(
      "div",
      { class: "attachment-meta" },
      createAttachmentNameElement(attachment),
      createElement("div", { class: "attachment-size" }, formatByteSize(attachment.size))
    ),
    createElement(
      "div",
      { class: "attachment-actions" },
      createElement("button", { dataset: { tooltip: "Save as" }, onclick: () => saveAttachmentAs(attachment) }, createIconElement("download", 14)),
      createElement(
        "button",
        { class: "danger", dataset: { tooltip: "Delete attachment" }, onclick: () => deleteAttachmentWithConfirmation(attachment) },
        createIconElement("trash", 14)
      )
    )
  );
}

export function createAttachmentSection(card) {
  return createElement(
    "div",
    { class: "modal-section" },
    createElement(
      "h3",
      {},
      "Attachments",
      createElement("button", { class: "button add-file-button", onclick: () => addAttachmentsThroughDialog(card) }, "+ Add file")
    ),
    createElement(
      "div",
      { class: "attachment-wrap" },
      card.attachments.length
        ? createElement("div", { class: "attachment-grid" }, card.attachments.map((attachment) => createAttachmentItemElement(card, attachment)))
        : createElement("div", { class: "attachment-empty" }, "Drop files anywhere on this card to attach them")
    )
  );
}
