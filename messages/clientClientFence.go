package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// ClientFence 支持 Fence扩展的客户端发送此扩展以请求数据流的同步。
type ClientFence struct {
	flags   uint32
	length  uint8
	payload []byte
}

func (that *ClientFence) Clone() rfb.Message {

	c := &ClientFence{
		flags:   that.flags,
		length:  that.length,
		payload: that.payload,
	}
	return c
}
func (that *ClientFence) Supported(session rfb.ISession) bool {
	return true
}
func (that *ClientFence) String() string {
	return fmt.Sprintf("(type=%d)", that.Type())
}

func (that *ClientFence) Type() rfb.MessageType {
	return rfb.MessageType(rfb.ClientFence)
}

// 读取数据
func (that *ClientFence) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &ClientFence{}
	bytes := make([]byte, 3)
	//c.Read(bytes)
	if _, err := session.Read(bytes); err != nil {
		return nil, err
	}
	if err := binary.Read(session, binary.BigEndian, &msg.flags); err != nil {
		return nil, err
	}
	if err := binary.Read(session, binary.BigEndian, &msg.length); err != nil {
		return nil, err
	}
	bytes = make([]byte, msg.length)
	if _, err := session.Read(bytes); err != nil {
		return nil, err
	}
	msg.payload = bytes
	return msg, nil
}

func (that *ClientFence) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	//写入填充
	var pad [3]byte
	if err := binary.Write(session, binary.BigEndian, pad); err != nil {
		return err
	}

	if err := binary.Write(session, binary.BigEndian, that.flags); err != nil {
		return err
	}

	if err := binary.Write(session, binary.BigEndian, that.length); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.payload); err != nil {
		return err
	}
	return session.Flush()
}
