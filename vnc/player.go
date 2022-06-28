package vnc

import (
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/os/gfile"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"io"
	"os"
	"time"
)

type Player struct {
	rfbSvrCfg     *rfb.Options           // proxy服务端监听vnc客户端的配置信息
	svrSession    *session.ServerSession // vnc客户端连接到proxy的会话
	playerSession *session.PlayerSession
	closed        chan struct{}
}

func NewPlayer(filePath string, svrCfg *rfb.Options) *Player {
	if svrCfg == nil {
		svrCfg = &rfb.Options{
			PixelFormat:      rfb.PixelFormat32bit,
			Messages:         messages.DefaultClientMessage,
			Encodings:        encodings.DefaultEncodings,
			Input:            make(chan rfb.Message),
			Output:           make(chan rfb.Message),
			ErrorCh:          make(chan error),
			SecurityHandlers: []rfb.ISecurityHandler{&security.ServerAuthNone{}},
		}
	}
	play := &Player{
		closed:    make(chan struct{}),
		rfbSvrCfg: svrCfg,
		playerSession: session.NewPlayerSession(
			&rfb.Options{
				ErrorCh:   svrCfg.ErrorCh,
				Encodings: encodings.DefaultEncodings,
				CreateConn: func() (io.ReadWriteCloser, error) {
					if !gfile.Exists(filePath) {
						return nil, fmt.Errorf("要保存的文件[%s]不存在", filePath)
					}
					return gfile.OpenFile(filePath, os.O_RDONLY, 0644)
				},
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
	that.rfbSvrCfg.CreateConn = func() (io.ReadWriteCloser, error) {
		return conn, nil
	}
	go func() {
		sess, err := session.NewServerSession(that.rfbSvrCfg)
		if err != nil {
			logger.Error(err)
			return
		}
		sess.Run()
	}()

	return nil
}

// Handle 建立远程链接
func (that *Player) Handle(sess rfb.ISession) error {
	that.playerSession.Run()
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.Desktop().SetWidth(that.playerSession.Desktop().Width())
	that.svrSession.Desktop().SetHeight(that.playerSession.Desktop().Height())
	that.svrSession.Desktop().SetDesktopName(that.playerSession.Desktop().DesktopName())
	that.svrSession.Desktop().SetPixelFormat(that.playerSession.Desktop().PixelFormat())

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
