package rfb

import (
	"fmt"
	"time"
)

type TargetConfig struct {
	Network  string        // 网络协议
	Timeout  time.Duration // 超时时间
	Host     string        // vnc服务端地址
	Port     int           // vnc服务端端口
	Password []byte        // vnc服务端密码
}

func (that TargetConfig) Addr() string {
	return fmt.Sprintf("%s:%d", that.Host, that.Port)
}
