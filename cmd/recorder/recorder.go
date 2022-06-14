package main

import (
	"fmt"
	"github.com/gogf/gf/os/gcfg"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/vnc"
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
	that.recorder = vnc.NewRecorder(that.cfg.GetString("rbsFile"),
		nil,
		rfb.TargetConfig{
			Network:  "tcp",
			Host:     that.cfg.GetString("vncHost"),
			Port:     that.cfg.GetInt("vncPort"),
			Password: that.cfg.GetBytes("vncPassword"),
			Timeout:  10 * time.Second,
		},
	)
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
