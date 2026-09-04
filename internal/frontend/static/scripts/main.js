import { BACKEND_AVAILABILITY_TIMEOUT_MILLISECONDS, PRESENCE_SYNCHRONIZATION_INTERVAL_MILLISECONDS } from "./constants.js";
import { invokeBackendMethod, isBackendAvailable } from "./backend.js";
import { querySelector } from "./dom.js";
import { refreshBoardsList } from "./board_actions.js";
import { renderMainApplicationView } from "./render.js";
import { syncDiscordPresence } from "./presence.js";
import { loadCustomFonts } from "./fonts.js";
import { applyRandomFooterPhrase } from "./footer.js";
import { installDocumentFileDropGuards } from "./file_drop.js";
import { installGlobalKeyboardShortcuts } from "./keyboard_shortcuts.js";
import { installBoardSpaceWheelScrolling } from "./board_scrolling.js";
import { installFormattingContextMenu } from "./formatting_menu.js";
import { installButtonTooltips, setTooltipsEnabled } from "./tooltips.js";

let isApplicationStarted = false;

async function applyPersistedSettings() {
  try {
    const settings = await invokeBackendMethod("get_settings");
    if (!settings) {
      return;
    }
    if (settings.theme) {
      document.documentElement.dataset.theme = settings.theme;
    }
    if (typeof settings.tooltips_enabled === "boolean") {
      setTooltipsEnabled(settings.tooltips_enabled);
    }
  } catch {}
}

async function startApplication() {
  if (isApplicationStarted || !isBackendAvailable()) {
    return;
  }
  isApplicationStarted = true;
  loadCustomFonts();
  await applyPersistedSettings();
  try {
    await refreshBoardsList();
  } catch {
    isApplicationStarted = false;
    return;
  }
  renderMainApplicationView();
  setInterval(syncDiscordPresence, PRESENCE_SYNCHRONIZATION_INTERVAL_MILLISECONDS);
}

function showBackendUnavailableMessage() {
  if (!isApplicationStarted && !window.pywebview) {
    querySelector("#board-root").innerHTML =
      '<div class="empty-state"><p>Backend not available — launch Orbit via executable.</p></div>';
  }
}

installDocumentFileDropGuards();
installGlobalKeyboardShortcuts();
installBoardSpaceWheelScrolling();
installFormattingContextMenu();
installButtonTooltips();
applyRandomFooterPhrase();

window.addEventListener("pywebviewready", startApplication);
document.addEventListener("DOMContentLoaded", () => startApplication());
setTimeout(showBackendUnavailableMessage, BACKEND_AVAILABILITY_TIMEOUT_MILLISECONDS);
