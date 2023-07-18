package main

import (
	"fmt"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"golang.org/x/net/context"
	"os"
)

var (
	helpContent = gstr.TrimLeft(`
USAGE
	./server [start|stop|quit] [tcpServer|wsServer] [OPTION]
OPTION
	--rbsFile       使用的rbs文件地址  必传
	--tcpHost       本地监听的tcp协议地址 默认0.0.0.0
	--tcpPort       本地监听的tcp协议端口 默认8989
	--proxyPassword 连接到proxy的密码   不传入密码则使用auth none
	--wsHost        启动websocket服务的本地地址  默认 0.0.0.0
	--wsPort        启动websocket服务的本地端口 默认8988
	--wsPath        启动websocket服务的url path 默认'/'
	--debug         是否开启debug 默认debug=false
	-d,--daemon     使用守护进程模式启动
	--pid           设置pid文件的地址，默认是/tmp/[server].pid
	-h,--help       获取帮助信息
	-v,--version    获取编译版本信息
	
EXAMPLES
	/path/to/server 
	/path/to/server start --env=dev --debug=true --pid=/tmp/server.pid
	/path/to/server start -c=config.product.toml
	/path/to/server start tcpServer,wsServer --config=config.product.toml
	/path/to/server start wsServer  --rbsFile=/path/to/foo.rbs
                                    --wsHost=0.0.0.0
                                    --wsPort=8988
                                    --proxyPassword=12345612
                                    --debug
	/path/to/server start tcpServer --rbsFile=/path/to/foo.rbs
                                    --tcpHost=0.0.0.0
                                    --tcpPort=8989
                                    --proxyPassword=12345612
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
			"tcpHost":       true, //本地监听的tcp协议地址 默认0.0.0.0
			"tcpPort":       true, //本地监听的tcp协议端口 默认8989
			"proxyPassword": true, //连接到proxy的密码   不传入密码则使用auth none
			"wsHost":        true, //启动websocket服务的本地地址  默认 0.0.0.0
			"wsPort":        true, //启动websocket服务的本地端口 默认8988
			"wsPath":        true, //启动websocket服务的url path 默认'/'
			"rbsFile":       true, // 使用的rbs文件地址  必传
		})

	easyservice.Setup(func(svr *easyservice.EasyService) {
		//注册服务停止时要执行法方法
		svr.BeforeStop(func(service *easyservice.EasyService) bool {
			fmt.Println("Vnc player server stop")
			return true
		})
		cfg := svr.Config()
		rbsFile := svr.CmdParser().GetOpt("rbsFile", "")
		if len(rbsFile.String()) <= 0 || !gfile.Exists(rbsFile.String()) {
			svr.Help()
			os.Exit(0)
		}
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("rbsFile", rbsFile.String())

		logger.SetDebug(cfg.MustGet(context.TODO(), "Debug").Bool())

		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("tcpHost", svr.CmdParser().GetOpt("tcpHost", "0.0.0.0").String())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("tcpPort", svr.CmdParser().GetOpt("tcpPort", 8989).Int())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("proxyPassword", svr.CmdParser().GetOpt("proxyPassword", "").String())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("wsHost", svr.CmdParser().GetOpt("wsHost", "0.0.0.0").String())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("wsPort", svr.CmdParser().GetOpt("wsPort", 8988).Int())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("wsPath", svr.CmdParser().GetOpt("wsPath", "/").String())

		if svr.SandboxNames().ContainsI("tcpserver") {
			svr.AddSandBox(NewTcpSandBox(cfg))
			return
		}
		if svr.SandboxNames().ContainsI("wsserver") {
			svr.AddSandBox(NewWSSandBox(cfg))
			return
		}
		svr.AddSandBox(NewTcpSandBox(cfg))
		svr.AddSandBox(NewWSSandBox(cfg))
	})
}
