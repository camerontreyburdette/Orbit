//go:build windows

package window

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"

	"orbit/internal/api"
	"orbit/internal/assets"
	"orbit/internal/configuration"
)

const (
	smallIconSize = 16
	bigIconSize   = 32
)

const initializationScript = `
(function() {
  window.pywebview = {
    api: new Proxy({}, {
      get: function(target, property) {
        return function(...argumentsList) {
          return window.__orbit_invoke(property, JSON.stringify(argumentsList));
        };
      }
    })
  };
  window.addEventListener('DOMContentLoaded', function() {
    window.dispatchEvent(new CustomEvent('pywebviewready'));
    if (window.__orbit_ready) {
      window.__orbit_ready();
    }
  });
})();
`

type nativeWindowResources struct {
	instanceHandle  windows.Handle
	smallIconHandle uintptr
	bigIconHandle   uintptr
	brushHandle     uintptr
}

func registerWindowClass(resources nativeWindowResources) *uint16 {
	cursorHandle, _, _ := loadCursorProcedure.Call(0, uintptr(cursorStandardArrowIdentifier))
	className, _ := windows.UTF16PtrFromString(windowClassName)
	windowClass := windowClassExtendedStructure{
		structureSize:         uint32(unsafe.Sizeof(windowClassExtendedStructure{})),
		instanceHandle:        resources.instanceHandle,
		className:             className,
		iconHandle:            windows.Handle(resources.bigIconHandle),
		smallIconHandle:       windows.Handle(resources.smallIconHandle),
		cursorHandle:          windows.Handle(cursorHandle),
		backgroundBrushHandle: resources.brushHandle,
		windowProcedure:       windows.NewCallback(windowProcedure),
	}
	registerClassExtendedProcedure.Call(uintptr(unsafe.Pointer(&windowClass)))
	return className
}

func createNativeWindow(className *uint16, resources nativeWindowResources, placement windowPlacement) uintptr {
	titlePointer, _ := windows.UTF16PtrFromString(windowTitle)
	windowHandle, _, _ := createWindowExtendedProcedure.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titlePointer)),
		uintptr(windowStyleOverlappedWindow),
		uintptr(placement.horizontalCoordinate),
		uintptr(placement.verticalCoordinate),
		uintptr(placement.width),
		uintptr(placement.height),
		0,
		0,
		uintptr(resources.instanceHandle),
		0,
	)
	return windowHandle
}

func applyWindowIcons(windowHandle uintptr, resources nativeWindowResources) {
	if resources.smallIconHandle != 0 {
		sendMessageProcedure.Call(windowHandle, uintptr(windowMessageSetIcon), uintptr(iconSmallIdentifier), resources.smallIconHandle)
	}
	if resources.bigIconHandle != 0 {
		sendMessageProcedure.Call(windowHandle, uintptr(windowMessageSetIcon), uintptr(iconBigIdentifier), resources.bigIconHandle)
	}
}

func loadNativeWindowResources() nativeWindowResources {
	var instanceHandle windows.Handle
	_ = windows.GetModuleHandleEx(0, nil, &instanceHandle)
	brushHandle, _, _ := createSolidBrushProcedure.Call(uintptr(darkThemeColorRGB))
	return nativeWindowResources{
		instanceHandle:  instanceHandle,
		smallIconHandle: createIconFromIconData(assets.ApplicationIconBytes, smallIconSize, smallIconSize),
		bigIconHandle:   createIconFromIconData(assets.ApplicationIconBytes, bigIconSize, bigIconSize),
		brushHandle:     brushHandle,
	}
}

func embedChromium(appWindow *WindowsAppWindow, initialTheme string, debugMode bool) error {
	chromiumInstance := edge.NewChromium()
	chromiumInstance.DataPath = filepath.Join(os.Getenv("AppData"), "Orbit")
	chromiumInstance.MessageCallback = appWindow.handleMessageCallback
	chromiumInstance.SetPermission(edge.CoreWebView2PermissionKindClipboardRead, edge.CoreWebView2PermissionStateAllow)
	appWindow.chromiumInstance = chromiumInstance

	if !chromiumInstance.Embed(appWindow.windowHandle) {
		return fmt.Errorf("failed to embed chromium in application window")
	}

	applyWebViewBackgroundColor(chromiumInstance, initialTheme)
	if controllerInstance := chromiumInstance.GetController(); controllerInstance != nil {
		_ = controllerInstance.PutIsVisible(true)
	}

	if settings, settingsError := chromiumInstance.GetSettings(); settingsError == nil && settings != nil {
		_ = settings.PutAreDefaultContextMenusEnabled(debugMode)
		_ = settings.PutAreDevToolsEnabled(debugMode)
	}

	chromiumInstance.Resize()
	return nil
}

func bindApplicationBridge(appWindow *WindowsAppWindow, handler *api.Handler) error {
	appWindow.Init(initializationScript)

	bindError := appWindow.Bind("__orbit_invoke", func(methodName string, argumentsJSON string) (interface{}, error) {
		var rawArguments []json.RawMessage
		if argumentsJSON != "" {
			_ = json.Unmarshal([]byte(argumentsJSON), &rawArguments)
		}

		result, executionError := handler.HandleMethodCall(methodName, rawArguments)
		if executionError != nil {
			return map[string]interface{}{"error": executionError.Error()}, nil
		}
		return result, nil
	})
	if bindError != nil {
		return fmt.Errorf("failed to bind rpc bridge: %w", bindError)
	}

	windowHandle := appWindow.windowHandle
	_ = appWindow.Bind("__orbit_ready", func() (interface{}, error) {
		sendMessageProcedure.Call(windowHandle, uintptr(windowMessageSize), 0, 0)
		return nil, nil
	})
	return nil
}

func StartAppWindow(serverUniformResourceLocator string, handler *api.Handler, configurationManager *configuration.Manager, debugMode bool) error {
	placement := determineInitialWindowPlacement(configurationManager)
	resources := loadNativeWindowResources()
	className := registerWindowClass(resources)

	windowHandle := createNativeWindow(className, resources, placement)
	if windowHandle == 0 {
		return fmt.Errorf("failed to create application window")
	}
	applyWindowIcons(windowHandle, resources)

	initialTheme := handler.GetTheme()
	applyWindowTheme(windowHandle, initialTheme)

	appWindow := &WindowsAppWindow{
		windowHandle:         windowHandle,
		configurationManager: configurationManager,
		bindings:             make(map[string]boundFunction),
		backgroundBrush:      resources.brushHandle,
	}
	appWindow.mainThreadIdentifier, _, _ = getCurrentThreadIdentifier.Call()
	appWindow.SetMinimumSize(defaultMinimumWindowWidth, defaultMinimumWindowHeight)

	registerApplicationWindow(windowHandle, appWindow)
	defer unregisterApplicationWindow(windowHandle)

	if embedError := embedChromium(appWindow, initialTheme, debugMode); embedError != nil {
		return embedError
	}

	handler.SetWindowController(appWindow)
	if bridgeError := bindApplicationBridge(appWindow, handler); bridgeError != nil {
		return bridgeError
	}

	applyInitialWindowPlacement(windowHandle, placement)
	sendMessageProcedure.Call(windowHandle, uintptr(windowMessageSize), 0, 0)

	appWindow.chromiumInstance.Navigate(serverUniformResourceLocator)
	appWindow.Run()
	return nil
}
