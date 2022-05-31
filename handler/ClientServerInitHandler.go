package handler

import (
	"encoding/binary"
	"github.com/vprix/vncproxy/logger"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
)

// ClientServerInitHandler vnc握手第四步
// 1. 读取vnc服务端发送的屏幕宽高，像素格式，桌面名称
type ClientServerInitHandler struct{}

func (*ClientServerInitHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debugf("[Proxy客户端->VNC服务端]: 执行vnc握手第四步:[ServerInit]")
	}
	var err error
	srvInit := messages.ServerInit{}

	if err = binary.Read(session, binary.BigEndian, &srvInit.FBWidth); err != nil {
		return err
	}
	if err = binary.Read(session, binary.BigEndian, &srvInit.FBHeight); err != nil {
		return err
	}
	if err = binary.Read(session, binary.BigEndian, &srvInit.PixelFormat); err != nil {
		return err
	}
	if err = binary.Read(session, binary.BigEndian, &srvInit.NameLength); err != nil {
		return err
	}

	srvInit.NameText = make([]byte, srvInit.NameLength)
	if err = binary.Read(session, binary.BigEndian, &srvInit.NameText); err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.Debugf("[Proxy客户端->VNC服务端]:  serverInit: %s", srvInit)
	}
	session.SetDesktopName(srvInit.NameText)
	// 如果协议是aten1，则执行特殊的逻辑
	if session.ProtocolVersion() == "aten1" {
		session.SetWidth(800)
		session.SetHeight(600)
		// 发送像素格式消息
		err = session.SetPixelFormat(rfb.NewPixelFormatAten())
		if err != nil {
			return err
		}
	} else {
		session.SetWidth(srvInit.FBWidth)
		session.SetHeight(srvInit.FBHeight)

		//告诉vnc服务端，proxy客户端支持的像素格式，发送`SetPixelFormat`消息
		pixelMsg := messages.SetPixelFormat{PF: rfb.PixelFormat32bit}
		err = pixelMsg.Write(session)
		if err != nil {
			return err
		}
		err = session.SetPixelFormat(rfb.PixelFormat32bit)
		if err != nil {
			return err
		}
	}
	// aten1协议需要再次读取扩展信息
	if session.ProtocolVersion() == "aten1" {
		ikvm := struct {
			_               [8]byte
			IKVMVideoEnable uint8
			IKVMKMEnable    uint8
			IKVMKickEnable  uint8
			VUSBEnable      uint8
		}{}
		if err = binary.Read(session, binary.BigEndian, &ikvm); err != nil {
			return err
		}
	}
	return nil
}
