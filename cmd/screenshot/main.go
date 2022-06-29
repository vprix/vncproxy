package main

import (
	"bytes"
	"fmt"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/vnc"
	"image/draw"
	"image/jpeg"
	"os"
	"time"
)

var (
	helpContent = gstr.TrimLeft(`
USAGE
	./server [start|stop|quit] [OPTION]
OPTION
	--imageFile     要生成的截图地址,暂时只支持jpeg格式  必传
	--vncHost       要连接的vnc服务端地址  必传
	--vncPort       要连接的vnc服务端端口 必传
	--vncPassword   要连接的vnc服务端密码 不传则使用auth none
	--debug         是否开启debug 默认debug=false
	-h,--help       获取帮助信息
	-v,--version    获取编译版本信息
	
EXAMPLES
	/path/to/server 
	/path/to/server start --env=dev --debug=true
	/path/to/server start -c=config.product.toml
	/path/to/server start  --config=config.product.toml
	/path/to/server start  --imageFile=/path/to/foo.jpeg
                                    --vncHost=192.168.1.2 
                                    --vncPort=5901
                                    --vncPassword=vprix
                                    --debug
	/path/to/server version
	/path/to/server help
`)
)

func main() {
	easyservice.Authors = "ClownFish"
	easyservice.SetHelpContent(helpContent)
	easyservice.SetOptions(
		map[string]bool{
			"imageFile":   true, // 要生成的截图地址,暂时只支持jpeg格式  必传
			"vncHost":     true, // 要连接的vnc服务端地址  必传
			"vncPort":     true, // 要连接的vnc服务端端口 必传
			"vncPassword": true, // 要连接的vnc服务端密码 不传则使用auth none
		})

	easyservice.Setup(func(svr *easyservice.EasyService) {
		//注册服务停止时要执行法方法
		svr.BeforeStop(func(service *easyservice.EasyService) bool {
			fmt.Println("Vnc player server stop")
			return true
		})
		cfg := svr.Config()
		rbsFile := svr.CmdParser().GetOptVar("imageFile", "")
		if len(rbsFile.String()) <= 0 {
			svr.Help()
			os.Exit(0)
		}
		vncHost := svr.CmdParser().GetOptVar("vncHost", "")
		if len(vncHost.String()) <= 0 {
			svr.Help()
			os.Exit(0)
		}
		vncPort := svr.CmdParser().GetOptVar("vncPort", 0)
		if vncPort.Int() <= 0 {
			svr.Help()
			os.Exit(0)
		}
		_ = cfg.Set("rbsFile", rbsFile.String())
		_ = cfg.Set("vncHost", vncHost.String())
		_ = cfg.Set("vncPort", vncPort.Int())
		_ = cfg.Set("vncPassword", svr.CmdParser().GetOptVar("vncPassword", ""))

		logger.SetDebug(cfg.GetBool("Debug"))

		v := vnc.NewScreenshot(
			rfb.TargetConfig{
				Network:  "tcp",
				Host:     vncHost.String(),
				Port:     vncPort.Int(),
				Password: svr.CmdParser().GetOptVar("vncPassword", "").Bytes(),
				Timeout:  5 * time.Second,
			},
		)
		img, err := v.GetImage()
		if err != nil {
			logger.Fatal(err)
		}

		j := &bytes.Buffer{}
		err = jpeg.Encode(j, img.(draw.Image), &jpeg.Options{Quality: 100})
		if err != nil {
			fmt.Println(err)
		}
		err = gfile.PutBytes(rbsFile.String(), j.Bytes())
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	})
}
