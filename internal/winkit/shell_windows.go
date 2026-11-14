//go:build windows

package winkit

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

//go:embed all:runtime
var runtime embed.FS

// prefix names the directory of a runtime: an earlier build named it
// busybox, as the shell was all it had.
const prefix = "winkit-"

// version names the runtime by the checksum of its files, which change
// with WinKit's shell and the profiles: a copy of another runtime is
// unpacked anew.
var (
	versionOnce sync.Once
	version     string
)

func runtimeVersion() string {
	versionOnce.Do(func() {
		h := sha256.New()
		fs.WalkDir(runtime, "runtime", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			b, err := runtime.ReadFile(p)
			if err != nil {
				return err
			}
			h.Write([]byte(p))
			h.Write(b)
			return nil
		})
		version = prefix + hex.EncodeToString(h.Sum(nil)[:6])
	})
	return version
}

// Shell returns the program of the runtime's shell, unpacked in base when
// it is not there: base\version\bin\busybox.exe. The runtimes of other
// builds go once nothing runs from them.
func Shell(base string) (string, bool) {
	dir := filepath.Join(base, runtimeVersion())
	shell := filepath.Join(dir, "bin", "busybox.exe")
	if _, err := os.Stat(shell); err != nil {
		if unpack(dir) != nil {
			return "", false
		}
	}
	if _, err := os.Stat(shell); err != nil {
		return "", false
	}
	removeOld(base, filepath.Base(dir))
	return shell, true
}

// removeOld removes the runtimes of base that nothing runs from: a build
// unpacks beside the runtime of another, whose panes may still run it.
func removeOld(base, keep string) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || name == keep || strings.HasSuffix(name, ".tmp") || !isRuntime(name) {
			continue
		}
		// A runtime another window unpacked may not have a shell yet.
		if info, err := e.Info(); err != nil || time.Since(info.ModTime()) < time.Minute {
			continue
		}
		dir := filepath.Join(base, name)
		if running(dir) {
			continue
		}
		os.RemoveAll(dir)
	}
}

// isRuntime tells that the name is one of a runtime, as this build and
// earlier ones name them.
func isRuntime(name string) bool {
	return strings.HasPrefix(name, prefix) || strings.HasPrefix(name, "busybox-")
}

// running tells that a shell of dir runs: Windows keeps an image from being
// opened without sharing.
func running(dir string) bool {
	p, err := windows.UTF16PtrFromString(filepath.Join(dir, "bin", "busybox.exe"))
	if err != nil {
		return true
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return !os.IsNotExist(err)
	}
	windows.CloseHandle(h)
	return false
}

// unpack writes the runtime into dir, through a directory beside it, so
// that a copy is whole or not there.
func unpack(dir string) error {
	tmp := dir + ".tmp"
	os.RemoveAll(tmp)
	err := fs.WalkDir(runtime, "runtime", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, "runtime")
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		target := filepath.Join(tmp, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := runtime.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o755)
	})
	if err != nil {
		os.RemoveAll(tmp)
		return err
	}
	if _, err := os.Stat(dir); err == nil {
		// Another window unpacked it meanwhile.
		os.RemoveAll(tmp)
		return nil
	}
	// Another window may unpack it as this one does: the first to rename
	// it wins, and the other takes what it wrote.
	if err := os.Rename(tmp, dir); err != nil {
		os.RemoveAll(tmp)
		if _, statErr := os.Stat(filepath.Join(dir, "bin", "busybox.exe")); statErr != nil {
			return err
		}
	}
	return nil
}
