package frontend

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed static/*
var embeddedFileSystem embed.FS

func StaticFileSystem() (fs.FS, error) {
	staticFileSystem, subError := fs.Sub(embeddedFileSystem, "static")
	if subError != nil {
		return nil, fmt.Errorf("failed to prepare static filesystem: %w", subError)
	}
	return staticFileSystem, nil
}
