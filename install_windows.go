//go:build windows

package main

import (
	"os"
	"path/filepath"
)

// binReadme is what the installation's bin directory says of itself.
const binReadme = `这里放你自己的程序。

torex 把这个目录排在窗格 PATH 的最前面，所以放在这里的可执行文件
（git、jq、自己的脚本……）在窗格里直接敲名字就能用，也会盖过同名的
内置程序。这个说明可以删。
`

// etcProfile is the profile of the installation, whose owner edits it.
const etcProfile = `# torex 在每个窗格的 shell 启动时加载这个文件：内置 profile 之后、
# 你的 bash 命令之前。适合设置环境变量、扩展 PATH，例如：
#
#   export EDITOR=vim
#   export PATH="$PATH:C:/tools/bin"
#
# $TOREX_HOME 是这个文件的上两级，即 torex 的安装目录；
# 随程序一起带的目录可以这样指向：
#
#   export PATH="$TOREX_HOME/share/mytool/bin:$PATH"
#
# 改完之后，新开的窗格生效；已开的窗格执行一次
#
#   . "$TOREX_HOME/etc/profile"
#
# 即可。升级不会覆盖这个文件。
`

// prepareInstall leaves the notes of the installation's bin and etc beside
// the program, once: they are its owner's to change, and an upgrade keeps
// what they made of them.
func prepareInstall() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	root := filepath.Dir(exe)
	writeIfMissing(filepath.Join(root, "bin", "README.txt"), binReadme)
	writeIfMissing(filepath.Join(root, "etc", "profile"), etcProfile)
}

func writeIfMissing(path, text string) {
	if _, err := os.Stat(path); err == nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	os.WriteFile(path, []byte(text), 0o644)
}
