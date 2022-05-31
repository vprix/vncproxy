package rfb

// ClientMessage vnc客户端发送给vnc服务端的消息接口
type ClientMessage interface {
	String() string
	Supported(ISession) bool
	Type() ClientMessageType
	Read(ISession) (ClientMessage, error)
	Write(ISession) error
	Clone() ClientMessage
}

// ServerMessage 服务端发送给客户端的消息接口
type ServerMessage interface {
	String() string
	Supported(ISession) bool
	Type() ServerMessageType
	Read(ISession) (ServerMessage, error)
	Write(ISession) error
	Clone() ServerMessage
}
