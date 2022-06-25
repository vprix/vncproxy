package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// EndOfContinuousUpdates Bell 结束连续更新
type EndOfContinuousUpdates struct{}

func (that *EndOfContinuousUpdates) Clone() rfb.Message {
	return &EndOfContinuousUpdates{}
}
func (that *EndOfContinuousUpdates) Supported(session rfb.ISession) bool {
	return true
}

// String return string
func (that *EndOfContinuousUpdates) String() string {
	return fmt.Sprintf("EndOfContinuousUpdates")
}

// Type 消息类型
func (that *EndOfContinuousUpdates) Type() rfb.MessageType {
	return rfb.MessageType(rfb.EndOfContinuousUpdates)
}

// Read 响铃消息只有消息类型，没有数据
func (that *EndOfContinuousUpdates) Read(session rfb.ISession) (rfb.Message, error) {
	return &EndOfContinuousUpdates{}, nil
}

// Write 写入响应消息类型
func (that *EndOfContinuousUpdates) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	return session.Flush()
}
