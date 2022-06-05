package vnc

import (
	"github.com/gogf/gf/util/gconv"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"io"
	"net"
	"time"
)

type Proxy struct {
	rfbSvrCfg              *rfb.ServerConfig      // proxy服务端监听vnc客户端的配置信息
	rfbCliCfg              *rfb.ClientConfig      // proxy客户端连接vnc服务端的配置信息
	targetConfig           rfb.TargetConfig       // vnc服务端的链接参数
	svrSession             *session.ServerSession // vnc客户端连接到proxy的会话
	cliSession             *session.ClientSession // 链接到vnc服务端的会话
	closed                 chan struct{}
	errorCh                chan error             // 错误通道
	proxyCli2VncSvrMsgChan chan rfb.ClientMessage // proxy客户端发送给vnc服务端的消息通道
	vncSvr2ProxyMsgChan    chan rfb.ServerMessage // vnc服务端发送给proxy客户端的消息通道
	vncCli2ProxyMsgChan    chan rfb.ClientMessage // vnc客户端发送给proxy服务端的消息通道
	proxySvr2VncCliMsgChan chan rfb.ServerMessage // proxy服务端发送给vnc客户端的消息通道
}

// NewVncProxy 生成vnc proxy服务对象
func NewVncProxy(svrCfg *rfb.ServerConfig, cliCfg *rfb.ClientConfig, targetCfg rfb.TargetConfig) *Proxy {
	errorChan := make(chan error, 32)
	vncProxy := &Proxy{
		errorCh:                errorChan,
		targetConfig:           targetCfg,
		closed:                 make(chan struct{}),
		proxyCli2VncSvrMsgChan: make(chan rfb.ClientMessage),
		vncSvr2ProxyMsgChan:    make(chan rfb.ServerMessage),
		vncCli2ProxyMsgChan:    make(chan rfb.ClientMessage),
		proxySvr2VncCliMsgChan: make(chan rfb.ServerMessage),
	}
	if svrCfg == nil {
		svrCfg = &rfb.ServerConfig{
			Encodings:   encodings.DefaultEncodings,
			DesktopName: []byte("Vprix VNC Proxy"),
			Width:       1024,
			Height:      768,
			SecurityHandlers: []rfb.ISecurityHandler{
				&security.ServerAuthNone{},
			},
			//DisableMessageType: []rfb.ServerMessageType{rfb.ServerCutText},
		}
	}

	if cliCfg == nil {
		cliCfg = &rfb.ClientConfig{
			SecurityHandlers: []rfb.ISecurityHandler{&security.ClientAuthVNC{Password: vncProxy.targetConfig.Password}},
			Encodings:        encodings.DefaultEncodings,
			ErrorCh:          make(chan error),
			Input:            vncProxy.vncSvr2ProxyMsgChan,
			Output:           vncProxy.proxyCli2VncSvrMsgChan,
			Handlers:         session.DefaultClientHandlers,
			Messages:         messages.DefaultServerMessages,
		}
	}
	vncProxy.rfbSvrCfg = svrCfg
	vncProxy.rfbCliCfg = cliCfg
	return vncProxy
}

// Start 启动
func (that *Proxy) Start(conn io.ReadWriteCloser) {

	that.rfbSvrCfg.Input = that.vncCli2ProxyMsgChan
	that.rfbSvrCfg.Output = that.proxySvr2VncCliMsgChan
	that.rfbSvrCfg.ErrorCh = make(chan error)
	if len(that.rfbSvrCfg.Messages) <= 0 {
		that.rfbSvrCfg.Messages = messages.DefaultClientMessage
	}

	that.rfbSvrCfg.Handlers = []rfb.IHandler{
		&handler.ServerVersionHandler{},
		&handler.ServerSecurityHandler{},
		that, // 把链接到vnc服务端的逻辑加入
		&handler.ServerClientInitHandler{},
		&handler.ServerServerInitHandler{},
		&handler.ServerMessageHandler{},
	}
	err := session.NewServerSession(conn, that.rfbSvrCfg).Server()
	if err != nil {
		that.errorCh <- err
	}
	return
}

