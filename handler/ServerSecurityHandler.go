package handler

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ServerSecurityHandler vnc握手步骤第二步
// 1.发送proxy服务端支持的安全认证套件数量及类型。
// 2.读取vnc客户端支持的安全认证套件类型，判断是否支持，
// 3.选择互相支持的安全认证套件进行认证，进入认证逻辑，如果认证成功则进入下一步，认证失败则报错。
type ServerSecurityHandler struct{}

func (*ServerSecurityHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debugf(context.TODO(), "[VNC客户端->Proxy服务端]: 执行vnc握手第二步:[Security]")
	}
	cfg := session.Options()
	var secType rfb.SecurityType
	if session.ProtocolVersion() == rfb.ProtoVersion37 || session.ProtocolVersion() == rfb.ProtoVersion38 {
		if err := binary.Write(session, binary.BigEndian, uint8(len(cfg.SecurityHandlers))); err != nil {
			return err
		}

		for _, sType := range cfg.SecurityHandlers {
			if err := binary.Write(session, binary.BigEndian, sType.Type()); err != nil {
				return err
			}
		}
	} else {
		st := uint32(0)
		for _, sType := range cfg.SecurityHandlers {
			if uint32(sType.Type()) > st {
				st = uint32(sType.Type())
				secType = sType.Type()
			}
		}
		if err := binary.Write(session, binary.BigEndian, st); err != nil {
			return err
		}
	}
	if err := session.Flush(); err != nil {
		return err
	}

	if session.ProtocolVersion() == rfb.ProtoVersion38 {
		if err := binary.Read(session, binary.BigEndian, &secType); err != nil {
			return err
		}
	}
	secTypes := make(map[rfb.SecurityType]rfb.ISecurityHandler)
	for _, sType := range cfg.SecurityHandlers {
		secTypes[sType.Type()] = sType
	}

	sType, ok := secTypes[secType]
	if !ok {
		return fmt.Errorf("security type %d not implemented", secType)
	}

	var authCode uint32
	authErr := sType.Auth(session)
	if authErr != nil {
		authCode = uint32(1)
	}

	if err := binary.Write(session, binary.BigEndian, authCode); err != nil {
		return err
	}

	if authErr == nil {
		if err := session.Flush(); err != nil {
			return err
		}
		session.SetSecurityHandler(sType)
		return nil
	}

	if session.ProtocolVersion() == rfb.ProtoVersion38 {
		if err := binary.Write(session, binary.BigEndian, uint32(len(authErr.Error()))); err != nil {
			return err
		}
		if err := binary.Write(session, binary.BigEndian, []byte(authErr.Error())); err != nil {
			return err
		}
		if err := session.Flush(); err != nil {
			return err
		}
	}
	return authErr
}
