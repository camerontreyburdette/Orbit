import { MODAL_ENTER_ANIMATION_WINDOW_MILLISECONDS } from "./constants.js";
import { createElement, createInlineElement, querySelector } from "./dom.js";
import { createIconElement } from "./icons.js";
import { renderInlineMarkup } from "./markup.js";
import { extractDatePortion } from "./formatting.js";
import { state, resetCardModalEditingState } from "./state.js";
import { findCardByIdentifier, findListByIdentifier } from "./board_queries.js";
import { syncDiscordPresence } from "./presence.js";
import { bindOverlayCloseHandler, closeOverlayElement, confirmDeletion } from "./overlays.js";
import { createInlineEditInput, scheduleFocusAtEnd } from "./inline_editing.js";
import { persistLocalChange, performBackendMutationThenRefresh } from "./board_actions.js";
import { renderCardOnBoard } from "./render.js";
import { createFileDropProperties } from "./file_drop.js";
import { createColorSwatchRow } from "./color_swatches.js";
import { createTagSection } from "./card_tags.js";
import { createDescriptionSection } from "./card_description.js";
import { createChecklistSection } from "./checklists.js";
import { createAttachmentSection } from "./attachment_section.js";

let modalOpenedTimestamp = 0;

export function openCardModalDialog(cardIdentifier) {
  state.openCardIdentifier = cardIdentifier;
  resetCardModalEditingState();
  modalOpenedTimestamp = Date.now();
  renderCardModalDialog();
}

export function closeCardModalDialog() {
  state.openCardIdentifier = null;
  renderCardModalDialog();
}

export function rerenderCardViews(card) {
  renderCardOnBoard(card);
  renderCardModalDialog();
}

export async function deleteOpenCardAction() {
  const currentCard = findCardByIdentifier(state.openCardIdentifier);
  if (!currentCard) {
    return;
  }
  const isConfirmed = await confirmDeletion(currentCard.title);
  if (!isConfirmed) {
    return;
  }
  state.openCardIdentifier = null;
  renderCardModalDialog();
  performBackendMutationThenRefresh("delete_card", currentCard.id);
}

function createTitleEditor(currentCard) {
  const titleInputElement = createInlineEditInput({
    className: "modal-title",
    value: currentCard.title,
    onInput: (inputEvent) => {
      const liveTitle = inputEvent.target.value.trim() || currentCard.title;
      syncDiscordPresence(liveTitle, true);
    },
    onCancel: () => {
      state.isEditingTitle = false;
      renderCardModalDialog();
    },
    onCommit: (rawValue) => {
      const trimmedValue = rawValue.trim();
      state.isEditingTitle = false;
      if (trimmedValue && trimmedValue !== currentCard.title) {
        currentCard.title = trimmedValue;
        persistLocalChange("update_card", currentCard.id, { title: trimmedValue });
      }
      rerenderCardViews(currentCard);
    },
  });
  scheduleFocusAtEnd(titleInputElement);
  return titleInputElement;
}

function createTitleView(currentCard) {
  const titleElement = createInlineElement("div", { class: "modal-title-view" }, currentCard.title);
  titleElement.addEventListener("click", () => {
    state.isEditingTitle = true;
    renderCardModalDialog();
    syncDiscordPresence(currentCard.title, true);
  });
  return titleElement;
}

function createSubtitleRow(currentCard) {
  const parentList = findListByIdentifier(currentCard.list_id);
  const subtitleContainer = createElement("div", { class: "modal-subtitle" }, "in ");
  const listNameElement = createElement("span");
  listNameElement.innerHTML = renderInlineMarkup(parentList ? parentList.name : "?");
  subtitleContainer.append(listNameElement, ` · created ${extractDatePortion(currentCard.created_at)}`);
  return createElement("div", { class: "modal-subrow" }, subtitleContainer, createColorSection(currentCard));
}

function createColorSection(card) {
  return createColorSwatchRow("color-row", card.color || "", (chosenColorHex) => {
    card.color = chosenColorHex;
    persistLocalChange("update_card", card.id, { color: chosenColorHex });
    rerenderCardViews(card);
  });
}

function createModalElement(currentCard, reanimationClass) {
  return createElement(
    "div",
    { class: "modal" + reanimationClass, ...createFileDropProperties(() => currentCard.id) },
    createElement(
      "div",
      { class: "modal-header" },
      state.isEditingTitle ? createTitleEditor(currentCard) : createTitleView(currentCard),
      createElement("button", { class: "icon-button", dataset: { tooltip: "Close" }, onclick: closeCardModalDialog }, createIconElement("x", 18))
    ),
    createSubtitleRow(currentCard),
    createTagSection(currentCard),
    createDescriptionSection(currentCard),
    createChecklistSection(currentCard),
    createAttachmentSection(currentCard),
    createElement(
      "div",
      { class: "modal-footer" },
      createElement("button", { class: "button button-ghost delete-card", onclick: deleteOpenCardAction }, "Delete card")
    )
  );
}

export function renderCardModalDialog() {
  syncDiscordPresence();
  const modalRootElement = querySelector("#modal-root");
  const previousModalElement = modalRootElement.querySelector(".modal");
  const previousScrollTop = previousModalElement ? previousModalElement.scrollTop : 0;
  const currentCard =
    state.openCardIdentifier != null && state.currentBoard ? findCardByIdentifier(state.openCardIdentifier) : null;
  if (!currentCard) {
    state.openCardIdentifier = null;
    closeOverlayElement(modalRootElement);
    return;
  }

  modalRootElement.innerHTML = "";
  const isEnteringTransition = Date.now() - modalOpenedTimestamp < MODAL_ENTER_ANIMATION_WINDOW_MILLISECONDS;
  const reanimationClass = isEnteringTransition ? "" : " re-render";

  const modalElement = createModalElement(currentCard, reanimationClass);
  const overlayElement = createElement("div", { class: "overlay" + reanimationClass }, modalElement);
  bindOverlayCloseHandler(overlayElement, closeCardModalDialog);
  modalRootElement.append(overlayElement);
  modalElement.scrollTop = previousScrollTop;
}
