package vnc

import (
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"io"
	"net"
	"time"
)

type Screenshot struct {
	cliCfg        *rfb.Option
	targetCfg     rfb.TargetConfig
	cliSession    *session.ClientSession // 链接到vnc服务端的会话
	canvasSession *session.CanvasSession
}

func NewScreenshot(targetCfg rfb.TargetConfig) *Screenshot {
	cliCfg := &rfb.Option{
		PixelFormat: rfb.PixelFormat32bit,
		Messages:    messages.DefaultServerMessages,
		Encodings:   encodings.DefaultEncodings,
		Output:      make(chan rfb.Message),
		Input:       make(chan rfb.Message),
		ErrorCh:     make(chan error),
	}
	if len(targetCfg.Password) > 0 {
		cliCfg.SecurityHandlers = []rfb.ISecurityHandler{
			&security.ClientAuthVNC{Password: targetCfg.Password},
		}
	} else {
		cliCfg.SecurityHandlers = []rfb.ISecurityHandler{
			&security.ClientAuthNone{},
		}
	}
	recorder := &Screenshot{
		canvasSession: session.NewCanvasSession(cliCfg),
		targetCfg:     targetCfg,
		cliCfg:        cliCfg,
	}
	return recorder
}

func (that *Screenshot) Start() (io.ReadWriteCloser, error) {

	timeout := 10 * time.Second
	if that.targetCfg.Timeout > 0 {
		timeout = that.targetCfg.Timeout
	}
	network := "tcp"
	if len(that.targetCfg.Network) > 0 {
		network = that.targetCfg.Network
	}
	clientConn, err := net.DialTimeout(network, that.targetCfg.Addr(), timeout)
	if err != nil {
		return nil, err
	}
	that.cliSession, err = session.NewClient(clientConn, that.cliCfg)
	if err != nil {
		return nil, err
	}

	that.cliSession.Run()
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
	that.canvasSession.Desktop().SetWidth(that.cliSession.Desktop().Width())
	that.canvasSession.Desktop().SetHeight(that.cliSession.Desktop().Height())
	that.canvasSession.Desktop().SetPixelFormat(that.cliSession.Desktop().PixelFormat())
	that.canvasSession.Desktop().SetDesktopName(that.cliSession.Desktop().DesktopName())
	that.canvasSession.Run()
	defer func() {
		_ = that.canvasSession.Close()
	}()
	reqMsg := messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Desktop().Width(), Height: that.cliSession.Desktop().Height()}
	err = reqMsg.Write(that.cliSession)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case msg := <-that.cliCfg.Input:
			if rfb.ServerMessageType(msg.Type()) == rfb.FramebufferUpdate {
				err = msg.Write(that.canvasSession)
				if err != nil {
					return nil, err
				}
				err = that.canvasSession.Flush()
				return that.canvasSession.Conn(), err
			}
		}
	}
}
