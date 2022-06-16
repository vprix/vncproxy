package encodings

import (
	"bytes"
	"errors"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/rfb"
)

// RawEncoding 采用原始地像素数据，而不进行任何的加工处理。
// 在这种情况下，对于一个宽度乘以高度（即面积）为N的矩形，数据就由N个像素值组成，这些值表示按照扫描线顺序从左到右排列的每个像素。
// 很明显，这种编码方式是最简单的，也是效率最低的。
// RFB要求所有的客户都必须能够处理这种原始编码的数据，并且在客户没有特别指定需要某种编码方式的时候，RFB服务器就默认生成原始编码。
type RawEncoding struct {
	buff *bytes.Buffer
}

var _ rfb.IEncoding = new(RawEncoding)

func (that *RawEncoding) Supported(rfb.ISession) bool {
	return true
}
func (that *RawEncoding) Type() rfb.EncodingType {
	return rfb.EncRaw
}

func (that *RawEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &RawEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *RawEncoding) Write(sess rfb.ISession, rect *rfb.Rectangle) error {
	if sess.Type() == rfb.CanvasSessionType {
		cv, ok := sess.Conn().(*canvas.VncCanvas)
		if !ok {
			return errors.New("canvas error")
		}
		pf := sess.Desktop().PixelFormat()
		return cv.DecodeRaw(that.buff, &pf, rect)
	}
	var err error
	_, err = that.buff.WriteTo(sess)
	that.buff.Reset()
	return err
}

// Read 读取原始色彩表示
func (that *RawEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	pf := session.Desktop().PixelFormat()
	// 表示单个像素是使用多少字节表示，分别为 8，16，32，对应的是1，2，4字节
	// 知道表示像素的字节长度，则根据宽高就能算出此次传输的总长度
	size := int(rect.Height) * int(rect.Width) * int(pf.BPP/8)

	b, err := ReadBytes(size, session)
	if err != nil {
		return err
	}
	_, err = that.buff.Write(b)
	if err != nil {
		return err
	}
	return nil
}
