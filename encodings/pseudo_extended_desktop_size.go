package encodings

import (
	"bytes"
	"encoding/binary"
	"github.com/vprix/vncproxy/rfb"
)

// ExtendedDesktopSizePseudo 扩展适应客户端桌面分辨率
type ExtendedDesktopSizePseudo struct {
	buff *bytes.Buffer
}

func (that *ExtendedDesktopSizePseudo) Supported(rfb.ISession) bool {
	return true
}

func (that *ExtendedDesktopSizePseudo) Clone(data ...bool) rfb.IEncoding {
	obj := &ExtendedDesktopSizePseudo{}
	if len(data) > 0 && data[0] {
		obj.buff = &bytes.Buffer{}
		_, _ = obj.buff.Write(that.buff.Bytes())
	}
	return obj
}

func (that *ExtendedDesktopSizePseudo) Type() rfb.EncodingType {
	return rfb.EncExtendedDesktopSizePseudo
}

func (that *ExtendedDesktopSizePseudo) Write(session rfb.ISession, rect *rfb.Rectangle) (err error) {
	_, err = that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}

func (that *ExtendedDesktopSizePseudo) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	//读取屏幕数量
	screensNumber, err := ReadUint8(session)
	if err != nil {
		return err
	}
	err = binary.Write(that.buff, binary.BigEndian, screensNumber)
	if err != nil {
		return err
	}
	// 填充
	pad, err := ReadBytes(3, session)
	if err != nil {
		return err
	}
	err = binary.Write(that.buff, binary.BigEndian, pad)

	b2, err := ReadBytes(int(screensNumber)*16, session)
	if err != nil {
		return err
	}
	_, _ = that.buff.Write(b2)
	return nil
}
