package api

import (
	"orbit/internal/database"
	"orbit/internal/discord"
)

const (
	maximumWindowTitleCharacters = 200
	defaultTheme                 = "dark"
	defaultTooltipsEnabled       = true
)

type TooltipPreferenceStore interface {
	IsTooltipsEnabled() bool
	SetTooltipsEnabled(enabled bool) error
}

func (handler *Handler) SetTooltipPreferenceStore(tooltipPreferenceStore TooltipPreferenceStore) {
	handler.tooltipPreferenceStore = tooltipPreferenceStore
}

func (handler *Handler) isTooltipsEnabled() bool {
	if handler.tooltipPreferenceStore == nil {
		return defaultTooltipsEnabled
	}
	return handler.tooltipPreferenceStore.IsTooltipsEnabled()
}

func (handler *Handler) SetTooltipsEnabled(enabled bool) (response, error) {
	if handler.tooltipPreferenceStore != nil {
		if saveError := handler.tooltipPreferenceStore.SetTooltipsEnabled(enabled); saveError != nil {
			return nil, saveError
		}
	}
	return response{"ok": true, "tooltips_enabled": enabled}, nil
}

func (handler *Handler) SetTitle(title string) (response, error) {
	title = database.TruncateRunes(title, maximumWindowTitleCharacters)
	if handler.windowController != nil {
		handler.windowController.SetTitle(title)
	}
	return okResponse(), nil
}

func (handler *Handler) GetFonts() ([]FontDescriptor, error) {
	result := make([]FontDescriptor, len(handler.cachedFonts))
	copy(result, handler.cachedFonts)
	return result, nil
}

func (handler *Handler) discordStatusResponse() response {
	if handler.discordClient == nil {
		return response{"enabled": false, "status": discord.StatusDisconnected}
	}
	return response{"enabled": handler.discordClient.IsEnabled(), "status": handler.discordClient.Status()}
}

func (handler *Handler) GetDiscordStatus() (response, error) {
	return handler.discordStatusResponse(), nil
}

func (handler *Handler) GetSettings() (response, error) {
	if handler.discordClient == nil {
		return response{
			"discord_rich_presence_enabled": false,
			"discord_status":                discord.StatusDisconnected,
			"theme":                         defaultTheme,
			"tooltips_enabled":              handler.isTooltipsEnabled(),
		}, nil
	}
	return response{
		"discord_rich_presence_enabled": handler.discordClient.IsEnabled(),
		"discord_status":                handler.discordClient.Status(),
		"theme":                         handler.discordClient.GetTheme(),
		"tooltips_enabled":              handler.isTooltipsEnabled(),
	}, nil
}

func (handler *Handler) GetTheme() string {
	if handler.discordClient != nil {
		return handler.discordClient.GetTheme()
	}
	return defaultTheme
}

func (handler *Handler) SetTheme(theme string) (response, error) {
	if theme != "light" && theme != defaultTheme {
		theme = defaultTheme
	}
	if handler.discordClient != nil {
		handler.discordClient.SetTheme(theme)
	}
	if handler.windowController != nil {
		handler.windowController.SetTheme(theme)
	}
	return response{"ok": true, "theme": theme}, nil
}

func (handler *Handler) SetDiscordEnabled(enabled bool) (response, error) {
	if handler.discordClient != nil {
		handler.discordClient.SetEnabled(enabled)
	}
	return handler.discordStatusResponse(), nil
}

func (handler *Handler) ToggleDiscordRPC() (response, error) {
	if handler.discordClient != nil {
		handler.discordClient.ToggleEnabled()
	}
	return handler.discordStatusResponse(), nil
}

func (handler *Handler) SetPresenceContext(presenceContext discord.Context) (response, error) {
	if handler.discordClient != nil {
		handler.discordClient.SetContext(presenceContext)
	}
	statusResponse := handler.discordStatusResponse()
	statusResponse["ok"] = true
	return statusResponse, nil
}
