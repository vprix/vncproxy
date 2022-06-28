package session

import (
	"bufio"
	"encoding/binary"
	"github.com/gogf/gf/container/gmap"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

const RBSVersion = "RBS 001.001\n"

type RecorderSession struct {
	c  io.ReadWriteCloser
	bw *bufio.Writer

	opts            rfb.Options          // 客户端配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap    *gmap.Map
	quitCh  chan struct{} // 退出
	errorCh chan error
}

var _ rfb.ISession = new(RecorderSession)

// NewRecorder 创建客户端会话
func NewRecorder(opts ...rfb.Option) *RecorderSession {
	recorder := &RecorderSession{
		swap: gmap.New(true),
	}
	for _, o := range opts {
		o(&recorder.opts)
	}
	if len(recorder.opts.Encodings) == 0 {
		recorder.opts.Encodings = []rfb.IEncoding{&encodings.RawEncoding{}}
	}
	desktop := &rfb.Desktop{}
	desktop.SetPixelFormat(recorder.opts.PixelFormat)
	recorder.desktop = desktop
	if recorder.opts.QuitCh == nil {
		recorder.opts.QuitCh = make(chan struct{})
	}
	if recorder.opts.ErrorCh == nil {
		recorder.opts.ErrorCh = make(chan error, 32)
	}
	return recorder
}

func (that *RecorderSession) Run() {
	var err error
	that.c, err = that.opts.GetConn()
	if err != nil {
		that.errorCh <- err
		return
	}

	that.bw = bufio.NewWriter(that.c)
	_, err = that.Write([]byte(RBSVersion))
	if err != nil {
		that.errorCh <- err
		return
	}
	_, err = that.Write([]byte(that.ProtocolVersion()))
	if err != nil {
		that.errorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, int32(rfb.SecTypeNone))
	if err != nil {
		that.errorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, int16(that.desktop.Width()))
	if err != nil {
		that.errorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, int16(that.desktop.Height()))
	if err != nil {
		that.errorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, that.desktop.PixelFormat())
	if err != nil {
		that.errorCh <- err
		return
	}
	nameSize := len(that.desktop.DesktopName())
	err = binary.Write(that.bw, binary.BigEndian, uint32(nameSize))
	if err != nil {
		that.errorCh <- err
		return
	}
	_, err = that.Write(that.desktop.DesktopName())
	if err != nil {
		that.errorCh <- err
		return
	}
	err = that.Flush()
	if err != nil {
		that.errorCh <- err
		return
	}
	return
}

// Conn 获取会话底层的网络链接
func (that *RecorderSession) Conn() io.ReadWriteCloser {
	return that.c
}

// Options 获取配置信息
func (that *RecorderSession) Options() rfb.Options {
	return that.opts
}

// ProtocolVersion 获取会话使用的协议版本
func (that *RecorderSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *RecorderSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// Desktop 获取桌面对象
func (that *RecorderSession) Desktop() *rfb.Desktop {
	return that.desktop
}

// Encodings 获取当前支持的编码格式
func (that *RecorderSession) Encodings() []rfb.IEncoding {
	return that.opts.Encodings
}

// SetEncodings 设置编码格式
func (that *RecorderSession) SetEncodings(encs []rfb.EncodingType) error {
	return nil
}

func (that *RecorderSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *RecorderSession) Wait() {
	<-that.quitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *RecorderSession) SecurityHandler() rfb.ISecurityHandler {
	return nil
}

// SetSecurityHandler 设置安全认证处理方法
func (that *RecorderSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) error {
	return nil
}

// GetEncoding 通过编码类型判断是否支持编码对象
func (that *RecorderSession) GetEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.opts.Encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *RecorderSession) Read(buf []byte) (int, error) {
	return 0, nil
}

// Write 写入数据到链接
func (that *RecorderSession) Write(buf []byte) (int, error) {
	return that.bw.Write(buf)
}

// Close 关闭会话
func (that *RecorderSession) Close() error {
	if that.quitCh != nil {
		close(that.quitCh)
		that.quitCh = nil
	}
	return that.c.Close()
}

// Swap session存储的临时变量
func (that *RecorderSession) Swap() *gmap.Map {
	return that.swap
}

// Type session类型
func (that *RecorderSession) Type() rfb.SessionType {
	return rfb.RecorderSessionType
}

func (that *RecorderSession) Messages() []rfb.Message {
	return that.opts.Messages
}
