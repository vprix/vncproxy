package handler

import (
	"encoding/binary"
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
)

// ClientSecurityHandler vnc握手步骤第二步
// 1. 读取vnc服务端支持的安全认证套件数量及类型
// 2. 匹配vnc服务端与proxy客户端的安全认证套件
// 3. 进入安全认证套件认证流程
// 4. 获取认证结果，如果认证失败，获取失败的原因。
type ClientSecurityHandler struct{}

func (*ClientSecurityHandler) Handle(session rfb.ISession) error {
	if logger.IsDebug() {
		logger.Debugf("[Proxy客户端->VNC服务端]: 执行vnc握手第二步:[Security]")
	}
	cfg := session.Options()
	// 读取vnc服务端支持的安全认证套件数量
	var numSecurityTypes uint8
	if err := binary.Read(session, binary.BigEndian, &numSecurityTypes); err != nil {
		return err
	}
	// 读取vnc服务端支持的安全认证套件类型
	secTypes := make([]rfb.SecurityType, numSecurityTypes)
	if err := binary.Read(session, binary.BigEndian, &secTypes); err != nil {
		return err
	}

	// 匹配vnc服务端与proxy客户端的安全认证套件
	var secType rfb.ISecurityHandler
	for _, st := range cfg.SecurityHandlers {
		for _, sc := range secTypes {
			if st.Type() == sc {
				secType = st
			}
		}
	}

	// 发送proxy客户端选中的安全认证套件
	if err := binary.Write(session, binary.BigEndian, cfg.SecurityHandlers[0].Type()); err != nil {
		return err
	}

	if err := session.Flush(); err != nil {
		return err
	}

	// 进入安全认证套件认证流程
	err := secType.Auth(session)
	if err != nil {
		return fmt.Errorf("安全认证失败, error:%v", err)
	}

	// 读取安全认证结果
	var authCode uint32
	if err := binary.Read(session, binary.BigEndian, &authCode); err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.Debugf("安全认证中, 安全认证套件类型: %d,认证结果(0为成功): %d", rfb.ClientMessageType(secType.Type()), authCode)
	}
	//如果认证失败，则读取失败原因
	if authCode == 1 {
		var reasonLength uint32
		if err = binary.Read(session, binary.BigEndian, &reasonLength); err != nil {
			return err
		}
		reasonText := make([]byte, reasonLength)
		if err = binary.Read(session, binary.BigEndian, &reasonText); err != nil {
			return err
		}
		return fmt.Errorf("%s", reasonText)
	}
	session.SetSecurityHandler(secType)
	return nil
}
