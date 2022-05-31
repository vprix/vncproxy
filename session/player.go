package session

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/os/gfile"
	"github.com/vprix/vncproxy/rfb"
	"io"
	"os"
)

type PlayerSession struct {
	RFBFile string
	c       io.ReadWriteCloser
	br      *bufio.Reader
	bw      *bufio.Writer

	cfg      *rfb.ServerConfig // 客户端配置信息
	protocol string            //协议版本
	colorMap rfb.ColorMap      // 颜色地图

	desktopName []byte          // 桌面名称
	encodings   []rfb.IEncoding // 支持的编码列

	fbHeight    uint16          // 缓冲帧高度
	fbWidth     uint16          // 缓冲帧宽度
	pixelFormat rfb.PixelFormat // 像素格式

	swap    *gmap.Map
	quit    chan struct{}
	errorCh chan error
}

func NewPlayerSession(saveFilePath string, cfg *rfb.ServerConfig) *PlayerSession {
	return &PlayerSession{
		RFBFile:   saveFilePath,
		cfg:       cfg,
		encodings: cfg.Encodings,
		errorCh:   cfg.ErrorCh,
		quit:      make(chan struct{}),
		swap:      gmap.New(true),
	}
}

func (that *PlayerSession) Connect() error {
	if !gfile.Exists(that.RFBFile) {
		//_ = gfile.Remove(that.RFBFile)
		return fmt.Errorf("要保存的文件[%s]不存在", that.RFBFile)
	}
	var err error
	that.c, err = gfile.OpenFile(that.RFBFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	that.br = bufio.NewReader(that.c)
	that.bw = bufio.NewWriter(that.c)
	version := make([]byte, len(RBSVersion))
	_, err = that.br.Read(version)
	if err != nil {
		return err
	}
	// 读取rfb协议
	version = make([]byte, len(rfb.ProtoVersion38))
	_, err = that.br.Read(version)
	if err != nil {
		return err
	}
	that.protocol = string(version)
	var secTypeNone int32
	err = binary.Read(that.br, binary.BigEndian, &secTypeNone)
	if err != nil {
		return err
	}
	err = binary.Read(that.br, binary.BigEndian, &that.fbWidth)
	if err != nil {
		return err
	}
	err = binary.Read(that.br, binary.BigEndian, &that.fbHeight)
	if err != nil {
		return err
	}
	err = binary.Read(that.br, binary.BigEndian, &that.pixelFormat)
	if err != nil {
		return err
	}
	//var pad [3]byte
	//err = binary.Read(that.br, binary.BigEndian, &pad)
	//if err != nil {
	//	return err
	//}
	var desktopNameSize uint32
	err = binary.Read(that.br, binary.BigEndian, &desktopNameSize)
	if err != nil {
		return err
	}
	that.desktopName = make([]byte, desktopNameSize)
	_, err = that.Read(that.desktopName)
	if err != nil {
		return err
	}
	return nil
}

// Conn 获取会话底层的网络链接
func (that *PlayerSession) Conn() io.ReadWriteCloser {
	return that.c
}

// Config 获取配置信息
func (that *PlayerSession) Config() interface{} {
	return that.cfg
}

// ProtocolVersion 获取会话使用的协议版本
func (that *PlayerSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *PlayerSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// PixelFormat 获取像素格式
func (that *PlayerSession) PixelFormat() rfb.PixelFormat {
	return that.pixelFormat
}

// SetPixelFormat 设置像素格式
func (that *PlayerSession) SetPixelFormat(pf rfb.PixelFormat) error {
	return nil
}

// ColorMap 获取颜色地图
func (that *PlayerSession) ColorMap() rfb.ColorMap {
	return that.colorMap
}

// SetColorMap 设置颜色地图
func (that *PlayerSession) SetColorMap(cm rfb.ColorMap) {
}

// Encodings 获取当前支持的编码格式
func (that *PlayerSession) Encodings() []rfb.IEncoding {
	return that.encodings
}

// SetEncodings 设置编码格式
func (that *PlayerSession) SetEncodings(encs []rfb.EncodingType) error {
	return nil
}

// Width 获取桌面宽度
func (that *PlayerSession) Width() uint16 {
	return that.fbWidth
}

// SetWidth 设置桌面宽度
func (that *PlayerSession) SetWidth(width uint16) {
	that.fbWidth = width
}

// Height 获取桌面高度
func (that *PlayerSession) Height() uint16 {
	return that.fbHeight
}

// SetHeight 设置桌面高度
func (that *PlayerSession) SetHeight(height uint16) {
}

// DesktopName 获取该会话的桌面名称
func (that *PlayerSession) DesktopName() []byte {
	return that.desktopName
}

// SetDesktopName 设置桌面名称
func (that *PlayerSession) SetDesktopName(name []byte) {
}

func (that *PlayerSession) Flush() error {

	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *PlayerSession) Wait() {
	<-that.quit
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
	if that.quit != nil {
		close(that.quit)
		that.quit = nil
	}
	return that.c.Close()
}

func (that *PlayerSession) Swap() *gmap.Map {
	return that.swap
}
func (that *PlayerSession) Type() rfb.SessionType {
	return rfb.PlayerSessionType
}
