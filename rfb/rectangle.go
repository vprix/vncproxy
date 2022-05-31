package rfb

import (
	"encoding/binary"
	"fmt"
)

// Rectangle 表示像素数据的矩形
type Rectangle struct {
	X       uint16
	Y       uint16
	Width   uint16
	Height  uint16
	EncType EncodingType
	Enc     IEncoding
}

func NewRectangle() *Rectangle {
	return &Rectangle{}
}
func (that *Rectangle) String() string {
	return fmt.Sprintf("X:%d,Y:%d,Width:%d,Height:%d,EncType:%s", that.X, that.Y, that.Width, that.Height, that.EncType)
}

// 读取矩形数据
func (that *Rectangle) Read(sess ISession) error {
	var err error

	//读取x坐标
	if err = binary.Read(sess, binary.BigEndian, &that.X); err != nil {
		return err
	}
	// 读取y坐标
	if err = binary.Read(sess, binary.BigEndian, &that.Y); err != nil {
		return err
	}
	// 读取x坐标上的宽度
	if err = binary.Read(sess, binary.BigEndian, &that.Width); err != nil {
		return err
	}
	// 读取y坐标上的高度
	if err = binary.Read(sess, binary.BigEndian, &that.Height); err != nil {
		return err
	}
	// 读取编码类型
	if err = binary.Read(sess, binary.BigEndian, &that.EncType); err != nil {
		return err
	}
	that.Enc = sess.GetEncoding(that.EncType)
	if that.Enc == nil {
		return fmt.Errorf("不支持的编码类型: %s", that.EncType)
	}
	return that.Enc.Read(sess, that)
}

// 写入矩形数据
func (that *Rectangle) Write(sess ISession) error {
	var err error

	if err = binary.Write(sess, binary.BigEndian, that.X); err != nil {
		return err
	}
	if err = binary.Write(sess, binary.BigEndian, that.Y); err != nil {
		return err
	}
	if err = binary.Write(sess, binary.BigEndian, that.Width); err != nil {
		return err
	}
	if err = binary.Write(sess, binary.BigEndian, that.Height); err != nil {
		return err
	}
	if err = binary.Write(sess, binary.BigEndian, that.EncType); err != nil {
		return err
	}

	// 通过预定义的编码格式写入
	return that.Enc.Write(sess, that)
}

func (that *Rectangle) Clone() *Rectangle {
	r := &Rectangle{
		X:       that.X,
		Y:       that.Y,
		Width:   that.Width,
		Height:  that.Height,
		EncType: that.EncType,
		Enc:     that.Enc.Clone(true),
	}
	return r
}
