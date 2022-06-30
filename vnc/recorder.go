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
	errorCh         chan error
	closed          chan struct{}
	cliSession      *session.ClientSession // 链接到vnc服务端的会话
	recorderSession *session.RecorderSession
}

func NewRecorder(recorderSess *session.RecorderSession, cliSession *session.ClientSession) *Recorder {
	recorder := &Recorder{
		recorderSession: recorderSess,
		cliSession:      cliSession,
		errorCh:         make(chan error, 32),
		closed:          make(chan struct{}),
	}
	return recorder
}

func (that *Recorder) Start() error {
	var err error
	that.cliSession.Start()
	encS := []rfb.EncodingType{
		rfb.EncCursorPseudo,
		rfb.EncPointerPosPseudo,
		rfb.EncCopyRect,
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
	that.recorderSession.SetWidth(that.cliSession.Options().Width)
	that.recorderSession.SetHeight(that.cliSession.Options().Height)
	that.recorderSession.SetPixelFormat(that.cliSession.Options().PixelFormat)
	that.recorderSession.SetDesktopName(that.cliSession.Options().DesktopName)
	that.recorderSession.Start()
	reqMsg := messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Options().Width, Height: that.cliSession.Options().Height}
	err = reqMsg.Write(that.cliSession)
	if err != nil {
		return err
	}
	var lastUpdate *gtime.Time
	for {
		select {
		case msg := <-that.recorderSession.Options().Output:
			logger.Debugf("client message received.messageType:%d,message:%s", msg.Type(), msg)
		case msg := <-that.cliSession.Options().Output:
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
				reqMsg = messages.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: that.cliSession.Options().Width, Height: that.cliSession.Options().Height}
				err = reqMsg.Write(that.cliSession)
				if err != nil {
					return err
				}
			}
		case <-that.cliSession.Wait():
			return nil
		case <-that.recorderSession.Wait():
			return nil
		case err = <-that.cliSession.Options().ErrorCh:
			that.errorCh <- err
			that.Close()
		case err = <-that.recorderSession.Options().ErrorCh:
			that.errorCh <- err
			that.Close()
		}
	}
}

func (that *Recorder) Wait() <-chan struct{} {
	return that.closed
}
func (that *Recorder) Close() {
	that.closed <- struct{}{}
}

func (that *Recorder) Error() <-chan error {
	return that.errorCh
}
