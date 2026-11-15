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
