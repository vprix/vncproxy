package encodings

import (
	"github.com/vprix/vncproxy/rfb"
)

type LastRectPseudo struct {
}

func (that *LastRectPseudo) Supported(_ rfb.ISession) bool {
	return true
}

func (that *LastRectPseudo) Clone(_ ...bool) rfb.IEncoding {
	obj := &LastRectPseudo{}
	return obj
}

func (that *LastRectPseudo) Type() rfb.EncodingType {
	return rfb.EncLastRectPseudo
}

func (that *LastRectPseudo) Read(_ rfb.ISession, _ *rfb.Rectangle) error {
	return nil
}

func (that *LastRectPseudo) Write(_ rfb.ISession, _ *rfb.Rectangle) error {

	return nil
}
