//go:build windows

package window

import (
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
)

const (
	dwmAttributeImmersiveDarkMode       = 20
	dwmAttributeImmersiveDarkModeLegacy = 19
	dwmAttributeBorderColor             = 34
	dwmAttributeCaptionColor            = 35
	dwmAttributeTextColor               = 36
	darkThemeColorRGB                   = 0x00161616
	lightThemeColorRGB                  = 0x00F7F5F5
	lightTextColorRGB                   = 0x00FFFFFF
	darkTextColorRGB                    = 0x001B1818
	lightThemeName                      = "light"
)

var (
	darkWebViewBackgroundColor  = edge.COREWEBVIEW2_COLOR{A: 0xFF, R: 0x16, G: 0x16, B: 0x16}
	lightWebViewBackgroundColor = edge.COREWEBVIEW2_COLOR{A: 0xFF, R: 0xF5, G: 0xF5, B: 0xF7}
)

type frameTheme struct {
	darkModeFlag int32
	frameColor   uint32
	textColor    uint32
}

func frameThemeFor(theme string) frameTheme {
	if theme == lightThemeName {
		return frameTheme{darkModeFlag: 0, frameColor: lightThemeColorRGB, textColor: darkTextColorRGB}
	}
	return frameTheme{darkModeFlag: 1, frameColor: darkThemeColorRGB, textColor: lightTextColorRGB}
}

func setDesktopWindowManagerAttribute(windowHandle uintptr, attribute uintptr, value unsafe.Pointer, valueSize uintptr) {
	desktopWindowManagerSetAttribute.Call(windowHandle, attribute, uintptr(value), valueSize)
}

func applyWindowTheme(windowHandle uintptr, theme string) {
	if windowHandle == 0 {
		return
	}

	frame := frameThemeFor(theme)
	captionColor := frame.frameColor
	borderColor := frame.frameColor
	textColor := frame.textColor

	setDesktopWindowManagerAttribute(windowHandle, dwmAttributeImmersiveDarkMode, unsafe.Pointer(&frame.darkModeFlag), unsafe.Sizeof(frame.darkModeFlag))
	setDesktopWindowManagerAttribute(windowHandle, dwmAttributeImmersiveDarkModeLegacy, unsafe.Pointer(&frame.darkModeFlag), unsafe.Sizeof(frame.darkModeFlag))
	setDesktopWindowManagerAttribute(windowHandle, dwmAttributeCaptionColor, unsafe.Pointer(&captionColor), unsafe.Sizeof(captionColor))
	setDesktopWindowManagerAttribute(windowHandle, dwmAttributeBorderColor, unsafe.Pointer(&borderColor), unsafe.Sizeof(borderColor))
	setDesktopWindowManagerAttribute(windowHandle, dwmAttributeTextColor, unsafe.Pointer(&textColor), unsafe.Sizeof(textColor))
}

func webViewBackgroundColorFor(theme string) edge.COREWEBVIEW2_COLOR {
	if theme == lightThemeName {
		return lightWebViewBackgroundColor
	}
	return darkWebViewBackgroundColor
}

func applyWebViewBackgroundColor(chromiumInstance *edge.Chromium, theme string) {
	if chromiumInstance == nil {
		return
	}
	controllerInstance := chromiumInstance.GetController()
	if controllerInstance == nil {
		return
	}
	controller2Instance := controllerInstance.GetICoreWebView2Controller2()
	if controller2Instance == nil {
		return
	}
	_ = controller2Instance.PutDefaultBackgroundColor(webViewBackgroundColorFor(theme))
}
