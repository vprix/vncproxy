package session

import (
	"bufio"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/os/glog"
	"github.com/osgochina/dmicro/logger"
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

type ClientSession struct {
	c  net.Conn // 网络链接
	br *bufio.Reader
	bw *bufio.Writer

	cfg      *rfb.ClientConfig // 客户端配置信息
	protocol string            //协议版本
	colorMap rfb.ColorMap      // 颜色地图

	desktopName     []byte               // 桌面名称
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	fbHeight    uint16          // 缓冲帧高度
	fbWidth     uint16          // 缓冲帧宽度
	pixelFormat rfb.PixelFormat // 像素格式

	swap *gmap.Map

	quitCh  chan struct{} // 退出
	quit    chan struct{}
	errorCh chan error
}

var _ rfb.ISession = new(ClientSession)

// NewClient 创建客户端会话
func NewClient(c net.Conn, cfg *rfb.ClientConfig) (*ClientSession, error) {
	if len(cfg.Encodings) == 0 {
		return nil, fmt.Errorf("必须要配置客户端支持的编码格式")
	}
	return &ClientSession{
		c:           c,
		cfg:         cfg,
		br:          bufio.NewReader(c),
		bw:          bufio.NewWriter(c),
		encodings:   cfg.Encodings,
		quitCh:      cfg.QuitCh,
		errorCh:     cfg.ErrorCh,
		pixelFormat: cfg.PixelFormat,
		quit:        make(chan struct{}),
		swap:        gmap.New(true),
	}, nil
}

func (that *ClientSession) Connect() error {
	if len(that.cfg.Handlers) == 0 {
		that.cfg.Handlers = DefaultClientHandlers
	}
	for _, h := range that.cfg.Handlers {
		if err := h.Handle(that); err != nil {
			glog.Error("握手失败，请检查服务是否启动: ", err)
			_ = that.Close()
			that.cfg.ErrorCh <- err
			return err
		}
	}
	return nil
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

// PixelFormat 获取像素格式
func (that *ClientSession) PixelFormat() rfb.PixelFormat {
	return that.pixelFormat
}

// SetPixelFormat 设置像素格式
func (that *ClientSession) SetPixelFormat(pf rfb.PixelFormat) error {
	that.pixelFormat = pf
	return nil
}

// ColorMap 获取颜色地图
func (that *ClientSession) ColorMap() rfb.ColorMap {
	return that.colorMap
}

// SetColorMap 设置颜色地图
func (that *ClientSession) SetColorMap(cm rfb.ColorMap) {
	that.colorMap = cm
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

// Width 获取桌面宽度
func (that *ClientSession) Width() uint16 {
	return that.fbWidth
}

// SetWidth 设置桌面宽度
func (that *ClientSession) SetWidth(width uint16) {
	that.fbWidth = width
}

// Height 获取桌面高度
func (that *ClientSession) Height() uint16 {
	return that.fbHeight
}

// SetHeight 设置桌面高度
func (that *ClientSession) SetHeight(height uint16) {
	that.fbHeight = height
}

// DesktopName 获取该会话的桌面名称
func (that *ClientSession) DesktopName() []byte {
	return that.desktopName
}

// SetDesktopName 设置桌面名称
func (that *ClientSession) SetDesktopName(name []byte) {
	that.desktopName = name
}

func (that *ClientSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *ClientSession) Wait() {
	<-that.quit
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

// GetEncoding 通过编码类型判断是否支持编码对象
func (that *ClientSession) GetEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// ResetAllEncodings 所有编码对象重置
//func (that *ClientSession) ResetAllEncodings() {
//	for _, enc := range that.encodings {
//		_ = enc.Reset()
//	}
//}

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
	if that.quit != nil {
		close(that.quit)
		that.quit = nil
	}
	if that.quitCh != nil {
		close(that.quitCh)
	}
	return that.c.Close()
}
func (that *ClientSession) Swap() *gmap.Map {
	return that.swap
}
func (that *ClientSession) Type() rfb.SessionType {
	return rfb.ClientSessionType
}
