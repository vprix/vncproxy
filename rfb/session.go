package rfb

import (
	"github.com/gogf/gf/container/gmap"
	"io"
)

// ISession vnc连接的接口
type ISession interface {
	io.ReadWriteCloser
	Conn() io.ReadWriteCloser
	Start()
	Flush() error          // 清空缓冲区
	Wait() <-chan struct{} // 等待会话处理结束
	Init(...Option) error
	Options() Options
	SetPixelFormat(PixelFormat)
	SetColorMap(ColorMap)
	SetWidth(uint16)
	SetHeight(uint16)
	SetDesktopName([]byte)
	ProtocolVersion() string             // 获取当前的rfb协议
	SetProtocolVersion(string)           // 设置rfb协议
	SetSecurityHandler(ISecurityHandler) // 设置安全认证处理方法
	SecurityHandler() ISecurityHandler   // 获取当前安全认证的处理方法
	Encodings() []IEncoding              // 获取该会话支持的图像编码类型
	SetEncodings([]EncodingType) error   // 设置该链接支持的图像编码类型
	NewEncoding(EncodingType) IEncoding
	Swap() *gmap.Map // 获取会话的自定义存储数据
	Type() SessionType
}

type SessionType uint8

//go:generate stringer -type=SessionType

const (
	ClientSessionType   SessionType = 0
	ServerSessionType   SessionType = 1
	RecorderSessionType SessionType = 2
	PlayerSessionType   SessionType = 3
	CanvasSessionType   SessionType = 4
)
