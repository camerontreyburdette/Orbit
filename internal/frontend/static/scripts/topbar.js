import { SEARCH_COLLAPSE_DELAY_MILLISECONDS, SUGGESTION_HIDE_DELAY_MILLISECONDS } from "./constants.js";
import { createElement, createInlineElement, querySelector } from "./dom.js";
import { createIconElement } from "./icons.js";
import { matchesTagFilter, normalizeTagQuery, renderInlineMarkup } from "./markup.js";
import { state, isInBoardView } from "./state.js";
import { collectAllBoardTags } from "./board_queries.js";
import { focusInputElementAtEnd } from "./inline_editing.js";
import { applyCardSearchHighlights } from "./card_search.js";
import { navigateToHomeScreen } from "./board_actions.js";

function createTagSearchControl() {
  const suggestionContainer = createElement("div", { class: "tag-suggest", hidden: true });

  function updateTagSuggestions() {
    const query = normalizeTagQuery(searchInputElement.value);
    const matchedOptions = collectAllBoardTags()
      .filter((tag) => tag.toLowerCase() !== query && matchesTagFilter(tag, query))
      .slice(0, 50);
    suggestionContainer.innerHTML = "";
    for (const matchedTag of matchedOptions) {
      suggestionContainer.append(
        createInlineElement(
          "button",
          {
            class: "tag-option",
            onmousedown: (mouseEvent) => {
              mouseEvent.preventDefault();
              pickTag(matchedTag);
            },
          },
          matchedTag
        )
      );
    }
    suggestionContainer.hidden = !matchedOptions.length;
    searchWrapperElement.classList.toggle("open", !suggestionContainer.hidden);
  }

  function pickTag(selectedTag) {
    searchInputElement.value = selectedTag;
    state.cardSearchQuery = selectedTag;
    applyCardSearchHighlights();
    suggestionContainer.hidden = true;
    searchWrapperElement.classList.remove("open");
    searchInputElement.blur();
  }

  function collapseSearch() {
    if (!state.isSearchOpen || searchWrapperElement.classList.contains("closing")) {
      return;
    }
    state.isSearchOpen = false;
    state.cardSearchQuery = "";
    applyCardSearchHighlights();
    searchWrapperElement.classList.add("closing");
    searchWrapperElement.classList.remove("open");
    searchToggleButton.classList.remove("active");
    searchInputElement.blur();
    setTimeout(() => {
      if (state.isSearchOpen || !searchWrapperElement.isConnected) {
        return;
      }
      searchWrapperElement.classList.add("collapsed");
      searchWrapperElement.classList.remove("closing", "enter");
      searchInputElement.value = "";
      suggestionContainer.hidden = true;
    }, SEARCH_COLLAPSE_DELAY_MILLISECONDS);
  }

  function expandSearch() {
    if (searchWrapperElement.classList.contains("closing")) {
      return;
    }
    state.isSearchOpen = true;
    searchWrapperElement.classList.remove("collapsed");
    searchWrapperElement.classList.add("enter");
    searchToggleButton.classList.add("active");
    searchInputElement.value = state.cardSearchQuery;
    setTimeout(() => focusInputElementAtEnd(searchInputElement), 0);
  }

  const searchInputElement = createElement("input", {
    class: "tag-search-input",
    value: state.cardSearchQuery,
    placeholder: "Search tags…",
    oninput: (inputEvent) => {
      state.cardSearchQuery = inputEvent.target.value;
      applyCardSearchHighlights();
      updateTagSuggestions();
    },
    onfocus: updateTagSuggestions,
    onblur: (blurEvent) => {
      setTimeout(() => {
        suggestionContainer.hidden = true;
        searchWrapperElement.classList.remove("open");
      }, SUGGESTION_HIDE_DELAY_MILLISECONDS);
      if (!blurEvent.target.value.trim()) {
        collapseSearch();
      }
    },
    onkeydown: (keyboardEvent) => {
      if (keyboardEvent.key === "Escape") {
        keyboardEvent.stopPropagation();
        collapseSearch();
      }
    },
  });

  const searchToggleButton = createElement(
    "button",
    {
      class: "icon-button tag-search-button" + (state.isSearchOpen ? " active" : ""),
      dataset: { tooltip: "Search tags" },
      onclick: () => {
        if (state.isSearchOpen) {
          collapseSearch();
        } else {
          expandSearch();
        }
      },
    },
    createIconElement("tag", 16)
  );

  const searchWrapperElement = createElement(
    "div",
    { class: "tag-search-wrap" + (state.isSearchOpen ? "" : " collapsed") },
    searchInputElement,
    searchToggleButton,
    suggestionContainer
  );
  return searchWrapperElement;
}

function createBoardContextElements(board) {
  const boardNameElement = createElement("span", { class: "topbar-name" });
  boardNameElement.innerHTML = renderInlineMarkup(board.name);
  const contextElements = [
    createElement(
      "button",
      { class: "icon-button back-button", dataset: { tooltip: "Back to boards" }, onclick: navigateToHomeScreen },
      createIconElement("arrowleft", 18)
    ),
    boardNameElement,
  ];
  if (board.description) {
    const descriptionElement = createElement("span", { class: "topbar-description" });
    descriptionElement.innerHTML = "- " + renderInlineMarkup(board.description);
    contextElements.push(descriptionElement);
  }
  return contextElements;
}

export function renderTopbarHeader() {
  const topbarElement = querySelector("#topbar");
  const inBoardView = isInBoardView();
  topbarElement.hidden = !inBoardView;
  const contextContainer = querySelector("#topbar-context");
  contextContainer.innerHTML = "";

  if (!inBoardView) {
    return;
  }

  contextContainer.append(...createBoardContextElements(state.currentBoard.board));
  contextContainer.append(createElement("div", { class: "topbar-right" }, createTagSearchControl()));
}
