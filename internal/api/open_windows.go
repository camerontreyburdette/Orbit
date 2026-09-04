//go:build windows

package api

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	shellLibrary          = syscall.NewLazyDLL("shell32.dll")
	shellExecuteProcedure = shellLibrary.NewProc("ShellExecuteW")
)

const showNormal = 1

func OpenPath(targetPath string) error {
	operationPointer, stringConversionError := syscall.UTF16PtrFromString("open")
	if stringConversionError != nil {
		return fmt.Errorf("failed to convert operation string: %w", stringConversionError)
	}

	pathPointer, stringConversionError := syscall.UTF16PtrFromString(targetPath)
	if stringConversionError != nil {
		return fmt.Errorf("failed to convert target path string: %w", stringConversionError)
	}

	result, _, _ := shellExecuteProcedure.Call(
		0,
		uintptr(unsafe.Pointer(operationPointer)),
		uintptr(unsafe.Pointer(pathPointer)),
		0,
		0,
		uintptr(showNormal),
	)

	if result <= 32 {
		return fmt.Errorf("ShellExecuteW failed with error code %d", result)
	}

	return nil
}
