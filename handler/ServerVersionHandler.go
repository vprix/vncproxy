package handler

import (
	"encoding/binary"
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ServerVersionHandler vnc握手步骤第一步。
// 1. vnc客户端链接到proxy服务端后，proxy服务端发送rfb版本信息。
// 2. 发送版本信息后，接受vnc客户端返回的版本信息，进行版本匹配。
// 3. 确定版本信息是相互支持的，如果不支持，则返回错误信息，如果支持则进行下一步。
type ServerVersionHandler struct{}

func (*ServerVersionHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debugf("[VNC客户端->Proxy服务端]: 执行vnc握手第一步:[Version]")
	}
	var version [rfb.ProtoVersionLength]byte
	if err := binary.Write(session, binary.BigEndian, []byte(rfb.ProtoVersion38)); err != nil {
		return err
	}
	if err := session.Flush(); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &version); err != nil {
		return err
	}
	major, minor, err := ParseProtoVersion(version[:])
	if err != nil {
		return err
	}

	pv := rfb.ProtoVersionUnknown
	if major == 3 {
		if minor >= 8 {
			pv = rfb.ProtoVersion38
		} else if minor >= 3 {
			pv = rfb.ProtoVersion33
		}
	}
	if pv == rfb.ProtoVersionUnknown {
		return fmt.Errorf("rfb协议握手; 不支持的协议版本 '%v'", string(version[:]))
	}

	session.SetProtocolVersion(pv)
	return nil
}
