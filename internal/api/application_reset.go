package api

import (
	"fmt"
	"os"
	"path/filepath"
)

const webViewProfileDirectoryName = "EBWebView"

type ConfigurationResetter interface {
	ResetToDefaults() error
}

func (handler *Handler) SetDataDirectory(dataDirectory string) {
	handler.dataDirectory = dataDirectory
}

func (handler *Handler) SetConfigurationResetter(configurationResetter ConfigurationResetter) {
	handler.configurationResetter = configurationResetter
}

func (handler *Handler) preservedDataDirectoryNames() map[string]struct{} {
	return map[string]struct{}{
		filepath.Base(handler.store.Directory()): {},
		webViewProfileDirectoryName:              {},
	}
}

func removeDataDirectoryEntries(dataDirectory string, preservedNames map[string]struct{}) error {
	entries, readError := os.ReadDir(dataDirectory)
	if readError != nil {
		return fmt.Errorf("failed to read data directory: %w", readError)
	}

	var firstError error
	for _, entry := range entries {
		if _, isPreserved := preservedNames[entry.Name()]; isPreserved {
			continue
		}
		if removeError := os.RemoveAll(filepath.Join(dataDirectory, entry.Name())); removeError != nil && firstError == nil {
			firstError = fmt.Errorf("failed to remove %s: %w", entry.Name(), removeError)
		}
	}
	return firstError
}

func (handler *Handler) restoreDefaultSettings() {
	if handler.configurationResetter != nil {
		_ = handler.configurationResetter.ResetToDefaults()
	}
	if handler.discordClient != nil {
		handler.discordClient.SetTheme(defaultTheme)
		handler.discordClient.SetEnabled(true)
	}
	if handler.windowController != nil {
		handler.windowController.SetTheme(defaultTheme)
	}
}

func (handler *Handler) ResetApplicationData() (response, error) {
	handler.boardTime.activeBoardIdentifier = 0

	if removeError := handler.store.RemoveAllBoards(); removeError != nil {
		return nil, fmt.Errorf("failed to remove boards: %w", removeError)
	}

	if handler.dataDirectory != "" {
		if removeError := removeDataDirectoryEntries(handler.dataDirectory, handler.preservedDataDirectoryNames()); removeError != nil {
			return nil, removeError
		}
	}

	handler.restoreDefaultSettings()
	return okResponse(), nil
}
