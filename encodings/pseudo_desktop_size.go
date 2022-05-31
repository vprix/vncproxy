package encodings

import (
	"github.com/vprix/vncproxy/rfb"
)

// DesktopSizePseudoEncoding 如果客户端请求桌面大小伪编码，那么就是说它能处理帧缓存宽/高的改变。
// 服务器通过发送带有桌面大小伪编码的伪矩形作为上一个矩形来完成一次更新。
// 伪矩形的x 和y 被忽略，而宽和高表示帧缓存新的宽和高。没有其他的数据与伪矩形有关。
type DesktopSizePseudoEncoding struct {
}

func (that *DesktopSizePseudoEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *DesktopSizePseudoEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &DesktopSizePseudoEncoding{}
	return obj
}

func (that *DesktopSizePseudoEncoding) Type() rfb.EncodingType { return rfb.EncDesktopSizePseudo }

// Read implements the Encoding interface.
func (that *DesktopSizePseudoEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	return nil
}

func (that *DesktopSizePseudoEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	return nil
}
