package encodings

import (
	"bytes"
	"encoding/binary"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/rfb"
	"image"
	"image/color"
	"math"
)

// CursorPseudoEncoding 如果客户端请求指针/鼠标伪编码，那么就是说它有能力进行本地绘制鼠标。
// 这样就可以明显改善传输性能。服务器通过发送带有伪鼠标编码的伪矩形来设置鼠标的形状作为更新的一部分。
// 伪矩形的x 和y 表示鼠标的热点，宽和高表示用像素来表示鼠标的宽和高。包含宽X高像素值的数据带有位掩码。
// 位掩码是由从左到右，从上到下的扫描线组成，而每一扫描线被填充为floor((width +7) / 8)。
// 对应每一字节最重要的位表示最左边像素，对应1 位表示相应指针的像素是正确的。
type CursorPseudoEncoding struct {
	buff *bytes.Buffer
}

func (that *CursorPseudoEncoding) Supported(session rfb.ISession) bool {
	return true
}

// Draw 绘制鼠标指针
func (that *CursorPseudoEncoding) draw(cv *canvas.VncCanvas, pf rfb.PixelFormat, rect *rfb.Rectangle) error {
	numColors := int(rect.Height) * int(rect.Width)
	colors := make([]color.Color, numColors)
	var err error
	for i := 0; i < numColors; i++ {
		colors[i], err = cv.ReadColor(that.buff, &pf)
		if err != nil {
			return err
		}
	}
	// 获取掩码信息
	bitmask := make([]byte, int((rect.Width+7)/8*rect.Height))
	if err = binary.Read(that.buff, binary.BigEndian, &bitmask); err != nil {
		return err
	}
	scanLine := (rect.Width + 7) / 8
	// 生成鼠标指针的形状
	cursorImg := image.NewRGBA(canvas.MakeRect(0, 0, int(rect.Width), int(rect.Height)))
	var cursorMask [][]bool
	for i := 0; i < int(rect.Width); i++ {
		cursorMask = append(cursorMask, make([]bool, rect.Height))
	}
	// 填充鼠标指针的颜色
	for y := 0; y < int(rect.Height); y++ {
		for x := 0; x < int(rect.Width); x++ {
			offset := y*int(rect.Width) + x
			if bitmask[y*int(scanLine)+x/8]&(1<<uint(7-x%8)) > 0 {
				cursorImg.Set(x, y, colors[offset])
				cursorMask[x][y] = true
			}
		}
	}
	// 设置鼠标指针
	cv.CursorOffset = &image.Point{X: int(rect.X), Y: int(rect.Y)}
	cv.Cursor = cursorImg
	cv.CursorBackup = image.NewRGBA(cursorImg.Bounds())
	cv.CursorMask = cursorMask

	return nil
}

func (that *CursorPseudoEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &CursorPseudoEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *CursorPseudoEncoding) Type() rfb.EncodingType {
	return rfb.EncCursorPseudo
}

func (that *CursorPseudoEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if rect.Width*rect.Height == 0 {
		return nil
	}
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	var bt []byte
	var err error

	bytesPixel := int(session.PixelFormat().BPP / 8) //calcTightBytePerPixel(pf)
	bt, err = ReadBytes(int(rect.Width*rect.Height)*bytesPixel, session)
	if err != nil {
		return err
	}
	_, _ = that.buff.Write(bt)
	mask := ((rect.Width + 7) / 8) * rect.Height
	bt, err = ReadBytes(int(math.Floor(float64(mask))), session)
	if err != nil {
		return err
	}
	_, _ = that.buff.Write(bt)
	return nil
}

func (that *CursorPseudoEncoding) Write(sess rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return nil
	}
	if sess.Type() == rfb.CanvasSessionType {
		return that.draw(sess.Conn().(*canvas.VncCanvas), sess.PixelFormat(), rect)
	}
	var err error
	_, err = that.buff.WriteTo(sess)
	that.buff.Reset()
	return err
}
