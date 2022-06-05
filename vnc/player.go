package vnc

import (
	"encoding/binary"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"io"
	"time"
)

type Player struct {
	rfbSvrCfg     *rfb.ServerConfig      // proxy服务端监听vnc客户端的配置信息
	svrSession    *session.ServerSession // vnc客户端连接到proxy的会话
	playerSession *session.PlayerSession
	closed        chan struct{}
}

func NewPlayer(filePath string, svrCfg *rfb.ServerConfig) *Player {
	if svrCfg == nil {
		svrCfg = &rfb.ServerConfig{
			PixelFormat:      rfb.PixelFormat32bit,
			Messages:         messages.DefaultClientMessage,
			Encodings:        encodings.DefaultEncodings,
			Input:            make(chan rfb.ClientMessage),
			Output:           make(chan rfb.ServerMessage),
			ErrorCh:          make(chan error),
			SecurityHandlers: []rfb.ISecurityHandler{&security.ServerAuthNone{}},
		}
	}
	play := &Player{
		closed:    make(chan struct{}),
		rfbSvrCfg: svrCfg,
		playerSession: session.NewPlayerSession(filePath,
			&rfb.ServerConfig{
				ErrorCh:   svrCfg.ErrorCh,
				Encodings: encodings.DefaultEncodings,
			},
		),
	}

	return play
}

// Start 启动
func (that *Player) Start(conn io.ReadWriteCloser) error {

	that.rfbSvrCfg.Handlers = []rfb.IHandler{
		&handler.ServerVersionHandler{},
		&handler.ServerSecurityHandler{},
		that, // 把链接到vnc服务端的逻辑加入
		&handler.ServerClientInitHandler{},
		&handler.ServerServerInitHandler{},
		&handler.ServerMessageHandler{},
	}
	go func() {
		err := session.NewServerSession(conn, that.rfbSvrCfg).Server()
		if err != nil {
			that.rfbSvrCfg.ErrorCh <- err
		}
	}()

	return nil
}

// Handle 建立远程链接
func (that *Player) Handle(sess rfb.ISession) error {
	err := that.playerSession.Connect()
	if err != nil {
		return err
	}
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.SetWidth(that.playerSession.Width())
	that.svrSession.SetHeight(that.playerSession.Height())
	that.svrSession.SetDesktopName(that.playerSession.DesktopName())
	_ = that.svrSession.SetPixelFormat(that.playerSession.PixelFormat())

	go that.handleIO()
	return nil
}

func (that *Player) handleIO() {
	for {
		select {
		case <-that.closed:
			return
		case msg := <-that.rfbSvrCfg.Input:
			if logger.IsDebug() {
				logger.Debugf("收到vnc客户端发送过来的消息,%s", msg)
			}
		default:
			// 从会话中读取消息类型
			var messageType rfb.ServerMessageType
			if err := binary.Read(that.playerSession, binary.BigEndian, &messageType); err != nil {
				that.rfbSvrCfg.ErrorCh <- err
				return
			}
			msg := &messages.FramebufferUpdate{}
			// 读取消息内容
			parsedMsg, err := msg.Read(that.playerSession)
			if err != nil {
				that.rfbSvrCfg.ErrorCh <- err
				return
			}
			err = parsedMsg.Write(that.svrSession)
			if err != nil {
				that.rfbSvrCfg.ErrorCh <- err
				return
			}
			var sleep int64
			_ = binary.Read(that.playerSession, binary.BigEndian, &sleep)
			if sleep > 0 {
				time.Sleep(time.Duration(sleep))
			}
		}
	}
}

func (that *Player) Close() {
	if that.closed != nil {
		close(that.closed)
		_ = that.svrSession.Close()
		_ = that.playerSession.Close()
	}
}

func (that *Player) Error() <-chan error {
	return that.rfbSvrCfg.ErrorCh
}
