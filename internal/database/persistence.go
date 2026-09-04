package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	boardFileName             = "board.json"
	identifierCounterFileName = "identifiers.json"
)

type identifierCounter struct {
	NextIdentifier int64 `json:"next_identifier"`
}

func writeFileSynchronized(path string, data []byte) error {
	file, openError := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if openError != nil {
		return openError
	}
	if _, writeError := file.Write(data); writeError != nil {
		_ = file.Close()
		return writeError
	}
	if syncError := file.Sync(); syncError != nil {
		_ = file.Close()
		return syncError
	}
	return file.Close()
}

func AtomicWriteJSON(targetPath string, payload interface{}) error {
	temporaryPath := targetPath + ".temporary"
	encodedBytes, marshalError := json.MarshalIndent(payload, "", "  ")
	if marshalError != nil {
		return fmt.Errorf("failed to marshal json payload: %w", marshalError)
	}

	encodedBytes = append(encodedBytes, '\n')
	if writeError := writeFileSynchronized(temporaryPath, encodedBytes); writeError != nil {
		return fmt.Errorf("failed to write temporary file: %w", writeError)
	}

	if renameError := os.Rename(temporaryPath, targetPath); renameError != nil {
		if removeError := os.Remove(temporaryPath); removeError != nil {
			return fmt.Errorf("failed to remove temporary file after rename error: %w", renameError)
		}
		return fmt.Errorf("failed to replace destination file: %w", renameError)
	}

	return nil
}

func readIdentifierCounter(boardsDirectory string) int64 {
	fileBytes, readError := os.ReadFile(filepath.Join(boardsDirectory, identifierCounterFileName))
	if readError != nil {
		return 0
	}
	var counter identifierCounter
	if unmarshalError := json.Unmarshal(fileBytes, &counter); unmarshalError != nil {
		return 0
	}
	return counter.NextIdentifier
}

func writeIdentifierCounter(boardsDirectory string, nextIdentifier int64) error {
	return AtomicWriteJSON(filepath.Join(boardsDirectory, identifierCounterFileName), identifierCounter{NextIdentifier: nextIdentifier})
}

func listDirectoryNames(boardsDirectory string) map[string]struct{} {
	names := make(map[string]struct{})
	entries, readError := os.ReadDir(boardsDirectory)
	if readError != nil {
		return names
	}
	for _, entry := range entries {
		if entry.IsDir() {
			names[entry.Name()] = struct{}{}
		}
	}
	return names
}

type loadedBoard struct {
	board         *Board
	directoryName string
}

func readBoardFile(boardsDirectory string, directoryName string) (*Board, bool) {
	fileBytes, readFileError := os.ReadFile(filepath.Join(boardsDirectory, directoryName, boardFileName))
	if readFileError != nil {
		return nil, false
	}

	var board Board
	if unmarshalError := json.Unmarshal(fileBytes, &board); unmarshalError != nil {
		return nil, false
	}

	if board.Lists == nil {
		board.Lists = make([]*List, 0)
	}
	return &board, true
}

func loadBoardsFromDirectory(boardsDirectory string) ([]loadedBoard, error) {
	entries, readError := os.ReadDir(boardsDirectory)
	if readError != nil {
		return nil, fmt.Errorf("failed to read boards directory: %w", readError)
	}

	sort.Slice(entries, func(firstIndex, secondIndex int) bool {
		return entries[firstIndex].Name() < entries[secondIndex].Name()
	})

	loaded := make([]loadedBoard, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		board, isValid := readBoardFile(boardsDirectory, entry.Name())
		if !isValid {
			continue
		}
		loaded = append(loaded, loadedBoard{board: board, directoryName: entry.Name()})
	}
	return loaded, nil
}
