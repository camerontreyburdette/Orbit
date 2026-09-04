package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"orbit/internal/database"
)

func buildAttachmentFileIndex(rootDirectory string) map[string]string {
	fileIndex := make(map[string]string)
	_ = filepath.Walk(rootDirectory, func(currentPath string, fileInformation os.FileInfo, walkError error) error {
		if walkError != nil || fileInformation.IsDir() {
			return nil
		}
		if relativeFilePath, relativePathError := filepath.Rel(rootDirectory, currentPath); relativePathError == nil {
			fileIndex[fileInformation.Name()] = filepath.ToSlash(relativeFilePath)
		}
		return nil
	})
	return fileIndex
}

func (handler *Handler) reconcileAttachments(board *database.Board) {
	rootDirectory := handler.store.AttachmentDirectory(board)
	var fileIndex map[string]string

	for _, list := range board.Lists {
		for _, card := range list.Cards {
			for _, attachment := range card.Attachments {
				if _, statError := os.Stat(filepath.Join(rootDirectory, filepath.FromSlash(attachment.File))); statError == nil {
					continue
				}
				if fileIndex == nil {
					fileIndex = buildAttachmentFileIndex(rootDirectory)
				}
				storedFilename := attachment.File[strings.LastIndex(attachment.File, "/")+1:]
				if matchedPath, exists := fileIndex[storedFilename]; exists {
					attachment.File = matchedPath
				}
			}
			handler.synchronizeCardDirectory(board, card)
		}
	}
}

func (handler *Handler) storeAttachment(cardIdentifier int64, originalName string, writeFile func(destinationPath string) error) (response, error) {
	board, _, card := handler.store.FindCard(cardIdentifier)
	if card == nil {
		return nil, fmt.Errorf("Card not found")
	}

	rootDirectory := handler.store.AttachmentDirectory(board)
	folder := database.CardAttachmentFolder(card)
	if folder == "" {
		folder = handler.resolveCardDirectoryName(board, card, rootDirectory, "")
	}

	cardDirectory := filepath.Join(rootDirectory, folder)
	if mkdirError := os.MkdirAll(cardDirectory, 0750); mkdirError != nil {
		return nil, fmt.Errorf("failed to create attachment directory: %w", mkdirError)
	}

	storedFilename := generateStoredFilename(originalName)
	destinationPath := filepath.Join(cardDirectory, storedFilename)
	if writeError := writeFile(destinationPath); writeError != nil {
		return nil, fmt.Errorf("failed to write attachment file: %w", writeError)
	}

	fileInformation, statError := os.Stat(destinationPath)
	if statError != nil {
		return nil, fmt.Errorf("failed to stat attachment file: %w", statError)
	}

	attachment := &database.Attachment{
		Identifier: handler.store.NewIdentifier(),
		Name:       originalName,
		File:       folder + "/" + storedFilename,
		Kind:       classifyFilename(originalName),
		Size:       fileInformation.Size(),
		CreatedAt:  database.FormatTimestampNow(),
	}

	card.Attachments = append(card.Attachments, attachment)
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board with attachment: %w", saveError)
	}

	return response{
		"id":   attachment.Identifier,
		"name": attachment.Name,
		"kind": attachment.Kind,
		"size": attachment.Size,
	}, nil
}

func (handler *Handler) findAttachmentFilePath(attachmentIdentifier int64) (*database.Attachment, string, error) {
	board, _, _, attachment := handler.store.FindAttachment(attachmentIdentifier)
	if attachment == nil {
		return nil, "", fmt.Errorf("Attachment not found")
	}

	fullPath := filepath.Join(handler.store.AttachmentDirectory(board), filepath.FromSlash(attachment.File))
	if _, statError := os.Stat(fullPath); statError != nil {
		return nil, "", fmt.Errorf("Attachment file missing: %s", attachment.Name)
	}

	return attachment, fullPath, nil
}

func (handler *Handler) AddAttachmentsDialog(cardIdentifier int64) (response, error) {
	filePaths, dialogError := OpenFileDialog(true, handler.nativeWindowHandle())
	if dialogError != nil {
		return nil, dialogError
	}

	addedAttachments := make([]response, 0, len(filePaths))
	for _, sourcePath := range filePaths {
		attachmentResponse, storeError := handler.storeAttachment(cardIdentifier, filepath.Base(sourcePath), fileCopier(sourcePath))
		if storeError == nil {
			addedAttachments = append(addedAttachments, attachmentResponse)
		}
	}

	return response{"attachments": addedAttachments}, nil
}

func (handler *Handler) StoreUploadedAttachment(cardIdentifier int64, filename string, content io.Reader) (map[string]interface{}, error) {
	handler.operationMutex.Lock()
	defer handler.operationMutex.Unlock()

	return handler.storeAttachment(cardIdentifier, filename, func(destinationPath string) error {
		return writeReaderToFile(destinationPath, content)
	})
}

func (handler *Handler) ResolveAttachmentFile(attachmentIdentifier int64) (string, string, bool) {
	handler.operationMutex.Lock()
	defer handler.operationMutex.Unlock()

	attachment, fullPath, findError := handler.findAttachmentFilePath(attachmentIdentifier)
	if findError != nil {
		return "", "", false
	}
	return fullPath, determineMediaType(attachment.Name), true
}

func (handler *Handler) OpenAttachment(attachmentIdentifier int64) (response, error) {
	_, fullPath, findError := handler.findAttachmentFilePath(attachmentIdentifier)
	if findError != nil {
		return nil, findError
	}

	if openError := OpenPath(fullPath); openError != nil {
		return nil, fmt.Errorf("failed to open attachment: %w", openError)
	}

	return okResponse(), nil
}

func (handler *Handler) SaveAttachmentAs(attachmentIdentifier int64) (response, error) {
	attachment, fullPath, findError := handler.findAttachmentFilePath(attachmentIdentifier)
	if findError != nil {
		return nil, findError
	}

	destinationPath, dialogError := SaveFileDialog(attachment.Name, handler.nativeWindowHandle())
	if dialogError != nil {
		return nil, dialogError
	}
	if destinationPath == "" {
		return response{"saved": false}, nil
	}

	if copyError := copyFileContents(fullPath, destinationPath); copyError != nil {
		return nil, copyError
	}

	return response{"saved": true, "path": destinationPath}, nil
}

func (handler *Handler) RenameAttachment(attachmentIdentifier int64, name string) (response, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("Name cannot be empty")
	}

	board, _, _, attachment := handler.store.FindAttachment(attachmentIdentifier)
	if attachment == nil {
		return nil, fmt.Errorf("Attachment not found")
	}

	originalExtension := filepath.Ext(attachment.Name)
	if originalExtension != "" && !strings.HasSuffix(strings.ToLower(name), strings.ToLower(originalExtension)) {
		name += originalExtension
	}

	attachment.Name = name
	if saveError := handler.store.SaveBoard(board); saveError != nil {
		return nil, fmt.Errorf("failed to save board: %w", saveError)
	}

	return response{"ok": true, "name": name}, nil
}

func (handler *Handler) DeleteAttachment(attachmentIdentifier int64) (response, error) {
	board, _, card, attachment := handler.store.FindAttachment(attachmentIdentifier)
	if attachment == nil {
		return nil, fmt.Errorf("Attachment not found")
	}

	card.Attachments = database.RemoveByIdentifier(card.Attachments, attachmentIdentifier)
	if card.CoverIdentifier != nil && *card.CoverIdentifier == attachmentIdentifier {
		card.CoverIdentifier = nil
	}

	return handler.saveBoardOrFail(board)
}
