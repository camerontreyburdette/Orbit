//go:build windows

package window

import (
	"sync"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"

	"orbit/internal/configuration"
)

type WindowsAppWindow struct {
	windowHandle         uintptr
	mainThreadIdentifier uintptr
	chromiumInstance     *edge.Chromium
	configurationManager *configuration.Manager
	backgroundBrush      uintptr
	hasSavedState        bool
	mutex                sync.Mutex
	dispatchQueue        []func()
	bindings             map[string]boundFunction
	minimumSize          pointStructure
	maximumSize          pointStructure
}

var (
	applicationWindowsMap  = map[uintptr]*WindowsAppWindow{}
	applicationWindowsLock sync.RWMutex
)

func getApplicationWindow(handle uintptr) *WindowsAppWindow {
	applicationWindowsLock.RLock()
	defer applicationWindowsLock.RUnlock()
	return applicationWindowsMap[handle]
}

func registerApplicationWindow(handle uintptr, window *WindowsAppWindow) {
	applicationWindowsLock.Lock()
	defer applicationWindowsLock.Unlock()
	applicationWindowsMap[handle] = window
}

func unregisterApplicationWindow(handle uintptr) {
	applicationWindowsLock.Lock()
	defer applicationWindowsLock.Unlock()
	delete(applicationWindowsMap, handle)
}

func callDefaultWindowProcedure(handle uintptr, message uintptr, wordParameter uintptr, longParameter uintptr) uintptr {
	result, _, _ := defaultWindowProcedureProcedure.Call(handle, message, wordParameter, longParameter)
	return result
}

func (windowInstance *WindowsAppWindow) eraseBackground(handle uintptr, deviceContext uintptr) {
	rectangle := rectangleStructure{}
	getClientRectangleProcedure.Call(handle, uintptr(unsafe.Pointer(&rectangle)))
	fillRectangleProcedure.Call(deviceContext, uintptr(unsafe.Pointer(&rectangle)), windowInstance.backgroundBrush)
}

func (windowInstance *WindowsAppWindow) applySizeConstraints(longParameter uintptr) {
	minimumMaximumInformation := (*minimumMaximumInformationStructure)(*(*unsafe.Pointer)(unsafe.Pointer(&longParameter)))
	if windowInstance.maximumSize.horizontalCoordinate > 0 && windowInstance.maximumSize.verticalCoordinate > 0 {
		minimumMaximumInformation.maximumSizePoint = windowInstance.maximumSize
		minimumMaximumInformation.maximumTrackSizePoint = windowInstance.maximumSize
	}
	if windowInstance.minimumSize.horizontalCoordinate > 0 && windowInstance.minimumSize.verticalCoordinate > 0 {
		minimumMaximumInformation.minimumTrackSizePoint = windowInstance.minimumSize
	}
}

func windowProcedure(handle uintptr, message uintptr, wordParameter uintptr, longParameter uintptr) uintptr {
	windowInstance := getApplicationWindow(handle)
	if windowInstance == nil {
		return callDefaultWindowProcedure(handle, message, wordParameter, longParameter)
	}

	switch uint32(message) {
	case windowMessageEraseBackground:
		windowInstance.eraseBackground(handle, wordParameter)
		return 1

	case windowMessageMove, windowMessageMoving:
		if windowInstance.chromiumInstance != nil {
			_ = windowInstance.chromiumInstance.NotifyParentWindowPositionChanged()
		}

	case windowMessageNonClientLeftButton:
		setFocusProcedure.Call(handle)
		return callDefaultWindowProcedure(handle, message, wordParameter, longParameter)

	case windowMessageSize:
		if windowInstance.chromiumInstance != nil {
			windowInstance.chromiumInstance.Resize()
		}

	case windowMessageActivate:
		if wordParameter != 0 && windowInstance.chromiumInstance != nil {
			windowInstance.chromiumInstance.Focus()
		}

	case windowMessageSetCursor:
		return callDefaultWindowProcedure(handle, message, wordParameter, longParameter)

	case windowMessageClose:
		windowInstance.saveWindowState()
		destroyWindowProcedure.Call(handle)

	case windowMessageDestroy:
		windowInstance.saveWindowState()
		postQuitMessageProcedure.Call(0)

	case windowMessageGetMinimumMaximumInformation:
		windowInstance.applySizeConstraints(longParameter)

	default:
		return callDefaultWindowProcedure(handle, message, wordParameter, longParameter)
	}
	return 0
}

