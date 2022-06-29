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

	// 客户端配置信息
	options rfb.Options

	//协议版本
	protocol string
	// 最终选择的安全认证方式
	securityHandler rfb.ISecurityHandler

	//交换区
	swap *gmap.Map
}

var _ rfb.ISession = new(ClientSession)

// NewClient 创建客户端会话
func NewClient(opts ...rfb.Option) *ClientSession {
	sess := &ClientSession{
		swap: gmap.New(true),
	}
	sess.configure(opts...)

	return sess
}

// Init 初始化参数
func (that *ClientSession) Init(opts ...rfb.Option) error {
	that.configure(opts...)
	return nil
}

func (that *ClientSession) configure(opts ...rfb.Option) {
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

func (that *ClientSession) Start() {
	var err error
	that.c, err = that.options.GetConn(that)
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
		logger.Debugf("[Proxy客户端->VNC服务端] 消息类型:%s,消息内容:%s", rfb.ClientMessageType(msg.Type()), msg.String())
	}
	return msg.Write(that)
}

func (that *ClientSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *ClientSession) Wait() <-chan struct{} {
	return that.options.QuitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *ClientSession) SecurityHandler() rfb.ISecurityHandler {
	return that.securityHandler
}

// SetSecurityHandler 设置安全认证处理方法
func (that *ClientSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) {
	that.securityHandler = securityHandler
}

// NewEncoding 获取编码对象
func (that *ClientSession) NewEncoding(typ rfb.EncodingType) rfb.IEncoding {
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
		that.options.QuitCh <- struct{}{}
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

// SetPixelFormat 设置像素格式
func (that *ClientSession) SetPixelFormat(pf rfb.PixelFormat) {
	that.options.PixelFormat = pf
}

// SetColorMap 设置颜色地图
func (that *ClientSession) SetColorMap(cm rfb.ColorMap) {
	that.options.ColorMap = cm
}

// SetWidth 设置桌面宽度
func (that *ClientSession) SetWidth(width uint16) {
	that.options.Width = width
}

// SetHeight 设置桌面高度
func (that *ClientSession) SetHeight(height uint16) {
	that.options.Height = height
}

// SetDesktopName 设置桌面名称
func (that *ClientSession) SetDesktopName(name []byte) {
	that.options.DesktopName = name
}
