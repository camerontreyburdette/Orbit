package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strconv"
)

const uploadFormFieldName = "file"

type AttachmentService interface {
	ResolveAttachmentFile(attachmentIdentifier int64) (filePath string, mediaType string, found bool)
	StoreUploadedAttachment(cardIdentifier int64, filename string, content io.Reader) (map[string]interface{}, error)
}

type BackendServices interface {
	AttachmentService
	BoardImportService
}

type StaticServer struct {
	httpServer *http.Server
	listener   net.Listener
	address    string
}

var explicitContentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "text/javascript; charset=utf-8",
	".mjs":  "text/javascript; charset=utf-8",
	".svg":  "image/svg+xml",
}

type staticAssetHandler struct {
	fileServer http.Handler
}

func (handler *staticAssetHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	headers := responseWriter.Header()
	headers.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if contentType, hasContentType := explicitContentTypes[path.Ext(request.URL.Path)]; hasContentType {
		headers.Set("Content-Type", contentType)
	}
	handler.fileServer.ServeHTTP(responseWriter, request)
}

func parseIdentifier(request *http.Request, name string) (int64, bool) {
	identifier, parseError := strconv.ParseInt(request.PathValue(name), 10, 64)
	return identifier, parseError == nil
}

type attachmentDownloadHandler struct {
	service AttachmentService
}

func (handler *attachmentDownloadHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	attachmentIdentifier, isValid := parseIdentifier(request, "identifier")
	if !isValid {
		http.NotFound(responseWriter, request)
		return
	}

	filePath, mediaType, found := handler.service.ResolveAttachmentFile(attachmentIdentifier)
	if !found {
		http.NotFound(responseWriter, request)
		return
	}

	headers := responseWriter.Header()
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Content-Type", mediaType)
	http.ServeFile(responseWriter, request, filePath)
}

type attachmentUploadHandler struct {
	service AttachmentService
}

func (handler *attachmentUploadHandler) storeParts(cardIdentifier int64, request *http.Request) ([]map[string]interface{}, error) {
	multipartReader, readerError := request.MultipartReader()
	if readerError != nil {
		return nil, readerError
	}

	stored := make([]map[string]interface{}, 0)
	for {
		part, partError := multipartReader.NextPart()
		if errors.Is(partError, io.EOF) {
			return stored, nil
		}
		if partError != nil {
			return nil, partError
		}
		if part.FormName() != uploadFormFieldName || part.FileName() == "" {
			_ = part.Close()
			continue
		}
		result, storeError := handler.service.StoreUploadedAttachment(cardIdentifier, part.FileName(), part)
		_ = part.Close()
		if storeError != nil {
			return nil, storeError
		}
		stored = append(stored, result)
	}
}

func (handler *attachmentUploadHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	cardIdentifier, isValid := parseIdentifier(request, "identifier")
	if !isValid {
		http.NotFound(responseWriter, request)
		return
	}

	stored, storeError := handler.storeParts(cardIdentifier, request)
	if storeError != nil {
		http.Error(responseWriter, storeError.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(responseWriter).Encode(map[string]interface{}{"attachments": stored})
}

func StartStaticServer(staticFileSystem fs.FS, services BackendServices) (*StaticServer, error) {
	listener, listenError := net.Listen("tcp", "127.0.0.1:0")
	if listenError != nil {
		return nil, fmt.Errorf("failed to bind static server listener: %w", listenError)
	}

	multiplexer := http.NewServeMux()
	multiplexer.Handle("GET /attachments/{identifier}", &attachmentDownloadHandler{service: services})
	multiplexer.Handle("POST /cards/{identifier}/attachments", &attachmentUploadHandler{service: services})
	multiplexer.Handle("POST /boards/import", &boardImportHandler{service: services})
	multiplexer.Handle("/", &staticAssetHandler{fileServer: http.FileServer(http.FS(staticFileSystem))})

	httpServer := &http.Server{Handler: multiplexer}
	go func() {
		_ = httpServer.Serve(listener)
	}()

	return &StaticServer{
		httpServer: httpServer,
		listener:   listener,
		address:    fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port),
	}, nil
}

func (server *StaticServer) Address() string {
	return server.address
}

func (server *StaticServer) Close() error {
	return server.httpServer.Close()
}
