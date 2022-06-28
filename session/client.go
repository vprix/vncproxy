package session

import (
	"bufio"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

var (
	DefaultClientHandlers = []rfb.IHandler{
		&handler.ClientVersionHandler{},
		&handler.ClientSecurityHandler{},
		&handler.ClientClientInitHandler{},
		&handler.ClientServerInitHandler{},
		&handler.ClientMessageHandler{},
	}
)

// ClientSession proxy 客户端
type ClientSession struct {
	c  io.ReadWriteCloser // 网络链接
	br *bufio.Reader
	bw *bufio.Writer

	options         rfb.Options          // 客户端配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap *gmap.Map
}

var _ rfb.ISession = new(ClientSession)

// NewClient 创建客户端会话
func NewClient(opts ...rfb.Option) *ClientSession {
	sess := &ClientSession{
		swap: gmap.New(true),
	}
	for _, o := range opts {
		o(&sess.options)
	}

	desktop := &rfb.Desktop{}
	desktop.SetPixelFormat(sess.options.PixelFormat)
	sess.desktop = desktop
	if sess.options.QuitCh == nil {
		sess.options.QuitCh = make(chan struct{})
	}
	if sess.options.ErrorCh == nil {
		sess.options.ErrorCh = make(chan error, 32)
	}
	if sess.options.Input == nil {
		sess.options.Input = make(chan rfb.Message)
	}
	if sess.options.Output == nil {
		sess.options.Output = make(chan rfb.Message)
	}
	if len(sess.options.Handlers) == 0 {
		sess.options.Handlers = DefaultClientHandlers
	}
	if len(sess.options.Messages) == 0 {
		sess.options.Messages = messages.DefaultServerMessages
	}
	if len(sess.options.Encodings) == 0 {
		sess.options.Encodings = encodings.DefaultEncodings
	}
	return sess
}

// Init 初始化参数
func (that *ClientSession) Init(opts ...rfb.Option) error {
	for _, o := range opts {
		o(&that.options)
	}
	return nil
}

func (that *ClientSession) Run() {
	var err error
	that.c, err = that.options.GetConn()
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.br = bufio.NewReader(that.c)
	that.bw = bufio.NewWriter(that.c)

	if len(that.options.Handlers) == 0 {
		that.options.Handlers = DefaultClientHandlers
	}
	for _, h := range that.options.Handlers {
		if err := h.Handle(that); err != nil {
			that.options.ErrorCh <- fmt.Errorf("握手失败，请检查服务是否启动: %v", err)
			err = that.Close()
			if err != nil {
				that.options.ErrorCh <- fmt.Errorf("关闭client失败: %v", err)
			}
			return
		}
	}
}

// Conn 获取会话底层的网络链接
func (that *ClientSession) Conn() io.ReadWriteCloser {
	return that.c
}

// Options 获取配置信息
func (that *ClientSession) Options() rfb.Options {
	return that.options
}

// ProtocolVersion 获取会话使用的协议版本
func (that *ClientSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *ClientSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// Desktop 获取桌面对象
func (that *ClientSession) Desktop() *rfb.Desktop {
	return that.desktop
}

// Encodings 获取当前支持的编码格式
func (that *ClientSession) Encodings() []rfb.IEncoding {
	return that.options.Encodings
}

// SetEncodings 设置编码格式
func (that *ClientSession) SetEncodings(encs []rfb.EncodingType) error {

	msg := &messages.SetEncodings{
		EncNum:    uint16(len(encs)),
		Encodings: encs,
	}
	if logger.IsDebug() {
		logger.Debugf("[Proxy客户端->VNC服务端] 消息类型:%s,消息内容:%s", msg.Type(), msg.String())
	}
	return msg.Write(that)
}

func (that *ClientSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *ClientSession) Wait() {
	<-that.options.QuitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *ClientSession) SecurityHandler() rfb.ISecurityHandler {
	return that.securityHandler
}

// SetSecurityHandler 设置安全认证处理方法
func (that *ClientSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) error {
	that.securityHandler = securityHandler
	return nil
}

// GetEncoding 获取编码对象
func (that *ClientSession) GetEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.options.Encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *ClientSession) Read(buf []byte) (int, error) {
	return that.br.Read(buf)
}

// Write 写入数据到链接
func (that *ClientSession) Write(buf []byte) (int, error) {
	return that.bw.Write(buf)
}

// Close 关闭会话
func (that *ClientSession) Close() error {
	if that.options.QuitCh != nil {
		close(that.options.QuitCh)
	}
	return that.c.Close()
}

// Swap session存储的临时变量
func (that *ClientSession) Swap() *gmap.Map {
	return that.swap
}

// Type session类型
func (that *ClientSession) Type() rfb.SessionType {
	return rfb.ClientSessionType
}
