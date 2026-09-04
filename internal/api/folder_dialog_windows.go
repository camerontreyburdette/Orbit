//go:build windows

package api

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	objectLinkingLibrary      = syscall.NewLazyDLL("ole32.dll")
	coInitializeExProcedure   = objectLinkingLibrary.NewProc("CoInitializeEx")
	coCreateInstanceProcedure = objectLinkingLibrary.NewProc("CoCreateInstance")
	coTaskMemFreeProcedure    = objectLinkingLibrary.NewProc("CoTaskMemFree")
)

type globallyUniqueIdentifier struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

var (
	fileOpenDialogClassIdentifier     = globallyUniqueIdentifier{0xDC1C5A9C, 0xE88A, 0x4DDE, [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	fileOpenDialogInterfaceIdentifier = globallyUniqueIdentifier{0xD57C7288, 0xD4AD, 0x4768, [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60}}
)

const (
	apartmentThreadedInitialization = 0x2
	inProcessServerContext          = 0x1
	optionPickFolders               = 0x20
	optionForceFileSystem           = 0x40
	optionPathMustExist             = 0x800
	displayNameFileSystemPath       = 0x80058000

	vtableRelease        = 2
	vtableShow           = 3
	vtableSetOptions     = 9
	vtableGetOptions     = 10
	vtableSetTitle       = 17
	vtableGetResult      = 20
	vtableGetDisplayName = 5
)

type componentObject struct {
	vtable *[32]uintptr
}

func (object *componentObject) call(methodIndex int, arguments ...uintptr) int32 {
	callArguments := append([]uintptr{uintptr(unsafe.Pointer(object))}, arguments...)
	result, _, _ := syscall.SyscallN(object.vtable[methodIndex], callArguments...)
	return int32(result)
}

func (object *componentObject) release() {
	object.call(vtableRelease)
}

func createFileOpenDialog() (*componentObject, error) {
	_, _, _ = coInitializeExProcedure.Call(0, apartmentThreadedInitialization)

	var dialog *componentObject
	result, _, _ := coCreateInstanceProcedure.Call(
		uintptr(unsafe.Pointer(&fileOpenDialogClassIdentifier)),
		0,
		inProcessServerContext,
		uintptr(unsafe.Pointer(&fileOpenDialogInterfaceIdentifier)),
		uintptr(unsafe.Pointer(&dialog)),
	)
	if int32(result) < 0 || dialog == nil {
		return nil, fmt.Errorf("failed to create folder dialog: 0x%x", result)
	}
	return dialog, nil
}

func applyFolderPickerOptions(dialog *componentObject, title string) {
	var currentOptions uint32
	dialog.call(vtableGetOptions, uintptr(unsafe.Pointer(&currentOptions)))
	dialog.call(vtableSetOptions, uintptr(currentOptions|optionPickFolders|optionForceFileSystem|optionPathMustExist))

	if titlePointer, conversionError := syscall.UTF16PtrFromString(title); conversionError == nil {
		dialog.call(vtableSetTitle, uintptr(unsafe.Pointer(titlePointer)))
	}
}

func selectedFileSystemPath(dialog *componentObject) string {
	var shellItem *componentObject
	if dialog.call(vtableGetResult, uintptr(unsafe.Pointer(&shellItem))) < 0 || shellItem == nil {
		return ""
	}
	defer shellItem.release()

	var pathPointer *uint16
	if shellItem.call(vtableGetDisplayName, displayNameFileSystemPath, uintptr(unsafe.Pointer(&pathPointer))) < 0 || pathPointer == nil {
		return ""
	}
	defer coTaskMemFreeProcedure.Call(uintptr(unsafe.Pointer(pathPointer)))
	return windows.UTF16PtrToString(pathPointer)
}

func OpenFolderDialog(title string, parentWindow uintptr) (string, error) {
	dialog, createError := createFileOpenDialog()
	if createError != nil {
		return "", createError
	}
	defer dialog.release()

	applyFolderPickerOptions(dialog, title)
	if dialog.call(vtableShow, parentWindow) < 0 {
		return "", nil
	}
	return selectedFileSystemPath(dialog), nil
}
