package messages

import (
	"encoding/binary"
	"fmt"
	"github.com/vprix/vncproxy/rfb"
)

type QEMUExtKeyEvent struct {
	SubMessageType uint8   // submessage type
	DownFlag       uint16  // down-flag
	KeySym         rfb.Key // key symbol
	KeyCode        uint32  // scan code
}

func (that *QEMUExtKeyEvent) Clone() rfb.ClientMessage {

	c := &QEMUExtKeyEvent{
		SubMessageType: that.SubMessageType,
		DownFlag:       that.DownFlag,
		KeySym:         that.KeySym,
		KeyCode:        that.KeyCode,
	}
	return c
}
func (that *QEMUExtKeyEvent) Supported(session rfb.ISession) bool {
	return true
}
func (that *QEMUExtKeyEvent) Type() rfb.ClientMessageType {
	return rfb.QEMUExtendedKeyEvent
}

func (that *QEMUExtKeyEvent) String() string {
	return fmt.Sprintf("SubMessageType=%d,DownFlag=%d,KeySym=%d,KeyCode=%d", that.SubMessageType, that.DownFlag, that.KeySym, that.KeyCode)
}

func (that *QEMUExtKeyEvent) Read(session rfb.ISession) (rfb.ClientMessage, error) {
	msg := &QEMUExtKeyEvent{}
	if err := binary.Read(session, binary.BigEndian, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (that *QEMUExtKeyEvent) Write(session rfb.ISession) error {
	if err := binary.Write(session, binary.BigEndian, that.Type()); err != nil {
		return err
	}
	if err := binary.Write(session, binary.BigEndian, that); err != nil {
		return err
	}
	return nil
}
