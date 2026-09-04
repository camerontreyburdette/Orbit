package discord

import (
	"sync"
	"time"

	"orbit/internal/configuration"
)

const (
	StatusConnected    = "connected"
	StatusConnecting   = "connecting"
	StatusDisconnected = "disconnected"
	defaultTheme       = "dark"
	synchronizePeriod  = 500 * time.Millisecond
)

type Client struct {
	mutex                sync.Mutex
	clientIdentifier     string
	configurationManager *configuration.Manager
	pipe                 *pipeConnection
	sessionStart         int64
	currentContext       Context
	stopChannel          chan struct{}
	wakeChannel          chan struct{}
	isRunning            bool
	isEnabled            bool
	theme                string
	status               string
}

func NewClient(clientIdentifier string, configurationManager *configuration.Manager) *Client {
	initialEnabled := true
	initialTheme := defaultTheme
	if configurationManager != nil {
		initialEnabled = configurationManager.IsDiscordRichPresenceEnabled()
		initialTheme = configurationManager.GetTheme()
	}

	return &Client{
		clientIdentifier:     clientIdentifier,
		configurationManager: configurationManager,
		pipe:                 newPipeConnection(),
		sessionStart:         time.Now().UnixMilli(),
		currentContext:       Context{View: "home"},
		stopChannel:          make(chan struct{}),
		wakeChannel:          make(chan struct{}, 10),
		isEnabled:            initialEnabled,
		theme:                initialTheme,
		status:               statusForEnabled(initialEnabled),
	}
}

func statusForEnabled(isEnabled bool) string {
	if isEnabled {
		return StatusConnecting
	}
	return StatusDisconnected
}

func (client *Client) wake() {
	select {
	case client.wakeChannel <- struct{}{}:
	default:
	}
}

func (client *Client) IsEnabled() bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.isEnabled
}

func (client *Client) applyEnabledStateLocked(enabled bool) {
	client.isEnabled = enabled
	client.status = statusForEnabled(enabled)
	if !enabled {
		client.pipe.close()
	}
}

func (client *Client) persistEnabledState(enabled bool) {
	if client.configurationManager != nil {
		_ = client.configurationManager.SetDiscordRichPresenceEnabled(enabled)
	}
	if enabled {
		client.wake()
	}
}

func (client *Client) SetEnabled(enabled bool) {
	client.mutex.Lock()
	if client.isEnabled == enabled {
		client.mutex.Unlock()
		return
	}
	client.applyEnabledStateLocked(enabled)
	client.mutex.Unlock()

	client.persistEnabledState(enabled)
}

func (client *Client) ToggleEnabled() bool {
	client.mutex.Lock()
	updatedEnabled := !client.isEnabled
	client.applyEnabledStateLocked(updatedEnabled)
	client.mutex.Unlock()

	client.persistEnabledState(updatedEnabled)
	return updatedEnabled
}

func (client *Client) GetTheme() string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.theme == "" {
		return defaultTheme
	}
	return client.theme
}

func (client *Client) SetTheme(theme string) {
	client.mutex.Lock()
	client.theme = theme
	client.mutex.Unlock()
	if client.configurationManager != nil {
		_ = client.configurationManager.SetTheme(theme)
	}
}

func (client *Client) Status() string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.status
}

func (client *Client) SetContext(context Context) {
	client.mutex.Lock()
	client.currentContext = context
	shouldWake := client.isRunning && client.isEnabled
	client.mutex.Unlock()

	if shouldWake {
		client.wake()
	}
}

func (client *Client) Start() {
	client.mutex.Lock()
	if client.isRunning {
		client.mutex.Unlock()
		return
	}
	client.isRunning = true
	client.mutex.Unlock()

	go client.runLoop()
}

func (client *Client) runLoop() {
	ticker := time.NewTicker(synchronizePeriod)
	defer ticker.Stop()

	client.synchronizeActivity()

	for {
		select {
		case <-client.stopChannel:
			client.pipe.close()
			return
		case <-ticker.C:
			client.synchronizeActivity()
		case <-client.wakeChannel:
			client.synchronizeActivity()
		}
	}
}

func (client *Client) synchronizeActivity() {
	client.mutex.Lock()
	if !client.isEnabled {
		client.status = StatusDisconnected
		client.mutex.Unlock()
		return
	}
	activity := client.BuildActivity(client.currentContext)
	clientIdentifier := client.clientIdentifier
	client.mutex.Unlock()

	sendError := client.pipe.sendActivity(clientIdentifier, activity)

	client.mutex.Lock()
	if !client.isEnabled || sendError != nil {
		client.status = StatusDisconnected
	} else {
		client.status = StatusConnected
	}
	client.mutex.Unlock()
}

func (client *Client) Close() {
	client.mutex.Lock()
	if !client.isRunning {
		client.mutex.Unlock()
		return
	}
	client.isRunning = false
	close(client.stopChannel)
	client.mutex.Unlock()

	client.pipe.close()
}
