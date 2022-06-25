package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

// SetColorMapEntries 设置颜色表的内容
//  See RFC 6143 Section 7.6.2
type SetColorMapEntries struct {
	_          [1]byte //填充
	FirstColor uint16  // 颜色的起始位置，
	ColorsNum  uint16  // 颜色的数目
	Colors     []rfb.Color
}

func (that *SetColorMapEntries) Clone() rfb.Message {

	c := &SetColorMapEntries{
		FirstColor: that.FirstColor,
		ColorsNum:  that.ColorsNum,
		Colors:     that.Colors,
	}
	return c
}
func (that *SetColorMapEntries) Supported(session rfb.ISession) bool {
	return true
}

// String returns string
func (that *SetColorMapEntries) String() string {
	return fmt.Sprintf("first color: %d, numcolors: %d, colors[]: { %v }", that.FirstColor, that.ColorsNum, that.Colors)
}

// Type returns MessageType
func (*SetColorMapEntries) Type() rfb.MessageType {
	return rfb.MessageType(rfb.SetColorMapEntries)
}

func (that *SetColorMapEntries) Read(session rfb.ISession) (rfb.Message, error) {
	msg := &SetColorMapEntries{}
	// 先读取一个字节的填充
	var pad [1]byte
	if err := binary.Read(session, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	// 单个消息不必指定整个色彩映射表，而可能能只更新几个条目。
	//例如，如果我想更新条目 5 和 6，我会在FirstColor中指定，后跟两组 RGB 值。first-colour:5 number-of-colours:2
	if err := binary.Read(session, binary.BigEndian, &msg.FirstColor); err != nil {
		return nil, err
	}
	// 获取此次要更新几个颜色
	if err := binary.Read(session, binary.BigEndian, &msg.ColorsNum); err != nil {
		return nil, err
	}

	msg.Colors = make([]rfb.Color, msg.ColorsNum)
	colorMap := session.Desktop().ColorMap()
	//读取指定的颜色数据
	for i := uint16(0); i < msg.ColorsNum; i++ {
		color := &msg.Colors[i]
		err := color.Read(session)
		if err != nil {
			return nil, err
		}
		colorMap[msg.FirstColor+i] = *color
	}
	session.Desktop().SetColorMap(colorMap)
	return msg, nil
}

func (that *SetColorMapEntries) Write(session rfb.ISession) error {

	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	// 填充
	var pad [1]byte
	if err := binary.Write(session, binary.BigEndian, &pad); err != nil {
		return err
	}

	// 首个颜色
	if err := binary.Write(session, binary.BigEndian, that.FirstColor); err != nil {
		return err
	}
	// 要更新的颜色数目
	if that.ColorsNum < uint16(len(that.Colors)) {
		that.ColorsNum = uint16(len(that.Colors))
	}
	if err := binary.Write(session, binary.BigEndian, that.ColorsNum); err != nil {
		return err
	}

	// 颜色数据
	for i := 0; i < len(that.Colors); i++ {
		color := that.Colors[i]
		if err := binary.Write(session, binary.BigEndian, color); err != nil {
			return err
		}
	}

	return session.Flush()
}
