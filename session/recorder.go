package session

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gtime"
	"github.com/vprix/vncproxy/rfb"
	"io"
	"os"
)

const RBSVersion = "RBS 001.001\n"

type RecorderSession struct {
	rbsFile string
	c       io.ReadWriteCloser
	bw      *bufio.Writer

	cfg      *rfb.ClientConfig // 客户端配置信息
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

var _ rfb.ISession = new(RecorderSession)

// NewRecorder 创建客户端会话
func NewRecorder(saveFilePath string, cfg *rfb.ClientConfig) *RecorderSession {
	return &RecorderSession{
		rbsFile:   saveFilePath,
		cfg:       cfg,
		errorCh:   cfg.ErrorCh,
		encodings: cfg.Encodings,
		quit:      make(chan struct{}),
		swap:      gmap.New(true),
	}
}

func (that *RecorderSession) Connect() error {
	if gfile.Exists(that.rbsFile) {
		that.rbsFile = fmt.Sprintf("%s%s%s_%d%s",
			gfile.Dir(that.rbsFile),
			gfile.Separator,
			gfile.Name(gfile.Basename(that.rbsFile)),
			gtime.Now().Unix(),
			gfile.Ext(gfile.Basename(that.rbsFile)),
		)
		//return fmt.Errorf("要保存的文件[%s]已存在", that.RFBFile)
	}
	var err error
	that.c, err = gfile.OpenFile(that.rbsFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	that.bw = bufio.NewWriter(that.c)
	_, err = that.Write([]byte(RBSVersion))
	if err != nil {
		return err
	}
	_, err = that.Write([]byte(that.ProtocolVersion()))
	if err != nil {
		return err
	}
	err = binary.Write(that.bw, binary.BigEndian, int32(rfb.SecTypeNone))
	if err != nil {
		return err
	}
	err = binary.Write(that.bw, binary.BigEndian, int16(that.Width()))
	if err != nil {
		return err
	}
	err = binary.Write(that.bw, binary.BigEndian, int16(that.Height()))
	if err != nil {
		return err
	}
	err = binary.Write(that.bw, binary.BigEndian, that.pixelFormat)
	if err != nil {
		return err
	}
	nameSize := len(that.desktopName)
	err = binary.Write(that.bw, binary.BigEndian, uint32(nameSize))
	if err != nil {
		return err
	}
	_, err = that.Write(that.desktopName)
	if err != nil {
		return err
	}
	return that.Flush()
}

// Conn 获取会话底层的网络链接
func (that *RecorderSession) Conn() io.ReadWriteCloser {
	return that.c
}

// Config 获取配置信息
func (that *RecorderSession) Config() interface{} {
	return that.cfg
}

// ProtocolVersion 获取会话使用的协议版本
func (that *RecorderSession) ProtocolVersion() string {
	return that.protocol
}

// SetProtocolVersion 设置支持的协议版本
func (that *RecorderSession) SetProtocolVersion(pv string) {
	that.protocol = pv
}

// PixelFormat 获取像素格式
func (that *RecorderSession) PixelFormat() rfb.PixelFormat {
	return that.pixelFormat
}

// SetPixelFormat 设置像素格式
func (that *RecorderSession) SetPixelFormat(pf rfb.PixelFormat) error {
	that.pixelFormat = pf
	return nil
}

// ColorMap 获取颜色地图
func (that *RecorderSession) ColorMap() rfb.ColorMap {
	return that.colorMap
}

// SetColorMap 设置颜色地图
func (that *RecorderSession) SetColorMap(cm rfb.ColorMap) {
	that.colorMap = cm
}

// Encodings 获取当前支持的编码格式
func (that *RecorderSession) Encodings() []rfb.IEncoding {
	return that.encodings
}

// SetEncodings 设置编码格式
func (that *RecorderSession) SetEncodings(encs []rfb.EncodingType) error {
	return nil
}

// Width 获取桌面宽度
func (that *RecorderSession) Width() uint16 {
	return that.fbWidth
}

// SetWidth 设置桌面宽度
func (that *RecorderSession) SetWidth(width uint16) {
	that.fbWidth = width
}

// Height 获取桌面高度
func (that *RecorderSession) Height() uint16 {
	return that.fbHeight
}

// SetHeight 设置桌面高度
func (that *RecorderSession) SetHeight(height uint16) {
	that.fbHeight = height
}

// DesktopName 获取该会话的桌面名称
func (that *RecorderSession) DesktopName() []byte {
	return that.desktopName
}

// SetDesktopName 设置桌面名称
func (that *RecorderSession) SetDesktopName(name []byte) {
	that.desktopName = name
}

func (that *RecorderSession) Flush() error {
	return that.bw.Flush()
}

// Wait 等待会话处理完成
func (that *RecorderSession) Wait() {
	<-that.quit
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
	if that.quit != nil {
		close(that.quit)
		that.quit = nil
	}
	return that.c.Close()
}
func (that *RecorderSession) Swap() *gmap.Map {
	return that.swap
}

func (that *RecorderSession) Type() rfb.SessionType {
	return rfb.RecorderSessionType
}
