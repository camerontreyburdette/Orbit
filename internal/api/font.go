package api

import (
	"encoding/base64"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maximumFontFileSizeBytes = 8 * 1024 * 1024
	defaultFontWeight        = 400
)

type FontDescriptor struct {
	DataUniformResourceIdentifier string `json:"data_uri"`
	Format                        string `json:"format"`
	Weight                        int    `json:"weight"`
	Style                         string `json:"style"`
}

type fontExtensionInformation struct {
	mediaType string
	format    string
}

var supportedFontExtensions = map[string]fontExtensionInformation{
	".woff2": {mediaType: "font/woff2", format: "woff2"},
	".woff":  {mediaType: "font/woff", format: "woff"},
	".ttf":   {mediaType: "font/ttf", format: "truetype"},
	".otf":   {mediaType: "font/otf", format: "opentype"},
}

type fontWeightRule struct {
	keyword string
	weight  int
}

var fontWeightRules = []fontWeightRule{
	{keyword: "extralight", weight: 200},
	{keyword: "semibold", weight: 600},
	{keyword: "demibold", weight: 600},
	{keyword: "extrabold", weight: 800},
	{keyword: "thin", weight: 100},
	{keyword: "light", weight: 300},
	{keyword: "medium", weight: 500},
	{keyword: "black", weight: 900},
	{keyword: "heavy", weight: 900},
	{keyword: "bold", weight: 700},
}

func fontWeightFromFilename(lowercaseFilename string) int {
	for _, rule := range fontWeightRules {
		if strings.Contains(lowercaseFilename, rule.keyword) {
			return rule.weight
		}
	}
	return defaultFontWeight
}

func fontStyleFromFilename(lowercaseFilename string) string {
	if strings.Contains(lowercaseFilename, "italic") || strings.Contains(lowercaseFilename, "oblique") {
		return "italic"
	}
	return "normal"
}

func loadFontDescriptor(fileSystem fs.FS, directoryPath string, filename string) (FontDescriptor, bool) {
	extensionInformation, isSupported := supportedFontExtensions[lowercaseExtension(filename)]
	if !isSupported {
		return FontDescriptor{}, false
	}

	fileBytes, readFileError := fs.ReadFile(fileSystem, filepath.ToSlash(filepath.Join(directoryPath, filename)))
	if readFileError != nil || len(fileBytes) > maximumFontFileSizeBytes {
		return FontDescriptor{}, false
	}

	lowercaseFilename := strings.ToLower(filename)
	return FontDescriptor{
		DataUniformResourceIdentifier: "data:" + extensionInformation.mediaType + ";base64," + base64.StdEncoding.EncodeToString(fileBytes),
		Format:                        extensionInformation.format,
		Weight:                        fontWeightFromFilename(lowercaseFilename),
		Style:                         fontStyleFromFilename(lowercaseFilename),
	}, true
}

func LoadFontsFromFileSystem(fileSystem fs.FS, directoryPath string) []FontDescriptor {
	fonts := make([]FontDescriptor, 0)
	entries, readError := fs.ReadDir(fileSystem, directoryPath)
	if readError != nil {
		return fonts
	}

	sort.Slice(entries, func(firstIndex, secondIndex int) bool {
		return entries[firstIndex].Name() < entries[secondIndex].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if descriptor, isLoaded := loadFontDescriptor(fileSystem, directoryPath, entry.Name()); isLoaded {
			fonts = append(fonts, descriptor)
		}
	}

	return fonts
}
