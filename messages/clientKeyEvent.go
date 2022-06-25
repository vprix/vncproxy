package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// KeyEvent 键盘按键事件
type KeyEvent struct {
	Down uint8   // 1 表示键位按下，0 表示弹起
	_    [2]byte // 对齐字节，方便解析
	Key  rfb.Key // 表示具体的键位，https://www.x.org/releases/X11R7.6/doc/xproto/x11protocol.html#keysym_encoding
}

func (that *KeyEvent) Clone() rfb.Message {

	c := &KeyEvent{
		Down: that.Down,
		Key:  that.Key,
	}
	return c
}
func (that *KeyEvent) Supported(session rfb.ISession) bool {
	return true
}

// String returns string
func (that *KeyEvent) String() string {
	return fmt.Sprintf("down: %d, key: %v", that.Down, that.Key)
}

// Type returns MessageType
func (that *KeyEvent) Type() rfb.MessageType {
	return rfb.MessageType(rfb.KeyEvent)
}

// Read 从会话中解析消息内容
func (that *KeyEvent) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &KeyEvent{}
	if err := binary.Read(session, binary.BigEndian, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write 把消息按协议格式写入会话
func (that *KeyEvent) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that); err != nil {
		return err
	}
	return session.Flush()
}
