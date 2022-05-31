package messages

import (
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// ServerInit  握手的时候服务端初始化消息
type ServerInit struct {
	FBWidth     uint16
	FBHeight    uint16
	PixelFormat rfb.PixelFormat
	NameLength  uint32
	NameText    []byte
}

func (srvInit ServerInit) String() string {
	return fmt.Sprintf("ServerInit->Width: %d, Height: %d, PixelFormat: %s, NameLength: %d, MameText: %s", srvInit.FBWidth, srvInit.FBHeight, srvInit.PixelFormat, srvInit.NameLength, srvInit.NameText)
}
