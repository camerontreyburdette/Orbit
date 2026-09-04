package api

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"orbit/internal/database"
)

type sortableSummary struct {
	summary database.BoardSummary
	sortKey string
}

func pinnedTimestamp(summary database.BoardSummary) string {
	if summary.PinnedAt == nil {
		return ""
	}
	return *summary.PinnedAt
}

func sortedSummaries(entries []sortableSummary, less func(first, second sortableSummary) bool) []database.BoardSummary {
	sort.SliceStable(entries, func(firstIndex, secondIndex int) bool {
		return less(entries[firstIndex], entries[secondIndex])
	})
	summaries := make([]database.BoardSummary, len(entries))
	for index, entry := range entries {
		summaries[index] = entry.summary
	}
	return summaries
}

func (handler *Handler) GetBoards() ([]database.BoardSummary, error) {
	boards := handler.store.Boards()
	pinnedEntries := make([]sortableSummary, 0, len(boards))
	standardEntries := make([]sortableSummary, 0, len(boards))

	for _, board := range boards {
		summary := board.Summary()
		if board.Pinned != 0 {
			pinnedEntries = append(pinnedEntries, sortableSummary{summary: summary, sortKey: pinnedTimestamp(summary)})
		} else {
			standardEntries = append(standardEntries, sortableSummary{summary: summary, sortKey: database.PlainName(board.Name)})
		}
	}

	pinnedSummaries := sortedSummaries(pinnedEntries, func(first, second sortableSummary) bool {
		return first.sortKey > second.sortKey
	})
	standardSummaries := sortedSummaries(standardEntries, func(first, second sortableSummary) bool {
		return first.sortKey < second.sortKey
	})

	return append(pinnedSummaries, standardSummaries...), nil
}

func (handler *Handler) CreateBoard(name string, description string) (response, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Untitled board"
	}

	board := &database.Board{
		Identifier:  handler.store.NewIdentifier(),
		Name:        name,
		Description: strings.TrimSpace(description),
		CreatedAt:   database.FormatTimestampNow(),
		Lists:       make([]*database.List, 0),
	}

	if addError := handler.store.AddBoard(board); addError != nil {
		return nil, fmt.Errorf("failed to add board: %w", addError)
	}
	_ = os.MkdirAll(handler.store.AttachmentDirectory(board), 0750)

	summaries, _ := handler.GetBoards()
	return response{"id": board.Identifier, "name": board.Name, "boards": summaries}, nil
}

func (handler *Handler) UpdateBoard(boardIdentifier int64, fields map[string]interface{}) (response, error) {
	updates, hasUpdates, parseError := parseBoardUpdates(fields)
	if parseError != nil {
		return nil, parseError
	}
	if !hasUpdates {
		return handler.boardListResponse(), nil
	}

	board := handler.store.Board(boardIdentifier)
	if board == nil {
		return nil, fmt.Errorf("Board not found")
	}

	updates.applyTo(board)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}

	return handler.boardListResponse(), nil
}

func (handler *Handler) DeleteBoard(boardIdentifier int64) (response, error) {
	board := handler.store.Board(boardIdentifier)
	if board == nil {
		return nil, fmt.Errorf("Board not found")
	}

	if removeError := handler.store.RemoveBoard(board); removeError != nil {
		return nil, fmt.Errorf("failed to delete board: %w", removeError)
	}

	return handler.boardListResponse(), nil
}

func (handler *Handler) GetBoard(boardIdentifier int64) (boardView, error) {
	board := handler.store.Board(boardIdentifier)
	if board == nil {
		return boardView{}, fmt.Errorf("Board not found")
	}
	return buildBoardView(board), nil
}

func parseSnapshotIdentifier(document map[string]interface{}) (int64, bool) {
	rawIdentifier, hasIdentifier := document["id"]
	if !hasIdentifier {
		return 0, false
	}
	switch value := rawIdentifier.(type) {
	case float64:
		return int64(value), true
	case int64:
		return value, true
	}
	return 0, true
}

func decodeBoardSnapshot(document map[string]interface{}) (*database.Board, error) {
	documentBytes, marshalError := json.Marshal(document)
	if marshalError != nil {
		return nil, fmt.Errorf("Invalid snapshot: %w", marshalError)
	}

	var restoredBoard database.Board
	if unmarshalError := json.Unmarshal(documentBytes, &restoredBoard); unmarshalError != nil {
		return nil, fmt.Errorf("Invalid snapshot: %w", unmarshalError)
	}
	return &restoredBoard, nil
}

func (handler *Handler) RestoreBoard(boardIdentifier int64, document map[string]interface{}) (response, error) {
	snapshotIdentifier, hasIdentifier := parseSnapshotIdentifier(document)
	if !hasIdentifier || snapshotIdentifier != boardIdentifier {
		return nil, fmt.Errorf("Invalid snapshot")
	}

	restoredBoard, decodeError := decodeBoardSnapshot(document)
	if decodeError != nil {
		return nil, decodeError
	}

	board := handler.store.Board(boardIdentifier)
	if board == nil {
		return nil, fmt.Errorf("Board not found")
	}

	board.Name = restoredBoard.Name
	board.Description = restoredBoard.Description
	board.Pinned = restoredBoard.Pinned
	board.PinnedAt = restoredBoard.PinnedAt
	board.CreatedAt = restoredBoard.CreatedAt
	board.Lists = restoredBoard.Lists

	handler.reconcileAttachments(board)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save restored board: %w", saveError)
	}

	return okResponse(), nil
}
