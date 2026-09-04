//go:build windows

package window

import (
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	userInterfaceLibrary             = syscall.NewLazyDLL("user32.dll")
	desktopWindowManagerLibrary      = syscall.NewLazyDLL("dwmapi.dll")
	graphicsDeviceInterfaceLibrary   = syscall.NewLazyDLL("gdi32.dll")
	kernelLibrary                    = syscall.NewLazyDLL("kernel32.dll")
	sendMessageProcedure             = userInterfaceLibrary.NewProc("SendMessageW")
	postQuitMessageProcedure         = userInterfaceLibrary.NewProc("PostQuitMessage")
	postThreadMessageProcedure       = userInterfaceLibrary.NewProc("PostThreadMessageW")
	getMessageProcedure              = userInterfaceLibrary.NewProc("GetMessageW")
	translateMessageProcedure        = userInterfaceLibrary.NewProc("TranslateMessage")
	dispatchMessageProcedure         = userInterfaceLibrary.NewProc("DispatchMessageW")
	defaultWindowProcedureProcedure  = userInterfaceLibrary.NewProc("DefWindowProcW")
	registerClassExtendedProcedure   = userInterfaceLibrary.NewProc("RegisterClassExW")
	createWindowExtendedProcedure    = userInterfaceLibrary.NewProc("CreateWindowExW")
	destroyWindowProcedure           = userInterfaceLibrary.NewProc("DestroyWindow")
	setWindowTextProcedure           = userInterfaceLibrary.NewProc("SetWindowTextW")
	setFocusProcedure                = userInterfaceLibrary.NewProc("SetFocus")
	getSystemMetricsProcedure        = userInterfaceLibrary.NewProc("GetSystemMetrics")
	createSolidBrushProcedure        = graphicsDeviceInterfaceLibrary.NewProc("CreateSolidBrush")
	desktopWindowManagerSetAttribute = desktopWindowManagerLibrary.NewProc("DwmSetWindowAttribute")
	getCurrentThreadIdentifier       = kernelLibrary.NewProc("GetCurrentThreadId")
	fillRectangleProcedure           = userInterfaceLibrary.NewProc("FillRect")
	getClientRectangleProcedure      = userInterfaceLibrary.NewProc("GetClientRect")
	loadCursorProcedure              = userInterfaceLibrary.NewProc("LoadCursorW")
	getAncestorProcedure             = userInterfaceLibrary.NewProc("GetAncestor")
	isDialogMessageProcedure         = userInterfaceLibrary.NewProc("IsDialogMessageW")
	createIconFromResourceExtended   = userInterfaceLibrary.NewProc("CreateIconFromResourceEx")
	getWindowPlacementProcedure      = userInterfaceLibrary.NewProc("GetWindowPlacement")
	setWindowPlacementProcedure      = userInterfaceLibrary.NewProc("SetWindowPlacement")
	monitorFromRectProcedure         = userInterfaceLibrary.NewProc("MonitorFromRect")
	getMonitorInfoProcedure          = userInterfaceLibrary.NewProc("GetMonitorInfoW")
)

const (
	windowMessageClose                        = 0x0010
	windowMessageDestroy                      = 0x0002
	windowMessageSize                         = 0x0005
	windowMessageMove                         = 0x0003
	windowMessageMoving                       = 0x0216
	windowMessageActivate                     = 0x0006
	windowMessageApplication                  = 0x8000
	windowMessageQuit                         = 0x0012
	windowMessageNonClientLeftButton          = 0x00A1
	windowMessageGetMinimumMaximumInformation = 0x0024
	windowMessageEraseBackground              = 0x0014
	windowMessageSetCursor                    = 0x0020
	windowMessageSetIcon                      = 0x0080
	iconSmallIdentifier                       = 0
	iconBigIdentifier                         = 1
	windowStyleOverlappedWindow               = 0x00CF0000
	showWindowNormal                          = 1
	showWindowMinimize                        = 2
	showWindowMaximize                        = 3
	windowPlacementFlagRestoreToMaximized     = 2
	monitorDefaultToNull                      = 0
	getAncestorRoot                           = 2
	cursorStandardArrowIdentifier             = 32512
	systemMetricScreenWidth                   = 0
	systemMetricScreenHeight                  = 1
	defaultMinimumWindowWidth                 = int32(960)
	defaultMinimumWindowHeight                = int32(540)
	defaultRestoredWindowWidth                = int32(1280)
	defaultRestoredWindowHeight               = int32(720)
	windowClassName                           = "OrbitMainWindowClass"
	windowTitle                               = "Orbit"
)

type pointStructure struct {
	horizontalCoordinate int32
	verticalCoordinate   int32
}

type minimumMaximumInformationStructure struct {
	reservedPoint         pointStructure
	maximumSizePoint      pointStructure
	maximumPositionPoint  pointStructure
	minimumTrackSizePoint pointStructure
	maximumTrackSizePoint pointStructure
}

type rectangleStructure struct {
	leftCoordinate   int32
	topCoordinate    int32
	rightCoordinate  int32
	bottomCoordinate int32
}

type messageStructure struct {
	windowHandle         uintptr
	messageIdentifier    uint32
	wordParameter        uintptr
	longParameter        uintptr
	time                 uint32
	point                pointStructure
	privateLongParameter uint32
}

type windowClassExtendedStructure struct {
	structureSize         uint32
	style                 uint32
	windowProcedure       uintptr
	classExtraBytes       int32
	windowExtraBytes      int32
	instanceHandle        windows.Handle
	iconHandle            windows.Handle
	cursorHandle          windows.Handle
	backgroundBrushHandle uintptr
	menuName              *uint16
	className             *uint16
	smallIconHandle       windows.Handle
}

type windowPlacementStructure struct {
	structureLength         uint32
	flags                   uint32
	showCommand             uint32
	minimumPositionPoint    pointStructure
	maximumPositionPoint    pointStructure
	normalPositionRectangle rectangleStructure
}

type monitorInformationStructure struct {
	structureSize     uint32
	monitorRectangle  rectangleStructure
	workAreaRectangle rectangleStructure
	flags             uint32
}

func (rectangle rectangleStructure) width() int32 {
	return rectangle.rightCoordinate - rectangle.leftCoordinate
}

func (rectangle rectangleStructure) height() int32 {
	return rectangle.bottomCoordinate - rectangle.topCoordinate
}
