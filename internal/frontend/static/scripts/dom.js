import { renderInlineMarkup } from "./markup.js";

const HASH_CHARACTER_CODE = 35;

export function querySelector(selector) {
  if (
    selector.charCodeAt(0) === HASH_CHARACTER_CODE &&
    !selector.includes(" ") &&
    !selector.includes(">") &&
    !selector.includes(".") &&
    !selector.includes(":")
  ) {
    return document.getElementById(selector.slice(1));
  }
  return document.querySelector(selector);
}

function applyAttribute(element, attributeKey, attributeValue) {
  if (attributeKey === "class") {
    element.className = attributeValue;
  } else if (attributeKey === "dataset") {
    Object.assign(element.dataset, attributeValue);
  } else if (attributeKey === "style") {
    Object.assign(element.style, attributeValue);
  } else if (attributeKey.startsWith("on")) {
    element.addEventListener(attributeKey.slice(2), attributeValue);
  } else if (attributeKey in element) {
    element[attributeKey] = attributeValue;
  } else {
    element.setAttribute(attributeKey, attributeValue);
  }
}

export function createElement(tagName, attributes = {}, ...children) {
  const element = document.createElement(tagName);
  for (const [attributeKey, attributeValue] of Object.entries(attributes)) {
    if (attributeValue == null) {
      continue;
    }
    applyAttribute(element, attributeKey, attributeValue);
  }
  for (const child of children.flat()) {
    if (child == null || child === false) {
      continue;
    }
    element.append(child.nodeType ? child : String(child));
  }
  return element;
}

export function createInlineElement(tagName, attributes, rawText) {
  const element = createElement(tagName, attributes);
  element.innerHTML = renderInlineMarkup(rawText);
  return element;
}

export function stopPropagationThen(callbackFunction) {
  return (eventObject) => {
    eventObject.stopPropagation();
    callbackFunction();
  };
}
