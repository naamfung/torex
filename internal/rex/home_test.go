//go:build windows

package rex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"torex/internal/winkit"
)

func TestWinkitShellHome(t *testing.T) {
	sh, ok := winkit.Shell(t.TempDir())
	if !ok {
		t.Skip("no runtime to unpack")
	}
	_, _, env, ok := winkitShell(sh, "")
	if !ok {
		t.Fatal("the unpacked shell is not a WinKit shell")
	}
	home := ""
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, "HOME="); ok {
			home = v
		}
	}
	want, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	if !strings.EqualFold(filepath.Clean(home), filepath.Clean(want)) {
		t.Fatalf("HOME = %q, want %q", home, want)
	}
}
