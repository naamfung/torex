package main

import (
	_ "embed"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// colors are the app's own, in light and dark windows.
type colors struct {
	// The window's background: a gradient from top through mid to bottom,
	// and a tint toward the right.
	bgTop, bgMid, bgLow, bgBottom, bgTint ui.Color

	card, cardFocused, cardBorder, cardBorderFocused ui.Color
	shadow, shadowFocused                            ui.Color

	text, textMuted, textFaint, iconMuted ui.Color
	hover, pressed                        ui.Color

	track, trackBorder, tabActive, tabSep ui.Color
	tileRim                               ui.Color

	busy, attention, exited ui.Color

	panel, panelBorder, panelSel, backdrop ui.Color
}

var lightColors = colors{
	bgTop:    ui.Hex("#f6e7f6"),
	bgMid:    ui.Hex("#efe1e6"),
	bgLow:    ui.Hex("#eee0d3"),
	bgBottom: ui.Hex("#ece1bd"),
	bgTint:   ui.RGBA(242, 214, 222, 0.35),

	card:              ui.RGBA(255, 255, 255, 0.66),
	cardFocused:       ui.RGBA(255, 255, 255, 0.95),
	cardBorder:        ui.RGBA(255, 255, 255, 0.72),
	cardBorderFocused: ui.RGBA(255, 255, 255, 1),
	shadow:            ui.RGBA(60, 40, 50, 0.07),
	shadowFocused:     ui.RGBA(60, 40, 50, 0.12),

	text:      ui.Hex("#1d1d1f"),
	textMuted: ui.Hex("#636366"),
	textFaint: ui.Hex("#8e8e93"),
	iconMuted: ui.Hex("#8a8a8f"),
	hover:     ui.RGBA(0, 0, 0, 0.055),
	pressed:   ui.RGBA(0, 0, 0, 0.1),

	track:       ui.RGBA(255, 255, 255, 0.34),
	trackBorder: ui.RGBA(255, 255, 255, 0.45),
	tabActive:   ui.RGBA(255, 255, 255, 0.97),
	tabSep:      ui.RGBA(0, 0, 0, 0.13),
	tileRim:     ui.RGBA(255, 255, 255, 0.95),

	busy:      ui.Hex("#34a853"),
	attention: ui.Hex("#f59e0b"),
	exited:    ui.Hex("#a1a1a6"),

	panel:       ui.RGBA(252, 252, 253, 0.97),
	panelBorder: ui.RGBA(0, 0, 0, 0.08),
	panelSel:    ui.RGBA(46, 111, 208, 0.12),
	backdrop:    ui.RGBA(40, 20, 40, 0.06),
}

var darkColors = colors{
	// Near black, as WinKit's console: the panes are black over a black
	// window, so that the terminal is the paper of the window.
	bgTop:    ui.Hex("#161618"),
	bgMid:    ui.Hex("#101012"),
	bgLow:    ui.Hex("#0c0c0e"),
	bgBottom: ui.Hex("#121214"),
	bgTint:   ui.RGBA(255, 255, 255, 0.02),

	card:              ui.RGBA(0, 0, 0, 0.3),
	cardFocused:       ui.RGBA(0, 0, 0, 0.94),
	cardBorder:        ui.RGBA(255, 255, 255, 0.06),
	cardBorderFocused: ui.RGBA(255, 255, 255, 0.16),
	shadow:            ui.RGBA(0, 0, 0, 0.3),
	shadowFocused:     ui.RGBA(0, 0, 0, 0.5),

	text:      ui.Hex("#f2f2f7"),
	textMuted: ui.Hex("#aeaeb2"),
	textFaint: ui.Hex("#8e8e93"),
	iconMuted: ui.Hex("#98989d"),
	hover:     ui.RGBA(255, 255, 255, 0.08),
	pressed:   ui.RGBA(255, 255, 255, 0.14),

	track:       ui.RGBA(255, 255, 255, 0.06),
	trackBorder: ui.RGBA(255, 255, 255, 0.06),
	tabActive:   ui.RGBA(255, 255, 255, 0.14),
	tabSep:      ui.RGBA(255, 255, 255, 0.14),
	tileRim:     ui.RGBA(255, 255, 255, 0.22),

	busy:      ui.Hex("#4ade80"),
	attention: ui.Hex("#fbbf24"),
	exited:    ui.Hex("#636366"),

	panel:       ui.RGBA(20, 20, 23, 0.97),
	panelBorder: ui.RGBA(255, 255, 255, 0.1),
	panelSel:    ui.RGBA(108, 178, 255, 0.18),
	backdrop:    ui.RGBA(0, 0, 0, 0.35),
}

func colorsOf(c *ui.Context) *colors {
	if c.Theme().Dark {
		return &darkColors
	}
	return &lightColors
}

// paintBackground paints the window's background.
func paintBackground(p *ui.Painter, r ui.Rect, k *colors) {
	h1, h2 := r.H*0.38, r.H*0.32
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: h1 + 1}, ui.LinearGradient{From: k.bgTop, To: k.bgMid, Angle: 180}, 0)
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y + h1, W: r.W, H: h2 + 1}, ui.LinearGradient{From: k.bgMid, To: k.bgLow, Angle: 180}, 0)
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y + h1 + h2, W: r.W, H: r.H - h1 - h2}, ui.LinearGradient{From: k.bgLow, To: k.bgBottom, Angle: 180}, 0)
	clear := k.bgTint
	clear.A = 0
	p.FillGradient(r, ui.LinearGradient{From: clear, To: k.bgTint, Angle: 90, Start: 0.35, End: 1}, 0)
}

