import { MAXIMUM_TAG_CHARACTERS, MAXIMUM_TAG_SUGGESTIONS, SUGGESTION_HIDE_DELAY_MILLISECONDS } from "./constants.js";
import { createElement, createInlineElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { normalizeTagQuery } from "./markup.js";
import { scheduleFocusAtEnd } from "./inline_editing.js";

export function sanitizeTagText(rawTagText) {
  const cleanedTag = (rawTagText || "").trim().replace(/^#/, "");
  return cleanedTag.length > MAXIMUM_TAG_CHARACTERS ? cleanedTag.slice(0, MAXIMUM_TAG_CHARACTERS) : cleanedTag;
}

export function hasTagIgnoringCase(tags, candidateTag) {
  const lowerCandidate = candidateTag.toLowerCase();
  return tags.some((existingTag) => existingTag.toLowerCase() === lowerCandidate);
}

export function createTagSuggestionInput({ getSuggestedTags, onSubmitTag }) {
  const suggestionContainer = createElement("div", { class: "tag-suggest", hidden: true });

  function updateSuggestions() {
    const query = normalizeTagQuery(tagInputElement.value);
    const matchingOptions = getSuggestedTags(query).slice(0, MAXIMUM_TAG_SUGGESTIONS);
    suggestionContainer.innerHTML = "";
    if (!matchingOptions.length) {
      suggestionContainer.hidden = true;
      return;
    }
    for (const optionTag of matchingOptions) {
      suggestionContainer.append(
        createInlineElement(
          "button",
          {
            class: "tag-option",
            onmousedown: (mouseEvent) => {
              mouseEvent.preventDefault();
              tagInputElement.value = "";
              suggestionContainer.hidden = true;
              tagInputElement.blur();
              onSubmitTag(optionTag, tagInputElement);
            },
          },
          optionTag
        )
      );
    }
    suggestionContainer.hidden = false;
  }

  const tagInputElement = createElement("input", {
    class: "tag-input",
    placeholder: "Add tag…",
    maxlength: MAXIMUM_TAG_CHARACTERS,
    oninput: updateSuggestions,
    onfocus: updateSuggestions,
    onblur: (blurEvent) => {
      onSubmitTag(blurEvent.target.value, tagInputElement);
      setTimeout(() => {
        suggestionContainer.hidden = true;
      }, SUGGESTION_HIDE_DELAY_MILLISECONDS);
    },
    onkeydown: (keyboardEvent) => {
      if (keyboardEvent.key === "Escape") {
        keyboardEvent.stopPropagation();
        suggestionContainer.hidden = true;
        keyboardEvent.target.value = "";
        keyboardEvent.target.blur();
        return;
      }
      if (keyboardEvent.key === "Enter") {
        keyboardEvent.preventDefault();
        const targetValue = keyboardEvent.target.value;
        keyboardEvent.target.value = "";
        suggestionContainer.hidden = true;
        keyboardEvent.target.blur();
        onSubmitTag(targetValue, tagInputElement);
      }
    },
  });

  return createElement("div", { class: "tag-input-wrap" }, tagInputElement, suggestionContainer);
}

function calculateTagInputWidth(text) {
  return Math.max(4, text.length + 2) + "ch";
}

export function createTagEditInput({ tagText, onCancel, onCommit }) {
  const editInputElement = createElement("input", {
    class: "tag-edit-input",
    value: tagText,
    maxlength: MAXIMUM_TAG_CHARACTERS,
    style: { width: calculateTagInputWidth(tagText) },
    oninput: (inputEvent) => {
      inputEvent.target.style.width = calculateTagInputWidth(inputEvent.target.value);
    },
    onkeydown: (keyboardEvent) => {
      if (keyboardEvent.key === "Enter") {
        keyboardEvent.target.blur();
      }
      if (keyboardEvent.key === "Escape") {
        keyboardEvent.stopPropagation();
        onCancel();
      }
    },
    onblur: (blurEvent) => onCommit(sanitizeTagText(blurEvent.target.value)),
  });
  scheduleFocusAtEnd(editInputElement);
  return editInputElement;
}

export function createTagRemoveButton(onRemove) {
  return createElement("button", { class: "tag-remove", dataset: { tooltip: "Remove tag" }, onclick: onRemove }, createIconElement("x", 11));
}
