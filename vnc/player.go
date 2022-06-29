package vnc

import (
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/os/gfile"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/session"
	"io"
	"os"
	"time"
)

type Player struct {
	svrSession    *session.ServerSession // vnc客户端连接到proxy的会话
	playerSession *session.PlayerSession
	errorCh       chan error
}

func NewPlayer(filePath string, svrSession *session.ServerSession) *Player {
	playerSession := session.NewPlayerSession(
		rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
			if !gfile.Exists(filePath) {
				return nil, fmt.Errorf("要读取的文件[%s]不存在", filePath)
			}
			return gfile.OpenFile(filePath, os.O_RDONLY, 0644)
		}),
	)

	return &Player{
		errorCh:       make(chan error, 32),
		svrSession:    svrSession,
		playerSession: playerSession,
	}
}

// Start 启动
func (that *Player) Start() error {

	_ = that.svrSession.Init(rfb.OptHandlers([]rfb.IHandler{
		&handler.ServerVersionHandler{},
		&handler.ServerSecurityHandler{},
		that, // 把链接到vnc服务端的逻辑加入
		&handler.ServerClientInitHandler{},
		&handler.ServerServerInitHandler{},
		&handler.ServerMessageHandler{},
	}...))

	that.svrSession.Start()
	return nil
}

// Handle 建立远程链接
func (that *Player) Handle(sess rfb.ISession) error {
	that.playerSession.Start()
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.SetWidth(that.playerSession.Options().Width)
	that.svrSession.SetHeight(that.playerSession.Options().Height)
	that.svrSession.SetDesktopName(that.playerSession.Options().DesktopName)
	that.svrSession.SetPixelFormat(that.playerSession.Options().PixelFormat)

	go that.handleIO()
	return nil
}

func (that *Player) handleIO() {
	for {
		select {
		case <-that.svrSession.Wait():
			return
		case <-that.playerSession.Wait():
			return
		case err := <-that.svrSession.Options().ErrorCh:
			that.errorCh <- err
		case err := <-that.playerSession.Options().ErrorCh:
			that.errorCh <- err
		case msg := <-that.svrSession.Options().Output:
			if logger.IsDebug() {
				logger.Debugf("收到vnc客户端发送过来的消息,%s", msg)
			}
		default:
			// 从会话中读取消息类型
			var messageType rfb.ServerMessageType
			if err := binary.Read(that.playerSession, binary.BigEndian, &messageType); err != nil {
				that.errorCh <- err
				return
			}
			msg := &messages.FramebufferUpdate{}
			// 读取消息内容
			parsedMsg, err := msg.Read(that.playerSession)
			if err != nil {
				that.errorCh <- err
				return
			}
			that.svrSession.Options().Input <- parsedMsg
			//err = parsedMsg.Write(that.svrSession)
			//if err != nil {
			//	that.errorCh <- err
			//	return
			//}
			var sleep int64
			_ = binary.Read(that.playerSession, binary.BigEndian, &sleep)
			if sleep > 0 {
				time.Sleep(time.Duration(sleep))
			}
		}
	}
}

func (that *Player) Close() {
	_ = that.svrSession.Close()
	_ = that.playerSession.Close()
}

func (that *Player) Error() <-chan error {
	return that.errorCh
}
