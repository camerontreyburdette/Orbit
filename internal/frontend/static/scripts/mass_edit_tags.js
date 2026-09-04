import { MAXIMUM_TAGS_PER_CARD } from "./constants.js";
import { invokeBackendMethod } from "./backend.js";
import { createElement, createInlineElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { matchesTagFilter } from "./markup.js";
import { selectedCardIdentifiers } from "./state.js";
import { collectAllBoardTags, collectSelectedCards } from "./board_queries.js";
import { pushCurrentSnapshotToUndoHistory, saveLocalBoardSnapshotState } from "./history.js";
import { refreshCurrentBoard } from "./board_actions.js";
import { renderMainApplicationView } from "./render.js";
import { openModalOverlay } from "./overlays.js";
import { createTagEditInput, createTagRemoveButton, createTagSuggestionInput, hasTagIgnoringCase, sanitizeTagText } from "./tag_editor.js";

let editingMassTagIndex = null;

function summarizeSelectedTags(selectedCards) {
  const uniqueTagsMap = new Map();
  const tagCardCountMap = new Map();
  for (const currentCard of selectedCards) {
    const cardSeenTags = new Set();
    for (const tagText of currentCard.tags || []) {
      const lowerCaseTag = tagText.toLowerCase();
      if (!uniqueTagsMap.has(lowerCaseTag)) {
        uniqueTagsMap.set(lowerCaseTag, tagText);
      }
      if (!cardSeenTags.has(lowerCaseTag)) {
        cardSeenTags.add(lowerCaseTag);
        tagCardCountMap.set(lowerCaseTag, (tagCardCountMap.get(lowerCaseTag) || 0) + 1);
      }
    }
  }
  return { tagsList: [...uniqueTagsMap.values()], tagCardCountMap };
}

function addTagLocally(cleanedTag) {
  for (const currentCard of collectSelectedCards()) {
    if (!currentCard.tags) {
      currentCard.tags = [];
    }
    if (!hasTagIgnoringCase(currentCard.tags, cleanedTag) && currentCard.tags.length < MAXIMUM_TAGS_PER_CARD) {
      currentCard.tags.push(cleanedTag);
    }
  }
}

function removeTagLocally(tagToRemove) {
  const lowerTag = tagToRemove.toLowerCase();
  for (const currentCard of collectSelectedCards()) {
    if (currentCard.tags) {
      currentCard.tags = currentCard.tags.filter((existingTag) => existingTag.toLowerCase() !== lowerTag);
    }
  }
}

function renameTagLocally(oldTag, newTag) {
  const lowerOldTag = oldTag.toLowerCase();
  for (const currentCard of collectSelectedCards()) {
    if (currentCard.tags) {
      currentCard.tags = currentCard.tags.map((tag) => (tag.toLowerCase() === lowerOldTag ? newTag : tag));
    }
  }
}

export function openMassEditTagsModalDialog() {
  editingMassTagIndex = null;

  const dialogBoxElement = createElement("div", { class: "confirm-box dialog-box mass-edit-tags-box" });
  const closeModal = openModalOverlay(dialogBoxElement);

  function applyBatchTagUpdate(fields, applyLocally) {
    pushCurrentSnapshotToUndoHistory();
    invokeBackendMethod("batch_update_cards", [...selectedCardIdentifiers], fields)
      .then(() => {
        applyLocally();
        saveLocalBoardSnapshotState();
        renderMainApplicationView();
        renderTagsDialogContent();
      })
      .catch(refreshCurrentBoard);
  }

  function renderTagsDialogContent() {
    dialogBoxElement.innerHTML = "";

    const selectedCardsList = collectSelectedCards();
    const { tagsList, tagCardCountMap } = summarizeSelectedTags(selectedCardsList);

    function isTagOnAllCards(tagText) {
      return (tagCardCountMap.get(tagText.toLowerCase()) || 0) === selectedCardsList.length;
    }

    function addTagToSelectedCards(rawTagText, tagInputElement) {
      const cleanedTag = sanitizeTagText(rawTagText);
      if (!cleanedTag) {
        return;
      }
      const isAlreadyOnAllCards =
        selectedCardsList.length > 0 &&
        selectedCardsList.every((currentCard) => hasTagIgnoringCase(currentCard.tags || [], cleanedTag));
      if (isAlreadyOnAllCards) {
        tagInputElement.value = "";
        return;
      }
      applyBatchTagUpdate({ add_tag: cleanedTag }, () => addTagLocally(cleanedTag));
    }

    function removeTagFromSelectedCards(tagToRemove) {
      applyBatchTagUpdate({ remove_tag: tagToRemove }, () => removeTagLocally(tagToRemove));
    }

    function renameTagOnSelectedCards(oldTag, newTag) {
      if (!newTag || oldTag === newTag) {
        renderTagsDialogContent();
        return;
      }
      applyBatchTagUpdate({ remove_tag: oldTag, add_tag: newTag }, () => renameTagLocally(oldTag, newTag));
    }

    const tagInputWrapper = createTagSuggestionInput({
      getSuggestedTags: (query) => {
        const fullySharedTagSet = new Set(tagsList.filter(isTagOnAllCards).map((tag) => tag.toLowerCase()));
        return collectAllBoardTags().filter(
          (tag) => !fullySharedTagSet.has(tag.toLowerCase()) && matchesTagFilter(tag, query)
        );
      },
      onSubmitTag: addTagToSelectedCards,
    });
    const tagInputElement = tagInputWrapper.firstElementChild;

    function createTagChipElement(tagText, tagIndex) {
      const isPresentOnAllCards = isTagOnAllCards(tagText);

      if (editingMassTagIndex === tagIndex) {
        return createTagEditInput({
          tagText,
          onCancel: () => {
            editingMassTagIndex = null;
            renderTagsDialogContent();
          },
          onCommit: (trimmedValue) => {
            editingMassTagIndex = null;
            if (!trimmedValue || trimmedValue === tagText) {
              renderTagsDialogContent();
              return;
            }
            renameTagOnSelectedCards(tagText, trimmedValue);
          },
        });
      }

      const handleTagTextClick = (clickEvent) => {
        if (!isPresentOnAllCards) {
          clickEvent.stopPropagation();
          addTagToSelectedCards(tagText, tagInputElement);
          return;
        }
        editingMassTagIndex = tagIndex;
        renderTagsDialogContent();
      };

      return createElement(
        "span",
        {
          class: "tag-chip" + (isPresentOnAllCards ? "" : " partial"),
          onclick: (clickEvent) => {
            if (isPresentOnAllCards) {
              return;
            }
            clickEvent.stopPropagation();
            addTagToSelectedCards(tagText, tagInputElement);
          },
        },
        createInlineElement("span", { class: "tag-text", onclick: handleTagTextClick, ondblclick: handleTagTextClick }, tagText),
        createTagRemoveButton((clickEvent) => {
          clickEvent.stopPropagation();
          removeTagFromSelectedCards(tagText);
        })
      );
    }

    dialogBoxElement.append(
      createElement(
        "div",
        { class: "dialog-head" },
        createElement("h3", {}, "Tags"),
        createElement("button", { class: "icon-button", dataset: { tooltip: "Close" }, onclick: () => closeModal() }, createIconElement("x", 16))
      ),
      createElement("div", { class: "tag-row" }, tagsList.map(createTagChipElement), tagInputWrapper)
    );
  }

  renderTagsDialogContent();
}
