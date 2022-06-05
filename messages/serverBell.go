package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// Bell 响铃
type Bell struct{}

func (that *Bell) Clone() rfb.ServerMessage {
	return &Bell{}
}
func (that *Bell) Supported(session rfb.ISession) bool {
	return true
}

// String return string
func (that *Bell) String() string {
	return fmt.Sprintf("bell")
}

// Type 消息类型
func (that *Bell) Type() rfb.ServerMessageType {
	return rfb.Bell
}

// Read 响铃消息只有消息类型，没有数据
func (that *Bell) Read(session rfb.ISession) (rfb.ServerMessage, error) {
	return &Bell{}, nil
}

// Write 写入响应消息类型
func (that *Bell) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	return session.Flush()
}
