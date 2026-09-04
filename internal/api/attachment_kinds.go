package api

import (
	"crypto/rand"
	"encoding/hex"
	"mime"
	"path/filepath"
	"strings"
)

const maximumSanitizedExtensionLength = 12

var (
	imageExtensions    = map[string]struct{}{".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}}
	videoExtensions    = map[string]struct{}{".mp4": {}, ".webm": {}, ".mov": {}}
	audioExtensions    = map[string]struct{}{".mp3": {}, ".wav": {}, ".ogg": {}}
	documentExtensions = map[string]struct{}{".pdf": {}, ".txt": {}, ".docx": {}, ".doc": {}, ".rtf": {}, ".md": {}, ".csv": {}, ".xlsx": {}, ".pptx": {}}
	fallbackMediaTypes = map[string]string{
		".webp": "image/webp",
		".ogg":  "audio/ogg",
		".wav":  "audio/wav",
		".mov":  "video/quicktime",
		".md":   "text/plain",
	}
)

func lowercaseExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

func classifyFilename(filename string) string {
	extension := lowercaseExtension(filename)
	if _, isImage := imageExtensions[extension]; isImage {
		return "image"
	}
	if _, isVideo := videoExtensions[extension]; isVideo {
		return "video"
	}
	if _, isAudio := audioExtensions[extension]; isAudio {
		return "audio"
	}
	if _, isDocument := documentExtensions[extension]; isDocument {
		return "document"
	}
	return "other"
}

func determineMediaType(filename string) string {
	extension := lowercaseExtension(filename)
	if detectedMediaType := mime.TypeByExtension(extension); detectedMediaType != "" {
		mediaType, _, _ := strings.Cut(detectedMediaType, ";")
		return mediaType
	}
	if fallback, hasFallback := fallbackMediaTypes[extension]; hasFallback {
		return fallback
	}
	return "application/octet-stream"
}

func sanitizeExtension(filename string) string {
	var builder strings.Builder
	for _, character := range lowercaseExtension(filename) {
		isLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if isLetter || isDigit || character == '.' {
			builder.WriteRune(character)
		}
	}
	sanitized := builder.String()
	if len(sanitized) > maximumSanitizedExtensionLength {
		return sanitized[:maximumSanitizedExtensionLength]
	}
	return sanitized
}

func generateStoredFilename(originalName string) string {
	randomBytes := make([]byte, 16)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + sanitizeExtension(originalName)
}
