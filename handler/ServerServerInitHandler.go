package handler

import (
	"encoding/binary"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ServerServerInitHandler vnc握手步骤第四步
// 1. 发送proxy服务端的参数信息，屏幕宽高，像素格式，桌面名称
type ServerServerInitHandler struct{}

func (*ServerServerInitHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debugf("[Proxy服务端->VNC客户端]: 执行vnc握手第四步:[ServerInit]")
	}
	if err := binary.Write(session, binary.BigEndian, session.Options().Width); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.Options().Height); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.Options().PixelFormat); err != nil {
		return err
	}
	desktopName := session.Options().DesktopName
	size := uint32(len(session.Options().DesktopName))
	if size == 0 {
		desktopName = []byte("vprix")
		size = uint32(len(desktopName))
	}
	if err := binary.Write(session, binary.BigEndian, size); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, desktopName); err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.Debugf("[Proxy服务端->VNC客户端]: ServerInit[Width:%d,Height:%d,PixelFormat:%s,DesktopName:%s]",
			session.Options().Width, session.Options().Height, session.Options().PixelFormat, desktopName)
	}
	return session.Flush()
}
