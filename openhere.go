package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"torex/internal/rex"
)

// folderArg returns the folder an argument names: Explorer hands the
// context menu the folder it was opened in.
func folderArg(args []string) string {
	for _, a := range args {
		if st, err := os.Stat(a); err == nil && st.IsDir() {
			// The menu passes the folder with a trailing slash, and a
			// drive's root keeps its backslash: Clean drops the one and
			// leaves the root whole.
			return filepath.Clean(a)
		}
	}
	return ""
}

// openDir returns the directory where a window that runs is asked to open
// folders, or to show itself.
func openDir() string { return filepath.Join(rex.Dir(), "open") }

// requestOpen asks the window that runs to open dir, or to show itself when
// dir is empty.
func requestOpen(dir string) error {
	d := openDir()
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	tmp := filepath.Join(d, "."+name)
	if err := os.WriteFile(tmp, []byte(dir), 0o644); err != nil {
		return err
	}
	// A request is renamed into place, whole, for the window reading it.
	return os.Rename(tmp, filepath.Join(d, name))
}

// takeOpen returns the folders that the requests ask to open, and removes
// them. An empty one asks for the window, as a second launch does.
func takeOpen() []string {
	entries, err := os.ReadDir(openDir())
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(openDir(), e.Name())
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		os.Remove(p)
		dirs = append(dirs, strings.TrimSpace(string(b)))
	}
	return dirs
}
