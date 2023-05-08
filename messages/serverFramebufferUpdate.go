package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/osgochina/dmicro/logger"
	"github.com/vprix/vncproxy/rfb"
	"golang.org/x/net/context"
)

// FramebufferUpdate 帧缓冲更新
type FramebufferUpdate struct {
	_       [1]byte          // 填充
	NumRect uint16           // 多少个像素数据的矩形
	Rects   []*rfb.Rectangle // 像素数据的矩形列表
}

func (that *FramebufferUpdate) String() string {
	return fmt.Sprintf("rects %d rectangle[]: { %v }", that.NumRect, that.Rects)
}
func (that *FramebufferUpdate) Supported(rfb.ISession) bool {
	return true
}

func (that *FramebufferUpdate) Type() rfb.MessageType {
	return rfb.MessageType(rfb.FramebufferUpdate)
}

// 读取帧数据
func (that *FramebufferUpdate) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &FramebufferUpdate{}
	var pad [1]byte
	if err := binary.Read(session, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(session, binary.BigEndian, &msg.NumRect); err != nil {
		return nil, err
	}
	if logger.IsDebug() {
		logger.Debugf(context.TODO(), "FramebufferUpdate->读取帧数据有 %d 个矩形-------", msg.NumRect)
	}

	for i := uint16(0); i < msg.NumRect; i++ {
		rect := rfb.NewRectangle()
		if logger.IsDebug() {
			logger.Debugf(context.TODO(), "开始读取第 %d 个矩形", i)
		}

		if err := rect.Read(session); err != nil {
			return nil, err
		}
		// 如果服务器告诉客户端这是最后一个rect，则停止解析
		if rect.EncType == rfb.EncLastRectPseudo {
			if logger.IsDebug() {
				logger.Debugf(context.TODO(), "读取第 %d 个矩形成功，但是是最后一帧:EncLastRectPseudo", i)
			}
			msg.Rects = append(msg.Rects, rect)
			break
		}
		//if rect.EncType == rfb.EncDesktopSizePseudo {
		//	session.ResetAllEncodings()
		//}
		if logger.IsDebug() {
			logger.Debugf(context.TODO(), "结束读取第 %d 个矩形,宽高:(%dx%d) 编码格式:%s", i, rect.Width, rect.Height, rect.EncType)
		}
		msg.Rects = append(msg.Rects, rect)
	}
	return msg, nil
}

// 写入帧数据
func (that *FramebufferUpdate) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	// 填充字节
	var pad [1]byte
	if err := binary.Write(session, binary.BigEndian, pad); err != nil {
		return err
	}
	// 写入矩形数量
	if err := binary.Write(session, binary.BigEndian, that.NumRect); err != nil {
		return err
	}
	// 编码后写入
	for _, rect := range that.Rects {
		if err := rect.Write(session); err != nil {
			return err
		}
	}
	return session.Flush()
}

func (that *FramebufferUpdate) Clone() rfb.Message {

	c := &FramebufferUpdate{
		NumRect: that.NumRect,
	}
	for _, rect := range that.Rects {
		c.Rects = append(c.Rects, rect.Clone())
	}
	return c
}
