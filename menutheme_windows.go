//go:build windows

package main

import (
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	uxtheme         = syscall.NewLazyDLL("uxtheme.dll")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procGetProcAddr = kernel32.NewProc("GetProcAddress")
)

// Menu appearance modes of uxtheme's SetPreferredAppMode.
const (
	menuForceDark = 2
	menuForceLite = 3
)

// setMenuTheme makes Windows draw menus in the app's appearance. Windows
// has no documented API for it; uxtheme's SetPreferredAppMode and
// FlushMenuThemes, by ordinal since Windows 10 1903, do, and a system that
// lacks them keeps its menus as they are.
func setMenuTheme(dark bool) {
	if windows.RtlGetVersion().BuildNumber < 18362 {
		return
	}
	mode := uintptr(menuForceLite)
	if dark {
		mode = menuForceDark
	}
	if callOrdinal(135, mode) {
		callOrdinal(136)
	}
}

// callOrdinal calls the function at ordinal in uxtheme, and reports whether
// it found it.
func callOrdinal(ordinal uintptr, args ...uintptr) bool {
	addr, _, _ := procGetProcAddr.Call(uintptr(uxtheme.Handle()), ordinal)
	if addr == 0 {
		return false
	}
	syscall.SyscallN(addr, args...)
	return true
}
