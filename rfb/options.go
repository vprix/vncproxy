package rfb

// Option 配置信息
type Option struct {
	// 公共配置
	Handlers           []IHandler
	SecurityHandlers   []ISecurityHandler // 安全验证
	Encodings          []IEncoding
	PixelFormat        PixelFormat // 像素格式
	ColorMap           ColorMap    // 颜色地图
	Input              chan Message
	Output             chan Message
	Messages           []Message
	DisableMessageType []MessageType // 禁用的消息，碰到这些消息，则跳过
	QuitCh             chan struct{} // 退出
	ErrorCh            chan error

	// 服务端配置
	DesktopName []byte
	Height      uint16
	Width       uint16

	// 客户端配置
	DrawCursor bool // 是否绘制鼠标指针
	Exclusive  bool // 是否独占
}
