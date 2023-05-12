package main

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/osgochina/dmicro/drpc"
	"github.com/osgochina/dmicro/drpc/status"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
	"golang.org/x/net/context"
	"io"
	"net"
	"time"
)

// TcpSandBox  Tcp的服务
type TcpSandBox struct {
	id       int
	name     string
	cfg      *gcfg.Config
	service  *easyservice.EasyService
	lis      net.Listener
	closed   chan struct{}
	proxyHub *gmap.StrAnyMap
}

// NewTcpSandBox 创建一个默认的服务沙盒
func NewTcpSandBox(cfg *gcfg.Config) *TcpSandBox {
	id := easyservice.GetNextSandBoxId()
	sBox := &TcpSandBox{
		id:       id,
		name:     fmt.Sprintf("tcp_%d", id),
		cfg:      cfg,
		closed:   make(chan struct{}),
		proxyHub: gmap.NewStrAnyMap(true),
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
	addr := fmt.Sprintf("%s:%d", that.cfg.MustGet(context.TODO(), "tcpHost").String(), that.cfg.MustGet(context.TODO(), "tcpPort").Int())
	that.lis, err = net.Listen("tcp", addr)
	if err != nil {
		glog.Fatalf(context.TODO(), "Error listen. %v", err)
	}
	fmt.Printf("Tcp proxy started! listening %s . vnc server %s:%d\n", that.lis.Addr().String(), that.cfg.MustGet(context.TODO(), "vncHost").String(), that.cfg.MustGet(context.TODO(), "vncPort"))
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
		go func(c net.Conn) {
			defer func() {
				//捕获错误，并且继续执行
				if p := recover(); p != nil {
					err = fmt.Errorf("panic:%v\n%s", p, status.PanicStackTrace())
				}
			}()
			svrSess := session.NewServerSession(
				rfb.OptDesktopName([]byte("Vprix VNC Proxy")),
				rfb.OptHeight(768),
				rfb.OptWidth(1024),
				rfb.OptSecurityHandlers(securityHandlers...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return c, nil
				}),
			)
			timeout := 10 * time.Second
			network := "tcp"
			cliSess := session.NewClient(
				rfb.OptSecurityHandlers([]rfb.ISecurityHandler{&security.ClientAuthVNC{Password: targetCfg.Password}}...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return net.DialTimeout(network, targetCfg.Addr(), timeout)
				}),
			)
			p := vnc.NewVncProxy(cliSess, svrSess)
			remoteKey := c.RemoteAddr().String()
			that.proxyHub.Set(remoteKey, p)
			err = p.Start()
			if err != nil {
				glog.Warning(context.TODO(), err)
				return
			}
			glog.Info(context.TODO(), "proxy session closed")
		}(conn)

	}
}

func (that *TcpSandBox) Shutdown() error {
	close(that.closed)
	return that.lis.Close()
}

func (that *TcpSandBox) Service() *easyservice.EasyService {
	return that.service
}
