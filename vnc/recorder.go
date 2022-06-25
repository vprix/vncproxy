package vnc

import (
	"encoding/binary"
	"github.com/gogf/gf/os/gtime"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"net"
	"time"
)

type Recorder struct {
	closed          chan struct{}
	cliCfg          *rfb.Option
	targetCfg       rfb.TargetConfig
	cliSession      *session.ClientSession // 链接到vnc服务端的会话
	recorderSession *session.RecorderSession
}

func NewRecorder(saveFilePath string, cliCfg *rfb.Option, targetCfg rfb.TargetConfig) *Recorder {
	if cliCfg == nil {
		cliCfg = &rfb.Option{
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
	recorder := &Recorder{
		recorderSession: session.NewRecorder(saveFilePath, cliCfg),
		targetCfg:       targetCfg,
		cliCfg:          cliCfg,
	}
	return recorder
}

func (that *Recorder) Start() error {

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
	that.recorderSession.SetProtocolVersion(that.cliSession.ProtocolVersion())
	that.recorderSession.Desktop().SetWidth(that.cliSession.Desktop().Width())
	that.recorderSession.Desktop().SetHeight(that.cliSession.Desktop().Height())
	that.recorderSession.Desktop().SetPixelFormat(that.cliSession.Desktop().PixelFormat())
	that.recorderSession.Desktop().SetDesktopName(that.cliSession.Desktop().DesktopName())
	that.recorderSession.Run()
	reqMsg := messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Desktop().Width(), Height: that.cliSession.Desktop().Height()}
	err = reqMsg.Write(that.cliSession)
	if err != nil {
		return err
	}
	var lastUpdate *gtime.Time
	for {
		select {
		case msg := <-that.cliCfg.Output:
			logger.Debugf("client message received.messageType:%d,message:%s", msg.Type(), msg)
		case msg := <-that.cliCfg.Input:
			if rfb.ServerMessageType(msg.Type()) == rfb.FramebufferUpdate {
				err = msg.Write(that.recorderSession)
				if err != nil {
					return err
				}
				if lastUpdate == nil {
					_ = binary.Write(that.recorderSession, binary.BigEndian, int64(0))
				} else {
					secsPassed := gtime.Now().UnixNano() - lastUpdate.UnixNano()
					_ = binary.Write(that.recorderSession, binary.BigEndian, secsPassed)
				}
				err = that.recorderSession.Flush()
				if err != nil {
					return err
				}
				lastUpdate = gtime.Now()
				reqMsg = messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Desktop().Width(), Height: that.cliSession.Desktop().Height()}
				err = reqMsg.Write(that.cliSession)
				if err != nil {
					return err
				}
			}
		case <-that.closed:
			return nil
		}
	}
}

func (that *Recorder) Close() {
	if that.closed != nil {
		close(that.closed)
		that.closed = nil
		_ = that.cliSession.Close()
		_ = that.recorderSession.Close()
	}

}

func (that *Recorder) Error() <-chan error {
	return that.cliCfg.ErrorCh
}
