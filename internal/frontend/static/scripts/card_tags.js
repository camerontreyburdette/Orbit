import { createElement, createInlineElement } from "./dom.js";
import { matchesTagFilter } from "./markup.js";
import { cardModalEditing } from "./state.js";
import { collectAllBoardTags } from "./board_queries.js";
import { persistLocalChange } from "./board_actions.js";
import { renderCardModalDialog, rerenderCardViews } from "./card_modal.js";
import { createTagEditInput, createTagRemoveButton, createTagSuggestionInput, hasTagIgnoringCase, sanitizeTagText } from "./tag_editor.js";

function deduplicateTagsIgnoringCase(tags) {
  const seenSet = new Set();
  return tags.filter((item) => {
    const lowerKey = item.toLowerCase();
    if (seenSet.has(lowerKey)) {
      return false;
    }
    seenSet.add(lowerKey);
    return true;
  });
}

export function createTagSection(card) {
  const cardTags = card.tags || [];

  function saveTagsList(nextTagsList) {
    card.tags = nextTagsList;
    persistLocalChange("update_card", card.id, { tags: nextTagsList });
    rerenderCardViews(card);
  }

  function addTagToCard(rawTagText, tagInputElement) {
    const cleanedTag = sanitizeTagText(rawTagText);
    if (!cleanedTag) {
      return;
    }
    if (hasTagIgnoringCase(cardTags, cleanedTag)) {
      tagInputElement.value = "";
      return;
    }
    saveTagsList([...cardTags, cleanedTag]);
  }

  const tagInputWrapper = createTagSuggestionInput({
    getSuggestedTags: (query) =>
      collectAllBoardTags().filter((tag) => !hasTagIgnoringCase(cardTags, tag) && matchesTagFilter(tag, query)),
    onSubmitTag: addTagToCard,
  });

  function createTagChipElement(tagText, tagIndex) {
    if (cardModalEditing.editingTagIndex === tagIndex) {
      return createTagEditInput({
        tagText,
        onCancel: () => {
          cardModalEditing.editingTagIndex = null;
          renderCardModalDialog();
        },
        onCommit: (trimmedValue) => {
          cardModalEditing.editingTagIndex = null;
          if (!trimmedValue || trimmedValue === tagText) {
            renderCardModalDialog();
            return;
          }
          const nextTags = [...cardTags];
          nextTags[tagIndex] = trimmedValue;
          saveTagsList(deduplicateTagsIgnoringCase(nextTags));
        },
      });
    }
    return createElement(
      "span",
      { class: "tag-chip" },
      createInlineElement(
        "span",
        {
          class: "tag-text",
          onclick: () => {
            cardModalEditing.editingTagIndex = tagIndex;
            renderCardModalDialog();
          },
        },
        tagText
      ),
      createTagRemoveButton(() => saveTagsList(cardTags.filter((item) => item !== tagText)))
    );
  }

  return createElement("div", { class: "tag-row" }, cardTags.map(createTagChipElement), tagInputWrapper);
}
