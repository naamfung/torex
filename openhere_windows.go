//go:build windows

package main

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"torex/internal/rex"
)

// menuKeys are where Explorer reads the entry that starts torex in a folder:
// the background of one, and the folder itself.
var menuKeys = []string{
	`Software\Classes\Directory\Background\shell\torex`,
	`Software\Classes\Directory\shell\torex`,
}

const menuText = "在此目录启动 torex"

// installMenu points Explorer's context menu entry at this program, as
// WinKit's install does.
func installMenu() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	for _, key := range menuKeys {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, key, registry.WRITE)
		if err != nil {
			continue
		}
		k.SetStringValue("", menuText)
		k.SetStringValue("Icon", exe+",0")
		k.Close()
		command, _, err := registry.CreateKey(registry.CURRENT_USER, key+`\command`, registry.WRITE)
		if err != nil {
			continue
		}
		// A drive's root ends in a backslash, which would escape the
		// closing quote: a slash before it keeps the quotes and passes
		// C:\ whole, and folderArg cleans it back to a Windows path.
		command.SetStringValue("", `"`+exe+`" "%V/"`)
		command.Close()
	}
}

// uninstallMenu removes the entry, as WinKit's uninstall does.
func uninstallMenu() {
	for _, key := range menuKeys {
		registry.DeleteKey(registry.CURRENT_USER, key+`\command`)
		registry.DeleteKey(registry.CURRENT_USER, key)
	}
}

// holdWindow takes the lock of the window of this data directory: the
// window that runs opens the folders other launches ask for.
func holdWindow() (func(), bool) {
	if err := os.MkdirAll(rex.Dir(), 0o755); err != nil {
		return func() {}, true
	}
	p, err := windows.UTF16PtrFromString(filepath.Join(rex.Dir(), "window.lock"))
	if err != nil {
		return func() {}, true
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil,
		windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return func() {}, false
	}
	return func() { windows.CloseHandle(h) }, true
}
