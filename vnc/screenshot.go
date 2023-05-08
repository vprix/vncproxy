package vnc

import (
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"golang.org/x/net/context"
	"io"
	"net"
	"time"
)

type Screenshot struct {
	cliSession    *session.ClientSession // 链接到vnc服务端的会话
	canvasSession *session.CanvasSession
	timeout       time.Duration
}

func NewScreenshot(targetCfg rfb.TargetConfig) *Screenshot {
	securityHandlers := []rfb.ISecurityHandler{
		&security.ClientAuthNone{},
	}
	if len(targetCfg.Password) > 0 {
		securityHandlers = []rfb.ISecurityHandler{
			&security.ClientAuthVNC{Password: targetCfg.Password},
		}
	}
	canvasSession := session.NewCanvasSession()
	cliSession := session.NewClient(
		rfb.OptSecurityHandlers(securityHandlers...),
		rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
			return net.DialTimeout(targetCfg.GetNetwork(), targetCfg.Addr(), targetCfg.GetTimeout())
		}),
	)
	recorder := &Screenshot{
		canvasSession: canvasSession,
		cliSession:    cliSession,
		timeout:       targetCfg.GetTimeout(),
	}
	return recorder
}

func (that *Screenshot) GetImage() (io.ReadWriteCloser, error) {
	var err error
	that.cliSession.Start()
	encS := []rfb.EncodingType{
		rfb.EncCursorPseudo,
		rfb.EncPointerPosPseudo,
		rfb.EncHexTile,
		rfb.EncTight,
		rfb.EncZRLE,
	}
	defer func() {
		_ = that.cliSession.Close()
	}()
	err = that.cliSession.SetEncodings(encS)
	if err != nil {
		return nil, err
	}
	// 设置参数信息
	that.canvasSession.SetProtocolVersion(that.cliSession.ProtocolVersion())
	that.canvasSession.SetWidth(that.cliSession.Options().Width)
	that.canvasSession.SetHeight(that.cliSession.Options().Height)
	that.canvasSession.SetPixelFormat(that.cliSession.Options().PixelFormat)
	that.canvasSession.SetDesktopName(that.cliSession.Options().DesktopName)
	that.canvasSession.Start()
	defer func() {
		_ = that.canvasSession.Close()
	}()
	reqMsg := messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Options().Width, Height: that.cliSession.Options().Height}
	err = reqMsg.Write(that.cliSession)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), that.timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("获取截图超时")
		case msg := <-that.cliSession.Options().Output:
			if rfb.ServerMessageType(msg.Type()) == rfb.FramebufferUpdate {
				err = msg.Write(that.canvasSession)
				if err != nil {
					return nil, err
				}
				err = that.canvasSession.Flush()
				return that.canvasSession.Conn(), err
			}
			if logger.IsDebug() {
				logger.Debugf(context.TODO(), "获取到来自vnc服务端的消息%v", msg)
			}
		}
	}
}
