package encodings

import (
	"encoding/binary"
	"github.com/vprix/vncproxy/rfb"
)

// DesktopNamePseudoEncoding 服务端设置桌面名字的消息
type DesktopNamePseudoEncoding struct {
	Name []byte
}

func (that *DesktopNamePseudoEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *DesktopNamePseudoEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &DesktopNamePseudoEncoding{}
	if len(data) > 0 && data[0] {
		obj.Name = that.Name
	}
	return obj
}

func (that *DesktopNamePseudoEncoding) Type() rfb.EncodingType {
	return rfb.EncDesktopNamePseudo
}

// Read 实现了编码接口
func (that *DesktopNamePseudoEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	var length uint32
	if err := binary.Read(session, binary.BigEndian, &length); err != nil {
		return err
	}
	name := make([]byte, length)
	if err := binary.Read(session, binary.BigEndian, &name); err != nil {
		return err
	}
	that.Name = name
	return nil
}

func (that *DesktopNamePseudoEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if err := binary.Write(session, binary.BigEndian, uint32(len(that.Name))); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.Name); err != nil {
		return err
	}

	return session.Flush()
}
