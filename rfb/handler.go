package rfb

type IHandler interface {
	Handle(session ISession) error
}

// ProtoVersionLength rfb协议长度
const ProtoVersionLength = 12

const (
	// ProtoVersionUnknown 未知协议
	ProtoVersionUnknown = ""
	// ProtoVersion33 版本 003.003
	ProtoVersion33 = "RFB 003.003\n"
	// ProtoVersion38 版本 003.008
	ProtoVersion38 = "RFB 003.008\n"
	// ProtoVersion37 版本 003.007
	ProtoVersion37 = "RFB 003.007\n"
)
