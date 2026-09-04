import { invokeBackendMethod } from "./backend.js";
import { createElement } from "./dom.js";
import { createIconElement } from "./icons.js";
import { confirmDestructiveAction, openModalOverlay } from "./overlays.js";
import { syncDiscordPresence } from "./presence.js";
import { setTooltipsEnabled } from "./tooltips.js";

const THEME_OPTIONS = [
  { name: "dark", label: "Dark" },
  { name: "light", label: "Light" },
];

async function loadCurrentSettings() {
  const settings = {
    isDiscordEnabled: true,
    isTooltipsEnabled: true,
    theme: document.documentElement.dataset.theme || "dark",
  };
  try {
    const currentSettings = await invokeBackendMethod("get_settings");
    if (currentSettings) {
      if (typeof currentSettings.discord_rich_presence_enabled === "boolean") {
        settings.isDiscordEnabled = currentSettings.discord_rich_presence_enabled;
      }
      if (typeof currentSettings.tooltips_enabled === "boolean") {
        settings.isTooltipsEnabled = currentSettings.tooltips_enabled;
      }
      if (currentSettings.theme) {
        settings.theme = currentSettings.theme;
        document.documentElement.dataset.theme = settings.theme;
      }
    }
  } catch {}
  return settings;
}

function createThemeSegmentControl(initialTheme) {
  let currentTheme = initialTheme;
  const segmentButtons = new Map();

  async function selectTheme(themeName) {
    if (currentTheme === themeName) {
      return;
    }
    currentTheme = themeName;
    document.documentElement.dataset.theme = themeName;
    for (const [buttonThemeName, buttonElement] of segmentButtons) {
      buttonElement.classList.toggle("active", buttonThemeName === themeName);
    }
    try {
      await invokeBackendMethod("set_theme", themeName);
    } catch (themeError) {
      console.error(themeError);
    }
  }

  for (const themeOption of THEME_OPTIONS) {
    segmentButtons.set(
      themeOption.name,
      createElement(
        "button",
        {
          class: "settings-segment-button" + (currentTheme === themeOption.name ? " active" : ""),
          onclick: () => selectTheme(themeOption.name),
        },
        themeOption.label
      )
    );
  }

  return createElement("div", { class: "settings-segmented" }, [...segmentButtons.values()]);
}

function createSettingsSwitch(initialEnabled, onToggle) {
  let isEnabled = initialEnabled;
  const switchButton = createElement(
    "button",
    {
      class: "settings-switch" + (isEnabled ? " active" : ""),
      onclick: () => {
        isEnabled = !isEnabled;
        switchButton.classList.toggle("active", isEnabled);
        onToggle(isEnabled);
      },
    },
    createElement("span", { class: "settings-switch-knob" })
  );
  return switchButton;
}

async function persistDiscordEnabled(isDiscordEnabled) {
  try {
    await invokeBackendMethod("set_discord_enabled", isDiscordEnabled);
  } catch (toggleError) {
    console.error(toggleError);
  }
  syncDiscordPresence();
}

async function persistTooltipsEnabled(isTooltipsEnabled) {
  setTooltipsEnabled(isTooltipsEnabled);
  try {
    await invokeBackendMethod("set_tooltips_enabled", isTooltipsEnabled);
  } catch (toggleError) {
    console.error(toggleError);
  }
}

function createResetConfirmationElement() {
  return createElement(
    "span",
    {},
    "Reset Orbit? This permanently deletes ",
    createElement("strong", {}, "every board, attachment and setting"),
    " stored on this computer."
  );
}

async function resetApplicationData() {
  const isConfirmed = await confirmDestructiveAction(createResetConfirmationElement(), "Reset");
  if (!isConfirmed) {
    return;
  }
  try {
    await invokeBackendMethod("reset_application_data");
  } catch (resetError) {
    console.error(resetError);
    return;
  }
  window.location.reload();
}

function createResetButton() {
  return createElement("button", { class: "button button-danger settings-reset-button", onclick: resetApplicationData }, "Reset");
}

function createSettingsRow(title, description, controlElement) {
  return createElement(
    "div",
    { class: "settings-row" },
    createElement(
      "div",
      { class: "settings-row-info" },
      createElement("span", { class: "settings-row-title" }, title),
      createElement("span", { class: "settings-row-description" }, description)
    ),
    controlElement
  );
}

export async function openSettingsModalDialog() {
  const settings = await loadCurrentSettings();

  const dialogBoxElement = createElement("div", { class: "confirm-box dialog-box settings-box" });
  const closeModal = openModalOverlay(dialogBoxElement);

  dialogBoxElement.append(
    createElement(
      "div",
      { class: "dialog-head" },
      createElement("h3", {}, "Settings"),
      createElement("button", { class: "icon-button", dataset: { tooltip: "Close" }, onclick: () => closeModal() }, createIconElement("x", 16))
    ),
    createElement(
      "div",
      { class: "settings-rows" },
      createSettingsRow("Theme", "Select appearance theme for the workspace", createThemeSegmentControl(settings.theme)),
      createSettingsRow(
        "Tooltips",
        "Show a short hint when hovering over buttons",
        createSettingsSwitch(settings.isTooltipsEnabled, persistTooltipsEnabled)
      ),
      createSettingsRow(
        "Discord Rich Presence",
        "Display your active board and card status on your Discord profile",
        createSettingsSwitch(settings.isDiscordEnabled, persistDiscordEnabled)
      ),
      createSettingsRow(
        "Reset Orbit",
        "Permanently delete all boards, attachments and settings stored on this computer",
        createResetButton()
      )
    )
  );
}
