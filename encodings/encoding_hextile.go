package encodings

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/errors/gerror"
	"github.com/vprix/vncproxy/canvas"
	"github.com/vprix/vncproxy/rfb"
	"image"
	"image/color"
)

const (
	HexTileRaw                 = 1 << 0 // Raw数据：不压缩，直接传送，一般此位置1时其它位都置0
	HexTileBackgroundSpecified = 1 << 1 // 包含背景色数据：标志位之后需要接收背景色数据
	HexTileForegroundSpecified = 1 << 2 // 包含前景色数据：背景色之后需要接收前景色数据
	HexTileAnySubRects         = 1 << 3 // 是否含有子块：只要该块中含有两种及两种以上颜色，则此位置1
	HexTileSubRectsColoured    = 1 << 4 // 子块的颜色：如果含有两种颜色，此位置0，子块颜色用前景色；若该块中含有两种以上的颜色，此位置1，子块颜色需要单独指明
)

// HexTileEncoding 是RRE编码的变种，把屏幕分成16x16象素的小块，每块用Raw或RRE方式转送.
// 通过解释HexTile算法，说明了简单而常用的屏幕传送和压缩算法，希望对屏幕监测、传送相关的工作有所启发
type HexTileEncoding struct {
	buff *bytes.Buffer
}

var _ rfb.IEncoding = new(HexTileEncoding)

func (that *HexTileEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *HexTileEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &HexTileEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *HexTileEncoding) Type() rfb.EncodingType {
	return rfb.EncHexTile
}

//
// 1. 分割
//    a. 传送的图像区域被分为若干个大小为16×16象素的块，如果整个矩形不是16的倍数，则最后1行（或列）的块宽度（或高度）变小
//    b. 这些块按从左到右，从上到下的顺序排列，屏幕中变化的块被传送，不变的不被传送.
//    c. 由于块大小是16x16，所以块内坐标XY可用一字节表示，WH可用一个字节表示
// 2. 块内部的编码
//    a. 计算块内部的颜色数：一种、两种、多种，并记录出现频率最多的颜色为背景色，如果仅有两种颜色，则另一颜色记为前景色
//    b. 判断块内颜色数是否为一种，如果是，则先修改传送标志位，然后传送整块大小（0，0，w，h）和背景色，此块传送即完成.
//    c. 如果不是一种，则把块拆分成颜色不同的小矩形，方法如下：
//       1. 先把块复制到一块内存区中，以免破坏原始数据，暂称tmpBuf
//       2. 从第一个象素开始判断，该点颜色是否与背景色相同.
//       3. 如果不同，则分别向右和向下求得与该点颜色连续的色块.
//       4. 对比右色块和下色块，取出其中较大的一个，做为一个矩形色块
//       5. 在tmpBuf中把此矩形填成背景色，以避免重复判断
//       6. 继续判断下一象素点……
//    d. 记录各个矩形色块的位置（x,y,w,h），如果块内含两种以上颜色，还要记录矩形色块的颜色值
//    e. 一边取得矩形色块，一边判断矩形色块描述数据的总长是否大于原始数据，如果大于原始数据，则放弃取色块，标志字节Raw(HexTileRaw)位置
//       1. 以Raw方式直接传送原始数据，此块传送完成
//    f. 如果块含两种以上颜色，则将标志位的子块位（HexTileSubRectsColoured）置1，否则置0
//    g. 传送标志位，传送矩形色块个数，然后传送各矩形块数据，块中颜色数为2时不需要传送每个矩形块的颜色数据，只传位置即可.
//    h. 注意：如果背景色或前景色与前一块(16x16块)相同，则可以不传送背景色或前景色，客户端会默认延用前一块的
func (that *HexTileEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	bytesPerPixel := int(session.PixelFormat().BPP) / 8
	// 从上到下
	for ty := rect.Y; ty < rect.Y+rect.Height; ty += 16 {
		th := 16
		//  如果整个矩形不是16的倍数，则最后1列的块高度为实际高度
		if rect.Y+rect.Height-ty < 16 {
			th = int(rect.Y) + int(rect.Height) - int(ty)
		}
		// 从左到右
		for tx := rect.X; tx < rect.X+rect.Width; tx += 16 {
			tw := 16
			//  如果整个矩形不是16的倍数，则最后1行的块宽度为实际宽度
			if rect.X+rect.Width-tx < 16 {
				tw = int(rect.X) + int(rect.Width) - int(tx)
			}
			var bt []byte
			// 读取标志位
			subEncoding, err := ReadUint8(session)
			if err != nil {
				return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
			}
			_ = binary.Write(that.buff, session.PixelFormat().Order(), subEncoding)
			// 如果是原始编码
			if (subEncoding & HexTileRaw) != 0 {
				bt, err = ReadBytes(tw*th*bytesPerPixel, session)
				if err != nil {
					return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
				}
				_, _ = that.buff.Write(bt)
				continue
			}

			// 包含背景色数据
			if (subEncoding & HexTileBackgroundSpecified) != 0 {
				bt, err = ReadBytes(bytesPerPixel, session)
				if err != nil {
					return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
				}
				_, _ = that.buff.Write(bt)
			}

			// 包含前景色数据
			if (subEncoding & HexTileForegroundSpecified) != 0 {
				bt, err = ReadBytes(bytesPerPixel, session)
				if err != nil {
					return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
				}
				_, _ = that.buff.Write(bt)
			}

			// 不包含子块则跳过
			if (subEncoding & HexTileAnySubRects) != 0 {
				nSubRects, err := ReadUint8(session)
				if err != nil {
					return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
				}
				_ = binary.Write(that.buff, session.PixelFormat().Order(), nSubRects)

				for i := 0; i < int(nSubRects); i++ {
					if (subEncoding & HexTileSubRectsColoured) != 0 {
						bt, err = ReadBytes(bytesPerPixel, session)
						if err != nil {
							return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
						}
						_, _ = that.buff.Write(bt)
					}
					xy, err := ReadUint8(session)
					if err != nil {
						return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
					}
					_ = binary.Write(that.buff, session.PixelFormat().Order(), xy)

					wh, err := ReadUint8(session)
					if err != nil {
						return gerror.Newf("HextileEncoding.Read: error in hextile reader: %v", err)
					}
					_ = binary.Write(that.buff, session.PixelFormat().Order(), wh)
				}
			}
		}
	}
	return nil
}

