// Command mkbusybox builds the POSIX shell of WinKit's runtime from its
// source with zig's compiler, and puts it where torex embeds it.
//
// WinKit's build patches busybox-w32 and configures it as w64devkit's
// mingw64u_defconfig does; this repeats that from the vendored tarball.
// zig compiles it for Windows, with tools of WinKit, MSYS2 or Git for
// Windows: make, a POSIX shell for make's recipes, and git to patch.
//
//	go run ./tools/mkbusybox -zig C:\path\to\zig.exe -make C:\path\to\make.exe
//
// With -app it builds torex after the shell, with go tool mygo build.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// What WinKit's Dockerfile builds: this source, and these switches off.
const (
	sourceTarball = "third_party/winkit/vendor/busybox-w32-FRP-5579-g5749feb35.tgz"
	sourceSHA256  = "cde68932d0680939977344259182482341adfffac4558cd1592bc26d2b558311"
	patchesGlob   = "third_party/winkit/src/busybox-*.patch"
	shellPath     = "internal/winkit/runtime/bin/busybox.exe"
)

func main() {
	work := flag.String("work", filepath.Join(".mygo", "winkit"), "build directory")
	zig := flag.String("zig", "", "zig program (default $ZIG, zig, or WinKit's share)")
	makeProg := flag.String("make", "", "make program (default $MAKE, make, or WinKit's bin)")
	sh := flag.String("sh", "", "POSIX shell for make's recipes (default $SHELL or Git's)")
	jobs := flag.Int("j", runtime.NumCPU(), "make jobs")
	app := flag.Bool("app", false, "build torex after the shell")
	flag.Parse()

	root, err := moduleRoot()
	if err != nil {
		fatal(err)
	}
	workDir := abs(root, *work)

	tools, err := findTools(*zig, *makeProg, *sh)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("mkbusybox: zig %s\nmkbusybox: make %s\nmkbusybox: shell %s\n", tools.zig, tools.make, tools.sh)

	src, err := extract(root, workDir)
	if err != nil {
		fatal(err)
	}
	if err := patch(root, workDir, src); err != nil {
		fatal(err)
	}
	if err := configure(src, tools, *jobs); err != nil {
		fatal(err)
	}
	bin, err := compile(root, src, workDir, tools, *jobs)
	if err != nil {
		fatal(err)
	}
	if err := install(root, bin); err != nil {
		fatal(err)
	}
	if *app {
		if err := buildApp(root); err != nil {
			fatal(err)
		}
	}
}

// buildApp builds torex, with the shell just built.
func buildApp(root string) error {
	fmt.Println("mkbusybox: go tool mygo build")
	cmd := exec.Command("go", "tool", "mygo", "build")
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go tool mygo build: %w", err)
	}
	return nil
}

type toolchain struct {
	zig, make, sh string
	env           []string
}

// findTools finds the programs the build runs: zig compiles, make drives
// the build, and the shell runs its recipes, with the tools they use.
func findTools(zig, makeProg, sh string) (toolchain, error) {
	var t toolchain
	t.zig = firstProgram(zig, "ZIG", "zig", winkitGlob("share/zig-*/zig.exe"))
	if t.zig == "" {
		return t, fmt.Errorf("no zig found: name it with -zig, $ZIG, or in PATH")
	}
	t.make = firstProgram(makeProg, "MAKE", "make", winkitFile("bin/make.exe"))
	if t.make == "" {
		return t, fmt.Errorf("no make found: name it with -make, $MAKE, or in PATH")
	}
	t.sh = sh
	if t.sh == "" {
		t.sh = os.Getenv("SHELL")
	}
	if t.sh == "" {
		t.sh = gitShell()
	}
	if t.sh == "" {
		t.sh = winkitFile("bin/sh.exe")
	}
	if t.sh == "" {
		return t, fmt.Errorf("no POSIX shell for make found: name it with -sh or $SHELL")
	}
	// make's recipes run sed, awk and more, which the shell's directory
	// of a WinKit or Git installation has.
	path := []string{filepath.Dir(t.make), filepath.Dir(t.sh)}
	if usr := gitUsrBin(); usr != "" {
		path = append(path, usr)
	}
	path = append(path, os.Getenv("PATH"))
	t.env = append(os.Environ(), "PATH="+strings.Join(path, string(os.PathListSeparator)), "SHELL="+t.sh, "ZIGCC_ZIG="+t.zig)
	return t, nil
}

