package canvas

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
	"image"
	"image/color"
	"image/draw"
	"io"
)

const (
	BlockWidth  = 16
	BlockHeight = 16
)

type VncCanvas struct {
	draw.Image
	imageBuffs     [2]draw.Image
	Cursor         draw.Image
	CursorMask     [][]bool
	CursorBackup   draw.Image
	CursorOffset   *image.Point
	CursorLocation *image.Point
	DrawCursor     bool
	Changed        map[string]bool
}

func NewVncCanvas(width, height int) *VncCanvas {
	writeImg := NewRGBImage(image.Rect(0, 0, width, height))
	canvas := &VncCanvas{
		Image: writeImg,
	}
	return canvas
}

// Read 从链接中读取数据
func (that *VncCanvas) Read(buf []byte) (int, error) {
	return 0, nil
}

// Write 写入数据到链接
func (that *VncCanvas) Write(buf []byte) (int, error) {
	return 0, nil
}

// Close 关闭会话
func (that *VncCanvas) Close() error {
	return nil
}

func (that *VncCanvas) SetChanged(rect *rfb.Rectangle) {
	if that.Changed == nil {
		that.Changed = make(map[string]bool)
	}
	for x := int(rect.X) / BlockWidth; x*BlockWidth < int(rect.X+rect.Width); x++ {
		for y := int(rect.Y) / BlockHeight; y*BlockHeight < int(rect.Y+rect.Height); y++ {
			key := fmt.Sprintf("%d,%d", x, y)
			//fmt.Println("setting block: ", key)
			that.Changed[key] = true
		}
	}
}

func (that *VncCanvas) Reset(rect *rfb.Rectangle) {
	that.Changed = nil
}

func (that *VncCanvas) RemoveCursor() image.Image {
	if that.Cursor == nil || that.CursorLocation == nil {
		return that.Image
	}
	if !that.DrawCursor {
		return that.Image
	}
	rect := that.Cursor.Bounds()
	loc := that.CursorLocation
	img := that.Image
	for y := rect.Min.Y; y < int(rect.Max.Y); y++ {
		for x := rect.Min.X; x < int(rect.Max.X); x++ {
			// offset := y*int(rect.Width) + x
			// if bitmask[y*int(scanLine)+x/8]&(1<<uint(7-x%8)) > 0 {
			col := that.CursorBackup.At(x, y)
			//mask := c.CursorMask.At(x, y).(color.RGBA)
			mask := that.CursorMask[x][y]
			//logger.Info("Drawing Cursor: ", x, y, col, mask)
			if mask {
				//logger.Info("Drawing Cursor for real: ", x, y, col)
				img.Set(x+loc.X-that.CursorOffset.X, y+loc.Y-that.CursorOffset.Y, col)
			}
			// 	//logger.Tracef("CursorPseudoEncoding.Read: setting pixel: (%d,%d) %v", x+int(rect.X), y+int(rect.Y), colors[offset])
			// }
		}
	}
	return img
}

func (that *VncCanvas) PaintCursor() image.Image {
	if that.Cursor == nil || that.CursorLocation == nil {
		return that.Image
	}
	if !that.DrawCursor {
		return that.Image
	}
	rect := that.Cursor.Bounds()
	if that.CursorBackup == nil {
		that.CursorBackup = image.NewRGBA(that.Cursor.Bounds())
	}

	loc := that.CursorLocation
	img := that.Image
	for y := rect.Min.Y; y < int(rect.Max.Y); y++ {
		for x := rect.Min.X; x < int(rect.Max.X); x++ {
			// offset := y*int(rect.Width) + x
			// if bitmask[y*int(scanLine)+x/8]&(1<<uint(7-x%8)) > 0 {
			col := that.Cursor.At(x, y)
			//mask := c.CursorMask.At(x, y).(RGBColor)
			mask := that.CursorMask[x][y]
			backup := that.Image.At(x+loc.X-that.CursorOffset.X, y+loc.Y-that.CursorOffset.Y)
			//c.CursorBackup.Set(x, y, backup)
			//backup the previous data at this point

			//logger.Info("Drawing Cursor: ", x, y, col, mask)
			if mask {

				that.CursorBackup.Set(x, y, backup)
				//logger.Info("Drawing Cursor for real: ", x, y, col)
				img.Set(x+loc.X-that.CursorOffset.X, y+loc.Y-that.CursorOffset.Y, col)
			}
			// 	//logger.Tracef("CursorPseudoEncoding.Read: setting pixel: (%d,%d) %v", x+int(rect.X), y+int(rect.Y), colors[offset])
			// }
		}
	}
	return img
}

// FillRect 为指定的矩形区域填充颜色
func (that *VncCanvas) FillRect(rect *image.Rectangle, c color.Color) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			that.Set(x, y, c)
		}
	}
}

// ReadColor Read unmarshal color from conn
func (that *VncCanvas) ReadColor(c io.Reader, pf *rfb.PixelFormat) (*color.RGBA, error) {
	if pf.TrueColor == 0 {
		return nil, errors.New("support for non true color formats was not implemented")
	}
	order := pf.Order()
	var pixel uint32

	switch pf.BPP {
	case 8:
		var px uint8
		if err := binary.Read(c, order, &px); err != nil {
			return nil, err
		}
		pixel = uint32(px)
	case 16:
		var px uint16
		if err := binary.Read(c, order, &px); err != nil {
			return nil, err
		}
		pixel = uint32(px)
	case 32:
		var px uint32
		if err := binary.Read(c, order, &px); err != nil {
			return nil, err
		}
		pixel = uint32(px)
	}
	rgb := color.RGBA{
		R: uint8((pixel >> pf.RedShift) & uint32(pf.RedMax)),
		G: uint8((pixel >> pf.GreenShift) & uint32(pf.GreenMax)),
		B: uint8((pixel >> pf.BlueShift) & uint32(pf.BlueMax)),
		A: 1,
	}

	return &rgb, nil
}

func (that *VncCanvas) DecodeRaw(reader io.Reader, pf *rfb.PixelFormat, rect *rfb.Rectangle) error {
	for y := 0; y < int(rect.Height); y++ {
		for x := 0; x < int(rect.Width); x++ {
			col, err := that.ReadColor(reader, pf)
			if err != nil {
				return err
			}
			that.Set(int(rect.X)+x, int(rect.Y)+y, col)
		}
	}
	return nil
}

func MakeRect(x, y, width, height int) image.Rectangle {
	return image.Rectangle{Min: image.Point{X: x, Y: y}, Max: image.Point{X: x + width, Y: y + height}}
}

func MakeRectFromVncRect(rect *rfb.Rectangle) image.Rectangle {
	return MakeRect(int(rect.X), int(rect.Y), int(rect.Width), int(rect.Height))
}