func (windowInstance *WindowsAppWindow) SetTitle(title string) {
	windowInstance.Dispatch(func() {
		titleUTF16, conversionError := windows.UTF16FromString(title)
		if conversionError != nil {
			titleUTF16, _ = windows.UTF16FromString("")
		}
		setWindowTextProcedure.Call(windowInstance.windowHandle, uintptr(unsafe.Pointer(&titleUTF16[0])))
	})
}

func (windowInstance *WindowsAppWindow) SetTheme(theme string) {
	windowInstance.Dispatch(func() {
		applyWindowTheme(windowInstance.windowHandle, theme)
		applyWebViewBackgroundColor(windowInstance.chromiumInstance, theme)
	})
}

func (windowInstance *WindowsAppWindow) NativeWindowHandle() uintptr {
	return windowInstance.windowHandle
}

func (windowInstance *WindowsAppWindow) Dispatch(function func()) {
	windowInstance.mutex.Lock()
	windowInstance.dispatchQueue = append(windowInstance.dispatchQueue, function)
	windowInstance.mutex.Unlock()
	postThreadMessageProcedure.Call(windowInstance.mainThreadIdentifier, uintptr(windowMessageApplication), 0, 0)
}

func (windowInstance *WindowsAppWindow) Init(script string) {
	windowInstance.chromiumInstance.Init(script)
}

func (windowInstance *WindowsAppWindow) Eval(script string) {
	windowInstance.chromiumInstance.Eval(script)
}

func (windowInstance *WindowsAppWindow) SetMinimumSize(width int32, height int32) {
	windowInstance.minimumSize = pointStructure{horizontalCoordinate: width, verticalCoordinate: height}
}

func (windowInstance *WindowsAppWindow) saveWindowState() {
	if windowInstance.hasSavedState || windowInstance.configurationManager == nil || windowInstance.windowHandle == 0 {
		return
	}
	windowInstance.hasSavedState = true

	placement, isValid := readWindowPlacement(windowInstance.windowHandle)
	if !isValid {
		return
	}
	_ = windowInstance.configurationManager.SetWindowState(windowStateFromPlacement(placement))
}

func (windowInstance *WindowsAppWindow) drainDispatchQueue() {
	windowInstance.mutex.Lock()
	queuedFunctions := windowInstance.dispatchQueue
	windowInstance.dispatchQueue = nil
	windowInstance.mutex.Unlock()

	for _, queuedFunction := range queuedFunctions {
		queuedFunction()
	}
}

func (windowInstance *WindowsAppWindow) Run() {
	var message messageStructure
	for {
		result, _, _ := getMessageProcedure.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if result == 0 || message.messageIdentifier == windowMessageQuit {
			return
		}

		if message.messageIdentifier == windowMessageApplication {
			windowInstance.drainDispatchQueue()
			continue
		}

		rootHandle, _, _ := getAncestorProcedure.Call(message.windowHandle, uintptr(getAncestorRoot))
		if isDialogResult, _, _ := isDialogMessageProcedure.Call(rootHandle, uintptr(unsafe.Pointer(&message))); isDialogResult != 0 {
			continue
		}

		translateMessageProcedure.Call(uintptr(unsafe.Pointer(&message)))
		dispatchMessageProcedure.Call(uintptr(unsafe.Pointer(&message)))
	}
}
