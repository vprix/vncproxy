package main

import (
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/recorder"
	"github.com/vprix/vncproxy/rfb"
	"time"
)

func main() {

	reco := recorder.NewRecorder("D:\\code\\GolandProjects\\vprix-vnc\\abc.rbs",
		nil,
		rfb.TargetConfig{
			Network:  "tcp",
			Host:     "127.0.0.1",
			Port:     5901,
			Password: []byte("@abc1234"),
			Timeout:  10 * time.Second,
		},
	)
	go func() {
		err := reco.Start()
		if err != nil {
			logger.Fatal(err)
		}
	}()
	for {
		err := <-reco.Error()
		logger.Error(err)
	}

}
