//go:build !windows

package window

import (
	"fmt"

	"orbit/internal/api"
	"orbit/internal/configuration"
)

func StartAppWindow(serverUniformResourceLocator string, handler *api.Handler, configurationManager *configuration.Manager, debugMode bool) error {
	return fmt.Errorf("platform not supported")
}
