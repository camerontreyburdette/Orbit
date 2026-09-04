package assets

import "embed"

//go:embed font/*
var EmbeddedFileSystem embed.FS

//go:embed icon.ico
var ApplicationIconBytes []byte
