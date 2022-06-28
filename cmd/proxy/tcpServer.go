package main

import (
	"fmt"
	"github.com/gogf/gf/os/gcfg"
	"github.com/gogf/gf/os/glog"
	"github.com/osgochina/dmicro/drpc"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"net"
)

// TcpSandBox  Tcp的服务
type TcpSandBox struct {
	id      int
	name    string
	cfg     *gcfg.Config
	service *easyservice.EasyService
	lis     net.Listener
	closed  chan struct{}
}

// NewTcpSandBox 创建一个默认的服务沙盒
func NewTcpSandBox(cfg *gcfg.Config) *TcpSandBox {
	id := easyservice.GetNextSandBoxId()
	sBox := &TcpSandBox{
		id:     id,
		name:   fmt.Sprintf("tcp_%d", id),
		cfg:    cfg,
		closed: make(chan struct{}),
	}
	return sBox
}

func (that *TcpSandBox) ID() int {
	return that.id
}

func (that *TcpSandBox) Name() string {
	return that.name
}

func (that *TcpSandBox) Setup() error {
	var err error
	addr := fmt.Sprintf("%s:%d", that.cfg.GetString("tcpHost"), that.cfg.GetInt("tcpPort"))
	that.lis, err = net.Listen("tcp", addr)
	if err != nil {
		glog.Fatalf("Error listen. %v", err)
	}
	fmt.Printf("Tcp proxy started! listening %s . vnc server %s:%d\n", that.lis.Addr().String(), that.cfg.GetString("vncHost"), that.cfg.GetInt("vncPort"))
	svrCfg := &rfb.Options{
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
	targetCfg := rfb.TargetConfig{
		Host:     that.cfg.GetString("vncHost"),
		Port:     that.cfg.GetInt("vncPort"),
		Password: that.cfg.GetBytes("vncPassword"),
	}
	for {
		conn, err := that.lis.Accept()
		if err != nil {
			select {
			case <-that.closed:
				return drpc.ErrListenClosed
			default:
			}
			return err
		}
		p := attachNewServerConn(conn, svrCfg, nil, targetCfg)
		go func() {
			for {
				err = <-p.Error()
				glog.Warning(err)
			}
		}()
	}
}

func (that *TcpSandBox) Shutdown() error {
	close(that.closed)
	return that.lis.Close()
}

func (that *TcpSandBox) Service() *easyservice.EasyService {
	return that.service
}
