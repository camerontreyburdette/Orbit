//go:build windows

package window

import (
	"unsafe"

	"orbit/internal/configuration"
)

type windowPlacement struct {
	horizontalCoordinate int32
	verticalCoordinate   int32
	width                int32
	height               int32
	isMaximized          bool
}

func (placement windowPlacement) rectangle() rectangleStructure {
	return rectangleStructure{
		leftCoordinate:   placement.horizontalCoordinate,
		topCoordinate:    placement.verticalCoordinate,
		rightCoordinate:  placement.horizontalCoordinate + placement.width,
		bottomCoordinate: placement.verticalCoordinate + placement.height,
	}
}

func (placement windowPlacement) showCommand() uint32 {
	if placement.isMaximized {
		return showWindowMaximize
	}
	return showWindowNormal
}

func clampMinimumSize(width int32, height int32) (int32, int32) {
	if width < defaultMinimumWindowWidth {
		width = defaultMinimumWindowWidth
	}
	if height < defaultMinimumWindowHeight {
		height = defaultMinimumWindowHeight
	}
	return width, height
}

func defaultWindowPlacement() windowPlacement {
	screenWidthValue, _, _ := getSystemMetricsProcedure.Call(uintptr(systemMetricScreenWidth))
	screenHeightValue, _, _ := getSystemMetricsProcedure.Call(uintptr(systemMetricScreenHeight))

	placement := windowPlacement{width: defaultRestoredWindowWidth, height: defaultRestoredWindowHeight}
	screenWidth := int32(screenWidthValue)
	screenHeight := int32(screenHeightValue)
	if screenWidth > placement.width && screenHeight > placement.height {
		placement.horizontalCoordinate = (screenWidth - placement.width) / 2
		placement.verticalCoordinate = (screenHeight - placement.height) / 2
	}
	return placement
}

func monitorWorkAreaFor(rectangle rectangleStructure) (rectangleStructure, bool) {
	monitorHandle, _, _ := monitorFromRectProcedure.Call(uintptr(unsafe.Pointer(&rectangle)), uintptr(monitorDefaultToNull))
	if monitorHandle == 0 {
		return rectangleStructure{}, false
	}

	var monitorInformation monitorInformationStructure
	monitorInformation.structureSize = uint32(unsafe.Sizeof(monitorInformation))
	informationResult, _, _ := getMonitorInfoProcedure.Call(monitorHandle, uintptr(unsafe.Pointer(&monitorInformation)))
	if informationResult == 0 {
		return rectangleStructure{}, false
	}
	return monitorInformation.workAreaRectangle, true
}

func fitPlacementToWorkArea(placement windowPlacement, workArea rectangleStructure) windowPlacement {
	if placement.width > workArea.width() {
		placement.width = workArea.width()
	}
	if placement.height > workArea.height() {
		placement.height = workArea.height()
	}
	if placement.horizontalCoordinate < workArea.leftCoordinate {
		placement.horizontalCoordinate = workArea.leftCoordinate
	}
	if placement.horizontalCoordinate+placement.width > workArea.rightCoordinate {
		placement.horizontalCoordinate = workArea.rightCoordinate - placement.width
	}
	if placement.verticalCoordinate < workArea.topCoordinate {
		placement.verticalCoordinate = workArea.topCoordinate
	}
	if placement.verticalCoordinate+placement.height > workArea.bottomCoordinate {
		placement.verticalCoordinate = workArea.bottomCoordinate - placement.height
	}
	return placement
}

func determineInitialWindowPlacement(configurationManager *configuration.Manager) windowPlacement {
	defaultPlacement := defaultWindowPlacement()
	if configurationManager == nil {
		return defaultPlacement
	}

	savedWindowState := configurationManager.GetWindowState()
	if savedWindowState == nil {
		return defaultPlacement
	}

	width, height := clampMinimumSize(savedWindowState.Width, savedWindowState.Height)
	savedPlacement := windowPlacement{
		horizontalCoordinate: savedWindowState.HorizontalCoordinate,
		verticalCoordinate:   savedWindowState.VerticalCoordinate,
		width:                width,
		height:               height,
		isMaximized:          savedWindowState.IsMaximized,
	}

	workArea, hasWorkArea := monitorWorkAreaFor(savedPlacement.rectangle())
	if !hasWorkArea {
		savedPlacement.horizontalCoordinate = defaultPlacement.horizontalCoordinate
		savedPlacement.verticalCoordinate = defaultPlacement.verticalCoordinate
		return savedPlacement
	}
	return fitPlacementToWorkArea(savedPlacement, workArea)
}

func readWindowPlacement(windowHandle uintptr) (windowPlacementStructure, bool) {
	var placement windowPlacementStructure
	placement.structureLength = uint32(unsafe.Sizeof(placement))
	placementResult, _, _ := getWindowPlacementProcedure.Call(windowHandle, uintptr(unsafe.Pointer(&placement)))
	return placement, placementResult != 0
}

func isPlacementMaximized(placement windowPlacementStructure) bool {
	if placement.showCommand == showWindowMaximize {
		return true
	}
	return placement.showCommand == showWindowMinimize && placement.flags&windowPlacementFlagRestoreToMaximized != 0
}

func windowStateFromPlacement(placement windowPlacementStructure) configuration.WindowState {
	normalRectangle := placement.normalPositionRectangle
	width, height := clampMinimumSize(normalRectangle.width(), normalRectangle.height())
	return configuration.WindowState{
		IsMaximized:          isPlacementMaximized(placement),
		HorizontalCoordinate: normalRectangle.leftCoordinate,
		VerticalCoordinate:   normalRectangle.topCoordinate,
		Width:                width,
		Height:               height,
	}
}

func applyInitialWindowPlacement(windowHandle uintptr, placement windowPlacement) {
	nativePlacement := windowPlacementStructure{
		structureLength:         uint32(unsafe.Sizeof(windowPlacementStructure{})),
		showCommand:             placement.showCommand(),
		normalPositionRectangle: placement.rectangle(),
	}
	setWindowPlacementProcedure.Call(windowHandle, uintptr(unsafe.Pointer(&nativePlacement)))
}
