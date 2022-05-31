package rfb

// SecurityType 安全认证类型
type SecurityType uint8

//go:generate stringer -type=SecurityType

const (
	SecTypeUnknown  SecurityType = SecurityType(0)  // 未知认证类型
	SecTypeNone     SecurityType = SecurityType(1)  // 不需要认证
	SecTypeVNC      SecurityType = SecurityType(2)  // vnc密码认证
	SecTypeTight    SecurityType = SecurityType(16) // tight vnc主导的认证模式
	SecTypeVeNCrypt SecurityType = SecurityType(19) // VeNCrypt 通用认证类型
)

// SecuritySubType 认证子类型
type SecuritySubType uint32

//go:generate stringer -type=SecuritySubType

// SecSubTypeUnknown 未知的子类型认证
const (
	SecSubTypeUnknown SecuritySubType = SecuritySubType(0)
)

// VeNCrypt 安全认证会有两种认证版本0.1和0.2
// 以下表示0.1
const (
	SecSubTypeVeNCrypt01Unknown   SecuritySubType = SecuritySubType(0)
	SecSubTypeVeNCrypt01Plain     SecuritySubType = SecuritySubType(19)
	SecSubTypeVeNCrypt01TLSNone   SecuritySubType = SecuritySubType(20)
	SecSubTypeVeNCrypt01TLSVNC    SecuritySubType = SecuritySubType(21)
	SecSubTypeVeNCrypt01TLSPlain  SecuritySubType = SecuritySubType(22)
	SecSubTypeVeNCrypt01X509None  SecuritySubType = SecuritySubType(23)
	SecSubTypeVeNCrypt01X509VNC   SecuritySubType = SecuritySubType(24)
	SecSubTypeVeNCrypt01X509Plain SecuritySubType = SecuritySubType(25)
)

// 以下表示0.2版本的类型
const (
	SecSubTypeVeNCrypt02Unknown   SecuritySubType = SecuritySubType(0)
	SecSubTypeVeNCrypt02Plain     SecuritySubType = SecuritySubType(256)
	SecSubTypeVeNCrypt02TLSNone   SecuritySubType = SecuritySubType(257)
	SecSubTypeVeNCrypt02TLSVNC    SecuritySubType = SecuritySubType(258)
	SecSubTypeVeNCrypt02TLSPlain  SecuritySubType = SecuritySubType(259)
	SecSubTypeVeNCrypt02X509None  SecuritySubType = SecuritySubType(260)
	SecSubTypeVeNCrypt02X509VNC   SecuritySubType = SecuritySubType(261)
	SecSubTypeVeNCrypt02X509Plain SecuritySubType = SecuritySubType(262)
)

// ISecurityHandler 认证方式的接口
type ISecurityHandler interface {
	Type() SecurityType
	SubType() SecuritySubType
	Auth(ISession) error
}
