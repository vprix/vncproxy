package encodings

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/rfb"
	"image/color"
	"io"
)

const (
	ZRLERawPixelData = 0
	ZRLESingleColour = 1
)

// ZRLEEncoding ZRLE(Zlib Run - Length Encoding),它结合了zlib 压缩，片技术、调色板和运行长度编码。
// 在传输中，矩形以4 字节长度区域开始，紧接着是zlib 压缩的数据，一个单一的 zlib“流”对象被用在RFB协议的连接上，
// 因此ZRLE矩形必须严格的按照顺序进行编码和译码。
type ZRLEEncoding struct {
	buff *bytes.Buffer
}

func (that *ZRLEEncoding) Type() rfb.EncodingType {
	return rfb.EncZRLE
}

func (that *ZRLEEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *ZRLEEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &ZRLEEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *ZRLEEncoding) Read(sess rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	size, err := ReadUint32(sess)
	if err != nil {
		return err
	}
	err = binary.Write(that.buff, binary.BigEndian, size)
	if err != nil {
		return err
	}
	b, err := ReadBytes(int(size), sess)
	if err != nil {
		return err
	}
	_, err = that.buff.Write(b)
	if err != nil {
		return err
	}
	return nil
}
func (that *ZRLEEncoding) Write(sess rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	if sess.Type() == rfb.CanvasSessionType {
		return that.draw(sess.Conn().(*canvas.VncCanvas), sess.Options().PixelFormat, rect)
	}
	_, err := that.buff.WriteTo(sess)
	that.buff.Reset()
	return err
}

