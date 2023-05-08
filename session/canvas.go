package session

import (
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

type CanvasSession struct {
	canvas *canvas.VncCanvas

	options   rfb.Options     // 客户端配置信息
	protocol  string          //协议版本
	encodings []rfb.IEncoding // 支持的编码列
	swap      *gmap.Map
}

// NewCanvasSession 创建客户端会话
func NewCanvasSession(opts ...rfb.Option) *CanvasSession {
	sess := &CanvasSession{
		swap: gmap.New(true),
	}
	sess.configure(opts...)
	return sess
}

// Init 初始化参数
func (that *CanvasSession) Init(opts ...rfb.Option) error {
	that.configure(opts...)
	return nil
}

func (that *CanvasSession) configure(opts ...rfb.Option) {
	for _, o := range opts {
		o(&that.options)
	}
	if that.options.PixelFormat.BPP == 0 {
		that.options.PixelFormat = rfb.PixelFormat32bit
	}
	if that.options.QuitCh == nil {
		that.options.QuitCh = make(chan struct{})
	}
	if that.options.ErrorCh == nil {
		that.options.ErrorCh = make(chan error, 32)
	}
	if that.options.Input == nil {
		that.options.Input = make(chan rfb.Message)
	}
	if that.options.Output == nil {
		that.options.Output = make(chan rfb.Message)
	}
	if len(that.options.Handlers) == 0 {
		that.options.Handlers = DefaultClientHandlers
	}
	if len(that.options.Messages) == 0 {
		that.options.Messages = messages.DefaultServerMessages
	}
	if len(that.options.Encodings) == 0 {
		that.options.Encodings = encodings.DefaultEncodings
	}
}

func (that *CanvasSession) Start() {
	that.canvas = canvas.NewVncCanvas(int(that.options.Width), int(that.options.Height))
	that.canvas.DrawCursor = that.options.DrawCursor
}

// Conn 获取会话底层的网络链接
func (that *CanvasSession) Conn() io.ReadWriteCloser {
	return that.canvas
}

// Options 获取配置信息
func (that *CanvasSession) Options() rfb.Options {
	return that.options
}

// ProtocolVersion 获取会话使用的协议版本
func (that *CanvasSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *CanvasSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// Encodings 获取当前支持的编码格式
func (that *CanvasSession) Encodings() []rfb.IEncoding {
	return that.encodings
}

// SetEncodings 设置编码格式
func (that *CanvasSession) SetEncodings(encs []rfb.EncodingType) error {

	msg := &messages.SetEncodings{
		EncNum:    uint16(len(encs)),
		Encodings: encs,
	}
	//if logger.IsDebug() {
	//	logger.Debugf("[Proxy客户端->VNC服务端] 消息类型:%s,消息内容:%s", msg.Type(), msg.String())
	//}
	return msg.Write(that)
}

func (that *CanvasSession) Flush() error {
	return nil
}

// Wait 等待会话处理完成
func (that *CanvasSession) Wait() <-chan struct{} {
	return that.options.QuitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *CanvasSession) SecurityHandler() rfb.ISecurityHandler {
	return nil
}

// SetSecurityHandler 设置安全认证处理方法
func (that *CanvasSession) SetSecurityHandler(_ rfb.ISecurityHandler) {
}

// NewEncoding 通过编码类型判断是否支持编码对象
func (that *CanvasSession) NewEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *CanvasSession) Read(buf []byte) (int, error) {
	return that.canvas.Read(buf)
}

// Write 写入数据到链接
func (that *CanvasSession) Write(buf []byte) (int, error) {
	return that.canvas.Write(buf)
}

// Close 关闭会话
func (that *CanvasSession) Close() error {
	if that.options.QuitCh != nil {
		close(that.options.QuitCh)
		that.options.QuitCh = nil
	}
	return that.canvas.Close()
}

// Swap session存储的临时变量
func (that *CanvasSession) Swap() *gmap.Map {
	return that.swap
}

// Type session类型
func (that *CanvasSession) Type() rfb.SessionType {
	return rfb.CanvasSessionType
}

// SetPixelFormat 设置像素格式
func (that *CanvasSession) SetPixelFormat(pf rfb.PixelFormat) {
	that.options.PixelFormat = pf
}

// SetColorMap 设置颜色地图
func (that *CanvasSession) SetColorMap(cm rfb.ColorMap) {
	that.options.ColorMap = cm
}

// SetWidth 设置桌面宽度
func (that *CanvasSession) SetWidth(width uint16) {
	that.options.Width = width
}

// SetHeight 设置桌面高度
func (that *CanvasSession) SetHeight(height uint16) {
	that.options.Height = height
}

// SetDesktopName 设置桌面名称
func (that *CanvasSession) SetDesktopName(name []byte) {
	that.options.DesktopName = name
}
