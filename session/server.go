package session

import (
	"bufio"
	"github.com/gogf/gf/container/gmap"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

var (
	DefaultServerHandlers = []rfb.IHandler{
		&handler.ServerVersionHandler{},
		&handler.ServerSecurityHandler{},
		&handler.ServerClientInitHandler{},
		&handler.ServerServerInitHandler{},
		&handler.ServerMessageHandler{},
	}
)

type ServerSession struct {
	c  io.ReadWriteCloser
	br *bufio.Reader
	bw *bufio.Writer

	options         rfb.Options          // 配置信息
	protocol        string               //协议版本
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap *gmap.Map
}

var _ rfb.ISession = new(ServerSession)

func NewServerSession(opts ...rfb.Option) *ServerSession {
	sess := &ServerSession{
		swap: gmap.New(true),
	}
	sess.configure(opts...)

	return sess
}

// Init 初始化参数
func (that *ServerSession) Init(opts ...rfb.Option) error {
	that.configure(opts...)
	return nil
}

func (that *ServerSession) configure(opts ...rfb.Option) {
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
	if len(that.options.Messages) == 0 {
		that.options.Messages = messages.DefaultClientMessage
	}
	if len(that.options.Encodings) == 0 {
		that.options.Encodings = encodings.DefaultEncodings
	}
}

func (that *ServerSession) Start() {
	var err error
	that.c, err = that.options.GetConn(that)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.br = bufio.NewReader(that.c)
	that.bw = bufio.NewWriter(that.c)

	if len(that.options.Handlers) == 0 {
		that.options.Handlers = DefaultServerHandlers
	}
	for _, h := range that.options.Handlers {
		if err = h.Handle(that); err != nil {
			that.options.ErrorCh <- err
			err = that.Close()
			if err != nil {
				that.options.ErrorCh <- err
			}
			return
		}
	}
	return
}
func (that *ServerSession) Conn() io.ReadWriteCloser {
	return that.c
}
func (that *ServerSession) Options() rfb.Options {
	return that.options
}

// ProtocolVersion 获取会话使用的协议版本
func (that *ServerSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *ServerSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// Encodings 获取当前支持的编码格式
func (that *ServerSession) Encodings() []rfb.IEncoding {
	return that.encodings
}

// SetEncodings 设置编码格式
func (that *ServerSession) SetEncodings(encs []rfb.EncodingType) error {
	es := make(map[rfb.EncodingType]rfb.IEncoding)
	for _, enc := range that.options.Encodings {
		es[enc.Type()] = enc
	}
	for _, encType := range encs {
		if enc, ok := es[encType]; ok {
			that.encodings = append(that.encodings, enc)
		}
	}
	return nil
}

func (that *ServerSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *ServerSession) Wait() <-chan struct{} {
	return that.options.QuitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *ServerSession) SecurityHandler() rfb.ISecurityHandler {
	return that.securityHandler
}

// SetSecurityHandler 设置安全认证处理方法
func (that *ServerSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) {
	that.securityHandler = securityHandler
}

// NewEncoding 通过编码类型判断是否支持编码对象
func (that *ServerSession) NewEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *ServerSession) Read(buf []byte) (int, error) {
	return that.br.Read(buf)
}

// Write 写入数据到链接
func (that *ServerSession) Write(buf []byte) (int, error) {
	return that.bw.Write(buf)
}

// Close 关闭会话
func (that *ServerSession) Close() error {
	if that.options.QuitCh != nil {
		that.options.QuitCh <- struct{}{}
	}
	return that.c.Close()
}

// Swap session存储的临时变量
func (that *ServerSession) Swap() *gmap.Map {
	return that.swap
}

// Type session类型
func (that *ServerSession) Type() rfb.SessionType {
	return rfb.ServerSessionType
}

// SetPixelFormat 设置像素格式
func (that *ServerSession) SetPixelFormat(pf rfb.PixelFormat) {
	that.options.PixelFormat = pf
}

// SetColorMap 设置颜色地图
func (that *ServerSession) SetColorMap(cm rfb.ColorMap) {
	that.options.ColorMap = cm
}

// SetWidth 设置桌面宽度
func (that *ServerSession) SetWidth(width uint16) {
	that.options.Width = width
}

// SetHeight 设置桌面高度
func (that *ServerSession) SetHeight(height uint16) {
	that.options.Height = height
}

// SetDesktopName 设置桌面名称
func (that *ServerSession) SetDesktopName(name []byte) {
	that.options.DesktopName = name
}
