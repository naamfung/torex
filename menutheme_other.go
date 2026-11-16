//go:build !windows

package main

// setMenuTheme does nothing where menus follow the app's appearance on
// their own.
func setMenuTheme(bool) {}
