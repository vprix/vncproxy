package main

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
	"golang.org/x/net/websocket"
	"io"
	"net"
)

// WSSandBox  Tcp的服务
type WSSandBox struct {
	id       int
	name     string
	cfg      *gcfg.Config
	service  *easyservice.EasyService
	svr      *ghttp.Server
	proxyHub *gmap.StrAnyMap
}

// NewWSSandBox 创建一个默认的服务沙盒
func NewWSSandBox(cfg *gcfg.Config) *WSSandBox {
	id := easyservice.GetNextSandBoxId()
	sBox := &WSSandBox{
		id:       id,
		name:     fmt.Sprintf("ws_%d", id),
		cfg:      cfg,
		proxyHub: gmap.NewStrAnyMap(true),
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
	that.svr.BindHandler(that.cfg.MustGet(context.TODO(), "wsPath", "/").String(), func(r *ghttp.Request) {
		h := websocket.Handler(func(conn *websocket.Conn) {
			conn.PayloadType = websocket.BinaryFrame
			securityHandlers := []rfb.ISecurityHandler{
				&security.ServerAuthNone{},
			}
			if len(that.cfg.MustGet(context.TODO(), "proxyPassword").Bytes()) > 0 {
				securityHandlers = append(securityHandlers, &security.ServerAuthVNC{Password: that.cfg.MustGet(context.TODO(), "proxyPassword").Bytes()})
			}
			targetCfg := rfb.TargetConfig{
				Host:     that.cfg.MustGet(context.TODO(), "vncHost").String(),
				Port:     that.cfg.MustGet(context.TODO(), "vncPort").Int(),
				Password: that.cfg.MustGet(context.TODO(), "vncPassword").Bytes(),
			}
			var err error
			svrSess := session.NewServerSession(
				rfb.OptDesktopName([]byte("Vprix VNC Proxy")),
				rfb.OptHeight(768),
				rfb.OptWidth(1024),
				rfb.OptSecurityHandlers(securityHandlers...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return conn, nil
				}),
			)
			cliSess := session.NewClient(
				rfb.OptSecurityHandlers([]rfb.ISecurityHandler{&security.ClientAuthVNC{Password: targetCfg.Password}}...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return net.DialTimeout(targetCfg.GetNetwork(), targetCfg.Addr(), targetCfg.GetTimeout())
				}),
			)
			p := vnc.NewVncProxy(cliSess, svrSess)
			remoteKey := conn.RemoteAddr().String()
			that.proxyHub.Set(remoteKey, p)
			err = p.Start()
			if err != nil {
				glog.Warning(context.TODO(), err)
				return
			}
			glog.Info(context.TODO(), "proxy session end")
		})
		h.ServeHTTP(r.Response.Writer, r.Request)
	})
	that.svr.SetAddr(fmt.Sprintf("%s:%d", that.cfg.MustGet(context.TODO(), "wsHost").String(), that.cfg.MustGet(context.TODO(), "wsPort").Int()))
	return that.svr.Start()
}

func (that *WSSandBox) Shutdown() error {
	return that.svr.Shutdown()
}

func (that *WSSandBox) Service() *easyservice.EasyService {
	return that.service
}
