package main

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"io"
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
	that.svr.BindHandler(that.cfg.MustGet(context.TODO(), "wsPath", "/").String(), func(r *ghttp.Request) {
		h := websocket.Handler(func(conn *websocket.Conn) {
			conn.PayloadType = websocket.BinaryFrame

			securityHandlers := []rfb.ISecurityHandler{&security.ServerAuthNone{}}
			if len(that.cfg.MustGet(context.TODO(), "proxyPassword").Bytes()) > 0 {
				securityHandlers = []rfb.ISecurityHandler{&security.ServerAuthVNC{Password: that.cfg.MustGet(context.TODO(), "proxyPassword").Bytes()}}
			}
			svrSession := session.NewServerSession(
				rfb.OptSecurityHandlers(securityHandlers...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return conn, nil
				}),
			)
			play := vnc.NewPlayer(that.cfg.MustGet(context.TODO(), "rfbFile").String(), svrSession)
			err := play.Start()
			if err != nil {
				glog.Warning(context.TODO(), err)
				return
			}
			glog.Info(context.TODO(), "play session end")
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
