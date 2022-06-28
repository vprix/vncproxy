package rfb

import "io"

type Option func(*Options)
type GetConn func() (io.ReadWriteCloser, error)

// Options 配置信息
type Options struct {
	// 公共配置
	Handlers           []IHandler         //  处理程序列表
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

	// 生成连接的方法
	GetConn GetConn
}

// OptHandlers 设置流程处理程序
func OptHandlers(opt ...IHandler) Option {
	return func(options *Options) {
		options.Handlers = append(options.Handlers, opt...)
	}
}

// OptSecurityHandlers 设置权限认证处理程序
func OptSecurityHandlers(opt ...ISecurityHandler) Option {
	return func(options *Options) {
		options.SecurityHandlers = append(options.SecurityHandlers, opt...)
	}
}

// OptEncodings 设置支持的编码格式
func OptEncodings(opt ...IEncoding) Option {
	return func(options *Options) {
		options.Encodings = append(options.Encodings, opt...)
	}
}

// OptMessages 设置支持的消息类型
func OptMessages(opt ...Message) Option {
	return func(options *Options) {
		options.Messages = append(options.Messages, opt...)
	}
}

// OptPixelFormat 设置像素格式
func OptPixelFormat(opt PixelFormat) Option {
	return func(options *Options) {
		options.PixelFormat = opt
	}
}

// OptGetConn 设置生成连接方法
func OptGetConn(opt GetConn) Option {
	return func(options *Options) {
		options.GetConn = opt
	}
}

func OptDesktopName(opt []byte) Option {
	return func(options *Options) {
		options.DesktopName = opt
	}
}

func OptHeight(opt int) Option {
	return func(options *Options) {
		options.Height = uint16(opt)
	}
}
func OptWidth(opt int) Option {
	return func(options *Options) {
		options.Width = uint16(opt)
	}
}
