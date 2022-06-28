package session

import (
	"bufio"
	"github.com/gogf/gf/container/gmap"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
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

	options         *rfb.Options         // 配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap    *gmap.Map
	quitCh  chan struct{} // 退出
	errorCh chan error
}

var _ rfb.ISession = new(ServerSession)

func NewServerSession(options *rfb.Options) (*ServerSession, error) {
	enc := options.Encodings
	if len(options.Encodings) == 0 {
		enc = []rfb.IEncoding{&encodings.RawEncoding{}}
	}
	desktop := &rfb.Desktop{}
	desktop.SetPixelFormat(options.PixelFormat)

	if options.QuitCh == nil {
		options.QuitCh = make(chan struct{})
	}
	if options.ErrorCh == nil {
		options.ErrorCh = make(chan error, 32)
	}
	c, err := options.CreateConn()
	if err != nil {
		return nil, err
	}

	return &ServerSession{
		c:         c,
		br:        bufio.NewReader(c),
		bw:        bufio.NewWriter(c),
		options:   options,
		desktop:   desktop,
		encodings: enc,
		quitCh:    options.QuitCh,
		errorCh:   options.ErrorCh,
		swap:      gmap.New(true),
	}, nil
}

func (that *ServerSession) Run() {
	if len(that.options.Handlers) == 0 {
		that.options.Handlers = DefaultServerHandlers
	}
	for _, h := range that.options.Handlers {
		if err := h.Handle(that); err != nil {
			that.errorCh <- err
			err = that.Close()
			if err != nil {
				that.errorCh <- err
			}
			return
		}
	}
	return
}
func (that *ServerSession) Conn() io.ReadWriteCloser {
	return that.c
}
func (that *ServerSession) Options() *rfb.Options {
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

// Desktop 获取桌面对象
func (that *ServerSession) Desktop() *rfb.Desktop {
	return that.desktop
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
func (that *ServerSession) Wait() {
	<-that.quitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *ServerSession) SecurityHandler() rfb.ISecurityHandler {
	return that.securityHandler
}

// SetSecurityHandler 设置安全认证处理方法
func (that *ServerSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) error {
	that.securityHandler = securityHandler
	return nil
}

// GetEncoding 通过编码类型判断是否支持编码对象
func (that *ServerSession) GetEncoding(typ rfb.EncodingType) rfb.IEncoding {
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
	if that.quitCh != nil {
		close(that.quitCh)
		that.quitCh = nil
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

func (that *ServerSession) Messages() []rfb.Message {
	return that.options.Messages
}
