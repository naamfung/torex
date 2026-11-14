// Command mkssh builds ssh, the ssh client of GoRex's Windows runtime, where
// GoRex embeds it beside the shell: the runtime of a pane has no ssh of
// Windows, and this one comes with the app.
//
//	go run ./tools/mkssh
//
// With -app it builds GoRex after the program, as mkbusybox and mkvim do.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const programPath = "internal/winkit/runtime/bin/ssh.exe"

func main() {
	app := flag.Bool("app", false, "build GoRex after the program")
	flag.Parse()
	root, err := moduleRoot()
	if err != nil {
		fatal(err)
	}
	build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(root, programPath), "./cmd/ssh")
	build.Dir = root
	// The runtime of Windows panes is x86-64, as is the shell beside it.
	build.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		fatal(fmt.Errorf("go build: %w", err))
	}
	st, err := os.Stat(filepath.Join(root, programPath))
	if err != nil {
		fatal(err)
	}
	fmt.Printf("mkssh: %s (%d bytes)\n", programPath, st.Size())
	if *app {
		fmt.Println("mkssh: go tool mygo build")
		cmd := exec.Command("go", "tool", "mygo", "build")
		cmd.Dir = root
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fatal(fmt.Errorf("go tool mygo build: %w", err))
		}
	}
}

// moduleRoot returns the directory of the go.mod of GoRex.
func moduleRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	mod := strings.TrimSpace(string(out))
	if mod == "" || mod == os.DevNull {
		return "", fmt.Errorf("not in GoRex's module: run it from its directory")
	}
	return filepath.Dir(mod), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mkssh:", err)
	os.Exit(1)
}
