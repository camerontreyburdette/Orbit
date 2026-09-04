package api

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	boardDocumentFileName    = "board.json"
	attachmentsDirectoryName = "attachments"
	folderDialogTitle        = "Choose a board folder"
)

func findBoardDocumentPaths(rootDirectory string) []string {
	documentPaths := make([]string, 0)
	_ = filepath.WalkDir(rootDirectory, func(currentPath string, entry fs.DirEntry, walkError error) error {
		if walkError != nil || entry.IsDir() {
			return nil
		}
		if strings.EqualFold(entry.Name(), boardDocumentFileName) {
			documentPaths = append(documentPaths, currentPath)
		}
		return nil
	})
	return documentPaths
}

func copyDirectoryFileToStore(attachmentDirectory string, currentPath string, storeAttachment attachmentStoreFunction) error {
	relativePath, relativeError := filepath.Rel(attachmentDirectory, currentPath)
	if relativeError != nil {
		return relativeError
	}
	sourceFile, openError := os.Open(currentPath)
	if openError != nil {
		return openError
	}
	defer sourceFile.Close()
	return storeAttachment(filepath.ToSlash(relativePath), sourceFile)
}

func streamDirectoryAttachments(attachmentDirectory string) func(storeAttachment attachmentStoreFunction) error {
	return func(storeAttachment attachmentStoreFunction) error {
		return filepath.WalkDir(attachmentDirectory, func(currentPath string, entry fs.DirEntry, walkError error) error {
			if walkError != nil || entry.IsDir() {
				return nil
			}
			return copyDirectoryFileToStore(attachmentDirectory, currentPath, storeAttachment)
		})
	}
}

func (handler *Handler) importBoardDirectoryLocked(boardDirectory string) error {
	documentBytes, readError := os.ReadFile(filepath.Join(boardDirectory, boardDocumentFileName))
	if readError != nil {
		return readError
	}
	session, beginError := handler.beginBoardImportLocked(documentBytes)
	if beginError != nil {
		return beginError
	}
	_, completeError := session.completeLocked(streamDirectoryAttachments(filepath.Join(boardDirectory, attachmentsDirectoryName)))
	return completeError
}

func (handler *Handler) importBoardDirectoriesLocked(rootDirectory string) (int, int) {
	importedCount := 0
	failedCount := 0
	for _, documentPath := range findBoardDocumentPaths(rootDirectory) {
		if importError := handler.importBoardDirectoryLocked(filepath.Dir(documentPath)); importError != nil {
			failedCount++
		} else {
			importedCount++
		}
	}
	return importedCount, failedCount
}

func (handler *Handler) ImportBoardsDialog() (response, error) {
	rootDirectory, dialogError := OpenFolderDialog(folderDialogTitle, handler.nativeWindowHandle())
	if dialogError != nil {
		return nil, dialogError
	}
	if rootDirectory == "" {
		return response{"cancelled": true}, nil
	}

	importedCount, failedCount := handler.importBoardDirectoriesLocked(rootDirectory)
	importResponse := handler.boardListResponse()
	importResponse["imported_count"] = importedCount
	importResponse["failed_count"] = failedCount
	return importResponse, nil
}
