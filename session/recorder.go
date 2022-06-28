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

	options         *rfb.Options         // 客户端配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap    *gmap.Map
	quitCh  chan struct{} // 退出
	errorCh chan error
}

var _ rfb.ISession = new(RecorderSession)

// NewRecorder 创建客户端会话
func NewRecorder(options *rfb.Options) *RecorderSession {
	enc := options.Encodings
	if len(options.Encodings) == 0 {
		enc = []rfb.IEncoding{&encodings.RawEncoding{}}
	}
	desktop := &rfb.Desktop{}
	desktop.SetPixelFormat(options.PixelFormat)

	if options.QuitCh == nil {
		options.QuitCh = make(chan struct{})
	}
	if options.ErrorCh == nil {
		options.ErrorCh = make(chan error, 32)
	}
	return &RecorderSession{
		options:   options,
		encodings: enc,
		desktop:   desktop,
		quitCh:    options.QuitCh,
		errorCh:   options.ErrorCh,
		swap:      gmap.New(true),
	}
}

func (that *RecorderSession) Run() {
	var err error
	that.c, err = that.options.CreateConn()
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
func (that *RecorderSession) Options() *rfb.Options {
	return that.options
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
	return that.encodings
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
	for _, enc := range that.encodings {
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
	return that.options.Messages
}
