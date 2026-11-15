// Mkicon draws torex's icon, resources/icon.png: a black tile with a
// terminal on it, the prompt in green and the cursor in cyan.
//
//	go run ./tools/mkicon
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"

	"golang.org/x/image/vector"
)

const size = 1024

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// The squircle of macOS's icon grid: 824 wide, inset by 100.
	x0, y0, w := float32(100), float32(100), float32(824)
	shadow(img, x0+6, y0+12, w-12, w-12, 186, 24, color.RGBA{0, 0, 0, 120})
	// The tile: black, lighter along the top, rimmed with a hairline so
	// that it reads on a dark desktop.
	fill(img, squircle(x0, y0, w, w), image.NewUniform(color.NRGBA{255, 255, 255, 34}))
	fill(img, squircle(x0+3, y0+3, w-6, w-6), gradient{y0: y0, y1: y0 + w, stops: []stop{
		{0, rgb(0x22262e)},
		{0.45, rgb(0x11141a)},
		{1, rgb(0x06070a)},
	}})
	// The terminal: a dark window with a grey rim.
	px, py, pw, ph := x0+120, y0+140, w-240, w-280
	shadow(img, px, py+12, pw, ph, 54, 26, color.RGBA{0, 0, 0, 150})
	fill(img, rounded(px, py, pw, ph, 54), image.NewUniform(rgb(0x394150)))
	fill(img, rounded(px+4, py+4, pw-8, ph-8, 50), gradient{y0: py, y1: py + ph, stops: []stop{
		{0, rgb(0x171b22)},
		{1, rgb(0x0a0c10)},
	}})
	// The window's controls, in the colors of a title bar.
	for i, c := range [3]color.NRGBA{rgb(0xff5f57), rgb(0xfebc2e), rgb(0x28c840)} {
		fill(img, disc(px+62+float32(i)*46, py+48, 14), image.NewUniform(c))
	}
	// The prompt: a chevron, a cursor, and two lines of output below.
	cy := py + 268
	green := image.NewUniform(rgb(0x3ee68f))
	for _, r := range stroke([][2]float32{{px + 108, cy - 104}, {px + 228, cy}, {px + 108, cy + 104}}, 52) {
		fill(img, r, green)
	}
	fill(img, rounded(px+272, cy-30, 172, 60, 12), image.NewUniform(rgb(0x5ad2ff)))
	fill(img, rounded(px+108, cy+140, pw*0.44, 22, 11), image.NewUniform(color.NRGBA{0x4a, 0x53, 0x61, 255}))
	fill(img, rounded(px+108, cy+186, pw*0.28, 22, 11), image.NewUniform(color.NRGBA{0x32, 0x39, 0x45, 255}))
	if err := os.MkdirAll("resources", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("resources/icon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func rgb(v uint32) color.NRGBA {
	return color.NRGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
}

type stop struct {
	at float64
	c  color.NRGBA
}

// gradient is a vertical gradient from y0 to y1 through stops, clamped
// outside the range.
type gradient struct {
	y0, y1 float32
	stops  []stop
}

func (g gradient) Bounds() image.Rectangle { return image.Rect(0, 0, size, size) }
func (g gradient) ColorModel() color.Model { return color.NRGBAModel }

func (g gradient) At(x, y int) color.Color {
	t := (float64(y) - float64(g.y0)) / (float64(g.y1) - float64(g.y0))
	t = math.Max(0, math.Min(1, t))
	i := 0
	for i < len(g.stops)-2 && t > g.stops[i+1].at {
		i++
	}
	a, b := g.stops[i], g.stops[i+1]
	f := (t - a.at) / (b.at - a.at)
	mix := func(p, q uint8) uint8 { return uint8(float64(p) + (float64(q)-float64(p))*f + 0.5) }
	return color.NRGBA{mix(a.c.R, b.c.R), mix(a.c.G, b.c.G), mix(a.c.B, b.c.B), 255}
}

func fill(dst *image.RGBA, r *vector.Rasterizer, src image.Image) {
	r.Draw(dst, dst.Bounds(), src, image.Point{})
}

// rounded is a rounded rectangle.
func rounded(x, y, w, h, rad float32) *vector.Rasterizer {
	r := vector.NewRasterizer(size, size)
	k := rad * 0.5523
	r.MoveTo(x+rad, y)
	r.LineTo(x+w-rad, y)
	r.CubeTo(x+w-rad+k, y, x+w, y+rad-k, x+w, y+rad)
	r.LineTo(x+w, y+h-rad)
	r.CubeTo(x+w, y+h-rad+k, x+w-rad+k, y+h, x+w-rad, y+h)
	r.LineTo(x+rad, y+h)
	r.CubeTo(x+rad-k, y+h, x, y+h-rad+k, x, y+h-rad)
	r.LineTo(x, y+rad)
	r.CubeTo(x, y+rad-k, x+rad-k, y, x+rad, y)
	r.ClosePath()
	return r
}

// squircle is a rounded rectangle of continuous curvature, as macOS's
// icons: a superellipse.
func squircle(x, y, w, h float32) *vector.Rasterizer {
	r := vector.NewRasterizer(size, size)
	const n = 5.0
	cx, cy, a, b := float64(x+w/2), float64(y+h/2), float64(w/2), float64(h/2)
	for i := 0; i <= 720; i++ {
		t := float64(i) / 720 * 2 * math.Pi
		c, s := math.Cos(t), math.Sin(t)
		px := cx + a*math.Copysign(math.Pow(math.Abs(c), 2/n), c)
		py := cy + b*math.Copysign(math.Pow(math.Abs(s), 2/n), s)
		if i == 0 {
			r.MoveTo(float32(px), float32(py))
		} else {
			r.LineTo(float32(px), float32(py))
		}
	}
	r.ClosePath()
	return r
}

// disc is a circle.
func disc(cx, cy, rad float32) *vector.Rasterizer {
	r := vector.NewRasterizer(size, size)
	circle(r, cx, cy, rad)
	return r
}

// stroke is a polyline of a width, with round joins and caps: shapes to
// fill one by one, as their windings would cancel where they overlap.
func stroke(pts [][2]float32, width float32) []*vector.Rasterizer {
	var out []*vector.Rasterizer
	hw := width / 2
	for i := 0; i+1 < len(pts); i++ {
		r := vector.NewRasterizer(size, size)
		out = append(out, r)
		a, b := pts[i], pts[i+1]
		dx, dy := b[0]-a[0], b[1]-a[1]
		l := float32(math.Hypot(float64(dx), float64(dy)))
		nx, ny := -dy/l*hw, dx/l*hw
		r.MoveTo(a[0]+nx, a[1]+ny)
		r.LineTo(b[0]+nx, b[1]+ny)
		r.LineTo(b[0]-nx, b[1]-ny)
		r.LineTo(a[0]-nx, a[1]-ny)
		r.ClosePath()
	}
	for _, p := range pts {
		r := vector.NewRasterizer(size, size)
		circle(r, p[0], p[1], hw)
		out = append(out, r)
	}
	return out
}

func circle(r *vector.Rasterizer, cx, cy, rad float32) {
	k := rad * 0.5523
	r.MoveTo(cx+rad, cy)
	r.CubeTo(cx+rad, cy+k, cx+k, cy+rad, cx, cy+rad)
	r.CubeTo(cx-k, cy+rad, cx-rad, cy+k, cx-rad, cy)
	r.CubeTo(cx-rad, cy-k, cx-k, cy-rad, cx, cy-rad)
	r.CubeTo(cx+k, cy-rad, cx+rad, cy-k, cx+rad, cy)
	r.ClosePath()
}

// shadow draws a soft shadow of a rounded rectangle, as layers of a
// little alpha each, growing.
func shadow(dst *image.RGBA, x, y, w, h, rad, blur float32, c color.RGBA) {
	steps := 12
	for i := steps; i >= 1; i-- {
		g := blur * float32(i) / float32(steps)
		a := float64(c.A) / float64(steps) * 0.9
		col := color.NRGBA{c.R, c.G, c.B, uint8(a)}
		m := rounded(x-g/2, y-g/2, w+g, h+g, rad+g/2)
		m.DrawOp = draw.Over
		fill(dst, m, image.NewUniform(col))
	}
}
