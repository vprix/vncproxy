package security

import "github.com/vprix/vncproxy/rfb"

// ClientAuthNone vnc客户端认证
type ClientAuthNone struct{}

// ServerAuthNone 服务端认证
type ServerAuthNone struct{}

var _ rfb.ISecurityHandler = new(ClientAuthNone)
var _ rfb.ISecurityHandler = new(ServerAuthNone)

func (*ClientAuthNone) Type() rfb.SecurityType {
	return rfb.SecTypeNone
}
func (*ClientAuthNone) SubType() rfb.SecuritySubType {
	return rfb.SecSubTypeUnknown
}

func (*ClientAuthNone) Auth(rfb.ISession) error {
	return nil
}

func (*ServerAuthNone) Type() rfb.SecurityType {
	return rfb.SecTypeNone
}

func (*ServerAuthNone) SubType() rfb.SecuritySubType {
	return rfb.SecSubTypeUnknown
}

func (*ServerAuthNone) Auth(rfb.ISession) error {
	return nil
}
