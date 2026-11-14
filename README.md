# GoRex

A replica of [Superlogical's Rex](https://www.superlogical.com/updates/public-testing-beginning)
terminal, written in Go with [MyGo](https://mygo.egoist.dev)'s native UI and its
terminal plugin (Ghostty's libghostty-vt). No webview, no cgo: one ~15 MB app.

![GoRex](docs/screenshot.png)

## What it does

- **Tabs and split panes.** A pane alone in its tab is a card over the
  window's gradient, its terminal filling it; its header, with its
  program's icon, name and directory and buttons to split, zoom and close,
  comes over the terminal's top as the pointer does. Splitting keeps a
  header on every pane. Drag the gaps between panes to resize them
  (double-click resets a split to half).
- **Persistent sessions.** Shells run in a session server, a background
  process of the same binary (`GoRex -server`). Quit GoRex and everything
  keeps running; open it again and every tab, split and pane comes back as
  it was, programs still running. The server keeps each session's screen in
  a headless libghostty-vt emulator and sends a snapshot on attach, so
  full-screen programs (lazygit, vim, htop…) come back exactly.
- **Program activity.** The server watches each terminal's foreground
  process (`tcgetpgrp`, `sysctl`, `proc_pidinfo`; on Windows, the shell's
  newest child, its command line and directory): headers and tabs name the
  program (`Node`, `Git Changes` for lazygit, `Codex`, `SSH host`…) and its
  working directory, a green dot shows a program printing, and an orange dot
  marks a pane whose program finished or rang the bell out of sight. When
  the window is in the background, a long command finishing posts a
  notification.
- **Tab icons** stack the tiles of a tab's programs, the focused pane's in
  front, as Rex does. Tabs reorder by dragging, rename by double-click, and
  have a context menu.
- **Host chip** with the machine's name and model; its popover shows the
  chip, memory, OS and the session server.
- **Command palette** (⇧⌘P, Ctrl+Shift+P, or the ⌘ button): every command, and every pane
  to jump to, fuzzy-matched.
- Light and dark appearances (View ▸ Appearance), text size (⌘+ ⌘− ⌘0),
  JetBrains Mono embedded.

## Shortcuts

| macOS | Windows, Linux | |
|---|---|---|
| ⌘T | Ctrl+Shift+T | New tab |
| ⌘D / ⇧⌘D | Ctrl+Shift+D / Ctrl+Shift+E | Split right / down |
| ⌘W / ⇧⌘W | Ctrl+Shift+W | Close pane / tab |
| ⇧⌘↩ | Ctrl+Shift+Enter | Zoom the pane |
| ⌥⌘ arrows | Alt+arrows | Focus the pane in that direction |
| ⌃⌘ arrows | Alt+Shift+arrows | Move the nearest divider |
| ⌃⌘= | | Equalize panes |
| ⌘1…⌘9, ⇧⌘[ ⇧⌘], ⌃Tab | Ctrl+Shift+1…9, Ctrl+PageUp/PageDown, Ctrl+Tab | Switch tabs |
| ⇧⌘R | Ctrl+Shift+R | Rename tab |
| ⇧⌘P, ⌘P | Ctrl+Shift+P | Command palette |
| ⌘K | Ctrl+Shift+K | Clear |
| ⌘+ ⌘− ⌘0 | Ctrl+= Ctrl+- Ctrl+0 | Text size |
| ⌥⌘Q | | Quit and end all sessions |

Elsewhere than on macOS, Control and a letter are the terminal's (Ctrl+C,
Ctrl+D, Ctrl+W…), so GoRex's shortcuts take Shift too, as copying and
pasting do (Ctrl+Shift+C, Ctrl+Shift+V), and none take Control and Alt,
which type characters (AltGr) on many keyboards.

## Develop and build

GoRex needs [Go](https://go.dev/dl/) 1.27+ and MyGo 0.3.0, whose CLI
`go.mod` pins as a tool:

```sh
go tool mygo dev     # GoRex Dev, rebuilt and restarted as the code changes
go test ./...        # the server, and the view without a window
go tool mygo build   # build/darwin-arm64/GoRex.app and a .dmg
go run ./tools/mkbusybox       # build WinKit's shell from source, with zig
go run ./tools/mkbusybox -app  # Windows: the shell, then GoRex
go run ./tools/mkvim           # build vim from source, with zig
go run ./tools/mkvim -app      # Windows: vim, then GoRex
go run ./tools/mkicon  # redraw resources/icon.png
```

`go get -tool github.com/egoist/mygo/cmd/mygo@latest` updates MyGo and its
CLI together.

The app keeps its state in its data directory: the server's socket and
log, `layout.json` and `settings.json`, in `~/Library/Application
Support/GoRex` for the built app and `GoRex Dev` for `mygo dev`'s, so
that developing never touches the sessions of the app you use
(`GOREX_DIR` names another directory). The sessions outlive the app, but
not a rebuild: a development build replaces a server that an older build
started, ending its sessions, which then start again in the same
directories.

### Windows

GoRex runs on Windows 10 1809 or later, whose ConPTY runs the shells; its
data directory is `%AppData%\GoRex`. Panes run WinKit's POSIX shell — its
`busybox` as `sh`, built from source and embedded in GoRex, unpacked under
the data directory — and no shell of Windows: no PowerShell, no `cmd`, so
that every pane behaves alike, whatever the machine has installed. Its
login shell sets WinKit's environment and its prompt: `root@…` when the
window is elevated, else the Windows user. The window is frameless and its
title bar draws minimize, maximize and close with the other buttons; it
resizes from its sides and bottom, not its top.

`internal/winkit` carries the runtime embedded in the app: WinKit's
`busybox.exe` as `sh`, with `winkit.ini` and the profiles a login shell
reads. `third_party/winkit` has WinKit's source (MIT) and the busybox-w32
source tarball, whose checksum `tools/mkbusybox` verifies; the tool applies
WinKit's patches and builds the shell with zig, as WinKit's Dockerfile does.
`go run ./tools/mkbusybox -app` builds the shell and then GoRex, in Go;
`go build` and `go tool mygo build` embed the shell that
`internal/winkit/runtime` has, and a new one is unpacked under the data
directory by its checksum. Building the shell needs zig, a `make`, and Git
(its POSIX shell runs make's recipes, and `git apply` applies the patches);
`tools/mkbusybox` finds WinKit's zig and make in `%WINKIT_HOME%` too.

The runtime carries vim as its editor: the console `vim.exe` and its runtime
files, embedded and unpacked beside the shell. `$VIMRUNTIME` names them, and
its `vimrc` — the system vimrc vim has no other way on Windows — gives vim
the defaults and the POSIX shell of the runtime, so a filter or `:!` runs in
`sh`, not `cmd`. `third_party/vim` has the source tarball and that `vimrc`;
`tools/mkvim` verifies the checksum, builds `vim.exe` with zig and windres
(of `%WINKIT_HOME%`), and copies the editor and its runtime into
`internal/winkit/runtime`. `vi` is an alias of `vim`.

## Layout

| | |
|---|---|
| `main.go` | the app, its window, and `-server` |
| `state.go` | tabs, the tree of splits, panes, saving and restoring the layout |
| `view.go`, `tabs.go`, `host.go`, `commands.go` | the interface: title bar, tabs, panes, host popover, menus and palette |
| `programs.go` | how programs show: names, glyphs, tile colors |
| `style.go`, `settings.go` | colors, fonts, terminal themes; appearance and text size |
| `internal/rex` | the session server and its client: PTYs, foreground processes, headless screens, the socket protocol |
| `internal/winkit` | the WinKit runtime embedded in the app, and its unpacking; source in `third_party/winkit` |
| `tools/mkbusybox`, `tools/mkvim`, `tools/zigcc` | building that runtime from WinKit's busybox-w32 and vim's source with zig |

Not replicated: Rex's connections to servers on other machines; GoRex's
server listens on a local socket only.

Icons: [Lucide](https://lucide.dev) (ISC) and [Simple Icons](https://simpleicons.org)
(CC0). Font: [JetBrains Mono](https://www.jetbrains.com/lp/mono/) (OFL).
