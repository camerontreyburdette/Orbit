package api

import (
	"fmt"
	"strings"

	"orbit/internal/database"
)

type boardUpdates struct {
	name        *string
	description *string
	pinned      *bool
}

func parseBoardUpdates(fields map[string]interface{}) (boardUpdates, bool, error) {
	var updates boardUpdates
	hasUpdates := false

	if rawName, hasName := fields["name"]; hasName {
		hasUpdates = true
		name := strings.TrimSpace(fmt.Sprintf("%v", rawName))
		if name == "" {
			return updates, hasUpdates, fmt.Errorf("Board name cannot be empty")
		}
		updates.name = &name
	}

	if rawDescription, hasDescription := fields["description"]; hasDescription {
		hasUpdates = true
		description := strings.TrimSpace(fmt.Sprintf("%v", rawDescription))
		updates.description = &description
	}

	if rawPinned, hasPinned := fields["pinned"]; hasPinned {
		hasUpdates = true
		pinned := parsePinnedFlag(rawPinned)
		updates.pinned = &pinned
	}

	return updates, hasUpdates, nil
}

func parsePinnedFlag(rawPinned interface{}) bool {
	switch value := rawPinned.(type) {
	case bool:
		return value
	case float64:
		return value != 0
	case int:
		return value != 0
	}
	return false
}

func (updates boardUpdates) applyTo(board *database.Board) {
	if updates.name != nil {
		board.Name = *updates.name
	}
	if updates.description != nil {
		board.Description = *updates.description
	}
	if updates.pinned != nil {
		applyPinnedState(board, *updates.pinned)
	}
}

func applyPinnedState(board *database.Board, isPinned bool) {
	newPinned := 0
	if isPinned {
		newPinned = 1
	}
	if newPinned == board.Pinned {
		return
	}
	board.Pinned = newPinned
	if isPinned {
		timestamp := database.FormatTimestampNow()
		board.PinnedAt = &timestamp
		return
	}
	board.PinnedAt = nil
}
