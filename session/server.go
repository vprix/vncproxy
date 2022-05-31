package session

import (
	"bufio"
	"github.com/gogf/gf/container/gmap"
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

	cfg             *rfb.ServerConfig    //配置信息
	protocol        string               //协议版本
	colorMap        rfb.ColorMap         // 颜色地图
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap *gmap.Map

	desktopName []byte          // 桌面名称
	fbHeight    uint16          // 缓冲帧高度
	fbWidth     uint16          // 缓冲帧宽度
	pixelFormat rfb.PixelFormat // 像素格式
	quit        chan struct{}
}

var _ rfb.ISession = new(ServerSession)

func NewServerSession(c io.ReadWriteCloser, cfg *rfb.ServerConfig) *ServerSession {
	return &ServerSession{
		c:           c,
		br:          bufio.NewReader(c),
		bw:          bufio.NewWriter(c),
		cfg:         cfg,
		desktopName: cfg.DesktopName,
		encodings:   cfg.Encodings,
		pixelFormat: cfg.PixelFormat,
		fbWidth:     cfg.Width,
		fbHeight:    cfg.Height,
		quit:        make(chan struct{}),
	}
}

func (that *ServerSession) Server() error {
	if len(that.cfg.Handlers) == 0 {
		that.cfg.Handlers = DefaultServerHandlers
	}
	for _, h := range that.cfg.Handlers {
		if err := h.Handle(that); err != nil {
			if that.cfg.ErrorCh != nil {
				that.cfg.ErrorCh <- err
			}
			return that.Close()
		}
	}
	return nil
}
func (that *ServerSession) Conn() io.ReadWriteCloser {
	return that.c
}
func (that *ServerSession) Config() interface{} {
	return that.cfg
}

// ProtocolVersion 获取会话使用的协议版本
func (that *ServerSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *ServerSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// PixelFormat 获取像素格式
func (that *ServerSession) PixelFormat() rfb.PixelFormat {
	return that.pixelFormat
}

// SetPixelFormat 设置像素格式
func (that *ServerSession) SetPixelFormat(pf rfb.PixelFormat) error {
	that.pixelFormat = pf
	return nil
}

// ColorMap 获取颜色地图
func (that *ServerSession) ColorMap() rfb.ColorMap {
	return that.colorMap
}

// SetColorMap 设置颜色地图
func (that *ServerSession) SetColorMap(cm rfb.ColorMap) {
	that.colorMap = cm
}

// Encodings 获取当前支持的编码格式
func (that *ServerSession) Encodings() []rfb.IEncoding {
	return that.encodings
}

// SetEncodings 设置编码格式
func (that *ServerSession) SetEncodings(encs []rfb.EncodingType) error {
	es := make(map[rfb.EncodingType]rfb.IEncoding)
	for _, enc := range that.cfg.Encodings {
		es[enc.Type()] = enc
	}
	for _, encType := range encs {
		if enc, ok := es[encType]; ok {
			that.encodings = append(that.encodings, enc)
		}
	}
	return nil
}

// Width 获取桌面宽度
func (that *ServerSession) Width() uint16 {
	return that.fbWidth
}

// SetWidth 设置桌面宽度
func (that *ServerSession) SetWidth(width uint16) {
	that.fbWidth = width
}

// Height 获取桌面高度
func (that *ServerSession) Height() uint16 {
	return that.fbHeight
}

// SetHeight 设置桌面高度
func (that *ServerSession) SetHeight(height uint16) {
	that.fbHeight = height
}

// DesktopName 获取该会话的桌面名称
func (that *ServerSession) DesktopName() []byte {
	return that.desktopName
}

// SetDesktopName 设置桌面名称
func (that *ServerSession) SetDesktopName(name []byte) {
	that.desktopName = name
}

func (that *ServerSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *ServerSession) Wait() {
	<-that.quit
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
	if that.quit != nil {
		close(that.quit)
		that.quit = nil
	}
	return that.c.Close()
}
func (that *ServerSession) Swap() *gmap.Map {
	return that.swap
}
func (that *ServerSession) Type() rfb.SessionType {
	return rfb.ServerSessionType
}
