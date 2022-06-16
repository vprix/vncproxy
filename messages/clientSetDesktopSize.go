package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/internal/dbuffer"
	"github.com/vprix/vncproxy/rfb"
)

// SetDesktopSize 客户端发起设置桌面大小
type SetDesktopSize struct {
	buff *dbuffer.ByteBuffer
}

func (that *SetDesktopSize) Clone() rfb.ClientMessage {

	c := &SetDesktopSize{
		buff: dbuffer.GetByteBuffer(),
	}
	_, _ = c.buff.Write(that.buff.Bytes())
	return c
}
func (that *SetDesktopSize) Supported(rfb.ISession) bool {
	return true
}
func (that *SetDesktopSize) String() string {
	return fmt.Sprintf("(type=%d)", that.Type())
}

func (that *SetDesktopSize) Type() rfb.ClientMessageType {
	return rfb.SetDesktopSize
}

// 读取数据
func (that *SetDesktopSize) Read(session rfb.ISession) (rfb.ClientMessage, error) {
	msg := &SetDesktopSize{buff: dbuffer.GetByteBuffer()}
	pad := make([]byte, 1)
	if _, err := session.Read(pad); err != nil {
		return nil, err
	}
	var width uint16
	_, _ = msg.buff.Write(pad)
	if err := binary.Read(session, binary.BigEndian, &width); err != nil {
		return nil, err
	}
	if err := binary.Write(msg.buff, binary.BigEndian, width); err != nil {
		return nil, err
	}
	var height uint16
	if err := binary.Read(session, binary.BigEndian, &height); err != nil {
		return nil, err
	}
	if err := binary.Write(msg.buff, binary.BigEndian, height); err != nil {
		return nil, err
	}
	var numberOfScreens uint8
	if err := binary.Read(session, binary.BigEndian, &numberOfScreens); err != nil {
		return nil, err
	}
	if err := binary.Write(msg.buff, binary.BigEndian, numberOfScreens); err != nil {
		return nil, err
	}
	pad = make([]byte, 1)
	if err := binary.Read(session, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	_, _ = msg.buff.Write(pad)
	for i := 0; i < int(numberOfScreens); i++ {
		b, err := that.readExtendedDesktopSize(session)
		if err != nil {
			return nil, err
		}
		_, _ = msg.buff.Write(b)
	}
	return msg, nil
}

func (that *SetDesktopSize) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	_, err := session.Write(that.buff.Bytes())
	if err != nil {
		return err
	}
	dbuffer.ReleaseByteBuffer(that.buff)
	that.buff = nil
	return session.Flush()
}

// No. of bytes		Type	Description
//	4				U32			id
//	2				U16			x-position
//	2				U16			y-position
//	2				U16			width
//	2				U16			height
//	4				U32			flags
func (that *SetDesktopSize) readExtendedDesktopSize(session rfb.ISession) ([]byte, error) {
	desktopSizeBuf := make([]byte, 16)
	if err := binary.Read(session, binary.BigEndian, &desktopSizeBuf); err != nil {
		return nil, err
	}
	return desktopSizeBuf, nil
}
