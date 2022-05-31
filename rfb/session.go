package rfb

import (
	"github.com/gogf/gf/container/gmap"
	"io"
)

// ISession vnc连接的接口
type ISession interface {
	io.ReadWriteCloser
	Conn() io.ReadWriteCloser
	Config() interface{}
	ProtocolVersion() string           // 获取当前的rfb协议
	SetProtocolVersion(string)         // 设置rfb协议
	PixelFormat() PixelFormat          // 获取该会话的像素格式
	SetPixelFormat(PixelFormat) error  // 设置该会话的像素格式
	ColorMap() ColorMap                // 获取该会话的颜色地图
	SetColorMap(ColorMap)              // 设置该会话的颜色地图
	Encodings() []IEncoding            // 获取该会话支持的图像编码类型
	SetEncodings([]EncodingType) error // 设置该链接支持的图像编码类型
	Width() uint16                     //获取桌面宽度
	Height() uint16                    //获取桌面高度
	SetWidth(uint16)                   //设置桌面宽度
	SetHeight(uint16)                  // 设置桌面高度
	DesktopName() []byte               // 获取桌面名称
	SetDesktopName([]byte)             // 设置桌面名称
	Flush() error
	Wait()
	SetSecurityHandler(ISecurityHandler) error // 设置安全认证处理方法
	SecurityHandler() ISecurityHandler         // 获取当前安全认证的处理方法
	GetEncoding(EncodingType) IEncoding
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
