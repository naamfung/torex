//go:build windows

package rex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"torex/internal/winkit"
)

// TestWinkitLoginShell runs the runtime's login shell: its home is the
// Windows user's own, and the mark of its first prompt is made in the user's
// temporary directory, which exists, rather than a /tmp that the shell has
// not.
func TestWinkitLoginShell(t *testing.T) {
	base := t.TempDir()
	sh, ok := winkit.Shell(base)
	if !ok {
		t.Skip("no runtime to unpack")
	}
	bb, argv, env, ok := winkitShell(sh, "")
	if !ok {
		t.Skip("not a winkit shell")
	}
	home, torexHome, path := "", "", ""
	for _, kv := range env {
		switch {
		case strings.HasPrefix(kv, "HOME="):
			home = strings.TrimPrefix(kv, "HOME=")
		case strings.HasPrefix(kv, "TOREX_HOME="):
			torexHome = strings.TrimPrefix(kv, "TOREX_HOME=")
		case strings.HasPrefix(kv, "PATH="):
			path = strings.TrimPrefix(kv, "PATH=")
		}
	}
	want, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	if !strings.EqualFold(filepath.Clean(home), filepath.Clean(want)) {
		t.Fatalf("HOME = %q, want %q", home, want)
	}
	// The programs of the installation come first, where its owner puts them.
	install, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	install = filepath.Dir(install)
	if !strings.EqualFold(torexHome, install) {
		t.Fatalf("TOREX_HOME = %q, want %q", torexHome, install)
	}
	if first, _, _ := strings.Cut(path, ";"); !strings.EqualFold(filepath.Clean(first), filepath.Join(install, "bin")) {
		t.Fatalf("PATH starts with %q, want %q", first, filepath.Join(install, "bin"))
	}

	p, err := startPTY(bb, argv, base, append(os.Environ(), env...), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()

	printed := make(chan string)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := p.read(buf)
			if n > 0 {
				printed <- string(buf[:n])
			}
			if err != nil {
				close(printed)
				return
			}
		}
	}()
	var out strings.Builder
	deadline := time.After(10 * time.Second)
read:
	for {
		select {
		case s, ok := <-printed:
			if !ok {
				break read
			}
			out.WriteString(s)
			if strings.Contains(out.String(), "#") {
				break read
			}
		case <-deadline:
			break read
		}
	}
	if strings.Contains(out.String(), "can't create") {
		t.Fatalf("the prompt mark could not be made: %q", out.String())
	}
	mark := filepath.Join(os.TempDir(), fmt.Sprintf(".torex-prompt-%d", p.pid()))
	if _, err := os.Stat(mark); err != nil {
		t.Fatalf("no %s: %q", mark, out.String())
	}
	p.write([]byte("exit\r"))
}
