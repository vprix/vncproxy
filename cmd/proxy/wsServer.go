package main

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcfg"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"time"
)

// WSSandBox  Tcp的服务
type WSSandBox struct {
	id      int
	name    string
	cfg     *gcfg.Config
	service *easyservice.EasyService
	svr     *ghttp.Server
}

// NewWSSandBox 创建一个默认的服务沙盒
func NewWSSandBox(cfg *gcfg.Config) *WSSandBox {
	id := easyservice.GetNextSandBoxId()
	sBox := &WSSandBox{
		id:   id,
		name: fmt.Sprintf("ws_%d", id),
		cfg:  cfg,
	}
	return sBox
}

func (that *WSSandBox) ID() int {
	return that.id
}

func (that *WSSandBox) Name() string {
	return that.name
}

func (that *WSSandBox) Setup() error {

	that.svr = g.Server()
	that.svr.BindHandler(that.cfg.GetString("wsPath", "/"), func(r *ghttp.Request) {
		h := websocket.Handler(func(conn *websocket.Conn) {
			conn.PayloadType = websocket.BinaryFrame

			securityHandlers := []rfb.ISecurityHandler{
				&security.ServerAuthNone{},
			}
			if len(that.cfg.GetBytes("proxyPassword")) > 0 {
				securityHandlers = append(securityHandlers, &security.ServerAuthVNC{Password: that.cfg.GetBytes("proxyPassword")})
			}
			svrSess := session.NewServerSession(
				rfb.OptDesktopName([]byte("Vprix VNC Proxy")),
				rfb.OptHeight(768),
				rfb.OptWidth(1024),
				rfb.OptSecurityHandlers(securityHandlers...),
				rfb.OptGetConn(func() (io.ReadWriteCloser, error) {
					return conn, nil
				}),
			)
			targetCfg := rfb.TargetConfig{
				Host:     that.cfg.GetString("vncHost"),
				Port:     that.cfg.GetInt("vncPort"),
				Password: that.cfg.GetBytes("vncPassword"),
			}
			timeout := 10 * time.Second
			network := "tcp"
			cliSess := session.NewClient(
				rfb.OptSecurityHandlers([]rfb.ISecurityHandler{&security.ClientAuthVNC{Password: targetCfg.Password}}...),
				rfb.OptGetConn(func() (io.ReadWriteCloser, error) {
					return net.DialTimeout(network, targetCfg.Addr(), timeout)
				}),
			)
			p := vnc.NewVncProxy(cliSess, svrSess)
			p.Start()
			for {
				err := <-p.Error()
				logger.Warning(err)
			}
		})
		h.ServeHTTP(r.Response.Writer, r.Request)
	})
	that.svr.SetAddr(fmt.Sprintf("%s:%d", that.cfg.GetString("wsHost"), that.cfg.GetInt("wsPort")))
	return that.svr.Start()
}

func (that *WSSandBox) Shutdown() error {
	return that.svr.Shutdown()
}

func (that *WSSandBox) Service() *easyservice.EasyService {
	return that.service
}