func (that *HexTileEncoding) Write(sess rfb.ISession, rect *rfb.Rectangle) error {
	if sess.Type() == rfb.CanvasSessionType {
		return that.draw(sess.Conn().(*canvas.VncCanvas), sess.PixelFormat(), rect)
	}
	var err error
	_, err = that.buff.WriteTo(sess)
	that.buff.Reset()
	return err
}

func (that *HexTileEncoding) draw(cv *canvas.VncCanvas, pf rfb.PixelFormat, rect *rfb.Rectangle) error {
	var bgCol *color.RGBA
	var fgCol *color.RGBA
	var err error
	var subEncoding byte
	var dimensions byte
	var nSubRects uint8

	// 从上到下
	for ty := rect.Y; ty < rect.Y+rect.Height; ty += 16 {
		th := 16
		//  如果整个矩形不是16的倍数，则最后1列的块高度为实际高度
		if rect.Y+rect.Height-ty < 16 {
			th = int(rect.Y) + int(rect.Height) - int(ty)
		}
		// 从左到右
		for tx := rect.X; tx < rect.X+rect.Width; tx += 16 {
			tw := 16
			//  如果整个矩形不是16的倍数，则最后1行的块宽度为实际宽度
			if rect.X+rect.Width-tx < 16 {
				tw = int(rect.X) + int(rect.Width) - int(tx)
			}
			subEncoding, err = ReadUint8(that.buff)
			if err != nil {
				return fmt.Errorf("HextileEncoding.Read: error in hextile reader: %v", err)
			}
			// 如果是原始编码
			if (subEncoding & HexTileRaw) != 0 {
				err = cv.DecodeRaw(that.buff, &pf, &rfb.Rectangle{X: tx, Y: ty, Width: uint16(tw), Height: uint16(th), EncType: rfb.EncRaw})
				if err != nil {
					return err
				}
				continue
			}
			// 读取单个背景颜色
			if (subEncoding & HexTileBackgroundSpecified) != 0 {
				bgCol, err = cv.ReadColor(that.buff, &pf)
				if err != nil {
					return fmt.Errorf("HexTileEncoding.Read: error in hexTile bg color reader: %v", err)
				}
			}
			// 绘制一个矩形
			rBounds := image.Rectangle{
				Min: image.Point{X: int(tx), Y: int(ty)},
				Max: image.Point{X: int(tx) + tw, Y: int(ty) + th},
			}
			// 填充背景色
			cv.FillRect(&rBounds, bgCol)

			// 读取前景色
			if (subEncoding & HexTileForegroundSpecified) != 0 {
				fgCol, err = cv.ReadColor(that.buff, &pf)
				if err != nil {
					return fmt.Errorf("HexTileEncoding.Read: error in hexTile fg color reader: %v", err)
				}
			}
			if (subEncoding & HexTileAnySubRects) == 0 {
				continue
			}
			// 读取子块的个数
			nSubRects, err = ReadUint8(that.buff)
			if err != nil {
				return err
			}
			// 是否指定子块的填充颜色，如果未指定，则使用前景色
			colorSpecified := (subEncoding & HexTileSubRectsColoured) != 0
			for i := 0; i < int(nSubRects); i++ {
				var co *color.RGBA
				if colorSpecified {
					co, err = cv.ReadColor(that.buff, &pf)
					if err != nil {
						return fmt.Errorf("HexTileEncoding.Read: problem reading color from connection: %v", err)
					}
				} else {
					co = fgCol
				}
				fgCol = co
				dimensions, err = ReadUint8(that.buff) // bits 7-4 for x, bits 3-0 for y
				if err != nil {
					return fmt.Errorf("HexTileEncoding.Read: problem reading dimensions from connection: %v", err)
				}
				subTileX := dimensions >> 4 & 0x0f
				subTileY := dimensions & 0x0f
				dimensions, err = ReadUint8(that.buff) // bits 7-4 for x, bits 3-0 for y
				if err != nil {
					return fmt.Errorf("HexTileEncoding.Read: problem reading dimensions from connection: %v", err)
				}
				subTileWidth := 1 + (dimensions >> 4 & 0x0f)
				subTileHeight := 1 + (dimensions & 0x0f)
				subRectBounds := image.Rectangle{
					Min: image.Point{X: int(tx) + int(subTileX), Y: int(ty) + int(subTileY)},
					Max: image.Point{X: int(tx) + int(subTileX) + int(subTileWidth), Y: int(ty) + int(subTileY) + int(subTileHeight)},
				}
				cv.FillRect(&subRectBounds, fgCol)
			}
		}
	}
	return nil
}
