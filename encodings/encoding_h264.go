package encodings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/vprix/vncproxy/rfb"
)

type H264Encoding struct {
	buff *bytes.Buffer
}

var _ rfb.IEncoding = new(H264Encoding)

func (that *H264Encoding) Type() rfb.EncodingType {
	return rfb.EncH264
}

func (that *H264Encoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *H264Encoding) Clone(data ...bool) rfb.IEncoding {
	obj := &ZLibEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *H264Encoding) Read(session rfb.ISession, rectangle *rfb.Rectangle) error {
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	size, err := ReadUint32(session)
	if err != nil {
		return err
	}
	err = binary.Write(that.buff, binary.BigEndian, size)
	if err != nil {
		return err
	}
	flags, err := ReadUint32(session)
	if err != nil {
		return err
	}
	err = binary.Write(that.buff, binary.BigEndian, flags)
	if err != nil {
		return err
	}
	b, err := ReadBytes(int(size), session)
	if err != nil {
		return err
	}
	_, err = that.buff.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (that *H264Encoding) Write(session rfb.ISession, rectangle *rfb.Rectangle) error {
	if that.buff == nil {
		return errors.New("ByteBuffer is nil")
	}
	_, err := that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}
