package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// SetPixelFormat 设置像素格式
type SetPixelFormat struct {
	_  [3]byte         // 填充
	PF rfb.PixelFormat // 像素格式
}

func (that *SetPixelFormat) Clone() rfb.ClientMessage {
	c := &SetPixelFormat{
		PF: that.PF,
	}
	return c
}

func (that *SetPixelFormat) Supported(session rfb.ISession) bool {
	return true
}

// String returns string
func (that *SetPixelFormat) String() string {
	return fmt.Sprintf("%s", that.PF)
}

// Type returns MessageType
func (that *SetPixelFormat) Type() rfb.ClientMessageType {
	return rfb.SetPixelFormat
}

// Write 写入像素格式
func (that *SetPixelFormat) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}

	if err := binary.Write(session, binary.BigEndian, that); err != nil {
		return err
	}

	pf := session.PixelFormat()
	// Invalidate the color map.
	if pf.TrueColor != 0 {
		session.SetColorMap(rfb.ColorMap{})
	}

	return session.Flush()
}

// Read 从链接中读取像素格式到当前对象
func (that *SetPixelFormat) Read(session rfb.ISession) (rfb.ClientMessage, error) {
	msg := &SetPixelFormat{}
	if err := binary.Read(session, binary.BigEndian, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