// 绘制画布
func (that *ZRLEEncoding) draw(cv *canvas.VncCanvas, pf rfb.PixelFormat, rect *rfb.Rectangle) error {
	var size uint32
	err := binary.Read(that.buff, binary.BigEndian, &size)
	if err != nil {
		return err
	}
	b, err := ReadBytes(int(size), that.buff)
	if err != nil {
		return err
	}
	bytesBuff := bytes.NewBuffer(b)
	unZipper, err := zlib.NewReader(bytesBuff)
	if err != nil {
		return err
	}

	for tileOffsetY := 0; tileOffsetY < int(rect.Height); tileOffsetY += 64 {

		tileHeight := min(64, int(rect.Height)-tileOffsetY)

		for tileOffsetX := 0; tileOffsetX < int(rect.Width); tileOffsetX += 64 {

			tileWidth := min(64, int(rect.Width)-tileOffsetX)
			// 获取二级编码格式
			subEnc, err := ReadUint8(unZipper)
			if err != nil {
				return fmt.Errorf("renderZRLE: error while reading subencoding: %v", err)
			}

			switch {
			case subEnc == ZRLERawPixelData: // 原始编码格式
				err = that.readZRLERaw(cv, unZipper, &pf, int(rect.X)+tileOffsetX, int(rect.Y)+tileOffsetY, tileWidth, tileHeight)
				if err != nil {
					return fmt.Errorf("renderZRLE: error while reading Raw tile: %v", err)
				}
			case subEnc == ZRLESingleColour: // 获取一个颜色，填充指定区域
				co, err := readCPixel(cv, unZipper, &pf)
				if err != nil {
					return fmt.Errorf("renderZRLE: error while reading CPixel for bgColor tile: %v", err)
				}
				myRect := canvas.MakeRect(int(rect.X)+tileOffsetX, int(rect.Y)+tileOffsetY, tileWidth, tileHeight)
				cv.FillRect(&myRect, co)
			case subEnc >= 2 && subEnc <= 16: // 调色版编码
				err = that.handlePaletteTile(cv, unZipper, tileOffsetX, tileOffsetY, tileWidth, tileHeight, subEnc, &pf, rect)
				if err != nil {
					return err
				}
			case subEnc == 128: //
				err = that.handlePlainRLETile(cv, unZipper, tileOffsetX, tileOffsetY, tileWidth, tileHeight, &pf, rect)
				if err != nil {
					return err
				}
			case subEnc >= 130:
				err = that.handlePaletteRLETile(cv, unZipper, tileOffsetX, tileOffsetY, tileWidth, tileHeight, subEnc, &pf, rect)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unknown ZRLE subencoding: %v ", subEnc)
			}
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (that *ZRLEEncoding) readZRLERaw(cv *canvas.VncCanvas, reader io.Reader, pf *rfb.PixelFormat, tx, ty, tw, th int) error {
	for y := 0; y < th; y++ {
		for x := 0; x < tw; x++ {
			col, err := readCPixel(cv, reader, pf)
			if err != nil {
				return err
			}
			cv.Set(tx+x, ty+y, col)
		}
	}

	return nil
}

// 获取像素格式
func readCPixel(cv *canvas.VncCanvas, c io.Reader, pf *rfb.PixelFormat) (*color.RGBA, error) {
	if pf.TrueColor == 0 {
		return nil, errors.New("support for non true color formats was not implemented")
	}

	isZRLEFormat := IsCPixelSpecific(pf)
	var col *color.RGBA
	if isZRLEFormat {
		tBytes, err := ReadBytes(3, c)
		if err != nil {
			return nil, err
		}
		if pf.BigEndian != 1 {
			col = &color.RGBA{
				B: tBytes[0],
				G: tBytes[1],
				R: tBytes[2],
				A: uint8(1),
			}
		} else {
			col = &color.RGBA{
				R: tBytes[0],
				G: tBytes[1],
				B: tBytes[2],
				A: uint8(1),
			}
		}
		return col, nil
	}

	col, err := cv.ReadColor(c, pf)
	if err != nil {
		return nil, fmt.Errorf("readCPixel: Error while reading zrle: %v", err)
	}

	return col, nil
}

func IsCPixelSpecific(pf *rfb.PixelFormat) bool {
	significant := int(pf.RedMax<<pf.RedShift | pf.GreenMax<<pf.GreenShift | pf.BlueMax<<pf.BlueShift)

	if pf.Depth <= 24 && 32 == pf.BPP && ((significant&0x00ff000000) == 0 || (significant&0x000000ff) == 0) {
		return true
	}
	return false
}

// 调色板编码
func (that *ZRLEEncoding) handlePaletteTile(cv *canvas.VncCanvas, unZipper io.Reader, tileOffsetX, tileOffsetY, tileWidth, tileHeight int, subEnc uint8, pf *rfb.PixelFormat, rect *rfb.Rectangle) error {
	paletteSize := subEnc
	palette := make([]*color.RGBA, paletteSize)
	var err error
	// Read palette
	for j := 0; j < int(paletteSize); j++ {
		palette[j], err = readCPixel(cv, unZipper, pf)
		if err != nil {
			return fmt.Errorf("renderZRLE: error while reading CPixel for palette tile: %v", err)
		}
	}
	// Calculate index size
	var indexBits, mask uint32
	if paletteSize == 2 {
		indexBits = 1
		mask = 0x80
	} else if paletteSize <= 4 {
		indexBits = 2
		mask = 0xC0
	} else {
		indexBits = 4
		mask = 0xF0
	}
	for y := 0; y < tileHeight; y++ {

		// Packing only occurs per-row
		bitsAvailable := uint32(0)
		buffer := uint32(0)

		for x := 0; x < tileWidth; x++ {

			// Buffer more bits if necessary
			if bitsAvailable == 0 {
				bits, err := ReadUint8(unZipper)
				if err != nil {
					return fmt.Errorf("renderZRLE: error while reading first uint8 into buffer: %v", err)
				}
				buffer = uint32(bits)
				bitsAvailable = 8
			}

			// Read next pixel
			index := (buffer & mask) >> (8 - indexBits)
			buffer <<= indexBits
			bitsAvailable -= indexBits

			// Write pixel to image
			cv.Set(tileOffsetX+int(rect.X)+x, tileOffsetY+int(rect.Y)+y, palette[index])
		}
	}
	return err
}

// 普通rle编码
func (that *ZRLEEncoding) handlePlainRLETile(cv *canvas.VncCanvas, unZipper io.Reader, tileOffsetX int, tileOffsetY int, tileWidth int, tileHeight int, pf *rfb.PixelFormat, rect *rfb.Rectangle) error {
	var col *color.RGBA
	var err error
	runLen := 0
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {

			if runLen == 0 {
				// Read length and color
				col, err = readCPixel(cv, unZipper, pf)
				if err != nil {
					return fmt.Errorf("handlePlainRLETile: error while reading CPixel in plain RLE subencoding: %v", err)
				}
				runLen, err = readRunLength(unZipper)
				if err != nil {
					return fmt.Errorf("handlePlainRLETile: error while reading runlength in plain RLE subencoding: %v", err)
				}
			}
			// Write pixel to image
			cv.Set(tileOffsetX+int(rect.X)+x, tileOffsetY+int(rect.Y)+y, col)
			runLen--
		}
	}
	return err
}

func readRunLength(r io.Reader) (int, error) {
	runLen := 1

	addition, err := ReadUint8(r)
	if err != nil {
		return 0, fmt.Errorf("renderZRLE: error while reading addition to runLen in plain RLE subencoding: %v", err)
	}
	runLen += int(addition)

	for addition == 255 {
		addition, err = ReadUint8(r)
		if err != nil {
			return 0, fmt.Errorf("renderZRLE: error while reading addition to runLen in-loop plain RLE subencoding: %v", err)
		}
		runLen += int(addition)
	}
	return runLen, nil
}

// 调色板rle编码
func (that *ZRLEEncoding) handlePaletteRLETile(cv *canvas.VncCanvas, unZipper io.Reader, tileOffsetX, tileOffsetY, tileWidth, tileHeight int, subEnc uint8, pf *rfb.PixelFormat, rect *rfb.Rectangle) error {
	// Palette RLE
	paletteSize := subEnc - 128
	palette := make([]*color.RGBA, paletteSize)
	var err error

	// Read RLE palette
	for j := 0; j < int(paletteSize); j++ {
		palette[j], err = readCPixel(cv, unZipper, pf)
		if err != nil {
			return fmt.Errorf("renderZRLE: error while reading color in palette RLE subencoding: %v", err)
		}
	}
	var index uint8
	runLen := 0
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {
			if runLen == 0 {
				// Read length and index
				index, err = ReadUint8(unZipper)
				if err != nil {
					return fmt.Errorf("renderZRLE: error while reading length and index in palette RLE subencoding: %v", err)
				}
				runLen = 1
				// Run is represented by index | 0x80
				// Otherwise, single pixel
				if (index & 0x80) != 0 {
					index -= 128
					runLen, err = readRunLength(unZipper)
					if err != nil {
						return fmt.Errorf("handlePlainRLETile: error while reading runlength in plain RLE subencoding: %v", err)
					}

				}
			}
			// Write pixel to image
			cv.Set(tileOffsetX+int(rect.X)+x, tileOffsetY+int(rect.Y)+y, palette[index])
			runLen--
		}
	}
	return nil
}
