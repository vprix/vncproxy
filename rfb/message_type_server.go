package rfb

// ServerMessageType 服务端发送给客户端的消息类型
type ServerMessageType uint8

//go:generate stringer -type=ServerMessageType

const (
	FramebufferUpdate      ServerMessageType = 0   // 帧缓冲区更新消息
	SetColorMapEntries     ServerMessageType = 1   // 设置颜色地图
	Bell                   ServerMessageType = 2   // 响铃
	ServerCutText          ServerMessageType = 3   // 设置剪切板数据
	EndOfContinuousUpdates                   = 150 //结束连续更新
	ServerFence                              = 248 //支持 Fence 扩展的服务器发送此扩展以请求数据流的同步
)
