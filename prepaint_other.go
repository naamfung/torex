//go:build !windows

package main

import "github.com/egoist/mygo"

// prepaint does nothing where the window does not show white before its
// first frame.
func prepaint(*mygo.Window) {}
