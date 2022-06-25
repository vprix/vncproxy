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
	"net"
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
	c  net.Conn // 网络链接
	br *bufio.Reader
	bw *bufio.Writer

	cfg             *rfb.Option          // 客户端配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap    *gmap.Map
	quitCh  chan struct{} // 退出
	errorCh chan error
}

var _ rfb.ISession = new(ClientSession)

// NewClient 创建客户端会话
func NewClient(c net.Conn, cfg *rfb.Option) (*ClientSession, error) {
	enc := cfg.Encodings
	if len(cfg.Encodings) == 0 {
		enc = []rfb.IEncoding{&encodings.RawEncoding{}}
	}
	desktop := &rfb.Desktop{}
	desktop.SetPixelFormat(cfg.PixelFormat)

	if cfg.QuitCh == nil {
		cfg.QuitCh = make(chan struct{})
	}
	if cfg.ErrorCh == nil {
		cfg.ErrorCh = make(chan error, 32)
	}
	return &ClientSession{
		c:         c,
		br:        bufio.NewReader(c),
		bw:        bufio.NewWriter(c),
		cfg:       cfg,
		desktop:   desktop,
		encodings: enc,
		quitCh:    cfg.QuitCh,
		errorCh:   cfg.ErrorCh,
		swap:      gmap.New(true),
	}, nil
}

func (that *ClientSession) Run() {
	if len(that.cfg.Handlers) == 0 {
		that.cfg.Handlers = DefaultClientHandlers
	}
	for _, h := range that.cfg.Handlers {
		if err := h.Handle(that); err != nil {
			that.cfg.ErrorCh <- fmt.Errorf("握手失败，请检查服务是否启动: %v", err)
			err = that.Close()
			if err != nil {
				that.cfg.ErrorCh <- fmt.Errorf("关闭client失败: %v", err)
			}
			return
		}
	}
}

// Conn 获取会话底层的网络链接
func (that *ClientSession) Conn() io.ReadWriteCloser {
	return that.c
}

// Config 获取配置信息
func (that *ClientSession) Config() interface{} {
	return that.cfg
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
	return that.encodings
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
	<-that.quitCh
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
	for _, enc := range that.encodings {
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
	if that.quitCh != nil {
		close(that.quitCh)
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
