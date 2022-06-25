package rfb

type MessageType uint8

// ClientMessageType vnc客户端发送给vnc服务端的消息类型
type ClientMessageType MessageType

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

// ServerMessageType 服务端发送给客户端的消息类型
type ServerMessageType MessageType

//go:generate stringer -type=ServerMessageType

const (
	FramebufferUpdate      ServerMessageType = 0   // 帧缓冲区更新消息
	SetColorMapEntries     ServerMessageType = 1   // 设置颜色地图
	Bell                   ServerMessageType = 2   // 响铃
	ServerCutText          ServerMessageType = 3   // 设置剪切板数据
	EndOfContinuousUpdates ServerMessageType = 150 //结束连续更新
	ServerFence            ServerMessageType = 248 //支持 Fence 扩展的服务器发送此扩展以请求数据流的同步
)
