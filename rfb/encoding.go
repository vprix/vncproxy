package rfb

// IEncoding vnc像素数据编码格式的接口定义
type IEncoding interface {
	Type() EncodingType
	Supported(ISession) bool
	Clone(...bool) IEncoding
	Read(ISession, *Rectangle) error
	Write(ISession, *Rectangle) error
}
