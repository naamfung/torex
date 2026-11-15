//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"github.com/egoist/mygo"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	procEnumChildWindows = user32.NewProc("EnumChildWindows")
	procGetClassNameW    = user32.NewProc("GetClassNameW")
	procGetDC            = user32.NewProc("GetDC")
	procReleaseDC        = user32.NewProc("ReleaseDC")
	procGetClientRect    = user32.NewProc("GetClientRect")
	procFillRect         = user32.NewProc("FillRect")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject     = gdi32.NewProc("DeleteObject")

	// surface is the child that the enumeration in prepaint finds.
	surface     uintptr
	surfaceProc uintptr
)

// prepaint fills the window's surface with the window's background. MyGo
// draws the surface with the GPU from its first frame on; until that frame
// Windows has nothing to paint there, and the surface is white: a black
// window would flash white as it starts.
func prepaint(win *mygo.Window) {
	hwnd := win.NativeHandle()
	if hwnd == 0 {
		return
	}
	if surfaceProc == 0 {
		surfaceProc = syscall.NewCallback(func(child, _ uintptr) uintptr {
			var name [32]uint16
			procGetClassNameW.Call(child, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
			// MyGo's surface, not the child a caption bar makes.
			if syscall.UTF16ToString(name[:]) != "MyGoSurface" {
				return 1
			}
			surface = child
			return 0
		})
	}
	surface = 0
	procEnumChildWindows.Call(hwnd, surfaceProc, 0)
	if surface == 0 {
		return
	}
	dc, _, _ := procGetDC.Call(surface)
	if dc == 0 {
		return
	}
	defer procReleaseDC.Call(surface, dc)
	var r struct{ Left, Top, Right, Bottom int32 }
	procGetClientRect.Call(surface, uintptr(unsafe.Pointer(&r)))
	// The window's background colors, light and dark, as COLORREF.
	color := uintptr(0x121010) // #101012
	if !mygo.Theme.IsDark() {
		color = 0xe6e1ef // #efe1e6
	}
	brush, _, _ := procCreateSolidBrush.Call(color)
	if brush == 0 {
		return
	}
	defer procDeleteObject.Call(brush)
	procFillRect.Call(dc, uintptr(unsafe.Pointer(&r)), brush)
}
