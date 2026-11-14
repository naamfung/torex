//go:build windows

package winkit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRemoveOld checks that the runtimes of other builds go, but not the
// one of this build, one being unpacked, or a directory that is no runtime.
func TestRemoveOld(t *testing.T) {
	base := t.TempDir()
	old := time.Now().Add(-time.Hour)
	mk := func(name string, mod time.Time) string {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		os.Chtimes(dir, mod, mod)
		return dir
	}
	mk("winkit-keep", old)
	mk("winkit-old", old)
	mk("busybox-old", old)
	mk("winkit-fresh", time.Now())
	mk("winkit-being.tmp", old)
	mk("other", old)

	removeOld(base, "winkit-keep")

	for _, name := range []string{"winkit-keep", "winkit-fresh", "winkit-being.tmp", "other"} {
		if _, err := os.Stat(filepath.Join(base, name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, name := range []string{"winkit-old", "busybox-old"} {
		if _, err := os.Stat(filepath.Join(base, name)); !os.IsNotExist(err) {
			t.Errorf("%s should be removed: %v", name, err)
		}
	}
}
