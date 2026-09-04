package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"orbit/internal/database"
)

const (
	maximumAttachmentPathSegments = 8
	illegalPathSegmentCharacters  = `\/:*?"<>|`
)

type attachmentStoreFunction = func(relativePath string, content io.Reader) error

type boardImportSession struct {
	handler             *Handler
	board               *database.Board
	attachmentDirectory string
}

func decodeImportedBoard(documentBytes []byte) (*database.Board, error) {
	var importedBoard database.Board
	if unmarshalError := json.Unmarshal(documentBytes, &importedBoard); unmarshalError != nil {
		return nil, fmt.Errorf("Invalid board file: %w", unmarshalError)
	}
	if importedBoard.Lists == nil {
		return nil, fmt.Errorf("Invalid board file: no lists found")
	}
	return &importedBoard, nil
}

func validateAttachmentRelativePath(relativePath string) ([]string, error) {
	segments := strings.Split(relativePath, "/")
	if len(segments) == 0 || len(segments) > maximumAttachmentPathSegments {
		return nil, fmt.Errorf("Invalid attachment path: %s", relativePath)
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." || strings.ContainsAny(segment, illegalPathSegmentCharacters) {
			return nil, fmt.Errorf("Invalid attachment path: %s", relativePath)
		}
	}
	return segments, nil
}

func (handler *Handler) beginBoardImportLocked(documentBytes []byte) (*boardImportSession, error) {
	board, decodeError := decodeImportedBoard(documentBytes)
	if decodeError != nil {
		return nil, decodeError
	}
	normalizeImportedBoard(board)
	handler.assignFreshBoardIdentifiers(board)

	if addError := handler.store.AddBoard(board); addError != nil {
		return nil, fmt.Errorf("failed to add imported board: %w", addError)
	}

	attachmentDirectory := handler.store.AttachmentDirectory(board)
	if mkdirError := os.MkdirAll(attachmentDirectory, 0750); mkdirError != nil {
		return nil, fmt.Errorf("failed to create attachment directory: %w", mkdirError)
	}

	return &boardImportSession{handler: handler, board: board, attachmentDirectory: attachmentDirectory}, nil
}

func (session *boardImportSession) storeAttachmentFile(relativePath string, content io.Reader) error {
	segments, validationError := validateAttachmentRelativePath(relativePath)
	if validationError != nil {
		return validationError
	}

	destinationPath := filepath.Join(append([]string{session.attachmentDirectory}, segments...)...)
	if mkdirError := os.MkdirAll(filepath.Dir(destinationPath), 0750); mkdirError != nil {
		return fmt.Errorf("failed to create attachment folder: %w", mkdirError)
	}
	return writeReaderToFile(destinationPath, content)
}

func (session *boardImportSession) finishLocked() (response, error) {
	handler := session.handler
	handler.reconcileAttachments(session.board)
	pruneMissingAttachments(session.board, session.attachmentDirectory)
	if saveError := handler.store.SaveBoard(session.board); saveError != nil {
		return nil, fmt.Errorf("failed to save imported board: %w", saveError)
	}

	importResponse := handler.boardListResponse()
	importResponse["imported"] = response{"id": session.board.Identifier, "name": session.board.Name}
	return importResponse, nil
}

func (session *boardImportSession) abortLocked() {
	_ = session.handler.store.RemoveBoard(session.board)
}

func (session *boardImportSession) completeLocked(streamAttachments func(storeAttachment attachmentStoreFunction) error) (response, error) {
	if streamError := streamAttachments(session.storeAttachmentFile); streamError != nil {
		session.abortLocked()
		return nil, streamError
	}
	return session.finishLocked()
}

func (handler *Handler) ImportBoard(
	documentBytes []byte,
	streamAttachments func(storeAttachment attachmentStoreFunction) error,
) (map[string]interface{}, error) {
	handler.operationMutex.Lock()
	session, beginError := handler.beginBoardImportLocked(documentBytes)
	handler.operationMutex.Unlock()
	if beginError != nil {
		return nil, beginError
	}

	if streamError := streamAttachments(session.storeAttachmentFile); streamError != nil {
		handler.operationMutex.Lock()
		session.abortLocked()
		handler.operationMutex.Unlock()
		return nil, streamError
	}

	handler.operationMutex.Lock()
	defer handler.operationMutex.Unlock()
	return session.finishLocked()
}
