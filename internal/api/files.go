package api

import (
	"io"
	"os"
)

func writeReaderToFile(destinationPath string, content io.Reader) error {
	destinationFile, createError := os.Create(destinationPath)
	if createError != nil {
		return createError
	}
	defer destinationFile.Close()

	_, copyError := io.Copy(destinationFile, content)
	return copyError
}

func copyFileContents(sourcePath string, destinationPath string) error {
	sourceFile, openError := os.Open(sourcePath)
	if openError != nil {
		return openError
	}
	defer sourceFile.Close()

	return writeReaderToFile(destinationPath, sourceFile)
}

func fileCopier(sourcePath string) func(destinationPath string) error {
	return func(destinationPath string) error {
		return copyFileContents(sourcePath, destinationPath)
	}
}
