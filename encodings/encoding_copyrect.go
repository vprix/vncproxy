package encodings

import (
	"encoding/binary"
	"github.com/vprix/vncproxy/rfb"
)

// CopyRectEncoding 该编码方式对于客户端在某些已经有了相同的象素数据的时候是非常简单和有效的。
// 这种编码方式在网络中表现为x,y 坐标。让客户端知道去拷贝那一个矩形的象素数据。
// 它可以应用于很多种情况。最明显的就是当用户在屏幕上移动某一个窗口的时候，还有在窗口内容滚动的时候。
// 在优化画的时候不是很明显，一个比较智能的服务器可能只会发送一次，因为它知道在客户端的帧缓存里已经存在了。
// 复制矩形编码并不是完全独立地发送所有的数据矩形，而是对于像素值完全相同的一组矩形，
// 只发送第一个矩形全部数据，随后的矩形则只需要发送左上角X、Y坐标。
// 实际上，复制矩形编码主要指的就是随后的这一系列X、Y坐标，而对于第一个矩形具体采用何种编码类型并没有限制，
// 仅仅需要知道第一个矩形在帧缓冲区中的位置，以便于完成复制操作。
// 因此，往往是把复制矩形编码和其它针对某一个矩形的编码类型结合使用。
type CopyRectEncoding struct {
	SX, SY uint16
}

func (that *CopyRectEncoding) Type() rfb.EncodingType {
	return rfb.EncCopyRect
}

func (that *CopyRectEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *CopyRectEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &CopyRectEncoding{}
	if len(data) > 0 && data[0] {
		obj.SX = that.SX
		obj.SY = that.SY
	}
	return obj
}

func (that *CopyRectEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if err := binary.Read(session, binary.BigEndian, &that.SX); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.SY); err != nil {
		return err
	}

	return nil
}

func (that *CopyRectEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if err := binary.Write(session, binary.BigEndian, that.SX); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.SY); err != nil {
		return err
	}
	return nil
}
