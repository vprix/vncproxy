package main

import (
	"fmt"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"golang.org/x/net/context"
	"os"
)

var (
	helpContent = gstr.TrimLeft(`
USAGE
	./recorder [start|stop|quit]  [OPTION]
OPTION
	--rbsFile       使用的rbs文件地址  必传
	--vncHost       要连接的vnc服务端地址  必传
	--vncPort       要连接的vnc服务端端口 必传
	--vncPassword   要连接的vnc服务端密码 不传则使用auth none
	--debug         是否开启debug 默认debug=false
	-d,--daemon     使用守护进程模式启动
	--pid           设置pid文件的地址，默认是/tmp/[server].pid
	-h,--help       获取帮助信息
	-v,--version    获取编译版本信息
	
EXAMPLES
	/path/to/recorder 
	/path/to/recorder start --env=dev --debug=true --pid=/tmp/server.pid
	/path/to/recorder start -c=config.product.toml
	/path/to/recorder start --config=config.product.toml
	/path/to/recorder start --rbsFile=/path/to/foo.rbs
							--vncHost=192.168.1.2 
							--vncPort=5901
							--vncPassword=vprix
							--debug
	/path/to/server stop
	/path/to/server quit
	/path/to/server reload
	/path/to/server version
	/path/to/server help
`)
)

func main() {
	easyservice.Authors = "ClownFish"
	easyservice.SetHelpContent(helpContent)
	easyservice.SetOptions(
		map[string]bool{
			"rbsFile":     true, // 使用的rbs文件地址  必传
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
		rbsFile := svr.CmdParser().GetOpt("rbsFile", "")
		if len(rbsFile.String()) <= 0 {
			svr.Help()
			os.Exit(0)
		}
		vncHost := svr.CmdParser().GetOpt("vncHost", "")
		if len(vncHost.String()) <= 0 {
			svr.Help()
			os.Exit(0)
		}
		vncPort := svr.CmdParser().GetOpt("vncPort", 0)
		if vncPort.Int() <= 0 {
			svr.Help()
			os.Exit(0)
		}
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("rbsFile", rbsFile.String())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("vncHost", vncHost.String())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("vncPort", vncPort.Int())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("vncPassword", svr.CmdParser().GetOpt("vncPassword", ""))

		logger.SetDebug(cfg.MustGet(context.TODO(), "Debug").Bool())

		svr.AddSandBox(NewRecorderSandBox(cfg))
	})

}
