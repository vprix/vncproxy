package rfb

type ServerConfig struct {
	Handlers           []IHandler
	SecurityHandlers   []ISecurityHandler
	Encodings          []IEncoding
	PixelFormat        PixelFormat
	ColorMap           ColorMap
	Input              chan ClientMessage
	Output             chan ServerMessage
	Messages           []ClientMessage
	DisableMessageType []ServerMessageType // 禁用的消息，碰到这些消息，则跳过
	DesktopName        []byte
	Height             uint16
	Width              uint16
	ErrorCh            chan error
}
