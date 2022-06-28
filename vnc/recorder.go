package vnc

import (
	"encoding/binary"
	"github.com/gogf/gf/os/gtime"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/session"
)

type Recorder struct {
	closed          chan struct{}
	cliSession      *session.ClientSession // 链接到vnc服务端的会话
	recorderSession *session.RecorderSession
}

func NewRecorder(recorderSess *session.RecorderSession, cliSession *session.ClientSession) *Recorder {
	recorder := &Recorder{
		recorderSession: recorderSess,
		cliSession:      cliSession,
	}
	return recorder
}

func (that *Recorder) Start() error {
	var err error
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
		case msg := <-that.cliSession.Options().Output:
			logger.Debugf("client message received.messageType:%d,message:%s", msg.Type(), msg)
		case msg := <-that.cliSession.Options().Input:
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
	return that.cliSession.Options().ErrorCh
}
