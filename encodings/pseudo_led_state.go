package encodings

import (
	"encoding/binary"
	"github.com/vprix/vncproxy/rfb"
)

// LedStatePseudo 切换客户端本地小键盘锁定的led灯
// 0 滚动锁
// 1 数字锁定
// 2 大写锁定
type LedStatePseudo struct {
	LedState uint8
}

func (that *LedStatePseudo) Supported(session rfb.ISession) bool {
	return true
}

func (that *LedStatePseudo) Clone(data ...bool) rfb.IEncoding {
	obj := &LedStatePseudo{}
	return obj
}

func (that *LedStatePseudo) Type() rfb.EncodingType {
	return rfb.EncLedStatePseudo
}

func (that *LedStatePseudo) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	u8, err := ReadUint8(session)
	if err != nil {
		return err
	}
	that.LedState = u8
	return nil
}

func (that *LedStatePseudo) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if err := binary.Write(session, binary.BigEndian, that.LedState); err != nil {
		return err
	}
	return nil
}
