import { createElement } from "./dom.js";
import { renderBlockMarkup } from "./markup.js";
import { state, cardModalEditing } from "./state.js";
import { syncDiscordPresence } from "./presence.js";
import { createInlineEditInput } from "./inline_editing.js";
import { persistLocalChange } from "./board_actions.js";
import { renderCardModalDialog, rerenderCardViews } from "./card_modal.js";

function adjustTextareaHeight(textarea) {
  textarea.style.height = "auto";
  textarea.style.height = Math.max(textarea.scrollHeight, cardModalEditing.descriptionEditFromHeight) + "px";
}

function createDescriptionEditor(card) {
  const textareaElement = createInlineEditInput({
    tagName: "textarea",
    className: "description-box",
    rows: 1,
    value: card.description || "",
    placeholder: "Add a description…",
    commitsOnShiftEnter: false,
    onInput: (inputEvent) => {
      adjustTextareaHeight(inputEvent.target);
      syncDiscordPresence(card.title, true);
    },
    onCancel: () => {
      state.isEditingDescription = false;
      renderCardModalDialog();
    },
    onCommit: (value) => {
      state.isEditingDescription = false;
      if (value !== card.description) {
        card.description = value;
        persistLocalChange("update_card", card.id, { description: value });
      }
      rerenderCardViews(card);
    },
  });
  if (cardModalEditing.descriptionEditFromHeight) {
    textareaElement.style.height = cardModalEditing.descriptionEditFromHeight + "px";
  }
  setTimeout(() => {
    textareaElement.focus();
    textareaElement.setSelectionRange(textareaElement.value.length, textareaElement.value.length);
    adjustTextareaHeight(textareaElement);
  }, 0);
  return textareaElement;
}

function createDescriptionPreview(card) {
  const previewElement = createElement("div", {
    class: "description-preview" + (card.description ? "" : " empty"),
    onclick: (clickEvent) => {
      cardModalEditing.descriptionEditFromHeight = clickEvent.currentTarget.offsetHeight;
      state.isEditingDescription = true;
      renderCardModalDialog();
      syncDiscordPresence(card.title, true);
    },
  });
  if (card.description) {
    previewElement.innerHTML = renderBlockMarkup(card.description);
  } else {
    previewElement.textContent = "Add a description…";
  }
  return previewElement;
}

export function createDescriptionSection(card) {
  return createElement(
    "div",
    { class: "modal-section" },
    createElement("h3", {}, "Description"),
    state.isEditingDescription ? createDescriptionEditor(card) : createDescriptionPreview(card)
  );
}
