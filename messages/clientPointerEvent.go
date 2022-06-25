package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// PointerEvent 鼠标事件
type PointerEvent struct {
	Mask uint8  //8 位掩码，表示键位状态，1为按下，0为弹起
	X, Y uint16 // 当前 X,Y 坐标
}

func (that *PointerEvent) Clone() rfb.Message {

	c := &PointerEvent{
		Mask: that.Mask,
		X:    that.X,
		Y:    that.Y,
	}
	return c
}
func (that *PointerEvent) Supported(session rfb.ISession) bool {
	return true
}

// String returns string
func (that *PointerEvent) String() string {
	return fmt.Sprintf("mask %d, x: %d, y: %d", that.Mask, that.X, that.Y)
}

// Type returns MessageType
func (that *PointerEvent) Type() rfb.MessageType {
	return rfb.MessageType(rfb.PointerEvent)
}

// Read 从会话中解析消息内容
func (that *PointerEvent) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &PointerEvent{}
	if err := binary.Read(session, binary.BigEndian, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write 把消息按协议格式写入会话
func (that *PointerEvent) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that); err != nil {
		return err
	}
	return session.Flush()
}
