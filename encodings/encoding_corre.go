package encodings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/vprix/vncproxy/rfb"
)

// CoRREEncoding CoRRE是RRE的变体，它把发送的最大矩形限制在255×255个像素以内，用一个字节就能表示子矩形的维度。
// 如果服务器想要发送一个超出限制的矩形，则只要把它划分成几个更小的RFB矩形即可。
// “对于通常的桌面，这样的方式具有比RRE更好的压缩度”。
// 实际上，如果进一步限制矩形的大小，就能够获得最好的压缩度。“矩形的最大值越小，决策的尺度就越好”。
// 但是，如果把矩形的最大值限制得太小，就增加了矩形的数量，而由于每个RFB矩形都会有一定的开销，结果反而会使压缩度变差。
// 所以应该选择一个比较恰当的数字。在目前的实现中，采用的最大值为48×48。
type CoRREEncoding struct {
	buff *bytes.Buffer
}

func (that *CoRREEncoding) Type() rfb.EncodingType {
	return rfb.EncCoRRE
}

func (that *CoRREEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *CoRREEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &CoRREEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}

	return obj
}

func (that *CoRREEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {

	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	pf := session.Desktop().PixelFormat()
	// 子矩形的数量
	var numOfSubRectangles uint32
	if err := binary.Read(session, binary.BigEndian, &numOfSubRectangles); err != nil {
		return err
	}
	if err := binary.Write(that.buff, binary.BigEndian, numOfSubRectangles); err != nil {
		return err
	}
	// (backgroundColor + (color=BPP + x=8b + y=8b + w=8b + h=8b))
	size := uint32(pf.BPP/8) + (uint32((pf.BPP/8)+4) * numOfSubRectangles)
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

func (that *CoRREEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) (err error) {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	_, err = that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}
