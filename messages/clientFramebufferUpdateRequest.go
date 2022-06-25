package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// FramebufferUpdateRequest 请求帧缓存更新消息
// incremental 通常为非 0 值，服务器只需要发有变化的图像信息。
// 当客户端丢失了缓存的帧缓冲信息，或者刚建立连接，需要完整的图像信息时，
// 将 incremental 置为 0，获取全量信息。
type FramebufferUpdateRequest struct {
	Inc           uint8  // 是否是增量请求
	X, Y          uint16 // 区域的起始坐标
	Width, Height uint16 // 区域的宽度和高度
}

func (that *FramebufferUpdateRequest) Clone() rfb.Message {

	c := &FramebufferUpdateRequest{
		Inc:    that.Inc,
		X:      that.X,
		Y:      that.Y,
		Width:  that.Width,
		Height: that.Height,
	}
	return c
}
func (that *FramebufferUpdateRequest) Supported(session rfb.ISession) bool {
	return true
}

// String returns string
func (that *FramebufferUpdateRequest) String() string {
	return fmt.Sprintf("incremental: %d, x: %d, y: %d, width: %d, height: %d", that.Inc, that.X, that.Y, that.Width, that.Height)
}

// Type returns MessageType
func (that *FramebufferUpdateRequest) Type() rfb.MessageType {
	return rfb.MessageType(rfb.FramebufferUpdateRequest)
}

// Read 从会话中解析消息内容
func (that *FramebufferUpdateRequest) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &FramebufferUpdateRequest{}
	if err := binary.Read(session, binary.BigEndian, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write 把消息按协议格式写入会话
func (that *FramebufferUpdateRequest) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that); err != nil {
		return err
	}
	return session.Flush()
}
