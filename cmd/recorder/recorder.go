package main

import (
	"fmt"
	"github.com/gogf/gf/os/gcfg"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gtime"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
	"io"
	"net"
	"os"
	"time"
)

// RecorderSandBox  记录服务
type RecorderSandBox struct {
	id       int
	name     string
	cfg      *gcfg.Config
	service  *easyservice.EasyService
	recorder *vnc.Recorder
	closed   chan struct{}
}

// NewRecorderSandBox 创建一个默认的服务沙盒
func NewRecorderSandBox(cfg *gcfg.Config) *RecorderSandBox {
	id := easyservice.GetNextSandBoxId()
	sBox := &RecorderSandBox{
		id:     id,
		name:   fmt.Sprintf("tcp_%d", id),
		cfg:    cfg,
		closed: make(chan struct{}),
	}
	return sBox
}

func (that *RecorderSandBox) ID() int {
	return that.id
}

func (that *RecorderSandBox) Name() string {
	return that.name
}

func (that *RecorderSandBox) Setup() error {
	saveFilePath := that.cfg.GetString("rbsFile")
	targetCfg := rfb.TargetConfig{
		Network:  "tcp",
		Host:     that.cfg.GetString("vncHost"),
		Port:     that.cfg.GetInt("vncPort"),
		Password: that.cfg.GetBytes("vncPassword"),
		Timeout:  10 * time.Second,
	}
	var securityHandlers = []rfb.ISecurityHandler{
		&security.ClientAuthNone{},
	}
	if len(targetCfg.Password) > 0 {
		securityHandlers = []rfb.ISecurityHandler{
			&security.ClientAuthVNC{Password: targetCfg.Password},
		}
	}
	// 创建会话
	recorderSess := session.NewRecorder(
		rfb.OptEncodings(encodings.DefaultEncodings...),
		rfb.OptMessages(messages.DefaultServerMessages...),
		rfb.OptPixelFormat(rfb.PixelFormat32bit),
		rfb.OptGetConn(func() (io.ReadWriteCloser, error) {
			if gfile.Exists(saveFilePath) {
				saveFilePath = fmt.Sprintf("%s%s%s_%d%s",
					gfile.Dir(saveFilePath),
					gfile.Separator,
					gfile.Name(gfile.Basename(saveFilePath)),
					gtime.Now().Unix(),
					gfile.Ext(gfile.Basename(saveFilePath)),
				)
			}
			return gfile.OpenFile(saveFilePath, os.O_RDWR|os.O_CREATE, 0644)
		}),
	)
	cliSession := session.NewClient(
		rfb.OptEncodings(encodings.DefaultEncodings...),
		rfb.OptMessages(messages.DefaultServerMessages...),
		rfb.OptPixelFormat(rfb.PixelFormat32bit),
		rfb.OptGetConn(func() (io.ReadWriteCloser, error) {
			return net.DialTimeout(targetCfg.Network, targetCfg.Addr(), targetCfg.Timeout)
		}),
		rfb.OptSecurityHandlers(securityHandlers...),
	)
	that.recorder = vnc.NewRecorder(recorderSess, cliSession)
	go func() {
		err := that.recorder.Start()
		if err != nil {
			logger.Fatal(err)
		}
	}()
	return nil
}

func (that *RecorderSandBox) Shutdown() error {
	close(that.closed)
	that.recorder.Close()
	return nil
}

func (that *RecorderSandBox) Error() <-chan error {
	return that.recorder.Error()
}

func (that *RecorderSandBox) Service() *easyservice.EasyService {
	return that.service
}
