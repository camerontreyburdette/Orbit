package api

import (
	"fmt"

	"orbit/internal/database"
)

type response = map[string]interface{}

func okResponse() response {
	return response{"ok": true}
}

func okResponseWithCreatedIdentifiers(createdIdentifiers []int64) response {
	return response{"ok": true, "created_identifiers": createdIdentifiers}
}

func (handler *Handler) boardListResponse() response {
	summaries, _ := handler.GetBoards()
	return response{"ok": true, "boards": summaries}
}

func (handler *Handler) saveBoards(boards map[int64]*database.Board) error {
	var firstError error
	for _, board := range boards {
		if saveError := handler.store.SaveBoard(board); saveError != nil && firstError == nil {
			firstError = fmt.Errorf("failed to save board: %w", saveError)
		}
	}
	return firstError
}

func (handler *Handler) saveBoardOrFail(board *database.Board) (response, error) {
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}
	return okResponse(), nil
}
