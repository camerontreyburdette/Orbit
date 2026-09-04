//go:build !windows

package discord

type pipeConnection struct{}

func newPipeConnection() *pipeConnection {
	return &pipeConnection{}
}

func (connection *pipeConnection) sendActivity(clientIdentifier string, activity Activity) error {
	return nil
}

func (connection *pipeConnection) clearActivity() {}

func (connection *pipeConnection) close() {}
