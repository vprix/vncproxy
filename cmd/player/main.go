package main

import (
	"fmt"
	"github.com/gogf/gf/os/glog"
	"github.com/vprix/vncproxy/vnc"
	"net"
)

func main() {

	paly := vnc.NewPlayer("D:\\code\\GolandProjects\\vprix-vnc\\abc.rbs", nil)
	var err error
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", 8989)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		glog.Fatalf("Error listen. %v", err)
	}
	for {
		conn, err := lis.Accept()
		_ = paly.Start(conn)
		go func() {
			for {
				err = <-paly.Error()
				glog.Warning(err)
			}
		}()
	}

}
