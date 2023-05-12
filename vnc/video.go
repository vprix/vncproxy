package vnc

import (
	"context"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"io"
	"net"
	"time"
)

type Video struct {
	cliCfg        *rfb.Options
	targetCfg     rfb.TargetConfig
	cliSession    *session.ClientSession // 链接到vnc服务端的会话
	canvasSession *session.CanvasSession
}

func NewVideo(cliCfg *rfb.Options, targetCfg rfb.TargetConfig) *Video {
	if cliCfg == nil {
		cliCfg = &rfb.Options{
			PixelFormat: rfb.PixelFormat32bit,
			Messages:    messages.DefaultServerMessages,
			Encodings:   encodings.DefaultEncodings,
			Output:      make(chan rfb.Message),
			Input:       make(chan rfb.Message),
			ErrorCh:     make(chan error),
		}
	}
	if cliCfg.Output == nil {
		cliCfg.Output = make(chan rfb.Message)
	}
	if cliCfg.Input == nil {
		cliCfg.Input = make(chan rfb.Message)
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
		//canvasSession: session.NewCanvasSession(*cliCfg),
		targetCfg: targetCfg,
		cliCfg:    cliCfg,
	}
	return recorder
}

func (that *Video) Start() error {
	var err error
	timeout := 10 * time.Second
	if that.targetCfg.Timeout > 0 {
		timeout = that.targetCfg.Timeout
	}
	network := "tcp"
	if len(that.targetCfg.Network) > 0 {
		network = that.targetCfg.Network
	}
	that.cliCfg.GetConn = func(sess rfb.ISession) (io.ReadWriteCloser, error) {
		return net.DialTimeout(network, that.targetCfg.Addr(), timeout)
	}
	that.cliSession = session.NewClient()

	that.cliSession.Start()
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
	that.canvasSession.SetWidth(that.cliSession.Options().Width)
	that.canvasSession.SetHeight(that.cliSession.Options().Height)
	that.canvasSession.SetPixelFormat(that.cliSession.Options().PixelFormat)
	that.canvasSession.SetDesktopName(that.cliSession.Options().DesktopName)
	that.canvasSession.Start()
	reqMsg := messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Options().Width, Height: that.cliSession.Options().Height}
	err = reqMsg.Write(that.cliSession)
	if err != nil {
		return err
	}
	for {
		select {
		case msg := <-that.cliCfg.Output:
			logger.Debugf(context.TODO(), "client message received.messageType:%d,message:%s", msg.Type(), msg)
		case msg := <-that.cliCfg.Input:
			if rfb.ServerMessageType(msg.Type()) == rfb.FramebufferUpdate {
				err = msg.Write(that.canvasSession)
				if err != nil {
					return err
				}
				err = that.canvasSession.Flush()
				if err != nil {
					return err
				}
				reqMsg = messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Options().Width, Height: that.cliSession.Options().Height}
				err = reqMsg.Write(that.cliSession)
				if err != nil {
					return err
				}
			}
		case err = <-that.cliCfg.ErrorCh:
			return err
		}
	}
}

func (that *Video) Close() {
	_ = that.cliSession.Close()
	_ = that.canvasSession.Close()

}
