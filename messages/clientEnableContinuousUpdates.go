package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// EnableContinuousUpdates 客户端发送连续更新消息
type EnableContinuousUpdates struct {
	flag   uint8
	x      uint16
	y      uint16
	width  uint16
	height uint16
}

func (that *EnableContinuousUpdates) Clone() rfb.Message {

	c := &EnableContinuousUpdates{
		flag:   that.flag,
		x:      that.x,
		y:      that.y,
		width:  that.width,
		height: that.height,
	}
	return c
}
func (that *EnableContinuousUpdates) Supported(rfb.ISession) bool {
	return true
}
func (that *EnableContinuousUpdates) String() string {
	return fmt.Sprintf("(type=%d,flag=%d,x=%d,y=%d,width=%d,height=%d)", that.Type(), that.flag, that.x, that.y, that.width, that.height)
}

func (that *EnableContinuousUpdates) Type() rfb.MessageType {
	return rfb.MessageType(rfb.EnableContinuousUpdates)
}

// 读取数据
func (that *EnableContinuousUpdates) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &EnableContinuousUpdates{}
	if err := binary.Read(session, binary.BigEndian, &msg.flag); err != nil {
		return nil, err
	}
	if err := binary.Read(session, binary.BigEndian, &msg.x); err != nil {
		return nil, err
	}
	if err := binary.Read(session, binary.BigEndian, &msg.y); err != nil {
		return nil, err
	}
	if err := binary.Read(session, binary.BigEndian, &msg.width); err != nil {
		return nil, err
	}
	if err := binary.Read(session, binary.BigEndian, &msg.height); err != nil {
		return nil, err
	}
	return msg, nil
}

func (that *EnableContinuousUpdates) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.flag); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.x); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.y); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.width); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.height); err != nil {
		return err
	}
	return session.Flush()
}
