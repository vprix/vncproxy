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
	if err := binary.Write(session, binary.BigEndian, session.Desktop().Width()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.Desktop().Height()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.Desktop().PixelFormat()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, uint32(len(session.Desktop().DesktopName()))); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.Desktop().DesktopName()); err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.Debugf("[Proxy服务端->VNC客户端]: ServerInit[Width:%d,Height:%d,PixelFormat:%s,DesktopName:%s]",
			session.Desktop().Width(), session.Desktop().Height(), session.Desktop().PixelFormat(), session.Desktop().DesktopName())
	}
	return session.Flush()
}
