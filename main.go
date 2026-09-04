package main

import (
	"fmt"
	"os"
	"path/filepath"

	"orbit/internal/api"
	"orbit/internal/assets"
	"orbit/internal/configuration"
	"orbit/internal/database"
	"orbit/internal/discord"
	"orbit/internal/frontend"
	"orbit/internal/server"
	"orbit/internal/window"
)

const (
	discordApplicationClientIdentifier = "1542920767556157571"
	applicationDirectoryName           = "Orbit"
	debugFlag                          = "--debug"
)

func resolveDataDirectory() string {
	userConfigurationDirectory, getUserConfigurationDirectoryError := os.UserConfigDir()
	if getUserConfigurationDirectoryError == nil && userConfigurationDirectory != "" {
		return filepath.Join(userConfigurationDirectory, applicationDirectoryName)
	}

	if appDataEnvironmentValue := os.Getenv("APPDATA"); appDataEnvironmentValue != "" {
		return filepath.Join(appDataEnvironmentValue, applicationDirectoryName)
	}

	return filepath.Join(".", applicationDirectoryName)
}

func isDebugModeRequested(arguments []string) bool {
	for _, argument := range arguments {
		if argument == debugFlag {
			return true
		}
	}
	return false
}

func exitWithError(message string, failure error) {
	fmt.Printf("%s: %v\n", message, failure)
	os.Exit(1)
}

func main() {
	dataDirectory := resolveDataDirectory()
	configurationManager := configuration.NewManager(filepath.Join(dataDirectory, "configuration.json"))

	storeInstance, storeError := database.NewStore(filepath.Join(dataDirectory, "boards"))
	if storeError != nil {
		exitWithError("failed to initialize store", storeError)
	}

	discordClient := discord.NewClient(discordApplicationClientIdentifier, configurationManager)
	discordClient.Start()
	defer discordClient.Close()

	staticFileSystem, staticFileSystemError := frontend.StaticFileSystem()
	if staticFileSystemError != nil {
		exitWithError("failed to prepare frontend", staticFileSystemError)
	}

	apiHandler := api.NewHandler(storeInstance, assets.EmbeddedFileSystem, "font")
	apiHandler.SetDiscordClient(discordClient)
	apiHandler.SetDataDirectory(dataDirectory)
	apiHandler.SetConfigurationResetter(configurationManager)
	apiHandler.SetTooltipPreferenceStore(configurationManager)
	apiHandler.StartBoardTimeTracking()
	defer apiHandler.Close()

	serverInstance, serverError := server.StartStaticServer(staticFileSystem, apiHandler)
	if serverError != nil {
		exitWithError("failed to start static server", serverError)
	}
	defer serverInstance.Close()

	serverUniformResourceLocator := fmt.Sprintf("%s/index.html", serverInstance.Address())
	if windowError := window.StartAppWindow(serverUniformResourceLocator, apiHandler, configurationManager, isDebugModeRequested(os.Args[1:])); windowError != nil {
		exitWithError("failed to run application window", windowError)
	}
}
