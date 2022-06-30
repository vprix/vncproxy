package main

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcfg"
	"github.com/gogf/gf/os/glog"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
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
	that.svr.BindHandler(that.cfg.GetString("wsPath", "/"), func(r *ghttp.Request) {
		h := websocket.Handler(func(conn *websocket.Conn) {
			conn.PayloadType = websocket.BinaryFrame

			securityHandlers := []rfb.ISecurityHandler{&security.ServerAuthNone{}}
			if len(that.cfg.GetBytes("proxyPassword")) > 0 {
				securityHandlers = []rfb.ISecurityHandler{&security.ServerAuthVNC{Password: that.cfg.GetBytes("proxyPassword")}}
			}
			svrSession := session.NewServerSession(
				rfb.OptSecurityHandlers(securityHandlers...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return conn, nil
				}),
			)
			play := vnc.NewPlayer(that.cfg.GetString("rfbFile"), svrSession)
			err := play.Start()
			if err != nil {
				glog.Warning(err)
				return
			}
			for {
				select {
				case err = <-play.Error():
					glog.Warning(err)
					return
				case <-play.Wait():
					return
				}
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
