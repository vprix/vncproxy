package session

import (
	"github.com/gogf/gf/container/gmap"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

type CanvasSession struct {
	canvas *canvas.VncCanvas

	cfg             *rfb.Option          // 客户端配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap    *gmap.Map
	quitCh  chan struct{} // 退出
	errorCh chan error
}

// NewCanvasSession 创建客户端会话
func NewCanvasSession(cfg *rfb.Option) *CanvasSession {
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
	return &CanvasSession{
		cfg:       cfg,
		desktop:   desktop,
		encodings: enc,
		quitCh:    cfg.QuitCh,
		errorCh:   cfg.ErrorCh,
		swap:      gmap.New(true),
	}
}

func (that *CanvasSession) Run() {
	that.canvas = canvas.NewVncCanvas(int(that.desktop.Width()), int(that.desktop.Height()))
	that.canvas.DrawCursor = that.cfg.DrawCursor
}

// Conn 获取会话底层的网络链接
func (that *CanvasSession) Conn() io.ReadWriteCloser {
	return that.canvas
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

// Desktop 获取桌面对象
func (that *CanvasSession) Desktop() *rfb.Desktop {
	return that.desktop
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
func (that *CanvasSession) Wait() {
	<-that.quitCh
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
	return that.canvas.Read(buf)
}

// Write 写入数据到链接
func (that *CanvasSession) Write(buf []byte) (int, error) {
	return that.canvas.Write(buf)
}

// Close 关闭会话
func (that *CanvasSession) Close() error {
	if that.quitCh != nil {
		close(that.quitCh)
		that.quitCh = nil
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
