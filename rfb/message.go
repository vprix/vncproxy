package rfb

type Message interface {
	Type() MessageType
	String() string
	Supported(ISession) bool
	Read(ISession) (Message, error)
	Write(ISession) error
	Clone() Message
}
