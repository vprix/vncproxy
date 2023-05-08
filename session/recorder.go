package session

import (
	"bufio"
	"encoding/binary"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

const RBSVersion = "RBS 001.001\n"

type RecorderSession struct {
	c  io.ReadWriteCloser
	bw *bufio.Writer

	options  rfb.Options // 客户端配置信息
	protocol string      //协议版本

	swap *gmap.Map
}

var _ rfb.ISession = new(RecorderSession)

// NewRecorder 创建客户端会话
func NewRecorder(opts ...rfb.Option) *RecorderSession {
	recorder := &RecorderSession{
		swap: gmap.New(true),
	}
	recorder.configure(opts...)
	return recorder
}

// Init 初始化参数
func (that *RecorderSession) Init(opts ...rfb.Option) error {
	that.configure(opts...)
	return nil
}

func (that *RecorderSession) configure(opts ...rfb.Option) {
	for _, o := range opts {
		o(&that.options)
	}
	if that.options.PixelFormat.BPP == 0 {
		that.options.PixelFormat = rfb.PixelFormat32bit
	}
	if that.options.QuitCh == nil {
		that.options.QuitCh = make(chan struct{})
	}
	if that.options.ErrorCh == nil {
		that.options.ErrorCh = make(chan error, 32)
	}
	if that.options.Input == nil {
		that.options.Input = make(chan rfb.Message)
	}
	if that.options.Output == nil {
		that.options.Output = make(chan rfb.Message)
	}
	if len(that.options.Handlers) == 0 {
		that.options.Handlers = DefaultClientHandlers
	}
	if len(that.options.Messages) == 0 {
		that.options.Messages = messages.DefaultServerMessages
	}
	if len(that.options.Encodings) == 0 {
		that.options.Encodings = encodings.DefaultEncodings
	}
}
func (that *RecorderSession) Start() {
	var err error
	that.c, err = that.options.GetConn(that)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}

	that.bw = bufio.NewWriter(that.c)
	_, err = that.Write([]byte(RBSVersion))
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	_, err = that.Write([]byte(that.ProtocolVersion()))
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, int32(rfb.SecTypeNone))
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, int16(that.options.Width))
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, int16(that.options.Height))
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	err = binary.Write(that.bw, binary.BigEndian, that.options.PixelFormat)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	nameSize := len(that.options.DesktopName)
	err = binary.Write(that.bw, binary.BigEndian, uint32(nameSize))
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	_, err = that.Write(that.options.DesktopName)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	err = that.Flush()
	if err != nil {
		that.options.ErrorCh <- err
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

// Encodings 获取当前支持的编码格式
func (that *RecorderSession) Encodings() []rfb.IEncoding {
	return that.options.Encodings
}

// SetEncodings 设置编码格式
func (that *RecorderSession) SetEncodings(_ []rfb.EncodingType) error {
	return nil
}

func (that *RecorderSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *RecorderSession) Wait() <-chan struct{} {
	return that.options.QuitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *RecorderSession) SecurityHandler() rfb.ISecurityHandler {
	return nil
}

// SetSecurityHandler 设置安全认证处理方法
func (that *RecorderSession) SetSecurityHandler(_ rfb.ISecurityHandler) {
}

// NewEncoding 通过编码类型判断是否支持编码对象
func (that *RecorderSession) NewEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.options.Encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *RecorderSession) Read(_ []byte) (int, error) {
	return 0, nil
}

// Write 写入数据到链接
func (that *RecorderSession) Write(buf []byte) (int, error) {
	return that.bw.Write(buf)
}

// Close 关闭会话
func (that *RecorderSession) Close() error {
	if that.options.QuitCh != nil {
		that.options.QuitCh <- struct{}{}
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

// SetPixelFormat 设置像素格式
func (that *RecorderSession) SetPixelFormat(pf rfb.PixelFormat) {
	that.options.PixelFormat = pf
}

// SetColorMap 设置颜色地图
func (that *RecorderSession) SetColorMap(cm rfb.ColorMap) {
	that.options.ColorMap = cm
}

// SetWidth 设置桌面宽度
func (that *RecorderSession) SetWidth(width uint16) {
	that.options.Width = width
}

// SetHeight 设置桌面高度
func (that *RecorderSession) SetHeight(height uint16) {
	that.options.Height = height
}

// SetDesktopName 设置桌面名称
func (that *RecorderSession) SetDesktopName(name []byte) {
	that.options.DesktopName = name
}
