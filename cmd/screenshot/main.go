package main

import (
	"bytes"
	"github.com/gogf/gf/os/gfile"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/vnc"
	"image/jpeg"
	"time"
)

func main() {

	v := vnc.NewScreenshot(
		rfb.TargetConfig{
			Network:  "tcp",
			Host:     "127.0.0.1",
			Port:     5901,
			Password: []byte("@abc1234"),
			Timeout:  10 * time.Second,
		},
	)
	img, err := v.Start()
	if err != nil {
		logger.Fatal(err)
	}
	j := &bytes.Buffer{}
	err = jpeg.Encode(j, img, &jpeg.Options{Quality: 100})
	if err != nil {
		logger.Fatal(err)
	}
	gfile.PutBytes("D:\\code\\GolandProjects\\vprix-vnc\\screenshot.jpeg", j.Bytes())
}
