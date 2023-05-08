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
	./proxy [start|stop|quit] [tcpServer|wsServer] [OPTION]
OPTION
	--vncHost       要连接的vnc服务端地址  必传
	--vncPort       要连接的vnc服务端端口 必传
	--vncPassword   要连接的vnc服务端密码 不传则使用auth none
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
	/path/to/proxy 
	/path/to/proxy start --env=dev --debug=true --pid=/tmp/server.pid
	/path/to/proxy start -c=config.product.toml
	/path/to/proxy start tcpServer,wsServer --config=config.product.toml
	/path/to/proxy start wsServer  --vncHost=192.168.1.2 
                                    --vncPort=5901
                                    --vncPassword=vprix
                                    --wsHost=0.0.0.0
                                    --wsPort=8988
                                    --wsPath=/
                                    --proxyPassword=12345612
                                    --debug
	/path/to/proxy start tcpServer --vncHost=192.168.1.2 
                                    --vncPort=5901
                                    --vncPassword=vprix
                                    --tcpHost=0.0.0.0
                                    --tcpPort=8989
                                    --proxyPassword=12345612
                                    --debug
	/path/to/proxy stop
	/path/to/proxy quit
	/path/to/proxy reload
	/path/to/proxy version
	/path/to/proxy help
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
			"vncHost":       true, // 要连接的vnc服务端地址  必传
			"vncPort":       true, // 要连接的vnc服务端端口 必传
			"vncPassword":   true, // 要连接的vnc服务端密码 不传则使用auth none
		})

	easyservice.Setup(func(svr *easyservice.EasyService) {
		//注册服务停止时要执行法方法
		svr.BeforeStop(func(service *easyservice.EasyService) bool {
			fmt.Println("Vnc proxy server stop")
			return true
		})
		cfg := svr.Config()
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
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("vncHost", vncHost.String())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("vncPort", vncPort.Int())
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("vncPassword", svr.CmdParser().GetOpt("vncPassword", ""))

		logger.SetDebug(cfg.MustGet(context.TODO(), "Debug").Bool())

		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("tcpHost", svr.CmdParser().GetOpt("tcpHost", "0.0.0.0"))
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("tcpPort", svr.CmdParser().GetOpt("tcpPort", 8989))
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("proxyPassword", svr.CmdParser().GetOpt("proxyPassword", ""))
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("wsHost", svr.CmdParser().GetOpt("wsHost", "0.0.0.0"))
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("wsPort", svr.CmdParser().GetOpt("wsPort", 8988))
		_ = cfg.GetAdapter().(*gcfg.AdapterFile).Set("wsPath", svr.CmdParser().GetOpt("wsPath", "/"))

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
