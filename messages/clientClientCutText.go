package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// ClientCutText 客户端发送剪切板内容到服务端
type ClientCutText struct {
	_      [3]byte // 填充
	Length uint32  // 剪切板内容长度
	Text   []byte  // 剪切板
}

func (that *ClientCutText) Clone() rfb.ClientMessage {
	c := &ClientCutText{
		Length: that.Length,
		Text:   that.Text,
	}
	return c
}
func (that *ClientCutText) Supported(rfb.ISession) bool {
	return true
}

// String
func (that *ClientCutText) String() string {
	return fmt.Sprintf("length: %d", that.Length)
}

// Type returns MessageType
func (that *ClientCutText) Type() rfb.ClientMessageType {
	return rfb.ClientCutText
}

// Read 从会话中解析消息内容
func (that *ClientCutText) Read(session rfb.ISession) (rfb.ClientMessage, error) {
	msg := &ClientCutText{}
	// 读取填充字节
	var pad [3]byte
	if err := binary.Read(session, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	// 读取消息长度
	if err := binary.Read(session, binary.BigEndian, &msg.Length); err != nil {
		return nil, err
	}
	// 读取指定长度的消息内容
	msg.Text = make([]byte, msg.Length)
	if err := binary.Read(session, binary.BigEndian, &msg.Text); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write 把消息按协议格式写入会话
func (that *ClientCutText) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}

	// 写入3给字节的填充
	var pad [3]byte
	if err := binary.Write(session, binary.BigEndian, &pad); err != nil {
		return err
	}

	if uint32(len(that.Text)) > that.Length {
		that.Length = uint32(len(that.Text))
	}

	// 写入剪切板内容长度
	if err := binary.Write(session, binary.BigEndian, that.Length); err != nil {
		return err
	}

	// 写入消息内容
	if err := binary.Write(session, binary.BigEndian, that.Text); err != nil {
		return err
	}

	return session.Flush()
}
