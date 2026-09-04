package api

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"

	"orbit/internal/database"
	"orbit/internal/discord"
)

const (
	maximumTitleCharacters = 120
	maximumItemCharacters  = 500
	maximumTagsCount       = 20
	maximumTagCharacters   = 30
)

type WindowController interface {
	SetTitle(title string)
	SetTheme(theme string)
	NativeWindowHandle() uintptr
}

type Handler struct {
	operationMutex         sync.Mutex
	store                  *database.Store
	windowController       WindowController
	discordClient          *discord.Client
	configurationResetter  ConfigurationResetter
	tooltipPreferenceStore TooltipPreferenceStore
	dataDirectory          string
	cachedFonts            []FontDescriptor
	boardTime              boardTimeTracker
}

func NewHandler(store *database.Store, fontFileSystem fs.FS, fontDirectory string) *Handler {
	return &Handler{
		store:       store,
		cachedFonts: LoadFontsFromFileSystem(fontFileSystem, fontDirectory),
	}
}

func (handler *Handler) SetWindowController(controller WindowController) {
	handler.windowController = controller
}

func (handler *Handler) SetDiscordClient(discordClient *discord.Client) {
	handler.discordClient = discordClient
}

func (handler *Handler) HandleMethodCall(method string, rawArguments []json.RawMessage) (interface{}, error) {
	handler.operationMutex.Lock()
	defer handler.operationMutex.Unlock()

	invoke, exists := methodRegistry[method]
	if !exists {
		return nil, fmt.Errorf("unknown api method: %s", method)
	}
	return invoke(handler, methodArguments(rawArguments))
}

func (handler *Handler) nativeWindowHandle() uintptr {
	if handler.windowController == nil {
		return 0
	}
	return handler.windowController.NativeWindowHandle()
}
