import { COLOR_HEX_DEFINITIONS } from "./constants.js";

export const WRAP_MARKERS = {
  bold: "**",
  italic: "*",
  underline: "_",
  strikethrough: "-",
};

const COLOR_WRAPPER_REGULAR_EXPRESSION = new RegExp(
  `^\\[(${Object.keys(COLOR_HEX_DEFINITIONS).join("|")})\\s*:\\s*([\\s\\S]*)\\]$`,
  "i"
);

function countLeadingRun(text, character) {
  let count = 0;
  while (count < text.length && text[count] === character) {
    count++;
  }
  return count;
}

function countTrailingRun(text, character) {
  let count = 0;
  while (count < text.length && text[text.length - 1 - count] === character) {
    count++;
  }
  return count;
}

function isWrappedWithMarker(text, marker) {
  const markerCharacter = marker[0];
  const boundaryRun = Math.min(countLeadingRun(text, markerCharacter), countTrailingRun(text, markerCharacter));
  if (boundaryRun * 2 >= text.length) {
    return false;
  }
  if (marker.length === 1) {
    return boundaryRun % 2 === 1;
  }
  return boundaryRun >= marker.length;
}

export function toggleWrapMarker(selectedText, marker) {
  if (isWrappedWithMarker(selectedText, marker)) {
    return selectedText.slice(marker.length, selectedText.length - marker.length);
  }
  return marker + selectedText + marker;
}

export function parseColorWrapper(selectedText) {
  const match = COLOR_WRAPPER_REGULAR_EXPRESSION.exec(selectedText);
  if (!match) {
    return null;
  }
  return { colorName: match[1].toLowerCase(), innerText: match[2] };
}

export function toggleColorWrapper(selectedText, colorName) {
  const existingWrapper = parseColorWrapper(selectedText);
  const innerText = existingWrapper ? existingWrapper.innerText : selectedText;
  if (!colorName || (existingWrapper && existingWrapper.colorName === colorName)) {
    return innerText;
  }
  return `[${colorName}: ${innerText}]`;
}

export function selectedTextOf(inputElement) {
  return inputElement.value.slice(inputElement.selectionStart, inputElement.selectionEnd);
}

export function replaceSelectedText(inputElement, transformSelectedText) {
  const selectionStart = inputElement.selectionStart;
  const selectionEnd = inputElement.selectionEnd;
  const replacementText = transformSelectedText(inputElement.value.slice(selectionStart, selectionEnd));
  inputElement.setRangeText(replacementText, selectionStart, selectionEnd, "select");
  inputElement.dispatchEvent(new Event("input", { bubbles: true }));
}
