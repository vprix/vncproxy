package handler

import (
	"encoding/binary"
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ClientVersionHandler vnc握手第一步
// 1. 连接到vnc服务端后,读取其支持的rfb协议版本。
// 2. 解析版本，判断该版本proxy客户端是否支持。
// 3. 如果支持该版本，则发送支持的版本给vnc服务端
type ClientVersionHandler struct{}

func (*ClientVersionHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debugf("[Proxy客户端->VNC服务端]: 执行vnc握手第一步:[Version]")
	}
	var version [rfb.ProtoVersionLength]byte

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
			pv = rfb.ProtoVersion38
		}
	}
	if pv == rfb.ProtoVersionUnknown {
		return fmt.Errorf("rfb协议握手失败; 不支持的版本 '%v'", string(version[:]))
	}
	session.SetProtocolVersion(string(version[:]))

	if err = binary.Write(session, binary.BigEndian, []byte(pv)); err != nil {
		return err
	}
	return session.Flush()
}

func ParseProtoVersion(pv []byte) (uint, uint, error) {
	var major, minor uint

	if len(pv) < rfb.ProtoVersionLength {
		return 0, 0, fmt.Errorf("协议版本的长度太短 (%v < %v)", len(pv), rfb.ProtoVersionLength)
	}

	l, err := fmt.Sscanf(string(pv), "RFB %d.%d\n", &major, &minor)
	if l != 2 {
		return 0, 0, fmt.Errorf("解析rfb协议失败")
	}
	if err != nil {
		return 0, 0, err
	}

	return major, minor, nil
}
