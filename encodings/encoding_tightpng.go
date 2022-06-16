package encodings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

type TightPngEncoding struct {
	buff *bytes.Buffer
}

func (that *TightPngEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *TightPngEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &TightPngEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *TightPngEncoding) Type() rfb.EncodingType {
	return rfb.EncTightPng
}

func (that *TightPngEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	_, err := that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}
func (that *TightPngEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	pf := session.Desktop().PixelFormat()
	bytesPixel := calcTightBytePerPixel(&pf)
	compressionControl, err := ReadUint8(session)
	if err != nil {
		return nil
	}
	_ = binary.Write(that.buff, binary.BigEndian, compressionControl)

	compType := compressionControl >> 4 & 0x0F

	switch compType {
	case tightCompressionPNG:
		size, err := that.ReadCompactLen(session)
		if err != nil {
			return err
		}
		bt, err := ReadBytes(size, session)
		if err != nil {
			return err
		}
		_, _ = that.buff.Write(bt)

	case tightCompressionFill:
		bt, err := ReadBytes(int(bytesPixel), session)
		if err != nil {
			return err
		}
		_, _ = that.buff.Write(bt)
	default:
		return fmt.Errorf("unknown tight compression %d", compType)
	}
	return nil
}

func (that *TightPngEncoding) ReadCompactLen(session rfb.ISession) (int, error) {
	var err error
	part, err := ReadUint8(session)
	if err := binary.Write(that.buff, binary.BigEndian, part); err != nil {
		return 0, err
	}
	size := uint32(part & 0x7F)
	if (part & 0x80) != 0 {
		part, err = ReadUint8(session)
		if err := binary.Write(that.buff, binary.BigEndian, part); err != nil {
			return 0, err
		}
		size |= uint32(part&0x7F) << 7
		if (part & 0x80) != 0 {
			part, err = ReadUint8(session)
			if err := binary.Write(that.buff, binary.BigEndian, part); err != nil {
				return 0, err
			}
			size |= uint32(part&0xFF) << 14
		}
	}

	return int(size), err
}