// firstProgram returns the first of a name from the flags and environment,
// the path, and a fallback in a WinKit installation.
func firstProgram(flagValue, env, name, fallback string) string {
	for _, cand := range []string{flagValue, os.Getenv(env), fallback} {
		if cand == "" {
			continue
		}
		if p, err := exec.LookPath(cand); err == nil {
			return p
		}
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			return cand
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}

// winkitHome returns the directory of the WinKit installation this
// machine builds with.
func winkitHome() string { return os.Getenv("WINKIT_HOME") }

// winkitFile returns a file of the WinKit installation, or "".
func winkitFile(rel string) string {
	home := winkitHome()
	if home == "" {
		return ""
	}
	p := filepath.Join(home, filepath.FromSlash(rel))
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

// winkitGlob returns the last file of the WinKit installation that a
// pattern names, or "".
func winkitGlob(pattern string) string {
	home := winkitHome()
	if home == "" {
		return ""
	}
	matches, _ := filepath.Glob(filepath.Join(home, filepath.FromSlash(pattern)))
	sort.Strings(matches)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

// gitShell returns Git's shell: its usr/bin tools come with it.
func gitShell() string {
	usr := gitUsrBin()
	if usr == "" {
		return ""
	}
	for _, p := range []string{filepath.Join(filepath.Dir(usr), "bin", "sh.exe"), filepath.Join(usr, "sh.exe")} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// gitUsrBin returns the directory of Git's POSIX tools, or "".
func gitUsrBin() string {
	git, err := exec.LookPath("git")
	if err != nil {
		return ""
	}
	root := filepath.Dir(filepath.Dir(git))
	for _, p := range []string{filepath.Join(root, "usr", "bin"), filepath.Join(root, "bin")} {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return ""
}

// extract unpacks the source tarball into the work directory and returns
// the directory of the source.
func extract(root, workDir string) (string, error) {
	tgz := filepath.Join(root, sourceTarball)
	b, err := os.ReadFile(tgz)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != sourceSHA256 {
		return "", fmt.Errorf("%s: sha256 %s, want %s", sourceTarball, got, sourceSHA256)
	}
	dst := filepath.Join(workDir, "src")
	os.RemoveAll(dst)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var top string
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if h.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		name := filepath.Clean(filepath.FromSlash(h.Name))
		if name == "." || strings.HasPrefix(name, "..") || filepath.IsAbs(name) {
			continue
		}
		if top == "" && h.FileInfo().IsDir() {
			top = strings.Split(name, string(filepath.Separator))[0]
		}
		target := filepath.Join(dst, name)
		if h.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(h.Mode)&0o777|0o200)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}
	if top == "" {
		return "", fmt.Errorf("%s: no files", sourceTarball)
	}
	src := filepath.Join(dst, top)
	if st, err := os.Stat(src); err != nil || !st.IsDir() {
		return "", fmt.Errorf("%s: no %s directory", sourceTarball, top)
	}
	return src, nil
}

// patch applies WinKit's patches to the source, as the version of the
// source it was made for: their lines end in CRLF, which git's diff wants
// in LF.
func patch(root, workDir, src string) error {
	patches, err := filepath.Glob(filepath.Join(root, patchesGlob))
	if err != nil || len(patches) == 0 {
		return fmt.Errorf("no patches in %s", patchesGlob)
	}
	sort.Strings(patches)
	dir := filepath.Join(workDir, "patches")
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var args []string
	for _, p := range patches {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		clean := filepath.Join(dir, filepath.Base(p))
		if err := os.WriteFile(clean, []byte(strings.ReplaceAll(string(b), "\r\n", "\n")), 0o644); err != nil {
			return err
		}
		args = append(args, clean)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("no git to apply WinKit's patches: %w", err)
	}
	cmd := exec.Command(git, append([]string{"apply", "--whitespace=nowarn", "-p1"}, args...)...)
	cmd.Dir = src
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("patches: %v\n%s", err, out)
	}
	return nil
}

// configure makes the configuration of w64devkit's mingw64u_defconfig, with
// every applet the Windows port builds: the runtime's shell has the tools
// that a WinKit installation's other programs give, and its own alone.
func configure(src string, t toolchain, jobs int) error {
	return run(t, src, t.make, "mingw64u_defconfig")
}

// compile builds busybox and returns its program.
func compile(root, src, workDir string, t toolchain, jobs int) (string, error) {
	cc := filepath.Join(workDir, "zigcc.exe")
	build := exec.Command("go", "build", "-o", cc, "./tools/zigcc")
	build.Dir = root
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building zigcc: %v\n%s", err, out)
	}
	ccArg := filepath.ToSlash(cc)
	err := run(t, src, t.make, fmt.Sprintf("-j%d", jobs),
		"CROSS_COMPILE=", "CC="+ccArg, "HOSTCC="+ccArg,
		"CONFIG_EXTRA_CFLAGS=-D_WIN32_WINNT=0x502")
	if err != nil {
		return "", err
	}
	return filepath.Join(src, "busybox.exe"), nil
}

// install copies the shell where torex embeds it: the runtime's version is
// its checksum, so a new one is unpacked anew.
func install(root, bin string) error {
	b, err := os.ReadFile(bin)
	if err != nil {
		return err
	}
	dst := filepath.Join(root, shellPath)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, b, 0o755); err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	fmt.Printf("mkbusybox: %s (%d bytes, sha256 %s)\n", shellPath, len(b), hex.EncodeToString(sum[:6]))
	return nil
}

func run(t toolchain, dir, prog string, args ...string) error {
	cmd := exec.Command(prog, args...)
	cmd.Dir = dir
	cmd.Env = t.env
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", filepath.Base(prog), strings.Join(args, " "), err)
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

func abs(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mkbusybox:", err)
	os.Exit(1)
}
