package encodings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/vprix/vncproxy/rfb"
)

type ZLibEncoding struct {
	buff *bytes.Buffer
}

func (that *ZLibEncoding) Supported(c rfb.ISession) bool {
	return true
}
func (that *ZLibEncoding) Type() rfb.EncodingType {
	return rfb.EncZlib
}

func (that *ZLibEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &ZLibEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *ZLibEncoding) Read(sess rfb.ISession, rect *rfb.Rectangle) error {
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

func (that *ZLibEncoding) Write(sess rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	_, err := that.buff.WriteTo(sess)
	that.buff.Reset()
	return err
}
