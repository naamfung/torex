# WinKit

![winkit](https://raw.gitcode.com/jorsion/winkit/raw/main/share/showcases/winkit.png)

![nvim-sim](https://raw.gitcode.com/jorsion/winkit/raw/main/share/showcases/NVIM-VS.png)

![nvim-sheme](https://raw.gitcode.com/jorsion/winkit/raw/main/share/showcases/nvim-sheme.png)

![nvim-simv](https://raw.gitcode.com/jorsion/winkit/raw/main/share/showcases/nvim-sim-v.png)

## 项目简介

WinKit 是一款适用于视窗系统的便携式集成编译环境终端，在原始 w64devkit 的基础上进行了增强，提供了完整的开发工具链，令开发者能够在视窗环境下便捷地进行 Scheme/C/C++/Fortran/Ada/Cgo/Lua 等程序的开发与调试，或作为其他编程语言的基础运行环境，无须繁琐的环境配置。

## 主要特点

-  **便携式设计**：无须安装，解压即可使用
-  **完整工具链**：集成了多种开发工具同实用程序
-  **开箱即用**：预装配置好的开发环境
-  **中文友好**：针对中文用户使用场景进行了优化
-  **便捷操作**：支持目录右键菜单快速启动

## 集成工具

WinKit 在原版 w64devkit 的基础上进行了增强，集成了以下工具：

### 核心开发工具
- **GCC/MINGW64**：C/C++/Fortran/Ada/Cgo 编译套件
- **GDB**：强大的调试工具
- **GNU Make**：自动化构建工具
- **Scheme**：提供思科开源版 (Chez (Scheme)) 环境
- **Lua**：提供官方 C 实现的 Lua 环境

### 编辑器与辅助工具
- **Vim**：经典的文本编辑器
- **NVIM**：预设 NeoVim 配置，提供开箱即用的体验
- **Ctags**：代码导航工具

## 优化工作

为了提升中文用户的使用体验，WinKit 进行了大量源码级别的优化：

- 解决了中文显示同输入问题
- 优化了终端界面同交互体验
- 提供了更符合中文用户习惯的配置
- 增强了原版 w64devkit 的功能

## 快速开始

```cmd
git clone --depth=1 https://gitcode.com/jorsion/winkit.git
```

1. 克隆或下载本仓库到本地
2. 在视窗系统下直接双击 `winkit.exe` 即可启动环境
3. 如需在任意目录快速启动，请参考下方的命令使用说明

## 命令使用

### 注册右键菜单

使用管理员权限打开命令提示符（CMD），进入项目目录，执行以下命令：

```cmd
winkit -install
```

执行成功后，在任意目录的右键菜单中将会出现「在此目录中启动WinKit」选项。

### 卸载右键菜单

如需移除右键菜单功能，请使用管理员权限打开命令提示符（CMD），执行以下命令：

```cmd
winkit -uninstall
```

执行成功后，WinKit 写入注册表的右键菜单信息将被清理。

## 注意事项

- 请确保以管理员权限执行安装同卸载命令，否则可能无法正常注册右键菜单
- 如有任何问题或建议，欢迎提交反馈

## 特别说明

- WinKit根目录下的winkit.ini可以设置家目录路径、标题、字体与其大小，环境变量的适配范围等等。
- 终端环境默认使用「ASH」登录，环境变量会在「bin」目录下找寻命令，如未在此处揾到命令，则会到达「Git Bash」附带的「/usr/bin」目录。
- NVIM的预设配置为「sim」，已在「.profile」文件配置命令别名为「vim」，Shell自带的编辑器可用「vi」启动。
- 从外部复制文本内容粘贴到终端里的NVIM时，请使用「Shift+Insert」粘贴操作；在NVIM内编辑文本时可以使用「Ctrl+Insert」进行拷贝操作。
- NVIM下可用「Shift+或h或j或k或l」、「Shift+或上或下或左或右的方向键」、「Shift+上下翻页键」选择文本内容。
- VI/VIM/NVIM默认「i」是插入编辑模式、「v」是可视模式、「u」是撤销操作。

### NVIM替换命令

```
	:[range]s[ubstitute]/pattern/replacement/[flags]
	[range]: 替换的范围，默认是当前行，可以是行号范围（例如 1,10 表示从第1行到第10行）。
	/pattern/: 要查找的文字或正则表达式。
	/replacement/: 用于替换找到的文字。
	[flags]: 可选标志，控制替换的行为（例如 g 表示全局替换，即替换所有匹配项；c 表示确认每个替换）。

	示例：

	:%s/foo/bar/g

	或者逐个确认：
	:%s/foo/bar/gc

```

或者，按下默认的空格<leader>键呼出主菜单进行操作。
