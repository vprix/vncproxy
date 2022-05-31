package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// ServerCutText 服务端剪切板发送到客户端
type ServerCutText struct {
	_      [3]byte // 填充
	Length uint32  // 剪切板内容长度
	Text   []byte  // 剪切板内容
}

func (that *ServerCutText) Clone() rfb.ServerMessage {
	return &ServerCutText{
		Length: that.Length,
		Text:   that.Text,
	}
}
func (that *ServerCutText) Supported(session rfb.ISession) bool {
	return true
}

// String returns string
func (that *ServerCutText) String() string {
	return fmt.Sprintf("lenght: %d", that.Length)
}

func (that *ServerCutText) Type() rfb.ServerMessageType {
	return rfb.ServerCutText
}

// 读取消息数据
func (that *ServerCutText) Read(session rfb.ISession) (rfb.ServerMessage, error) {
	// 每次读取以后生成的都是一个新的对象
	msg := &ServerCutText{}
	var pad [3]byte
	if err := binary.Read(session, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(session, binary.BigEndian, &msg.Length); err != nil {
		return nil, err
	}

	msg.Text = make([]byte, msg.Length)
	if err := binary.Read(session, binary.BigEndian, &msg.Text); err != nil {
		return nil, err
	}
	return msg, nil
}

func (that *ServerCutText) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	//写入填充
	var pad [3]byte
	if err := binary.Write(session, binary.BigEndian, pad); err != nil {
		return err
	}

	if that.Length < uint32(len(that.Text)) {
		that.Length = uint32(len(that.Text))
	}
	if err := binary.Write(session, binary.BigEndian, that.Length); err != nil {
		return err
	}

	if err := binary.Write(session, binary.BigEndian, that.Text); err != nil {
		return err
	}
	return session.Flush()
}
