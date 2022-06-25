package main

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcfg"
	"github.com/gogf/gf/os/glog"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"golang.org/x/net/websocket"
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
			svrCfg := &rfb.Option{
				Encodings:   encodings.DefaultEncodings,
				DesktopName: []byte("Vprix VNC Proxy"),
				Width:       1024,
				Height:      768,
				SecurityHandlers: []rfb.ISecurityHandler{
					&security.ServerAuthNone{},
				},
				//DisableMessageType: []rfb.ServerMessageType{rfb.ServerCutText},
			}
			if len(that.cfg.GetBytes("proxyPassword")) > 0 {
				svrCfg.SecurityHandlers = append(svrCfg.SecurityHandlers, &security.ServerAuthVNC{Password: that.cfg.GetBytes("proxyPassword")})
			}
			p := attachNewServerConn(conn, svrCfg, nil, that.cfg.GetString("rbsFile"))
			for {
				err := <-p.Error()
				glog.Error(err)
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
