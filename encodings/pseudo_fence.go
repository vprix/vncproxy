package encodings

import (
	"github.com/vprix/vncproxy/rfb"
)

type FencePseudo struct {
}

func (that *FencePseudo) Supported(_ rfb.ISession) bool {
	return true
}

func (that *FencePseudo) Clone(_ ...bool) rfb.IEncoding {
	obj := &FencePseudo{}
	return obj
}

func (that *FencePseudo) Type() rfb.EncodingType {
	return rfb.EncFencePseudo
}

func (that *FencePseudo) Read(_ rfb.ISession, _ *rfb.Rectangle) error {
	return nil
}

func (that *FencePseudo) Write(_ rfb.ISession, _ *rfb.Rectangle) error {

	return nil
}
