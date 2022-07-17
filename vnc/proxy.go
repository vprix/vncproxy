package vnc

import (
	"github.com/gogf/gf/util/gconv"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/session"
)

type Proxy struct {
	remoteSession rfb.ISession // 链接到vnc远端服务的会话
	svrSession    rfb.ISession // vnc客户端连接到proxy的会话

	errorCh chan error
	closed  chan struct{}
}

// NewVncProxy 生成vnc proxy服务对象
func NewVncProxy(remoteSession *session.ClientSession, serverSession *session.ServerSession) *Proxy {
	vncProxy := &Proxy{
		svrSession:    serverSession,
		remoteSession: remoteSession,
		errorCh:       make(chan error, 1024),
		closed:        make(chan struct{}),
	}
	return vncProxy
}

// Start 启动
func (that *Proxy) Start() error {

	hds := []rfb.IHandler{
		&handler.ServerVersionHandler{},
		&handler.ServerSecurityHandler{},
		that, // 把链接到vnc服务端的逻辑加入
		&handler.ServerClientInitHandler{},
		&handler.ServerServerInitHandler{},
		&handler.ServerMessageHandler{},
	}
	err := that.svrSession.Init(
		rfb.OptHandlers(hds...),
	)
	if err != nil {
		return err
	}
	that.svrSession.Start()
	return nil
}

func (that *Proxy) handleIO() {

	for {
		select {
		case msg := <-that.remoteSession.Options().ErrorCh:
			// 如果链接到vnc服务端的会话报错，则需要把链接到proxy的vnc客户端全部关闭
			_ = that.svrSession.Close()
			that.errorCh <- msg
		case msg := <-that.svrSession.Options().ErrorCh:
			//  链接到proxy的vnc客户端链接报错，则把错误转发给vnc proxy
			that.errorCh <- msg
		case msg := <-that.remoteSession.Options().Output:
			// 收到vnc服务端发送给proxy客户端的消息，转发给proxy服务端, proxy服务端内部会把该消息转发给vnc客户端
			sSessCfg := that.svrSession.Options()
			disabled := false
			// 如果该消息禁用，则跳过不转发该消息
			for _, t := range sSessCfg.DisableServerMessageType {
				if rfb.MessageType(t) == msg.Type() {
					disabled = true
					break
				}
			}
			if !disabled {
				sSessCfg.Input <- msg
			}
		case msg := <-that.svrSession.Options().Output:
			// vnc客户端发送消息到proxy服务端的时候,需要对消息进行检查
			// 有些消息不支持转发给vnc服务端
			switch rfb.ClientMessageType(msg.Type()) {
			case rfb.SetPixelFormat:
				// 发现是设置像素格式的消息，则忽略
				//that.rfbCliCfg.PixelFormat = msg.(*messages.SetPixelFormat).PF
				that.remoteSession.SetPixelFormat(msg.(*messages.SetPixelFormat).PF)
				that.remoteSession.Options().Input <- msg
				continue
			case rfb.SetEncodings:
				// 设置编码格式的消息
				var encTypes []rfb.EncodingType
				// 判断编码是否再支持的列表
				for _, s := range that.remoteSession.Encodings() {
					for _, cEnc := range msg.(*messages.SetEncodings).Encodings {
						if cEnc == s.Type() {
							encTypes = append(encTypes, s.Type())
						}
					}
				}
				// 发送编码消息给vnc服务端
				that.remoteSession.Options().Input <- &messages.SetEncodings{EncNum: gconv.Uint16(len(encTypes)), Encodings: encTypes}
			default:
				cliCfg := that.remoteSession.Options()
				disabled := false
				for _, t := range cliCfg.DisableClientMessageType {
					if rfb.MessageType(t) == msg.Type() {
						disabled = true
						break
					}
				}
				if !disabled {
					that.remoteSession.Options().Input <- msg
				}
			}
		}
	}
}

// Handle 建立远程链接
func (that *Proxy) Handle(sess rfb.ISession) (err error) {

	that.remoteSession.Start()
	if err != nil {
		return err
	}
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.SetWidth(that.remoteSession.Options().Width)
	that.svrSession.SetHeight(that.remoteSession.Options().Height)
	that.svrSession.SetDesktopName(that.remoteSession.Options().DesktopName)
	that.svrSession.SetPixelFormat(that.remoteSession.Options().PixelFormat)

	go that.handleIO()
	return nil
}

func (that *Proxy) Wait() <-chan struct{} {
	return that.closed
}

func (that *Proxy) Close() {
	that.closed <- struct{}{}
}

func (that *Proxy) Error() <-chan error {
	return that.errorCh
}
