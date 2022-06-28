package session

import (
	"bufio"
	"encoding/binary"
	"github.com/gogf/gf/container/gmap"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/rfb"
	"io"
)

type PlayerSession struct {
	c  io.ReadWriteCloser
	br *bufio.Reader
	bw *bufio.Writer

	options         rfb.Options          // 配置信息
	protocol        string               //协议版本
	desktop         *rfb.Desktop         // 桌面对象
	encodings       []rfb.IEncoding      // 支持的编码列
	securityHandler rfb.ISecurityHandler // 安全认证方式

	swap    *gmap.Map
	quitCh  chan struct{}
	errorCh chan error
}

func NewPlayerSession(options rfb.Options) *PlayerSession {
	enc := options.Encodings
	if len(options.Encodings) == 0 {
		enc = []rfb.IEncoding{&encodings.RawEncoding{}}
	}
	desktop := &rfb.Desktop{}
	if options.QuitCh == nil {
		options.QuitCh = make(chan struct{})
	}
	if options.ErrorCh == nil {
		options.ErrorCh = make(chan error, 32)
	}
	return &PlayerSession{
		options:   options,
		desktop:   desktop,
		encodings: enc,
		errorCh:   options.ErrorCh,
		quitCh:    options.QuitCh,
		swap:      gmap.New(true),
	}
}

// Init 初始化参数
func (that *PlayerSession) Init(opts ...rfb.Option) error {
	for _, o := range opts {
		o(&that.options)
	}
	return nil
}

func (that *PlayerSession) Run() {
	var err error
	that.c, err = that.options.GetConn()
	if err != nil {
		that.errorCh <- err
		return
	}

	that.br = bufio.NewReader(that.c)
	that.bw = bufio.NewWriter(that.c)
	version := make([]byte, len(RBSVersion))
	_, err = that.br.Read(version)
	if err != nil {
		that.errorCh <- err
		return
	}
	// 读取rfb协议
	version = make([]byte, len(rfb.ProtoVersion38))
	_, err = that.br.Read(version)
	if err != nil {
		that.errorCh <- err
		return
	}
	that.protocol = string(version)
	var secTypeNone int32
	err = binary.Read(that.br, binary.BigEndian, &secTypeNone)
	if err != nil {
		that.errorCh <- err
		return
	}
	var fbWeight uint16
	err = binary.Read(that.br, binary.BigEndian, &fbWeight)
	if err != nil {
		that.errorCh <- err
		return
	}
	that.desktop.SetWidth(fbWeight)

	var fbHeight uint16
	err = binary.Read(that.br, binary.BigEndian, &fbHeight)
	if err != nil {
		that.errorCh <- err
		return
	}
	that.desktop.SetHeight(fbWeight)

	var pixelFormat rfb.PixelFormat
	err = binary.Read(that.br, binary.BigEndian, &pixelFormat)
	if err != nil {
		that.errorCh <- err
		return
	}
	that.desktop.SetPixelFormat(pixelFormat)
	var desktopNameSize uint32
	err = binary.Read(that.br, binary.BigEndian, &desktopNameSize)
	if err != nil {
		that.errorCh <- err
		return
	}
	desktopName := make([]byte, desktopNameSize)
	_, err = that.Read(desktopName)
	if err != nil {
		that.errorCh <- err
		return
	}
	that.desktop.SetDesktopName(desktopName)
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

// Desktop 获取桌面对象
func (that *PlayerSession) Desktop() *rfb.Desktop {
	return that.desktop
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
	return that.encodings
}

// SetEncodings 设置编码格式
func (that *PlayerSession) SetEncodings(encs []rfb.EncodingType) error {
	return nil
}

func (that *PlayerSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *PlayerSession) Wait() {
	<-that.quitCh
}

// SecurityHandler 返回安全认证处理方法
func (that *PlayerSession) SecurityHandler() rfb.ISecurityHandler {
	return nil
}

// SetSecurityHandler 设置安全认证处理方法
func (that *PlayerSession) SetSecurityHandler(securityHandler rfb.ISecurityHandler) error {
	return nil
}

// GetEncoding 通过编码类型判断是否支持编码对象
func (that *PlayerSession) GetEncoding(typ rfb.EncodingType) rfb.IEncoding {
	for _, enc := range that.encodings {
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
	if that.quitCh != nil {
		close(that.quitCh)
		that.quitCh = nil
	}
	return that.c.Close()
}

func (that *PlayerSession) Swap() *gmap.Map {
	return that.swap
}
func (that *PlayerSession) Type() rfb.SessionType {
	return rfb.PlayerSessionType
}
