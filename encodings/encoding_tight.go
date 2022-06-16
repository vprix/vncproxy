package encodings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

var TightMinToCompress int = 12

const (
	tightCompressionBasic = 0
	tightCompressionFill  = 0x08
	tightCompressionJPEG  = 0x09
	tightCompressionPNG   = 0x0A
)

const (
	TightFilterCopy     = 0
	TightFilterPalette  = 1
	TightFilterGradient = 2
)

type TightEncoding struct {
	buff *bytes.Buffer
}

func (that *TightEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *TightEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &TightEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}
func (that *TightEncoding) Type() rfb.EncodingType {
	return rfb.EncTight
}

func calcTightBytePerPixel(pf *rfb.PixelFormat) int {
	bytesPerPixel := int(pf.BPP / 8)

	var bytesPerPixelTight int
	if 24 == pf.Depth && 32 == pf.BPP {
		bytesPerPixelTight = 3
	} else {
		bytesPerPixelTight = bytesPerPixel
	}
	return bytesPerPixelTight
}

func (that *TightEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	_, err := that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}
func (that *TightEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
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
	case tightCompressionFill:

		bt, err := ReadBytes(bytesPixel, session)
		if err != nil {
			return err
		}
		_, _ = that.buff.Write(bt)

	case tightCompressionJPEG:
		if pf.BPP == 8 {
			return errors.New("Tight encoding: JPEG is not supported in 8 bpp mode. ")
		}
		size, err := that.ReadCompactLen(session)
		jpegBytes, err := ReadBytes(size, session)
		if err != nil {
			return err
		}
		_, _ = that.buff.Write(jpegBytes)
	default:
		if compType > tightCompressionJPEG {
			return errors.New("Compression control byte is incorrect!")
		}
		err = that.handleTightFilters(session, rect, &pf, compressionControl)
		return err
	}
	return nil
}

func (that *TightEncoding) ReadCompactLen(session rfb.ISession) (int, error) {
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

func (that *TightEncoding) handleTightFilters(session rfb.ISession, rect *rfb.Rectangle, pf *rfb.PixelFormat, compCtl uint8) error {

	var FilterIdMask uint8 = 0x40

	var filterId uint8
	var err error

	if (compCtl & FilterIdMask) > 0 {
		filterId, err = ReadUint8(session)

		if err != nil {
			return fmt.Errorf("error in handling tight encoding, reading filterid: %s", err.Error())
		}
		_ = binary.Write(that.buff, binary.BigEndian, filterId)
	}
	bytesPixel := calcTightBytePerPixel(pf)
	lengthCurrentBPP := bytesPixel * int(rect.Width) * int(rect.Height)
	switch filterId {
	case TightFilterPalette:
		palette, err := that.readTightPalette(session, bytesPixel)
		if err != nil {
			return err
		}
		var dataLength int
		if palette == 2 {
			dataLength = int(rect.Height) * ((int(rect.Width) + 7) / 8)
		} else {
			dataLength = int(rect.Width) * int(rect.Height)
		}
		err = that.ReadTightData(dataLength, session)
		if err != nil {
			return err
		}
	case TightFilterGradient:
		err = that.ReadTightData(lengthCurrentBPP, session)
		if err != nil {
			return fmt.Errorf("handleTightFilters: error in handling tight encoding, Reading GRADIENT_FILTER: %v", err)
		}

	case TightFilterCopy:
		err = that.ReadTightData(lengthCurrentBPP, session)
		if err != nil {
			return fmt.Errorf("handleTightFilters: error in handling tight encoding, Reading BASIC_FILTER: %v", err)
		}
	default:
		return fmt.Errorf("handleTightFilters: Bad tight filter id: %d", filterId)
	}
	return nil
}

func (that *TightEncoding) readTightPalette(session rfb.ISession, bytesPixel int) (int, error) {
	colorCount, err := ReadUint8(session)
	if err != nil {
		return 0, fmt.Errorf("handleTightFilters: error in handling tight encoding, reading TightFilterPalette: %v", err)
	}
	_ = binary.Write(that.buff, binary.BigEndian, colorCount)
	paletteSize := colorCount + 1
	paletteColorBytes, err := ReadBytes(int(paletteSize)*bytesPixel, session)
	_, _ = that.buff.Write(paletteColorBytes)
	return int(paletteSize), nil
}

func (that *TightEncoding) ReadTightData(dataSize int, session rfb.ISession) error {
	if dataSize < TightMinToCompress {
		b, err := ReadBytes(dataSize, session)
		if err == nil {
			_, _ = that.buff.Write(b)
		}
		return err
	}
	zlibDataLen, err := that.ReadCompactLen(session)
	if err != nil {
		return err
	}
	zippedBytes, err := ReadBytes(zlibDataLen, session)
	if err != nil {
		return err
	}
	_, _ = that.buff.Write(zippedBytes)
	return nil
}
