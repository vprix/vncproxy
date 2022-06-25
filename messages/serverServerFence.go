package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// ServerFence 支持 Fence扩展的服务器发送此扩展以请求数据流的同步。
type ServerFence struct {
	flags   uint32
	length  uint8
	payload []byte
}

func (that *ServerFence) Clone() rfb.Message {

	c := &ServerFence{
		flags:   that.flags,
		length:  that.length,
		payload: that.payload,
	}
	return c
}
func (that *ServerFence) Supported(session rfb.ISession) bool {
	return true
}
func (that *ServerFence) String() string {
	return fmt.Sprintf("type=%d", that.Type())
}

func (that *ServerFence) Type() rfb.MessageType {
	return rfb.MessageType(rfb.ServerFence)
}

// 读取数据
func (that *ServerFence) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &ServerFence{}
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

func (that *ServerFence) Write(session rfb.ISession) error {
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
