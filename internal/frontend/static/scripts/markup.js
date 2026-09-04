import { COLOR_HEX_DEFINITIONS, COLOR_TAG_REGULAR_EXPRESSION } from "./constants.js";

const MEMOIZATION_CAPACITY = 2000;

function createMemoizedTextTransform(transform) {
  const cache = new Map();
  return (rawText) => {
    const text = String(rawText);
    const cachedResult = cache.get(text);
    if (cachedResult !== undefined) {
      return cachedResult;
    }
    const result = transform(text);
    if (cache.size >= MEMOIZATION_CAPACITY) {
      cache.clear();
    }
    cache.set(text, result);
    return result;
  };
}

function escapeHypertextMarkup(text) {
  return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function renderInlineMarkupUncached(text) {
  let formattedString = escapeHypertextMarkup(text);
  formattedString = formattedString.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
  formattedString = formattedString.replace(/(^|[^\w])_([^_]+)_(?=[^\w]|$)/g, "$1<u>$2</u>");
  formattedString = formattedString.replace(/(^|[^\w\-])-(?![\s-])([^-]+?)(?<![\s-])-(?=[^\w\-]|$)/g, "$1<s>$2</s>");
  formattedString = formattedString.replace(/(^|[^*])\*([^*]+)\*(?!\*)/g, "$1<em>$2</em>");
  return formattedString.replace(
    COLOR_TAG_REGULAR_EXPRESSION,
    (_, colorName, matchedText) =>
      `<span style="color:${COLOR_HEX_DEFINITIONS[colorName.toLowerCase()]}">${matchedText}</span>`
  );
}

function stripMarkupFormattingUncached(text) {
  let sanitizedString = text;
  sanitizedString = sanitizedString.replace(/\*\*(.+?)\*\*/g, "$1");
  sanitizedString = sanitizedString.replace(/(^|[^\w])_([^_]+)_(?=[^\w]|$)/g, "$1$2");
  sanitizedString = sanitizedString.replace(/(^|[^\w\-])-(?![\s-])([^-]+?)(?<![\s-])-(?=[^\w\-]|$)/g, "$1$2");
  sanitizedString = sanitizedString.replace(/(^|[^*])\*([^*]+)\*(?!\*)/g, "$1$2");
  return sanitizedString.replace(COLOR_TAG_REGULAR_EXPRESSION, "$2");
}

export const renderInlineMarkup = createMemoizedTextTransform(renderInlineMarkupUncached);
export const stripMarkupFormatting = createMemoizedTextTransform(stripMarkupFormattingUncached);

export function renderBlockMarkup(rawText) {
  return String(rawText)
    .split("\n")
    .map((line) => "<p>" + (renderInlineMarkup(line) || "&nbsp;") + "</p>")
    .join("");
}

export function matchesTagFilter(tagText, query) {
  if (!query) {
    return true;
  }
  return (
    tagText.toLowerCase().includes(query) ||
    stripMarkupFormatting(tagText).toLowerCase().includes(stripMarkupFormatting(query).toLowerCase())
  );
}

export function normalizeTagQuery(rawQuery) {
  return rawQuery.trim().toLowerCase().replace(/^#/, "");
}
