package api

import (
	"os"
	"path/filepath"
	"strings"

	"orbit/internal/database"
)

func cardAttachmentFolders(cards []*database.Card) []string {
	folders := make([]string, 0, len(cards))
	for _, card := range cards {
		if folder := database.CardAttachmentFolder(card); folder != "" {
			folders = append(folders, folder)
		}
	}
	return folders
}

func joinDirectoryPaths(rootDirectory string, folders []string) []string {
	paths := make([]string, 0, len(folders))
	for _, folder := range folders {
		paths = append(paths, filepath.Join(rootDirectory, folder))
	}
	return paths
}

func removeDirectories(directoryPaths []string) {
	for _, directoryPath := range directoryPaths {
		_ = os.RemoveAll(directoryPath)
	}
}

func (handler *Handler) removeCardAttachmentFolders(board *database.Board, folders []string) {
	if len(folders) == 0 {
		return
	}
	removeDirectories(joinDirectoryPaths(handler.store.AttachmentDirectory(board), folders))
}

func (handler *Handler) occupiedCardDirectories(board *database.Board, card *database.Card, rootDirectory string, excludedFolder string) map[string]struct{} {
	occupiedDirectories := make(map[string]struct{})
	for _, list := range board.Lists {
		for _, currentCard := range list.Cards {
			if currentCard.Identifier == card.Identifier {
				continue
			}
			if folder := database.CardAttachmentFolder(currentCard); folder != "" {
				occupiedDirectories[strings.ToLower(folder)] = struct{}{}
			}
		}
	}

	entries, readError := os.ReadDir(rootDirectory)
	if readError != nil {
		return occupiedDirectories
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if excludedFolder != "" && strings.EqualFold(entry.Name(), excludedFolder) {
			continue
		}
		occupiedDirectories[strings.ToLower(entry.Name())] = struct{}{}
	}
	return occupiedDirectories
}

func (handler *Handler) resolveCardDirectoryName(board *database.Board, card *database.Card, rootDirectory string, excludedFolder string) string {
	return database.CardDirectoryName(card, handler.occupiedCardDirectories(board, card, rootDirectory, excludedFolder))
}

func relocateAttachmentFiles(card *database.Card, newFolder string) {
	for _, attachment := range card.Attachments {
		if _, remainder, hasSeparator := strings.Cut(attachment.File, "/"); hasSeparator {
			attachment.File = newFolder + "/" + remainder
		}
	}
}

func (handler *Handler) synchronizeCardDirectory(board *database.Board, card *database.Card) {
	oldFolder := database.CardAttachmentFolder(card)
	if oldFolder == "" {
		return
	}

	rootDirectory := handler.store.AttachmentDirectory(board)
	newFolder := handler.resolveCardDirectoryName(board, card, rootDirectory, oldFolder)
	if newFolder == oldFolder {
		return
	}

	if renameError := os.Rename(filepath.Join(rootDirectory, oldFolder), filepath.Join(rootDirectory, newFolder)); renameError != nil {
		return
	}
	relocateAttachmentFiles(card, newFolder)
}

func (handler *Handler) moveCardDirectory(sourceBoard *database.Board, targetBoard *database.Board, card *database.Card) {
	oldFolder := database.CardAttachmentFolder(card)
	if oldFolder == "" {
		return
	}

	sourcePath := filepath.Join(handler.store.AttachmentDirectory(sourceBoard), oldFolder)
	if _, statError := os.Stat(sourcePath); statError != nil {
		return
	}

	targetRoot := handler.store.AttachmentDirectory(targetBoard)
	if mkdirError := os.MkdirAll(targetRoot, 0750); mkdirError != nil {
		return
	}

	newFolder := handler.resolveCardDirectoryName(targetBoard, card, targetRoot, "")
	if moveError := os.Rename(sourcePath, filepath.Join(targetRoot, newFolder)); moveError != nil {
		return
	}

	if newFolder != oldFolder {
		relocateAttachmentFiles(card, newFolder)
	}
}

func (handler *Handler) copyCardAttachments(board *database.Board, originalCard *database.Card, clonedCard *database.Card) {
	oldFolder := database.CardAttachmentFolder(originalCard)
	if oldFolder == "" {
		return
	}

	rootDirectory := handler.store.AttachmentDirectory(board)
	if _, statError := os.Stat(filepath.Join(rootDirectory, oldFolder)); statError != nil {
		return
	}

	newFolder := handler.resolveCardDirectoryName(board, clonedCard, rootDirectory, "")
	targetPath := filepath.Join(rootDirectory, newFolder)
	if mkdirError := os.MkdirAll(targetPath, 0750); mkdirError != nil {
		return
	}

	clonedAttachments := make([]*database.Attachment, 0, len(originalCard.Attachments))
	for _, originalAttachment := range originalCard.Attachments {
		sourceFile := filepath.Join(rootDirectory, filepath.FromSlash(originalAttachment.File))
		fileName := filepath.Base(sourceFile)
		if copyError := copyFileContents(sourceFile, filepath.Join(targetPath, fileName)); copyError != nil {
			continue
		}

		clonedAttachment := &database.Attachment{
			Identifier: handler.store.NewIdentifier(),
			Name:       originalAttachment.Name,
			File:       newFolder + "/" + fileName,
			Kind:       originalAttachment.Kind,
			Size:       originalAttachment.Size,
			CreatedAt:  database.FormatTimestampNow(),
		}
		clonedAttachments = append(clonedAttachments, clonedAttachment)

		if originalCard.CoverIdentifier != nil && *originalCard.CoverIdentifier == originalAttachment.Identifier {
			coverIdentifier := clonedAttachment.Identifier
			clonedCard.CoverIdentifier = &coverIdentifier
		}
	}
	clonedCard.Attachments = clonedAttachments
}
