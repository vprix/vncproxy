package rfb

// ClientMessageType vnc客户端发送给vnc服务端的消息类型
type ClientMessageType uint8

//go:generate stringer -type=ClientMessageType

const (
	SetPixelFormat           ClientMessageType = 0   // 设置像素格式
	SetEncodings             ClientMessageType = 2   // 设置消息的编码格式
	FramebufferUpdateRequest ClientMessageType = 3   // 请求帧缓冲内容
	KeyEvent                 ClientMessageType = 4   // 键盘事件消息
	PointerEvent             ClientMessageType = 5   // 鼠标事件消息
	ClientCutText            ClientMessageType = 6   // 剪切板消息
	EnableContinuousUpdates  ClientMessageType = 150 // 打开连续更新
	ClientFence              ClientMessageType = 248 //客户端到服务端的数据同步请求
	SetDesktopSize           ClientMessageType = 251 //客户端设置桌面大小
	QEMUExtendedKeyEvent     ClientMessageType = 255 // qumu虚拟机的扩展按键消息
)
