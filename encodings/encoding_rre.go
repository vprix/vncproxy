package encodings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/vprix/vncproxy/rfb"
)

// RREEncoding RRE表示提升和运行长度，正如它名字暗示的那样，它实质上表示二维向量的运行长度编码。
// RRE把矩形编码成可以被客户机的图形引擎翻译的格式。RRE不适合复杂的桌面，但在一些情况下比较有用。
// RRE的思想就是把像素矩形的数据分成一些子区域，和一些压缩原始区域的单元。最近最佳的分区方式一般是比较容易计算的。
// 编码是由像素值组成的，Vb(基本上是在矩形中最常用的像素值）和一个计数N，紧接着是N的子矩形列表，这些里面由数组组成，(x,y)是对应子矩形的坐标，
// 表示子矩形上-左的坐标值，(w,h) 则表示子矩形的宽高。客户端可以通过绘制使用背景像素数据值，然后再根据子矩形来绘制原始矩形。
// 二维行程编码本质上是对行程编码的一个二维模拟，而其压缩度可以保证与行程编码相同甚至更好。
// 而且更重要的是，采用RRE编码的矩形被传送到客户端以后，可以立即有效地被最简单的图形引擎所还原。
type RREEncoding struct {
	buff *bytes.Buffer
}

func (that *RREEncoding) Type() rfb.EncodingType {
	return rfb.EncRRE
}
func (that *RREEncoding) Supported(session rfb.ISession) bool {
	return true
}
func (that *RREEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &RREEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *RREEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	pf := session.Options().PixelFormat
	// 子矩形的数量
	var numOfSubRectangles uint32
	if err := binary.Read(session, binary.BigEndian, &numOfSubRectangles); err != nil {
		return err
	}
	if err := binary.Write(that.buff, binary.BigEndian, numOfSubRectangles); err != nil {
		return err
	}

	// (backgroundColor + (color=BPP + x=16b + y=16b + w=16b + h=16b))
	size := uint32(pf.BPP/8) + (uint32((pf.BPP/8)+8) * numOfSubRectangles)
	b, err := ReadBytes(int(size), session)
	if err != nil {
		return err
	}
	_, err = that.buff.Write(b)
	if err != nil {
		return err
	}

	return nil
}
func (that *RREEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	_, err := that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}
