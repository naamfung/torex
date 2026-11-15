// Command mkapp builds torex with MyGo, as mygo build does, and names its
// installer "Torex Setup <version>.exe": the product's name, while the
// program stays torex.exe, the name on the command line.
//
//	go run ./tools/mkapp
//
// Arguments are mygo build's, as -platform.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		fatal(err)
	}
	build := exec.Command("go", append([]string{"tool", "mygo", "build"}, os.Args[1:]...)...)
	build.Dir = root
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		fatal(fmt.Errorf("go tool mygo build: %w", err))
	}
	if err := nameInstallers(root); err != nil {
		fatal(err)
	}
}

// nameInstallers gives the installers of a build the name of the product, and
// says where they are.
func nameInstallers(root string) error {
	matches, err := filepath.Glob(filepath.Join(root, "build", "*", "torex Setup *.exe"))
	if err != nil {
		return err
	}
	for _, m := range matches {
		// Windows takes the old and the new name for one file, so the
		// installer goes through a name that is neither.
		tmp := m + ".rename"
		if err := os.Rename(m, tmp); err != nil {
			return err
		}
		dst := filepath.Join(filepath.Dir(m), "Torex"+strings.TrimPrefix(filepath.Base(m), "torex"))
		if err := os.Rename(tmp, dst); err != nil {
			return err
		}
		rel, err := filepath.Rel(root, dst)
		if err != nil {
			rel = dst
		}
		fmt.Printf("mkapp: %s\n", rel)
	}
	return nil
}

// moduleRoot returns the directory of the go.mod of torex.
func moduleRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	mod := strings.TrimSpace(string(out))
	if mod == "" || mod == os.DevNull {
		return "", fmt.Errorf("not in torex's module: run it from its directory")
	}
	return filepath.Dir(mod), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mkapp:", err)
	os.Exit(1)
}