func (that *Proxy) handleIO() {

	for {
		select {
		case msg := <-that.rfbCliCfg.ErrorCh:
			// 如果链接到vnc服务端的会话报错，则需要把链接到proxy的vnc客户端全部关闭
			_ = that.svrSession.Close()
			that.errorCh <- msg
		case msg := <-that.rfbSvrCfg.ErrorCh:
			//  链接到proxy的vnc客户端链接报错，则把错误转发给vnc proxy
			that.errorCh <- msg
		case msg := <-that.vncSvr2ProxyMsgChan:
			// 收到vnc服务端发送给proxy客户端的消息，转发给proxy服务端, proxy服务端内部会把该消息转发给vnc客户端
			sSessCfg := that.svrSession.Config().(*rfb.ServerConfig)
			disabled := false
			// 如果该消息禁用，则跳过不转发该消息
			for _, t := range sSessCfg.DisableMessageType {
				if t == msg.Type() {
					disabled = true
					break
				}
			}
			if !disabled {
				sSessCfg.Output <- msg
			}
		case msg := <-that.vncCli2ProxyMsgChan:
			// vnc客户端发送消息到proxy服务端的时候,需要对消息进行检查
			// 有些消息不支持转发给vnc服务端
			switch msg.Type() {
			case rfb.SetPixelFormat:
				// 发现是设置像素格式的消息，则忽略
				//that.rfbCliCfg.PixelFormat = msg.(*messages.SetPixelFormat).PF
				_ = that.cliSession.SetPixelFormat(msg.(*messages.SetPixelFormat).PF)
				that.proxyCli2VncSvrMsgChan <- msg
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
				that.proxyCli2VncSvrMsgChan <- &messages.SetEncodings{EncNum: gconv.Uint16(len(encTypes)), Encodings: encTypes}
			default:
				cliCfg := that.cliSession.Config().(*rfb.ClientConfig)
				disabled := false
				for _, t := range cliCfg.DisableMessageType {
					if t == msg.Type() {
						disabled = true
						break
					}
				}
				if !disabled {
					that.proxyCli2VncSvrMsgChan <- msg
				}
			}
		}
	}
}

// Handle 建立远程链接
func (that *Proxy) Handle(sess rfb.ISession) error {
	timeout := 10 * time.Second
	if that.targetConfig.Timeout > 0 {
		timeout = that.targetConfig.Timeout
	}
	network := "tcp"
	if len(that.targetConfig.Network) > 0 {
		network = that.targetConfig.Network
	}

	clientConn, err := net.DialTimeout(network, that.targetConfig.Addr(), timeout)
	if err != nil {
		return err
	}
	that.cliSession, err = session.NewClient(clientConn, that.rfbCliCfg)
	if err != nil {
		return err
	}
	err = that.cliSession.Connect()
	if err != nil {
		return err
	}
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.SetWidth(that.cliSession.Width())
	that.svrSession.SetHeight(that.cliSession.Height())
	that.svrSession.SetDesktopName(that.cliSession.DesktopName())
	_ = that.svrSession.SetPixelFormat(that.cliSession.PixelFormat())

	go that.handleIO()
	//go func() {
	//
	//	ticker := time.Tick(2 * time.Second)
	//	for true {
	//		select {
	//		case <-ticker:
	//			buff := &bytes.Buffer{}
	//			err = jpeg.Encode(buff, that.cliSession.Canvas, &jpeg.Options{Quality: 100})
	//			if err != nil {
	//				glog.Error(err)
	//				return
	//			}
	//			filename := fmt.Sprintf("D:\\code\\GolandProjects\\vprix-vnc\\abc_%d.jpeg", grand.Intn(10000))
	//			err = gfile.PutBytes(filename, buff.Bytes())
	//			if err != nil {
	//				glog.Error(err)
	//				return
	//			}
	//		}
	//	}
	//
	//}()
	return nil
}

func (that *Proxy) Close() {
	that.closed <- struct{}{}
	close(that.proxySvr2VncCliMsgChan)
	close(that.proxyCli2VncSvrMsgChan)
	close(that.vncCli2ProxyMsgChan)
	close(that.vncSvr2ProxyMsgChan)
}

func (that *Proxy) Error() <-chan error {
	return that.errorCh
}
