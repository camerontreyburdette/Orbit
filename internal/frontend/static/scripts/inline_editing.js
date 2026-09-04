import { createElement } from "./dom.js";

export function focusInputElementAtEnd(inputElement) {
  inputElement.focus();
  const textLength = inputElement.value.length;
  try {
    inputElement.setSelectionRange(textLength, textLength);
  } catch {}
}

export function scheduleFocusAtEnd(inputElement) {
  setTimeout(() => focusInputElementAtEnd(inputElement), 0);
}

export function scheduleFocus(inputElement) {
  setTimeout(() => inputElement.focus(), 0);
}

export function isInputElementFocused() {
  const activeElement = document.activeElement;
  return Boolean(
    activeElement &&
      (activeElement.tagName === "INPUT" || activeElement.tagName === "TEXTAREA" || activeElement.isContentEditable)
  );
}

function shouldCommitOnEnter(keyboardEvent, commitsOnShiftEnter) {
  if (keyboardEvent.key !== "Enter") {
    return false;
  }
  return commitsOnShiftEnter || !keyboardEvent.shiftKey;
}

export function createInlineEditInput({
  tagName = "input",
  className,
  value,
  placeholder,
  rows,
  onInput,
  onCommit,
  onCancel,
  shouldStopEscapePropagation = true,
  commitsOnShiftEnter = true,
}) {
  return createElement(tagName, {
    class: className,
    rows,
    value,
    placeholder,
    oninput: onInput,
    onkeydown: (keyboardEvent) => {
      if (shouldCommitOnEnter(keyboardEvent, commitsOnShiftEnter)) {
        if (!commitsOnShiftEnter) {
          keyboardEvent.preventDefault();
        }
        keyboardEvent.target.blur();
      }
      if (keyboardEvent.key === "Escape") {
        if (shouldStopEscapePropagation) {
          keyboardEvent.stopPropagation();
        }
        onCancel(keyboardEvent);
      }
    },
    onblur: (blurEvent) => onCommit(blurEvent.target.value, blurEvent),
  });
}
