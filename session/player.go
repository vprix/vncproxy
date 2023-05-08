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

type PlayerSession struct {
	c  io.ReadWriteCloser
	br *bufio.Reader
	bw *bufio.Writer

	options         rfb.Options          // 配置信息
	protocol        string               //协议版本
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap *gmap.Map
}

func NewPlayerSession(opts ...rfb.Option) *PlayerSession {
	sess := &PlayerSession{
		swap: gmap.New(true),
	}
	sess.configure(opts...)
	return sess
}

// Init 初始化参数
func (that *PlayerSession) Init(opts ...rfb.Option) error {
	that.configure(opts...)
	return nil
}

func (that *PlayerSession) configure(opts ...rfb.Option) {
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
	if len(that.options.Messages) == 0 {
		that.options.Messages = messages.DefaultClientMessage
	}
	if len(that.options.Encodings) == 0 {
		that.options.Encodings = encodings.DefaultEncodings
	}
}

func (that *PlayerSession) Start() {
	var err error
	that.c, err = that.options.GetConn(that)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}

	that.br = bufio.NewReader(that.c)
	that.bw = bufio.NewWriter(that.c)
	version := make([]byte, len(RBSVersion))
	_, err = that.br.Read(version)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	// 读取rfb协议
	version = make([]byte, len(rfb.ProtoVersion38))
	_, err = that.br.Read(version)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.protocol = string(version)
	var secTypeNone int32
	err = binary.Read(that.br, binary.BigEndian, &secTypeNone)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	var fbWeight uint16
	err = binary.Read(that.br, binary.BigEndian, &fbWeight)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.SetWidth(fbWeight)

	var fbHeight uint16
	err = binary.Read(that.br, binary.BigEndian, &fbHeight)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.SetHeight(fbHeight)

	var pixelFormat rfb.PixelFormat
	err = binary.Read(that.br, binary.BigEndian, &pixelFormat)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.SetPixelFormat(pixelFormat)
	var desktopNameSize uint32
	err = binary.Read(that.br, binary.BigEndian, &desktopNameSize)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	desktopName := make([]byte, desktopNameSize)
	_, err = that.Read(desktopName)
	if err != nil {
		that.options.ErrorCh <- err
		return
	}
	that.SetDesktopName(desktopName)
	return
}

// Conn 获取会话底层的网络链接
func (that *PlayerSession) Conn() io.ReadWriteCloser {
	return that.c
}

// Options 获取配置信息
func (that *PlayerSession) Options() rfb.Options {
	return that.options
}

// ProtocolVersion 获取会话使用的协议版本
func (that *PlayerSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *PlayerSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// Encodings 获取当前支持的编码格式
func (that *PlayerSession) Encodings() []rfb.IEncoding {
	return that.options.Encodings
}

// SetEncodings 设置编码格式
func (that *PlayerSession) SetEncodings(_ []rfb.EncodingType) error {

	return nil
}

func (that *PlayerSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *PlayerSession) Wait() <-chan struct{} {
	return that.options.QuitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *PlayerSession) SecurityHandler() rfb.ISecurityHandler {
	return nil
}

// SetSecurityHandler 设置安全认证处理方法
func (that *PlayerSession) SetSecurityHandler(_ rfb.ISecurityHandler) {
}

// NewEncoding 通过编码类型判断是否支持编码对象
func (that *PlayerSession) NewEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.options.Encodings {
		if enc.Type() == typ && enc.Supported(that) {
			return enc.Clone()
		}
	}
	return nil
}

// Read 从链接中读取数据
func (that *PlayerSession) Read(buf []byte) (int, error) {
	return that.br.Read(buf)
}

// Write 写入数据到链接
func (that *PlayerSession) Write(buf []byte) (int, error) {
	return that.bw.Write(buf)
}

// Close 关闭会话
func (that *PlayerSession) Close() error {
	if that.options.QuitCh != nil {
		that.options.QuitCh <- struct{}{}
	}
	return that.c.Close()
}

func (that *PlayerSession) Swap() *gmap.Map {
	return that.swap
}
func (that *PlayerSession) Type() rfb.SessionType {
	return rfb.PlayerSessionType
}

// SetPixelFormat 设置像素格式
func (that *PlayerSession) SetPixelFormat(pf rfb.PixelFormat) {
	that.options.PixelFormat = pf
}

// SetColorMap 设置颜色地图
func (that *PlayerSession) SetColorMap(cm rfb.ColorMap) {
	that.options.ColorMap = cm
}

// SetWidth 设置桌面宽度
func (that *PlayerSession) SetWidth(width uint16) {
	that.options.Width = width
}

// SetHeight 设置桌面高度
func (that *PlayerSession) SetHeight(height uint16) {
	that.options.Height = height
}

// SetDesktopName 设置桌面名称
func (that *PlayerSession) SetDesktopName(name []byte) {
	that.options.DesktopName = name
}
