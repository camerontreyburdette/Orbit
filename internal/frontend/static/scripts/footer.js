import { FOOTER_PHRASES } from "./constants.js";
import { querySelector } from "./dom.js";

export function applyRandomFooterPhrase() {
  querySelector("#footer-phrase").textContent = FOOTER_PHRASES[Math.floor(Math.random() * FOOTER_PHRASES.length)];
}
