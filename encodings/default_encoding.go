package encodings

import "github.com/vprix/vncproxy/rfb"

var (
	DefaultEncodings = []rfb.IEncoding{
		&ZRLEEncoding{},
		//&TightEncoding{},
		&HexTileEncoding{},
		//&TightPngEncoding{},
		//&RREEncoding{},
		//&ZLibEncoding{},
		//&CopyRectEncoding{},
		//&CoRREEncoding{},
		&RawEncoding{},
		&CursorPseudoEncoding{},
		&DesktopNamePseudoEncoding{},
		&DesktopSizePseudoEncoding{},
		&CursorPosPseudoEncoding{},
		&ExtendedDesktopSizePseudo{},
		&CursorWithAlphaPseudoEncoding{},
		&LedStatePseudo{},
		&XCursorPseudoEncoding{},
	}
)
