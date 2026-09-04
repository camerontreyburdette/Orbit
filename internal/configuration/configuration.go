package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const defaultTheme = "dark"

type WindowState struct {
	IsMaximized          bool  `json:"is_maximized"`
	HorizontalCoordinate int32 `json:"horizontal_coordinate"`
	VerticalCoordinate   int32 `json:"vertical_coordinate"`
	Width                int32 `json:"width"`
	Height               int32 `json:"height"`
}

type Configuration struct {
	DiscordRichPresenceEnabled bool         `json:"discord_rich_presence_enabled"`
	TooltipsEnabled            bool         `json:"tooltips_enabled"`
	Theme                      string       `json:"theme"`
	WindowState                *WindowState `json:"window_state,omitempty"`
}

type Manager struct {
	mutex                 sync.RWMutex
	configurationFilePath string
	currentConfiguration  Configuration
}

func defaultConfiguration() Configuration {
	return Configuration{
		DiscordRichPresenceEnabled: true,
		TooltipsEnabled:            true,
		Theme:                      defaultTheme,
	}
}

func NewManager(configurationFilePath string) *Manager {
	manager := &Manager{
		configurationFilePath: configurationFilePath,
		currentConfiguration:  defaultConfiguration(),
	}
	manager.loadFromFile()
	return manager
}

func (manager *Manager) loadFromFile() {
	if manager.configurationFilePath == "" {
		return
	}

	fileBytes, readFileError := os.ReadFile(manager.configurationFilePath)
	if readFileError != nil {
		_ = manager.saveToFile()
		return
	}

	parsedConfiguration := defaultConfiguration()
	if unmarshalError := json.Unmarshal(fileBytes, &parsedConfiguration); unmarshalError != nil {
		return
	}

	if parsedConfiguration.Theme == "" {
		parsedConfiguration.Theme = defaultTheme
	}
	manager.currentConfiguration = parsedConfiguration
}

func (manager *Manager) saveToFile() error {
	if manager.configurationFilePath == "" {
		return nil
	}

	encodedBytes, marshalError := json.MarshalIndent(manager.currentConfiguration, "", "  ")
	if marshalError != nil {
		return marshalError
	}

	if makeDirectoryError := os.MkdirAll(filepath.Dir(manager.configurationFilePath), 0750); makeDirectoryError != nil {
		return makeDirectoryError
	}

	temporaryFilePath := fmt.Sprintf("%s.%d.temporary", manager.configurationFilePath, time.Now().UnixNano())
	if writeError := os.WriteFile(temporaryFilePath, encodedBytes, 0600); writeError != nil {
		return writeError
	}

	if renameError := os.Rename(temporaryFilePath, manager.configurationFilePath); renameError != nil {
		_ = os.Remove(temporaryFilePath)
		return renameError
	}

	return nil
}

func (manager *Manager) update(mutate func(configuration *Configuration)) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	mutate(&manager.currentConfiguration)
	return manager.saveToFile()
}

func (manager *Manager) ResetToDefaults() error {
	return manager.update(func(configuration *Configuration) {
		*configuration = defaultConfiguration()
	})
}

func (manager *Manager) GetConfiguration() Configuration {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.currentConfiguration
}

func (manager *Manager) GetTheme() string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	if manager.currentConfiguration.Theme == "" {
		return defaultTheme
	}
	return manager.currentConfiguration.Theme
}

func (manager *Manager) SetTheme(theme string) error {
	return manager.update(func(configuration *Configuration) {
		configuration.Theme = theme
	})
}

func (manager *Manager) IsDiscordRichPresenceEnabled() bool {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.currentConfiguration.DiscordRichPresenceEnabled
}

func (manager *Manager) SetDiscordRichPresenceEnabled(enabled bool) error {
	return manager.update(func(configuration *Configuration) {
		configuration.DiscordRichPresenceEnabled = enabled
	})
}

func (manager *Manager) IsTooltipsEnabled() bool {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.currentConfiguration.TooltipsEnabled
}

func (manager *Manager) SetTooltipsEnabled(enabled bool) error {
	return manager.update(func(configuration *Configuration) {
		configuration.TooltipsEnabled = enabled
	})
}

func (manager *Manager) GetWindowState() *WindowState {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	if manager.currentConfiguration.WindowState == nil {
		return nil
	}
	windowStateCopy := *manager.currentConfiguration.WindowState
	return &windowStateCopy
}

func (manager *Manager) SetWindowState(windowState WindowState) error {
	return manager.update(func(configuration *Configuration) {
		configuration.WindowState = &windowState
	})
}
