# WinKit 控制台字体自定义指南

## 概述

WinKit 现在支持通过配置文件自定义控制台字体和字体大小。此功能允许用户根据自己的偏好调整终端的外观，以获得更好的中文显示效果和视觉体验。

## 配置文件位置

配置文件 `winkit.ini` 应该位于 WinKit 可执行文件的同一目录中。

## 字体配置选项

在 `winkit.ini` 文件中，可以在 `[winkit]` 部分下配置以下字体相关选项：

### font
设置控制台使用的字体名称。必须是系统已安装的等宽字体（monospace font）。

默认值: `Consolas`

示例:
```ini
font = Consolas
font = Courier New
font = Lucida Console
```

### font size
设置字体大小（以像素为单位）。

默认值: `16`

示例:
```ini
font size = 14
font size = 16
font size = 18
```

## 完整配置文件示例

```ini
[winkit]
home = .
title = WinKit Terminal
font = Consolas
font size = 16
path type = inherit
```

## 支持的字体

控制台字体必须是系统已安装的等宽字体。以下是一些常见的支持字体：

- Consolas（推荐）
- Courier New
- Lucida Console
- Monaco（如果已安装）
- Source Code Pro（如果已安装）

## 注意事项

1. **字体要求**：只能使用系统已安装的字体，不能直接从文件路径加载字体文件。
2. **等宽字体**：为了确保文本对齐，建议只使用等宽字体。
3. **Unicode 支持**：为了更好地显示中文，建议使用支持 Unicode 的字体，如 Consolas。
4. **字体大小**：字体大小以像素为单位，建议在 12-20 之间调整以获得最佳效果。

## 自定义字体安装

如果您想使用不在系统默认字体列表中的字体：

1. 下载并安装所需的等宽字体文件（.ttf 或 .otf）
2. 重启 WinKit 以使新字体生效
3. 在 `winkit.ini` 中指定新字体的名称

例如，要使用 "Source Code Pro" 字体：
```ini
[winkit]
font = Source Code Pro
font size = 16
```

## 故障排除

如果字体设置没有生效：

1. 确保指定的字体已正确安装在系统中
2. 确保使用的是等宽字体
3. 检查 `winkit.ini` 文件语法是否正确
4. 重启 WinKit 以确保配置生效

## 技术说明

字体设置通过 Windows API `SetCurrentConsoleFontEx` 实现，该 API 允许程序动态更改控制台窗口的字体和字体大小。