// Command zigcc is the C compiler of GoRex's WinKit runtime build: zig's
// compiler driver for Windows, with the depfile that kbuild's fixdep reads.
//
// A depfile names its target with a drive letter, as zig writes it from a
// temporary object, and fixdep takes the first colon of the file for the
// end of the target: it would read the rest of the drive letter as a file.
// This rewrites that target to the one the command builds.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	zig := os.Getenv("ZIGCC_ZIG")
	if zig == "" {
		zig = "zig"
	}
	args := os.Args[1:]
	var deps []string
	var out string
	linking := true
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-o" && i+1 < len(args):
			out = args[i+1]
			i++
		case args[i] == "-MF" && i+1 < len(args):
			deps = append(deps, args[i+1])
			i++
		case strings.HasPrefix(args[i], "-Wp,-MD,"):
			deps = append(deps, strings.TrimPrefix(args[i], "-Wp,-MD,"))
		case args[i] == "-c" || args[i] == "-E" || args[i] == "-S":
			linking = false
		}
	}
	// Zig compiles WinKit's C runtime as it links it, and its SEH handlers
	// need the unwind tables that kbuild turns off; WinKit's busybox does
	// too, to unwind as its fork emulation ends a child.
	args = without(args, "-fno-unwind-tables", "-fno-asynchronous-unwind-tables")
	// Makefiles that derive -march from the target triple spell the CPU as
	// gcc does, x86-64; zig's clang knows it as x86_64.
	args = renameArg(args, "-march=x86-64", "-march=x86_64")
	// Zig rejects __DATE__ and __TIME__ for reproducible builds; programs such
	// as vim embed them in their version string on purpose.
	args = append(args, "-Wno-error=date-time")
	if linking {
		// The linker does not know the map file that kbuild probes for.
		args = without(args, "-Wl,--verbose")
		args = withoutPrefix(args, "-Wl,-Map,", "-Wl,-Map=")
	}
	cmd := exec.Command(zig, append([]string{"cc", "-target", "x86_64-windows-gnu"}, args...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if err == nil {
		for _, dep := range deps {
			fixDepfile(dep, out)
		}
	}
	if ee, ok := err.(*exec.ExitError); ok {
		os.Exit(ee.ExitCode())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "zigcc:", err)
		os.Exit(1)
	}
}

// without returns args without the named flags.
func without(args []string, drop ...string) []string {
	out := args[:0]
	for _, a := range args {
		keep := true
		for _, d := range drop {
			if a == d {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, a)
		}
	}
	return out
}

// renameArg returns args with the argument old renamed to new.
func renameArg(args []string, old, renamed string) []string {
	for i, a := range args {
		if a == old {
			args[i] = renamed
		}
	}
	return args
}

// withoutPrefix returns args without those that start with a prefix.
func withoutPrefix(args []string, prefixes ...string) []string {
	out := args[:0]
	for _, a := range args {
		keep := true
		for _, p := range prefixes {
			if strings.HasPrefix(a, p) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, a)
		}
	}
	return out
}

// fixDepfile names the target of a depfile as out, which has no colon
// unless it is absolute: fixdep knows the target it builds anyway.
func fixDepfile(path, out string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	i := bytes.Index(b, []byte(": "))
	if i < 0 {
		return
	}
	target := out
	if target == "" || strings.Contains(target, ":") {
		target = strings.TrimSuffix(filepath.Base(out), ".exe")
		if target == "" || strings.Contains(target, ":") {
			target = "target"
		}
	}
	os.WriteFile(path, append([]byte(target), b[i:]...), 0o644)
}
