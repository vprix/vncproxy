package rfb

import "io"

type Option func(*Options)
type GetConn func(sess ISession) (io.ReadWriteCloser, error)

// Options 配置信息
type Options struct {
	// 公共配置
	Handlers                 []IHandler          //  处理程序列表
	SecurityHandlers         []ISecurityHandler  // 安全验证
	Encodings                []IEncoding         // 支持的编码类型
	PixelFormat              PixelFormat         // 像素格式
	ColorMap                 ColorMap            // 颜色地图
	Input                    chan Message        // 输入消息
	Output                   chan Message        // 输出消息
	Messages                 []Message           // 支持的消息类型
	DisableServerMessageType []ServerMessageType // 禁用的消息，碰到这些消息，则跳过
	DisableClientMessageType []ClientMessageType // 禁用的消息，碰到这些消息，则跳过
	QuitCh                   chan struct{}       // 退出
	ErrorCh                  chan error          // 错误通道

	// 服务端配置
	DesktopName []byte // 桌面名称，作为服务端配置的时候，需要设置
	Height      uint16 // 缓冲帧高度，作为服务端配置的时候，需要设置
	Width       uint16 // 缓冲帧宽度，作为服务端配置的时候，需要设置

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

// OptDisableServerMessageType 要屏蔽的服务端消息
func OptDisableServerMessageType(opt ...ServerMessageType) Option {
	return func(options *Options) {
		options.DisableServerMessageType = opt
	}
}

// OptDisableClientMessageType 要屏蔽的客户端消息
func OptDisableClientMessageType(opt ...ClientMessageType) Option {
	return func(options *Options) {
		options.DisableClientMessageType = opt
	}
}
