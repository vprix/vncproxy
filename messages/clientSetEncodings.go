package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/gogf/gf/text/gstr"
	"github.com/vprix/vncproxy/rfb"
)

// SetEncodings 设置编码类型消息
type SetEncodings struct {
	_         [1]byte // padding
	EncNum    uint16  // number-of-encodings
	Encodings []rfb.EncodingType
}

func (that *SetEncodings) Clone() rfb.ClientMessage {
	c := &SetEncodings{
		EncNum:    that.EncNum,
		Encodings: that.Encodings,
	}
	return c
}
func (that *SetEncodings) Supported(session rfb.ISession) bool {
	return true
}

// String return string
func (that *SetEncodings) String() string {
	s := fmt.Sprintf("encNum: %d, encodings[]: ", that.EncNum)
	var s1 []string
	for _, e := range that.Encodings {
		s1 = append(s1, fmt.Sprintf("%s", e))
	}
	return s + gstr.Implode(",", s1)
}

// Type returns MessageType
func (that *SetEncodings) Type() rfb.ClientMessageType {
	return rfb.SetEncodings
}

// Read 从会话中解析消息内容
func (that *SetEncodings) Read(session rfb.ISession) (rfb.ClientMessage, error) {
	msg := &SetEncodings{}
	//读取一个字节的填充数据
	var pad [1]byte
	if err := binary.Read(session, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	//读取编码格式数量
	if err := binary.Read(session, binary.BigEndian, &msg.EncNum); err != nil {
		return nil, err
	}
	var enc rfb.EncodingType
	//读取指定数据量的编码信息
	for i := uint16(0); i < msg.EncNum; i++ {
		if err := binary.Read(session, binary.BigEndian, &enc); err != nil {
			return nil, err
		}
		msg.Encodings = append(msg.Encodings, enc)
	}
	if err := session.SetEncodings(msg.Encodings); err != nil {
		return nil, err
	}

	return msg, nil
}

// Write 把消息按协议格式写入会话
func (that *SetEncodings) Write(session rfb.ISession) error {
	// 写入消息类型
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	// 写入一个字节的填充数据
	var pad [1]byte
	if err := binary.Write(session, binary.BigEndian, pad); err != nil {
		return err
	}
	// 写入当前支持的编码类型的数量
	if uint16(len(that.Encodings)) > that.EncNum {
		that.EncNum = uint16(len(that.Encodings))
	}
	if err := binary.Write(session, binary.BigEndian, that.EncNum); err != nil {
		return err
	}
	// 写入当前支持的编码类型的列表
	for _, enc := range that.Encodings {
		if err := binary.Write(session, binary.BigEndian, enc); err != nil {
			return err
		}
	}
	return session.Flush()
}
