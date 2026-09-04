import { createElement, querySelector } from "./dom.js";
import { formatByteSize } from "./formatting.js";
import { attachmentUniformResourceLocator } from "./attachments.js";
import { bindOverlayCloseHandler, closeOverlayElement } from "./overlays.js";
import { deleteAttachmentWithConfirmation, saveAttachmentAs } from "./attachment_section.js";

export function isLightboxModalOpen() {
  const overlayElement = querySelector("#lightbox-root").firstElementChild;
  return Boolean(overlayElement) && !overlayElement.classList.contains("closing");
}

export function closeLightboxModal() {
  closeOverlayElement(querySelector("#lightbox-root"));
}

function createMediaElement(attachment) {
  const sourceUniformResourceLocator = attachmentUniformResourceLocator(attachment.id);
  if (attachment.kind === "image") {
    return createElement("img", { src: sourceUniformResourceLocator, alt: attachment.name });
  }
  const mediaTagName = attachment.kind === "video" ? "video" : "audio";
  return createElement(mediaTagName, { src: sourceUniformResourceLocator, controls: true, autoplay: true });
}

async function deleteFromLightbox(attachment) {
  const isDeleted = await deleteAttachmentWithConfirmation(attachment);
  if (isDeleted) {
    closeLightboxModal();
  }
}

function createLightboxActions(attachment) {
  return createElement(
    "div",
    { class: "lightbox-actions" },
    createElement("button", { class: "button", onclick: () => saveAttachmentAs(attachment) }, "Save"),
    createElement("button", { class: "button button-danger", onclick: () => deleteFromLightbox(attachment) }, "Delete"),
    createElement("button", { class: "button", onclick: closeLightboxModal }, "Close")
  );
}

export function openLightboxModal(attachment) {
  const rootElement = querySelector("#lightbox-root");
  rootElement.innerHTML = "";
  const boxElement = createElement(
    "div",
    { class: "lightbox" },
    createMediaElement(attachment),
    createElement("div", { class: "lightbox-name" }, `${attachment.name} · ${formatByteSize(attachment.size)}`),
    createLightboxActions(attachment)
  );
  const overlayElement = createElement("div", { class: "overlay" }, boxElement);
  bindOverlayCloseHandler(overlayElement, closeLightboxModal);
  rootElement.append(overlayElement);
}
