import { createElement } from "./dom.js";

export function attachmentUniformResourceLocator(attachmentIdentifier) {
  return window.location.origin + "/attachments/" + attachmentIdentifier;
}

export function attachmentUploadUniformResourceLocator(cardIdentifier) {
  return window.location.origin + "/cards/" + cardIdentifier + "/attachments";
}

export function createAttachmentImageElement(attachment, attributes, shouldRemoveOnError) {
  return createElement("img", {
    ...attributes,
    src: attachmentUniformResourceLocator(attachment.id),
    onerror: shouldRemoveOnError ? (errorEvent) => errorEvent.target.remove() : null,
  });
}
