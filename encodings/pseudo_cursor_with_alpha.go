package encodings

import (
	"bytes"
	"github.com/vprix/vncproxy/rfb"
)

type CursorWithAlphaPseudoEncoding struct {
	buff *bytes.Buffer
}

func (that *CursorWithAlphaPseudoEncoding) Supported(session rfb.ISession) bool {
	return true
}

func (that *CursorWithAlphaPseudoEncoding) Clone(data ...bool) rfb.IEncoding {
	obj := &CursorWithAlphaPseudoEncoding{}
	if len(data) > 0 && data[0] {
		if that.buff != nil {
			obj.buff = &bytes.Buffer{}
			_, _ = obj.buff.Write(that.buff.Bytes())
		}
	}
	return obj
}

func (that *CursorWithAlphaPseudoEncoding) Type() rfb.EncodingType {
	return rfb.EncCursorWithAlphaPseudo
}

func (that *CursorWithAlphaPseudoEncoding) Read(session rfb.ISession, rect *rfb.Rectangle) error {
	if rect.Width*rect.Height == 0 {
		return nil
	}
	if that.buff == nil {
		that.buff = &bytes.Buffer{}
	}
	var bt []byte
	var err error
	bt, err = ReadBytes(4, session)
	if err != nil {
		return err
	}
	_, _ = that.buff.Write(bt)
	return nil
}

func (that *CursorWithAlphaPseudoEncoding) Write(session rfb.ISession, rect *rfb.Rectangle) error {
	if that.buff == nil {
		return nil
	}
	var err error
	_, err = that.buff.WriteTo(session)
	that.buff.Reset()
	return err
}
