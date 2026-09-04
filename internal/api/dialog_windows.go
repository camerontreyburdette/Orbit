//go:build windows

package api

import (
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	commonDialogLibrary      = syscall.NewLazyDLL("comdlg32.dll")
	getOpenFileNameProcedure = commonDialogLibrary.NewProc("GetOpenFileNameW")
	getSaveFileNameProcedure = commonDialogLibrary.NewProc("GetSaveFileNameW")
)

const (
	flagExplorer          = 0x00080000
	flagFileMustExist     = 0x00001000
	flagPathMustExist     = 0x00000800
	flagAllowMultiSelect  = 0x00000200
	flagOverwritePrompt   = 0x00000002
	flagHideReadOnly      = 0x00000004
	bufferCharacterLength = 65536
	allFilesFilter        = "All Files (*.*)\x00*.*\x00\x00"
)

type openFileNameStructure struct {
	structureSize             uint32
	ownerWindow               uintptr
	instance                  uintptr
	filter                    *uint16
	customFilter              *uint16
	maximumCustomFilterLength uint32
	filterIndex               uint32
	file                      *uint16
	maximumFileLength         uint32
	fileTitle                 *uint16
	maximumFileTitleLength    uint32
	initialDirectory          *uint16
	title                     *uint16
	flags                     uint32
	fileOffset                uint16
	fileExtension             uint16
	defaultExtension          *uint16
	customData                uintptr
	hookFunction              uintptr
	templateName              *uint16
	reservedEdit              uintptr
	reservedAttributes        uint32
	extendedFlags             uint32
}

type fileDialogRequest struct {
	structure openFileNameStructure
	buffer    []uint16
	filter    []uint16
}

func newFileDialogRequest(parentWindow uintptr, flags uint32, initialFilename string) *fileDialogRequest {
	request := &fileDialogRequest{
		buffer: make([]uint16, bufferCharacterLength),
		filter: utf16.Encode([]rune(allFilesFilter)),
	}
	if initialFilename != "" {
		copy(request.buffer, utf16.Encode([]rune(initialFilename)))
	}
	request.structure.structureSize = uint32(unsafe.Sizeof(request.structure))
	request.structure.ownerWindow = parentWindow
	request.structure.filter = &request.filter[0]
	request.structure.file = &request.buffer[0]
	request.structure.maximumFileLength = uint32(len(request.buffer))
	request.structure.flags = flags
	return request
}

func (request *fileDialogRequest) invoke(procedure *syscall.LazyProc) bool {
	returnCode, _, _ := procedure.Call(uintptr(unsafe.Pointer(&request.structure)))
	return returnCode != 0
}

func (request *fileDialogRequest) nullSeparatedTokens() []string {
	var tokens []string
	buffer := request.buffer
	startOffset := 0
	for index := 0; index < len(buffer)-1; index++ {
		if buffer[index] != 0 {
			continue
		}
		if startOffset == index {
			break
		}
		tokens = append(tokens, string(utf16.Decode(buffer[startOffset:index])))
		startOffset = index + 1
		if buffer[startOffset] == 0 {
			break
		}
	}
	return tokens
}

func (request *fileDialogRequest) singleToken() string {
	endIndex := 0
	for endIndex < len(request.buffer) && request.buffer[endIndex] != 0 {
		endIndex++
	}
	return string(utf16.Decode(request.buffer[:endIndex]))
}

func OpenFileDialog(allowMultipleSelection bool, parentWindow uintptr) ([]string, error) {
	var flags uint32 = flagExplorer | flagFileMustExist | flagPathMustExist | flagHideReadOnly
	if allowMultipleSelection {
		flags |= flagAllowMultiSelect
	}

	request := newFileDialogRequest(parentWindow, flags, "")
	if !request.invoke(getOpenFileNameProcedure) {
		return []string{}, nil
	}

	tokens := request.nullSeparatedTokens()
	if len(tokens) == 0 {
		return []string{}, nil
	}
	if len(tokens) == 1 {
		return tokens, nil
	}

	baseDirectory := tokens[0]
	filePaths := make([]string, 0, len(tokens)-1)
	for _, filename := range tokens[1:] {
		filePaths = append(filePaths, filepath.Join(baseDirectory, filename))
	}
	return filePaths, nil
}

func SaveFileDialog(defaultFilename string, parentWindow uintptr) (string, error) {
	request := newFileDialogRequest(parentWindow, flagExplorer|flagPathMustExist|flagOverwritePrompt|flagHideReadOnly, defaultFilename)
	if !request.invoke(getSaveFileNameProcedure) {
		return "", nil
	}
	return request.singleToken(), nil
}
