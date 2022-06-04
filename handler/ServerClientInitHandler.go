package handler

import (
	"encoding/binary"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ServerClientInitHandler vnc握手步骤第三步
// 读取vnc客户端发送的是否支持共享屏幕标识
type ServerClientInitHandler struct{}

func (*ServerClientInitHandler) Handle(session rfb.ISession) error {

	if logger.IsDebug() {
		logger.Debugf("[VNC客户端->Proxy服务端]: 执行vnc握手第三步:[ClientInit]")
	}
	// 读取分享屏幕标识符，proxy会无视该标识，因为通过proxy链接的vnc服务端都是默认支持分享的。
	var shared uint8
	if err := binary.Read(session, binary.BigEndian, &shared); err != nil {
		return err
	}
	return nil
}
