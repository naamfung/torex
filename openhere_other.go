//go:build !windows

package main

// installMenu does nothing: only Windows has the context menu entry.
func installMenu() {}

// uninstallMenu does nothing: only Windows has the context menu entry.
func uninstallMenu() {}

// holdWindow does nothing: the system runs one instance of the app.
func holdWindow() (func(), bool) { return func() {}, true }
