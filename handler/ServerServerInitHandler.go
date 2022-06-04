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
	if err := binary.Write(session, binary.BigEndian, session.Width()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.Height()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.PixelFormat()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, uint32(len(session.DesktopName()))); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, session.DesktopName()); err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.Debugf("[Proxy服务端->VNC客户端]: ServerInit[Width:%d,Height:%d,PixelFormat:%s,DesktopName:%s]",
			session.Width(), session.Height(), session.PixelFormat(), session.DesktopName())
	}
	return session.Flush()
}
