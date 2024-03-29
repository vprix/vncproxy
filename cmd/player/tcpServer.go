package main

import (
	"fmt"
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
	addr := fmt.Sprintf("%s:%d", that.cfg.MustGet(context.TODO(), "tcpHost"), that.cfg.MustGet(context.TODO(), "tcpPort"))
	that.lis, err = net.Listen("tcp", addr)
	if err != nil {
		glog.Fatalf(context.TODO(), "Error listen. %v", err)
	}
	fmt.Printf("Tcp proxy started! listening %s . vnc server %s:%d\n", that.lis.Addr().String(), that.cfg.MustGet(context.TODO(), "vncHost"), that.cfg.MustGet(context.TODO(), "vncPort"))
	securityHandlers := []rfb.ISecurityHandler{&security.ServerAuthNone{}}
	if len(that.cfg.MustGet(context.TODO(), "proxyPassword").Bytes()) > 0 {
		securityHandlers = append(securityHandlers, &security.ServerAuthVNC{Password: that.cfg.MustGet(context.TODO(), "proxyPassword").Bytes()})
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
			svrSession := session.NewServerSession(
				rfb.OptSecurityHandlers(securityHandlers...),
				rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
					return c, nil
				}),
			)
			play := vnc.NewPlayer(that.cfg.MustGet(context.TODO(), "rbsFile").String(), svrSession)
			err = play.Start()
			if err != nil {
				glog.Warning(context.TODO(), err)
				return
			}
			glog.Info(context.TODO(), "play finished")
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
