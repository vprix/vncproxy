package encodings

import (
	"encoding/binary"
	"github.com/vprix/vncproxy/rfb"
	"math"
)

type XCursorPseudoEncoding struct {
	PrimaryR, PrimaryG, PrimaryB       uint8  // 主颜色
	SecondaryR, SecondaryG, SecondaryB uint8  // 次颜色
	Bitmap                             []byte //颜色位图
	Bitmask                            []byte //透明度位掩码
}

func (that *XCursorPseudoEncoding) Supported(session rfb.ISession) bool {
	return true
}
func (that *XCursorPseudoEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &XCursorPseudoEncoding{}
	if len(data) > 0 && data[0] {
		obj.PrimaryR = that.PrimaryR
		obj.PrimaryG = that.PrimaryG
		obj.PrimaryB = that.PrimaryB
		obj.SecondaryR = that.SecondaryR
		obj.SecondaryG = that.SecondaryG
		obj.SecondaryB = that.SecondaryB
		Bitmap := make([]byte, len(that.Bitmap))
		Bitmask := make([]byte, len(that.Bitmask))
		copy(Bitmap, that.Bitmap)
		copy(Bitmask, that.Bitmask)
		obj.Bitmap = Bitmap
		obj.Bitmask = Bitmask
	}
	return obj
}
func (that *XCursorPseudoEncoding) Type() rfb.EncodingType {
	return rfb.EncXCursorPseudo
}

func (that *XCursorPseudoEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if err := binary.Read(session, binary.BigEndian, &that.PrimaryR); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.PrimaryG); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.PrimaryB); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.SecondaryR); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.SecondaryG); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.SecondaryB); err != nil {
		return err
	}

	bitMapSize := int(math.Floor((float64(rect.Width)+7)/8) * float64(rect.Height))
	bitMaskSize := int(math.Floor((float64(rect.Width)+7)/8) * float64(rect.Height))

	that.Bitmap = make([]byte, bitMapSize)
	that.Bitmask = make([]byte, bitMaskSize)
	if err := binary.Read(session, binary.BigEndian, &that.Bitmap); err != nil {
		return err
	}
	if err := binary.Read(session, binary.BigEndian, &that.Bitmask); err != nil {
		return err
	}

	return nil
}

func (that *XCursorPseudoEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if err := binary.Write(session, binary.BigEndian, that.PrimaryR); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.PrimaryG); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.PrimaryB); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.SecondaryR); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.SecondaryG); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.SecondaryB); err != nil {
		return err
	}

	if err := binary.Write(session, binary.BigEndian, that.Bitmap); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that.Bitmask); err != nil {
		return err
	}

	return nil
}
