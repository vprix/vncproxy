package session

import (
	"github.com/gogf/gf/container/gmap"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

type CanvasSession struct {
	cfg      *rfb.ClientConfig // 客户端配置信息
	protocol string            //协议版本
	colorMap rfb.ColorMap      // 颜色地图
	Canvas   *canvas.VncCanvas

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

// NewCanvasSession 创建客户端会话
func NewCanvasSession(cfg *rfb.ClientConfig) *CanvasSession {
	return &CanvasSession{
		cfg:         cfg,
		encodings:   cfg.Encodings,
		quitCh:      cfg.QuitCh,
		errorCh:     cfg.ErrorCh,
		pixelFormat: cfg.PixelFormat,
		quit:        make(chan struct{}),
		swap:        gmap.New(true),
	}
}

func (that *CanvasSession) Connect() error {
	that.Canvas = canvas.NewVncCanvas(int(that.Width()), int(that.Height()))
	that.Canvas.DrawCursor = that.cfg.DrawCursor
	return nil
}

// Conn 获取会话底层的网络链接
func (that *CanvasSession) Conn() io.ReadWriteCloser {
	return that.Canvas
}

// Config 获取配置信息
func (that *CanvasSession) Config() interface{} {
	return that.cfg
}

// ProtocolVersion 获取会话使用的协议版本
func (that *CanvasSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *CanvasSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// PixelFormat 获取像素格式
func (that *CanvasSession) PixelFormat() rfb.PixelFormat {
	return that.pixelFormat
}

// SetPixelFormat 设置像素格式
func (that *CanvasSession) SetPixelFormat(pf rfb.PixelFormat) error {
	that.pixelFormat = pf
	return nil
}

// ColorMap 获取颜色地图
func (that *CanvasSession) ColorMap() rfb.ColorMap {
	return that.colorMap
}

// SetColorMap 设置颜色地图
func (that *CanvasSession) SetColorMap(cm rfb.ColorMap) {
	that.colorMap = cm
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

// Width 获取桌面宽度
func (that *CanvasSession) Width() uint16 {
	return that.fbWidth
}

// SetWidth 设置桌面宽度
func (that *CanvasSession) SetWidth(width uint16) {
	that.fbWidth = width
}

// Height 获取桌面高度
func (that *CanvasSession) Height() uint16 {
	return that.fbHeight
}

// SetHeight 设置桌面高度
func (that *CanvasSession) SetHeight(height uint16) {
	that.fbHeight = height
}

// DesktopName 获取该会话的桌面名称
func (that *CanvasSession) DesktopName() []byte {
	return that.desktopName
}

// SetDesktopName 设置桌面名称
func (that *CanvasSession) SetDesktopName(name []byte) {
	that.desktopName = name
}

func (that *CanvasSession) Flush() error {
	return nil
}

// Wait 等待会话处理完成
func (that *CanvasSession) Wait() {
	<-that.quit
}

// SecurityHandler 返回安全认证处理方法
func (that *CanvasSession) SecurityHandler() rfb.ISecurityHandler {
	return that.securityHandler
}

// SetSecurityHandler 设置安全认证处理方法
func (that *CanvasSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) error {
	that.securityHandler = securityHandler
	return nil
}

// GetEncoding 通过编码类型判断是否支持编码对象
func (that *CanvasSession) GetEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *CanvasSession) Read(buf []byte) (int, error) {
	return that.Canvas.Read(buf)
}

// Write 写入数据到链接
func (that *CanvasSession) Write(buf []byte) (int, error) {
	return that.Canvas.Write(buf)
}

// Close 关闭会话
func (that *CanvasSession) Close() error {
	if that.quit != nil {
		close(that.quit)
		that.quit = nil
	}
	if that.quitCh != nil {
		close(that.quitCh)
	}
	return that.Canvas.Close()
}
func (that *CanvasSession) Swap() *gmap.Map {
	return that.swap
}
func (that *CanvasSession) Type() rfb.SessionType {
	return rfb.CanvasSessionType
}
