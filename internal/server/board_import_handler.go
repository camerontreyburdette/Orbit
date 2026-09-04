package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	importBoardFormFieldName          = "board"
	importAttachmentPathFormFieldName = "attachment_path"
	importAttachmentFileFormFieldName = "attachment_file"
	maximumBoardDocumentBytes         = 64 << 20
	maximumAttachmentPathBytes        = 4096
)

type attachmentStoreFunction = func(relativePath string, content io.Reader) error

type BoardImportService interface {
	ImportBoard(documentBytes []byte, streamAttachments func(storeAttachment attachmentStoreFunction) error) (map[string]interface{}, error)
}

type boardImportHandler struct {
	service BoardImportService
}

func readBoardDocumentPart(multipartReader *multipart.Reader) ([]byte, error) {
	part, partError := multipartReader.NextPart()
	if partError != nil {
		return nil, fmt.Errorf("missing board document")
	}
	defer part.Close()

	if part.FormName() != importBoardFormFieldName {
		return nil, fmt.Errorf("board document must be the first part")
	}
	return io.ReadAll(io.LimitReader(part, maximumBoardDocumentBytes))
}

func consumeAttachmentPart(part *multipart.Part, pendingPath *string, storeAttachment attachmentStoreFunction) error {
	switch part.FormName() {
	case importAttachmentPathFormFieldName:
		pathBytes, readError := io.ReadAll(io.LimitReader(part, maximumAttachmentPathBytes))
		if readError != nil {
			return readError
		}
		*pendingPath = string(pathBytes)
		return nil
	case importAttachmentFileFormFieldName:
		if *pendingPath == "" {
			return fmt.Errorf("attachment file received without a path")
		}
		relativePath := *pendingPath
		*pendingPath = ""
		return storeAttachment(relativePath, part)
	}
	return nil
}

func streamAttachmentParts(multipartReader *multipart.Reader) func(storeAttachment attachmentStoreFunction) error {
	return func(storeAttachment attachmentStoreFunction) error {
		pendingPath := ""
		for {
			part, partError := multipartReader.NextPart()
			if errors.Is(partError, io.EOF) {
				return nil
			}
			if partError != nil {
				return partError
			}
			consumeError := consumeAttachmentPart(part, &pendingPath, storeAttachment)
			_ = part.Close()
			if consumeError != nil {
				return consumeError
			}
		}
	}
}

func (handler *boardImportHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	multipartReader, readerError := request.MultipartReader()
	if readerError != nil {
		http.Error(responseWriter, readerError.Error(), http.StatusBadRequest)
		return
	}

	documentBytes, documentError := readBoardDocumentPart(multipartReader)
	if documentError != nil {
		http.Error(responseWriter, documentError.Error(), http.StatusBadRequest)
		return
	}

	result, importError := handler.service.ImportBoard(documentBytes, streamAttachmentParts(multipartReader))
	if importError != nil {
		http.Error(responseWriter, importError.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(responseWriter).Encode(result)
}
