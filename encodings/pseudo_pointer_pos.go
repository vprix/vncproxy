package encodings

import (
	"errors"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/rfb"
	"image"
	"image/draw"
)

type CursorPosPseudoEncoding struct {
}

func (that *CursorPosPseudoEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *CursorPosPseudoEncoding) Draw(img draw.Image, rect *rfb.Rectangle) error {
	cv, ok := img.(*canvas.VncCanvas)
	if !ok {
		return errors.New("canvas error")
	}
	// 本地鼠标指针的位置
	cv.CursorLocation = &image.Point{X: int(rect.X), Y: int(rect.Y)}
	return nil
}

func (that *CursorPosPseudoEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &CursorPosPseudoEncoding{}
	return obj
}

func (that *CursorPosPseudoEncoding) Type() rfb.EncodingType {
	return rfb.EncPointerPosPseudo
}

func (that *CursorPosPseudoEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	return nil
}

func (that *CursorPosPseudoEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	return nil
}
