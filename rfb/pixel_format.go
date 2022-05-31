package rfb

import (
	"encoding/binary"
	"fmt"
)

const PixelFormatLen = 16

var (
	// PixelFormat8bit 获取8bit像素格式
	PixelFormat8bit = NewPixelFormat(8)
	// PixelFormat16bit 获取15bit像素格式
	PixelFormat16bit = NewPixelFormat(16)
	// PixelFormat32bit 获取32bit像素格式
	PixelFormat32bit = NewPixelFormat(32)
	// PixelFormatAten returns pixel format used in Aten IKVM
	PixelFormatAten = NewPixelFormatAten()
)

// PixelFormat 像素格式结构体
type PixelFormat struct {
	BPP        uint8   // 1 byte,像素的位数，位数越大，色彩越丰富。只支持[8|16|32] 该值必须大于Depth
	Depth      uint8   // 1 byte,色深，像素中表示色彩的位数
	BigEndian  uint8   // 1 byte,多字节像素的字节序，非零即大端序
	TrueColor  uint8   // 1 byte,1 表示真彩色，pixel 的值表示 RGB 颜色；0 表示调色板，pexel 的值表示颜色在调色板的偏移量
	RedMax     uint16  // 2 byte,红色的长度
	GreenMax   uint16  // 2 byte,绿色的长度
	BlueMax    uint16  // 2 byte,蓝色的长度
	RedShift   uint8   // 1 byte,红色的位移量
	GreenShift uint8   // 1 byte,绿色的位移量
	BlueShift  uint8   // 1 byte,蓝色的偏移量
	_          [3]byte // 填充字节
}

func (that PixelFormat) String() string {
	return fmt.Sprintf("{ bpp: %d depth: %d big-endian: %d true-color: %d red-max: %d green-max: %d blue-max: %d red-shift: %d green-shift: %d blue-shift: %d }",
		that.BPP, that.Depth, that.BigEndian, that.TrueColor, that.RedMax, that.GreenMax, that.BlueMax, that.RedShift, that.GreenShift, that.BlueShift)
}

// Order 确定像素格式是使用了大端字节序还是小端字节序
func (that PixelFormat) Order() binary.ByteOrder {
	if that.BigEndian == 1 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

func NewPixelFormat(bpp uint8) PixelFormat {
	bigEndian := uint8(0)
	//	rgbMax := uint16(math.Exp2(float64(bpp))) - 1
	rMax := uint16(255)
	gMax := uint16(255)
	bMax := uint16(255)
	var (
		tc         = uint8(1)
		rs, gs, bs uint8
		depth      uint8
	)
	switch bpp {
	case 8:
		tc = 0
		depth = 8
		rs, gs, bs = 0, 0, 0
	case 16:
		depth = 16
		rs, gs, bs = 0, 4, 8
	case 32:
		depth = 24
		//	rs, gs, bs = 0, 8, 16
		rs, gs, bs = 16, 8, 0
	}
	return PixelFormat{
		BPP:        bpp,
		Depth:      depth,
		BigEndian:  bigEndian,
		TrueColor:  tc,
		RedMax:     rMax,
		GreenMax:   gMax,
		BlueMax:    bMax,
		RedShift:   rs,
		GreenShift: gs,
		BlueShift:  bs,
		//_:          [3]byte{},
	}
}

func NewPixelFormatAten() PixelFormat {
	return PixelFormat{
		BPP:        16,
		Depth:      15,
		BigEndian:  0,
		TrueColor:  1,
		RedMax:     (1 << 5) - 1,
		GreenMax:   (1 << 5) - 1,
		BlueMax:    (1 << 5) - 1,
		RedShift:   10,
		GreenShift: 5,
		BlueShift:  0,
		//_:          [3]byte{},
	}
}