//go:embed assets/fonts/JetBrainsMono-Regular.ttf
var fontRegular []byte

//go:embed assets/fonts/JetBrainsMono-Bold.ttf
var fontBold []byte

//go:embed assets/fonts/JetBrainsMono-Italic.ttf
var fontItalic []byte

//go:embed assets/fonts/JetBrainsMono-BoldItalic.ttf
var fontBoldItalic []byte

func registerFonts() {
	for _, f := range [][]byte{fontRegular, fontBold, fontItalic, fontBoldItalic} {
		ui.RegisterFont(f, "JetBrains Mono")
	}
}

// WinKit's console font: Cascadia Mono SemiBold, at a 24-pixel cell.
var termFont = terminal.Font{Family: "Cascadia Mono, monospace", Size: 18, Weight: 600, LineHeight: 1.0}

// The terminals' colors: soft ink on the panes' paper, and the dark kind.
var lightTerm = &terminal.Theme{
	Foreground: ui.Hex("#2a2d31"),
	Background: ui.Hex("#fbfbfb"),
	Cursor:     ui.Hex("#3a3d42"),
	Selection:  ui.RGBA(46, 111, 208, 0.2),
	Palette: [16]ui.Color{
		ui.Hex("#2a2d31"), ui.Hex("#c2465a"), ui.Hex("#4e8e5f"), ui.Hex("#b07a1e"),
		ui.Hex("#2e6fd0"), ui.Hex("#9050c8"), ui.Hex("#1e8c9c"), ui.Hex("#8a8d93"),
		ui.Hex("#6e7178"), ui.Hex("#d9566b"), ui.Hex("#5ba671"), ui.Hex("#c89026"),
		ui.Hex("#4a8fe0"), ui.Hex("#a56bd8"), ui.Hex("#2ba2b3"), ui.Hex("#b5b8be"),
	},
}

var darkTerm = &terminal.Theme{
	Foreground: ui.Hex("#cccccc"),
	Background: ui.Hex("#0c0c0c"),
	Cursor:     ui.Hex("#cccccc"),
	Selection:  ui.RGBA(108, 178, 255, 0.28),
	// The console's palette, as Windows Terminal's Campbell: WinKit's
	// green prompts and blue directories show as they do there.
	Palette: [16]ui.Color{
		ui.Hex("#0c0c0c"), ui.Hex("#c50f1f"), ui.Hex("#13a10e"), ui.Hex("#c19c00"),
		ui.Hex("#0037da"), ui.Hex("#881798"), ui.Hex("#3a96dd"), ui.Hex("#cccccc"),
		ui.Hex("#767676"), ui.Hex("#e74856"), ui.Hex("#16c60c"), ui.Hex("#f9f1a5"),
		ui.Hex("#3b78ff"), ui.Hex("#b4009e"), ui.Hex("#61d6d6"), ui.Hex("#f2f2f2"),
	},
}
