package rfb

import "encoding/binary"

// ColorMap 颜色地图
type ColorMap [256]Color

// Color 表示颜色地图中的一个颜色。
type Color struct {
	pf      *PixelFormat
	cm      *ColorMap
	cmIndex uint32 // Only valid if pf.TrueColor is false.
	R, G, B uint16
}

// 写入颜色数据
func (that *Color) Write(session ISession) error {
	var err error
	pf := session.Options().PixelFormat
	order := pf.Order()
	pixel := that.cmIndex
	if that.pf.TrueColor != 0 {
		pixel = uint32(that.R) << pf.RedShift
		pixel |= uint32(that.G) << pf.GreenShift
		pixel |= uint32(that.B) << pf.BlueShift
	}

	switch pf.BPP {
	case 8:
		err = binary.Write(session, order, byte(pixel))
	case 16:
		err = binary.Write(session, order, uint16(pixel))
	case 32:
		err = binary.Write(session, order, uint32(pixel))
	}

	return err
}

// 从链接中读取颜色偏移量
func (that *Color) Read(session ISession) error {
	order := that.pf.Order()
	var pixel uint32

	switch that.pf.BPP {
	case 8:
		var px uint8
		if err := binary.Read(session, order, &px); err != nil {
			return err
		}
		pixel = uint32(px)
	case 16:
		var px uint16
		if err := binary.Read(session, order, &px); err != nil {
			return err
		}
		pixel = uint32(px)
	case 32:
		var px uint32
		if err := binary.Read(session, order, &px); err != nil {
			return err
		}
		pixel = px
	}

	if that.pf.TrueColor != 0 {
		that.R = uint16((pixel >> that.pf.RedShift) & uint32(that.pf.RedMax))
		that.G = uint16((pixel >> that.pf.GreenShift) & uint32(that.pf.GreenMax))
		that.B = uint16((pixel >> that.pf.BlueShift) & uint32(that.pf.BlueMax))
	} else {
		*that = that.cm[pixel]
		that.cmIndex = pixel
	}
	return nil
}
