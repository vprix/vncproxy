package handler

import (
	"encoding/binary"
	"github.com/vprix/vncproxy/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ClientClientInitHandler vnc握手步骤第三步
// 1. 根据配置信息判断该vnc会话是否独占，
// 2. 发送是否独占标识给vnc服务端
type ClientClientInitHandler struct{}

func (that *ClientClientInitHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debug("[Proxy客户端->VNC服务端]: 执行vnc握手步骤第三步[ClientInit]")
	}
	cfg := session.Config().(*rfb.ClientConfig)
	var shared uint8
	if cfg.Exclusive {
		shared = 0
	} else {
		shared = 1
	}
	if err := binary.Write(session, binary.BigEndian, shared); err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.Debugf("[Proxy客户端->VNC服务端]: 执行ClientInit步骤，发送shared=%d", shared)
	}
	return session.Flush()
}
