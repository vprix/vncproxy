package canvas

import (
	"image"
	"image/color"
)

type RGBColor struct {
	R, G, B uint8
}

func (that RGBColor) RGBA() (r, g, b, a uint32) {
	return uint32(that.R), uint32(that.G), uint32(that.B), 1
}

type RGBImage struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*3].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

func (that RGBImage) ColorModel() color.Model {
	return nil
}

func (that RGBImage) Bounds() image.Rectangle {
	return that.Rect
}

func (that RGBImage) At(x, y int) color.Color {
	col := that.RGBAt(x, y)
	return color.RGBA{R: col.R, G: col.G, B: col.B, A: 1}
}

func (that *RGBImage) RGBAt(x, y int) *RGBColor {
	if !(image.Point{X: x, Y: y}.In(that.Rect)) {
		return &RGBColor{}
	}
	i := that.PixOffset(x, y)
	return &RGBColor{that.Pix[i+0], that.Pix[i+1], that.Pix[i+2]}
}

func (that *RGBImage) PixOffset(x, y int) int {
	return (y-that.Rect.Min.Y)*that.Stride + (x-that.Rect.Min.X)*3
}

func (that RGBImage) Set(x, y int, c color.Color) {
	if !(image.Point{X: x, Y: y}.In(that.Rect)) {
		return
	}
	i := that.PixOffset(x, y)
	c1 := color.RGBAModel.Convert(c).(color.RGBA)
	that.Pix[i+0] = c1.R
	that.Pix[i+1] = c1.G
	that.Pix[i+2] = c1.B
}

func (that *RGBImage) SetRGB(x, y int, c color.RGBA) {
	if !(image.Point{X: x, Y: y}.In(that.Rect)) {
		return
	}
	i := that.PixOffset(x, y)
	that.Pix[i+0] = c.R
	that.Pix[i+1] = c.G
	that.Pix[i+2] = c.B
}

func NewRGBImage(r image.Rectangle) *RGBImage {
	w, h := r.Dx(), r.Dy()
	buf := make([]uint8, 3*w*h)
	return &RGBImage{buf, 3 * w, r}
}
