package messages

import "github.com/vprix/vncproxy/rfb"

var (
	// DefaultClientMessage 默认client支持的消息
	DefaultClientMessage = []rfb.ClientMessage{
		&SetPixelFormat{},
		&SetEncodings{},
		&FramebufferUpdateRequest{},
		&KeyEvent{},
		&PointerEvent{},
		&ClientCutText{},
		&ClientFence{},
		&SetDesktopSize{},
		&EnableContinuousUpdates{},
	}
	// DefaultServerMessages 默认server支持的消息
	DefaultServerMessages = []rfb.ServerMessage{
		&FramebufferUpdate{},
		&SetColorMapEntries{},
		&Bell{},
		&ServerCutText{},
		&EndOfContinuousUpdates{},
		&ServerFence{},
	}
)
