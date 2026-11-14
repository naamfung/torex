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
)

//go:embed all:runtime
var runtime embed.FS

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
		version = "busybox-" + hex.EncodeToString(h.Sum(nil)[:6])
	})
	return version
}

// Shell returns the program of the runtime's shell, unpacked in base when
// it is not there: base\version\bin\busybox.exe.
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
	return shell, true
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
