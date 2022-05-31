package main

import (
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/video"
	"time"
)

func main() {

	v := video.NewVideo(nil,
		rfb.TargetConfig{
			Network:  "tcp",
			Host:     "127.0.0.1",
			Port:     5901,
			Password: []byte("@abc1234"),
			Timeout:  10 * time.Second,
		},
	)
	go func() {
		err := v.Start()
		if err != nil {
			logger.Fatal(err)
		}
	}()
	for {
		err := <-v.Error()
		logger.Error(err)
	}

}
