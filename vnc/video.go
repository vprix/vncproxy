package vnc

import (
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"net"
	"time"
)

type Video struct {
	cliCfg        *rfb.ClientConfig
	targetCfg     rfb.TargetConfig
	cliSession    *session.ClientSession // 链接到vnc服务端的会话
	canvasSession *session.CanvasSession
}

func NewVideo(cliCfg *rfb.ClientConfig, targetCfg rfb.TargetConfig) *Video {
	if cliCfg == nil {
		cliCfg = &rfb.ClientConfig{
			PixelFormat: rfb.PixelFormat32bit,
			Messages:    messages.DefaultServerMessages,
			Encodings:   encodings.DefaultEncodings,
			Output:      make(chan rfb.ClientMessage),
			Input:       make(chan rfb.ServerMessage),
			ErrorCh:     make(chan error),
		}
	}
	if cliCfg.Output == nil {
		cliCfg.Output = make(chan rfb.ClientMessage)
	}
	if cliCfg.Input == nil {
		cliCfg.Input = make(chan rfb.ServerMessage)
	}
	if cliCfg.ErrorCh == nil {
		cliCfg.ErrorCh = make(chan error)
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
	recorder := &Video{
		canvasSession: session.NewCanvasSession(cliCfg),
		targetCfg:     targetCfg,
		cliCfg:        cliCfg,
	}
	return recorder
}

func (that *Video) Start() error {

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
		return err
	}
	that.cliSession, err = session.NewClient(clientConn, that.cliCfg)
	if err != nil {
		return err
	}

	that.cliSession.Run()
	encS := []rfb.EncodingType{
		rfb.EncCursorPseudo,
		rfb.EncPointerPosPseudo,
		rfb.EncCopyRect,
		rfb.EncTight,
		rfb.EncZRLE,
		rfb.EncHexTile,
		rfb.EncZlib,
		rfb.EncRRE,
	}
	err = that.cliSession.SetEncodings(encS)
	if err != nil {
		return err
	}
	// 设置参数信息
	that.canvasSession.SetProtocolVersion(that.cliSession.ProtocolVersion())
	that.canvasSession.Desktop().SetWidth(that.cliSession.Desktop().Width())
	that.canvasSession.Desktop().SetHeight(that.cliSession.Desktop().Height())
	that.canvasSession.Desktop().SetPixelFormat(that.cliSession.Desktop().PixelFormat())
	that.canvasSession.Desktop().SetDesktopName(that.cliSession.Desktop().DesktopName())
	that.canvasSession.Run()
	reqMsg := messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Desktop().Width(), Height: that.cliSession.Desktop().Height()}
	err = reqMsg.Write(that.cliSession)
	if err != nil {
		return err
	}
	for {
		select {
		case msg := <-that.cliCfg.Output:
			logger.Debugf("client message received.messageType:%d,message:%s", msg.Type(), msg)
		case msg := <-that.cliCfg.Input:
			if msg.Type() == rfb.FramebufferUpdate {
				err = msg.Write(that.canvasSession)
				if err != nil {
					return err
				}
				err = that.canvasSession.Flush()
				if err != nil {
					return err
				}
				reqMsg = messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Desktop().Width(), Height: that.cliSession.Desktop().Height()}
				err = reqMsg.Write(that.cliSession)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (that *Video) Close() {
	_ = that.cliSession.Close()
	_ = that.canvasSession.Close()

}

func (that *Video) Error() <-chan error {
	return that.cliCfg.ErrorCh
}
