package rfb

type ClientConfig struct {
	Handlers           []IHandler
	SecurityHandlers   []ISecurityHandler  // 安全验证
	PixelFormat        PixelFormat         // 像素格式
	Encodings          []IEncoding         // 像素编码格式
	DisableMessageType []ClientMessageType // 禁用的消息，碰到这些消息，则跳过
	ColorMap           ColorMap            // 颜色地图
	Output             chan ClientMessage  // 客户端消息通道
	Input              chan ServerMessage  // 服务端消息通道
	Exclusive          bool                // 是否独占
	DrawCursor         bool                // 是否绘制鼠标指针
	Messages           []ServerMessage     // 支持的服务端消息列表
	QuitCh             chan struct{}       // 退出
	ErrorCh            chan error          // 错误通道
	quit               chan struct{}
}
