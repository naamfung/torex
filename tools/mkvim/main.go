// Command mkvim builds vim, the editor of WinKit's runtime, from its source
// with zig's compiler, and puts it where torex embeds it: the program with the
// runtime's files beside it.
//
// Vim's own Make_ming.mak drives the build, with tools of WinKit, MSYS2 or Git
// for Windows: make, a POSIX shell for make's recipes, and windres for vim's
// resource, the icon and the manifest.
//
//	go run ./tools/mkvim -zig C:\path\to\zig.exe -make C:\path\to\make.exe
//
// With -app it builds torex after the editor, with go tool mygo build.
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
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// The source, and where torex embeds what the build makes of it.
const (
	sourceTarball = "third_party/vim/vendor/vim-9.2.1345.tar.gz"
	sourceSHA256  = "ddae84ef8d9e8fcca1ec6e501a41235a2466e935a27d6a95e3a84677f84b8fb4"
	editorPath    = "internal/winkit/runtime/bin/vim.exe"
	runtimePath   = "internal/winkit/runtime/vim"
	// The system vimrc, which $VIM names: it makes the defaults of vim and
	// gives them the POSIX shell of the runtime.
	vimrcPath = "third_party/vim/vimrc"
	// Vim's runtime is for a whole editor; help and filetype detection are
	// part of using it, the spelling files and the tutorial are not.
	skipRuntime = "tutor spell"
)

func main() {
	work := flag.String("work", filepath.Join(".mygo", "vim"), "build directory")
	zig := flag.String("zig", "", "zig program (default $ZIG, zig, or WinKit's share)")
	makeProg := flag.String("make", "", "make program (default $MAKE, make, or WinKit's bin)")
	sh := flag.String("sh", "", "POSIX shell for make's recipes (default $SHELL or Git's)")
	jobs := flag.Int("j", runtime.NumCPU(), "make jobs")
	app := flag.Bool("app", false, "build torex after the editor")
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
	fmt.Printf("mkvim: zig %s\nmkvim: make %s\nmkvim: shell %s\nmkvim: windres %s\n", tools.zig, tools.make, tools.sh, tools.windres)

	src, err := extract(root, workDir)
	if err != nil {
		fatal(err)
	}
	bin, err := compile(root, src, workDir, tools, *jobs)
	if err != nil {
		fatal(err)
	}
	if err := install(root, bin); err != nil {
		fatal(err)
	}
	if err := installRuntime(root, src); err != nil {
		fatal(err)
	}
	if *app {
		if err := buildApp(root); err != nil {
			fatal(err)
		}
	}
}

// buildApp builds torex, with the editor just built.
func buildApp(root string) error {
	fmt.Println("mkvim: go tool mygo build")
	cmd := exec.Command("go", "tool", "mygo", "build")
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go tool mygo build: %w", err)
	}
	return nil
}

type toolchain struct {
	zig, make, sh, windres string
	env                    []string
}

// findTools finds the programs the build runs: zig compiles, make drives the
// build, the shell runs its recipes, and windres compiles vim's resource.
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
	t.windres = firstProgram("", "", "windres", winkitFile("bin/windres.exe"))
	if t.windres == "" {
		return t, fmt.Errorf("no windres found: it comes with a WinKit installation")
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

// winkitFile returns a file of the WinKit installation this machine builds
// with, or "".
func winkitFile(rel string) string {
	home := os.Getenv("WINKIT_HOME")
	if home == "" {
		return ""
	}
	p := filepath.Join(home, filepath.FromSlash(rel))
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

// winkitGlob returns the last file of the WinKit installation that a pattern
// names, or "".
func winkitGlob(pattern string) string {
	home := os.Getenv("WINKIT_HOME")
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

// extract unpacks the source tarball into the work directory and returns the
// directory of the source.
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

// compile builds the console vim of WinKit's runtime and returns its program.
//
// Without a GUI there is no GvimExt, no gvim, and no msvcrt; WINVER is the
// Windows 10 of this torex, whose headers name the mouse wheel event that vim
// reports. ARCH is left to the makefile, whose -march=x86-64 zigcc respells.
//
// Vim has no system vimrc on MS-Windows; SYS_VIMRC_FILE names the one of the
// runtime, which vim reads before the user's. The single quotes are the
// shell's, and $$ is make's.
func compile(root, src, workDir string, t toolchain, jobs int) (string, error) {
	cc := filepath.Join(workDir, "zigcc.exe")
	build := exec.Command("go", "build", "-o", cc, "./tools/zigcc")
	build.Dir = root
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building zigcc: %v\n%s", err, out)
	}
	err := run(t, filepath.Join(src, "src"), t.make, "-f", "Make_ming.mak", fmt.Sprintf("-j%d", jobs),
		"vim.exe", "CC="+filepath.ToSlash(cc), "CROSS=yes", "CROSS_COMPILE=",
		"GUI=no", "VIMDLL=no", "NETBEANS=no",
		"FEATURES=HUGE", "OPTIMIZE=SIZE", "STATIC_STDCPLUS=yes", "HAS_GCC_EH=no",
		"WINVER=0x0a00", "WINDRES="+filepath.ToSlash(t.windres),
		`EXTRA_DEFINES='-DSYS_VIMRC_FILE="$$VIMRUNTIME/vimrc"'`)
	if err != nil {
		return "", err
	}
	return filepath.Join(src, "src", "vim.exe"), nil
}

// install copies the editor where torex embeds it: the runtime's version is
// its checksum, so a new one is unpacked anew.
func install(root, bin string) error {
	b, err := os.ReadFile(bin)
	if err != nil {
		return err
	}
	dst := filepath.Join(root, editorPath)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, b, 0o755); err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	fmt.Printf("mkvim: %s (%d bytes, sha256 %s)\n", editorPath, len(b), hex.EncodeToString(sum[:6]))
	return nil
}

// installRuntime copies vim's runtime files, and our system vimrc, where
// torex embeds them: $VIMRUNTIME and $VIM name that directory.
func installRuntime(root, src string) error {
	dst := filepath.Join(root, runtimePath)
	os.RemoveAll(dst)
	skip := strings.Fields(skipRuntime)
	err := filepath.WalkDir(filepath.Join(src, "runtime"), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(filepath.Join(src, "runtime"), p)
		if err != nil || rel == "." {
			return err
		}
		for _, s := range skip {
			if rel == s || strings.HasPrefix(rel, s+string(filepath.Separator)) {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		return err
	}
	vimrc, err := os.ReadFile(filepath.Join(root, vimrcPath))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dst, "vimrc"), vimrc, 0o644); err != nil {
		return err
	}
	fmt.Printf("mkvim: %s (%d bytes)\n", runtimePath, dirSize(dst))
	return nil
}

// dirSize returns the bytes of the files under dir.
func dirSize(dir string) int64 {
	var n int64
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if st, err := d.Info(); err == nil {
				n += st.Size()
			}
		}
		return nil
	})
	return n
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
	fmt.Fprintln(os.Stderr, "mkvim:", err)
	os.Exit(1)
}
