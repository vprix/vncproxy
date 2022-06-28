package vnc

import (
	"github.com/gogf/gf/util/gconv"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/session"
)

type Proxy struct {
	svrSession rfb.ISession // vnc客户端连接到proxy的会话
	cliSession rfb.ISession // 链接到vnc服务端的会话
	errorCh    chan error
	closed     chan struct{}
	//proxyCli2VncSvrMsgChan chan rfb.Message // proxy客户端发送给vnc服务端的消息通道
	//vncSvr2ProxyMsgChan    chan rfb.Message // vnc服务端发送给proxy客户端的消息通道
	//vncCli2ProxyMsgChan    chan rfb.Message // vnc客户端发送给proxy服务端的消息通道
	//proxySvr2VncCliMsgChan chan rfb.Message // proxy服务端发送给vnc客户端的消息通道
}

// NewVncProxy 生成vnc proxy服务对象
func NewVncProxy(serverSession *session.ServerSession, cliSession *session.ClientSession) *Proxy {
	vncProxy := &Proxy{
		svrSession: serverSession,
		cliSession: cliSession,
		errorCh:    make(chan error, 1024),
		closed:     make(chan struct{}),
	}
	//if svrCfg == nil {
	//	svrCfg = &rfb.Options{
	//		Encodings:   encodings.DefaultEncodings,
	//		DesktopName: []byte("Vprix VNC Proxy"),
	//		Width:       1024,
	//		Height:      768,
	//		SecurityHandlers: []rfb.ISecurityHandler{
	//			&security.ServerAuthNone{},
	//		},
	//		//DisableMessageType: []rfb.ServerMessageType{rfb.ServerCutText},
	//	}
	//}
	//
	//if cliCfg == nil {
	//	cliCfg = &rfb.Options{
	//		SecurityHandlers: []rfb.ISecurityHandler{&security.ClientAuthVNC{Password: vncProxy.targetConfig.Password}},
	//		Encodings:        encodings.DefaultEncodings,
	//		ErrorCh:          make(chan error),
	//		Input:            vncProxy.vncSvr2ProxyMsgChan,
	//		Output:           vncProxy.proxyCli2VncSvrMsgChan,
	//		Handlers:         session.DefaultClientHandlers,
	//		Messages:         messages.DefaultServerMessages,
	//	}
	//}
	//vncProxy.rfbSvrCfg = svrCfg
	//vncProxy.rfbCliCfg = cliCfg
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
	that.svrSession.Run()
	return nil
}

func (that *Proxy) handleIO() {

	for {
		select {
		case msg := <-that.cliSession.Options().ErrorCh:
			// 如果链接到vnc服务端的会话报错，则需要把链接到proxy的vnc客户端全部关闭
			_ = that.svrSession.Close()
			that.errorCh <- msg
		case msg := <-that.svrSession.Options().ErrorCh:
			//  链接到proxy的vnc客户端链接报错，则把错误转发给vnc proxy
			that.errorCh <- msg
		case msg := <-that.cliSession.Options().Output:
			// 收到vnc服务端发送给proxy客户端的消息，转发给proxy服务端, proxy服务端内部会把该消息转发给vnc客户端
			sSessCfg := that.svrSession.Options()
			disabled := false
			// 如果该消息禁用，则跳过不转发该消息
			for _, t := range sSessCfg.DisableMessageType {
				if t == msg.Type() {
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
				that.cliSession.Desktop().SetPixelFormat(msg.(*messages.SetPixelFormat).PF)
				that.cliSession.Options().Input <- msg
				continue
			case rfb.SetEncodings:
				// 设置编码格式的消息
				var encTypes []rfb.EncodingType
				// 判断编码是否再支持的列表
				for _, s := range that.cliSession.Encodings() {
					for _, cEnc := range msg.(*messages.SetEncodings).Encodings {
						if cEnc == s.Type() {
							encTypes = append(encTypes, s.Type())
						}
					}
				}
				// 发送编码消息给vnc服务端
				that.cliSession.Options().Input <- &messages.SetEncodings{EncNum: gconv.Uint16(len(encTypes)), Encodings: encTypes}
			default:
				cliCfg := that.cliSession.Options()
				disabled := false
				for _, t := range cliCfg.DisableMessageType {
					if t == msg.Type() {
						disabled = true
						break
					}
				}
				if !disabled {
					that.svrSession.Options().Input <- msg
				}
			}
		}
	}
}

// Handle 建立远程链接
func (that *Proxy) Handle(sess rfb.ISession) (err error) {

	that.cliSession.Run()
	if err != nil {
		return err
	}
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.Desktop().SetWidth(that.cliSession.Desktop().Width())
	that.svrSession.Desktop().SetHeight(that.cliSession.Desktop().Height())
	that.svrSession.Desktop().SetDesktopName(that.cliSession.Desktop().DesktopName())
	that.svrSession.Desktop().SetPixelFormat(that.cliSession.Desktop().PixelFormat())

	go that.handleIO()
	return nil
}

func (that *Proxy) Close() {
	that.closed <- struct{}{}
}

func (that *Proxy) Error() <-chan error {
	return that.errorCh
}
